package parser

import (
	"context"
	"errors"
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

func TestNDJSONParallelPreservesSequence(t *testing.T) {
	got, err := ParseNDJSONParallel(strings.NewReader("{\"a\":1}\n{\"b\":2}\n"), Options{}, 2)
	if err != nil || got.Children["a"] == nil || got.Children["b"] == nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestParseWithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ParseWithContext(ctx, strings.NewReader(`{"a":1}`), Options{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v, want cancellation", err)
	}
}

func TestLimitStringAndNumber(t *testing.T) {
	_, err := ParseWithOptions(strings.NewReader(`{"value":"12345"}`), Options{Limits: option.Limits{MaxStringBytes: 3}})
	if !errors.Is(err, ErrMaxStringBytes) {
		t.Fatalf("error=%v", err)
	}
	_, err = ParseWithOptions(strings.NewReader(`{"value":12345}`), Options{Limits: option.Limits{MaxNumberBytes: 3}})
	if !errors.Is(err, ErrMaxNumberBytes) {
		t.Fatalf("error=%v", err)
	}
}

func TestMaxNodesAndDeepInput(t *testing.T) {
	_, err := ParseWithOptions(strings.NewReader(`{"a":{"b":{"c":1}}}`), Options{Limits: option.Limits{MaxNodes: 2}})
	if !errors.Is(err, ErrMaxNodes) {
		t.Fatalf("error=%v, want ErrMaxNodes", err)
	}
}
