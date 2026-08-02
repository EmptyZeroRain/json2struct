package parser

import "github.com/EmptyZeroRain/json2struct/option"

type Options struct {
	Limits        option.Limits
	DuplicateKeys DuplicateKeyPolicy
}

type DuplicateKeyPolicy string

const (
	DuplicateKeyLast  DuplicateKeyPolicy = "last"
	DuplicateKeyFirst DuplicateKeyPolicy = "first"
	DuplicateKeyError DuplicateKeyPolicy = "error"
)
