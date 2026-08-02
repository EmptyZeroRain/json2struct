package parser

import (
	"github.com/EmptyZeroRain/json2struct/option"
	"strings"
	"testing"
)

func TestLimitsCountNestedFields(t *testing.T) {
	_, err := ParseWithOptions(strings.NewReader(`{"a":{"b":1}}`), Options{Limits: option.Limits{MaxFields: 1}})
	if err == nil {
		t.Fatal("expected field limit")
	}
}

func TestNDJSONLineByteLimit(t *testing.T) {
	_, err := ParseNDJSONWithOptions(strings.NewReader(`{"long":"12345"}`), Options{Limits: option.Limits{MaxLineBytes: 10}})
	if err == nil {
		t.Fatal("expected line limit")
	}
}
