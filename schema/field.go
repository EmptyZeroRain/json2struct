package schema

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
	if f == nil {
		return nil
	}
	r := *f
	if f.Children != nil {
		r.Children = make(map[string]*Field, len(f.Children))
		for k, v := range f.Children {
			r.Children[k] = v.clone()
		}
	}
	if f.Tags != nil {
		r.Tags = make(map[string]string, len(f.Tags))
		for k, v := range f.Tags {
			r.Tags[k] = v
		}
	}
	r.Element = f.Element.clone()
	return &r
}
