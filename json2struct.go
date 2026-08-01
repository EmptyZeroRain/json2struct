// Package json2struct infers Go structs from JSON and NDJSON samples.
package json2struct

import (
	"bytes"
	"errors"
	"github.com/EmptyZeroRain/json2struct/generator"
	"github.com/EmptyZeroRain/json2struct/option"
	"github.com/EmptyZeroRain/json2struct/parser"
	"github.com/EmptyZeroRain/json2struct/schema"
	"io"
	"os"
	"sync"
)

type Options = option.Options
type Field = schema.Field
type Generator struct {
	mu            sync.RWMutex
	options       option.Options
	schemas       []*schema.Field
	merged        *schema.Field
	version       uint64
	mergedVersion uint64
}

func New(o Options) *Generator                 { o.Defaults(); return &Generator{options: o} }
func (g *Generator) AddJSON(data []byte) error { return g.AddReader(bytes.NewReader(data)) }
func (g *Generator) AddBatch(data [][]byte, workers int) error {
	s, err := InferBatch(data, BatchOptions{Workers: workers})
	if err != nil {
		return err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.schemas = append(g.schemas, s)
	if g.options.Merge {
		if g.merged == nil {
			g.merged = s.Clone()
		} else {
			schema.MergeInto(g.merged, s)
		}
	}
	g.version++
	return nil
}
func (g *Generator) AddReader(r io.Reader) error {
	s, err := (parser.JSONParser{}).Parse(r)
	if err != nil {
		return err
	}
	g.mu.Lock()
	g.schemas = append(g.schemas, s)
	if g.options.Merge {
		if g.merged == nil {
			g.merged = s.Clone()
		} else {
			schema.MergeInto(g.merged, s)
		}
	}
	g.version++
	g.mu.Unlock()
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
func (g *Generator) AddNDJSON(r io.Reader) error {
	s, err := parser.ParseNDJSON(r)
	if err != nil {
		return err
	}
	g.mu.Lock()
	g.schemas = append(g.schemas, s)
	if g.options.Merge {
		if g.merged == nil {
			g.merged = s.Clone()
		} else {
			schema.MergeInto(g.merged, s)
		}
	}
	g.version++
	g.mu.Unlock()
	return nil
}
func (g *Generator) Schema() *Field { return g.mergedSchema() }
func (g *Generator) Generate() ([]byte, error) {
	return generator.Generate(g.mergedSchema(), g.options)
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
