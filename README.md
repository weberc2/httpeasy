httpeasy
--------

[![GoDoc](https://godoc.org/github.com/weberc2/httpeasy?status.svg)](https://godoc.org/github.com/weberc2/httpeasy)
[![Go Report Card](https://goreportcard.com/badge/github.com/weberc2/httpeasy)](https://goreportcard.com/report/github.com/weberc2/httpeasy)
[![Coverage Status](https://coveralls.io/repos/github/weberc2/httpeasy/badge.svg?branch=master)](https://coveralls.io/github/weberc2/httpeasy?branch=master)

`httpeasy` is an easy-peasy HTTP library for Go. It aims to jump start your
web development projects and generally be simpler, easier, and more ergonomic
than `net/http` and other libraries.

`httpeasy` provides complete request logging out of the box as well as a suite
of helper functions for operating on requests and responses, including a suite
of utilities for working with requests...

* `Request.Bytes()`
* `Request.Text()`,
* `Request.JSON()`
* `Request.Vars`

...for serializing data...

* `JSON()`
* `String()`
* `HTMLTemplate()`
* etc

...and for creating responses...

* `Ok()`
* `InternalServerError()`
* `NotFound()`
* `BadRequest()`

## Installation

`go get -u github.com/weberc2/httpeasy`

## Testing

Just like any Go project, `go test ./...` (from the project root directory)

## Usage

You can run the following program yourself. Run the server with
`go run ./examples/hello.go` and then, from another terminal, `curl
localhost:8080/(html|plaintext|json)/{name}`.

```go
package main

import (
	html "html/template"
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
			Path:   "/plaintext/{name}",
			Method: "GET",
			Handler: func(r Request) Response {
				return Ok(String("Hello, " + r.Vars["name"] + "!"))
			},
		},
		Route{
			Path:   "/json/{name}",
			Method: "GET",
			Handler: func(r Request) Response {
				return Ok(JSON(struct {
					Greeting string `json:"greeting"`
				}{Greeting: "Hello, " + r.Vars["name"] + "!"}))
			},
		},
		Route{
			Path:   "/html/{name}",
			Method: "GET",
			Handler: func(r Request) Response {
				return Ok(HTMLTemplate(
					html.Must(html.New("greeting.html").Parse(
						`<html>
							<body>
								<h1>Hello, {{.Name}}</h1>
							</body>
						</html>`,
					)),
					struct{ Name string }{r.Vars["name"]},
                ))
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
