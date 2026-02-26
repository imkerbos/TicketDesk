// Package telegram 提供 Telegram 通知服务
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	configService "github.com/kerbos/ticketdesk/internal/system-config/service"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
)

// TelegramService Telegram 通知服务接口
type TelegramService interface {
	// SendNotification 发送 Telegram 通知
	SendNotification(ctx context.Context, event string, data interface{}) error
	// SendTestMessage 发送测试消息
	SendTestMessage(ctx context.Context) error
	// IsEnabled 检查 Telegram 通知是否启用
	IsEnabled(ctx context.Context) bool
}

// telegramService Telegram 通知服务实现
type telegramService struct {
	configSvc  configService.ConfigService
	httpClient *http.Client
}

// telegramSendMessageRequest Telegram sendMessage API 请求体
type telegramSendMessageRequest struct {
	ChatID             string                  `json:"chat_id"`
	Text               string                  `json:"text"`
	ParseMode          string                  `json:"parse_mode"`
	LinkPreviewOptions *linkPreviewOptions     `json:"link_preview_options,omitempty"`
	ReplyMarkup        *telegramInlineKeyboard `json:"reply_markup,omitempty"`
}

// linkPreviewOptions 链接预览选项
type linkPreviewOptions struct {
	IsDisabled bool `json:"is_disabled"`
}

// telegramInlineKeyboard Telegram 内联键盘
type telegramInlineKeyboard struct {
	InlineKeyboard [][]telegramInlineButton `json:"inline_keyboard"`
}

// telegramInlineButton Telegram 内联按钮
type telegramInlineButton struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// NewTelegramService 创建 Telegram 通知服务实例
func NewTelegramService(configSvc configService.ConfigService) TelegramService {
	return &telegramService{
		configSvc: configSvc,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsEnabled 检查 Telegram 通知是否启用
func (s *telegramService) IsEnabled(ctx context.Context) bool {
	enabled, err := s.configSvc.GetConfigValue(ctx, configService.KeyTelegramEnabled)
	if err != nil {
		return false
	}
	return enabled == "true"
}

// SendNotification 发送 Telegram 通知
func (s *telegramService) SendNotification(ctx context.Context, event string, data interface{}) error {
	if !s.IsEnabled(ctx) {
		logger.Debug("telegram notification is disabled, skip sending")
		return nil
	}

	botToken, err := s.configSvc.GetConfigValue(ctx, configService.KeyTelegramBotToken)
	if err != nil || botToken == "" {
		return fmt.Errorf("Telegram Bot Token 未配置")
	}

	chatID, err := s.configSvc.GetConfigValue(ctx, configService.KeyTelegramChatID)
	if err != nil || chatID == "" {
		return fmt.Errorf("Telegram Chat ID 未配置")
	}

	// 构建消息
	text, replyMarkup := s.buildMessage(event, data)

	return s.doSend(ctx, botToken, chatID, text, replyMarkup)
}

// SendTestMessage 发送测试消息
func (s *telegramService) SendTestMessage(ctx context.Context) error {
	botToken, err := s.configSvc.GetConfigValue(ctx, configService.KeyTelegramBotToken)
	if err != nil || botToken == "" {
		return fmt.Errorf("Telegram Bot Token 未配置")
	}

	chatID, err := s.configSvc.GetConfigValue(ctx, configService.KeyTelegramChatID)
	if err != nil || chatID == "" {
		return fmt.Errorf("Telegram Chat ID 未配置")
	}

	text := "🔔 <b>TicketDesk 通知测试</b>\n\n" +
		"✅ 恭喜！Telegram 通知配置成功。\n\n" +
		"此消息由 TicketDesk 系统发送，用于验证 Telegram 通知功能是否正常工作。\n\n" +
		fmt.Sprintf("<i>%s</i>", html.EscapeString(time.Now().Format("2006-01-02 15:04:05")))

	return s.doSend(ctx, botToken, chatID, text, nil)
}

// buildMessage 根据事件类型构建 Telegram 消息（HTML 格式）
func (s *telegramService) buildMessage(event string, data interface{}) (string, *telegramInlineKeyboard) {
	dataMap := toMap(data)

	// 获取站点 URL
	ctx := context.Background()
	siteURL, _ := s.configSvc.GetConfigValue(ctx, configService.KeyGeneralSiteURL)
	if siteURL == "" {
		siteURL = "https://ticketdesk.example.com"
	}

	title := s.getEventTitle(event)
	var text string
	var replyMarkup *telegramInlineKeyboard

	// 告警来源的工单创建，使用独立标题
	if event == "issue.created" {
		if source, _ := dataMap["source"].(string); source == "alert" {
			title = "🚨 告警建单"
		}
	}

	switch {
	case event == "issue.created" && dataMap["source"] == "alert":
		text, replyMarkup = buildAlertIssueTelegram(title, dataMap, siteURL)

	case len(event) > 6 && event[:6] == "issue.":
		issueKey, _ := dataMap["issue_key"].(string)
		issueTitle, _ := dataMap["issue_title"].(string)
		projectName, _ := dataMap["project_name"].(string)
		status, _ := dataMap["status"].(string)
		priority, _ := dataMap["priority"].(string)

		text = fmt.Sprintf("%s\n\n", title)
		text += fmt.Sprintf("<b>%s</b> %s\n\n", html.EscapeString(issueKey), html.EscapeString(issueTitle))
		if projectName != "" {
			text += fmt.Sprintf("📁 项目：%s\n", html.EscapeString(projectName))
		}
		if status != "" {
			text += fmt.Sprintf("📊 状态：%s\n", html.EscapeString(status))
		}
		if priority != "" {
			text += fmt.Sprintf("🔴 优先级：%s\n", html.EscapeString(priority))
		}
		if comment, ok := dataMap["comment"].(string); ok && comment != "" {
			if len(comment) > 200 {
				comment = comment[:200] + "..."
			}
			text += fmt.Sprintf("💬 评论：%s\n", html.EscapeString(comment))
		}

		// 添加查看按钮
		if issueKey != "" {
			replyMarkup = &telegramInlineKeyboard{
				InlineKeyboard: [][]telegramInlineButton{
					{
						{
							Text: "📋 查看工单",
							URL:  fmt.Sprintf("%s/issues/%s", siteURL, issueKey),
						},
					},
				},
			}
		}

	case len(event) > 6 && event[:6] == "alert.":
		alertName, _ := dataMap["alert_name"].(string)
		severity, _ := dataMap["severity"].(string)
		alertStatus, _ := dataMap["status"].(string)
		issueKey, _ := dataMap["issue_key"].(string)

		text = fmt.Sprintf("%s\n\n", title)
		text += fmt.Sprintf("<b>%s</b>\n\n", html.EscapeString(alertName))
		if severity != "" {
			text += fmt.Sprintf("⚠️ 级别：%s\n", html.EscapeString(severity))
		}
		if alertStatus != "" {
			text += fmt.Sprintf("📊 状态：%s\n", html.EscapeString(alertStatus))
		}
		if issueKey != "" {
			text += fmt.Sprintf("🎫 关联工单：%s\n", html.EscapeString(issueKey))
			replyMarkup = &telegramInlineKeyboard{
				InlineKeyboard: [][]telegramInlineButton{
					{
						{
							Text: "📋 查看工单",
							URL:  fmt.Sprintf("%s/issues/%s", siteURL, issueKey),
						},
					},
				},
			}
		}

	default:
		text = fmt.Sprintf("%s\n\n", title)
		jsonBytes, _ := json.MarshalIndent(dataMap, "", "  ")
		text += fmt.Sprintf("<pre>%s</pre>", html.EscapeString(string(jsonBytes)))
	}

	return text, replyMarkup
}

// buildAlertIssueTelegram 构建告警建单的 Telegram 消息
func buildAlertIssueTelegram(title string, data map[string]interface{}, siteURL string) (string, *telegramInlineKeyboard) {
	issueKey, _ := data["issue_key"].(string)
	alertName, _ := data["alert_name"].(string)
	projectName, _ := data["project_name"].(string)
	if projectName == "" {
		projectName, _ = data["project_key"].(string)
	}
	status, _ := data["status"].(string)
	priority, _ := data["priority"].(string)
	severity, _ := data["severity"].(string)
	assignee, _ := data["assignee"].(string)
	alertTime, _ := data["alert_time"].(string)

	priEmoji := "🟡"
	switch priority {
	case "P0":
		priEmoji = "🔴"
	case "P1":
		priEmoji = "🟠"
	case "P2":
		priEmoji = "🟡"
	case "P3":
		priEmoji = "🟢"
	}

	text := fmt.Sprintf("%s\n\n", title)
	text += fmt.Sprintf("<b>%s</b>\n", html.EscapeString(issueKey))
	text += fmt.Sprintf("%s\n\n", html.EscapeString(alertName))
	text += fmt.Sprintf("%s 优先级：<b>%s</b>　　📊 状态：<b>%s</b>\n", priEmoji, html.EscapeString(priority), html.EscapeString(status))
	if projectName != "" || assignee != "" {
		text += fmt.Sprintf("📁 项目：<b>%s</b>", html.EscapeString(projectName))
		if assignee != "" {
			text += fmt.Sprintf("　　👤 指派：<b>%s</b>", html.EscapeString(assignee))
		}
		text += "\n"
	}
	if severity != "" {
		text += fmt.Sprintf("⚠️ 级别：<b>%s</b>\n", html.EscapeString(severity))
	}
	if alertTime != "" {
		text += fmt.Sprintf("⏰ 告警时间：%s\n", html.EscapeString(alertTime))
	}

	var replyMarkup *telegramInlineKeyboard
	if issueKey != "" {
		replyMarkup = &telegramInlineKeyboard{
			InlineKeyboard: [][]telegramInlineButton{
				{
					{
						Text: "📋 查看工单",
						URL:  fmt.Sprintf("%s/issues/%s", siteURL, issueKey),
					},
				},
			},
		}
	}

	return text, replyMarkup
}

// getEventTitle 获取事件标题
func (s *telegramService) getEventTitle(event string) string {
	switch event {
	case "issue.created":
		return "🎫 <b>工单创建</b>"
	case "issue.updated":
		return "✏️ <b>工单更新</b>"
	case "issue.transitioned":
		return "🔄 <b>工单状态变更</b>"
	case "issue.assigned":
		return "👤 <b>工单指派</b>"
	case "issue.commented":
		return "💬 <b>工单评论</b>"
	case "alert.firing":
		return "🔥 <b>告警触发</b>"
	case "alert.resolved":
		return "✅ <b>告警恢复</b>"
	case "alert.acked":
		return "👁️ <b>告警确认</b>"
	default:
		return "📢 <b>系统通知</b>"
	}
}

// doSend 执行 Telegram API 调用
func (s *telegramService) doSend(ctx context.Context, botToken, chatID, text string, replyMarkup *telegramInlineKeyboard) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	reqBody := telegramSendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
		LinkPreviewOptions: &linkPreviewOptions{
			IsDisabled: true,
		},
		ReplyMarkup: replyMarkup,
	}

	payloadBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化 Telegram 消息失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("创建 Telegram 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("failed to send telegram notification", zap.Error(err))
		return fmt.Errorf("发送 Telegram 通知失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// 解析 Telegram API 响应
	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && !result.OK {
		logger.Error("telegram API returned error",
			zap.String("description", result.Description),
			zap.Int("status_code", resp.StatusCode),
		)
		return fmt.Errorf("Telegram API 错误: %s", result.Description)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Telegram 返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	logger.Info("telegram notification sent successfully", zap.String("chat_id", chatID))
	return nil
}

// ============ Direct Sender（项目级通知渠道使用，不依赖 ConfigService）============

// DirectTelegramSender 直接 Telegram 发送器（接受凭据参数，不依赖系统配置）
type DirectTelegramSender struct {
	botToken   string
	chatID     string
	siteURL    string
	httpClient *http.Client
}

// NewDirectTelegramSender 创建直接 Telegram 发送器
func NewDirectTelegramSender(botToken, chatID string, siteURL ...string) *DirectTelegramSender {
	url := ""
	if len(siteURL) > 0 {
		url = siteURL[0]
	}
	return &DirectTelegramSender{
		botToken: botToken,
		chatID:   chatID,
		siteURL:  url,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendNotification 发送 Telegram 通知
func (d *DirectTelegramSender) SendNotification(ctx context.Context, event string, data interface{}) error {
	text, replyMarkup := d.buildMessage(event, data)
	return d.doSend(ctx, text, replyMarkup)
}

// SendTestMessage 发送测试消息（使用与真实通知相同的模板）
func (d *DirectTelegramSender) SendTestMessage(ctx context.Context) error {
	testData := map[string]interface{}{
		"issue_key":    "TEST-1",
		"issue_title":  "这是一条测试通知，用于验证 Telegram 通知渠道是否配置正确",
		"project_name": "测试项目",
		"status":       "待处理",
		"priority":     "P2",
	}
	return d.SendNotification(ctx, "issue.created", testData)
}

// buildMessage 根据事件类型构建 Telegram 消息（HTML 格式）
func (d *DirectTelegramSender) buildMessage(event string, data interface{}) (string, *telegramInlineKeyboard) {
	dataMap := toMap(data)
	siteURL := d.siteURL
	if siteURL == "" {
		siteURL = "https://ticketdesk.example.com"
	}

	title := directGetEventTitle(event)
	var text string
	var replyMarkup *telegramInlineKeyboard

	// 告警来源的工单创建，使用独立标题
	if event == "issue.created" {
		if source, _ := dataMap["source"].(string); source == "alert" {
			title = "🚨 <b>告警建单</b>"
		}
	}

	switch {
	case event == "issue.created" && dataMap["source"] == "alert":
		text, replyMarkup = buildAlertIssueTelegram(title, dataMap, siteURL)

	case len(event) > 6 && event[:6] == "issue.":
		issueKey, _ := dataMap["issue_key"].(string)
		issueTitle, _ := dataMap["issue_title"].(string)
		projectName, _ := dataMap["project_name"].(string)
		status, _ := dataMap["status"].(string)
		priority, _ := dataMap["priority"].(string)

		text = fmt.Sprintf("%s\n\n", title)
		text += fmt.Sprintf("<b>%s</b> %s\n\n", html.EscapeString(issueKey), html.EscapeString(issueTitle))
		if projectName != "" {
			text += fmt.Sprintf("📁 项目：%s\n", html.EscapeString(projectName))
		}
		if status != "" {
			text += fmt.Sprintf("📊 状态：%s\n", html.EscapeString(status))
		}
		if priority != "" {
			text += fmt.Sprintf("🔴 优先级：%s\n", html.EscapeString(priority))
		}
		if comment, ok := dataMap["comment"].(string); ok && comment != "" {
			if len(comment) > 200 {
				comment = comment[:200] + "..."
			}
			text += fmt.Sprintf("💬 评论：%s\n", html.EscapeString(comment))
		}

		if issueKey != "" {
			replyMarkup = &telegramInlineKeyboard{
				InlineKeyboard: [][]telegramInlineButton{
					{
						{
							Text: "📋 查看工单",
							URL:  fmt.Sprintf("%s/issues/%s", siteURL, issueKey),
						},
					},
				},
			}
		}

	case len(event) > 6 && event[:6] == "alert.":
		alertName, _ := dataMap["alert_name"].(string)
		severity, _ := dataMap["severity"].(string)
		alertStatus, _ := dataMap["status"].(string)
		issueKey, _ := dataMap["issue_key"].(string)

		text = fmt.Sprintf("%s\n\n", title)
		text += fmt.Sprintf("<b>%s</b>\n\n", html.EscapeString(alertName))
		if severity != "" {
			text += fmt.Sprintf("⚠️ 级别：%s\n", html.EscapeString(severity))
		}
		if alertStatus != "" {
			text += fmt.Sprintf("📊 状态：%s\n", html.EscapeString(alertStatus))
		}
		if issueKey != "" {
			text += fmt.Sprintf("🎫 关联工单：%s\n", html.EscapeString(issueKey))
			replyMarkup = &telegramInlineKeyboard{
				InlineKeyboard: [][]telegramInlineButton{
					{
						{
							Text: "📋 查看工单",
							URL:  fmt.Sprintf("%s/issues/%s", siteURL, issueKey),
						},
					},
				},
			}
		}

	default:
		text = fmt.Sprintf("%s\n\n", title)
		jsonBytes, _ := json.MarshalIndent(dataMap, "", "  ")
		text += fmt.Sprintf("<pre>%s</pre>", html.EscapeString(string(jsonBytes)))
	}

	return text, replyMarkup
}

// doSend 执行 Telegram API 调用
func (d *DirectTelegramSender) doSend(ctx context.Context, text string, replyMarkup *telegramInlineKeyboard) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", d.botToken)

	reqBody := telegramSendMessageRequest{
		ChatID:    d.chatID,
		Text:      text,
		ParseMode: "HTML",
		LinkPreviewOptions: &linkPreviewOptions{
			IsDisabled: true,
		},
		ReplyMarkup: replyMarkup,
	}

	payloadBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化 Telegram 消息失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("创建 Telegram 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送 Telegram 通知失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && !result.OK {
		return fmt.Errorf("Telegram API 错误: %s", result.Description)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Telegram 返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// directGetEventTitle 获取事件标题
func directGetEventTitle(event string) string {
	switch event {
	case "issue.created":
		return "🎫 <b>工单创建</b>"
	case "issue.updated":
		return "✏️ <b>工单更新</b>"
	case "issue.transitioned":
		return "🔄 <b>工单状态变更</b>"
	case "issue.assigned":
		return "👤 <b>工单指派</b>"
	case "issue.commented":
		return "💬 <b>工单评论</b>"
	case "alert.firing":
		return "🔥 <b>告警触发</b>"
	case "alert.resolved":
		return "✅ <b>告警恢复</b>"
	case "alert.acked":
		return "👁️ <b>告警确认</b>"
	default:
		return "📢 <b>系统通知</b>"
	}
}

// toMap 将 interface{} 转换为 map[string]interface{}
func toMap(data interface{}) map[string]interface{} {
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return map[string]interface{}{}
	}
	var m map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &m); err != nil {
		return map[string]interface{}{}
	}
	return m
}
