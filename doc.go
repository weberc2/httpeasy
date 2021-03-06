// Package httpeasy provides an easy-peasy API for building HTTP servers.
//
// Out of the box it provides detailed request logging and a variety of
// serializers (json, string, html template, etc) for rendering data as well as
// convenient functions for working with requests and responses.
//
// You can run the following example program yourself. Run the server with
// `go run ./examples/hello.go` and then, from another terminal, `curl
// localhost:8080/(html|plaintext|json)/{name}`:
//
//     package main
//
//     import (
//         html "html/template"
//         "log"
//         "net/http"
//         "os"
//
//         . "github.com/weberc2/httpeasy"
//     )
//
//     func main() {
//         log.Println("Listening at :8080")
//         if err := http.ListenAndServe(":8080", Register(
//             JSONLog(os.Stderr),
//             Route{
//                 Path:   "/plaintext/{name}",
//                 Method: "GET",
//                 Handler: func(r Request) Response {
//                     return Ok(String("Hello, " + r.Vars["name"] + "!"))
//                 },
//             },
//             Route{
//                 Path:   "/json/{name}",
//                 Method: "GET",
//                 Handler: func(r Request) Response {
//                     return Ok(JSON(struct {
//                         Greeting string `json:"greeting"`
//                     }{Greeting: "Hello, " + r.Vars["name"] + "!"}))
//                 },
//             },
//             Route{
//                 Path:   "/html/{name}",
//                 Method: "GET",
//                 Handler: func(r Request) Response {
//                     return Ok(HTMLTemplate(
//                         html.Must(html.New("greeting.html").Parse(
//                             `<html>
//                                 <body>
//                                     <h1>Hello, {{.Name}}</h1>
//                                 </body>
//                             </html>`,
//                         )),
//                         struct{ Name string }{r.Vars["name"]}))
//                 },
//             },
//             Route{
//                 Path:   "/error",
//                 Method: "GET",
//                 Handler: func(r Request) Response {
//                     return InternalServerError("Error details...")
//                 },
//             },
//         )); err != nil {
//             log.Fatal(err)
//         }
//     }
package httpeasy
