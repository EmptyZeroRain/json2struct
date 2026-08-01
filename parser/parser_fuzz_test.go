package parser

import (
	"bytes"
	"testing"
)

func FuzzJSONParser(f *testing.F) {
	f.Add([]byte(`{"a":[1,true,null]}`))
	f.Add([]byte(`null`))
	f.Fuzz(func(t *testing.T, data []byte) { _, _ = (JSONParser{}).Parse(bytes.NewReader(data)) })
}
