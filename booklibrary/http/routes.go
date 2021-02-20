package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gorilla/mux"
	"github.com/joergjo/go-samples/booklibrary"
)

const applicationJSON = "application/json"

func (api *APIHandler) routes() {
	api.Handle("/api/books", handlerFor(api.allBooks(), "allBooks")).Methods(http.MethodGet)
	api.Handle("/api/books/{id}", handlerFor(api.getBook(), "getBook")).Methods(http.MethodGet)
	api.Handle("/api/books", handlerFor(api.addBook(), "addBook")).Methods(http.MethodPost)
	api.Handle("/api/books/{id}", handlerFor(api.updateBook(), "updateBook")).Methods(http.MethodPut)
	api.Handle("/api/books/{id}", handlerFor(api.deleteBook(), "deleteBook")).Methods(http.MethodDelete)
	api.Handle("/metrics", promhttp.Handler())
}

func handlerFor(handlerFunc http.HandlerFunc, handlerName string) http.Handler {
	return promhttp.InstrumentHandlerInFlight(inFlightGauge,
		promhttp.InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": handlerName}),
			promhttp.InstrumentHandlerCounter(counter,
				promhttp.InstrumentHandlerResponseSize(responseSize, handlerFunc),
			),
		),
	)
}

func (api *APIHandler) allBooks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := r.URL.Query().Get("limit")
		limit, err := strconv.Atoi(l)
		if err != nil || limit < 1 {
			limit = 100
		}
		log.Printf("Limiting: result to %d entries\n", limit)

		all, err := api.store.All(r.Context(), int64(limit))
		if err != nil {
			log.Printf("Database error: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respond(w, all, http.StatusOK, nil)
	}
}

func (api *APIHandler) getBook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		id := v["id"]
		book, err := api.store.Book(r.Context(), id)
		if err != nil {
			if errors.Is(err, booklibrary.ErrInvalidID) {
				log.Printf("Client provided invalid ID for document: %s\n", id)
				http.NotFound(w, r)
				return
			}
			if errors.Is(err, booklibrary.ErrNotFound) {
				log.Printf("Book with ID %s not found\n", id)
				http.NotFound(w, r)
				return
			}
			log.Printf("Database error: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respond(w, book, http.StatusOK, nil)
	}
}

func (api *APIHandler) addBook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read Book JSON object from HTTP body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON to domain object
		var book booklibrary.Book
		err = json.Unmarshal(body, &book)
		if err != nil {
			log.Printf("Error unmarshalling book: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Add to storage
		added, err := api.store.Add(r.Context(), &book)
		if err != nil {
			log.Printf("Error adding book to database: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return Created with JSON payload
		loc := fmt.Sprintf("%s/%s", r.URL.String(), added.ID)
		h := map[string]string{"Location": loc}
		respond(w, added, http.StatusCreated, h)
	}
}

func (api *APIHandler) updateBook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read Book JSON object from HTTP body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON to domain object
		var book booklibrary.Book
		err = json.Unmarshal(body, &book)
		if err != nil {
			log.Printf("Error unmarshalling book: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Updated book by ID in request URI
		v := mux.Vars(r)
		id := v["id"]
		updated, err := api.store.Update(r.Context(), id, &book)
		if err != nil {
			if errors.Is(err, booklibrary.ErrInvalidID) || errors.Is(err, booklibrary.ErrNotFound) {
				log.Printf("Book with ID %s not found\n", id)
				http.NotFound(w, r)
				return
			}
			log.Printf("Database error: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respond(w, updated, http.StatusOK, nil)
	}
}

func (api *APIHandler) deleteBook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		id := v["id"]
		if _, err := api.store.Remove(r.Context(), id); err != nil {
			if errors.Is(err, booklibrary.ErrInvalidID) || errors.Is(err, booklibrary.ErrNotFound) {
				log.Printf("Book with ID %s not found\n", id)
				http.NotFound(w, r)
				return
			}
			log.Printf("Database error: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respond(w, nil, http.StatusNoContent, nil)
	}
}

func respond(w http.ResponseWriter, obj interface{}, status int, headers map[string]string) {
	var content []byte
	if obj != nil {
		var err error
		content, err = json.Marshal(obj)
		if err != nil {
			log.Printf("Error marshalling object: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	for h, v := range headers {
		w.Header().Set(h, v)
	}
	if len(content) == 0 {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", applicationJSON)
	w.WriteHeader(status)
	w.Write([]byte(content))
}
