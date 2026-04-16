package mailer

import (
	"context"
	"net/textproto"
)

type Message struct {
	ID             string
	MessageType    string
	RecipientEmail string
	Subject        string
	HTMLBody       string
	TextBody       string
	Headers        textproto.MIMEHeader
	IdempotencyKey string
}

type SendResult struct {
	Provider          string
	ProviderMessageID string
}

type Provider interface {
	Send(ctx context.Context, msg Message) (SendResult, error)
}
