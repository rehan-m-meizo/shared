package ledgerdb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Ledger struct {
	DB *mongo.Database
}

func NewLedger(db *mongo.Database) *Ledger {
	return &Ledger{DB: db}
}

func (l *Ledger) InsertOne(ctx context.Context, collection, entityType, entityID string, createBy string, data bson.M) error {
	col := l.DB.Collection(collection)

	hash := ComputeHash(data, "")
	entry := LedgerEntry{
		ID:         primitive.NewObjectID(),
		EntityType: entityType,
		EntityID:   entityID,
		Data:       data,
		Version:    1,
		CreatedBy:  createBy,
		CreatedAt:  time.Now(),
		Hash:       hash,
	}

	_, err := col.InsertOne(ctx, entry)
	return err
}

func (l *Ledger) UpdateOne(ctx context.Context, collection, entityType, entityID string, createBy string, newData bson.M) error {
	col := l.DB.Collection(collection)

	latest, err := l.FindLatest(ctx, collection, entityType, entityID)
	if err != nil {
		return err
	}

	hash := ComputeHash(newData, latest.Hash)
	entry := LedgerEntry{
		ID:           primitive.NewObjectID(),
		EntityType:   entityType,
		EntityID:     entityID,
		Data:         newData,
		Version:      latest.Version + 1,
		PreviousHash: latest.Hash,
		Hash:         hash,
		CreatedBy:    createBy,
		CreatedAt:    time.Now(),
	}

	_, err = col.InsertOne(ctx, entry)
	return err
}

func (m *Ledger) Approve(ctx context.Context, collection, entityType, entityID string, approvedBy string, formID string) error {
	col := m.DB.Collection(collection)

	latest, err := m.FindLatest(ctx, collection, entityType, entityID)
	if err != nil {
		return err
	}

	if latest.ApprovedBy != "" {
		return errors.New("ledger entry already approved")
	}

	hash := ComputeHash(latest.Data, latest.Hash)
	entry := LedgerEntry{
		ID:           primitive.NewObjectID(),
		EntityType:   entityType,
		EntityID:     entityID,
		Data:         latest.Data,
		Version:      latest.Version + 1,
		PreviousHash: latest.Hash,
		Hash:         hash,
		CreatedBy:    latest.CreatedBy,
		ApprovedBy:   approvedBy,
		CreatedAt:    time.Now(),
		ApprovedAt:   time.Now(),
	}

	_, err = col.InsertOne(ctx, entry)

	return err
}

func (m *Ledger) Reject(ctx context.Context, collection, entityType, entityID string, rejectedBy string, formID string) error {
	col := m.DB.Collection(collection)

	latest, err := m.FindLatest(ctx, collection, entityType, entityID)
	if err != nil {
		return err
	}

	if latest.RejectedBy != "" {
		return errors.New("ledger entry already reviewed")
	}

	if latest.ApprovedBy != "" {
		return errors.New("ledger entry already approved, cannot reject")
	}

	hash := ComputeHash(latest.Data, latest.Hash)
	entry := LedgerEntry{
		ID:           primitive.NewObjectID(),
		EntityType:   entityType,
		EntityID:     entityID,
		Data:         latest.Data,
		Version:      latest.Version + 1,
		PreviousHash: latest.Hash,
		Hash:         hash,
		CreatedBy:    latest.CreatedBy,
		CreatedAt:    time.Now(),
		RejectedBy:   rejectedBy,
		RejectedAt:   time.Now(),
	}

	_, err = col.InsertOne(ctx, entry)
	return err
}

func (l *Ledger) FindLatest(ctx context.Context, collection, entityType, entityID string) (*LedgerEntry, error) {
	col := l.DB.Collection(collection)

	var result LedgerEntry
	err := col.FindOne(ctx, bson.M{
		"entity_type": entityType,
		"entity_id":   entityID,
	}, mongoOptionsLatest()).Decode(&result)

	return &result, err
}

func (l *Ledger) History(ctx context.Context, collection, entityType, entityID string) ([]LedgerEntry, error) {
	col := l.DB.Collection(collection)

	cur, err := col.Find(ctx, bson.M{
		"entity_type": entityType,
		"entity_id":   entityID,
	}, mongoOptionsAllVersions())
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []LedgerEntry
	for cur.Next(ctx) {
		var entry LedgerEntry
		if err := cur.Decode(&entry); err == nil {
			results = append(results, entry)
		}
	}
	return results, nil
}

func (l *Ledger) Revert(ctx context.Context, collection, entityType, entityID string, revertedBy string) error {
	col := l.DB.Collection(collection)

	// Find the latest (live) entry
	latest, err := l.FindLatest(ctx, collection, entityType, entityID)
	if err != nil {
		return err
	}

	// Find the previous version (latest.Version - 1)
	var prev LedgerEntry
	err = col.FindOne(ctx, bson.M{
		"entity_type": entityType,
		"entity_id":   entityID,
		"version":     latest.Version - 1,
	}).Decode(&prev)
	if err != nil {
		return err
	}

	return l.UpdateOne(ctx, collection, entityType, entityID, revertedBy, prev.Data)
}

func (l *Ledger) Delete(ctx context.Context, collection, entityType, entityID, deletedBy string, data bson.M) error {
	col := l.DB.Collection(collection)

	latest, err := l.FindLatest(ctx, collection, entityType, entityID)

	if err != nil {
		return err
	}

	hash := ComputeHash(data, latest.Hash)
	entry := LedgerEntry{
		ID:           primitive.NewObjectID(),
		EntityType:   entityType,
		EntityID:     entityID,
		Data:         data,
		Version:      latest.Version + 1,
		PreviousHash: latest.Hash,
		Hash:         hash,
		DeletedBy:    deletedBy,
		DeletedAt:    time.Now(),
	}

	_, err = col.InsertOne(ctx, entry)

	return err
}

func (l *Ledger) Diff(ctx context.Context, collection, entityType, entityID string, v1, v2 int) (map[string][2]any, error) {
	col := l.DB.Collection(collection)

	var entry1, entry2 LedgerEntry
	err := col.FindOne(ctx, bson.M{
		"entity_type": entityType,
		"entity_id":   entityID,
		"version":     v1,
	}).Decode(&entry1)
	if err != nil {
		return nil, err
	}

	err = col.FindOne(ctx, bson.M{
		"entity_type": entityType,
		"entity_id":   entityID,
		"version":     v2,
	}).Decode(&entry2)
	if err != nil {
		return nil, err
	}

	return Diff(entry1.Data, entry2.Data), nil
}
