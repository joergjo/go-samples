package webapi

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/joergjo/go-samples/booklibrary/internal/model"
)

// NewServer creates a new HTTP server with the given handler and port.
func NewServer(crud model.CrudService, port int) *http.Server {
	mux := NewMux(crud)
	addr := net.JoinHostPort("", strconv.Itoa(port))
	s := http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
	return &s
}
