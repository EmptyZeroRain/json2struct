package parser

import (
	"github.com/EmptyZeroRain/json2struct/option"
	"strings"
	"testing"
)

func TestDeepSampledArrayDoesNotRecurse(t *testing.T) {
	input := `{"items":[` + strings.Repeat(`[`, 200) + `1` + strings.Repeat(`]`, 200) + `]}`
	_, err := ParseWithOptions(strings.NewReader(input), Options{Limits: option.Limits{MaxDepth: 32, SampleArrayItems: 1}})
	if err == nil {
		t.Fatal("expected depth limit")
	}
}
