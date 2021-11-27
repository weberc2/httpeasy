package httpeasy

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Error interface {
	HTTPError() *HTTPError
}

type HTTPError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (err *HTTPError) HTTPError() *HTTPError { return err }

func (err *HTTPError) Error() string { return string(err.Message) }

func (err *HTTPError) Compare(other error) error {
	var e *HTTPError
	if !errors.As(other, &e) {
		return fmt.Errorf("wanted `%v`; found `%v`", err, other)
	}
	return nil
}

func (wanted *HTTPError) CompareData(data []byte) error {
	var other HTTPError
	if err := json.Unmarshal(data, &other); err != nil {
		return fmt.Errorf("unmarshaling `HTTPError`: %w", err)
	}
	return wanted.Compare(&other)
}

func HandleError(message string, err error, logging ...interface{}) Response {
	logging = append(logging, struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}{
		Message: message,
		Error:   err.Error(),
	})

	cause := err
	for {
		if unwrapped := errors.Unwrap(cause); unwrapped != nil {
			cause = unwrapped
			continue
		}
		break
	}

	if e, ok := cause.(Error); ok {
		httpErr := e.HTTPError()
		return Response{
			Status:  httpErr.Status,
			Data:    JSON(httpErr),
			Logging: logging,
		}
	}

	return InternalServerError(logging)
}