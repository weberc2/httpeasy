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
	Cause_  error  `json:"cause,omitempty"`
}

func (err *HTTPError) Cause() error { return err.Cause_ }

func (err *HTTPError) HTTPError() *HTTPError { return err }

func (err *HTTPError) Error() string {
	if err.Cause_ == nil {
		return err.Message
	}
	return fmt.Sprintf("%s: %s", err.Message, err.Cause_)
}

func (err *HTTPError) Compare(other *HTTPError) error {
	if err == other {
		return nil
	}

	if err != nil && other == nil {
		return fmt.Errorf("wanted not-nil; found `nil`")
	}

	if err == nil && other != nil {
		return fmt.Errorf("wanted `nil`; found not-nil")
	}

	if err.Status != other.Status {
		return fmt.Errorf(
			"HTTPError.Status: wanted `%d`; found `%d`",
			err.Status,
			other.Status,
		)
	}

	if err.Message != other.Message {
		return fmt.Errorf(
			"HTTPError.Message: wanted `%s`; found `%s`",
			err.Message,
			other.Message,
		)
	}

	if err.Cause_ != nil && other.Cause_ != nil {
		wanted, found := err.Cause_.Error(), other.Cause_.Error()
		if wanted != found {
			return fmt.Errorf(
				"HTTPError.Cause_: wanted `%s`; found `%s`",
				wanted,
				found,
			)
		}
		return nil
	}
	if err.Cause_ == nil && other.Cause_ == nil {
		return nil
	}

	// exactly one of the `Cause_` fields are nil
	return fmt.Errorf(
		"HTTPError.Cause_: wanted `%v`; found `%v`",
		err.Cause_,
		other.Cause_,
	)
}

func (err *HTTPError) CompareErr(other error) error {
	var e *HTTPError
	if !errors.As(other, &e) {
		return fmt.Errorf("wanted `%v`; found `%v`", err, other)
	}
	return err.Compare(e)
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
