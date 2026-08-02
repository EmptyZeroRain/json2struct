package json2struct_test

import (
	json2struct "github.com/EmptyZeroRain/json2struct"
	"testing"
)

func TestStatsForJSON(t *testing.T) {
	g := json2struct.New(json2struct.Options{Name: "S"})
	if err := g.AddJSON([]byte(`{"a":1}`)); err != nil {
		t.Fatal(err)
	}
	s := g.Stats()
	if s.Samples != 1 || s.Bytes == 0 || s.Nodes == 0 {
		t.Fatalf("stats=%+v", s)
	}
}
