package booklibrary

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"
)

func Routes(crud CrudService) chi.Router {
	rs := Resource{crud: crud}
	r := chi.NewRouter()
	r.Use(middleware.AllowContentType("application/json"))
	r.Method(http.MethodGet, "/", instrument(rs.List, "list_books"))
	r.Method(http.MethodPost, "/", instrument(rs.Create, "create_book"))
	r.Route("/{id}", func(r chi.Router) {
		r.Method(http.MethodGet, "/", instrument(rs.Get, "get_book"))
		r.Method(http.MethodPut, "/", instrument(rs.Update, "update_book"))
		r.Method(http.MethodDelete, "/", instrument(rs.Delete, "delete_book"))
	})
	return r
}

type Resource struct {
	crud CrudService
}

func (rs Resource) List(w http.ResponseWriter, r *http.Request) {
	l := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(l)
	if err != nil || limit < 1 {
		limit = 100
	}
	slog.Debug("limiting results", slog.Int("limit", limit))

	all, err := rs.crud.List(r.Context(), limit)
	if err != nil {
		slog.Error("database access", ErrorKey, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Debug(
			"handler complete",
			slog.Int("status", http.StatusInternalServerError),
			slog.Group("handler",
				slog.String("resource", "Book"),
				slog.String("method", "List")))
		return
	}

	respond(w, all, http.StatusOK)
	slog.Debug(
		"handler complete",
		slog.Int("status", http.StatusOK),
		slog.Group("handler",
			slog.String("resource", "Book"),
			slog.String("method", "List")))
}

func (rs Resource) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	book, err := rs.crud.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrInvalidID) {
			slog.Info("invalid ID", slog.String("id", id))
			http.NotFound(w, r)
			slog.Debug(
				"handler complete",
				slog.Int("status", http.StatusNotFound),
				slog.Group("handler",
					slog.String("resource", "Book"),
					slog.String("method", "Get")))
			return
		}
		if errors.Is(err, ErrNotFound) {
			slog.Info("book not found", slog.String("id", id))
			http.NotFound(w, r)
			slog.Debug(
				"handler complete",
				slog.Int("status", http.StatusNotFound),
				slog.Group("handler",
					slog.String("resource", "Book"),
					slog.String("method", "Get")))
			return
		}
		slog.Error("database access", ErrorKey, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Debug(
			"handler complete",
			slog.Int("status", http.StatusInternalServerError),
			slog.Group("handler",
				slog.String("resource", "Book"),
				slog.String("method", "Get")))
		return
	}

	respond(w, book, http.StatusOK)
	slog.Debug(
		"handler complete",
		slog.Int("status", http.StatusOK),
		slog.Group("handler",
			slog.String("resource", "Book"),
			slog.String("method", "Get")))
}

func (rs Resource) Create(w http.ResponseWriter, r *http.Request) {
	// Unmarshal JSON to domain object
	var book Book
	err := bind(r, &book)
	if err != nil {
		slog.Error("binding request payload", ErrorKey, err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		slog.Debug(
			"handler complete",
			slog.Int("status", http.StatusBadRequest),
			slog.Group("handler",
				slog.String("resource", "Book"),
				slog.String("method", "Create")))

		return
	}

	// Add to storage
	added, err := rs.crud.Add(r.Context(), book)
	if err != nil {
		slog.Error("adding book to database", ErrorKey, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Debug(
			"handler complete",
			slog.Int("status", http.StatusInternalServerError),
			slog.Group("handler",
				slog.String("resource", "Book"),
				slog.String("method", "Create")))
		return
	}

	// Return Created with JSON payload
	path := strings.TrimSuffix(r.URL.String(), "/")
	loc := header{
		name: "Location",
		val:  fmt.Sprintf("%s/%s", path, added.IDStr()),
	}
	respond(w, added, http.StatusCreated, loc)
	slog.Debug(
		"handler complete",
		slog.Int("status", http.StatusCreated),
		slog.Group("handler",
			slog.String("resource", "Book"),
			slog.String("method", "Create")))
}

func (rs Resource) Update(w http.ResponseWriter, r *http.Request) {
	var book Book
	err := bind(r, &book)
	if err != nil {
		slog.Error("binding request payload", ErrorKey, err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		slog.Debug(
			"handler complete",
			slog.Int("status", http.StatusBadRequest),
			slog.Group("handler",
				slog.String("resource", "Book"),
				slog.String("method", "Update")))
		return
	}

	// Updated book by ID in request URI
	id := chi.URLParam(r, "id")
	updated, err := rs.crud.Update(r.Context(), id, book)
	if err != nil {
		if errors.Is(err, ErrInvalidID) || errors.Is(err, ErrNotFound) {
			slog.Info("book not found", slog.String("id", id))
			http.NotFound(w, r)
			slog.Debug(
				"handler complete",
				slog.Int("status", http.StatusNotFound),
				slog.Group("handler",
					slog.String("resource", "Book"),
					slog.String("method", "Update")))
			return
		}
		slog.Error("database access", ErrorKey, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Debug(
			"handler complete",
			slog.Int("status", http.StatusInternalServerError),
			slog.Group("handler",
				slog.String("resource", "Book"),
				slog.String("method", "Update")))
		return
	}

	respond(w, updated, http.StatusOK)
	slog.Debug(
		"handler complete",
		slog.Int("status", http.StatusOK),
		slog.Group("handler",
			slog.String("resource", "Book"),
			slog.String("method", "Update")))
}

func (rs Resource) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := rs.crud.Remove(r.Context(), id); err != nil {
		if errors.Is(err, ErrInvalidID) || errors.Is(err, ErrNotFound) {
			slog.Info("book not found", slog.String("id", id))
			http.NotFound(w, r)
			slog.Debug(
				"handler complete",
				slog.Int("status", http.StatusNotFound),
				slog.Group("handler", slog.String("resource", "Book"), slog.String("method", "Delete")))
			return
		}
		slog.Error("database access", ErrorKey, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	respond(w, nil, http.StatusNoContent)
	slog.Debug(
		"handler complete",
		slog.Int("status", http.StatusNoContent),
		slog.Group("handler", slog.String("resource", "Book"), slog.String("method", "Delete")))
}

type header struct {
	name string
	val  string
}

func respond(w http.ResponseWriter, data any, status int, headers ...header) {
	b, err := json.Marshal(data)
	if err != nil {
		slog.Error("encoding response", ErrorKey, err, slog.Any("data", data))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	for _, h := range headers {
		w.Header().Add(h.name, h.val)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

func bind(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
