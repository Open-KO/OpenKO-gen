package igenerator

import (
	"github.com/Open-KO/kodb-godef"
)

type Identifier interface {
	GetType(property jsonSchema.Column) (string, error)
}
