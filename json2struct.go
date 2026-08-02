// Package json2struct infers Go structs from JSON and NDJSON samples.
package json2struct

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/EmptyZeroRain/json2struct/generator"
	"github.com/EmptyZeroRain/json2struct/option"
	"github.com/EmptyZeroRain/json2struct/parser"
	"github.com/EmptyZeroRain/json2struct/schema"
)

type Options = option.Options
type Field = schema.Field
type Generator struct {
	mu             sync.RWMutex
	options        option.Options
	schemas        []*schema.Field
	merged         *schema.Field
	version        uint64
	mergedVersion  uint64
	totalBytes     int64
	samples        int
	nameCache      map[string]string
	generatorCache *generator.NameCache
	stats          Stats
}

func New(o Options) *Generator {
	o.Defaults()
	return &Generator{options: o, generatorCache: generator.NewNameCache()}
}
func (g *Generator) AddJSON(data []byte) error { return g.AddReader(bytes.NewReader(data)) }
func (g *Generator) Stats() Stats              { g.mu.RLock(); defer g.mu.RUnlock(); return g.stats }
func (g *Generator) AddBatch(data [][]byte, workers int) error {
	if g.options.Limits.MaxSamples > 0 {
		g.mu.RLock()
		current := g.samples
		g.mu.RUnlock()
		if current+len(data) > g.options.Limits.MaxSamples {
			return fmt.Errorf("%w: %d", parser.ErrMaxSamples, g.options.Limits.MaxSamples)
		}
	}
	if g.options.Limits.MaxTotalBytes > 0 {
		var total int64
		for _, d := range data {
			total += int64(len(d))
		}
		g.mu.RLock()
		current := g.totalBytes
		g.mu.RUnlock()
		if current+total > g.options.Limits.MaxTotalBytes {
			return fmt.Errorf("%w: %d", parser.ErrMaxTotalBytes, g.options.Limits.MaxTotalBytes)
		}
	}
	s, err := InferBatch(data, BatchOptions{Workers: workers})
	if err != nil {
		return err
	}
	return g.addSchema(s)
}

// AddBatchWithOptions is the configurable batch ingestion API.
func (g *Generator) AddBatchWithOptions(data [][]byte, opts BatchOptions) error {
	s, err := InferBatch(data, opts)
	if err != nil {
		return err
	}
	return g.addSchema(s)
}
func (g *Generator) AddReader(r io.Reader) error {
	s, ps, err := parser.ParseWithStats(r, parser.Options{Limits: g.options.Limits})
	if err != nil {
		return err
	}
	err = g.addSchema(s)
	if err == nil {
		g.mu.Lock()
		g.stats.Samples++
		g.stats.Nodes += ps.Nodes
		g.stats.Fields += ps.Fields
		g.stats.Bytes += ps.Bytes
		g.stats.Duration += ps.Duration
		g.mu.Unlock()
	}
	return err
}

func (g *Generator) addSchema(s *schema.Field) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.options.Limits.MaxSamples > 0 && g.samples+1 > g.options.Limits.MaxSamples {
		return fmt.Errorf("%w: %d", parser.ErrMaxSamples, g.options.Limits.MaxSamples)
	}
	if g.options.Limits.MaxTotalBytes > 0 && g.totalBytes >= g.options.Limits.MaxTotalBytes {
		return fmt.Errorf("%w: %d", parser.ErrMaxTotalBytes, g.options.Limits.MaxTotalBytes)
	}
	if g.options.Limits.MaxSchemaNodes > 0 && schema.CountNodes(s) > g.options.Limits.MaxSchemaNodes {
		return fmt.Errorf("%w: %d", parser.ErrMaxSchemaNodes, g.options.Limits.MaxSchemaNodes)
	}
	if g.options.Merge {
		if g.merged == nil {
			g.merged = s.Clone()
		} else {
			schema.MergeInto(g.merged, s)
		}
	} else {
		g.schemas = append(g.schemas, s)
	}
	g.version++
	g.samples++
	// Byte accounting for stream inputs is updated by AddReader; byte-slice and
	// batch APIs update it at their boundary before calling addSchema.
	return nil
}
func (g *Generator) AddFile(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return g.AddReader(f)
}
func (g *Generator) AddFileUnder(root, name string) error {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	pathAbs, err := filepath.Abs(filepath.Join(root, name))
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("path escapes root")
	}
	rootReal, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return err
	}
	pathReal, err := filepath.EvalSymlinks(pathAbs)
	if err != nil {
		return err
	}
	realRel, err := filepath.Rel(rootReal, pathReal)
	if err != nil || realRel == ".." || strings.HasPrefix(realRel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("path escapes root through symlink")
	}
	return g.AddFile(pathReal)
}
func (g *Generator) AddNDJSON(r io.Reader) error {
	s, err := parser.ParseNDJSONWithOptions(r, parser.Options{Limits: g.options.Limits})
	if err != nil {
		return err
	}
	return g.addSchema(s)
}
func (g *Generator) AddNDJSONParallel(r io.Reader, opts parser.Options, workers int) error {
	s, err := parser.ParseNDJSONParallel(r, opts, workers)
	if err != nil {
		return err
	}
	return g.addSchema(s)
}
func (g *Generator) AddNDJSONParallelContext(ctx context.Context, r io.Reader, opts parser.Options, workers int) error {
	s, err := parser.ParseNDJSONParallelContext(ctx, r, opts, workers)
	if err != nil {
		return err
	}
	return g.addSchema(s)
}
func (g *Generator) AddNDJSONParallelReadCloser(ctx context.Context, r io.ReadCloser, opts parser.Options, workers int) error {
	s, err := parser.ParseNDJSONParallelReadCloser(ctx, r, opts, workers)
	if err != nil {
		return err
	}
	return g.addSchema(s)
}
func (g *Generator) Schema() *Field { return g.mergedSchema() }
func (g *Generator) Generate() ([]byte, error) {
	return generator.GenerateWithCache(g.mergedSchema(), g.options, g.generatorCache)
}
func (g *Generator) GenerateFromSchema(s *Field) ([]byte, error) {
	return generator.Generate(s, g.options)
}
func (g *Generator) GenerateAll() ([][]byte, error) {
	g.mu.RLock()
	schemas := append([]*schema.Field(nil), g.schemas...)
	opts := g.options
	g.mu.RUnlock()
	if len(schemas) == 0 {
		return nil, ErrNoSchema
	}
	result := make([][]byte, 0, len(schemas))
	for _, s := range schemas {
		code, err := generator.Generate(s, opts)
		if err != nil {
			return nil, err
		}
		result = append(result, code)
	}
	return result, nil
}

func (g *Generator) mergedSchema() *schema.Field {
	g.mu.RLock()
	if g.options.Merge && g.merged != nil {
		r := g.merged.Clone()
		g.mu.RUnlock()
		return r
	}
	if g.merged != nil && g.mergedVersion == g.version {
		r := g.merged.Clone()
		g.mu.RUnlock()
		return r
	}
	schemas := append([]*schema.Field(nil), g.schemas...)
	merge := g.options.Merge
	version := g.version
	g.mu.RUnlock()
	var r *schema.Field
	for _, s := range schemas {
		if r == nil {
			r = s
		} else if merge {
			r = schema.Merge(r, s)
		}
	}
	g.mu.Lock()
	if g.merged == nil && g.version == version && r != nil {
		g.merged = r.Clone()
		g.mergedVersion = version
	}
	g.mu.Unlock()
	return r
}

var ErrNoSchema = errors.New("json2struct: no schema; add JSON before generating")
