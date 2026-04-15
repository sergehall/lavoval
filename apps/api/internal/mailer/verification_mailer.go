package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strings"

	"github.com/sergehall/lavoval/apps/api/internal/config"
)

type VerificationEmail struct {
	ToEmail     string
	ToName      string
	VerifyURL   string
	ProductName string
}

type VerificationSender interface {
	SendVerificationEmail(context.Context, VerificationEmail) error
	SendPasswordResetEmail(context.Context, PasswordResetEmail) error
	SendPasswordChangedEmail(context.Context, PasswordChangedEmail) error
}

type SMTPVerificationMailer struct {
	cfg config.Config
}

func NewSMTPVerificationMailer(cfg config.Config) *SMTPVerificationMailer {
	return &SMTPVerificationMailer{cfg: cfg}
}

func (m *SMTPVerificationMailer) SendVerificationEmail(_ context.Context, email VerificationEmail) (err error) {
	if m.cfg.SMTPHost == "" || m.cfg.SMTPUsername == "" || m.cfg.SMTPPassword == "" || m.cfg.SMTPFromEmail == "" {
		return fmt.Errorf("email delivery is not configured")
	}

	fromName := m.cfg.SMTPFromName
	if fromName == "" {
		fromName = email.ProductName
	}

	rendered, err := renderVerificationEmail(email, m.cfg.AppURL)
	if err != nil {
		return fmt.Errorf("render verification email: %w", err)
	}

	message, err := buildMultipartMessage(fromName, m.cfg.SMTPFromEmail, email.ToEmail, rendered)
	if err != nil {
		return fmt.Errorf("build verification email: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", m.cfg.SMTPHost, m.cfg.SMTPPort)
	auth := smtp.PlainAuth("", m.cfg.SMTPUsername, m.cfg.SMTPPassword, m.cfg.SMTPHost)

	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("dial smtp: %w", err)
	}
	defer func() {
		if quitErr := client.Quit(); quitErr != nil && err == nil {
			err = fmt.Errorf("smtp quit: %w", quitErr)
		}
	}()

	if m.cfg.SMTPRequireTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("smtp server does not advertise STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{
			ServerName:         m.cfg.SMTPHost,
			InsecureSkipVerify: m.cfg.AppEnv == "development" && m.cfg.SMTPAllowInsecureAuth,
			MinVersion:         tls.VersionTLS12,
		}); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if ok, _ := client.Extension("AUTH"); ok {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(m.cfg.SMTPFromEmail); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(email.ToEmail); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		if closeErr := writer.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("smtp close writer after write failure: %w", closeErr)
		}
		return fmt.Errorf("smtp write message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp close writer: %w", err)
	}

	return nil
}

func (m *SMTPVerificationMailer) SendPasswordResetEmail(_ context.Context, email PasswordResetEmail) (err error) {
	if m.cfg.SMTPHost == "" || m.cfg.SMTPUsername == "" || m.cfg.SMTPPassword == "" || m.cfg.SMTPFromEmail == "" {
		return fmt.Errorf("email delivery is not configured")
	}

	fromName := m.cfg.SMTPFromName
	if fromName == "" {
		fromName = email.ProductName
	}

	rendered, err := renderPasswordResetEmail(email, m.cfg.AppURL)
	if err != nil {
		return fmt.Errorf("render password reset email: %w", err)
	}

	message, err := buildMultipartMessage(fromName, m.cfg.SMTPFromEmail, email.ToEmail, rendered)
	if err != nil {
		return fmt.Errorf("build password reset email: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", m.cfg.SMTPHost, m.cfg.SMTPPort)
	auth := smtp.PlainAuth("", m.cfg.SMTPUsername, m.cfg.SMTPPassword, m.cfg.SMTPHost)

	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("dial smtp: %w", err)
	}
	defer func() {
		if quitErr := client.Quit(); quitErr != nil && err == nil {
			err = fmt.Errorf("smtp quit: %w", quitErr)
		}
	}()

	if m.cfg.SMTPRequireTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("smtp server does not advertise STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{
			ServerName:         m.cfg.SMTPHost,
			InsecureSkipVerify: m.cfg.AppEnv == "development" && m.cfg.SMTPAllowInsecureAuth,
			MinVersion:         tls.VersionTLS12,
		}); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if ok, _ := client.Extension("AUTH"); ok {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(m.cfg.SMTPFromEmail); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(email.ToEmail); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		if closeErr := writer.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("smtp close writer after write failure: %w", closeErr)
		}
		return fmt.Errorf("smtp write message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp close writer: %w", err)
	}

	return nil
}

func (m *SMTPVerificationMailer) SendPasswordChangedEmail(_ context.Context, email PasswordChangedEmail) (err error) {
	if m.cfg.SMTPHost == "" || m.cfg.SMTPUsername == "" || m.cfg.SMTPPassword == "" || m.cfg.SMTPFromEmail == "" {
		return fmt.Errorf("email delivery is not configured")
	}

	fromName := m.cfg.SMTPFromName
	if fromName == "" {
		fromName = email.ProductName
	}

	rendered, err := renderPasswordChangedEmail(email, m.cfg.AppURL)
	if err != nil {
		return fmt.Errorf("render password changed email: %w", err)
	}

	message, err := buildMultipartMessage(fromName, m.cfg.SMTPFromEmail, email.ToEmail, rendered)
	if err != nil {
		return fmt.Errorf("build password changed email: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", m.cfg.SMTPHost, m.cfg.SMTPPort)
	auth := smtp.PlainAuth("", m.cfg.SMTPUsername, m.cfg.SMTPPassword, m.cfg.SMTPHost)

	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("dial smtp: %w", err)
	}
	defer func() {
		if quitErr := client.Quit(); quitErr != nil && err == nil {
			err = fmt.Errorf("smtp quit: %w", quitErr)
		}
	}()

	if m.cfg.SMTPRequireTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("smtp server does not advertise STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{
			ServerName:         m.cfg.SMTPHost,
			InsecureSkipVerify: m.cfg.AppEnv == "development" && m.cfg.SMTPAllowInsecureAuth,
			MinVersion:         tls.VersionTLS12,
		}); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if ok, _ := client.Extension("AUTH"); ok {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(m.cfg.SMTPFromEmail); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(email.ToEmail); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		if closeErr := writer.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("smtp close writer after write failure: %w", closeErr)
		}
		return fmt.Errorf("smtp write message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp close writer: %w", err)
	}

	return nil
}

func buildMultipartMessage(fromName string, fromEmail string, toEmail string, email RenderedEmail) ([]byte, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	textHeader := textproto.MIMEHeader{}
	textHeader.Set("Content-Type", `text/plain; charset="UTF-8"`)
	textPart, err := writer.CreatePart(textHeader)
	if err != nil {
		return nil, fmt.Errorf("create text part: %w", err)
	}
	if _, err := textPart.Write([]byte(email.TextBody)); err != nil {
		return nil, fmt.Errorf("write text part: %w", err)
	}

	htmlHeader := textproto.MIMEHeader{}
	htmlHeader.Set("Content-Type", `text/html; charset="UTF-8"`)
	htmlPart, err := writer.CreatePart(htmlHeader)
	if err != nil {
		return nil, fmt.Errorf("create html part: %w", err)
	}
	if _, err := htmlPart.Write([]byte(email.HTMLBody)); err != nil {
		return nil, fmt.Errorf("write html part: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	headers := strings.Join([]string{
		fmt.Sprintf("From: %s <%s>", fromName, fromEmail),
		fmt.Sprintf("To: %s", toEmail),
		fmt.Sprintf("Subject: %s", email.Subject),
		"MIME-Version: 1.0",
		fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q", writer.Boundary()),
		"",
	}, "\r\n")

	return append([]byte(headers), body.Bytes()...), nil
}

func displayName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "there"
	}
	return trimmed
}

var _ VerificationSender = (*SMTPVerificationMailer)(nil)
