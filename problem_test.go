package problem_test

import (
	"net/http"
	"testing"

	"github.com/kodeart/go-problem"
	"github.com/stretchr/testify/assert"
)

func TestProblem(t *testing.T) {
	t.Run("should construct the empty instance", func(t *testing.T) {
		p := problem.Problem{}
		assert.Empty(t, p.Status)
		assert.Empty(t, p.Instance)
		assert.Empty(t, p.Detail)
		assert.Empty(t, p.Title)
		assert.Empty(t, p.Type)
		assert.Empty(t, p.Extensions)
	})

	t.Run("should populate the instance with helper methods", func(t *testing.T) {
		p := problem.New().
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
		assert.Equal(t, 30, p.GetExtension("balance"))
		assert.Equal(t, []string{"/account/12345", "/account/67890"}, p.GetExtension("accounts"))
	})

	t.Run("should not read the standard fields", func(t *testing.T) {
		p := problem.New().
			WithStatus(http.StatusForbidden).
			WithType("https://example.com/probs/out-of-credit").
			WithTitle("You do not have enough credit.").
			WithDetail("Your current balance is 30, but that costs 50.").
			WithInstance("/account/12345/msgs/abc")

		assert.Nil(t, p.GetExtension("status"))
		assert.Nil(t, p.GetExtension("type"))
		assert.Nil(t, p.GetExtension("title"))
		assert.Nil(t, p.GetExtension("detail"))
		assert.Nil(t, p.GetExtension("instance"))
	})

	t.Run("should return nil for non-existing extension", func(t *testing.T) {
		p := problem.Problem{}
		assert.Nil(t, p.GetExtension("non-existing-key"))
	})

	t.Run("should add and remove extensions to instance", func(t *testing.T) {
		p := problem.New().
			WithExtension("customField", "customValue").
			WithExtension("balance", 42).
			WithoutExtension("customField")

		assert.Nil(t, p.GetExtension("customField"))
		assert.Len(t, p.Extensions, 1)
	})
}
