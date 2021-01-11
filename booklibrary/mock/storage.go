package mock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joergjo/go-samples/booklibrary"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mockStore abstracts the data access from the underlying data store
type mockStore struct {
	items map[string]*booklibrary.Book
}

// Compile-time check to verify we implement Storage
var _ booklibrary.Storage = &mockStore{}

// NewStorage creates a new Storage instance
func NewStorage() (booklibrary.Storage, error) {
	m := &mockStore{items: make(map[string]*booklibrary.Book)}
	m.sampleData()
	return m, nil
}

// All finds all books
func (m *mockStore) All(_ context.Context, limit int64) ([]*booklibrary.Book, error) {
	all := []*booklibrary.Book{}
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
func (m *mockStore) Book(_ context.Context, id string) (*booklibrary.Book, error) {
	b, ok := m.items[id]
	if !ok {
		msg := fmt.Sprintf("Book with ID %s not found.", id)
		err := errors.New(msg)
		return nil, err
	}
	return b, nil
}

// Add ads a new Book
func (m *mockStore) Add(_ context.Context, book *booklibrary.Book) (*booklibrary.Book, error) {
	oid := primitive.NewObjectID()
	book.ID = oid
	id := oid.Hex()
	m.items[id] = book
	return book, nil
}

// Update updates an existing Book
func (m *mockStore) Update(_ context.Context, id string, book *booklibrary.Book) (*booklibrary.Book, error) {
	if _, ok := m.items[id]; !ok {
		msg := fmt.Sprintf("Adding Book with ID %s failed.", id)
		err := errors.New(msg)
		return nil, err
	}
	var err error
	book.ID, err = primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, booklibrary.ErrInvalidID
	}
	m.items[id] = book
	return book, nil
}

// Remove removes a Book
func (m *mockStore) Remove(_ context.Context, id string) (*booklibrary.Book, error) {
	book, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	delete(m.items, id)
	return book, nil
}

func (m *mockStore) sampleData() {
	bb := []*booklibrary.Book{
		{
			Author:      "Jörg Jooss",
			Title:       "Go in 24 Minutes",
			ReleaseDate: time.Date(2019, 5, 1, 0, 0, 0, 0, time.UTC),
			Keywords:    []booklibrary.Keyword{{Value: "Go"}},
		},
		{
			Author:      "Jonah Jooss",
			Title:       "Dragons Unleashed",
			ReleaseDate: time.Date(2021, 11, 15, 0, 0, 0, 0, time.UTC),
			Keywords:    []booklibrary.Keyword{{Value: "Dragons"}, {Value: "Toys"}},
		},
		{
			Author:      "Paul Jooss",
			Title:       "Professional BattleTech",
			ReleaseDate: time.Date(2022, 7, 30, 0, 0, 0, 0, time.UTC),
			Keywords:    []booklibrary.Keyword{{Value: "SciFi"}, {Value: "Boardgames"}},
		}}
	for _, b := range bb {
		m.Add(nil, b)
	}
}
