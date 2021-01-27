package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	in   []*booklibrary.Book
	want int
}{
	{mock.SampleData(), len(mock.SampleData())},
	{[]*booklibrary.Book{}, 0},
}

func TestGetAllBooks(t *testing.T) {
	for _, tt := range allBooksTest {
		t.Run(fmt.Sprintf("%d books", len(tt.in)), func(t *testing.T) {
			store, _ := mock.NewStorage(tt.in)
			api := NewHandler(store)
			r := httptest.NewRequest(http.MethodGet, "/api/books", nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, r)
			resp := w.Result()

			if got := resp.StatusCode; got != http.StatusOK {
				t.Logf("Received unexpected HTTP status code, got %d, want %d", got, http.StatusOK)
				t.FailNow()
			}

			if got := resp.Header.Get("Content-Type"); got != applicationJSON {
				t.Logf("Received unexpected HTTP content, got %q, want %q", got, applicationJSON)
				t.FailNow()
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Logf("Error reading response body: %v", err)
				t.FailNow()
			}
			var books []*booklibrary.Book
			if err := json.Unmarshal(body, &books); err != nil {
				t.Logf("Error unmarshaling JSON response: %v", err)
				t.FailNow()
			}
			if got := len(books); got != tt.want {
				t.Errorf("Received an unexpected number of items, got %d, want %d", got, tt.want)
			}
		})
	}
}

var getBookTests = []struct {
	in   string
	want int
}{
	{"000000000000000000000000", 404},
	{"000000000000000000000001", 200},
	{"000000000000000000000002", 200},
	{"000000000000000000000003", 200},
	{"000000000000000000000004", 404},
	{"012345678901234567890123", 404},
}

func TestGetBookByID(t *testing.T) {
	for _, tt := range getBookTests {
		t.Run(tt.in, func(t *testing.T) {
			store, _ := mock.NewStorage(mock.SampleData())
			api := NewHandler(store)
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
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Logf("Error reading response body: %v", err)
				t.FailNow()
			}
			var book booklibrary.Book
			if err := json.Unmarshal(body, &book); err != nil {
				t.Logf("Error unmarshaling JSON response: %v", err)
				t.FailNow()
			}
			if book.ID != want {
				t.Errorf("Received unexected Book, got %q, want %q", book.ID, want)
			}
		})
	}
}

var deleteBookTests = []struct {
	in   string
	want int
}{
	{"000000000000000000000000", 404},
	{"000000000000000000000001", 204},
	{"000000000000000000000002", 204},
	{"000000000000000000000003", 204},
	{"000000000000000000000004", 404},
	{"012345678901234567890123", 404},
}

func TestDeleteBook(t *testing.T) {
	for _, tt := range deleteBookTests {
		t.Run(tt.in, func(t *testing.T) {
			store, _ := mock.NewStorage(mock.SampleData())
			api := NewHandler(store)
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
	store, _ := mock.NewStorage(mock.SampleData())
	api := NewHandler(store)
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
		t.FailNow()
	}

	if got := resp.Header.Get("Content-Type"); got != applicationJSON {
		t.Logf("Received unexpected HTTP content, got %q, want %q", got, applicationJSON)
		t.FailNow()
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error reading response body: %v", err)
		t.FailNow()
	}

	book = &booklibrary.Book{}
	if err := json.Unmarshal(body, book); err != nil {
		t.Logf("Error unmarshaling JSON response: %v", err)
		t.FailNow()
	}

	got := resp.Header.Get("Location")
	if got == "" {
		t.Logf("No Location header present in response")
		t.FailNow()
	}

	want := "/api/books/" + book.ID.Hex()
	if got != want {
		t.Logf("Incorrect Location header, want %q, got %q", want, got)
	}
}

var updateBookTests = []struct {
	in   string
	want int
}{
	{"000000000000000000000001", 200},
	{"000000000000000000000004", 404},
}

func TestUpdateBook(t *testing.T) {
	for _, tt := range updateBookTests {
		t.Run(tt.in, func(t *testing.T) {
			store, _ := mock.NewStorage(mock.SampleData())
			api := NewHandler(store)
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
				t.Logf("Error marshaling Book: %v.", err)
				t.FailNow()
			}
			r := httptest.NewRequest(http.MethodPut, "/api/books/"+book.ID.Hex(), bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			api.ServeHTTP(w, r)
			resp := w.Result()

			if got := resp.StatusCode; got != tt.want {
				t.Fatalf("Received unexpected HTTP status code, got %d, want %d", got, tt.want)
				t.FailNow()
			}

			if resp.StatusCode != http.StatusOK {
				// The remainder of this test only apply to HTTP 200 OK
				return
			}

			if got := resp.Header.Get("Content-Type"); got != applicationJSON {
				t.Logf("Received unexpected HTTP content, got %q, want %q", got, applicationJSON)
				t.FailNow()
			}

			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Logf("Error reading response body: %v", err)
				t.FailNow()
			}

			got := &booklibrary.Book{}
			if err := json.Unmarshal(body, got); err != nil {
				t.Logf("Error unmarshaling JSON response: %v", err)
				t.FailNow()
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
