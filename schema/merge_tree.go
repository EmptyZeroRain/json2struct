package schema

// MergeAll merges schemas using a balanced tree, reducing repeated work on
// large batches compared with a long left-fold.
func MergeAll(items []*Field) *Field {
	if len(items) == 0 {
		return nil
	}
	current := make([]*Field, len(items))
	copy(current, items)
	for len(current) > 1 {
		next := make([]*Field, 0, (len(current)+1)/2)
		for i := 0; i < len(current); i += 2 {
			if i+1 == len(current) {
				next = append(next, current[i])
			} else {
				next = append(next, Merge(current[i], current[i+1]))
			}
		}
		current = next
	}
	return current[0]
}

// MergeAllInto merges all items into dst without cloning the accumulator.
func MergeAllInto(dst *Field, items []*Field) *Field {
	for _, item := range items {
		if item != nil {
			MergeInto(dst, item)
		}
	}
	return dst
}

// CountNodes returns the number of nodes in a schema tree.
func CountNodes(root *Field) int {
	if root == nil {
		return 0
	}
	n := 0
	seen := make(map[*Field]bool)
	stack := []*Field{root}
	for len(stack) > 0 {
		i := len(stack) - 1
		f := stack[i]
		stack = stack[:i]
		if seen[f] {
			continue
		}
		seen[f] = true
		n++
		for _, c := range f.Children {
			stack = append(stack, c)
		}
		if f.Element != nil {
			stack = append(stack, f.Element)
		}
	}
	return n
}
