package igenerator

import (
	"github.com/kenner2/OpenKO-db/jsonSchema"
)

type Identifier interface {
	GetType(property jsonSchema.Column) (string, error)
}
