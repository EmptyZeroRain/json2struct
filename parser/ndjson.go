package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"runtime"
	"sync"

	"github.com/EmptyZeroRain/json2struct/schema"
)

// ParseNDJSONParallel parses bounded batches concurrently and merges results
// in input order. It uses a bounded queue, so a large stream is not retained
// in memory by the parser.
func ParseNDJSONParallel(r io.Reader, opts Options, workers int) (*schema.Field, error) {
	if r == nil {
		return nil, fmt.Errorf("reader is nil")
	}
	if workers == 0 {
		workers = runtime.GOMAXPROCS(0)
	}
	if workers < 1 {
		return nil, fmt.Errorf("parser: workers must be positive")
	}
	max := opts.Limits.MaxLineBytes
	if max <= 0 {
		max = 16 * 1024 * 1024
	}
	type job struct {
		line int
		data []byte
	}
	type result struct {
		line  int
		field *schema.Field
		err   error
	}
	jobs := make(chan job, workers*2)
	results := make(chan result, workers*2)
	done := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				line := bytes.TrimSpace(j.data)
				if len(line) == 0 {
					continue
				}
				f, err := ParseWithOptions(bytes.NewReader(line), opts)
				select {
				case results <- result{j.line, f, err}:
				case <-done:
					return
				}
			}
		}()
	}
	go func() { wg.Wait(); close(results) }()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), max+1)
	line := 0
	for scanner.Scan() {
		line++
		raw := append([]byte(nil), scanner.Bytes()...)
		if len(raw) > max {
			close(done)
			close(jobs)
			wg.Wait()
			return nil, fmt.Errorf("ndjson line %d: %w", line, ErrMaxBytes)
		}
		jobs <- job{line, raw}
	}
	close(jobs)
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	var merged *schema.Field
	for item := range results {
		if item.err != nil {
			return nil, fmt.Errorf("ndjson line %d: %w", item.line, item.err)
		}
		if item.field != nil {
			if merged == nil {
				merged = item.field
			} else {
				schema.MergeInto(merged, item.field)
			}
		}
	}
	if merged == nil {
		return nil, fmt.Errorf("no JSON values")
	}
	return merged, nil
}
