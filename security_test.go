package json2struct_test

import (
	json2struct "github.com/EmptyZeroRain/json2struct"
	"github.com/EmptyZeroRain/json2struct/option"
	"github.com/EmptyZeroRain/json2struct/parser"
	"strings"
	"testing"
)

func TestParserLimits(t *testing.T) {
	_, err := parser.ParseWithOptions(strings.NewReader(`{"a":{"b":1}}`), parser.Options{Limits: option.Limits{MaxDepth: 1}})
	if err == nil {
		t.Fatal("expected depth limit error")
	}
}

func TestGeneratedTagsEscapeBackticks(t *testing.T) {
	g := json2struct.New(json2struct.Options{Name: "Safe"})
	if err := g.AddJSON([]byte("{\"bad`key\":1}")); err != nil {
		t.Fatal(err)
	}
	code, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(code), "json:\\\"bad`key\\\"") {
		t.Fatalf("unsafe tag output: %s", code)
	}
}
