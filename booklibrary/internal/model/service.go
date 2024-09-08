package model

import (
	"context"
	"errors"
)

var (
	// ErrInvalidID is returned when an ID is not in a valid form.
	ErrInvalidID = errors.New("string is not valid book ID")
	// ErrNotFound is returned when a book is not found.
	ErrNotFound = errors.New("book not found")
)

// CrudService is the interface for all book library data stores.
// operations
type CrudService interface {
	List(ctx context.Context, limit int) ([]Book, error)
	Get(ctx context.Context, id string) (Book, error)
	Add(ctx context.Context, book Book) (Book, error)
	Update(ctx context.Context, id string, book Book) (Book, error)
	Remove(ctx context.Context, id string) (Book, error)
	Ping(ctx context.Context) error
}
