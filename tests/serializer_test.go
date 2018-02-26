package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	. "github.com/weberc2/httpeasy"
)

func TestSerializer(t *testing.T) {
	testCases := []struct {
		Name         string
		Serializer   Serializer
		WantedOutput string
		WantedError  func(err error) error
	}{{
		Name:         "string",
		Serializer:   String("Hello"),
		WantedOutput: "Hello",
	}, {
		Name:         "bytes",
		Serializer:   Bytes([]byte("Lorem ipsum")),
		WantedOutput: "Lorem ipsum",
	}, {
		Name:         "sprint",
		Serializer:   Sprint(struct{ Foo int }{2}),
		WantedOutput: fmt.Sprint(struct{ Foo int }{2}),
	}, {
		Name: "json",
		Serializer: JSON(struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}{"Bob", 54}),
		WantedOutput: `{"name":"Bob","age":54}`,
	}, {
		Name:       "json-marshal-error",
		Serializer: JSON(marshalErrorer{}),
		WantedError: func(err error) error {
			if err, ok := err.(*json.MarshalerError); ok {
				if err.Err != sentinelErr {
					return fmt.Errorf(
						"Expected the sentinel error; got '%v'",
						err.Err,
					)
				}
				return nil
			}
			return fmt.Errorf("Expected json.MarshalerError; got '%v'", err)
		},
	}}

	var buf bytes.Buffer
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			buf.Reset()
			writerTo, err := testCase.Serializer()
			if err != nil {
				if testCase.WantedError != nil {
					if err := testCase.WantedError(err); err != nil {
						t.Fatal(err)
					}
					return
				}
				t.Fatal("Unexpected error:", err)
			}
			if testCase.WantedError != nil {
				t.Fatal("Expected an error but got none")
			}
			if _, err := writerTo.WriteTo(&buf); err != nil {
				t.Fatal("Unexpected error writing to buffer:", err)
			}

			if testCase.WantedOutput != buf.String() {
				t.Fatalf(
					"Wanted output:\n%s\n\nGot output: %s",
					testCase.WantedOutput,
					buf.String(),
				)
			}
		})
	}
}

type marshalErrorer struct{}

var sentinelErr = errors.New("Sentinel error")

func (me marshalErrorer) MarshalJSON() ([]byte, error) {
	return nil, sentinelErr
}
