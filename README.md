httpeasy
--------

`httpeasy` is an easy-peasy HTTP framework for Go. It's designed to be quite a
bit easier to use than the standard library's `net/http` framework without
compromising performance. It is deliberately less fully-featured than
`net/http`.

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
	if err := http.ListenAndServe(":8080", NewRouter().Register(
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
	)); err != nil {
		log.Fatal(err)
	}
}
```
