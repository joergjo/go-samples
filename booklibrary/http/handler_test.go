package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/joergjo/go-samples/booklibrary"
	"github.com/joergjo/go-samples/booklibrary/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var allBooksTest = []struct {
	name  string
	in    []booklibrary.Book
	limit int
	want  int
}{
	{"get_multiple_books", mock.SampleData(), -1, len(mock.SampleData())},
	{"get_first_book", mock.SampleData(), 1, 1},
	{"get_multiple_books_with_limit", mock.SampleData(), 50, len(mock.SampleData())},
	{"get_empty_collection", []booklibrary.Book{}, -1, 0},
}

func TestGetAllBooks(t *testing.T) {
	for _, tt := range allBooksTest {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := mock.NewStorage(tt.in)
			api := NewAPIHandler(store)
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
				t.Errorf("Received an unexpected number of items, got %d, want %d", got, tt.want)
			}
		})
	}
}

var getBookTests = []struct {
	name string
	in   string
	want int
}{
	{"get_by_id", "000000000000000000000001", 200},
	{"get_unknown_id", "000000000000000000000004", 404},
}

func TestGetBookByID(t *testing.T) {
	for _, tt := range getBookTests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := mock.NewStorage(mock.SampleData())
			api := NewAPIHandler(store)
			r := httptest.NewRequest(http.MethodGet, "/api/books/"+tt.in, nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, r)
			resp := w.Result()

			got := resp.StatusCode
			if got != tt.want {
				t.Errorf("Received unexpected HTTP status code, got %d, want %d", got, tt.want)
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
				t.Errorf("Received unexected Book, got %q, want %q", book.ID, want)
			}
		})
	}
}

var deleteBookTests = []struct {
	name string
	in   string
	want int
}{
	{"delete_by_id", "000000000000000000000002", 204},
	{"delete_unknown_id", "000000000000000000000004", 404},
}

func TestDeleteBook(t *testing.T) {
	for _, tt := range deleteBookTests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := mock.NewStorage(mock.SampleData())
			api := NewAPIHandler(store)
			r := httptest.NewRequest(http.MethodDelete, "/api/books/"+tt.in, nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, r)

			if got := w.Result().StatusCode; got != tt.want {
				t.Errorf("Received unexpected HTTP status code, got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestAddBook(t *testing.T) {
	store, _ := mock.NewStorage(mock.SampleData())
	api := NewAPIHandler(store)
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
		t.Errorf("Incorrect Location header, want %q, got %q", want, got)
	}
}

var updateBookTests = []struct {
	name string
	in   string
	want int
}{
	{"update_by_id", "000000000000000000000003", 200},
	{"update_invalid_id", "000000000000000000000004", 404},
}

func TestUpdateBook(t *testing.T) {
	for _, tt := range updateBookTests {
		t.Run(tt.name, func(t *testing.T) {
			store, _ := mock.NewStorage(mock.SampleData())
			api := NewAPIHandler(store)
			id, _ := primitive.ObjectIDFromHex(tt.in)
			book := &booklibrary.Book{
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

			if got.ID != book.ID {
				t.Errorf("Updated book has wrong ID, got %q, want %q", got.ID, book.ID)
			}
			if got.Author != book.Author {
				t.Errorf("Updated book has wrong author, got %q, want %q", got.Author, book.Author)
			}
			if got.Title != book.Title {
				t.Errorf("Updated book has wrong title, got %q, want %q", got.Title, book.Title)
			}
			if got.ReleaseDate.UTC() != book.ReleaseDate.UTC() {
				t.Errorf("Updated book has wrong release date, got %v, want %v", got.ReleaseDate.UTC(), book.ReleaseDate.UTC())
			}

			gotKW := []string{}
			for _, kw := range got.Keywords {
				gotKW = append(gotKW, kw.Value)
			}
			sort.Strings(gotKW)

			wantKW := []string{}
			for _, kw := range got.Keywords {
				wantKW = append(wantKW, kw.Value)
			}
			sort.Strings(wantKW)

			gotStr := strings.Join(gotKW, " ")
			wantStr := strings.Join(wantKW, " ")

			if gotStr != wantStr {
				t.Errorf("Updated book has wrong set of keywords, got %q, want %q", gotStr, wantStr)
			}
		})
	}
}
