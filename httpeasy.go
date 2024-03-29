package httpeasy

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
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

	// URL contains the parsed URL information. See net/http.Request.URL for
	// more information.
	URL *url.URL
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

// Cookie returns the named cookie if it exists, otherwise http.ErrNoCookie.
func (r Request) Cookie(name string) (*http.Cookie, error) {
	for _, c := range readCookies(r.Headers, name) {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, http.ErrNoCookie
}

// Cookies returns the cookies attached to the request.
func (r Request) Cookies() []*http.Cookie {
	return readCookies(r.Headers, "")
}

// InvalidJSONErr wraps an error encountered while trying to unmarshal JSON.
type InvalidJSONErr struct {
	Err error
}

// Error implements the error interface for InvalidJSONErr
func (err InvalidJSONErr) Error() string {
	return fmt.Sprintf("Invalid JSON: %v", err.Err)
}

// JSON deserializes the request body into `v`. `v` must be a pointer; all the
// standard `encoding/json.Unmarshal()` rules apply. If an error is encountered
// while unmarshaling, `InvalidJSONErr` is returned to distinguish it from
// errors encountered while reading the request body.
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
	if err := json.Unmarshal(data, v); err != nil {
		return InvalidJSONErr{err}
	}
	return nil
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

	// Cookies are the list of cookies to be set on the response
	Cookies []*http.Cookie
}

// WithHeaders returns a copy of the response with the specified headers
// attached. The headers provided by this method will be merged with any
// existing headers on the response.
func (r Response) WithHeaders(headers http.Header) Response {
	if r.Headers == nil {
		r.Headers = http.Header{}
	}
	for key, values := range headers {
		r.Headers[key] = append(r.Headers[key], values...)
	}
	return r
}

// WithCookies returns a copy of the response with the specified cookies
// attached. The cookies provided by this method will be appended onto any
// existing cookies on the response.
func (r Response) WithCookies(cookies ...*http.Cookie) Response {
	r.Cookies = append(r.Cookies, cookies...)
	return r
}

// WithLogging returns a copy of the response with the specified logging
// attached. The logging provided by this method will be appended onto any
// existing cookies on the response.
func (r Response) WithLogging(logging ...interface{}) Response {
	r.Logging = append(r.Logging, logging...)
	return r
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

	// RequestHeaders holds the headers for the request
	RequestHeaders http.Header `json:"requestHeaders"`

	// ResponseHeaders holds the headers for the request
	ResponseHeaders http.Header `json:"responseHeaders"`

	// Status holds the HTTP status code returned by the request handler
	Status int `json:"status"`

	// Message holds the logging message returned by the request handler
	Message []interface{} `json:"message"`

	// WriteError holds any errors that were encountered during writing the
	// response to the output socket.
	WriteError interface{} `json:"writeError"`
}

// JSONLog returns a `LogFunc` which logs JSON to `w`.
func JSONLog(w io.Writer) LogFunc {
	return func(v interface{}) {
		data, err := json.Marshal(v)
		if err != nil {
			data, err = json.Marshal(struct {
				Context   string `json:"context"`
				DebugData string `json:"debugData"`
				Error     string `json:"error"`
			}{
				Context:   "Error marshaling 'debugData' into JSON",
				DebugData: spew.Sdump(v),
				Error:     err.Error(),
			})
			if err != nil {
				// We really REALLY should never get here
				log.Println("ERROR MARSHALLING THE MARSHALLING ERROR!:", err)
				return
			}
		}
		if _, err := fmt.Fprintf(w, "%s\n", data); err != nil {
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
		contentLength := r.Header.Get("Content-Length")
		i, err := strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			i = 0 // just being explicit
			rsp.Logging = append(
				rsp.Logging,
				struct {
					Context string `json:"context"`
					Error   string `json:"error"`
				}{
					Context: "Invalid or missing `Content-Length` header; " +
						"defaulting to Content-Length=0",
					Error: err.Error(),
				},
			)
		}
		rsp = h(Request{
			Vars:    mux.Vars(r),
			Body:    io.LimitReader(r.Body, i),
			Headers: r.Header,
			URL:     r.URL,
		})

		writerTo, err := rsp.Data()
		if err != nil {
			rsp.Status = http.StatusInternalServerError
			writerTo = strings.NewReader("500 Internal Server Error")
			rsp.Logging = []interface{}{
				struct {
					Context         string        `json:"context"`
					Error           string        `json:"error"`
					OriginalLogging []interface{} `json:"originalLogging"`
				}{
					Context:         "Error serializing response data",
					Error:           err.Error(),
					OriginalLogging: rsp.Logging,
				},
			}
		}

		// Copy HTTP headers from the response object to the response writer.
		// This has to go before the WriteHeader invocation or it won't take
		// effect (quirk of net/http.ResponseWriter).
		header := w.Header()
		for key, values := range rsp.Headers {
			for _, value := range values {
				header.Add(key, value)
			}
		}

		for _, cookie := range rsp.Cookies {
			http.SetCookie(w, cookie)
		}

		w.WriteHeader(rsp.Status)
		_, err = writerTo.WriteTo(w)

		log(requestLog{
			Started:         start,
			Duration:        time.Since(start),
			Method:          r.Method,
			URL:             *r.URL,
			RequestHeaders:  r.Header,
			ResponseHeaders: w.Header(),
			Status:          rsp.Status,
			Message:         rsp.Logging,
			WriteError:      err,
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

// StdlibRoute holds the complete routing information. It is the same as a
// `Route` except that the handler type is an `http.HandlerFunc` instead of a
// `Handler`.
type StdlibRoute struct {
	// Method is the HTTP method for the route
	Method string

	// Path is the path to the handler. See github.com/gorilla/mux.Route.Path
	// for additional details.
	Path string

	// Handler is the function which handles the request
	Handler http.HandlerFunc
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

// RegisterStdlib registers `StdlibRoute`s with the provided Router and returns
// the same modified Router.
func (r *Router) RegisterStdlib(routes ...StdlibRoute) *Router {
	for _, route := range routes {
		r.inner.Path(route.Path).
			Methods(route.Method).
			HandlerFunc(route.Handler)
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
