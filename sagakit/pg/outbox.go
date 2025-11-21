package pg

import (
	"context"
	"encoding/json"
	"fmt"
	"shared/sagakit/db"
	"shared/sagakit/outbox"

	"github.com/ThreeDotsLabs/watermill/message"
)

// SQL schema (run once per DB):
// CREATE TABLE IF NOT EXISTS outbox (
//   id BIGSERIAL PRIMARY KEY,
//   topic TEXT NOT NULL,
//   payload BYTEA NOT NULL,
//   headers JSONB,
//   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
//   sent_at TIMESTAMPTZ
// );

type Outbox struct{}

func NewOutbox() *Outbox { return &Outbox{} }

func (o *Outbox) InsertTx(ctx context.Context, tx db.Tx, topic string, msg *message.Message) error {
	headers, err := json.Marshal(msg.Metadata)
	if err != nil {
		return err
	}

	return tx.Exec(ctx,
		`INSERT INTO outbox(topic, payload, headers) VALUES ($1, $2, $3)`,
		topic, msg.Payload, headers,
	)
}

func (o *Outbox) GetPendingTx(ctx context.Context, tx db.Tx, limit int) ([]outbox.Entry, error) {
	rows, err := tx.Query(ctx,
		`SELECT id, topic, payload, headers
           FROM outbox
          WHERE sent_at IS NULL
          ORDER BY id
          LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []outbox.Entry
	for rows.Next() {
		var (
			id      int64
			topic   string
			payload []byte
			headers []byte
		)
		if err := rows.Scan(&id, &topic, &payload, &headers); err != nil {
			return nil, err
		}
		entry := outbox.Entry{
			ID:      fmt.Sprintf("%d", id),
			Topic:   topic,
			Payload: payload,
			Headers: map[string]string{},
		}
		if len(headers) > 0 {
			_ = json.Unmarshal(headers, &entry.Headers)
		}
		res = append(res, entry)
	}
	return res, nil
}

func (o *Outbox) MarkSentTx(ctx context.Context, tx db.Tx, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return tx.Exec(ctx,
		`UPDATE outbox SET sent_at = now() WHERE id::text = ANY($1)`,
		ids,
	)
}
