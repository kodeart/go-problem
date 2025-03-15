package problem_test

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kodeart/go-problem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXmlMarshal(t *testing.T) {
	t.Run("should create a new empty instance", func(t *testing.T) {
		p := problem.Problem{}
		xmlData, err := xml.Marshal(p)

		require.Nil(t, err)
		assert.Equal(t, `<problem xmlns="urn:ietf:rfc:7807"></problem>`, string(xmlData))
	})

	t.Run("should marshal to xml", func(t *testing.T) {
        p := problem.Problem{
            Status:   http.StatusForbidden,
            Title:    "Balance Error",
            Detail:   "You do not have enough credit.",
            Instance: "/",
            Type:     "/errors/balance",
            // Extensions: problem.Extensions{
            Extensions: map[string]any{
                "balance":  42,
                "accounts": []any{"/account/12345", "/account/67890"},
            },
        }
        xmlData, err := xml.Marshal(p)

        assert.Nil(t, err)
        expected := `<problem xmlns="urn:ietf:rfc:7807"><status>403</status><type>/errors/balance</type><title>Balance Error</title><detail>You do not have enough credit.</detail><instance>/</instance><balance>42</balance></problem>`
        assert.Equal(t, expected, string(xmlData))
    })

	t.Run("should add and remove extensions to instance", func(t *testing.T) {
		p := problem.Problem{
			Instance: "/",
			Detail:   "Balance error",
			Status:   http.StatusBadRequest,
			Type:     "/errors/balance",
		}

		p.WithExtension("customField", "customValue").
		WithExtension("balance", float64(42)).
		WithExtension("accounts", []string{"account1", "account2"}).
		WithoutExtension("customField")

		body, err := xml.Marshal(p)
		expected := `<problem xmlns="urn:ietf:rfc:7807"><status>400</status><type>/errors/balance</type><detail>Balance error</detail><instance>/</instance><accounts><i>account1</i><i>account2</i></accounts><balance>42</balance></problem>`

		require.Nil(t, err)
		assert.Equal(t, expected, string(body))
	})

	t.Run("should not set a nil extension value", func(t *testing.T) {
    	p := problem.Problem{
   			Extensions: map[string]any{"errors":  nil},
     	}
		body, err := xml.Marshal(p)

		require.Nil(t, err)
		assert.Equal(t, `<problem xmlns="urn:ietf:rfc:7807"></problem>`, string(body))
    })
}

func TestXmlUnmarshal(t *testing.T) {
	t.Run("should fail unmarshalling an invalid xml", func(t *testing.T) {
		var p problem.Problem
		xmlData := `<foo></bar>`
		err := xml.Unmarshal([]byte(xmlData), &p)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "XML syntax error")
	})

	t.Run("should fail unmarshalling nil", func(t *testing.T) {
        var e problem.Problem
        err := xml.Unmarshal(nil, &e)

        require.Error(t, err)
		assert.Contains(t, err.Error(), "EOF")
    })

	t.Run("should fail unmarshalling an empty string", func(t *testing.T) {
        var e problem.Problem
        err := xml.Unmarshal([]byte(""), &e)

        require.Error(t, err)
		assert.Contains(t, err.Error(), "EOF")
    })

	t.Run("should construct from empty xml problem", func(t *testing.T) {
        var e problem.Problem
        err := xml.Unmarshal([]byte(`<problem></problem>`), &e)

        require.Nil(t, err)
        assert.Equal(t, e.Status, 0)
        assert.Empty(t, e.Detail)
        assert.Empty(t, e.Instance)
        assert.Empty(t, e.Title)
        assert.Empty(t, e.Type)
        assert.Empty(t, e.Extensions)
    })

	t.Run("should return error if status is empty string", func(t *testing.T) {
        var e problem.Problem
        err := xml.Unmarshal([]byte(`<problem><status></status></problem>`), &e)
        require.EqualError(t, err, "invalid status type: string")
    })

    t.Run("should return error if status is a string", func(t *testing.T) {
        var e problem.Problem
        err := xml.Unmarshal([]byte(`<problem><status>invalid</status></problem>`), &e)

        require.EqualError(t, err, "invalid status type: string")
    })

    t.Run("should unmarshal non rfc-9457 xml string", func(t *testing.T) {
        var e problem.Problem
        legacyError := `<problem>
        	<message>refactor this message</message>
         	<code>500</code>
          </problem>`
        err := xml.Unmarshal([]byte(legacyError), &e)

        require.Nil(t, err)
        assert.Equal(t, "refactor this message", e.Extensions["message"])
        assert.Equal(t, "500", e.Extensions["code"])
    })

	t.Run("should unmarshal only provided fields without extensions", func(t *testing.T) {
        var e problem.Problem
        xmlData := `<problem>
        	<status>200</status>
         	<detail>Hello World</detail>
        </problem>`

        err := xml.Unmarshal([]byte(xmlData), &e)

        assert.Nil(t, err)
        assert.Equal(t, http.StatusOK, e.Status)
        assert.Equal(t, "Hello World", e.Detail)
        assert.Equal(t, map[string]any{}, e.Extensions)
        // other fields should be empty
        assert.Empty(t, e.Instance)
        assert.Empty(t, e.Title)
        assert.Empty(t, e.Type)
        assert.Empty(t, e.Extensions)
    })

	t.Run("should unmarshal only provided fields with extensions", func(t *testing.T) {
		var e problem.Problem
        xmlData := `<?xml version="1.0" encoding="UTF-8" ?>
        <problem>
	        <status>422</status>
	        <type>https://example.net/validation-error</type>
	        <title>Your request is not valid.</title>
	        <errors>
	            <detail>must be a positive integer</detail>
	            <pointer>#/age</pointer>
	        </errors>
	        <errors>
	            <detail>must be 'green', 'red' or 'blue'</detail>
	            <pointer>#/profile/color</pointer>
	        </errors>
        </problem>`

        err := xml.Unmarshal( []byte(xmlData), &e)

        assert.Nil(t, err)
        // these should be empty
        assert.Empty(t, e.Instance)
        assert.Empty(t, e.Detail)

        assert.Equal(t, http.StatusUnprocessableEntity, e.Status)
        assert.Equal(t, "Your request is not valid.", e.Title)
        assert.Equal(t, "https://example.net/validation-error", e.Type)
        assert.Equal(t, map[string]any{
            "errors": []any{
                map[string]any{"detail": "must be a positive integer", "pointer": "#/age"},
                map[string]any{"detail": "must be 'green', 'red' or 'blue'", "pointer": "#/profile/color"},
            },
        }, e.Extensions)

        extErrors := e.Extensions["errors"].([]any)
        assert.Equal(t, map[string]any{
            "detail":  "must be a positive integer",
            "pointer": "#/age",
        }, extErrors[0])

        assert.Equal(t, map[string]any{
            "detail":  "must be 'green', 'red' or 'blue'",
            "pointer": "#/profile/color",
        }, extErrors[1])

    })

 	t.Run("should unmarshal all fields", func(t *testing.T) {
        var p problem.Problem
        xmlData := `<?xml version="1.0" encoding="UTF-8"?>
           <problem xmlns="urn:ietf:rfc:7807">
             <type>https://example.com/probs/out-of-credit</type>
             <title>You do not have enough credit.</title>
             <detail>Your current balance is 30, but that costs 50.</detail>
             <instance>https://example.net/account/12345/msgs/abc</instance>
             <status>403</status>
             <balance>30.99</balance>
             <accounts>
               <i>https://example.net/account/12345</i>
               <i>https://example.net/account/67890</i>
             </accounts>
           </problem>`

        err := xml.Unmarshal([]byte(xmlData), &p)

        require.Nil(t, err)
        assert.Equal(t, 403, p.Status)
        assert.Equal(t, "You do not have enough credit.", p.Title)
        assert.Equal(t, "Your current balance is 30, but that costs 50.", p.Detail)
        assert.Equal(t, "https://example.net/account/12345/msgs/abc", p.Instance)
        assert.Equal(t, "https://example.com/probs/out-of-credit", p.Type)
        assert.Equal(t, map[string]any{
        	"balance":  "30.99",
         	"accounts": []any{"https://example.net/account/12345", "https://example.net/account/67890"},
        }, p.Extensions)
  	})

}

func TestXmlRenderer(t *testing.T) {
    t.Run("should render an empty xml if problem is empty", func(t *testing.T) {
        w := httptest.NewRecorder()
        p := problem.Problem{}
        p.XML(w)
        resp := w.Result()

        assert.Equal(t, http.StatusOK, resp.StatusCode)
        assert.Equal(t, "application/problem+xml", resp.Header.Get("Content-Type"))
        assert.Empty(t, resp.Header.Get("Cache-Control"))
        assert.Equal(t, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<problem xmlns=\"urn:ietf:rfc:7807\"></problem>", w.Body.String())
    })

    t.Run("should render as xml response", func(t *testing.T) {
        w := httptest.NewRecorder()
        p := problem.Problem{
            Status:   http.StatusServiceUnavailable,
            Title:    "Service Maintenance",
            Detail:   "API is under maintenance",
            Instance: "/ping",
        }
        p.WithExtension("version", "1.0.0")
        p.WithExtension("maintenance", true)
        p.XML(w)
        resp := w.Result()

        assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
        assert.Equal(t, "application/problem+xml", resp.Header.Get("Content-Type"))
        assert.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get("Cache-Control"))

        // the XML output is non-deterministic, so we cannot assert on it
        body := w.Body.String()
        assert.Contains(t, body, `<problem xmlns="urn:ietf:rfc:7807">`)
        assert.Contains(t, body, "<status>503</status>")
        assert.Contains(t, body, "<detail>API is under maintenance</detail>")
        assert.Contains(t, body, "<title>Service Maintenance</title>")
        assert.Contains(t, body, "<instance>/ping</instance>")
        assert.Contains(t, body, "<version>1.0.0</version>")
        assert.Contains(t, body, "<maintenance>true</maintenance>")
        assert.Contains(t, body, "</problem>")
    })

    t.Run("should render generic xml error if cannot encode the struct", func(t *testing.T) {
        w := httptest.NewRecorder()
        p := problem.Problem{}
        p.WithExtension("bogus", func() {})
        p.XML(w)
        resp := w.Result()

        assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
        assert.Equal(t, "application/problem+xml", resp.Header.Get("Content-Type"))
        assert.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get("Cache-Control"))

        body := w.Body.String()
        assert.Contains(t, body, `<problem xmlns="urn:ietf:rfc:7807">`)
        assert.Contains(t, body, "<status>422</status>")
        assert.Contains(t, body, "<detail>xml: unsupported type: func()</detail>")
        assert.Contains(t, body, "<title>XML Encoding Error</title>")
        assert.Contains(t, body, "</problem>")
    })

    t.Run("should not add cache control header for non cacheable response", func(t *testing.T) {
        w := httptest.NewRecorder()
        p := problem.Problem{
            Status:   http.StatusNotFound,
            Title:    "Page Not Found",
            Detail:   "The page you are looking for does not exist",
            Instance: "/fubar",
        }
        p.XML(w)
        resp := w.Result()

        assert.Empty(t, resp.Header.Get("Cache-Control"), "Cache-Control header is not set")
    })

    t.Run("should render slice extensions", func(t *testing.T) {
        w := httptest.NewRecorder()
        p := problem.Problem{}
        p.WithExtension("accounts", []string{
            "/account/1234",
            "/account/5678",
        })
        p.XML(w)

        assert.Contains(t, w.Body.String(), "<accounts><i>/account/1234</i><i>/account/5678</i></accounts>")
    })

    t.Run("should render array extensions", func(t *testing.T) {
        w := httptest.NewRecorder()
        p := problem.Problem{}
        p.WithExtension("accounts", [2]string{
            "/account/1234",
            "/account/5678",
        })
        p.XML(w)

        assert.Contains(t, w.Body.String(), "<accounts><i>/account/1234</i><i>/account/5678</i></accounts>")
    })

    t.Run("should render map extensions", func(t *testing.T) {
        w := httptest.NewRecorder()
        p := problem.Problem{}
        p.WithExtension("errors", []any{
            map[string]any{"detail": "must be a positive integer", "pointer": "#/age"},
            map[string]any{"detail": "must be green, red or blue", "pointer": "#/profile/color"},
        })
        p.XML(w)
        body := w.Body.String()

        assert.Contains(t, body, "<errors><detail>must be a positive integer</detail><pointer>#/age</pointer></errors>")
        assert.Contains(t, body, "<errors><detail>must be green, red or blue</detail><pointer>#/profile/color</pointer></errors>")
    })
}
