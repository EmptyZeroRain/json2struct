package option

// Limits prevents hostile or unexpectedly large input from exhausting memory.
// A zero value means unlimited.
type Limits struct {
	MaxBytes      int64
	MaxDepth      int
	MaxFields     int
	MaxArrayItems int
	MaxLineBytes  int
	MaxNodes      int
}
