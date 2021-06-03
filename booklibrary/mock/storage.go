package mock

import (
	"context"
	"time"

	"github.com/joergjo/go-samples/booklibrary"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mockStore abstracts the data access from the underlying data store
type mockStore struct {
	items map[string]booklibrary.Book
}

// Compile-time check to verify we implement Storage
var _ booklibrary.Storage = &mockStore{}

// NewStorage creates a new Storage instance
func NewStorage(books []booklibrary.Book) (booklibrary.Storage, error) {
	m := &mockStore{items: make(map[string]booklibrary.Book)}
	for _, b := range books {
		m.Add(context.TODO(), b)
	}
	return m, nil
}

// All finds all books
func (m *mockStore) All(_ context.Context, limit int64) ([]booklibrary.Book, error) {
	all := []booklibrary.Book{}
	for _, b := range m.items {
		all = append(all, b)
	}
	last := int64(len(all))
	if limit < last {
		last = limit
	}
	return all[:last], nil
}

// Book finds a specific book
func (m *mockStore) Book(_ context.Context, id string) (booklibrary.Book, error) {
	b, ok := m.items[id]
	if !ok {
		return booklibrary.Book{}, booklibrary.ErrNotFound
	}
	return b, nil
}

// Add ads a new Book
func (m *mockStore) Add(ctx context.Context, book booklibrary.Book) (booklibrary.Book, error) {
	if book.ID == primitive.NilObjectID {
		book.ID = primitive.NewObjectID()
	}
	return m.insert(ctx, book)
}

// Update updates an existing Book
func (m *mockStore) Update(_ context.Context, id string, book booklibrary.Book) (booklibrary.Book, error) {
	if _, ok := m.items[id]; !ok {
		return booklibrary.Book{}, booklibrary.ErrNotFound
	}
	m.items[id] = book
	return book, nil
}

// Remove removes a Book
func (m *mockStore) Remove(_ context.Context, id string) (booklibrary.Book, error) {
	book, ok := m.items[id]
	if !ok {
		return booklibrary.Book{}, booklibrary.ErrNotFound
	}
	delete(m.items, id)
	return book, nil
}

func (m *mockStore) insert(_ context.Context, book booklibrary.Book) (booklibrary.Book, error) {
	id := book.ID.Hex()
	m.items[id] = book
	return book, nil
}

// SampleData generates a sample Book objects for testing
func SampleData() []booklibrary.Book {
	id1, _ := primitive.ObjectIDFromHex("000000000000000000000001")
	id2, _ := primitive.ObjectIDFromHex("000000000000000000000002")
	id3, _ := primitive.ObjectIDFromHex("000000000000000000000003")
	bb := []booklibrary.Book{
		{
			ID:          id1,
			Author:      "Jörg Jooss",
			Title:       "Go in 24 Minutes",
			ReleaseDate: time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC),
			Keywords:    []booklibrary.Keyword{{Value: "Go"}},
		},
		{
			ID:          id2,
			Author:      "Jonah Jooss",
			Title:       "Dragons Unleashed",
			ReleaseDate: time.Date(2021, 11, 15, 0, 0, 0, 0, time.UTC),
			Keywords:    []booklibrary.Keyword{{Value: "Dragons"}, {Value: "Toys"}},
		},
		{
			ID:          id3,
			Author:      "Paul Jooss",
			Title:       "Professional BattleTech",
			ReleaseDate: time.Date(2022, 7, 30, 0, 0, 0, 0, time.UTC),
			Keywords:    []booklibrary.Keyword{{Value: "SciFi"}, {Value: "Boardgames"}},
		}}
	return bb
}
