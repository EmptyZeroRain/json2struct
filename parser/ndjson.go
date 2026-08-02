package parser

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"runtime"
	"sync"

	"github.com/EmptyZeroRain/json2struct/schema"
)

// ParseNDJSONParallel parses lines concurrently and merges them by sequence.
// The input queue and result queue are bounded; the whole stream is never
// retained in memory.
func ParseNDJSONParallel(r io.Reader, opts Options, workers int) (*schema.Field, error) {
	return ParseNDJSONParallelContext(context.Background(), r, opts, workers)
}

// ParseNDJSONParallelContext supports cancellation while reading and parsing.
func ParseNDJSONParallelContext(ctx context.Context, r io.Reader, opts Options, workers int) (*schema.Field, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
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
		seq  int
		data []byte
	}
	type result struct {
		seq   int
		field *schema.Field
		err   error
	}
	jobs := make(chan job, workers*2)
	results := make(chan result, workers*2)
	done := make(chan struct{})
	var stopOnce sync.Once
	stop := func() { stopOnce.Do(func() { close(done) }) }
	var workersWG sync.WaitGroup
	for i := 0; i < workers; i++ {
		workersWG.Add(1)
		go func() {
			defer workersWG.Done()
			for j := range jobs {
				line := bytes.TrimSpace(j.data)
				if len(line) == 0 {
					select {
					case results <- result{seq: j.seq}:
					case <-done:
						return
					}
					continue
				}
				f, err := ParseWithContext(ctx, bytes.NewReader(line), opts)
				select {
				case results <- result{j.seq, f, err}:
				case <-done:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go func() { workersWG.Wait(); close(results) }()
	producerErr := make(chan error, 1)
	go func() {
		defer close(jobs)
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 64*1024), max+1)
		seq := 0
		var totalBytes int64
		for scanner.Scan() {
			if err := ctx.Err(); err != nil {
				producerErr <- err
				stop()
				return
			}
			raw := append([]byte(nil), scanner.Bytes()...)
			seq++
			if opts.Limits.MaxSamples > 0 && seq > opts.Limits.MaxSamples {
				producerErr <- fmt.Errorf("%w: %d", ErrMaxSamples, opts.Limits.MaxSamples)
				stop()
				return
			}
			totalBytes += int64(len(raw))
			if opts.Limits.MaxTotalBytes > 0 && totalBytes > opts.Limits.MaxTotalBytes {
				producerErr <- fmt.Errorf("%w: %d", ErrMaxTotalBytes, opts.Limits.MaxTotalBytes)
				stop()
				return
			}
			if len(raw) > max {
				producerErr <- fmt.Errorf("ndjson line %d: %w", seq, ErrMaxBytes)
				stop()
				return
			}
			select {
			case jobs <- job{seq - 1, raw}:
			case <-done:
				return
			case <-ctx.Done():
				producerErr <- ctx.Err()
				stop()
				return
			}
		}
		producerErr <- scanner.Err()
	}()

	pending := make(map[int]result, workers*2)
	next := 0
	var merged *schema.Field
	var firstErr error
	for item := range results {
		if firstErr != nil {
			continue
		}
		pending[item.seq] = item
		for {
			ready, ok := pending[next]
			if !ok {
				break
			}
			delete(pending, next)
			next++
			if ready.err != nil {
				firstErr = fmt.Errorf("ndjson line %d: %w", ready.seq+1, ready.err)
				stop()
				break
			}
			if ready.field != nil {
				if merged == nil {
					merged = ready.field
				} else {
					schema.MergeInto(merged, ready.field)
				}
			}
		}
	}
	if err := <-producerErr; err != nil && firstErr == nil {
		firstErr = err
	}
	if firstErr == nil && len(pending) != 0 {
		firstErr = fmt.Errorf("parser: incomplete NDJSON results")
	}
	if firstErr == nil {
		firstErr = ctx.Err()
	}
	if firstErr != nil {
		return nil, firstErr
	}
	if merged == nil {
		return nil, fmt.Errorf("no JSON values")
	}
	return merged, nil
}
