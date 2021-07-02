package mongo

import (
	"context"
	"log"
	"time"

	"github.com/joergjo/go-samples/booklibrary"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongoCollectionStore stores Book instances in a MongoDB collection.
type mongoCollectionStore struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

var (
	// Compile-time check to verify we implement Storage
	_              booklibrary.Storage = (*mongoCollectionStore)(nil)
	timeout                            = 2 * time.Second
	startupTimeout                     = 10 * time.Second
)

// NewStorage creates a new Storage backed by MongoDB
func NewStorage(mongoURI, database, collection string) (booklibrary.Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
	defer cancel()

	// Set client options
	opts := options.Client().ApplyURI(mongoURI)
	if err := opts.Validate(); err != nil {
		log.Printf("Validating client options failed: %+v\n", opts)
		return nil, err
	}

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Printf("Connecting to MongoDB failed: %s\n", err)
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Printf("Pinging MongoDB failed: %s\n", err)
		return nil, err
	}

	db := client.Database(database)
	coll := db.Collection(collection)
	store := &mongoCollectionStore{
		client:     client,
		database:   db,
		collection: coll,
	}
	return store, nil
}

// All returns all books up to 'limit' instances.
func (m *mongoCollectionStore) All(ctx context.Context, limit int) ([]booklibrary.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	books, err := m.find(ctx, bson.M{}, limit)
	if err != nil {
		return nil, err
	}
	return books, nil
}

// Book finds a book by its ID.
func (m *mongoCollectionStore) Book(ctx context.Context, id string) (booklibrary.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Parsing ObjectID %s failed: %s\n", id, err)
		return booklibrary.Book{}, booklibrary.ErrInvalidID
	}

	filter := bson.M{"_id": oid}
	books, err := m.find(ctx, filter, 1)
	if err != nil {
		return booklibrary.Book{}, err
	}
	if len(books) == 0 {
		return booklibrary.Book{}, booklibrary.ErrNotFound
	}
	return books[0], nil
}

// Add adds a new book
func (m *mongoCollectionStore) Add(ctx context.Context, book booklibrary.Book) (booklibrary.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := m.collection.InsertOne(ctx, book)
	if err != nil {
		log.Printf("Inserting document failed: %s\n", err)
		return booklibrary.Book{}, err
	}
	book.ID = res.InsertedID.(primitive.ObjectID)
	return book, nil
}

// Update a book for specific ID
func (m *mongoCollectionStore) Update(ctx context.Context, id string, book booklibrary.Book) (booklibrary.Book, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Parsing ObjectID %s failed: %s\n", id, err)
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
	res := m.collection.FindOneAndUpdate(ctx, filter, update, options)
	if err := res.Err(); err != nil {
		log.Printf("Updating document %s failed: %s\n", id, err)
		if err != mongo.ErrNoDocuments {
			return booklibrary.Book{}, err
		}
		return booklibrary.Book{}, booklibrary.ErrNotFound
	}

	var b booklibrary.Book
	err = res.Decode(&b)
	if err != nil {
		log.Printf("Decoding document failed: %s\n", err)
		return booklibrary.Book{}, err
	}
	return b, nil
}

// Remove deletes a book from the database
func (m *mongoCollectionStore) Remove(ctx context.Context, id string) (booklibrary.Book, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("Parsing ObjectID %s failed: %s\n", id, err)
		return booklibrary.Book{}, booklibrary.ErrInvalidID
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	filter := bson.M{"_id": oid}
	res := m.collection.FindOneAndDelete(ctx, filter)
	if err := res.Err(); err != nil {
		log.Printf("Deleting document %s failed: %s\n", id, err)
		if err != mongo.ErrNoDocuments {
			return booklibrary.Book{}, err
		}
		return booklibrary.Book{}, booklibrary.ErrNotFound
	}

	var b booklibrary.Book
	err = res.Decode(&b)
	if err != nil {
		log.Printf("Decoding document failed: %s\n", err)
		return booklibrary.Book{}, err
	}
	return b, nil
}

func (m *mongoCollectionStore) find(ctx context.Context, filter primitive.M, limit int) ([]booklibrary.Book, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	findOptions := options.Find().SetLimit(int64(limit))
	cur, err := m.collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("Finding document(s) failed: %s\n", err)
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
			log.Printf("Decoding document failed: %s\n", err)
			break
		}
		books = append(books, b)
	}

	if err := cur.Err(); err != nil {
		log.Printf("Iterating over cursor failed: %s\n", err)
		return books, err
	}
	return books, nil
}
