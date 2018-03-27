package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	. "github.com/weberc2/httpeasy"
)

func TestHelpers(t *testing.T) {
	testCases := []struct {
		Name   string
		Actual Response
		Wanted Response
	}{{
		Name:   "ok",
		Actual: Ok(String("200 OK")),
		Wanted: Response{
			Status:  200,
			Data:    String("200 OK"),
			Logging: nil,
		},
	}, {
		Name:   "ok-with-logging",
		Actual: Ok(String("200 OK"), "foo", 1, nil),
		Wanted: Response{
			Status:  200,
			Data:    String("200 OK"),
			Logging: []interface{}{"foo", 1, nil},
		},
	}, {
		Name:   "created",
		Actual: Created(String("201 CREATED")),
		Wanted: Response{
			Status:  201,
			Data:    String("201 CREATED"),
			Logging: nil,
		},
	}, {
		Name:   "created-with-logging",
		Actual: Created(String("201 CREATED"), "some", "logging"),
		Wanted: Response{
			Status:  201,
			Data:    String("201 CREATED"),
			Logging: []interface{}{"some", "logging"},
		},
	}, {
		Name:   "created-nil-data",
		Actual: Created(nil),
		Wanted: Response{
			Status:  201,
			Data:    String("201 Created"),
			Logging: nil,
		},
	}, {
		Name:   "no-content",
		Actual: NoContent(),
		Wanted: Response{
			Status:  204,
			Data:    String(""),
			Logging: nil,
		},
	}, {
		Name:   "no-content-with-logging",
		Actual: NoContent("foo", 1, "bar", 2),
		Wanted: Response{
			Status:  204,
			Data:    String(""),
			Logging: []interface{}{"foo", 1, "bar", 2},
		},
	}, {
		Name:   "temporary-redirect",
		Actual: TemporaryRedirect("http://google.com"),
		Wanted: Response{
			Status:  307,
			Data:    String("307 Temporary Redirect"),
			Logging: nil,
			Headers: http.Header{"Location": []string{"http://google.com"}},
		},
	}, {
		Name:   "temporary-redirect-with-logging",
		Actual: TemporaryRedirect("http://yahoo.com", "logs", "go", "here"),
		Wanted: Response{
			Status:  307,
			Data:    String("307 Temporary Redirect"),
			Logging: []interface{}{"logs", "go", "here"},
			Headers: http.Header{"Location": []string{"http://yahoo.com"}},
		},
	}, {
		Name:   "bad-request",
		Actual: BadRequest(String("400 BAD REQUEST")),
		Wanted: Response{
			Status:  400,
			Data:    String("400 BAD REQUEST"),
			Logging: nil,
		},
	}, {
		Name:   "bad-request-nil-data",
		Actual: BadRequest(nil),
		Wanted: Response{
			Status:  400,
			Data:    String("400 Bad Request"),
			Logging: nil,
		},
	}, {
		Name:   "bad-request-with-logging",
		Actual: BadRequest(nil, "x", "y", "z"),
		Wanted: Response{
			Status:  400,
			Data:    String("400 Bad Request"),
			Logging: []interface{}{"x", "y", "z"},
		},
	}, {
		Name:   "unauthorized",
		Actual: Unauthorized(String("401 UNAUTHORIZED")),
		Wanted: Response{
			Status:  401,
			Data:    String("401 UNAUTHORIZED"),
			Logging: nil,
		},
	}, {
		Name:   "unauthorized-nil-data",
		Actual: Unauthorized(nil),
		Wanted: Response{
			Status:  401,
			Data:    String("401 Unauthorized"),
			Logging: nil,
		},
	}, {
		Name:   "unauthorized-with-logging",
		Actual: Unauthorized(nil, "some", "logs", "here"),
		Wanted: Response{
			Status:  401,
			Data:    String("401 Unauthorized"),
			Logging: []interface{}{"some", "logs", "here"},
		},
	}, {
		Name:   "not-found",
		Actual: NotFound(String("404 NOT FOUND")),
		Wanted: Response{
			Status:  404,
			Data:    String("404 NOT FOUND"),
			Logging: nil,
		},
	}, {
		Name:   "not-found-nil-data",
		Actual: NotFound(nil),
		Wanted: Response{
			Status:  404,
			Data:    String("404 Not Found"),
			Logging: nil,
		},
	}, {
		Name:   "not-found-with-logging",
		Actual: NotFound(nil, "logging"),
		Wanted: Response{
			Status:  404,
			Data:    String("404 Not Found"),
			Logging: []interface{}{"logging"},
		},
	}, {
		Name:   "internal-server-error",
		Actual: InternalServerError(),
		Wanted: Response{
			Status:  500,
			Data:    String("500 Internal Server Error"),
			Logging: nil,
		},
	}, {
		Name:   "internal-server-error-with-logging",
		Actual: InternalServerError("logging", "goes", "here"),
		Wanted: Response{
			Status:  500,
			Data:    String("500 Internal Server Error"),
			Logging: []interface{}{"logging", "goes", "here"},
		},
	}}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			if err := compareResponses(
				testCase.Wanted,
				testCase.Actual,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func compareResponses(wanted, actual Response) error {
	if wanted.Status != actual.Status {
		return fmt.Errorf(
			"Response status mismatch; wanted '%d', got '%d'",
			wanted.Status,
			actual.Status,
		)
	}

	if err := compareHeaders(wanted.Headers, actual.Headers); err != nil {
		return err
	}

	// NOTE: Validating by marshaling to JSON and comparing the resulting
	// bytes. This is probably fragile, so we'll need to find something better
	// going forward.
	wantedLoggingData, err := json.MarshalIndent(wanted.Logging, "", "    ")
	if err != nil {
		return fmt.Errorf(
			"Unexpected error marshalling 'wanted' logging: '%v'",
			err,
		)
	}
	actualLoggingData, err := json.MarshalIndent(actual.Logging, "", "    ")
	if err != nil {
		return fmt.Errorf(
			"Unexpected error marshalling 'actual' logging: '%v'",
			err,
		)
	}

	if bytes.Compare(wantedLoggingData, actualLoggingData) != 0 {
		return fmt.Errorf(
			"Wanted logging:\n%s\n\nGot logging:\n%s",
			wantedLoggingData,
			actualLoggingData,
		)
	}

	return compareSerializers(wanted.Data, actual.Data)
}

func compareHeaders(wanted, actual http.Header) error {
	if len(wanted) != len(actual) {
		return fmt.Errorf(
			"Wanted headers:\n%s\n\nGot headers:\n%s",
			stringHeaders(wanted),
			stringHeaders(actual),
		)
	}
	for header, values := range wanted {
		if otherValues, found := actual[header]; found {
			if len(values) != len(otherValues) {
				return fmt.Errorf(
					"Wanted values for header '%s': [%s], got [%s]",
					header,
					strings.Join(values, ", "),
					strings.Join(otherValues, ", "),
				)
			}
			for i, value := range values {
				if value != otherValues[i] {
					return fmt.Errorf(
						"Wanted values for header '%s': [%s], got [%s]",
						header,
						strings.Join(values, ", "),
						strings.Join(otherValues, ", "),
					)
				}
			}
		} else {
			return fmt.Errorf("Missing header '%s'", header)
		}
	}
	return nil
}

func stringHeaders(headers http.Header) string {
	var buf bytes.Buffer
	for header, values := range headers {
		fmt.Fprintf(&buf, "%s: %s\n", header, strings.Join(values, ", "))
	}
	return buf.String()
}

func compareSerializers(wanted, actual Serializer) error {
	wantedData, err := readSerializer(wanted)
	if err != nil {
		return fmt.Errorf("Error reading 'wanted' serializer: %v", err)
	}
	actualData, err := readSerializer(actual)
	if err != nil {
		return fmt.Errorf("Error reading 'actual' serializer: %v", err)
	}
	if bytes.Compare(wantedData, actualData) != 0 {
		return fmt.Errorf(
			"Wanted serialized data:\n%s\n\nGot serialized data:\n%s",
			wantedData,
			actualData,
		)
	}
	return nil
}

func readSerializer(s Serializer) ([]byte, error) {
	writerTo, err := s()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	writerTo.WriteTo(&buf)
	return buf.Bytes(), nil
}
