package problem

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProblem(t *testing.T) {
	t.Run("should construct with empty json string", func(t *testing.T) {
		var e Problem
		err := json.Unmarshal([]byte(`{}`), &e)
		assert.Nil(t, err)
		assert.Equal(t, e.Status, 0)
		assert.Empty(t, e.Detail)
		assert.Empty(t, e.Instance)
		assert.Empty(t, e.Title)
		assert.Empty(t, e.Type)
		assert.Equal(t, e.Extensions, map[string]any{})
	})

	t.Run("should return error if status is empty string", func(t *testing.T) {
		var e Problem
		err := json.Unmarshal([]byte(`{"status": ""}`), &e)
		require.EqualError(t, err, "invalid status type: string")
	})

	t.Run("should return error if status is a string", func(t *testing.T) {
		var e Problem
		err := json.Unmarshal([]byte(`{"status": "invalid"}`), &e)
		require.EqualError(t, err, "invalid status type: string")
	})

	t.Run("should return error if status is not a number", func(t *testing.T) {
		var e Problem
		err := json.Unmarshal([]byte(`{"status": true}`), &e)
		require.EqualError(t, err, "invalid status type: bool")
	})

	t.Run("should return error if status is nil", func(t *testing.T) {
		var e Problem
		err := json.Unmarshal([]byte(`{"status": null}`), &e)
		require.EqualError(t, err, "invalid status type: <nil>")
	})

	t.Run("should convert the status from string integer", func(t *testing.T) {
		var e Problem
		err := json.Unmarshal([]byte(`{"status": "200"}`), &e)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, e.Status)
	})

	t.Run("should return nil for non-existing extension", func(t *testing.T) {
		var e Problem
		err := json.Unmarshal([]byte(`{"status": 200}`), &e)
		assert.Nil(t, err)
		assert.Nil(t, e.Extensions["non-existing-key"])
	})

	t.Run("should add and remove extensions to instance", func(t *testing.T) {
		p := Problem{
			Instance: "/",
			Detail:   "Balance error",
			Status:   http.StatusBadRequest,
			Type:     "/errors/balance",
		}

		p.
			WithExtension("customField", "customValue").
			WithExtension("balance", 42).
			WithExtension("accounts", []string{"account1", "account2"}).
			WithoutExtension("customField")

		body, err := json.Marshal(p)
		require.Nil(t, err)

		expected := `{"status":400,"instance":"/","detail":"Balance error","balance":42,"accounts":["account1","account2"],"type":"/errors/balance"}`
		assert.JSONEq(t, expected, string(body))
	})

	t.Run("should unmarshal some fields without extensions", func(t *testing.T) {
		var e Problem
		jsonData := []byte(`{"status": 200, "detail": "Hello World"}`)

		err := json.Unmarshal(jsonData, &e)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusOK, e.Status)
		assert.Empty(t, e.Instance)
		assert.Equal(t, "Hello World", e.Detail)
		assert.Empty(t, e.Title)
		assert.Empty(t, e.Type)
		assert.IsType(t, map[string]any{}, e.Extensions)
		assert.Empty(t, e.Extensions)
	})

	t.Run("should unmarshal some fields with extensions", func(t *testing.T) {
		var e Problem
		jsonData := []byte(`{
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
		}`)

		err := json.Unmarshal(jsonData, &e)
		assert.Nil(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, e.Status)
		assert.Empty(t, e.Instance)
		assert.Empty(t, e.Detail)
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
		var e Problem
		jsonData := []byte(`{
			"status": 403,
			"type": "https://example.com/probs/out-of-credit",
			"title": "You do not have enough credit.",
			"detail": "Your current balance is 30, but that costs 50.",
			"instance": "/account/12345/msgs/abc",
			"balance": 30,
			"accounts": ["/account/12345", "/account/67890"]
		}`)

		err := json.Unmarshal(jsonData, &e)
		assert.Nil(t, err)

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

	t.Run("should create a new empty instance", func(t *testing.T) {
		p := New()

		assert.Empty(t, p.Status)
		assert.Empty(t, p.Instance)
		assert.Empty(t, p.Extensions)
		assert.Empty(t, p.Title)
		assert.Empty(t, p.Detail)
		assert.Empty(t, p.Type)
		assert.IsType(t, map[string]any{}, p.Extensions)

		jsonData, err := json.Marshal(p)
		require.Nil(t, err)
		assert.JSONEq(t, `{}`, string(jsonData))
	})

	t.Run("should populate the instance with helper methods", func(t *testing.T) {
		p := New().
			WithStatus(http.StatusForbidden).
			WithType("https://example.com/probs/out-of-credit").
			WithTitle("You do not have enough credit.").
			WithDetail("Your current balance is 30, but that costs 50.").
			WithInstance("/account/12345/msgs/abc").
			WithExtension("balance", 30).
			WithExtension("accounts", []string{"/account/12345", "/account/67890"})

		assert.Equal(t, http.StatusForbidden, p.Status)
		assert.Equal(t, "You do not have enough credit.", p.Title)
		assert.Equal(t, "Your current balance is 30, but that costs 50.", p.Detail)
		assert.Equal(t, "https://example.com/probs/out-of-credit", p.Type)
		assert.Equal(t, "/account/12345/msgs/abc", p.Instance)

		assert.Equal(t, 30, p.Extensions["balance"])
		assert.Equal(t, []string{"/account/12345", "/account/67890"}, p.Extensions["accounts"])
	})
}
