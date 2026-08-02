package json2struct

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync"

	"github.com/EmptyZeroRain/json2struct/parser"
	"github.com/EmptyZeroRain/json2struct/schema"
)

var ErrInvalidWorkers = errors.New("json2struct: workers must be positive")

// BatchOptions controls parallel JSON inference.
type BatchOptions struct {
	Workers int
	Parser  parser.Options
}

// InferBatch parses samples concurrently and merges local schemas in a
// deterministic order. Input bytes are not modified.
func InferBatch(data [][]byte, opts BatchOptions) (*Field, error) {
	if len(data) == 0 {
		return nil, ErrNoSchema
	}
	if opts.Parser.Limits.MaxSamples > 0 && len(data) > opts.Parser.Limits.MaxSamples {
		return nil, fmt.Errorf("%w: %d", parser.ErrMaxSamples, opts.Parser.Limits.MaxSamples)
	}
	if opts.Parser.Limits.MaxTotalBytes > 0 {
		var total int64
		for _, sample := range data {
			total += int64(len(sample))
			if total > opts.Parser.Limits.MaxTotalBytes {
				return nil, fmt.Errorf("%w: %d", parser.ErrMaxTotalBytes, opts.Parser.Limits.MaxTotalBytes)
			}
		}
	}
	if opts.Workers == 0 {
		opts.Workers = runtime.GOMAXPROCS(0)
	}
	if opts.Workers < 1 {
		return nil, ErrInvalidWorkers
	}
	if opts.Workers > len(data) {
		opts.Workers = len(data)
	}
	results := make([]*schema.Field, len(data))
	errs := make([]error, len(data))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < opts.Workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i], errs[i] = parser.ParseWithOptions(bytesReader(data[i]), opts.Parser)
			}
		}()
	}
	for i := range data {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	for i := range results {
		if errs[i] != nil {
			return nil, errs[i]
		}
	}
	merged := schema.MergeAll(results)
	if opts.Parser.Limits.MaxSchemaNodes > 0 && schema.CountNodes(merged) > opts.Parser.Limits.MaxSchemaNodes {
		return nil, fmt.Errorf("%w: %d", parser.ErrMaxSchemaNodes, opts.Parser.Limits.MaxSchemaNodes)
	}
	return merged, nil
}

func bytesReader(b []byte) *byteReader { return &byteReader{b: b} }

type byteReader struct {
	b    []byte
	done bool
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.done = true
	return n, nil
}
