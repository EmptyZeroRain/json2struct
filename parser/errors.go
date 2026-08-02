package parser

import "errors"

type LimitKind string

const (
	LimitBytes       LimitKind = "bytes"
	LimitDepth       LimitKind = "depth"
	LimitFields      LimitKind = "fields"
	LimitArrayItems  LimitKind = "array_items"
	LimitNodes       LimitKind = "nodes"
	LimitStringBytes LimitKind = "string_bytes"
	LimitNumberBytes LimitKind = "number_bytes"
	LimitTotalBytes  LimitKind = "total_bytes"
	LimitSamples     LimitKind = "samples"
	LimitSchemaNodes LimitKind = "schema_nodes"
)

type LimitError struct {
	Kind  LimitKind
	Limit int64
	Path  string
}

func (e *LimitError) Error() string {
	if e.Path == "" {
		return "parser: maximum " + string(e.Kind) + " exceeded"
	}
	return "parser: maximum " + string(e.Kind) + " exceeded at " + e.Path
}
func (e *LimitError) Unwrap() error {
	switch e.Kind {
	case LimitBytes:
		return ErrMaxBytes
	case LimitDepth:
		return ErrMaxDepth
	case LimitFields:
		return ErrMaxFields
	case LimitArrayItems:
		return ErrMaxArrayItems
	case LimitNodes:
		return ErrMaxNodes
	case LimitStringBytes:
		return ErrMaxStringBytes
	case LimitNumberBytes:
		return ErrMaxNumberBytes
	case LimitTotalBytes:
		return ErrMaxTotalBytes
	case LimitSamples:
		return ErrMaxSamples
	case LimitSchemaNodes:
		return ErrMaxSchemaNodes
	}
	return nil
}

var (
	ErrMaxBytes       = errors.New("parser: maximum input bytes exceeded")
	ErrMaxDepth       = errors.New("parser: maximum depth exceeded")
	ErrMaxFields      = errors.New("parser: maximum fields exceeded")
	ErrMaxArrayItems  = errors.New("parser: maximum array items exceeded")
	ErrMaxNodes       = errors.New("parser: maximum nodes exceeded")
	ErrMaxStringBytes = errors.New("parser: maximum string bytes exceeded")
	ErrMaxNumberBytes = errors.New("parser: maximum number bytes exceeded")
	ErrMaxSchemaNodes = errors.New("parser: maximum schema nodes exceeded")
	ErrMaxTotalBytes  = errors.New("parser: maximum total bytes exceeded")
	ErrMaxSamples     = errors.New("parser: maximum samples exceeded")
)
