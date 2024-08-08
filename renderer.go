package problem

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var cacheableStatuses = []int{
	0,
	http.StatusOK,
	http.StatusNoContent,
	http.StatusPartialContent,
	http.StatusMultipleChoices,
	http.StatusMovedPermanently,
	http.StatusNotFound,
	http.StatusMethodNotAllowed,
	http.StatusGone,
	http.StatusRequestURITooLong,
	http.StatusNotImplemented,
}

// JSON writes a JSON response to the client.
// Handles JSON encoding errors and returns an
// RFC-9457 compliant JSON error object.
func (p *Problem) JSON(w http.ResponseWriter) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(p); err != nil {
		p.Status = http.StatusUnprocessableEntity
		_ = enc.Encode(Problem{
			Status: p.Status,
			Detail: err.Error(),
			Title:  "JSON Encoding Error",
		})
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w = setCacheControl(w, p.Status)
	if p.Status > 0 {
		w.WriteHeader(p.Status)
	}
	// remove the trailing "\n"
	b := buf.Bytes()
	_, _ = w.Write(b[:len(b)-1]) //nolint:errcheck
}

func setCacheControl(w http.ResponseWriter, status int) http.ResponseWriter {
	for _, s := range cacheableStatuses {
		if status == s {
			return w
		}
	}
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	return w
}
