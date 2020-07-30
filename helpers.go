package httpeasy

import "net/http"

// Ok is a convenience function for building HTTP 200 OK responses.
func Ok(data Serializer, logging ...interface{}) Response {
	if data == nil {
		data = String("200 OK")
	}
	return Response{Status: http.StatusOK, Data: data, Logging: logging}
}

// Created is a convenience function for building HTTP 201 responses. If data
// is nil, a default serializer will be used.
func Created(data Serializer, logging ...interface{}) Response {
	if data == nil {
		data = String("201 Created")
	}
	return Response{Status: http.StatusCreated, Data: data, Logging: logging}
}

// NoContent is a convenience function for building HTTP 204 No Content
// responses.
func NoContent(logging ...interface{}) Response {
	return Response{
		Status:  http.StatusNoContent,
		Data:    Bytes(nil),
		Logging: logging,
	}
}

// TemporaryRedirect is a convenience function for building HTTP 307 Temporary
// Redirect responses. It takes no data argument because there isn't much point
// in custom status text for a redirect response. Instead, it takes a URL that
// will be used as the Location header, which should be used by clients as the
// redirect location.
func TemporaryRedirect(location string, logging ...interface{}) Response {
	return Response{
		Status:  http.StatusTemporaryRedirect,
		Data:    String("307 Temporary Redirect"),
		Logging: logging,
		Headers: http.Header{"Location": []string{location}},
	}
}

// BadRequest is a convenience function for building HTTP 400 Bad Request
// responses. If data is nil, a default serializer will be used.
func BadRequest(data Serializer, logging ...interface{}) Response {
	if data == nil {
		data = String("400 Bad Request")
	}
	return Response{
		Status:  http.StatusBadRequest,
		Data:    data,
		Logging: logging,
	}
}

// Unauthorized is a convenience function for building HTTP 401 Unauthorized
// responses. If data is nil, a default serializer will be used.
func Unauthorized(data Serializer, logging ...interface{}) Response {
	if data == nil {
		data = String("401 Unauthorized")
	}
	return Response{
		Status:  http.StatusUnauthorized,
		Data:    data,
		Logging: logging,
	}
}

// NotFound is a convenience function for building HTTP 404 Not Found
// responses. If data is nil, a default serializer will be used.
func NotFound(data Serializer, logging ...interface{}) Response {
	if data == nil {
		data = String("404 Not Found")
	}
	return Response{Status: http.StatusNotFound, Data: data, Logging: logging}
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
