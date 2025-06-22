package igenerator

import (
	"github.com/kenner2/OpenKO-db/jsonSchema"
)

type Template interface {
	SetTableDef(def jsonSchema.TableDef)
	AddMethod(def MethodDef)
	AddInclude(string)
	Generate() (string, error)
	GetFileName() string
	AddConst(string, string)
}

type MethodDef struct {
	ReturnType  string
	Name        string
	Params      [][]string
	Body        string
	Description string
}
