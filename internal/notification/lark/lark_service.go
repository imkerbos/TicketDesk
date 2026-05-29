// Package lark 提供飞书（Lark）通知服务
package lark

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	configService "github.com/kerbos/ticketdesk/internal/system-config/service"
	"github.com/kerbos/ticketdesk/pkg/logger"
)

// LarkService 飞书通知服务接口
type LarkService interface {
	// SendNotification 发送飞书通知
	SendNotification(ctx context.Context, event string, data interface{}) error
	// SendTestMessage 发送测试消息
	SendTestMessage(ctx context.Context) error
	// IsEnabled 检查飞书通知是否启用
	IsEnabled(ctx context.Context) bool
}

// larkService 飞书通知服务实现
type larkService struct {
	configSvc  configService.ConfigService
	httpClient *http.Client
}

// NewLarkService 创建飞书通知服务实例
func NewLarkService(configSvc configService.ConfigService) LarkService {
	return &larkService{
		configSvc: configSvc,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsEnabled 检查飞书通知是否启用
func (s *larkService) IsEnabled(ctx context.Context) bool {
	enabled, err := s.configSvc.GetConfigValue(ctx, configService.KeyLarkEnabled)
	if err != nil {
		return false
	}
	return enabled == "true"
}

// SendNotification 发送飞书通知
func (s *larkService) SendNotification(ctx context.Context, event string, data interface{}) error {
	if !s.IsEnabled(ctx) {
		logger.Debug("lark notification is disabled, skip sending")
		return nil
	}

	webhookURL, err := s.configSvc.GetConfigValue(ctx, configService.KeyLarkWebhookURL)
	if err != nil || webhookURL == "" {
		return fmt.Errorf("飞书 Webhook URL 未配置")
	}

	// 构建卡片消息
	card := s.buildCard(event, data)

	// 构建请求体
	body := map[string]interface{}{
		"msg_type": "interactive",
		"card":     card,
	}

	// 添加签名（如果配置了密钥）
	secret, _ := s.configSvc.GetConfigValue(ctx, configService.KeyLarkSecret)
	if secret != "" {
		timestamp := time.Now().Unix()
		sign, signErr := s.generateSign(secret, timestamp)
		if signErr != nil {
			return fmt.Errorf("生成飞书签名失败: %w", signErr)
		}
		body["timestamp"] = fmt.Sprintf("%d", timestamp)
		body["sign"] = sign
	}

	return s.doSend(ctx, webhookURL, body)
}

// SendTestMessage 发送测试消息
func (s *larkService) SendTestMessage(ctx context.Context) error {
	webhookURL, err := s.configSvc.GetConfigValue(ctx, configService.KeyLarkWebhookURL)
	if err != nil || webhookURL == "" {
		return fmt.Errorf("飞书 Webhook URL 未配置")
	}

	card := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"title": map[string]interface{}{
				"tag":     "plain_text",
				"content": "🔔 TicketDesk 通知测试",
			},
			"template": "blue",
		},
		"elements": []interface{}{
			map[string]interface{}{
				"tag": "div",
				"text": map[string]interface{}{
					"tag":     "lark_md",
					"content": "✅ 恭喜！飞书通知配置成功。\n\n此消息由 TicketDesk 系统发送，用于验证飞书通知功能是否正常工作。",
				},
			},
			map[string]interface{}{
				"tag": "hr",
			},
			map[string]interface{}{
				"tag": "note",
				"elements": []interface{}{
					map[string]interface{}{
						"tag":     "plain_text",
						"content": fmt.Sprintf("来自 TicketDesk · %s", time.Now().Format("2006-01-02 15:04:05")),
					},
				},
			},
		},
	}

	body := map[string]interface{}{
		"msg_type": "interactive",
		"card":     card,
	}

	// 添加签名
	secret, _ := s.configSvc.GetConfigValue(ctx, configService.KeyLarkSecret)
	if secret != "" {
		timestamp := time.Now().Unix()
		sign, signErr := s.generateSign(secret, timestamp)
		if signErr != nil {
			return fmt.Errorf("生成飞书签名失败: %w", signErr)
		}
		body["timestamp"] = fmt.Sprintf("%d", timestamp)
		body["sign"] = sign
	}

	return s.doSend(ctx, webhookURL, body)
}

// buildCard 根据事件类型构建飞书卡片消息
func (s *larkService) buildCard(event string, data interface{}) map[string]interface{} {
	// 获取站点 URL
	ctx := context.Background()
	siteURL, _ := s.configSvc.GetConfigValue(ctx, configService.KeyGeneralSiteURL)
	if siteURL == "" {
		siteURL = "https://ticketdesk.example.com"
	}
	siteURL = strings.TrimRight(siteURL, "/")

	// 解析事件数据
	dataMap := toMap(data)

	// 根据事件类型确定标题和颜色
	title, template := s.getEventMeta(event)

	// 告警来源的工单创建，使用独立标题和颜色
	if event == "issue.created" {
		if source, _ := dataMap["source"].(string); source == "alert" {
			title = "🚨 告警建单"
			template = "red"
		}
	}

	// 构建内容字段
	contentLines := s.buildContentLines(event, dataMap)

	elements := []interface{}{
		map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": contentLines,
			},
		},
	}

	// 如果有工单链接，添加按钮
	if issueKey, ok := dataMap["issue_key"].(string); ok && issueKey != "" {
		elements = append(elements,
			map[string]interface{}{
				"tag": "hr",
			},
			map[string]interface{}{
				"tag": "action",
				"actions": []interface{}{
					map[string]interface{}{
						"tag": "button",
						"text": map[string]interface{}{
							"tag":     "plain_text",
							"content": "查看工单",
						},
						"url":  fmt.Sprintf("%s/issues/%s", siteURL, issueKey),
						"type": "primary",
					},
				},
			},
		)
	}

	// 添加底部备注
	elements = append(elements,
		map[string]interface{}{
			"tag": "note",
			"elements": []interface{}{
				map[string]interface{}{
					"tag":     "plain_text",
					"content": fmt.Sprintf("TicketDesk · %s", time.Now().Format("2006-01-02 15:04:05")),
				},
			},
		},
	)

	return map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"title": map[string]interface{}{
				"tag":     "plain_text",
				"content": title,
			},
			"template": template,
		},
		"elements": elements,
	}
}

// getEventMeta 获取事件的标题和卡片颜色模板
func (s *larkService) getEventMeta(event string) (title, template string) {
	switch event {
	case "issue.created":
		return "🎫 工单创建", "blue"
	case "issue.updated":
		return "✏️ 工单更新", "wathet"
	case "issue.transitioned":
		return "🔄 工单流转", "turquoise"
	case "issue.assigned":
		return "👤 工单指派", "indigo"
	case "issue.commented":
		return "💬 工单评论", "violet"
	case "alert.firing":
		return "🔥 告警触发", "red"
	case "alert.resolved":
		return "✅ 告警恢复", "green"
	case "alert.merged":
		return "🔗 告警合并", "orange"
	case "alert.acked":
		return "👁️ 告警确认", "orange"
	default:
		return "📢 系统通知", "blue"
	}
}

// buildContentLines 根据事件类型构建消息内容
func (s *larkService) buildContentLines(event string, data map[string]interface{}) string {
	return sharedBuildContentLines(event, data)
}

// sharedBuildContentLines 统一的消息内容构建逻辑
func sharedBuildContentLines(event string, data map[string]interface{}) string {
	var content string

	// 告警来源的工单创建，使用独立模板
	if event == "issue.created" {
		if source, _ := data["source"].(string); source == "alert" {
			return buildAlertIssueContent(data)
		}
	}

	switch {
	// 工单状态变更：专用模板，显示新旧状态对比
	case event == "issue.transitioned":
		issueKey, _ := data["issue_key"].(string)
		issueTitle, _ := data["issue_title"].(string)
		projectName, _ := data["project_name"].(string)
		priority, _ := data["priority"].(string)
		assignee, _ := data["assignee"].(string)
		dueDate, _ := data["due_date"].(string)
		statusName := getStatusDisplayName(data, "status", "status_name")
		oldStatusName := getStatusDisplayName(data, "old_status", "old_status_name")

		// 优先使用工作流节点名称，fallback 到状态名
		nodeName, _ := data["node_name"].(string)
		oldNodeName, _ := data["old_node_name"].(string)
		displayNew := nodeName
		displayOld := oldNodeName
		if displayNew == "" {
			displayNew = statusName
		}
		if displayOld == "" {
			displayOld = oldStatusName
		}

		content = fmt.Sprintf("**%s** %s\n\n", issueKey, issueTitle)
		if displayOld != "" && displayNew != "" {
			content += fmt.Sprintf("📊 %s → **%s**\n", displayOld, displayNew)
		} else if displayNew != "" {
			content += fmt.Sprintf("📊 状态：**%s**\n", displayNew)
		}
		if projectName != "" {
			content += fmt.Sprintf("📁 项目：%s\n", projectName)
		}
		if priority != "" {
			content += fmt.Sprintf("🔴 优先级：%s\n", priority)
		}
		if assignee != "" {
			content += fmt.Sprintf("👤 处理人：%s\n", assignee)
		}
		if dueDate != "" {
			content += fmt.Sprintf("⏰ 截止时间：**%s**\n", dueDate)
		}

	// 工单指派：专用模板，显示操作人和指派人
	case event == "issue.assigned":
		issueKey, _ := data["issue_key"].(string)
		issueTitle, _ := data["issue_title"].(string)
		projectName, _ := data["project_name"].(string)
		priority, _ := data["priority"].(string)
		assignee, _ := data["assignee"].(string)
		operator, _ := data["operator"].(string)
		dueDate, _ := data["due_date"].(string)
		statusName := getStatusDisplayName(data, "status", "status_name")

		content = fmt.Sprintf("**%s** %s\n\n", issueKey, issueTitle)
		if projectName != "" {
			content += fmt.Sprintf("📁 项目：%s\n", projectName)
		}
		if operator != "" && assignee != "" {
			content += fmt.Sprintf("👤 %s 指派给 **%s**\n", operator, assignee)
		} else if assignee != "" {
			content += fmt.Sprintf("👤 指派给：**%s**\n", assignee)
		}
		if priority != "" {
			content += fmt.Sprintf("🔴 优先级：%s\n", priority)
		}
		if statusName != "" {
			content += fmt.Sprintf("📊 状态：%s\n", statusName)
		}
		if dueDate != "" {
			content += fmt.Sprintf("⏰ 截止时间：**%s**\n", dueDate)
		}

	// 告警合并：专用模板，显示合并详情
	case event == "alert.merged":
		issueKey, _ := data["issue_key"].(string)
		issueTitle, _ := data["issue_title"].(string)
		alertName, _ := data["alert_name"].(string)
		instance, _ := data["instance"].(string)
		// alert_count 可能是 int 或 float64（JSON 反序列化）
		var alertCount string
		switch v := data["alert_count"].(type) {
		case float64:
			alertCount = fmt.Sprintf("%d", int(v))
		case int:
			alertCount = fmt.Sprintf("%d", v)
		case int64:
			alertCount = fmt.Sprintf("%d", v)
		}

		content = fmt.Sprintf("**%s** %s\n\n", issueKey, issueTitle)
		if alertName != "" {
			content += fmt.Sprintf("⚠️ 告警：%s\n", alertName)
		}
		if instance != "" {
			content += fmt.Sprintf("📊 新增实例：**%s**\n", instance)
		}
		if alertCount != "" {
			content += fmt.Sprintf("🔢 当前实例数：**%s**\n", alertCount)
		}

	// 通用工单事件
	case len(event) > 6 && event[:6] == "issue.":
		issueKey, _ := data["issue_key"].(string)
		issueTitle, _ := data["issue_title"].(string)
		projectName, _ := data["project_name"].(string)
		priority, _ := data["priority"].(string)
		statusName := getStatusDisplayName(data, "status", "status_name")

		content = fmt.Sprintf("**%s** %s\n", issueKey, issueTitle)
		if projectName != "" {
			content += fmt.Sprintf("📁 项目：%s\n", projectName)
		}
		if statusName != "" {
			content += fmt.Sprintf("📊 状态：%s\n", statusName)
		}
		if priority != "" {
			content += fmt.Sprintf("🔴 优先级：%s\n", priority)
		}
		if comment, ok := data["comment"].(string); ok && comment != "" {
			if len(comment) > 200 {
				comment = comment[:200] + "..."
			}
			content += fmt.Sprintf("💬 评论：%s\n", comment)
		}

	case len(event) > 6 && event[:6] == "alert.":
		alertName, _ := data["alert_name"].(string)
		severity, _ := data["severity"].(string)
		alertStatus, _ := data["status"].(string)
		issueKey, _ := data["issue_key"].(string)

		content = fmt.Sprintf("**%s**\n", alertName)
		if severity != "" {
			content += fmt.Sprintf("⚠️ 级别：%s\n", severity)
		}
		if alertStatus != "" {
			content += fmt.Sprintf("📊 状态：%s\n", alertStatus)
		}
		if issueKey != "" {
			content += fmt.Sprintf("🎫 关联工单：%s\n", issueKey)
		}

	default:
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		content = string(jsonBytes)
	}

	return content
}

// buildAlertIssueContent 构建告警建单的消息内容
func buildAlertIssueContent(data map[string]interface{}) string {
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

	// 优先级 emoji
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

	content := fmt.Sprintf("**%s**\n", issueKey)
	content += fmt.Sprintf("%s\n\n", alertName)
	content += fmt.Sprintf("%s 优先级：**%s**　　📊 状态：**%s**\n", priEmoji, priority, status)
	if projectName != "" || assignee != "" {
		content += fmt.Sprintf("📁 项目：**%s**", projectName)
		if assignee != "" {
			content += fmt.Sprintf("　　👤 指派：**%s**", assignee)
		}
		content += "\n"
	}
	if severity != "" {
		content += fmt.Sprintf("⚠️ 级别：**%s**\n", severity)
	}
	if alertTime != "" {
		content += fmt.Sprintf("⏰ 告警时间：%s\n", alertTime)
	}

	return content
}

// generateSign 生成飞书签名
func (s *larkService) generateSign(secret string, timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write([]byte{})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// doSend 执行 HTTP 发送
func (s *larkService) doSend(ctx context.Context, webhookURL string, body map[string]interface{}) error {
	payloadBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化飞书消息失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("创建飞书请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error("failed to send lark notification", zap.Error(err))
		return fmt.Errorf("发送飞书通知失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// 飞书返回 200 但可能 code != 0
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && result.Code != 0 {
		logger.Error("lark webhook returned error",
			zap.Int("code", result.Code),
			zap.String("msg", result.Msg),
		)
		return fmt.Errorf("飞书返回错误: code=%d, msg=%s", result.Code, result.Msg)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("飞书返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	logger.Info("lark notification sent successfully", zap.String("webhook_url", webhookURL))
	return nil
}

// ============ Direct Sender（项目级通知渠道使用，不依赖 ConfigService）============

// DirectLarkSender 直接飞书发送器（接受凭据参数，不依赖系统配置）
type DirectLarkSender struct {
	webhookURL string
	secret     string
	siteURL    string
	httpClient *http.Client
}

// NewDirectLarkSender 创建直接飞书发送器
func NewDirectLarkSender(webhookURL, secret string, siteURL ...string) *DirectLarkSender {
	url := ""
	if len(siteURL) > 0 {
		url = siteURL[0]
	}
	return &DirectLarkSender{
		webhookURL: webhookURL,
		secret:     secret,
		siteURL:    url,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendNotification 发送飞书通知
func (d *DirectLarkSender) SendNotification(ctx context.Context, event string, data interface{}) error {
	dataMap := toMap(data)
	title, template := directGetEventMeta(event)

	// 告警来源的工单创建，使用独立标题和颜色
	if event == "issue.created" {
		if source, _ := dataMap["source"].(string); source == "alert" {
			title = "🚨 告警建单"
			template = "red"
		}
	}

	siteURL := d.siteURL
	if siteURL == "" {
		siteURL = "https://ticketdesk.example.com"
	}
	siteURL = strings.TrimRight(siteURL, "/")

	// 每日日报有独立模板：标题 + 多组列表 + 项目看板按钮
	if event == "issue.daily_digest" {
		return d.sendDigest(ctx, dataMap, siteURL)
	}

	contentLines := sharedBuildContentLines(event, dataMap)
	contentLines = appendLarkMentions(contentLines, dataMap)

	elements := []interface{}{
		map[string]interface{}{
			"tag": "div",
			"text": map[string]interface{}{
				"tag":     "lark_md",
				"content": contentLines,
			},
		},
	}

	// 如果有工单链接，添加按钮
	if issueKey, ok := dataMap["issue_key"].(string); ok && issueKey != "" {
		elements = append(elements,
			map[string]interface{}{
				"tag": "hr",
			},
			map[string]interface{}{
				"tag": "action",
				"actions": []interface{}{
					map[string]interface{}{
						"tag": "button",
						"text": map[string]interface{}{
							"tag":     "plain_text",
							"content": "查看工单",
						},
						"url":  fmt.Sprintf("%s/issues/%s", siteURL, issueKey),
						"type": "primary",
					},
				},
			},
		)
	}

	elements = append(elements,
		map[string]interface{}{
			"tag": "note",
			"elements": []interface{}{
				map[string]interface{}{
					"tag":     "plain_text",
					"content": fmt.Sprintf("TicketDesk · %s", time.Now().Format("2006-01-02 15:04:05")),
				},
			},
		},
	)

	card := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"title": map[string]interface{}{
				"tag":     "plain_text",
				"content": title,
			},
			"template": template,
		},
		"elements": elements,
	}

	body := map[string]interface{}{
		"msg_type": "interactive",
		"card":     card,
	}

	if d.secret != "" {
		timestamp := time.Now().Unix()
		sign, signErr := directGenerateSign(d.secret, timestamp)
		if signErr != nil {
			return fmt.Errorf("生成飞书签名失败: %w", signErr)
		}
		body["timestamp"] = fmt.Sprintf("%d", timestamp)
		body["sign"] = sign
	}

	return d.doSend(ctx, body)
}

// sendDigest 发送每日日报卡片（按 assignee 分组 + 每组前 @ 对应人）
// 期望 data 含：project_key, project_name, total, groups, mention_all
func (d *DirectLarkSender) sendDigest(ctx context.Context, dataMap map[string]interface{}, siteURL string) error {
	projectKey, _ := dataMap["project_key"].(string)
	projectName, _ := dataMap["project_name"].(string)
	total := toInt(dataMap["total"])

	header := fmt.Sprintf("📋 **[%s] %s** 未完结工单共 **%d** 条", projectKey, projectName, total)
	if all, _ := dataMap["mention_all"].(bool); all {
		header = `<at id="all"></at>` + "\n" + header
	}

	elements := []interface{}{
		map[string]interface{}{
			"tag":  "div",
			"text": map[string]interface{}{"tag": "lark_md", "content": header},
		},
	}

	groups := extractGroupList(dataMap)
	for _, g := range groups {
		var sb strings.Builder
		// 分组标题：纯文本指派人名称（每行末尾会单独 @ 对应人）
		assigneeName, _ := g["assignee_name"].(string)
		sb.WriteString(fmt.Sprintf("**%s**\n", assigneeName))

		// 每条 issue 单独渲染，末尾追加 @ 对应指派人（缺 mention 时仅纯文本）
		items := extractItemList(g["items"])
		for _, it := range items {
			key, _ := it["issue_key"].(string)
			title, _ := it["title"].(string)
			typ, _ := it["issue_type"].(string)
			priority, _ := it["priority"].(string)
			status, _ := it["status"].(string)
			nodeName, _ := it["node_name"].(string)
			display := status
			if nodeName != "" {
				display = nodeName
			}
			meta := joinNonEmpty([]string{typ, priority, display}, " | ")
			line := fmt.Sprintf("• [%s](%s/issues/%s) [%s] %s",
				key, siteURL, key, meta, title)
			if m, ok := it["mention"].(map[string]interface{}); ok {
				line += " " + larkMentionFromMap(m)
			}
			sb.WriteString(line + "\n")
		}

		elements = append(elements,
			map[string]interface{}{"tag": "hr"},
			map[string]interface{}{
				"tag":  "div",
				"text": map[string]interface{}{"tag": "lark_md", "content": sb.String()},
			},
		)
	}

	// 项目看板按钮
	elements = append(elements,
		map[string]interface{}{"tag": "hr"},
		map[string]interface{}{
			"tag": "action",
			"actions": []interface{}{
				map[string]interface{}{
					"tag":  "button",
					"text": map[string]interface{}{"tag": "plain_text", "content": "查看项目"},
					"url":  fmt.Sprintf("%s/projects/%s", siteURL, projectKey),
					"type": "primary",
				},
			},
		},
		map[string]interface{}{
			"tag": "note",
			"elements": []interface{}{
				map[string]interface{}{
					"tag":     "plain_text",
					"content": fmt.Sprintf("TicketDesk · %s", time.Now().Format("2006-01-02 15:04:05")),
				},
			},
		},
	)

	card := map[string]interface{}{
		"config": map[string]interface{}{"wide_screen_mode": true},
		"header": map[string]interface{}{
			"title":    map[string]interface{}{"tag": "plain_text", "content": "📋 每日工单日报"},
			"template": "blue",
		},
		"elements": elements,
	}

	body := map[string]interface{}{"msg_type": "interactive", "card": card}
	if d.secret != "" {
		timestamp := time.Now().Unix()
		sign, signErr := directGenerateSign(d.secret, timestamp)
		if signErr != nil {
			return fmt.Errorf("生成飞书签名失败: %w", signErr)
		}
		body["timestamp"] = fmt.Sprintf("%d", timestamp)
		body["sign"] = sign
	}
	return d.doSend(ctx, body)
}

// larkMentionFromMap 把单个 mention map 渲染为 <at> 标签（飞书 lark_md 用 id 而非 user_id）
// 优先 lark_open_id -> email -> 纯文本
func larkMentionFromMap(m map[string]interface{}) string {
	name, _ := m["display_name"].(string)
	openID, _ := m["lark_open_id"].(string)
	email, _ := m["email"].(string)
	switch {
	case openID != "":
		return fmt.Sprintf("<at id=%q>%s</at>", openID, name)
	case email != "":
		return fmt.Sprintf("<at email=%q>%s</at>", email, name)
	case name != "":
		return "@" + name
	}
	return ""
}

// extractGroupList 兼容 []map[string]any / []any
func extractGroupList(data map[string]interface{}) []map[string]interface{} {
	raw, ok := data["groups"]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []map[string]interface{}:
		return v
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		return out
	}
	return nil
}

// extractItemList 兼容 []map[string]any / []any
func extractItemList(raw interface{}) []map[string]interface{} {
	switch v := raw.(type) {
	case []map[string]interface{}:
		return v
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		return out
	}
	return nil
}

// toInt 把 any 数字字段转 int
func toInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}

// joinNonEmpty 拼接非空字符串
func joinNonEmpty(parts []string, sep string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, sep)
}

// SendTestMessage 发送测试消息（使用与真实通知相同的模板）
func (d *DirectLarkSender) SendTestMessage(ctx context.Context) error {
	testData := map[string]interface{}{
		"issue_key":    "TEST-1",
		"issue_title":  "这是一条测试通知，用于验证飞书通知渠道是否配置正确",
		"project_name": "测试项目",
		"status":       "待处理",
		"priority":     "P2",
	}
	return d.SendNotification(ctx, "issue.created", testData)
}

// doSend 执行 HTTP 发送
func (d *DirectLarkSender) doSend(ctx context.Context, body map[string]interface{}) error {
	payloadBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化飞书消息失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.webhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("创建飞书请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送飞书通知失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && result.Code != 0 {
		return fmt.Errorf("飞书返回错误: code=%d, msg=%s", result.Code, result.Msg)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("飞书返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// ============ Direct 辅助函数 ============

func directGetEventMeta(event string) (title, template string) {
	switch event {
	case "issue.created":
		return "🎫 工单创建", "blue"
	case "issue.updated":
		return "✏️ 工单更新", "wathet"
	case "issue.transitioned":
		return "🔄 工单流转", "turquoise"
	case "issue.assigned":
		return "👤 工单指派", "indigo"
	case "issue.commented":
		return "💬 工单评论", "violet"
	case "alert.firing":
		return "🔥 告警触发", "red"
	case "alert.resolved":
		return "✅ 告警恢复", "green"
	case "alert.merged":
		return "🔗 告警合并", "orange"
	case "alert.acked":
		return "👁️ 告警确认", "orange"
	default:
		return "📢 系统通知", "blue"
	}
}

func directGenerateSign(secret string, timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write([]byte{})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// statusDisplayNames 状态中文显示名映射
var statusDisplayNames = map[string]string{
	"open":           "待处理",
	"in_progress":    "进行中",
	"resolved":       "已解决",
	"closed":         "已关闭",
	"reviewing":      "待确认",
	"pending_review": "待确认",
	"merged":         "已合并",
}

// getStatusDisplayName 获取状态的中文显示名，优先使用 nameKey，fallback 到 statusKey 的映射
func getStatusDisplayName(data map[string]interface{}, statusKey, nameKey string) string {
	// 优先使用已有的中文名
	if name, _ := data[nameKey].(string); name != "" {
		return name
	}
	// fallback：通过英文状态映射
	if status, _ := data[statusKey].(string); status != "" {
		if name, ok := statusDisplayNames[status]; ok {
			return name
		}
		return status
	}
	return ""
}

// appendLarkMentions 在飞书卡片内容尾部追加 @ 提及段落
// 期望 data["mentions"] 为 []any 或 []map[string]any，每项含 display_name / lark_open_id
// 缺 lark_open_id 时退化为纯文本 @display_name
func appendLarkMentions(content string, data map[string]interface{}) string {
	mentions := extractMentionList(data)
	if len(mentions) == 0 {
		return content
	}

	var parts []string
	// @ 全员（飞书 lark_md 卡片必须用 id="all"，user_id="all" 不触发）
	if all, _ := data["mention_all"].(bool); all {
		parts = append(parts, `<at id="all"></at>`)
	}
	for _, m := range mentions {
		name, _ := m["display_name"].(string)
		openID, _ := m["lark_open_id"].(string)
		email, _ := m["email"].(string)
		switch {
		case openID != "":
			// 注意：lark_md 卡片必须用 id 属性，user_id 不触发提醒
			parts = append(parts, fmt.Sprintf("<at id=%q>%s</at>", openID, name))
		case email != "":
			// 飞书自定义机器人 webhook 卡片支持 <at email>，邮箱与飞书账号一致时可触发提醒
			parts = append(parts, fmt.Sprintf("<at email=%q>%s</at>", email, name))
		case name != "":
			parts = append(parts, fmt.Sprintf("@%s", name))
		}
	}
	if len(parts) == 0 {
		return content
	}
	return content + "\n\n" + strings.Join(parts, " ")
}

// extractMentionList 兼容 []any 与 []map[string]any 两种形态
func extractMentionList(data map[string]interface{}) []map[string]interface{} {
	raw, ok := data["mentions"]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []map[string]interface{}:
		return v
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		return out
	}
	return nil
}

// toMap 将 interface{} 转换为 map[string]interface{}
func toMap(data interface{}) map[string]interface{} {
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}
	// 通过 JSON 序列化/反序列化转换
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
