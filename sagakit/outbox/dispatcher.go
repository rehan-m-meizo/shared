package outbox

import (
	"context"
	"shared/pkgs/uuids"
	"shared/sagakit/db"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Dispatcher struct {
	DB        db.DB
	Store     Store
	Pub       message.Publisher
	Logger    watermill.LoggerAdapter
	Limit     int
	Delay     time.Duration
	StopOnErr bool
}

func (d *Dispatcher) Start(ctx context.Context) error {
	if d.Limit <= 0 {
		d.Limit = 100
	}
	if d.Delay <= 0 {
		d.Delay = time.Second
	}
	log := d.Logger

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		tx, err := d.DB.BeginTx(ctx)
		if err != nil {
			return err
		}

		entries, err := d.Store.GetPendingTx(ctx, tx, d.Limit)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if len(entries) == 0 {
			_ = tx.Rollback()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(d.Delay):
				continue
			}
		}

		var sentIDs []string

		for _, e := range entries {
			msg := message.NewMessage(uuids.NewUUID(), e.Payload)
			for k, v := range e.Headers {
				msg.Metadata.Set(k, v)
			}

			if err := d.Pub.Publish(e.Topic, msg); err != nil {
				_ = tx.Rollback()
				log.Error("outbox publish failed", err, watermill.LogFields{
					"id":    e.ID,
					"topic": e.Topic,
				})
				if d.StopOnErr {
					return err
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(d.Delay):
					goto NEXT
				}
			}

			sentIDs = append(sentIDs, e.ID)
		}

		if err := d.Store.MarkSentTx(ctx, tx, sentIDs); err != nil {
			_ = tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

	NEXT:
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d.Delay):
		}
	}
}
