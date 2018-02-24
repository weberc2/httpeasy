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
