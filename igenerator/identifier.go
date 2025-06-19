package igenerator

import (
	"ko-codegen/jsonSchema"
)

type Identifier interface {
	GetType(property jsonSchema.Column) (string, error)
}
