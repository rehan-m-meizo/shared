package mdb

import (
	"context"
	"fmt"
	"log"
	"math"
	"shared/constants"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	client *mongo.Client
}

var (
	mongoInstance *MongoStore
	once          sync.Once
)

type PaginatedResult struct {
	Cursor      *mongo.Cursor `json:"-"`
	TotalCount  int64         `json:"totalCount"`
	TotalPages  int64         `json:"totalPages"`
	CurrentPage int64         `json:"currentPage"`
	Limit       int64         `json:"limit"`
	HasNextPage bool          `json:"hasNextPage"`
	HasPrevPage bool          `json:"hasPrevPage"`
}

func ensureBsonType(arg interface{}, name string) error {
	switch arg.(type) {
	case bson.M, bson.D, nil:
		return nil
	default:
		return fmt.Errorf("%s must be bson.M, bson.D, or nil", name)
	}
}

// InitMongo creates a singleton Mongo client
func InitMongo() error {
	var err error
	once.Do(func() {
		ctx := context.Background()
		uri := fmt.Sprintf("mongodb://%s:%s@%s:%s",
			constants.MongoUser, constants.MongoPassword,
			constants.MongoHost, constants.MongoPort,
		)

		clientOpts := options.Client().ApplyURI(uri).SetMaxPoolSize(20)
		client, connErr := mongo.Connect(ctx, clientOpts)
		if connErr != nil {
			err = connErr
			return
		}

		if pingErr := client.Ping(ctx, nil); pingErr != nil {
			err = pingErr
			return
		}

		mongoInstance = &MongoStore{
			client: client,
		}
		log.Println("✅ MongoDB connected")
	})
	return err
}

func GetMongo() *MongoStore {
	if mongoInstance == nil {
		log.Fatal("Mongo not initialized. Call InitMongo() first.")
	}
	return mongoInstance
}

func (m *MongoStore) Disconnect(ctx context.Context) {
	if m.client != nil {
		_ = m.client.Disconnect(ctx)
		log.Println("🛑 MongoDB disconnected")
	}
}

func (m *MongoStore) GetClient() *mongo.Client {
	if m.client == nil {
		log.Fatal("Mongo client is not initialized. Call InitMongo() first.")
	}
	return m.client
}

// ──────────────── CRUD METHODS ────────────────

func (m *MongoStore) InsertOne(ctx context.Context, database, collection string, doc interface{}) (*mongo.InsertOneResult, error) {
	return m.client.Database(database).Collection(collection).InsertOne(ctx, doc)
}

func (m *MongoStore) InsertMany(ctx context.Context, database, collection string, docs []interface{}) (*mongo.InsertManyResult, error) {
	return m.client.Database(database).Collection(collection).InsertMany(ctx, docs)
}

func (m *MongoStore) InsertManyWithResult(ctx context.Context, database, collection string, docs []interface{}) (*mongo.InsertManyResult, error) {
	return m.client.Database(database).Collection(collection).InsertMany(ctx, docs)
}

func (m *MongoStore) FindOneWithSortedId(
	ctx context.Context,
	database, collection string,
	filter interface{},
	column string,
) (*mongo.SingleResult, error) {

	if err := ensureBsonType(filter, "filter"); err != nil {
		return nil, err
	}

	opts := options.FindOne().
		SetSort(bson.D{{Key: column, Value: -1}}).
		SetProjection(bson.D{{Key: column, Value: 1}})

	res := m.client.
		Database(database).
		Collection(collection).
		FindOne(ctx, filter, opts)

	if err := res.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (m *MongoStore) FindOne(ctx context.Context, database, collection string, filter interface{}, projection interface{}) (*mongo.SingleResult, error) {

	opts := options.FindOne()
	if projection != nil {
		switch projection.(type) {
		case bson.M, bson.D:
			opts.SetProjection(projection)
		default:
			return nil, fmt.Errorf("projection must be bson.M or bson.D")
		}
	}

	res := m.client.Database(database).Collection(collection).FindOne(ctx, filter, opts)
	return res, res.Err()
}

func (m *MongoStore) UpdateOne(ctx context.Context, database, collection string, filter interface{}, update interface{}, upsert bool) (*mongo.UpdateResult, error) {
	opts := options.Update().SetUpsert(upsert)
	return m.client.Database(database).Collection(collection).UpdateOne(
		ctx,
		filter,
		bson.M{"$set": update},
		opts,
	)
}

func (m *MongoStore) UpdateMany(ctx context.Context, database, collection string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return m.client.Database(database).Collection(collection).UpdateMany(ctx, filter, bson.M{"$set": update})
}

func (m *MongoStore) DeleteOne(ctx context.Context, database, collection string, filter interface{}) (*mongo.DeleteResult, error) {
	return m.client.Database(database).Collection(collection).DeleteOne(ctx, filter)
}

func (m *MongoStore) Count(ctx context.Context, database, collection string, filter interface{}) (int64, error) {
	return m.client.Database(database).Collection(collection).CountDocuments(ctx, filter)
}

func (m *MongoStore) Find(ctx context.Context, database, collection string, filter, projection, sort interface{}, limit int64) (*mongo.Cursor, error) {
	opts := options.Find()
	if projection != nil {
		opts.SetProjection(projection)
	}
	if sort != nil {
		opts.SetSort(sort)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}
	return m.client.Database(database).Collection(collection).Find(ctx, filter, opts)
}

func (m *MongoStore) Paginate(ctx context.Context, database, collection string, filter, projection, sort interface{}, page, limit int64) (*mongo.Cursor, error) {
	opts := options.Find()
	if projection != nil {
		opts.SetProjection(projection)
	}
	if sort != nil {
		opts.SetSort(sort)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}
	opts.SetSkip((page - 1) * limit)
	return m.client.Database(database).Collection(collection).Find(ctx, filter, opts)
}

func (m *MongoStore) AdvancePagination(
	ctx context.Context,
	database, collection string,
	filter, projection, sort interface{},
	page, limit int64,
) (*PaginatedResult, error) {

	opts := options.Find()
	if projection != nil {
		opts.SetProjection(projection)
	}
	if sort != nil {
		opts.SetSort(sort)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if page > 0 {
		opts.SetSkip((page - 1) * limit)
	}

	col := m.client.Database(database).Collection(collection)

	// Get total count
	totalCount, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Get cursor for current page
	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	totalPages := int64(0)
	if limit > 0 {
		totalPages = int64(math.Ceil(float64(totalCount) / float64(limit)))
	}

	return &PaginatedResult{
		Cursor:      cursor,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: page,
		Limit:       limit,
		HasNextPage: page < totalPages,
		HasPrevPage: page > 1,
	}, nil
}

func (m *MongoStore) Aggregate(ctx context.Context, database, collection string, pipeline interface{}) (*mongo.Cursor, error) {
	opts := options.Aggregate().SetAllowDiskUse(true)
	return m.client.Database(database).Collection(collection).Aggregate(ctx, pipeline, opts)
}

// AggregatePaginated executes an aggregation pipes with pagination and returns total count & cursor.
// AggregatePaginated executes an aggregation pipeline with pagination and optional sorting.
// It returns a PaginatedResult with cursor and metadata similar to AdvancePagination.
func (m *MongoStore) AggregatePaginated(
	ctx context.Context,
	database, collection string,
	basePipeline mongo.Pipeline,
	sort interface{}, // e.g. bson.D{{"created_at", -1}}
	page, limit int64,
) (*PaginatedResult, error) {

	if m.client == nil {
		return nil, fmt.Errorf("Mongo client is not initialized")
	}

	if basePipeline == nil {
		return nil, fmt.Errorf("pipeline cannot be nil")
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	col := m.client.Database(database).Collection(collection)

	// ---------- 1️⃣ Get total count ----------
	countPipeline := append(basePipeline, bson.D{{Key: "$count", Value: "total"}})
	countCursor, err := col.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregation count failed: %w", err)
	}
	var countResult []bson.M
	if err := countCursor.All(ctx, &countResult); err != nil {
		return nil, fmt.Errorf("failed to decode count results: %w", err)
	}

	var totalCount int64
	if len(countResult) > 0 {
		switch v := countResult[0]["total"].(type) {
		case int32:
			totalCount = int64(v)
		case int64:
			totalCount = v
		}
	}

	// ---------- 2️⃣ Apply sort, skip & limit ----------
	skip := (page - 1) * limit

	pagedPipeline := append(mongo.Pipeline{}, basePipeline...)

	if sort != nil {
		pagedPipeline = append(pagedPipeline, bson.D{{Key: "$sort", Value: sort}})
	}

	pagedPipeline = append(pagedPipeline,
		bson.D{{Key: "$skip", Value: skip}},
		bson.D{{Key: "$limit", Value: limit}},
	)

	// ---------- 3️⃣ Execute the paginated pipeline ----------
	opts := options.Aggregate().SetAllowDiskUse(true)
	cursor, err := col.Aggregate(ctx, pagedPipeline, opts)
	if err != nil {
		return nil, fmt.Errorf("aggregation with pagination failed: %w", err)
	}

	// ---------- 4️⃣ Build pagination metadata ----------
	totalPages := int64(math.Ceil(float64(totalCount) / float64(limit)))

	return &PaginatedResult{
		Cursor:      cursor,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: page,
		Limit:       limit,
		HasNextPage: page < totalPages,
		HasPrevPage: page > 1,
	}, nil
}

func (m *MongoStore) StartTransaction(ctx context.Context) (mongo.Session, mongo.SessionContext, error) {
	session, err := m.client.StartSession()
	if err != nil {
		return nil, nil, err
	}

	// Create a session context to use for all operations
	sessCtx := mongo.NewSessionContext(ctx, session)

	// Start the transaction
	if err := session.StartTransaction(); err != nil {
		session.EndSession(ctx)
		return nil, nil, err
	}

	return session, sessCtx, nil
}
