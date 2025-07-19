package igenerator

import (
	"github.com/Open-KO/kodb-godef/jsonSchema"
)

type Identifier interface {
	GetType(property jsonSchema.Column) (string, error)
}
