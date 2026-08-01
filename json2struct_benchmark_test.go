package json2struct_test

import (
	json2struct "github.com/EmptyZeroRain/json2struct"
	"testing"
)

func BenchmarkAddAndGenerate(b *testing.B) {
	data := []byte(`{"id":123,"name":"sample","active":true,"profile":{"ip":"127.0.0.1"},"items":[{"code":1}]}`)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		g := json2struct.New(json2struct.Options{Name: "Sample", Merge: true})
		if err := g.AddJSON(data); err != nil {
			b.Fatal(err)
		}
		if _, err := g.Generate(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMergeTenThousand(b *testing.B) {
	data := []byte(`{"id":123,"name":"sample","active":true,"profile":{"ip":"127.0.0.1"}}`)
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		g := json2struct.New(json2struct.Options{Name: "Sample", Merge: true})
		for i := 0; i < 10000; i++ {
			if err := g.AddJSON(data); err != nil {
				b.Fatal(err)
			}
		}
	}
}
