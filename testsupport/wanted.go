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

func (wanted WantedSerializer) CompareSerializer(
	found httpeasy.Serializer,
) error {
	data, err := ReadAll(found)
	if err != nil {
		return err
	}

	return wanted.CompareData(data)
}
