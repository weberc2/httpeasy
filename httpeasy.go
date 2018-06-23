package httpeasy

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
)

// Request represents a simplified HTTP request
type Request struct {
	// Vars are the variables parsed out of the URL path.
	Vars map[string]string

	// Body is the contents of the HTTP request body.
	Body io.Reader

	// Headers are the HTTP headers
	Headers http.Header
}

// Text consumes the request body and returns it as a string.
func (r Request) Text() (string, error) {
	data, err := r.Bytes()
	return string(data), err
}

// Bytes consumes the request body and returns it as a byte slice.
func (r Request) Bytes() ([]byte, error) {
	return ioutil.ReadAll(r.Body)
}

// JSON deserializes the request body into `v`. `v` must be a pointer; all the
// standard `encoding/json.Unmarshal()` rules apply.
//
//     var person struct {
//         Name string `json:"name"`
//         Age  int    `json:"age"`
//     }
//     if err := r.JSON(&person); err != nil {
//         return err
//     }
//     fmt.Printf("Name='%s'; Age=%d", person.Name, person.Age)
//
func (r Request) JSON(v interface{}) error {
	data, err := r.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Response represents a simplified HTTP response
type Response struct {
	// Status is the HTTP status code to write
	Status int

	// Data is the data to be written to the client
	Data Serializer

	// Logging is the information to pass to the logger
	Logging []interface{}

	// Headers is the HTTP headers for the response
	Headers http.Header
}

// requestLog represents a standard HTTP request log
type requestLog struct {
	// Started holds the start time for the request
	Started time.Time `json:"started"`

	// Duration holds the duration to service the request
	Duration time.Duration `json:"duration"`

	// Method holds the HTTP method for the request
	Method string `json:"method"`

	// URL holds the URL for the request
	URL url.URL `json:"url"`

	// Headers holds the headers for the request
	Headers http.Header `json:"headers"`

	// Status holds the HTTP status code returned by the request handler
	Status int `json:"status"`

	// Message holds the logging message returned by the request handler
	Message []interface{} `json:"message"`

	// WriteError holds any errors that were encountered during writing the
	// response to the output socket.
	WriteError interface{} `json:"write_error"`
}

// JSONLog returns a `LogFunc` which logs JSON to `w`.
func JSONLog(w io.Writer) LogFunc {
	return func(v interface{}) {
		data, err := json.MarshalIndent(v, "", "    ")
		if err != nil {
			data, err = json.Marshal(struct {
				Context   string `json:"context"`
				DebugData string `json:"debug_data"`
				Error     string `json:"error"`
			}{
				Context:   "Error marshaling 'debug_data' into JSON",
				DebugData: spew.Sdump(v),
				Error:     err.Error(),
			})
			if err != nil {
				// We really REALLY should never get here
				log.Println("ERROR MARSHALLING THE MARSHALLING ERROR!:", err)
				return
			}
		}
		if _, err := w.Write(data); err != nil {
			log.Println("ERROR WRITING TO LOGGER:", err)
		}
	}
}

// Handler handles HTTP requests
type Handler func(r Request) Response

// LogFunc logs its argument
type LogFunc func(v interface{})

// HTTP converts an httpeasy.Handler into an http.HandlerFunc. The returned
// function will collect a bunch of standard HTTP information and pass it to
// the provided log function.
func (h Handler) HTTP(log LogFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer r.Body.Close()

		rsp := h(Request{Vars: mux.Vars(r), Body: r.Body, Headers: r.Header})

		writerTo, err := rsp.Data()
		if err != nil {
			rsp.Status = http.StatusInternalServerError
			writerTo = strings.NewReader("500 Internal Server Error")
			rsp.Logging = []interface{}{
				struct {
					Context         string        `json:"context"`
					Error           string        `json:"error"`
					OriginalLogging []interface{} `json:"original_logging"`
				}{
					Context:         "Error serializing response data",
					Error:           err.Error(),
					OriginalLogging: rsp.Logging,
				},
			}
		}
		w.WriteHeader(rsp.Status)
		_, err = writerTo.WriteTo(w)

		log(requestLog{
			Started:    start,
			Duration:   time.Since(start),
			Method:     r.Method,
			URL:        *r.URL,
			Headers:    r.Header,
			Status:     rsp.Status,
			Message:    rsp.Logging,
			WriteError: err,
		})
	}
}

// Route holds the complete routing information
type Route struct {
	// Method is the HTTP method for the route
	Method string

	// Path is the path to the handler. See github.com/gorilla/mux.Route.Path
	// for additional details.
	Path string

	// Handler is the function which handles the request
	Handler Handler
}

// Router is an HTTP mux for httpeasy.
type Router struct {
	inner *mux.Router
}

// NewRouter constructs a new router.
func NewRouter() *Router { return &Router{mux.NewRouter()} }

// ServeHTTP implements the http.Handler interface for Router.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.inner.ServeHTTP(w, req)
}

// Register registers routes with the provided Router and LogFunc and returns
// the same modified Router.
func (r *Router) Register(log LogFunc, routes ...Route) *Router {
	for _, route := range routes {
		r.inner.Path(route.Path).
			Methods(route.Method).
			HandlerFunc(route.Handler.HTTP(log))
	}
	return r
}

// Register creates a new router and uses it to register all of the provided
// routes before returning it. It's purely a convenience wrapper around
//
//     r := NewRouter()
//     r.Register(log, routes...)
func Register(log LogFunc, routes ...Route) *Router {
	return NewRouter().Register(log, routes...)
}
