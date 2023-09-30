package mongo

import (
	"context"
	"errors"
	"time"

	"log/slog"

	"github.com/joergjo/go-samples/booklibrary/internal/log"
	"github.com/joergjo/go-samples/booklibrary/internal/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoCollectionStore stores Book instances in a MongoDB collection.
type CrudService struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

var (
	// Compile-time check to verify we implement Storage
	_                  model.CrudService = (*CrudService)(nil)
	timeout                              = 2 * time.Second
	startupTimeout                       = 10 * time.Second
	connectionIDKey                      = "connectionID"
	heartbeatSucceeded                   = promauto.NewCounter(prometheus.CounterOpts{
		Name: "booklibrary_mongodb_heartbeat_succeeded_total",
		Help: "The total number of successful MongoDB server heartbeats",
	})
	heartbeatFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "booklibrary_mongodb_server_heartbeat_failed_total",
		Help: "The total number of failed MongoDB server heartbeats",
	})
)

func newMonitor() *event.ServerMonitor {
	return &event.ServerMonitor{
		ServerHeartbeatFailed: func(evt *event.ServerHeartbeatFailedEvent) {
			heartbeatFailed.Inc()
			slog.Warn("MongoDB server heartbeat failed", log.ErrorKey, evt.Failure, connectionIDKey, evt.ConnectionID)
		},
		ServerHeartbeatSucceeded: func(evt *event.ServerHeartbeatSucceededEvent) {
			heartbeatSucceeded.Inc()
			slog.Debug("server heartbeat succeeded", connectionIDKey, evt.ConnectionID)
		},
	}
}

// NewStorage creates a new Storage backed by MongoDB
func NewCrudService(mongoURI, database, collection string) (*CrudService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	// Set client options
	opts := options.Client().ApplyURI(mongoURI).SetServerMonitor(newMonitor())
	if err := opts.Validate(); err != nil {
		slog.Error("validating client options", log.ErrorKey, err, slog.Any("options", opts))
		return nil, err
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		slog.Error("connecting to MongoDB", log.ErrorKey, err)
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		slog.Error("pinging MongoDB", log.ErrorKey, err)
		return nil, err
	}

	db := client.Database(database)
	coll := db.Collection(collection)
	crud := CrudService{
		client:     client,
		database:   db,
		collection: coll,
	}
	return &crud, nil
}

// All returns all books up to 'limit' instances.
func (cs *CrudService) List(ctx context.Context, limit int) ([]model.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	books, err := cs.find(ctx, bson.M{}, limit)
	if err != nil {
		return nil, err
	}
	return books, nil
}

// Book finds a book by its ID.
func (cs *CrudService) Get(ctx context.Context, id string) (model.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		slog.Error("parsing ObjectID", log.ErrorKey, err, slog.String("id", id))
		return model.Book{}, model.ErrInvalidID
	}

	filter := bson.M{"_id": oid}
	books, err := cs.find(ctx, filter, 1)
	if err != nil {
		return model.Book{}, err
	}
	if len(books) == 0 {
		return model.Book{}, model.ErrNotFound
	}
	return books[0], nil
}

// Add adds a new book
func (cs *CrudService) Add(ctx context.Context, book model.Book) (model.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := cs.collection.InsertOne(ctx, book)
	if err != nil {
		slog.Error("inserting document", log.ErrorKey, err)
		return model.Book{}, err
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		panic("inserted ID is not an ObjectID")
	}
	book.ID = oid.Hex()
	return book, nil
}

// Update a book for specific ID
func (cs *CrudService) Update(ctx context.Context, id string, book model.Book) (model.Book, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		slog.Error("parsing ObjectID", log.ErrorKey, err, log.IdKey, id)
		return model.Book{}, model.ErrInvalidID
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": bson.M{
		"title":       book.Title,
		"author":      book.Author,
		"releaseDate": book.ReleaseDate,
		"keywords":    book.Keywords}}
	res := cs.collection.FindOneAndUpdate(ctx, filter, update, options)
	if err := res.Err(); err != nil {
		slog.Error("updating document", log.ErrorKey, err, log.IdKey, id)
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return model.Book{}, err
		}
		return model.Book{}, model.ErrNotFound
	}

	var b model.Book
	err = res.Decode(&b)
	if err != nil {
		slog.Error("decoding document", log.ErrorKey, err)
		return model.Book{}, err
	}
	return b, nil
}

// Remove deletes a book from the database
func (cs *CrudService) Remove(ctx context.Context, id string) (model.Book, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		slog.Error("parsing ObjectID", log.ErrorKey, err, log.IdKey, id)
		return model.Book{}, model.ErrInvalidID
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	filter := bson.M{"_id": oid}
	res := cs.collection.FindOneAndDelete(ctx, filter)
	if err := res.Err(); err != nil {
		slog.Error("deleting document", log.ErrorKey, err, log.IdKey, id)
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return model.Book{}, err
		}
		return model.Book{}, model.ErrNotFound
	}

	var b model.Book
	err = res.Decode(&b)
	if err != nil {
		slog.Error("decoding document", log.ErrorKey, err)
		return model.Book{}, err
	}
	return b, nil
}

func (cs *CrudService) find(ctx context.Context, filter primitive.M, limit int) ([]model.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	findOptions := options.Find().SetLimit(int64(limit))
	cur, err := cs.collection.Find(ctx, filter, findOptions)
	if err != nil {
		slog.Error("finding document(s)", log.ErrorKey, err)
		return nil, err
	}
	defer cur.Close(ctx)

	books := []model.Book{}
	for cur.Next(ctx) {
		var b model.Book
		err := cur.Decode(&b)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return []model.Book{}, nil
			}
			slog.Error("decoding document", log.ErrorKey, err)
			break
		}
		books = append(books, b)
	}

	if err := cur.Err(); err != nil {
		slog.Error("iterating over cursor", log.ErrorKey, err)
		return nil, err
	}
	return books, nil
}

func (m *CrudService) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

func (m *CrudService) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, readpref.Primary())
}
