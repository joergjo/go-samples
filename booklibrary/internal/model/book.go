package model

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"time"
)

// Book represent a book in the library.
type Book struct {
	ID          string    `json:"_id" bson:"_id,omitempty"`
	Author      string    `json:"author" bson:"author"`
	Title       string    `json:"title" bson:"title"`
	ReleaseDate time.Time `json:"releaseDate" bson:"releaseDate"`
	Keywords    []Keyword `json:"keywords" bson:"keywords"`
}

// MarhsalJSONTo deserializes a Book with its ReleaseDate rendered as Unix time using streaming.
func (b Book) MarshalJSONTo(enc *jsontext.Encoder) error {
	type Dto Book
	dto := struct {
		ReleaseDate int64 `json:"releaseDate"`
		Dto
	}{
		ReleaseDate: b.ReleaseDate.Unix(),
		Dto:         (Dto)(b),
	}
	return json.MarshalEncode(enc, dto)
}

// UnmarshalJSONFrom deserializes a Book with its ReleaseDate rendered as Unix time using streaming.
func (b *Book) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	type Dto Book
	dto := struct {
		ReleaseDate int64 `json:"releaseDate"`
		*Dto
	}{
		Dto: (*Dto)(b),
	}
	if err := json.UnmarshalDecode(dec, &dto); err != nil {
		return err
	}
	b.ReleaseDate = time.Unix(dto.ReleaseDate, 0)
	return nil
}

// Keyword represents a book's topic.
type Keyword struct {
	Value string `json:"keyword" bson:"keyword"`
}

// String returns the keyword value.
func (kw Keyword) String() string {
	return kw.Value
}
