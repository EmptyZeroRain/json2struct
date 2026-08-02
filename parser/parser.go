package parser

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/EmptyZeroRain/json2struct/option"
	"github.com/EmptyZeroRain/json2struct/schema"
)

type Parser interface {
	Parse(io.Reader) (*schema.Field, error)
}

type JSONParser struct{}

func (JSONParser) Parse(r io.Reader) (*schema.Field, error) {
	return ParseWithOptions(r, Options{})
}

// ParseWithOptions parses JSON while enforcing resource limits.
func ParseWithOptions(r io.Reader, opts Options) (*schema.Field, error) {
	if r == nil {
		return nil, fmt.Errorf("reader is nil")
	}
	counted := &countingReader{reader: r}
	var source io.Reader = counted
	if opts.Limits.MaxBytes > 0 {
		source = io.LimitReader(counted, opts.Limits.MaxBytes+1)
	}
	dec := json.NewDecoder(source)
	dec.UseNumber()
	state := &parseState{}
	f, err := inferDecoder(dec, "", opts.Limits, 0, state)
	if err != nil {
		return nil, err
	}
	if _, err = dec.Token(); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("multiple JSON values")
		}
		return nil, fmt.Errorf("invalid trailing JSON: %w", err)
	}
	if opts.Limits.MaxBytes > 0 && counted.n > opts.Limits.MaxBytes {
		return nil, fmt.Errorf("%w: %d", ErrMaxBytes, opts.Limits.MaxBytes)
	}
	return f, nil
}

// ParseWithContext is ParseWithOptions with cancellation support.
func ParseWithContext(ctx context.Context, r io.Reader, opts Options) (*schema.Field, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return ParseWithOptions(&contextReader{ctx: ctx, r: r}, opts)
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(p)
}

type parseState struct{ nodes, fields int }

func inferDecoder(dec *json.Decoder, name string, l option.Limits, depth int, state *parseState) (*schema.Field, error) {
	state.nodes++
	if l.MaxNodes > 0 && state.nodes > l.MaxNodes {
		return nil, fmt.Errorf("%w: %d", ErrMaxNodes, l.MaxNodes)
	}
	if l.MaxDepth > 0 && depth > l.MaxDepth {
		return nil, fmt.Errorf("%w: %d", ErrMaxDepth, l.MaxDepth)
	}
	t, err := dec.Token()
	if err != nil {
		return nil, err
	}
	f := schema.NewField(name, schema.TypeUnknown)
	switch v := t.(type) {
	case nil:
		f.Type, f.Nullable = schema.TypeNull, true
	case string:
		if l.MaxStringBytes > 0 && len(v) > l.MaxStringBytes {
			return nil, fmt.Errorf("%w: %d", ErrMaxStringBytes, l.MaxStringBytes)
		}
		f.Type = schema.TypeString
	case bool:
		f.Type = schema.TypeBoolean
	case json.Number:
		if l.MaxNumberBytes > 0 && len(v) > l.MaxNumberBytes {
			return nil, fmt.Errorf("%w: %d", ErrMaxNumberBytes, l.MaxNumberBytes)
		}
		f.Type = schema.TypeString
	case json.Delim:
		switch v {
		case '{':
			f.Type = schema.TypeObject
			f.Children = map[string]*schema.Field{}
			for dec.More() {
				keyToken, e := dec.Token()
				key, ok := keyToken.(string)
				if e != nil {
					return nil, e
				}
				if !ok {
					return nil, fmt.Errorf("invalid object key")
				}
				state.fields++
				if l.MaxFields > 0 && state.fields > l.MaxFields {
					return nil, fmt.Errorf("%w: %d", ErrMaxFields, l.MaxFields)
				}
				child, e := inferDecoder(dec, key, l, depth+1, state)
				if e != nil {
					return nil, e
				}
				f.Children[key] = child
			}
			if _, e := dec.Token(); e != nil {
				return nil, e
			}
		case '[':
			f.Type, f.Array = schema.TypeArray, true
			var element *schema.Field
			n := 0
			for dec.More() {
				n++
				if l.MaxArrayItems > 0 && n > l.MaxArrayItems {
					return nil, fmt.Errorf("%w: %d", ErrMaxArrayItems, l.MaxArrayItems)
				}
				if l.SampleArrayItems > 0 && n > l.SampleArrayItems {
					if _, e := skipValue(dec); e != nil {
						return nil, e
					}
					continue
				}
				child, e := inferDecoder(dec, "", l, depth+1, state)
				if e != nil {
					return nil, e
				}
				if element == nil {
					element = child
				} else {
					schema.MergeInto(element, child)
				}
			}
			if _, e := dec.Token(); e != nil {
				return nil, e
			}
			f.Element = element
		default:
			return nil, fmt.Errorf("invalid JSON delimiter %q", v)
		}
	default:
		return nil, fmt.Errorf("unsupported JSON token")
	}
	return f, nil
}

func skipValue(dec *json.Decoder) (int, error) {
	t, e := dec.Token()
	if e != nil {
		return 0, e
	}
	if d, ok := t.(json.Delim); ok && (d == '{' || d == '[') {
		close := json.Delim('}')
		if d == '[' {
			close = ']'
		}
		for dec.More() {
			if _, e = skipValue(dec); e != nil {
				return 0, e
			}
		}
		_, e = dec.Token()
		if e != nil {
			return 0, e
		}
		_ = close
	}
	return 0, nil
}

type countingReader struct {
	reader io.Reader
	n      int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.n += int64(n)
	return n, err
}

// ParseBytes is an allocation-conscious convenience API for a single sample.
func ParseBytes(data []byte) (*schema.Field, error) {
	return ParseWithOptions(bytes.NewReader(data), Options{})
}

// validateValue is kept for compatibility with internal tests and callers of
// the old helper; production parsing uses inferDecoder directly.
func validateValue(root interface{}, l option.Limits) error {
	type item struct {
		value interface{}
		depth int
	}
	stack := []item{{root, 0}}
	fields, nodes := 0, 0
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		nodes++
		if l.MaxNodes > 0 && nodes > l.MaxNodes {
			return fmt.Errorf("%w: %d", ErrMaxNodes, l.MaxNodes)
		}
		if l.MaxDepth > 0 && current.depth > l.MaxDepth {
			return fmt.Errorf("%w: %d", ErrMaxDepth, l.MaxDepth)
		}
		switch x := current.value.(type) {
		case string:
			if l.MaxStringBytes > 0 && len(x) > l.MaxStringBytes {
				return fmt.Errorf("%w: %d", ErrMaxStringBytes, l.MaxStringBytes)
			}
		case json.Number:
			if l.MaxNumberBytes > 0 && len(x) > l.MaxNumberBytes {
				return fmt.Errorf("%w: %d", ErrMaxNumberBytes, l.MaxNumberBytes)
			}
		case map[string]interface{}:
			fields += len(x)
			if l.MaxFields > 0 && fields > l.MaxFields {
				return fmt.Errorf("%w: %d", ErrMaxFields, l.MaxFields)
			}
			for _, child := range x {
				stack = append(stack, item{child, current.depth + 1})
			}
		case []interface{}:
			if l.MaxArrayItems > 0 && len(x) > l.MaxArrayItems {
				return fmt.Errorf("%w: %d", ErrMaxArrayItems, l.MaxArrayItems)
			}
			for _, child := range x {
				stack = append(stack, item{child, current.depth + 1})
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
	var totalBytes int64
	for s.Scan() {
		line++
		totalBytes += int64(len(s.Bytes()))
		if opts.Limits.MaxTotalBytes > 0 && totalBytes > opts.Limits.MaxTotalBytes {
			return nil, fmt.Errorf("%w: %d", ErrMaxTotalBytes, opts.Limits.MaxTotalBytes)
		}
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
		if opts.Limits.MaxSchemaNodes > 0 && schema.CountNodes(result) > opts.Limits.MaxSchemaNodes {
			return nil, fmt.Errorf("%w: %d", ErrMaxSchemaNodes, opts.Limits.MaxSchemaNodes)
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
