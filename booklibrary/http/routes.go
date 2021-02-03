package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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
		all, err := api.store.All(r.Context(), 100)
		if err != nil {
			log.Printf("Error reading from database: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ok(w, all)
	}
}

func (api *APIHandler) getBook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		id := v["id"]
		book, err := api.store.Book(r.Context(), id)
		if err != nil {
			switch err {
			case booklibrary.ErrInvalidID:
				log.Printf("Client provided invalid ID for document: %s\n", id)
				http.NotFound(w, r)
				return
			case booklibrary.ErrNotFound:
				log.Printf("Book with ID %s not found\n", id)
				http.NotFound(w, r)
				return
			default:
				log.Printf("Error reading from database: %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Return OK with JSON payload
		ok(w, book)
	}
}

func (api *APIHandler) addBook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read Book JSON object from HTTP body
		body, err := ioutil.ReadAll(r.Body)
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
		created(w, added, loc)
	}
}

func (api *APIHandler) updateBook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read Book JSON object from HTTP body
		body, err := ioutil.ReadAll(r.Body)
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
			switch err {
			case booklibrary.ErrInvalidID:
				log.Printf("Client provided invalid ID for document: %s\n", id)
				http.NotFound(w, r)
				return
			case booklibrary.ErrNotFound:
				log.Printf("Book with ID %s not found\n", id)
				http.NotFound(w, r)
				return
			default:
				log.Printf("Error reading from database: %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if updated == nil {
			log.Printf("Book with ID %s not found\n", id)
			http.NotFound(w, r)
			return
		}

		// Return OK with JSON payload
		ok(w, updated)
	}
}

func (api *APIHandler) deleteBook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		id := v["id"]
		if _, err := api.store.Remove(r.Context(), id); err != nil {
			switch err {
			case booklibrary.ErrInvalidID:
				log.Printf("Client provided invalid ID for document: %s\n", id)
				http.NotFound(w, r)
				return
			case booklibrary.ErrNotFound:
				log.Printf("Book with ID %s not found\n", id)
				http.NotFound(w, r)
				return
			default:
				log.Printf("Error reading from database: %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func ok(w http.ResponseWriter, v interface{}) {
	content(w, v, http.StatusOK, nil)
}

func created(w http.ResponseWriter, v interface{}, location string) {
	h := map[string]string{"Location": location}
	content(w, v, http.StatusCreated, h)
}

func content(w http.ResponseWriter, v interface{}, status int, headers map[string]string) {
	js, err := json.Marshal(v)
	if err != nil {
		log.Printf("Error marshalling object: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", applicationJSON)
	for h, v := range headers {
		w.Header().Set(h, v)
	}
	w.WriteHeader(status)
	w.Write(js)
}
