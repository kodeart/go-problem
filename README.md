# Problem

[![Codecov](https://codecov.io/gh/kodeart/go-problem/branch/master/graph/badge.svg)](https://codecov.io/gh/kodeart/go-problem)
[![Go Report Card](https://goreportcard.com/badge/github.com/kodeart/go-problem)](https://goreportcard.com/report/github.com/kodeart/go-problem)
[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/kodeart/go-problem/blob/master/LICENSE)

## Problem details for HTTP APIs per [RFC-9457][RFC9457] standard.

This module provides the `Problem` struct which can be used to represent a problem
in HTTP APIs. It implements the [RFC-9457][RFC9457] standard.

It supports serializing and deserializing the `Problem` struct to and from JSON.

## Usage

`go-problem` module provides an easy way to send the `Problem` struct as a response to the client.

### Example with HTTP handler and middleware 

```go
package middleware

import (
    "net/http"
	
    "github.com/kodeart/go-problem"
)

func NotFoundHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        p := problem.Problem{
            Status:   http.StatusNotFound,
            Detail:   "No such API route",
            Title:    "Route Not Found",
            Instance: r.URL.Path,
        }
        p.JSON(w)
    }
}

// or with helper methods

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
    problem.New().
        WithStatus(http.StatusNotFound).
        WithDetail("No such API route").
        WithTitle("Route Not Found").
        WithInstance(r.URL.Path).
        JSON(w)
}
```

```go
package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kodeart/go-problem"
)

func main() {
    mux := chi.NewRouter()
    mux.NotFound(middleware.NotFoundHandler)

    ...

    mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
        problem.New().
            WithStatus(http.StatusServiceUnavailable).
            WithExtension("maintenance", true).
            WithExtension("version", "1.0.0").
            JSON(w)
    })
}
```


### Create a `Problem` with helpers

```go
...
p := problem.New().
    WithStatus(http.StatusUnprocessableEntity).
    WithType("https://example.com/probs/out-of-credit").
    WithTitle("You do not have enough credit.").
    WithDetail("Your current balance is 30, but that costs 50.").
    WithInstance("/account/12345/msgs/abc").
    WithExtension("balance", 30).
    WithExtension("accounts", []string{
        "/account/12345",
        "/account/67890",
    })
}
```


### Create a `Problem` directly

```go
p := problem.Problem{
    Status:   http.StatusUnprocessableEntity,
    Type:     "https://example.com/probs/out-of-credit",
    Title:    "You do not have enough credit.",
    Detail:   "Your current balance is 30, but that costs 50.",
    Instance: "/account/12345/msgs/abc",
    Extensions: map[string]interface{}{
        "balance": 30,
        "accounts": []string{
            "/account/12345",
            "/account/67890",
		},
	},
}

```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


[RFC9457]: https://tools.ietf.org/html/rfc9457
