package testsupport

import (
	"bytes"
	"fmt"

	"github.com/weberc2/httpeasy"
)

type WantedData interface {
	CompareData([]byte) error
}

func ReadAll(s httpeasy.Serializer) ([]byte, error) {
	writerTo, err := s()
	if err != nil {
		return nil, fmt.Errorf("serializing: %w", err)
	}

	var buf bytes.Buffer
	if _, err := writerTo.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("serializing to buffer: %w", err)
	}

	return buf.Bytes(), nil
}

type WantedSerializer struct {
	WantedData
}

type CompareSerializerError struct {
	Data []byte
	Err  error
}

func (err *CompareSerializerError) Error() string { return err.Err.Error() }

func (wanted WantedSerializer) CompareSerializer(
	found httpeasy.Serializer,
) error {
	data, err := ReadAll(found)
	if err != nil {
		return err
	}

	if err := wanted.CompareData(data); err != nil {
		return &CompareSerializerError{data, err}
	}

	return nil
}

func CompareSerializer(wanted WantedData, found httpeasy.Serializer) error {
	return WantedSerializer{wanted}.CompareSerializer(found)
}
