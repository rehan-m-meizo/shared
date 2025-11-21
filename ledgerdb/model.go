package ledgerdb

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LedgerEntry struct {
	ID           primitive.ObjectID `bson:"_id"`
	EntityType   string             `bson:"entity_type"`
	EntityID     string             `bson:"entity_id"`
	Version      int                `bson:"version"`
	Data         bson.M             `bson:"data"`
	PreviousHash string             `bson:"previous_hash,omitempty"`
	Hash         string             `bson:"hash"`
	CreatedBy    string             `bson:"created_by"`
	CreatedAt    time.Time          `bson:"created_at"`
	ApprovedAt   time.Time          `bson:"approved_at,omitempty"`
	ApprovedBy   string             `bson:"approved_by,omitempty"`
	RejectedAt   time.Time          `bson:"rejected_at,omitempty"`
	RejectedBy   string             `bson:"rejected_by,omitempty"`
	RevertedAt   time.Time          `bson:"reverted_at,omitempty"`
	RevertedBy   string             `bson:"reverted_by,omitempty"`
	DeletedBy    string             `bson:"deleted_by,omitempty"`
	DeletedAt    time.Time          `bson:"deleted_at,omitempty"`
}
