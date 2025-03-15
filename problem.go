package problem

// Problem is a struct that represents a problem instance
// as defined in RFC-9457 (https://tools.ietf.org/html/rfc9457).
type Problem struct {
    Status     int            `json:"status,omitempty" xml:"status,omitempty"`
    Instance   string         `json:"instance,omitempty" xml:"instance,omitempty"`
    Detail     string         `json:"detail,omitempty" xml:"detail,omitempty"`
    Title      string         `json:"title,omitempty" xml:"title,omitempty"`
    Type       string         `json:"type,omitempty" xml:"type,omitempty"`
    Extensions map[string]any `json:"-" xml:"-"`
}

// New returns a new Problem instance.
func New() *Problem {
    return &Problem{Extensions: make(map[string]any)}
}

// WithStatus sets the status code value.
func (p *Problem) WithStatus(v int) *Problem {
    p.Status = v
    return p
}

// WithInstance sets the instance URI value.
func (p *Problem) WithInstance(v string) *Problem {
    p.Instance = v
    return p
}

// WithDetail sets the problem detail value.
func (p *Problem) WithDetail(v string) *Problem {
    p.Detail = v
    return p
}

// WithTitle sets the problem title value.
func (p *Problem) WithTitle(v string) *Problem {
    p.Title = v
    return p
}

// WithType sets the problem type.
func (p *Problem) WithType(v string) *Problem {
    p.Type = v
    return p
}

// WithExtension adds key:value pairs to internal Extensions map.
// When JSON serialization is performed, these pairs are
// included in the final JSON response as additional values.
func (p *Problem) WithExtension(key string, val any) *Problem {
    if p.Extensions == nil {
        p.Extensions = make(map[string]any)
    }
    p.Extensions[key] = val
    return p
}

// WithoutExtension removes a key from internal Extensions map.
// If key is nil or there is no such element, WithoutExtension is a no-op.
func (p *Problem) WithoutExtension(key string) *Problem {
    delete(p.Extensions, key)
    return p
}

// GetExtension retrieves a value from Extensions map.
// If there is no such element, nil is returned.
// If you know the type of the value, you can assert it.
//
// Examples:
// p := problem.New().WithExtension("key", "value")
//
// v := p.GetExtension("key")
// s := p.GetExtension("key".(string)
// i := p.GetExtension("key").(int)
// b := p.GetExtension("key").(bool)
func (p *Problem) GetExtension(key string) any {
    if v, ok := (p.Extensions)[key]; ok {
        return v
    }
    return nil
}
