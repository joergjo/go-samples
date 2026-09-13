package webapi

import (
	"encoding/json/v2"
	"net/http"
)

type header struct {
	name string
	val  string
}

func respond(w http.ResponseWriter, v any, status int, headers ...header) {
	for _, h := range headers {
		w.Header().Add(h.name, h.val)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.MarshalWrite(w, v)
}

func bind(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.UnmarshalRead(r.Body, v)
}
