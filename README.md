httpeasy
--------

[![GoDoc](https://godoc.org/github.com/weberc2/httpeasy?status.svg)](https://godoc.org/github.com/weberc2/httpeasy)

`httpeasy` is an easy-peasy HTTP framework for Go. It's designed to be quite a
bit easier to use than the standard library's `net/http` framework without
compromising performance. It is deliberately less fully-featured than
`net/http`.

`httpeasy` provides complete request logging out of the box.

## Installation

`go get -u github.com/weberc2/httpeasy`

## Usage

```go
package main

import (
	"log"
	"net/http"
	"os"

	. "github.com/weberc2/httpeasy"
)

func main() {
	log.Println("Listening at :8080")
	if err := http.ListenAndServe(":8080", Register(
		JSONLog(os.Stderr),
		Route{
			Path:   "/plaintext",
			Method: "GET",
			Handler: func(r Request) Response {
				return Ok(String("Hello, world!"))
			},
		},
		Route{
			Path:   "/json",
			Method: "GET",
			Handler: func(r Request) Response {
				return Ok(JSON(struct {
					Greeting string `json:"greeting"`
				}{Greeting: "Hello, world!"}))
			},
		},
		Route{
			Path:   "/error",
			Method: "GET",
			Handler: func(r Request) Response {
				return InternalServerError("Error details...")
			},
		},
	)); err != nil {
		log.Fatal(err)
	}
}
```
