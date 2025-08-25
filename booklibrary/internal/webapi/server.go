package webapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/joergjo/go-samples/booklibrary/internal/model"
)

// NewServer creates a new HTTP server with the given handler and port.
func NewServer(crud model.CrudService, port int) *http.Server {
	mux := NewMux(crud)
	s := http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
	return &s
}
