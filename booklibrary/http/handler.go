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

// NewAPIHandler creates a new Server and injects a Storage implementation
func NewAPIHandler(store booklibrary.Storage) *APIHandler {
	api := &APIHandler{
		Router: mux.NewRouter(),
		store:  store,
	}
	api.Router.StrictSlash(true)
	api.routes()
	return api
}
