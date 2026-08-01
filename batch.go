package json2struct

import (
	"errors"
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
	var merged *schema.Field
	for i, f := range results {
		if errs[i] != nil {
			return nil, errs[i]
		}
		if merged == nil {
			merged = f
		} else {
			schema.MergeInto(merged, f)
		}
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
