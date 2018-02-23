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
	Vars    map[string]string
	Body    []byte
	Headers http.Header
}

// Response represents a simplifiied HTTP response
type Response struct {
	// Status is the HTTP status code to write
	Status int

	// Data is the data to be written to the client
	Data Serializer

	// Logging is the information to pass to the logger
	Logging []interface{}
}

// Ok is a convenience function for building HTTP 200 OK responses.
func Ok(data Serializer, logging ...interface{}) Response {
	return Response{Status: http.StatusOK, Data: data, Logging: logging}
}

// InternalServerError is a convenience function for building HTTP 500 Internal
// Server Error responses.
func InternalServerError(logging ...interface{}) Response {
	return Response{
		Status:  http.StatusInternalServerError,
		Data:    String("500 Internal Server Error"),
		Logging: logging,
	}
}

// NotFound is a convenience function for building HTTP 404 Not Found
// responses.
func NotFound(data Serializer, logging ...interface{}) Response {
	return Response{Status: http.StatusNotFound, Data: data, Logging: logging}
}

// Log represents a standard HTTP request log
type RequestLog struct {
	// Started holds the start time for the request
	Started time.Time `json:"started"`

	// Duration holds the duration to service the request
	Duration time.Duration `json:"duration"`

	// Method holds the HTTP method for the request
	Method string `json:"method"`

	// URL holds the URL for the request
	URL url.URL `json:"url"`

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

		var rsp Response

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rsp = InternalServerError(struct {
				Context string `json:"context"`
				Error   string `json:"error"`
			}{Context: "Failed to read request body", Error: err.Error()})
		} else {
			rsp = h(Request{Vars: mux.Vars(r), Body: data, Headers: r.Header})
		}

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

		log(RequestLog{
			Started:    start,
			Duration:   time.Since(start),
			Method:     r.Method,
			URL:        *r.URL,
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

// Register registers routes with the provided Router and returns the same
// modified Router.
func (r *Router) Register(log LogFunc, routes ...Route) *Router {
	for _, route := range routes {
		r.inner.Path(route.Path).
			Methods(route.Method).
			HandlerFunc(route.Handler.HTTP(log))
	}
	return r
}
