package booklibrary

import (
	"context"
	"errors"
)

// ErrInvalidID represents an invalid book ID
var (
	ErrInvalidID = errors.New("string is not valid book ID")
	ErrNotFound  = errors.New("book not found")
)

// Storage provides access to manage Book instances
type Storage interface {
	All(context context.Context, limit int) ([]Book, error)
	Book(context context.Context, id string) (Book, error)
	Add(context context.Context, book Book) (Book, error)
	Update(context context.Context, id string, book Book) (Book, error)
	Remove(context context.Context, id string) (Book, error)
}
