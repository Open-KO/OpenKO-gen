package igenerator

import (
	"github.com/Open-KO/OpenKO-db/jsonSchema"
)

type Identifier interface {
	GetType(property jsonSchema.Column) (string, error)
}
