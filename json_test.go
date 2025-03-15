package problem_test

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/kodeart/go-problem/v2"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestJsonMarshal(t *testing.T) {
	t.Run("should create a new empty instance", func(t *testing.T) {
		p := problem.Problem{}
		jsonData, err := json.Marshal(p)

		require.Nil(t, err)
		assert.JSONEq(t, `{}`, string(jsonData))
	})

    t.Run("should marshal to json", func(t *testing.T) {
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
        jsonData, err := json.Marshal(p)

        assert.Nil(t, err)
        expected := `{"status":403,"instance":"/","title":"Balance Error","detail":"You do not have enough credit.","balance":42,"accounts":["/account/12345","/account/67890"],"type":"/errors/balance"}`
        assert.JSONEq(t, expected, string(jsonData))
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

		body, err := json.Marshal(p)
		expected := `{"status":400,"instance":"/","detail":"Balance error","balance":42,"accounts":["account1","account2"],"type":"/errors/balance"}`

		require.Nil(t, err)
		assert.JSONEq(t, expected, string(body))
	})

    t.Run("should not set extension value if nil", func(t *testing.T) {
    	p := problem.Problem{
   			Extensions: map[string]any{"errors":  nil},
     	}
		body, err := json.Marshal(p)

		require.Nil(t, err)
		assert.JSONEq(t, `{}`, string(body))
    })
}

func TestJsonUnmarshal(t *testing.T) {
	t.Run("should fail unmarshalling an invalid json", func(t *testing.T) {
		var e problem.Problem
		jsonData := `{"type": null,`
		err := e.UnmarshalJSON( []byte(jsonData))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected end of JSON input")
	})

	t.Run("should fail unmarshalling nil", func(t *testing.T) {
        var e problem.Problem
        err := json.Unmarshal(nil, &e)

        require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected end of JSON input")
    })

	t.Run("should fail unmarshalling an empty string", func(t *testing.T) {
        var e problem.Problem
        err := json.Unmarshal([]byte(""), &e)

        require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected end of JSON input")
    })

	t.Run("should fail unmarshalling a json array", func(t *testing.T) {
        var e problem.Problem
        err := json.Unmarshal([]byte(`[]`), &e)

        require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot unmarshal array")
    })

	t.Run("should construct from empty json object", func(t *testing.T) {
        var e problem.Problem
        err := json.Unmarshal([]byte(`{}`), &e)

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
        err := json.Unmarshal([]byte(`{"status": ""}`), &e)
        require.EqualError(t, err, "invalid status type: string")
    })

    t.Run("should return error if status is a string", func(t *testing.T) {
        var e problem.Problem
        err := json.Unmarshal([]byte(`{"status": "invalid"}`), &e)

        require.EqualError(t, err, "invalid status type: string")
    })

    t.Run("should return error if status is not a number", func(t *testing.T) {
        var e problem.Problem
        err := json.Unmarshal([]byte(`{"status": true}`), &e)

        require.EqualError(t, err, "invalid status type: bool")
    })

    t.Run("should return error if status is nil", func(t *testing.T) {
        var e problem.Problem
        err := json.Unmarshal([]byte(`{"status": null}`), &e)

        require.EqualError(t, err, "invalid status type: <nil>")
    })

    t.Run("should convert the status from string integer", func(t *testing.T) {
        var e problem.Problem
        err := json.Unmarshal([]byte(`{"status": "200"}`), &e)
        assert.Nil(t, err)

        assert.Equal(t, http.StatusOK, e.Status)
    })

    t.Run("should unmarshal non rfc-9457 json string", func(t *testing.T) {
        var e problem.Problem
        legacyError := `{"message": "refactor this message", "code": 500}`
        err := json.Unmarshal([]byte(legacyError), &e)

        require.Nil(t, err)
        assert.Equal(t, "refactor this message", e.Extensions["message"])
        assert.Equal(t, float64(500), e.Extensions["code"], "json.Unmarshal() creates a float64")
    })


    t.Run("should unmarshal only provided fields without extensions", func(t *testing.T) {
        var e problem.Problem
        jsonData := []byte(`{"status": 200, "detail": "Hello World"}`)

        err := json.Unmarshal(jsonData, &e)

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
        jsonData := `{
			"status": 422,
			"type": "https://example.net/validation-error",
			"title": "Your request is not valid.",
			"errors": [
				{
					"detail": "must be a positive integer",
					"pointer": "#/age"
				},
				{
					"detail": "must be 'green', 'red' or 'blue'",
					"pointer": "#/profile/color"
				}
			]
		}`

        err := json.Unmarshal( []byte(jsonData), &e)

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
        var e problem.Problem
        jsonData :=`{
			"status": 403,
			"type": "https://example.com/probs/out-of-credit",
			"title": "You do not have enough credit.",
			"detail": "Your current balance is 30, but that costs 50.",
			"instance": "/account/12345/msgs/abc",
			"balance": 30,
			"accounts": ["/account/12345", "/account/67890"]
		}`
        err := json.Unmarshal( []byte(jsonData), &e)

        require.Nil(t, err)
        assert.Equal(t, http.StatusForbidden, e.Status)
        assert.Equal(t, "/account/12345/msgs/abc", e.Instance)
        assert.Equal(t, "Your current balance is 30, but that costs 50.", e.Detail)
        assert.Equal(t, "You do not have enough credit.", e.Title)
        assert.Equal(t, "https://example.com/probs/out-of-credit", e.Type)
        assert.Equal(t, map[string]any{
            "balance":  float64(30),
            "accounts": []any{"/account/12345", "/account/67890"},
        }, e.Extensions)

        assert.Equal(t, float64(30), e.Extensions["balance"])
        assert.Equal(t, []any{"/account/12345", "/account/67890"}, e.Extensions["accounts"])
    })
}

func TestJsonRenderer(t *testing.T) {
    t.Run("should render an empty json object if problem is empty", func(t *testing.T) {
        w := httptest.NewRecorder()
        p := problem.Problem{}
        p.JSON(w)
        resp := w.Result()

        assert.Equal(t, http.StatusOK, resp.StatusCode)
        assert.Equal(t, "application/problem+json", resp.Header.Get("Content-Type"))
        assert.Empty(t, resp.Header.Get("Cache-Control"), "caching is fine for empty problem")
        assert.Equal(t, `{}`, w.Body.String())
    })

    t.Run("should render as json response", func(t *testing.T) {
        w := httptest.NewRecorder()
        p := problem.Problem{
            Status:   http.StatusServiceUnavailable,
            Title:    "Service Maintenance",
            Detail:   "API is under maintenance",
            Instance: "/ping",
        }
        p.WithExtension("version", "1.0.0")
        p.WithExtension("maintenance", true)
        p.JSON(w)
        resp := w.Result()

        assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
        assert.Equal(t, "application/problem+json", resp.Header.Get("Content-Type"))
        assert.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get("Cache-Control"))
        assert.JSONEq(t, `{"title":"Service Maintenance","detail":"API is under maintenance","instance":"/ping","status":503,"version":"1.0.0","maintenance":true}`, w.Body.String())
    })

    t.Run("should render generic json error if cannot encode the struct", func(t *testing.T) {
        w := httptest.NewRecorder()
        p := problem.Problem{}
        p.WithExtension("bogus", func() {})

        p.JSON(w)
        resp := w.Result()

        assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
        assert.Equal(t, "application/problem+json", resp.Header.Get("Content-Type"))
        assert.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get("Cache-Control"))
        assert.JSONEq(t, `{"detail":"json: error calling MarshalJSON for type *problem.Problem: json: unsupported type: func()", "status":422, "title":"JSON Encoding Error"}`, w.Body.String())
    })

    t.Run("should not add cache control header for non cacheable response", func(t *testing.T) {
        w := httptest.NewRecorder()
        p := problem.Problem{
            Status:   http.StatusNotFound,
            Title:    "Page Not Found",
            Detail:   "The page you are looking for does not exist",
            Instance: "/fubar",
        }

        p.JSON(w)
        resp := w.Result()

        assert.Empty(t, resp.Header.Get("Cache-Control"), "Cache-Control header is not set")
    })
}
