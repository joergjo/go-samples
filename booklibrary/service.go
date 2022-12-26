package booklibrary

import (
	"context"
	"errors"
)

var (
	ErrInvalidID = errors.New("string is not valid book ID")
	ErrNotFound  = errors.New("book not found")
)

type CrudService interface {
	List(ctx context.Context, limit int) ([]Book, error)
	Get(ctx context.Context, id string) (Book, error)
	Add(ctx context.Context, book Book) (Book, error)
	Update(ctx context.Context, id string, book Book) (Book, error)
	Remove(ctx context.Context, id string) (Book, error)
	Ping(ctx context.Context) error
}
