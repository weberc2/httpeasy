package testsupport

import (
	"encoding/json"
	"testing"

	"github.com/weberc2/httpeasy"
)

func TestLog(t *testing.T) httpeasy.LogFunc {
	return func(v interface{}) {
		data, err := json.MarshalIndent(v, "", "    ")
		if err != nil {
			t.Logf("error logging: failed to marshal log data: %v", err)
			t.Logf("DATA: %# v", v)
		}
		t.Logf("%s", data)
	}
}
