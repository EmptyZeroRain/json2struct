package schema

// Merge combines two observations without mutating either input.
func Merge(a, b *Field) *Field {
	if a == nil {
		return b.Clone()
	}
	if b == nil {
		return a.Clone()
	}
	r := a.Clone()
	return MergeInto(r, b)
}

// MergeInto merges b into a in place. The caller must exclusively own a.
func MergeInto(a, b *Field) *Field {
	if a == nil {
		return b.Clone()
	}
	if b == nil {
		return a
	}
	a.Required = a.Required && b.Required
	a.Nullable = a.Nullable || b.Nullable || a.Type == TypeNull || b.Type == TypeNull
	if a.Type == TypeNull {
		a.Type = b.Type
	}
	if b.Type == TypeNull {
		return a
	}

	// A conflict is terminal: do not reinterpret scalar+object or scalar+array
	// as a container, otherwise the generated model silently loses the conflict.
	if a.Type == TypeAny || b.Type == TypeAny {
		a.Type = TypeAny
		a.Array = false
		return a
	}
	if a.Type != b.Type {
		if (a.Type == TypeInteger && b.Type == TypeNumber) || (a.Type == TypeNumber && b.Type == TypeInteger) {
			a.Type = TypeNumber
		} else {
			a.Type = TypeAny
			a.Array = false
		}
		return a
	}

	switch a.Type {
	case TypeObject:
		if a.Children == nil {
			a.Children = make(map[string]*Field)
		}
		for name, child := range a.Children {
			if _, ok := b.Children[name]; !ok {
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
	case TypeArray:
		a.Array = true
		if a.Element == nil && b.Element != nil {
			a.Element = b.Element.Clone()
		} else if a.Element != nil && b.Element != nil {
			MergeInto(a.Element, b.Element)
		}
	}
	return a
}
