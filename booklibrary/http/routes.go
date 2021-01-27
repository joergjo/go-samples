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
	return func(writer http.ResponseWriter, request *http.Request) {
		all, err := api.store.All(request.Context(), 100)
		if err != nil {
			log.Printf("Error reading from database: %v\n", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(all)
		if err != nil {
			log.Printf("Error marshalling book: %v\n", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", applicationJSON)
		writer.Write(b)
	}
}

func (api *APIHandler) getBook() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		v := mux.Vars(request)
		id := v["id"]
		book, err := api.store.Book(request.Context(), id)
		if err != nil {
			switch err {
			case booklibrary.ErrInvalidID:
				log.Printf("Client provided invalid ID for document: %s\n", id)
				http.NotFound(writer, request)
				return
			case booklibrary.ErrNotFound:
				log.Printf("Book with ID %s not found\n", id)
				http.NotFound(writer, request)
				return
			default:
				log.Printf("Error reading from database: %v\n", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		b, err := json.Marshal(book)
		if err != nil {
			log.Printf("Error marshalling book: %v\n", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", applicationJSON)
		writer.Write(b)
	}
}

func (api *APIHandler) addBook() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Read Book JSON object from HTTP body
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			log.Printf("Error reading request body: %v\n", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON to domain object
		var book booklibrary.Book
		err = json.Unmarshal(body, &book)
		if err != nil {
			log.Printf("Error unmarshalling book: %v\n", err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		// Add to storage
		_, err = api.store.Add(request.Context(), &book)
		if err != nil {
			log.Printf("Error adding book to database: %v\n", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Marshal book back as JSON
		b, err := json.Marshal(&book)
		if err != nil {
			log.Printf("Error marshalling book: %v\n", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write response
		loc := fmt.Sprintf("%s/%s", request.URL.String(), book.ID)
		writer.Header().Set("Location", loc)
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		writer.Write(b)
	}
}

func (api *APIHandler) updateBook() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Read Book JSON object from HTTP body
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			log.Printf("Error reading request body: %v\n", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON to domain object
		var book booklibrary.Book
		err = json.Unmarshal(body, &book)
		if err != nil {
			log.Printf("Error unmarshalling book: %v\n", err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		// Updated book by ID in request URI
		v := mux.Vars(request)
		id := v["id"]
		updatedBook, err := api.store.Update(request.Context(), id, &book)
		if err != nil {
			switch err {
			case booklibrary.ErrInvalidID:
				log.Printf("Client provided invalid ID for document: %s\n", id)
				http.NotFound(writer, request)
				return
			case booklibrary.ErrNotFound:
				log.Printf("Book with ID %s not found\n", id)
				http.NotFound(writer, request)
				return
			default:
				log.Printf("Error reading from database: %v\n", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if updatedBook == nil {
			log.Printf("Book with ID %s not found\n", id)
			http.NotFound(writer, request)
			return
		}

		// Marshal book back as JSON
		b, err := json.Marshal(updatedBook)
		if err != nil {
			log.Printf("Error marshalling book: %v\n", err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", applicationJSON)
		writer.Write(b)
	}
}

func (api *APIHandler) deleteBook() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		v := mux.Vars(request)
		id := v["id"]
		if _, err := api.store.Remove(request.Context(), id); err != nil {
			switch err {
			case booklibrary.ErrInvalidID:
				log.Printf("Client provided invalid ID for document: %s\n", id)
				http.NotFound(writer, request)
				return
			case booklibrary.ErrNotFound:
				log.Printf("Book with ID %s not found\n", id)
				http.NotFound(writer, request)
				return
			default:
				log.Printf("Error reading from database: %v\n", err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		writer.WriteHeader(http.StatusNoContent)
	}
}
