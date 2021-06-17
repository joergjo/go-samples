package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/joergjo/go-samples/booklibrary"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mockStore abstracts the data access from the underlying data store
type MockStore struct {
	AllFn    func(ctx context.Context, limit int) ([]booklibrary.Book, error)
	BookFn   func(ctx context.Context, id string) (booklibrary.Book, error)
	AddFn    func(ctx context.Context, book booklibrary.Book) (booklibrary.Book, error)
	UpdateFn func(ctx context.Context, id string, book booklibrary.Book) (booklibrary.Book, error)
	RemoveFn func(ctx context.Context, id string) (booklibrary.Book, error)
}

// Compile-time check to verify we implement Storage
var _ booklibrary.Storage = (*MockStore)(nil)

// NewStorage creates a new Storage instance

// All finds all books
func (m *MockStore) All(ctx context.Context, limit int) ([]booklibrary.Book, error) {
	return m.AllFn(ctx, limit)
}

// Book finds a specific book
func (m *MockStore) Book(ctx context.Context, id string) (booklibrary.Book, error) {
	return m.BookFn(ctx, id)
}

// Add ads a new Book
func (m *MockStore) Add(ctx context.Context, book booklibrary.Book) (booklibrary.Book, error) {
	return m.AddFn(ctx, book)
}

// Update updates an existing Book
func (m *MockStore) Update(ctx context.Context, id string, book booklibrary.Book) (booklibrary.Book, error) {
	return m.UpdateFn(ctx, id, book)
}

func (m *MockStore) Remove(ctx context.Context, id string) (booklibrary.Book, error) {
	return m.RemoveFn(ctx, id)
}

// SampleData generates a sample Book objects for testing
func SampleData(count int) map[string]booklibrary.Book {
	m := make(map[string]booklibrary.Book, count)
	for i := 0; i < count; i++ {
		id := primitive.NewObjectID()
		m[id.Hex()] = booklibrary.Book{
			ID:          id,
			Author:      "John Doe",
			Title:       fmt.Sprintf("Unit Testing, Volume %d", i),
			ReleaseDate: time.Now(),
			Keywords:    []booklibrary.Keyword{{Value: "Go"}, {Value: "Test"}},
		}
	}
	return m
}
