package mailer

import (
	"fmt"
	"net/textproto"
)

func buildVerificationMessage(jobID string, email VerificationEmail, appURL string) (Message, error) {
	rendered, err := renderVerificationEmail(email, appURL)
	if err != nil {
		return Message{}, permanentDeliveryError("render_failed", fmt.Errorf("render verification email: %w", err))
	}
	return newMessage(jobID, mailJobTypeVerification, email.ToEmail, rendered), nil
}

func buildPasswordResetMessage(jobID string, email PasswordResetEmail, appURL string) (Message, error) {
	rendered, err := renderPasswordResetEmail(email, appURL)
	if err != nil {
		return Message{}, permanentDeliveryError("render_failed", fmt.Errorf("render password reset email: %w", err))
	}
	return newMessage(jobID, mailJobTypePasswordReset, email.ToEmail, rendered), nil
}

func buildPasswordChangedMessage(jobID string, email PasswordChangedEmail, appURL string) (Message, error) {
	rendered, err := renderPasswordChangedEmail(email, appURL)
	if err != nil {
		return Message{}, permanentDeliveryError("render_failed", fmt.Errorf("render password changed email: %w", err))
	}
	return newMessage(jobID, mailJobTypePasswordChange, email.ToEmail, rendered), nil
}

func newMessage(jobID string, messageType string, recipientEmail string, rendered RenderedEmail) Message {
	headers := textproto.MIMEHeader{}
	headers.Set("X-Lavoval-Message-ID", jobID)
	headers.Set("X-Lavoval-Message-Type", messageType)

	return Message{
		ID:             jobID,
		MessageType:    messageType,
		RecipientEmail: recipientEmail,
		Subject:        rendered.Subject,
		HTMLBody:       rendered.HTMLBody,
		TextBody:       rendered.TextBody,
		Headers:        headers,
		IdempotencyKey: jobID,
	}
}
