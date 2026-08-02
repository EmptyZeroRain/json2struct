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
