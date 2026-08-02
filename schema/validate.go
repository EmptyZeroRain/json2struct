package schema

import "fmt"

// Validate checks a user-supplied schema before code generation.
func Validate(root *Field, maxNodes, maxDepth int) error {
	if root == nil {
		return fmt.Errorf("schema: nil root")
	}
	type item struct {
		f     *Field
		depth int
	}
	stack := []item{{root, 0}}
	seen := make(map[*Field]bool)
	nodes := 0
	for len(stack) > 0 {
		i := len(stack) - 1
		cur := stack[i]
		stack = stack[:i]
		if cur.f == nil {
			return fmt.Errorf("schema: nil field")
		}
		if seen[cur.f] {
			return fmt.Errorf("schema: cycle detected at %q", cur.f.Name)
		}
		seen[cur.f] = true
		nodes++
		if maxNodes > 0 && nodes > maxNodes {
			return fmt.Errorf("schema: maximum nodes %d exceeded", maxNodes)
		}
		if maxDepth > 0 && cur.depth > maxDepth {
			return fmt.Errorf("schema: maximum depth %d exceeded", maxDepth)
		}
		if cur.f.Type == TypeUnknown {
			return fmt.Errorf("schema: field %q has unknown type", cur.f.Name)
		}
		if cur.f.Type != TypeObject && len(cur.f.Children) > 0 {
			return fmt.Errorf("schema: scalar field %q has children", cur.f.Name)
		}
		if cur.f.Type != TypeArray && cur.f.Element != nil {
			return fmt.Errorf("schema: non-array field %q has an element", cur.f.Name)
		}
		for _, child := range cur.f.Children {
			stack = append(stack, item{child, cur.depth + 1})
		}
		if cur.f.Element != nil {
			stack = append(stack, item{cur.f.Element, cur.depth + 1})
		}
	}
	return nil
}
