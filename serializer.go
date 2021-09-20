package httpeasy

import (
	"bytes"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"strings"
	text "text/template"

	"github.com/davecgh/go-spew/spew"
)

// Serializer is a type of functions which can return either an io.WriterTo or
// an error. This abstraction helps to make this package ergonomic, for
// example:
//
//     return Ok(String("Hello, world!"))
//
// OR:
//
//     return Ok(Bytes(data))
//
// OR:
//
//     return Ok(JSON(Person{Name: "Bob", Age: 58}))
//
type Serializer func() (io.WriterTo, error)

// String wraps a string in a serializer.
func String(s string) Serializer {
	return func() (io.WriterTo, error) { return strings.NewReader(s), nil }
}

// Stringf formats a string and wraps it in a serializer. Conceptually, it's
// `String(fmt.Sprintf(format, vs...)))`.
func Stringf(format string, vs ...interface{}) Serializer {
	return String(fmt.Sprintf(format, vs...))
}

// Bytes wraps a byte slice in a serializer. The returned serializer always
// succeeds.
func Bytes(bs []byte) Serializer {
	return func() (io.WriterTo, error) { return bytes.NewReader(bs), nil }
}

type reader struct {
	r io.Reader
}

func (r reader) WriteTo(w io.Writer) (int64, error) {
	return io.CopyBuffer(w, r.r, make([]byte, 4096))
}

// Reader wraps an `io.Reader` in a serializer. The returned serializer always
// succeeds.
func Reader(r io.Reader) Serializer {
	return func() (io.WriterTo, error) { return reader{r}, nil }
}

// Sprint wraps N values in a serializer. Its serialization mechanism is
// `fmt.Sprint`. This is probably just useful for debugging. The returned
// serializer always succeeds.
func Sprint(vs ...interface{}) Serializer {
	return func() (io.WriterTo, error) {
		return strings.NewReader(fmt.Sprint(vs...)), nil
	}
}

// JSON wraps a value in a JSON serializer. The returned serializer will only
// fail if the value isn't JSON serializable.
func JSON(v interface{}) Serializer {
	return func() (io.WriterTo, error) {
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(data), nil
	}
}

// Debug wraps a series of values in a serializer. The serialization mechanism
// is github.com/davecgh/go-spew/spew.Sdump(). The returned serializer always
// succeeds.
func Debug(vs ...interface{}) Serializer { return String(spew.Sdump(vs...)) }

// HTMLTemplate takes an html/template.Template and some data and returns a
// serializer. The serializer will execute the template with the data and
// return any errors it encounters. See examples/hello.go for an example.
func HTMLTemplate(t *html.Template, v interface{}) Serializer {
	return func() (io.WriterTo, error) {
		var buf bytes.Buffer
		err := t.Execute(&buf, v)
		return &buf, err
	}
}

// TextTemplate takes a text/template.Template and some data and returns a
// serializer. The serializer will execute the template with the data and
// return any errors it encounters. See the HTMLTemplate() example in
// examples/hello.go for an analogous example.
func TextTemplate(t *text.Template, v interface{}) Serializer {
	return func() (io.WriterTo, error) {
		var buf bytes.Buffer
		err := t.Execute(&buf, v)
		return &buf, err
	}
}
