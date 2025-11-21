package sagakit

import (
	"context"
	"encoding/json"
	"errors"
	"shared/sagakit/db"
	"shared/sagakit/outbox"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type UnitOfWork interface {
	Tx() db.Tx
	Publish(topic string, payload any, metadata map[string]string) error
}

type unitOfWork struct {
	tx    db.Tx
	store outbox.Store
}

func (u *unitOfWork) Tx() db.Tx { return u.tx }

func (u *unitOfWork) Publish(topic string, payload any, metadata map[string]string) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), b)

	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	return u.store.InsertTx(context.Background(), u.tx, topic, msg)
}

func RunInTx(ctx context.Context, database db.DB, store outbox.Store, fn func(uow UnitOfWork) error) error {
	tx, err := database.BeginTx(ctx)
	if err != nil {
		return err
	}

	uow := &unitOfWork{tx: tx, store: store}
	if err := fn(uow); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func PublishWithMeta(ctx context.Context, topic string, payload any, metadata map[string]string) error {
	if globalDB == nil || globalStore == nil {
		return errors.New("sagakit not initialized")
	}

	return RunInTx(ctx, globalDB, globalStore, func(uow UnitOfWork) error {
		return uow.Publish(topic, payload, metadata)
	})
}
