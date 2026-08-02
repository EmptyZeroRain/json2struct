package schema

import "testing"

func TestMergeContainerConflictsBecomeAny(t *testing.T) {
	object := NewField("x", TypeObject)
	object.Children = map[string]*Field{"a": NewField("a", TypeString)}
	array := NewField("x", TypeArray)
	array.Array = true
	array.Element = NewField("", TypeString)
	for _, pair := range [][2]*Field{{NewField("x", TypeString), object}, {object, array}, {array, object}} {
		got := Merge(pair[0], pair[1])
		if got.Type != TypeAny || got.Array {
			t.Fatalf("got type=%q array=%v", got.Type, got.Array)
		}
	}
}
