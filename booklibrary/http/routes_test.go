package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/joergjo/go-samples/booklibrary"
	api "github.com/joergjo/go-samples/booklibrary/http"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const applicationJSON = "application/json"

type storeStub struct {
	AllFn    func(ctx context.Context, limit int) ([]booklibrary.Book, error)
	GetFn    func(ctx context.Context, id string) (booklibrary.Book, error)
	AddFn    func(ctx context.Context, book booklibrary.Book) (booklibrary.Book, error)
	UpdateFn func(ctx context.Context, id string, book booklibrary.Book) (booklibrary.Book, error)
	RemoveFn func(ctx context.Context, id string) (booklibrary.Book, error)
}

// Compile-time check to verify we implement Storage
var _ booklibrary.Store = (*storeStub)(nil)

// NewStorage creates a new Storage instance

// All finds all books
func (m *storeStub) All(ctx context.Context, limit int) ([]booklibrary.Book, error) {
	return m.AllFn(ctx, limit)
}

// Book finds a specific book
func (m *storeStub) Get(ctx context.Context, id string) (booklibrary.Book, error) {
	return m.GetFn(ctx, id)
}

// Add ads a new Book
func (m *storeStub) Add(ctx context.Context, book booklibrary.Book) (booklibrary.Book, error) {
	return m.AddFn(ctx, book)
}

// Update updates an existing Book
func (m *storeStub) Update(ctx context.Context, id string, book booklibrary.Book) (booklibrary.Book, error) {
	return m.UpdateFn(ctx, id, book)
}

func (m *storeStub) Remove(ctx context.Context, id string) (booklibrary.Book, error) {
	return m.RemoveFn(ctx, id)
}

func testData(count int) map[string]booklibrary.Book {
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

func TestGetAllBooks(t *testing.T) {
	tests := []struct {
		name  string
		in    map[string]booklibrary.Book
		limit int
		want  int
	}{
		{"get_multiple_books", testData(5), -1, 5},
		{"get_first_book", testData(5), 1, 1},
		{"get_multiple_books_with_limit", testData(10), 50, 10},
		{"get_empty_collection", make(map[string]booklibrary.Book), -1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &storeStub{}
			store.AllFn = func(_ context.Context, _ int) ([]booklibrary.Book, error) {
				i := 0
				bb := make([]booklibrary.Book, tt.want)
				for _, b := range tt.in {
					if i == tt.want {
						break
					}
					bb[i] = b
					i++
				}
				return bb, nil
			}
			api := api.NewAPIHandler(store)
			path := "/api/books"
			if tt.limit != -1 {
				path = fmt.Sprintf("%s?limit=%d", path, tt.limit)
			}
			r := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, r)
			resp := w.Result()

			if got := resp.StatusCode; got != http.StatusOK {
				t.Fatalf("Received unexpected HTTP status code, got %d, want %d", got, http.StatusOK)
			}

			if got := resp.Header.Get("Content-Type"); got != applicationJSON {
				t.Fatalf("Received unexpected HTTP content, got %q, want %q", got, applicationJSON)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Error reading response body: %v", err)
			}

			var books []*booklibrary.Book
			if err := json.Unmarshal(body, &books); err != nil {
				t.Fatalf("Error unmarshaling JSON response: %v", err)
			}

			if got := len(books); got > tt.want {
				t.Fatalf("Received an unexpected number of items, got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetBookByID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"get_by_id", "000000000000000000000001", 200},
		{"get_unknown_id", "000000000000000000000004", 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &storeStub{}
			id, err := primitive.ObjectIDFromHex("000000000000000000000001")
			if err != nil {
				t.Fatalf("Error creating ObjectID")
			}
			b := booklibrary.Book{
				ID: id,
			}
			store.GetFn = func(_ context.Context, id string) (booklibrary.Book, error) {
				if id != string(b.ID.Hex()) {
					return booklibrary.Book{}, booklibrary.ErrNotFound
				}
				return b, nil
			}

			api := api.NewAPIHandler(store)
			r := httptest.NewRequest(http.MethodGet, "/api/books/"+tt.in, nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, r)
			resp := w.Result()

			got := resp.StatusCode
			if got != tt.want {
				t.Fatalf("Received unexpected HTTP status code, got %d, want %d", got, tt.want)
			}

			if got != http.StatusOK {
				// The remainder of this test only apply to HTTP 200 OK
				return
			}

			want, _ := primitive.ObjectIDFromHex(tt.in)
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Error reading response body: %v", err)
			}

			var book booklibrary.Book
			if err := json.Unmarshal(body, &book); err != nil {
				t.Fatalf("Error unmarshaling JSON response: %v", err)
			}

			if book.ID != want {
				t.Fatalf("Received unexected Book, got %q, want %q", book.ID, want)
			}
		})
	}
}

func TestDeleteBook(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"delete_by_id", "000000000000000000000001", 204},
		{"delete_unknown_id", "000000000000000000000004", 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &storeStub{}
			id, err := primitive.ObjectIDFromHex("000000000000000000000001")
			if err != nil {
				t.Fatalf("Error creating ObjectID")
			}
			b := booklibrary.Book{
				ID: id,
			}
			store.RemoveFn = func(_ context.Context, id string) (booklibrary.Book, error) {
				if id != string(b.ID.Hex()) {
					return booklibrary.Book{}, booklibrary.ErrNotFound
				}
				return b, nil
			}

			api := api.NewAPIHandler(store)
			r := httptest.NewRequest(http.MethodDelete, "/api/books/"+tt.in, nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, r)

			if got := w.Result().StatusCode; got != tt.want {
				t.Fatalf("Received unexpected HTTP status code, got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestAddBook(t *testing.T) {
	store := &storeStub{}
	store.AddFn = func(_ context.Context, book booklibrary.Book) (booklibrary.Book, error) {
		book.ID = primitive.NewObjectID()
		return book, nil
	}

	api := api.NewAPIHandler(store)
	book := &booklibrary.Book{
		Author:      "Jörg Jooss",
		Title:       "Go Testing in Action",
		ReleaseDate: time.Now(),
		Keywords:    []booklibrary.Keyword{{Value: "Golang"}, {Value: "Testing"}},
	}
	body, err := json.Marshal(book)
	if err != nil {
		t.Logf("Error marshaling Book: %v", err)
		t.FailNow()
	}
	r := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	api.ServeHTTP(w, r)
	resp := w.Result()

	if got := resp.StatusCode; got != http.StatusCreated {
		t.Fatalf("Received unexpected HTTP status code, got %d, want %d", got, http.StatusCreated)
	}

	if got := resp.Header.Get("Content-Type"); got != applicationJSON {
		t.Fatalf("Received unexpected HTTP content, got %q, want %q", got, applicationJSON)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %v", err)
	}

	book = &booklibrary.Book{}
	if err := json.Unmarshal(body, book); err != nil {
		t.Fatalf("Error unmarshaling JSON response: %v", err)
	}

	got := resp.Header.Get("Location")
	if got == "" {
		t.Fatalf("No Location header present in response")
	}

	want := "/api/books/" + book.ID.Hex()
	if got != want {
		t.Fatalf("Incorrect Location header, want %q, got %q", want, got)
	}
}

func TestUpdateBook(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"update_by_id", "000000000000000000000003", 200},
		{"update_invalid_id", "000000000000000000000004", 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &storeStub{}
			store.UpdateFn = func(_ context.Context, id string, book booklibrary.Book) (booklibrary.Book, error) {
				if id != "000000000000000000000003" {
					return booklibrary.Book{}, booklibrary.ErrNotFound
				}
				return book, nil
			}

			id, err := primitive.ObjectIDFromHex(tt.in)
			if err != nil {
				t.Fatalf("Error creating ObjectID")
			}

			book := booklibrary.Book{
				ID:          id,
				Author:      "Jörg Jooss",
				Title:       "Go Testing in 24 Minutes",
				ReleaseDate: time.Date(2021, 1, 10, 0, 0, 0, 0, time.UTC),
				Keywords:    []booklibrary.Keyword{{Value: "Golang"}, {Value: "Testing"}},
			}
			body, err := json.Marshal(book)
			if err != nil {
				t.Fatalf("Error marshaling Book: %v.", err)
			}

			api := api.NewAPIHandler(store)
			r := httptest.NewRequest(http.MethodPut, "/api/books/"+book.ID.Hex(), bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			api.ServeHTTP(w, r)
			resp := w.Result()

			if got := resp.StatusCode; got != tt.want {
				t.Fatalf("Received unexpected HTTP status code, got %d, want %d", got, tt.want)
			}

			if resp.StatusCode != http.StatusOK {
				// The remainder of this test only apply to HTTP 200 OK
				return
			}

			if got := resp.Header.Get("Content-Type"); got != applicationJSON {
				t.Fatalf("Received unexpected HTTP content, got %q, want %q", got, applicationJSON)
			}

			body, err = io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Error reading response body: %v", err)
			}

			got := &booklibrary.Book{}
			if err := json.Unmarshal(body, got); err != nil {
				t.Fatalf("Error unmarshaling JSON response: %v", err)
			}
			// cmp.Diff() considers value and pointer types to be different,
			// so we pass a pointer to the expected value
			if diff := cmp.Diff(&book, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
