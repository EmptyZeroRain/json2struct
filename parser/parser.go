package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/EmptyZeroRain/json2struct/inference"
	"github.com/EmptyZeroRain/json2struct/option"
	"github.com/EmptyZeroRain/json2struct/schema"
)

type Parser interface {
	Parse(io.Reader) (*schema.Field, error)
}

type JSONParser struct{}

func (JSONParser) Parse(r io.Reader) (*schema.Field, error) {
	if r == nil {
		return nil, fmt.Errorf("reader is nil")
	}
	dec := json.NewDecoder(r)
	dec.UseNumber()
	var value interface{}
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	var extra interface{}
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("multiple JSON values; use ParseNDJSON for line-delimited input")
		}
		return nil, fmt.Errorf("invalid trailing JSON: %w", err)
	}
	return inference.Infer(value), nil
}

// ParseWithOptions parses JSON while enforcing resource limits.
func ParseWithOptions(r io.Reader, opts Options) (*schema.Field, error) {
	if r == nil {
		return nil, fmt.Errorf("reader is nil")
	}
	var data []byte
	var err error
	if opts.Limits.MaxBytes > 0 {
		data, err = io.ReadAll(io.LimitReader(r, opts.Limits.MaxBytes+1))
	} else {
		data, err = io.ReadAll(r)
	}
	if err != nil {
		return nil, err
	}
	if opts.Limits.MaxBytes > 0 && int64(len(data)) > opts.Limits.MaxBytes {
		return nil, fmt.Errorf("parser: maximum input size %d exceeded", opts.Limits.MaxBytes)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var value interface{}
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	var extra interface{}
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("multiple JSON values")
		}
		return nil, fmt.Errorf("invalid trailing JSON: %w", err)
	}
	count := 0
	if err := validateValue(value, opts.Limits, 0, "$", &count); err != nil {
		return nil, err
	}
	return inference.Infer(value), nil
}

// ParseBytes is an allocation-conscious convenience API for a single sample.
func ParseBytes(data []byte) (*schema.Field, error) {
	return (JSONParser{}).Parse(bytes.NewReader(data))
}

func validateValue(v interface{}, l option.Limits, depth int, path string, count *int) error {
	if l.MaxDepth > 0 && depth > l.MaxDepth {
		return fmt.Errorf("parser: maximum depth %d exceeded at %s", l.MaxDepth, path)
	}
	switch x := v.(type) {
	case map[string]interface{}:
		for k, child := range x {
			*count = *count + 1
			if l.MaxFields > 0 && *count > l.MaxFields {
				return fmt.Errorf("parser: maximum fields %d exceeded at %s", l.MaxFields, path)
			}
			if err := validateValue(child, l, depth+1, path+"."+k, count); err != nil {
				return err
			}
		}
	case []interface{}:
		if l.MaxArrayItems > 0 && len(x) > l.MaxArrayItems {
			return fmt.Errorf("parser: maximum array items %d exceeded at %s", l.MaxArrayItems, path)
		}
		for i, child := range x {
			if err := validateValue(child, l, depth+1, fmt.Sprintf("%s[%d]", path, i), count); err != nil {
				return err
			}
		}
	}
	return nil
}

func ValidateLimits(f *schema.Field, l option.Limits) error {
	fields := 0
	var walk func(*schema.Field, int) error
	walk = func(n *schema.Field, depth int) error {
		if n == nil {
			return nil
		}
		if l.MaxDepth > 0 && depth > l.MaxDepth {
			return fmt.Errorf("parser: maximum depth %d exceeded", l.MaxDepth)
		}
		fields += len(n.Children)
		if l.MaxFields > 0 && fields > l.MaxFields {
			return fmt.Errorf("parser: maximum fields %d exceeded", l.MaxFields)
		}
		if l.MaxArrayItems > 0 && n.Array && n.Element != nil { /* element schema is already sampled */
		}
		for _, c := range n.Children {
			if err := walk(c, depth+1); err != nil {
				return err
			}
		}
		return walk(n.Element, depth+1)
	}
	return walk(f, 0)
}

// ParseNDJSON reads one JSON value per non-empty line and merges observations.
func ParseNDJSON(r io.Reader) (*schema.Field, error) {
	return ParseNDJSONWithOptions(r, Options{})
}

func ParseNDJSONWithOptions(r io.Reader, opts Options) (*schema.Field, error) {
	if r == nil {
		return nil, fmt.Errorf("reader is nil")
	}
	s := bufio.NewScanner(r)
	max := opts.Limits.MaxLineBytes
	if max <= 0 {
		max = 16 * 1024 * 1024
	}
	// Scanner's maximum token size includes a small implementation margin;
	// enforce the public limit explicitly after scanning as well.
	s.Buffer(make([]byte, 64*1024), max+1)
	var result *schema.Field
	line := 0
	for s.Scan() {
		line++
		lineBytes := bytes.TrimSpace(s.Bytes())
		if opts.Limits.MaxLineBytes > 0 && len(s.Bytes()) > opts.Limits.MaxLineBytes {
			return nil, fmt.Errorf("ndjson line %d: maximum line size %d exceeded", line, opts.Limits.MaxLineBytes)
		}
		if len(lineBytes) == 0 {
			continue
		}
		if opts.Limits.MaxBytes > 0 && int64(len(lineBytes)) > opts.Limits.MaxBytes {
			return nil, fmt.Errorf("ndjson line %d: maximum input size %d exceeded", line, opts.Limits.MaxBytes)
		}
		lineOpts := opts
		lineOpts.Limits.MaxBytes = 0
		f, err := ParseWithOptions(bytes.NewReader(lineBytes), lineOpts)
		if err != nil {
			return nil, fmt.Errorf("ndjson line %d: %w", line, err)
		}
		if result == nil {
			result = f
		} else {
			schema.MergeInto(result, f)
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("no JSON values")
	}
	return result, nil
}
