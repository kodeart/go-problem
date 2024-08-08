package problem

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

// Problem is a struct that represents a problem instance
// as defined in RFC 9457 (https://tools.ietf.org/html/rfc9457).
// All fields are optional.
type Problem struct {
	Status     int            `json:"status,omitempty"`
	Instance   string         `json:"instance,omitempty"`
	Detail     string         `json:"detail,omitempty"`
	Title      string         `json:"title,omitempty"`
	Type       string         `json:"type,omitempty"`
	Extensions map[string]any `json:"-"`
}

func New() *Problem {
	return new(Problem)
}

func (p *Problem) WithStatus(v int) *Problem {
	p.Status = v
	return p
}

func (p *Problem) WithInstance(v string) *Problem {
	p.Instance = v
	return p
}

func (p *Problem) WithDetail(v string) *Problem {
	p.Detail = v
	return p
}

func (p *Problem) WithTitle(v string) *Problem {
	p.Title = v
	return p
}

func (p *Problem) WithType(v string) *Problem {
	p.Type = v
	return p
}

// WithExtension adds key:value pairs to internal Extensions map.
// When JSON serialization is performed, these pairs are
// included in the JSON response as key:value to the final response.
func (p *Problem) WithExtension(key string, val any) *Problem {
	if p.Extensions == nil {
		p.Extensions = make(map[string]any)
	}
	p.Extensions[key] = val
	return p
}

// WithoutExtension removes a key from internal Extensions map.
// If the key does not exist, it does nothing.
func (p *Problem) WithoutExtension(key string) *Problem {
	delete(p.Extensions, key)
	return p
}

// MarshalJSON implements json.Marshaler interface to serialize
// the Problem instance into RFC-9457 JSON format.
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
		m[k] = v
	}
	return json.Marshal(m)
}

// UnmarshalJSON implements json.Unmarshaler interface to unserialize
// the JSON string into Problem structure as RFC-9457 implementation.
// [IMPORTANT]: built-in json.Unmarshaler converts numeric values
// to float64, so we need to convert status code back to int.
// The extension values are not converted, but are available.
func (p *Problem) UnmarshalJSON(data []byte) error {
	var (
		m   map[string]any
		err error
	)

	if err = json.Unmarshal(data, &m); err != nil {
		return err
	}
	if status, ok := m["status"]; ok {
		if p.Status, err = p.toInt(status); err != nil {
			return err
		}
	}
	// get the values, cleanup and set the extensions (if any)
	p.Extensions = make(map[string]any)
	p.Instance, _ = m["instance"].(string)
	p.Detail, _ = m["detail"].(string)
	p.Title, _ = m["title"].(string)
	p.Type, _ = m["type"].(string)
	for _, f := range []string{"status", "instance", "detail", "title", "type"} {
		delete(m, f)
	}
	for k, v := range m {
		p.Extensions[k] = v
	}
	return nil
}

func (p *Problem) JSON(w http.ResponseWriter) {
	body, err := json.Marshal(p)
	if err != nil {
		p.Status = http.StatusUnprocessableEntity
		body, _ = json.Marshal(Problem{
			Status: p.Status,
			Detail: err.Error(),
			Title:  "JSON Encoding Error",
		})
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_, _ = w.Write(body)
}

// toInt converts various numeric types into int.
func (p Problem) toInt(value any) (int, error) {
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.String:
		i, err := strconv.Atoi(v.String())
		if err != nil {
			return 0, fmt.Errorf("invalid status type: %v", v.Kind())
		}
		return i, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return int(v.Float()), nil
	default:
		return 0, fmt.Errorf("invalid status type: %T", value)
	}
}
