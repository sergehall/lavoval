package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
	"sort"
	"strings"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/config"
)

type VerificationEmail struct {
	ToEmail     string
	ToName      string
	VerifyURL   string
	ProductName string
}

type PasswordResetEmail struct {
	ToEmail     string
	ToName      string
	ResetURL    string
	ProductName string
}

type PasswordChangedEmail struct {
	ToEmail     string
	ToName      string
	SignInURL   string
	ProductName string
}

type VerificationSender interface {
	SendVerificationEmail(context.Context, VerificationEmail) error
	SendPasswordResetEmail(context.Context, PasswordResetEmail) error
	SendPasswordChangedEmail(context.Context, PasswordChangedEmail) error
}

type SMTPProvider struct {
	cfg config.Config
}

func NewSMTPProvider(cfg config.Config) *SMTPProvider {
	return &SMTPProvider{cfg: cfg}
}

func (m *SMTPProvider) Send(ctx context.Context, msg Message) (SendResult, error) {
	if m.cfg.SMTPHost == "" || m.cfg.SMTPUsername == "" || m.cfg.SMTPPassword == "" || m.cfg.SMTPFromEmail == "" {
		return SendResult{}, permanentDeliveryError("not_configured", fmt.Errorf("email delivery is not configured"))
	}

	fromName := m.cfg.SMTPFromName
	if fromName == "" {
		fromName = m.cfg.AppName
	}

	message, err := buildMultipartMessage(fromName, m.cfg.SMTPFromEmail, msg.RecipientEmail, RenderedEmail{
		Subject:  msg.Subject,
		TextBody: msg.TextBody,
		HTMLBody: msg.HTMLBody,
	}, msg.Headers)
	if err != nil {
		return SendResult{}, permanentDeliveryError("mime_build_failed", fmt.Errorf("build %s email: %w", msg.MessageType, err))
	}

	timeout := m.cfg.SMTPDialTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", m.cfg.SMTPHost, m.cfg.SMTPPort)
	auth := smtp.PlainAuth("", m.cfg.SMTPUsername, m.cfg.SMTPPassword, m.cfg.SMTPHost)

	log.Printf("mailer: sending %s email to %s via %s (ssl=%v)", msg.MessageType, msg.RecipientEmail, addr, m.cfg.SMTPUseSSL)

	tlsCfg := &tls.Config{
		ServerName:         m.cfg.SMTPHost,
		InsecureSkipVerify: m.cfg.AppEnv == "development" && m.cfg.SMTPAllowInsecureAuth,
		MinVersion:         tls.VersionTLS12,
	}

	dialer := &net.Dialer{Timeout: timeout}

	var client *smtp.Client
	if m.cfg.SMTPUseSSL {
		tlsDialer := tls.Dialer{NetDialer: dialer, Config: tlsCfg}
		conn, err := tlsDialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			log.Printf("mailer: %s email ssl dial failed for %s: %v", msg.MessageType, msg.RecipientEmail, err)
			return SendResult{}, temporaryDeliveryError("smtp_ssl_dial_failed", fmt.Errorf("dial smtp ssl: %w", err))
		}
		defer conn.Close()

		client, err = smtp.NewClient(conn, m.cfg.SMTPHost)
		if err != nil {
			log.Printf("mailer: %s email client init failed for %s: %v", msg.MessageType, msg.RecipientEmail, err)
			return SendResult{}, temporaryDeliveryError("smtp_client_init_failed", fmt.Errorf("smtp new client: %w", err))
		}
	} else {
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			log.Printf("mailer: %s email dial failed for %s: %v", msg.MessageType, msg.RecipientEmail, err)
			return SendResult{}, temporaryDeliveryError("smtp_dial_failed", fmt.Errorf("dial smtp: %w", err))
		}
		defer conn.Close()

		if deadline, ok := ctx.Deadline(); ok {
			if err := conn.SetDeadline(deadline); err != nil {
				log.Printf("mailer: could not set smtp deadline for %s: %v", msg.RecipientEmail, err)
			}
		}

		client, err = smtp.NewClient(conn, m.cfg.SMTPHost)
		if err != nil {
			log.Printf("mailer: %s email client init failed for %s: %v", msg.MessageType, msg.RecipientEmail, err)
			return SendResult{}, temporaryDeliveryError("smtp_client_init_failed", fmt.Errorf("smtp new client: %w", err))
		}
	}
	defer func() {
		if quitErr := client.Quit(); quitErr != nil {
			log.Printf("mailer: smtp quit warning for %s email to %s: %v", msg.MessageType, msg.RecipientEmail, quitErr)
		}
	}()

	if !m.cfg.SMTPUseSSL && m.cfg.SMTPRequireTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			err := fmt.Errorf("smtp server does not advertise STARTTLS")
			log.Printf("mailer: %s email starttls unavailable for %s", msg.MessageType, msg.RecipientEmail)
			return SendResult{}, temporaryDeliveryError("smtp_starttls_unavailable", err)
		}
		if err := client.StartTLS(tlsCfg); err != nil {
			log.Printf("mailer: %s email starttls failed for %s: %v", msg.MessageType, msg.RecipientEmail, err)
			return SendResult{}, temporaryDeliveryError("smtp_starttls_failed", fmt.Errorf("smtp starttls: %w", err))
		}
	}

	if ok, _ := client.Extension("AUTH"); ok {
		if err := client.Auth(auth); err != nil {
			log.Printf("mailer: %s email auth failed for %s: %v", msg.MessageType, msg.RecipientEmail, err)
			return SendResult{}, permanentDeliveryError("smtp_auth_failed", fmt.Errorf("smtp auth: %w", err))
		}
	}

	if err := client.Mail(m.cfg.SMTPFromEmail); err != nil {
		log.Printf("mailer: %s email MAIL FROM failed for %s: %v", msg.MessageType, msg.RecipientEmail, err)
		return SendResult{}, classifySMTPCommandError("smtp_mail_from_failed", err, false)
	}
	if err := client.Rcpt(msg.RecipientEmail); err != nil {
		log.Printf("mailer: %s email RCPT TO failed for %s: %v", msg.MessageType, msg.RecipientEmail, err)
		return SendResult{}, classifySMTPCommandError("smtp_rcpt_failed", err, true)
	}

	writer, err := client.Data()
	if err != nil {
		log.Printf("mailer: %s email DATA failed for %s: %v", msg.MessageType, msg.RecipientEmail, err)
		return SendResult{}, classifySMTPCommandError("smtp_data_failed", err, false)
	}
	if _, err := writer.Write(message); err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			log.Printf("mailer: %s email writer close warning for %s after write failure: %v", msg.MessageType, msg.RecipientEmail, closeErr)
		}
		log.Printf("mailer: %s email write failed for %s: %v", msg.MessageType, msg.RecipientEmail, err)
		return SendResult{}, temporaryDeliveryError("smtp_write_failed", fmt.Errorf("smtp write message: %w", err))
	}
	if err := writer.Close(); err != nil {
		log.Printf("mailer: %s email close failed for %s: %v", msg.MessageType, msg.RecipientEmail, err)
		return SendResult{}, temporaryDeliveryError("smtp_writer_close_failed", fmt.Errorf("smtp close writer: %w", err))
	}

	log.Printf("mailer: sent %s email to %s", msg.MessageType, msg.RecipientEmail)
	return SendResult{Provider: "smtp", ProviderMessageID: msg.ID}, nil
}

func classifySMTPCommandError(defaultCode string, err error, fourXXPermanent bool) error {
	var smtpErr *textproto.Error
	if errors.As(err, &smtpErr) {
		switch {
		case smtpErr.Code >= 500:
			return permanentDeliveryError(fmt.Sprintf("smtp_%d", smtpErr.Code), err)
		case smtpErr.Code >= 400:
			if fourXXPermanent {
				return permanentDeliveryError(fmt.Sprintf("smtp_%d", smtpErr.Code), err)
			}
			return temporaryDeliveryError(fmt.Sprintf("smtp_%d", smtpErr.Code), err)
		}
	}

	return temporaryDeliveryError(defaultCode, err)
}

func buildMultipartMessage(fromName string, fromEmail string, toEmail string, email RenderedEmail, extraHeaders textproto.MIMEHeader) ([]byte, error) {
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

	headerLines := []string{
		fmt.Sprintf("From: %s <%s>", fromName, fromEmail),
		fmt.Sprintf("To: %s", toEmail),
		fmt.Sprintf("Subject: %s", email.Subject),
		"MIME-Version: 1.0",
		fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q", writer.Boundary()),
	}

	if len(extraHeaders) > 0 {
		keys := make([]string, 0, len(extraHeaders))
		for key := range extraHeaders {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			for _, value := range extraHeaders[key] {
				headerLines = append(headerLines, fmt.Sprintf("%s: %s", key, value))
			}
		}
	}

	headers := strings.Join(append(headerLines, ""), "\r\n")
	return append([]byte(headers), body.Bytes()...), nil
}

var _ Provider = (*SMTPProvider)(nil)
