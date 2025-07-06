package webapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joergjo/go-samples/booklibrary/internal/model"
	"github.com/joergjo/go-samples/booklibrary/internal/webapi"
)

func TestSystemEndpoints(t *testing.T) {
	tests := []struct {
		name string
		path string
		want int
	}{
		{
			name: "get_metrics",
			path: "/metrics",
			want: http.StatusOK,
		},
		{
			name: "get_liveness",
			path: "/healthz/live",
			want: http.StatusOK,
		},
		{
			name: "get_readiness",
			path: "/healthz/ready",
			want: http.StatusOK,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			crud := crudStub{}
			crud.GetFn = func(_ context.Context, id string) (model.Book, error) {
				return model.Book{}, nil
			}
			router := webapi.NewMux(&crud)
			ts := httptest.NewServer(router)
			defer ts.Close()

			r, err := http.NewRequest(http.MethodGet, ts.URL+tc.path, nil)
			if err != nil {
				t.Fatalf("Error creating request: %v", err)
			}
			res, err := ts.Client().Do(r)
			if err != nil {
				t.Fatalf("Error sending request: %v", err)
			}
			if res.StatusCode != tc.want {
				t.Errorf("Received unexpected HTTP status code, got %d, want %d", res.StatusCode, http.StatusOK)
			}
		})
	}
}
