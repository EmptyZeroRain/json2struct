package schema

import "fmt"

// FieldType is the inferred JSON type.
type FieldType string

const (
	TypeUnknown FieldType = "unknown"
	TypeNull    FieldType = "null"
	TypeString  FieldType = "string"
	TypeInteger FieldType = "integer"
	TypeNumber  FieldType = "number"
	TypeBoolean FieldType = "boolean"
	TypeObject  FieldType = "object"
	TypeArray   FieldType = "array"
	TypeAny     FieldType = "any"
)

// Field describes one JSON value or object property.
type Field struct {
	Name     string            `json:"name"`
	Type     FieldType         `json:"type"`
	Required bool              `json:"required"`
	Nullable bool              `json:"nullable"`
	Array    bool              `json:"array"`
	Children map[string]*Field `json:"children,omitempty"`
	Element  *Field            `json:"element,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
}

func NewField(name string, typ FieldType) *Field {
	return &Field{Name: name, Type: typ, Required: true}
}

// Clone returns an independent copy of the field tree. It is safe for callers
// to modify the returned value without changing the schema held by a Generator.
func (f *Field) Clone() *Field { return f.clone() }

func (f *Field) clone() *Field {
	r, err := cloneChecked(f, make(map[*Field]bool))
	if err != nil {
		return nil
	}
	return r
}
func cloneChecked(f *Field, seen map[*Field]bool) (*Field, error) {
	if f == nil {
		return nil, nil
	}
	if seen[f] {
		return nil, fmt.Errorf("schema: cycle detected")
	}
	seen[f] = true
	r := *f
	if f.Children != nil {
		r.Children = make(map[string]*Field, len(f.Children))
		for k, v := range f.Children {
			c, e := cloneChecked(v, seen)
			if e != nil {
				return nil, e
			}
			r.Children[k] = c
		}
	}
	if f.Element != nil {
		c, e := cloneChecked(f.Element, seen)
		if e != nil {
			return nil, e
		}
		r.Element = c
	}
	if f.Tags != nil {
		r.Tags = make(map[string]string, len(f.Tags))
		for k, v := range f.Tags {
			r.Tags[k] = v
		}
	}
	delete(seen, f)
	return &r, nil
}
func cloneWithSeen(f *Field, seen map[*Field]bool) *Field {
	if f == nil {
		return nil
	}
	if seen[f] {
		return nil
	}
	seen[f] = true
	r := *f
	if f.Children != nil {
		r.Children = make(map[string]*Field, len(f.Children))
		for k, v := range f.Children {
			r.Children[k] = cloneWithSeen(v, seen)
		}
	}
	if f.Tags != nil {
		r.Tags = make(map[string]string, len(f.Tags))
		for k, v := range f.Tags {
			r.Tags[k] = v
		}
	}
	r.Element = cloneWithSeen(f.Element, seen)
	delete(seen, f)
	return &r
}
