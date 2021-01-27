package http

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/joergjo/go-samples/booklibrary"
)

// APIHandler represents a runnable HTTP application server
type APIHandler struct {
	*mux.Router
	store booklibrary.Storage
}

var _ http.Handler = &APIHandler{}

// NewHandler creates a new Server and injects a Storage implementation
func NewHandler(store booklibrary.Storage) *APIHandler {
	api := &APIHandler{
		Router: mux.NewRouter(),
		store:  store,
	}
	api.routes()
	return api
}
