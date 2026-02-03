// Package email 提供邮件发送服务
package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/kerbos/ticketdesk/internal/system-config/service"
	"github.com/kerbos/ticketdesk/pkg/logger"
	"go.uber.org/zap"
)

// EmailService 邮件服务接口
type EmailService interface {
	SendEmail(ctx context.Context, to []string, subject, body string) error
	SendHTMLEmail(ctx context.Context, to []string, subject, htmlBody string) error
	IsEnabled(ctx context.Context) bool
}

// emailService 邮件服务实现
type emailService struct {
	configService service.ConfigService
}

// NewEmailService 创建邮件服务实例
func NewEmailService(configService service.ConfigService) EmailService {
	return &emailService{
		configService: configService,
	}
}

// IsEnabled 检查邮件服务是否启用
func (s *emailService) IsEnabled(ctx context.Context) bool {
	enabled, err := s.configService.GetConfigValue(ctx, service.KeyEmailEnabled)
	if err != nil {
		return false
	}
	return enabled == "true"
}

// SendEmail 发送纯文本邮件
func (s *emailService) SendEmail(ctx context.Context, to []string, subject, body string) error {
	return s.send(ctx, to, subject, body, false)
}

// SendHTMLEmail 发送 HTML 邮件
func (s *emailService) SendHTMLEmail(ctx context.Context, to []string, subject, htmlBody string) error {
	return s.send(ctx, to, subject, htmlBody, true)
}

// send 发送邮件
func (s *emailService) send(ctx context.Context, to []string, subject, body string, isHTML bool) error {
	if !s.IsEnabled(ctx) {
		logger.Warn("email service is disabled, skip sending")
		return nil
	}

	// 获取配置
	config, err := s.configService.GetEmailConfig(ctx)
	if err != nil {
		return fmt.Errorf("获取邮件配置失败: %w", err)
	}

	if config.SMTPHost == "" {
		return fmt.Errorf("SMTP 服务器未配置")
	}

	// 获取密码（敏感配置需要单独获取）
	password, _ := s.configService.GetConfigValue(ctx, service.KeyEmailSMTPPassword)

	// 构建邮件内容
	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}

	fromHeader := config.FromAddress
	if config.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", config.FromName, config.FromAddress)
	}

	message := fmt.Sprintf("From: %s\r\n", fromHeader)
	message += fmt.Sprintf("To: %s\r\n", strings.Join(to, ", "))
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += fmt.Sprintf("Content-Type: %s; charset=UTF-8\r\n", contentType)
	message += "\r\n" + body

	// 连接 SMTP 服务器
	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	var auth smtp.Auth
	if config.SMTPUsername != "" && password != "" {
		auth = smtp.PlainAuth("", config.SMTPUsername, password, config.SMTPHost)
	}

	var sendErr error
	if config.UseTLS {
		sendErr = s.sendWithTLS(addr, auth, config.FromAddress, to, []byte(message), config.SMTPHost)
	} else {
		sendErr = smtp.SendMail(addr, auth, config.FromAddress, to, []byte(message))
	}

	if sendErr != nil {
		logger.Error("failed to send email",
			zap.Strings("to", to),
			zap.String("subject", subject),
			zap.Error(sendErr),
		)
		return fmt.Errorf("发送邮件失败: %w", sendErr)
	}

	logger.Info("email sent successfully",
		zap.Strings("to", to),
		zap.String("subject", subject),
	)

	return nil
}

// sendWithTLS 使用 TLS 发送邮件
func (s *emailService) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
	// 解析地址
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid address: %s", addr)
	}

	port, _ := strconv.Atoi(parts[1])

	// 如果端口是 465，使用隐式 TLS
	if port == 465 {
		return s.sendWithImplicitTLS(addr, auth, from, to, msg, host)
	}

	// 否则使用 STARTTLS
	return s.sendWithStartTLS(addr, auth, from, to, msg, host)
}

// sendWithStartTLS 使用 STARTTLS 发送邮件
func (s *emailService) sendWithStartTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	// STARTTLS
	tlsConfig := &tls.Config{
		ServerName: host,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return err
	}

	// 认证
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	// 发送邮件
	if err := client.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// sendWithImplicitTLS 使用隐式 TLS 发送邮件（端口 465）
func (s *emailService) sendWithImplicitTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
	tlsConfig := &tls.Config{
		ServerName: host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	// 认证
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	// 发送邮件
	if err := client.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}
