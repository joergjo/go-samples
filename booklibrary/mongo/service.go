package mongo

import (
	"context"
	"time"

	"github.com/joergjo/go-samples/booklibrary"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/exp/slog"
)

// MongoCollectionStore stores Book instances in a MongoDB collection.
type CrudService struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

var (
	// Compile-time check to verify we implement Storage
	_              booklibrary.CrudService = (*CrudService)(nil)
	timeout                                = 2 * time.Second
	startupTimeout                         = 10 * time.Second
)

// NewStorage creates a new Storage backed by MongoDB
func NewCrudService(mongoURI, database, collection string) (*CrudService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	// Set client options
	opts := options.Client().ApplyURI(mongoURI)
	if err := opts.Validate(); err != nil {
		slog.Error("validating client options", booklibrary.ErrorKey, err, slog.Any("options", opts))
		return nil, err
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		slog.Error("connecting to MongoDB", booklibrary.ErrorKey, err)
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		slog.Error("pinging MongoDB", booklibrary.ErrorKey, err)
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
func (cs *CrudService) List(ctx context.Context, limit int) ([]booklibrary.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	books, err := cs.find(ctx, bson.M{}, limit)
	if err != nil {
		return nil, err
	}
	return books, nil
}

// Book finds a book by its ID.
func (cs *CrudService) Get(ctx context.Context, id string) (booklibrary.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		slog.Error("parsing ObjectID", booklibrary.ErrorKey, err, slog.String("id", id))
		return booklibrary.Book{}, booklibrary.ErrInvalidID
	}

	filter := bson.M{"_id": oid}
	books, err := cs.find(ctx, filter, 1)
	if err != nil {
		return booklibrary.Book{}, err
	}
	if len(books) == 0 {
		return booklibrary.Book{}, booklibrary.ErrNotFound
	}
	return books[0], nil
}

// Add adds a new book
func (cs *CrudService) Add(ctx context.Context, book booklibrary.Book) (booklibrary.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := cs.collection.InsertOne(ctx, book)
	if err != nil {
		slog.Error("inserting document", booklibrary.ErrorKey, err)
		return booklibrary.Book{}, err
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		panic("inserted ID is not an ObjectID")
	}
	book.ID = oid.Hex()
	return book, nil
}

// Update a book for specific ID
func (cs *CrudService) Update(ctx context.Context, id string, book booklibrary.Book) (booklibrary.Book, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		slog.Error("parsing ObjectID", booklibrary.ErrorKey, err, booklibrary.IdKey, id)
		return booklibrary.Book{}, booklibrary.ErrInvalidID
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
		slog.Error("updating document", booklibrary.ErrorKey, err, booklibrary.IdKey, id)
		if err != mongo.ErrNoDocuments {
			return booklibrary.Book{}, err
		}
		return booklibrary.Book{}, booklibrary.ErrNotFound
	}

	var b booklibrary.Book
	err = res.Decode(&b)
	if err != nil {
		slog.Error("decoding document", booklibrary.ErrorKey, err)
		return booklibrary.Book{}, err
	}
	return b, nil
}

// Remove deletes a book from the database
func (cs *CrudService) Remove(ctx context.Context, id string) (booklibrary.Book, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		slog.Error("parsing ObjectID", booklibrary.ErrorKey, err, booklibrary.IdKey, id)
		return booklibrary.Book{}, booklibrary.ErrInvalidID
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	filter := bson.M{"_id": oid}
	res := cs.collection.FindOneAndDelete(ctx, filter)
	if err := res.Err(); err != nil {
		slog.Error("deleting document", booklibrary.ErrorKey, err, booklibrary.IdKey, id)
		if err != mongo.ErrNoDocuments {
			return booklibrary.Book{}, err
		}
		return booklibrary.Book{}, booklibrary.ErrNotFound
	}

	var b booklibrary.Book
	err = res.Decode(&b)
	if err != nil {
		slog.Error("decoding document", booklibrary.ErrorKey, err)
		return booklibrary.Book{}, err
	}
	return b, nil
}

func (cs *CrudService) find(ctx context.Context, filter primitive.M, limit int) ([]booklibrary.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	findOptions := options.Find().SetLimit(int64(limit))
	cur, err := cs.collection.Find(ctx, filter, findOptions)
	if err != nil {
		slog.Error("finding document(s)", booklibrary.ErrorKey, err)
		return nil, err
	}
	defer cur.Close(ctx)

	books := []booklibrary.Book{}
	for cur.Next(ctx) {
		var b booklibrary.Book
		err := cur.Decode(&b)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return []booklibrary.Book{}, nil
			}
			slog.Error("decoding document", booklibrary.ErrorKey, err)
			break
		}
		books = append(books, b)
	}

	if err := cur.Err(); err != nil {
		slog.Error("iterating over cursor", booklibrary.ErrorKey, err)
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
