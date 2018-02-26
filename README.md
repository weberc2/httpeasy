httpeasy
--------

[![GoDoc](https://godoc.org/github.com/weberc2/httpeasy?status.svg)](https://godoc.org/github.com/weberc2/httpeasy)
[![Go Report Card](https://goreportcard.com/badge/github.com/weberc2/httpeasy)](https://goreportcard.com/report/github.com/weberc2/httpeasy)

`httpeasy` is an easy-peasy HTTP framework for Go. It's designed to be quite a
bit easier to use than the standard library's `net/http` framework without
compromising performance. It is deliberately less fully-featured than
`net/http`.

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
					struct{ Name string }{r.Vars["name"]}))
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
