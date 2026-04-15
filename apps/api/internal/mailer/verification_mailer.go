package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
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

	subject := fmt.Sprintf("%s: confirm your email", email.ProductName)
	plainBody := strings.Join([]string{
		fmt.Sprintf("Hi %s,", displayName(email.ToName)),
		"",
		fmt.Sprintf("Confirm your %s email address to activate your account:", email.ProductName),
		email.VerifyURL,
		"",
		"If you did not create this account, you can ignore this email.",
	}, "\r\n")

	message := strings.Join([]string{
		fmt.Sprintf("From: %s <%s>", fromName, m.cfg.SMTPFromEmail),
		fmt.Sprintf("To: %s", email.ToEmail),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		plainBody,
	}, "\r\n")

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
	if _, err := writer.Write([]byte(message)); err != nil {
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

func displayName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "there"
	}
	return trimmed
}

var _ VerificationSender = (*SMTPVerificationMailer)(nil)
