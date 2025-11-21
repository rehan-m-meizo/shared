package sagakit

import (
	"context"
	"fmt"
	"time"

	"shared/sagakit/db"
	"shared/sagakit/outbox"
	pgdb "shared/sagakit/pg"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	globalDB    db.DB
	globalStore outbox.Store
	globalPub   message.Publisher
	logger      watermill.LoggerAdapter
)

// Config defines everything needed to run sagakit
type Config struct {
	PostgresDSN  string
	KafkaBrokers []string
	KafkaUser    string
	KafkaPass    string
}

// Init bootstraps sagakit (DB + store + Kafka)
func Init(cfg Config) error {
	logger = NewStdLogger()

	// ---- PostgreSQL ----
	pool, err := pgxpool.New(context.Background(), cfg.PostgresDSN)
	if err != nil {
		return fmt.Errorf("failed to init pgx pool: %w", err)
	}

	dbx := pgdb.NewDB(pool)
	store := pgdb.NewOutbox()

	// Ensure table exists
	err = RunInTx(context.Background(), dbx, store, func(uow UnitOfWork) error {
		return uow.Tx().Exec(context.Background(), `
			CREATE TABLE IF NOT EXISTS outbox (
				id BIGSERIAL PRIMARY KEY,
				topic TEXT NOT NULL,
				payload BYTEA NOT NULL,
				headers JSONB,
				created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
				sent_at TIMESTAMPTZ
			)
		`)
	})
	if err != nil {
		return err
	}

	globalDB = dbx
	globalStore = store

	// ---- Kafka ----
	pub, err := NewKafkaPublisher(logger, cfg.KafkaBrokers, cfg.KafkaUser, cfg.KafkaPass)
	if err != nil {
		return fmt.Errorf("cannot create kafka publisher: %w", err)
	}
	globalPub = pub

	return nil
}

// StartDispatcher starts outbox → Kafka background delivery
func StartDispatcher(ctx context.Context) error {
	d := &outbox.Dispatcher{
		DB:     globalDB,
		Store:  globalStore,
		Pub:    globalPub,
		Logger: logger,
		Limit:  100,
		Delay:  time.Second,
	}
	return d.Start(ctx)
}

// Publish helper (global)
func Publish(ctx context.Context, topic string, payload any) error {
	return PublishWithMeta(ctx, topic, payload, nil)
}

// GetGlobalStore returns the global outbox store
func GetGlobalStore() outbox.Store {
	return globalStore
}

// GetLogger returns the global logger
func GetLogger() watermill.LoggerAdapter {
	return logger
}

// GetDB returns the global database
func GetDB() db.DB {
	return globalDB
}
