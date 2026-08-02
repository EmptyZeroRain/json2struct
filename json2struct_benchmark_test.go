package json2struct_test

import (
	"bytes"
	json2struct "github.com/EmptyZeroRain/json2struct"
	"testing"
)

func BenchmarkLargeJSON(b *testing.B) {
	data := bytes.Repeat([]byte(`{"value":"sample"}`), 1024)
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		g := json2struct.New(json2struct.Options{Name: "Sample"})
		if err := g.AddJSON(data); err == nil {
			_, _ = g.Generate()
		}
	}
}
