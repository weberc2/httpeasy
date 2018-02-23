package httpeasy

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
)

type Response struct {
	Status int
	Data   []byte
	Log    interface{}
}

func Ok(out interface{}, log interface{}) Response {
	data, err := json.Marshal(out)
	if err != nil {
		return InternalServerError(struct {
			Reason                 string      `json:"reason"`
			Err                    error       `json:"error"`
			OriginalLog            interface{} `json:"original_log"`
			StringifiedOriginalOut string      `json:"stringified_original_out"`
		}{
			Reason:                 "Error marshaling data in httpeasy.Ok()",
			Err:                    err,
			OriginalLog:            log,
			StringifiedOriginalOut: spew.Sdump(out),
		})
	}

	return Response{
		Status: http.StatusOK,
		Data:   data,
		Log:    log,
	}
}

func InternalServerError(log interface{}) Response {
	return Response{
		Status: http.StatusInternalServerError,
		Data:   []byte("500 Internal Server Error"),
		Log:    log,
	}
}

// NotFound is a convenience function for constructing HTTP 404 Not Found
// responses. Its `text` argument will be rendered to the client while its
// `log` argument will be logged.
func NotFound(text string, log interface{}) Response {
	return Response{
		Status: http.StatusNotFound,
		Data:   []byte(text),
		Log:    log,
	}
}

// Log represents a standard HTTP log
type Log struct {
	Started    time.Time     `json:"started"`
	Duration   time.Duration `json:"duration"`
	Method     string        `json:"method"`
	URL        url.URL       `json:"url"`
	Status     int           `json:"status"`
	Message    interface{}   `json:"message"`
	WriteError interface{}   `json:"write_error"`
}

// Request represents a simplified HTTP request
type Request struct {
	Vars    map[string]string
	Body    []byte
	Headers http.Header
}

// Handler handles HTTP requests
type Handler func(r Request) Response

// LogFunc logs its arguments
type LogFunc func(vs ...interface{})

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

		w.WriteHeader(rsp.Status)
		_, err = w.Write(rsp.Data)

		log(Log{
			Started:    start,
			Duration:   time.Since(start),
			Method:     r.Method,
			URL:        *r.URL,
			Status:     rsp.Status,
			Message:    rsp.Log,
			WriteError: err,
		})
	}
}

// Route holds the complete routing information
type Route struct {
	Method  string
	Path    string
	Handler Handler
}

// Register a route with the provided mux.Router
func RegisterRoute(muxr *mux.Router, log LogFunc, r Route) {
	muxr.Path(r.Path).Methods(r.Method).HandlerFunc(r.Handler.HTTP(log))
}
