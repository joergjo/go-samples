package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestMarshalJSON(t *testing.T) {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		t.Fatalf("Fatal error loading UTC location: %v\n", err)
	}
	tests := []struct {
		name string
		want string
		in   Book
	}{
		{
			name: "book_with_single_tag",
			in: Book{
				ID:          "000000000000000000000001",
				Author:      "John Doe",
				Title:       "Unit Testing in Go",
				ReleaseDate: time.Date(2020, time.February, 1, 11, 0, 0, 0, loc),
				Keywords:    []Keyword{{Value: "Golang"}},
			},
			want: "{\"releaseDate\":1580554800,\"_id\":\"000000000000000000000001\",\"author\":\"John Doe\",\"title\":\"Unit Testing in Go\",\"keywords\":[{\"keyword\":\"Golang\"}]}",
		},
		{
			name: "book_with_multiple_tags",
			in: Book{
				ID:          "000000000000000000000001",
				Author:      "John Doe",
				Title:       "Unit Testing in Go and Python",
				ReleaseDate: time.Date(2020, time.February, 1, 11, 0, 0, 0, loc),
				Keywords:    []Keyword{{Value: "Golang"}, {Value: "Python"}},
			},
			want: "{\"releaseDate\":1580554800,\"_id\":\"000000000000000000000001\",\"author\":\"John Doe\",\"title\":\"Unit Testing in Go and Python\",\"keywords\":[{\"keyword\":\"Golang\"},{\"keyword\":\"Python\"}]}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Fatal error marshalling to JSON: %v\n", err)
			}
			got := string(b)
			if got != tt.want {
				t.Errorf("JSON does match expected result, got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		t.Fatalf("Fatal error loading UTC location: %v\n", err)
	}
	tests := []struct {
		name string
		in   string
		want Book
	}{
		{
			name: "book_with_single_tag",
			in:   "{\"releaseDate\":1580554800,\"_id\":\"000000000000000000000001\",\"author\":\"John Doe\",\"title\":\"Unit Testing in Go\",\"keywords\":[{\"keyword\":\"Golang\"}]}",
			want: Book{
				ID:          "000000000000000000000001",
				Author:      "John Doe",
				Title:       "Unit Testing in Go",
				ReleaseDate: time.Date(2020, time.February, 1, 11, 0, 0, 0, loc),
				Keywords:    []Keyword{{Value: "Golang"}},
			},
		},
		{
			name: "book_with_multiple_tags",
			in:   "{\"releaseDate\":1580554800,\"_id\":\"000000000000000000000001\",\"author\":\"John Doe\",\"title\":\"Unit Testing in Go and Python\",\"keywords\":[{\"keyword\":\"Golang\"},{\"keyword\":\"Python\"}]}",
			want: Book{
				ID:          "000000000000000000000001",
				Author:      "John Doe",
				Title:       "Unit Testing in Go and Python",
				ReleaseDate: time.Date(2020, time.February, 1, 11, 0, 0, 0, loc),
				Keywords:    []Keyword{{Value: "Golang"}, {Value: "Python"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := []byte(tt.in)
			var got Book
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatalf("Fatal error unmarshalling from JSON: %v\n", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
