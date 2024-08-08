# Problem

## Problem details for HTTP APIs per [RFC-9457][RFC9457] standard.

This module provides the `Problem` struct which can be used to represent a problem
in HTTP APIs. It implements the [RFC-9457][RFC9457] standard.

It supports serializing and deserializing the `Problem` struct to and from JSON.

## Usage

```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/kodeart/go-problem"
)

func main() {
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
		}
	}

	body, _ := json.Marshal(&p)
	// send the response body to the client
	// ...
}
```


[RFC9457]: https://tools.ietf.org/html/rfc9457
