package booklibrary

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Book represent a book in our library
type Book struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Author      string             `json:"author" bson:"author"`
	Title       string             `json:"title" bson:"title"`
	ReleaseDate time.Time          `json:"releaseDate" bson:"releaseDate"`
	Keywords    []Keyword          `json:"keywords" bson:"keywords"`
}

// Keyword represent a certain topic a book covers
type Keyword struct {
	Value string `json:"keyword" bson:"keyword"`
}

// MarshalJSON serializes a Book with its ReleaseDate rendered as Unix time
func (b Book) MarshalJSON() ([]byte, error) {
	type Dto Book
	return json.Marshal(struct {
		ReleaseDate int64 `json:"releaseDate"`
		Dto
	}{
		ReleaseDate: b.ReleaseDate.Unix(),
		Dto:         (Dto)(b),
	})
}

// UnmarshalJSON deserializes a Book with its ReleaseDate rendered as Unix time
func (b *Book) UnmarshalJSON(data []byte) error {
	type Dto Book
	dto := struct {
		ReleaseDate int64 `json:"releaseDate"`
		*Dto
	}{
		Dto: (*Dto)(b),
	}
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}
	b.ReleaseDate = time.Unix(dto.ReleaseDate, 0)
	return nil
}

func (kw Keyword) String() string {
	return kw.Value
}
