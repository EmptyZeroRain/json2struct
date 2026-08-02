package json2struct_test

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	json2struct "github.com/EmptyZeroRain/json2struct"
	"github.com/EmptyZeroRain/json2struct/inference"
	"github.com/EmptyZeroRain/json2struct/schema"
)

func TestGenerateMergedStruct(t *testing.T) {
	g := json2struct.New(json2struct.Options{Name: "User", Merge: true, Omitempty: true})
	if err := g.AddJSON([]byte(`{"id":1,"user_name":"a","profile":{"ip_address":"127.0.0.1"}}`)); err != nil {
		t.Fatal(err)
	}
	if err := g.AddJSON([]byte(`{"id":2,"email":"a@b.test","profile":{"ip_address":"127.0.0.2"}}`)); err != nil {
		t.Fatal(err)
	}
	code, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}
	out := string(code)
	for _, want := range []string{"type User struct", "ID      string `json:\"id\"`", "Email   string `json:\"email,omitempty\"`", "IPAddress string"} {
		if !strings.Contains(out, want) {
			t.Errorf("generated code missing %q:\n%s", want, out)
		}
	}
}

func TestNDJSON(t *testing.T) {
	g := json2struct.New(json2struct.Options{Name: "Event", Merge: true})
	if err := g.AddNDJSON(strings.NewReader("{\"a\":1}\n{\"b\":true}\n")); err != nil {
		t.Fatal(err)
	}
	if g.Schema().Children["a"].Required {
		t.Error("a should be optional after merge")
	}
}

func TestGenerateFromSchema(t *testing.T) {
	g := json2struct.New(json2struct.Options{Name: "Thing"})
	if _, err := g.GenerateFromSchema(nil); err == nil {
		t.Error("nil schema should fail")
	}
}

func TestInvalidAndEmptyInputs(t *testing.T) {
	g := json2struct.New(json2struct.Options{Name: "Thing"})
	if err := g.AddJSON([]byte(`{"a":1}{"b":2}`)); err == nil {
		t.Error("expected trailing JSON error")
	}
	if _, err := g.Generate(); err == nil {
		t.Error("expected no schema error")
	}
}

func TestUnusualJSONKeysProduceValidIdentifiers(t *testing.T) {
	g := json2struct.New(json2struct.Options{Name: "Weird"})
	if err := g.AddJSON([]byte(`{"123 name":1,"a.b":2,"type":3,"a-b":4,"a_b":5}`)); err != nil {
		t.Fatal(err)
	}
	code, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}
	out := string(code)
	for _, want := range []string{"Field123Name", "AB", "Type         string", "AB2", "AB3", "json:\"123 name\"", "json:\"a.b\""} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in generated code:\n%s", want, out)
		}
	}
}

func TestTimeAndUnknownValuesDefaultToString(t *testing.T) {
	g := json2struct.New(json2struct.Options{Name: "Record"})
	if err := g.AddJSON([]byte(`{"created_at":"2025-01-01T00:00:00Z"}`)); err != nil {
		t.Fatal(err)
	}
	code, err := g.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(code), "CreatedAt string") {
		t.Fatalf("time value should remain string:\n%s", code)
	}
	if got := inference.Infer(struct{ Value int }{Value: 1}).Type; got != schema.TypeString {
		t.Fatalf("unknown value type = %q, want string", got)
	}
	if got := inference.Infer(json.Number("123")).Type; got != schema.TypeString {
		t.Fatalf("number type = %q, want string", got)
	}
}

func TestConcurrentUse(t *testing.T) {
	g := json2struct.New(json2struct.Options{Name: "Concurrent", Merge: true})
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := g.AddJSON([]byte(`{"value":"x"}`)); err != nil {
				t.Error(err)
			}
			if _, err := g.Generate(); err != nil {
				t.Error(err)
			}
			_ = g.Schema()
		}(i)
	}
	wg.Wait()
}

func TestInferBatch(t *testing.T) {
	s, err := json2struct.InferBatch([][]byte{[]byte(`{"id":1}`), []byte(`{"name":"x"}`)}, json2struct.BatchOptions{Workers: 2})
	if err != nil {
		t.Fatal(err)
	}
	if s.Children["id"].Required {
		t.Error("id should be optional")
	}
}

func TestAddBatchWithOptions(t *testing.T) {
	g := json2struct.New(json2struct.Options{Name: "Batch", Merge: true})
	if err := g.AddBatchWithOptions([][]byte{[]byte(`{"a":1}`), []byte(`{"b":2}`)}, json2struct.BatchOptions{Workers: 2}); err != nil {
		t.Fatal(err)
	}
	if g.Schema().Children["a"].Required {
		t.Error("a should be optional")
	}
}

func TestSchemaCycleIsRejected(t *testing.T) {
	f := &json2struct.Field{Name: "root", Type: schema.TypeObject, Children: map[string]*json2struct.Field{}}
	f.Children["self"] = f
	g := json2struct.New(json2struct.Options{Name: "Cycle"})
	if _, err := g.GenerateFromSchema(f); err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestAddFileUnderRejectsTraversal(t *testing.T) {
	g := json2struct.New(json2struct.Options{Name: "Safe"})
	if err := g.AddFileUnder("/tmp", "../etc/passwd"); err == nil {
		t.Fatal("expected path traversal rejection")
	}
}
