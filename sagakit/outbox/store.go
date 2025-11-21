package outbox

import (
	"context"
	"shared/sagakit/db"

	"github.com/ThreeDotsLabs/watermill/message"
)

type Entry struct {
	ID      string
	Topic   string
	Payload []byte
	Headers map[string]string
}

type Store interface {
	InsertTx(ctx context.Context, tx db.Tx, topic string, msg *message.Message) error
	GetPendingTx(ctx context.Context, tx db.Tx, limit int) ([]Entry, error)
	MarkSentTx(ctx context.Context, tx db.Tx, ids []string) error
}
