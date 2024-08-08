package problem

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonRenderer(t *testing.T) {
	t.Run("should render as json if problem struct is empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		problem := Problem{}
		problem.JSON(w)
		resp := w.Result()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/problem+json", resp.Header.Get("Content-Type"))
		assert.Empty(t, resp.Header.Get("Cache-Control"))
		assert.Equal(t, `{}`, w.Body.String())
	})

	t.Run("should render as json response", func(t *testing.T) {
		w := httptest.NewRecorder()
		problem := Problem{
			Status:   http.StatusServiceUnavailable,
			Title:    "Service Maintenance",
			Detail:   "API is under maintenance",
			Instance: "/ping",
		}
		problem.WithExtension("version", "1.0.0")
		problem.WithExtension("maintenance", true)

		problem.JSON(w)
		resp := w.Result()

		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		assert.Equal(t, "application/problem+json", resp.Header.Get("Content-Type"))
		assert.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get("Cache-Control"))
		assert.JSONEq(t, `{"title":"Service Maintenance","detail":"API is under maintenance","instance":"/ping","status":503,"version":"1.0.0","maintenance":true}`, w.Body.String())
	})

	t.Run("should create generic error if cannot encode the struct", func(t *testing.T) {
		w := httptest.NewRecorder()
		problem := Problem{}
		problem.WithExtension("bogus", func() {})

		problem.JSON(w)
		resp := w.Result()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		assert.Equal(t, "application/problem+json", resp.Header.Get("Content-Type"))
		assert.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get("Cache-Control"))
		assert.JSONEq(t, `{"detail":"json: error calling MarshalJSON for type *problem.Problem: json: unsupported type: func()", "status":422, "title":"JSON Encoding Error"}`, w.Body.String())
	})
}
