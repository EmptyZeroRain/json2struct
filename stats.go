package json2struct

import "time"

type Stats struct {
	Samples  int64
	Bytes    int64
	Nodes    int64
	Fields   int64
	Duration time.Duration
}
