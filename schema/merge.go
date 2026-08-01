package schema

// Merge combines two observations. Missing properties become optional.
func Merge(a, b *Field) *Field {
	if a == nil {
		return b.clone()
	}
	if b == nil {
		return a.clone()
	}
	r := a.clone()
	r.Required = a.Required && b.Required
	r.Nullable = a.Nullable || b.Nullable || a.Type == TypeNull || b.Type == TypeNull
	r.Type = mergeType(a.Type, b.Type)
	if r.Type == TypeObject || a.Type == TypeObject || b.Type == TypeObject {
		r.Type = TypeObject
		if r.Children == nil {
			r.Children = map[string]*Field{}
		}
		for name, child := range r.Children {
			if _, exists := b.Children[name]; !exists {
				child.Required = false
			}
		}
		for name, child := range b.Children {
			if old, ok := r.Children[name]; ok {
				r.Children[name] = Merge(old, child)
			} else {
				c := child.clone()
				c.Required = false
				r.Children[name] = c
			}
		}
	}
	if a.Array || b.Array || a.Type == TypeArray || b.Type == TypeArray {
		r.Array, r.Type = true, TypeArray
		if a.Element != nil && b.Element != nil {
			r.Element = Merge(a.Element, b.Element)
		} else if r.Element == nil {
			if a.Element != nil {
				r.Element = a.Element.clone()
			}
			if b.Element != nil {
				r.Element = b.Element.clone()
			}
		}
	}
	return r
}

// MergeInto merges b into a in place. The caller must exclusively own a.
// It avoids cloning the accumulated tree and is intended for high-volume
// ingestion such as large NDJSON streams.
func MergeInto(a, b *Field) *Field {
	if a == nil {
		return b.Clone()
	}
	if b == nil {
		return a
	}
	a.Required = a.Required && b.Required
	a.Nullable = a.Nullable || b.Nullable || a.Type == TypeNull || b.Type == TypeNull
	a.Type = mergeType(a.Type, b.Type)
	if a.Type == TypeObject || b.Type == TypeObject {
		a.Type = TypeObject
		if a.Children == nil {
			a.Children = make(map[string]*Field)
		}
		for name, child := range a.Children {
			if _, exists := b.Children[name]; !exists {
				child.Required = false
			}
		}
		for name, child := range b.Children {
			if old, ok := a.Children[name]; ok {
				MergeInto(old, child)
			} else {
				c := child.Clone()
				c.Required = false
				a.Children[name] = c
			}
		}
	}
	if a.Array || b.Array || a.Type == TypeArray || b.Type == TypeArray {
		a.Array, a.Type = true, TypeArray
		if a.Element == nil && b.Element != nil {
			a.Element = b.Element.Clone()
		} else if a.Element != nil && b.Element != nil {
			MergeInto(a.Element, b.Element)
		}
	}
	return a
}

func mergeType(a, b FieldType) FieldType {
	if a == b {
		return a
	}
	if a == TypeNull {
		return b
	}
	if b == TypeNull {
		return a
	}
	if (a == TypeInteger && b == TypeNumber) || (a == TypeNumber && b == TypeInteger) {
		return TypeNumber
	}
	return TypeAny
}
