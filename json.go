package problem

import (
    "encoding/json"
    "net/http"
)

// JSON function writes a json response to the client.
// Handles the encoding errors and returns a RFC-9457
// compliant error response.
func (p *Problem) JSON(w http.ResponseWriter) {
    b, err := json.Marshal(p)
    if err != nil {
        p.Status = http.StatusUnprocessableEntity
        b, _ = json.Marshal(Problem{
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
    _, _ = w.Write(b) //nolint:errcheck
}

// MarshalJSON implements json.Marshaler interface to serialize
// the Problem struct into RFC-9457 JSON format.
func (p Problem) MarshalJSON() ([]byte, error) {
    m := map[string]any{}
    if p.Status > 0 {
        m["status"] = p.Status
    }
    if p.Instance != "" {
        m["instance"] = p.Instance
    }
    if p.Detail != "" {
        m["detail"] = p.Detail
    }
    if p.Title != "" {
        m["title"] = p.Title
    }
    if p.Type != "" {
        m["type"] = p.Type
    }
    for k, v := range p.Extensions {
	    if v != nil {
	        m[k] = v
	    }
    }
    return json.Marshal(m)
}

// UnmarshalJSON implements json.Unmarshaler interface to deserialize
// the JSON string into Problem struct as in the RFC-9457 spec.
// The extension values are not converted, but are available.
func (p *Problem) UnmarshalJSON(data []byte) error {
    var (
        v   map[string]any
        err error
    )
    if err = json.Unmarshal(data, &v); err != nil {
        return err
    }
    if status, ok := v["status"]; ok {
        if p.Status, err = p.statusCode(status); err != nil {
            return err
        }
    }
    // Get the spec definitions
    p.Extensions = make(map[string]any)
    p.Instance, _ = v["instance"].(string)
    p.Detail, _ = v["detail"].(string)
    p.Title, _ = v["title"].(string)
    p.Type, _ = v["type"].(string)
    for _, f := range []string{"status", "instance", "detail", "title", "type"} {
        delete(v, f)
    }
    // Set anything else as extension
    for key, val := range v {
        p.Extensions[key] = val
    }
    return nil
}
