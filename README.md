# Problem Details

[![Codecov](https://codecov.io/gh/kodeart/go-problem/branch/master/graph/badge.svg)](https://codecov.io/gh/kodeart/go-problem)
[![Go Report Card](https://goreportcard.com/badge/github.com/kodeart/go-problem)](https://goreportcard.com/report/github.com/kodeart/go-problem)
[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/kodeart/go-problem/blob/master/LICENSE)

## Problem details for HTTP APIs per [RFC-9457][RFC9457] standard.

This module provides a `Problem` struct which can be used to represent a problem
in HTTP APIs. It implements the [RFC-9457][RFC9457] standard that creates error responses.

### The benefits of using `go-problem`

- RFC 9457 and 7807 compliant error responses
- Consistent error format across the API
- Built-in support for HTTP status codes
- Extensible error details
- Clean and readable error handling
- Built-in JSON and XML marshaling

## Installation

```bash
go get -u github.com/kodeart/go-problem
```

### What is it again?

It's for creating a standardized error responses, like this JSON for example:

```json
{
  "status": 403,
  "type": "https://example.com/probs/out-of-credit",
  "title": "You do not have enough credit.",
  "detail": "Your current balance is 30, but that costs 50.",
  "instance": "/account/12345/msgs/abc",
  "balance": 30,
  "accounts": ["/account/12345", "/account/67890"]
}
```
or

```xml
<?xml version="1.0" encoding="UTF-8"?>
<problem xmlns="urn:ietf:rfc:7807">
  <type>https://example.com/probs/out-of-credit</type>
  <title>You do not have enough credit.</title>
  <detail>Your current balance is 30, but that costs 50.</detail>
  <instance>https://example.net/account/12345/msgs/abc</instance>
  <balance>30</balance>
  <accounts>
    <i>https://example.net/account/12345</i>
    <i>https://example.net/account/67890</i>
  </accounts>
</problem>
```

or even this, **which defeats the whole purpose of the RFC**, but it may be useful for your current/legacy code:

```json
{
  "message": "Failed to use the problem details",
  "code": 500
}
```

```xml
<problem>
    <message>Failed to use the problem details</message>
    <code>500</code>
</problem>
```

Don't do this, please.


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

    // ...

    mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
        problem.New().
            WithStatus(http.StatusServiceUnavailable).
            WithExtension("maintenance", true).
            WithExtension("version", "1.0.0").
            JSON(w)
    })
}
```

### Helpers

Any key-value pair outide the standard fields can be accessed with

```go
p := problem.New().WithExtension("key", "value")

v := p.GetExtension("key")

// if you know the type of the value, assert it
intVal := p.GetExtension("key name").(int)
```
If there is no such element, `nil` is returned.

### Create a `Problem` with helpers

```go
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
    Extensions: map[string]any{
        "balance": 30,
        "accounts": []string{
            "/account/12345",
            "/account/67890",
		},
	},
}
```

## Rendering

`Problem` supports serializing and deserializing the data to and from JSON and XML.

### JSON

The `JSON()` method will render the `Problem` struct to JSON string,

```go
p.JSON(w)
```

while the standard `json.Unmarshal()` method will try to parse a JSON string into a `Problem` struct:

```go
var p problem.Problem
jsonString := `{...}` // some error message

err := json.Unmarshal([]byte(jsonString), &p)
```

### XML

The `XML()` method will render the `Problem` struct to XML string,

```go
p.XML(w)
```

```go
var p problem.Problem
xmlString := `<problem>...</problem>`

err := xml.Unmarshal([]byte(xmlString), &p)
```

> When unmarshalling the XML into `Problem` struct,
> the unmarshaller will try it's best to create the extensions.
> Expect all values to be of a `string` type.

### Content-Type header

The final response should have a `Content-Type: application/problem+json` or
`Content-Type: application/problem+xml` header:

<pre>
HTTP/1.1 404 Not Found
<b style="color:green">Content-Type: application/problem+json</b>
Vary: Origin
Date: Sat, 01 Jan 1970 17:21:18 GMT
Content-Length: 87

{"detail":"No such API route","instance":"/foo","status":404,"title":"Route Not Found"}
</pre>

### Cache-Control header

The response **will not have** a `Cache-Control: no-cache, no-store, must-revalidate` header
according to the [RFC-7231][RFC7231] for HTTP/1.1, unless otherwise explicitly set.

### Thread-satey considerations

`UnmarshalJSON` and `UnmarshalXML` modifies the receiver, so the concurrent calls
with the same receiver would be unsafe.

> - `Problem` instances should not be modified after creation
> - Each request should use its own `Problem` instance
> - `UnmarshalJSON` and `UnmarshalXML` should not be called concurrently on the same instance

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


[RFC9457]: https://tools.ietf.org/html/rfc9457
[RFC7231]: https://tools.ietf.org/html/rfc7231
