package problem

import (
	"encoding/xml"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

// XML writes an XML response to the client.
// Handles the encoding errors and returns a RFC-9457
// compliant error response.
func (p *Problem) XML(w http.ResponseWriter) {
    b, err := xml.Marshal(p)
    if err != nil {
        p.Status = http.StatusUnprocessableEntity
        b, _ = xml.Marshal(Problem{
            Status: p.Status,
            Detail: err.Error(),
            Title:  "XML Encoding Error",
        })
    }
    b = append([]byte(xml.Header), b...)
    w.Header().Set("Content-Type", "application/problem+xml")
    w = setCacheControl(w, p.Status)
    if p.Status > 0 {
        w.WriteHeader(p.Status)
    }
    _, _ = w.Write(b)
}

// MarshalXML implements xml.Marshaler interface to serialize
// the Problem struct into RFC-9457 XML format.
func (p Problem) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	var err error
    start := xml.StartElement{
        Name: xml.Name{Local: "problem"},
        Attr: []xml.Attr{{Name: xml.Name{Local: "xmlns"}, Value: "urn:ietf:rfc:7807"}},
    }
    _ = e.EncodeToken(start)
    if p.Status > 0 {
        _ = e.EncodeElement(p.Status, xml.StartElement{Name: xml.Name{Local: "status"}})
    }
    if p.Type != "" {
		_ = e.EncodeElement(p.Type, xml.StartElement{Name: xml.Name{Local: "type"}})
	}
	if p.Title != "" {
		_ = e.EncodeElement(p.Title, xml.StartElement{Name: xml.Name{Local: "title"}})
	}
	if p.Detail != "" {
		_ = e.EncodeElement(p.Detail, xml.StartElement{Name: xml.Name{Local: "detail"}})
	}
	if p.Instance != "" {
		_ = e.EncodeElement(p.Instance, xml.StartElement{Name: xml.Name{Local: "instance"}})
	}
    if err = marshalXmlValues(e, p.Extensions); err != nil {
    	return err
    }
    return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// UnmarshalXML implements xml.Unmarshaler interface to deserialize
// the XML string into Problem struct as in the RFC-9457 spec.
func (p *Problem) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var (
		token xml.Token
		err   error
	)
	if p.Extensions == nil {
		p.Extensions = make(map[string]any)
	}
	// Track extension elements by name to handle duplicates
	extensions := make(map[string][]any)

	for {
		token, err = d.Token()
		if err != nil {
			return err
		}

		switch t := token.(type) {
		case xml.StartElement:
			name := t.Name.Local
			switch name {
			case "status":
				var status string
				_ = d.DecodeElement(&status, &t)
				statusCode, err := p.statusCode(status)
				if err != nil {
					return err
				}
				p.Status = statusCode
				case "instance":
					_ = d.DecodeElement(&p.Instance, &t)
				case "detail":
					_ = d.DecodeElement(&p.Detail, &t)
				case "title":
					_ = d.DecodeElement(&p.Title, &t)
				case "type":
					_ = d.DecodeElement(&p.Type, &t)
			default:
				// This is an extension field
				var content struct {
					Data []byte `xml:",innerxml"`
				}
				_ = d.DecodeElement(&content, &t)
				if val := unmarshalXmlValue(string(content.Data)); val != "" {
					extensions[name] = append(extensions[name], val)
				}
			}
		case xml.EndElement:
			if t == start.End() {
				// Process all extension elements.
				// Use an array if multiple elements have the same name
				for name, values := range extensions {
					if len(values) == 1 {
						p.Extensions[name] = values[0]
					} else {
						p.Extensions[name] = values
					}
				}
				return nil
			}
		}
	}
}

// unmarshalXmlValue parses extension content into an appropriate Go value.
func unmarshalXmlValue(v string) any {
	v = strings.TrimSpace(v)
	if !strings.Contains(v, "<") {
        return v
    }
	if strings.Contains(v, "<i>") {
		var arr struct {
			Items []string `xml:"i"`
		}
		if err := xml.Unmarshal([]byte("<root>"+v+"</root>"), &arr); err == nil && len(arr.Items) > 0 {
			var items []any
			for _, item := range arr.Items {
				items = append(items, item)
			}
			return items
		}
	}
	// Try to parse as a map structure
	var m struct {
		Elements []struct {
			XMLName xml.Name
			Content []byte `xml:",innerxml"`
		} `xml:",any"`
	}
	if err := xml.Unmarshal([]byte("<root>"+v+"</root>"), &m); err == nil && len(m.Elements) > 0 {
		r := make(map[string]any)
		for _, elem := range m.Elements {
			name := elem.XMLName.Local
			elemContent := string(elem.Content)
			parsedContent := unmarshalXmlValue(elemContent)
			r[name] = parsedContent
		}
		return r
	}
	return v
}

// marshalXmlValues serializes extension values into XML format.
func marshalXmlValues(e *xml.Encoder, ext map[string]any) error {
	var err error
	// Sort extension keys for deterministic output
    keys := make([]string, 0, len(ext))
    for k := range ext {
    	keys = append(keys, k)
    }
    sort.Strings(keys)
	for _, k := range keys {
    	v := ext[k]
        switch val := v.(type) {
        case []any:
            for _, item := range val {
                if m, ok := item.(map[string]any); ok {
	                // Get sorted keys for consistent order
	                mapKeys := make([]string, 0, len(m))
	                for mk := range m {
	                    mapKeys = append(mapKeys, mk)
	                }
	                sort.Strings(mapKeys)
                    // Start a new tag for each map
                    _ = e.EncodeToken(xml.StartElement{Name: xml.Name{Local: k}})
                    // Encode map key-value pairs in sorted order
                    for _, mk := range mapKeys {
                        _ = e.EncodeElement(m[mk], xml.StartElement{Name: xml.Name{Local: mk}})
                    }
                    _ = e.EncodeToken(xml.EndElement{Name: xml.Name{Local: k}})
                }
            }
        case []string:
            _ = e.EncodeToken(xml.StartElement{Name: xml.Name{Local: k}})
            for _, item := range val {
                _ = e.EncodeElement(item, xml.StartElement{Name: xml.Name{Local: "i"}})
            }
            _ = e.EncodeToken(xml.EndElement{Name: xml.Name{Local: k}})
        default:
            rv := reflect.ValueOf(v)
            if rv.Kind() == reflect.Array {
                _ = e.EncodeToken(xml.StartElement{Name: xml.Name{Local: k}})
                j := rv.Len()
                for i := 0; i < j; i++ {
                    _ = e.EncodeElement(rv.Index(i).Interface(), xml.StartElement{Name: xml.Name{Local: "i"}})
                }
                _ = e.EncodeToken(xml.EndElement{Name: xml.Name{Local: k}})
            } else {
                if err = e.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: k}}); err != nil {
                    return err
                }
            }
        }
    }
    return nil
}
