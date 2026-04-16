package mailer

import (
	"context"
	"log"

	"github.com/sergehall/lavoval/apps/api/internal/config"
)

type NoopProvider struct {
	cfg config.Config
}

func NewNoopProvider(cfg config.Config) *NoopProvider {
	return &NoopProvider{cfg: cfg}
}

func (p *NoopProvider) Send(_ context.Context, msg Message) (SendResult, error) {
	log.Printf("mailer: status=noop_sent provider=noop message_type=%s recipient_domain=%s message_id=%s", msg.MessageType, recipientDomain(msg.RecipientEmail), msg.ID)
	return SendResult{
		Provider:          "noop",
		ProviderMessageID: msg.ID,
	}, nil
}

var _ Provider = (*NoopProvider)(nil)
