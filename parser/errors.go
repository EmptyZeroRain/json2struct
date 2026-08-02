package parser

import "errors"

var (
	ErrMaxBytes       = errors.New("parser: maximum input bytes exceeded")
	ErrMaxDepth       = errors.New("parser: maximum depth exceeded")
	ErrMaxFields      = errors.New("parser: maximum fields exceeded")
	ErrMaxArrayItems  = errors.New("parser: maximum array items exceeded")
	ErrMaxNodes       = errors.New("parser: maximum nodes exceeded")
	ErrMaxStringBytes = errors.New("parser: maximum string bytes exceeded")
	ErrMaxNumberBytes = errors.New("parser: maximum number bytes exceeded")
)
