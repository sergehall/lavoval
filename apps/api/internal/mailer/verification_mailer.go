package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
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

type SMTPVerificationMailer struct {
	cfg config.Config
}

func NewSMTPVerificationMailer(cfg config.Config) *SMTPVerificationMailer {
	return &SMTPVerificationMailer{cfg: cfg}
}

func (m *SMTPVerificationMailer) SendVerificationEmail(ctx context.Context, email VerificationEmail) error {
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

	return m.sendRenderedEmail(ctx, "verification", fromName, email.ToEmail, rendered)
}

func (m *SMTPVerificationMailer) SendPasswordResetEmail(ctx context.Context, email PasswordResetEmail) error {
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

	return m.sendRenderedEmail(ctx, "password-reset", fromName, email.ToEmail, rendered)
}

func (m *SMTPVerificationMailer) SendPasswordChangedEmail(ctx context.Context, email PasswordChangedEmail) error {
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

	return m.sendRenderedEmail(ctx, "password-changed", fromName, email.ToEmail, rendered)
}

func (m *SMTPVerificationMailer) sendRenderedEmail(
	ctx context.Context,
	kind string,
	fromName string,
	toEmail string,
	email RenderedEmail,
) error {
	message, err := buildMultipartMessage(fromName, m.cfg.SMTPFromEmail, toEmail, email)
	if err != nil {
		return fmt.Errorf("build %s email: %w", kind, err)
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

	log.Printf("mailer: sending %s email to %s via %s (ssl=%v)", kind, toEmail, addr, m.cfg.SMTPUseSSL)

	tlsCfg := &tls.Config{
		ServerName:         m.cfg.SMTPHost,
		InsecureSkipVerify: m.cfg.AppEnv == "development" && m.cfg.SMTPAllowInsecureAuth,
		MinVersion:         tls.VersionTLS12,
	}

	dialer := &net.Dialer{Timeout: timeout}

	var client *smtp.Client
	if m.cfg.SMTPUseSSL {
		// Implicit TLS (port 465)
		tlsDialer := tls.Dialer{NetDialer: dialer, Config: tlsCfg}
		conn, err := tlsDialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			log.Printf("mailer: %s email ssl dial failed for %s: %v", kind, toEmail, err)
			return fmt.Errorf("dial smtp ssl: %w", err)
		}
		defer conn.Close()
		client, err = smtp.NewClient(conn, m.cfg.SMTPHost)
		if err != nil {
			log.Printf("mailer: %s email client init failed for %s: %v", kind, toEmail, err)
			return fmt.Errorf("smtp new client: %w", err)
		}
	} else {
		// STARTTLS (port 587)
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			log.Printf("mailer: %s email dial failed for %s: %v", kind, toEmail, err)
			return fmt.Errorf("dial smtp: %w", err)
		}
		defer conn.Close()
		if deadline, ok := ctx.Deadline(); ok {
			if err := conn.SetDeadline(deadline); err != nil {
				log.Printf("mailer: could not set smtp deadline for %s: %v", toEmail, err)
			}
		}
		client, err = smtp.NewClient(conn, m.cfg.SMTPHost)
		if err != nil {
			log.Printf("mailer: %s email client init failed for %s: %v", kind, toEmail, err)
			return fmt.Errorf("smtp new client: %w", err)
		}
	}
	defer func() {
		if quitErr := client.Quit(); quitErr != nil {
			log.Printf("mailer: smtp quit warning for %s email to %s: %v", kind, toEmail, quitErr)
		}
	}()

	if !m.cfg.SMTPUseSSL && m.cfg.SMTPRequireTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			err := fmt.Errorf("smtp server does not advertise STARTTLS")
			log.Printf("mailer: %s email starttls unavailable for %s", kind, toEmail)
			return err
		}
		if err := client.StartTLS(tlsCfg); err != nil {
			log.Printf("mailer: %s email starttls failed for %s: %v", kind, toEmail, err)
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if ok, _ := client.Extension("AUTH"); ok {
		if err := client.Auth(auth); err != nil {
			log.Printf("mailer: %s email auth failed for %s: %v", kind, toEmail, err)
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(m.cfg.SMTPFromEmail); err != nil {
		log.Printf("mailer: %s email MAIL FROM failed for %s: %v", kind, toEmail, err)
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(toEmail); err != nil {
		log.Printf("mailer: %s email RCPT TO failed for %s: %v", kind, toEmail, err)
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		log.Printf("mailer: %s email DATA failed for %s: %v", kind, toEmail, err)
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		if closeErr := writer.Close(); closeErr != nil {
			log.Printf("mailer: %s email writer close warning for %s after write failure: %v", kind, toEmail, closeErr)
		}
		log.Printf("mailer: %s email write failed for %s: %v", kind, toEmail, err)
		return fmt.Errorf("smtp write message: %w", err)
	}
	if err := writer.Close(); err != nil {
		log.Printf("mailer: %s email close failed for %s: %v", kind, toEmail, err)
		return fmt.Errorf("smtp close writer: %w", err)
	}

	log.Printf("mailer: sent %s email to %s", kind, toEmail)
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

var _ VerificationSender = (*SMTPVerificationMailer)(nil)
