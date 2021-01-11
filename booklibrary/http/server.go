package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/joergjo/go-samples/booklibrary"
)

// Server represents a runnable HTTP application server
type Server struct {
	*mux.Router
	store booklibrary.Storage
}

var _ http.Handler = &Server{}

// NewServer creates a new Server and injects a Storage implementation
func NewServer(store booklibrary.Storage) *Server {
	s := &Server{
		Router: mux.NewRouter(),
		store:  store,
	}
	s.routes()
	return s
}
