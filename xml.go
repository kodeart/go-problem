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
    if err = e.EncodeToken(start); err != nil {
        return err
    }
    if p.Status > 0 {
        if err = e.EncodeElement(p.Status, xml.StartElement{Name: xml.Name{Local: "status"}}); err != nil {
            return err
        }
    }
    if p.Type != "" {
        if err = e.EncodeElement(p.Type, xml.StartElement{Name: xml.Name{Local: "type"}}); err != nil {
            return err
        }
    }
    if p.Title != "" {
        if err = e.EncodeElement(p.Title, xml.StartElement{Name: xml.Name{Local: "title"}}); err != nil {
            return err
        }
    }
    if p.Detail != "" {
        if err = e.EncodeElement(p.Detail, xml.StartElement{Name: xml.Name{Local: "detail"}}); err != nil {
            return err
        }
    }
    if p.Instance != "" {
        if err = e.EncodeElement(p.Instance, xml.StartElement{Name: xml.Name{Local: "instance"}}); err != nil {
            return err
        }
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
				if err = d.DecodeElement(&status, &t); err != nil {
					return err
				}
				statusCode, err := p.statusCode(status)
				if err != nil {
					return err
				}
				p.Status = statusCode
			case "instance":
				if err = d.DecodeElement(&p.Instance, &t); err != nil {
					return err
				}
			case "detail":
				if err = d.DecodeElement(&p.Detail, &t); err != nil {
					return err
				}
			case "title":
				if err = d.DecodeElement(&p.Title, &t); err != nil {
					return err
				}
			case "type":
				if err = d.DecodeElement(&p.Type, &t); err != nil {
					return err
				}
			default:
				// This is an extension field
				var content struct {
					Data []byte `xml:",innerxml"`
				}
				if err = d.DecodeElement(&content, &t); err != nil {
					return err
				}
				val := unmarshalXmlValue(string(content.Data))
				if val != nil {
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
                    if err = e.EncodeToken(xml.StartElement{Name: xml.Name{Local: k}}); err != nil {
                        return err
                    }
                    // Encode map key-value pairs in sorted order
                    for _, mk := range mapKeys {
                        if err = e.EncodeElement(m[mk], xml.StartElement{Name: xml.Name{Local: mk}}); err != nil {
                            return err
                        }
                    }
                    if err = e.EncodeToken(xml.EndElement{Name: xml.Name{Local: k}}); err != nil {
                        return err
                    }
                }
            }
        case []string:
            if err = e.EncodeToken(xml.StartElement{Name: xml.Name{Local: k}}); err != nil {
                return err
            }
            for _, item := range val {
                if err = e.EncodeElement(item, xml.StartElement{Name: xml.Name{Local: "i"}}); err != nil {
                    return err
                }
            }
            if err = e.EncodeToken(xml.EndElement{Name: xml.Name{Local: k}}); err != nil {
                return err
            }
        default:
            rv := reflect.ValueOf(v)
            if rv.Kind() == reflect.Array {
                if err = e.EncodeToken(xml.StartElement{Name: xml.Name{Local: k}}); err != nil {
                    return err
                }
                j := rv.Len()
                for i := 0; i < j; i++ {
                    if err = e.EncodeElement(rv.Index(i).Interface(), xml.StartElement{Name: xml.Name{Local: "i"}}); err != nil {
                        return err
                    }
                }
                if err = e.EncodeToken(xml.EndElement{Name: xml.Name{Local: k}}); err != nil {
                    return err
                }
            } else {
                if err = e.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: k}}); err != nil {
                    return err
                }
            }
        }
    }
    return nil
}
