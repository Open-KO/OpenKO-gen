package igenerator

import (
	"github.com/Open-KO/OpenKO-db/jsonSchema"
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
	ReturnType    string
	Name          string
	Params        [][]string
	Body          string
	Description   string
	IsPtrReceiver bool
	IsStatic      bool
}
