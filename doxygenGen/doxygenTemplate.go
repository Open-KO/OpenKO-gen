package doxygenGen

import (
	"fmt"
	"github.com/Open-KO/OpenKO-db/jsonSchema"
	"openko-gen/igenerator"
	"strings"
)

const (
	fileNameFmt           string = "%[1]s.ixx"
	primaryModuleFileName        = "doxygen_module.ixx"
)

type DoxygenTemplate struct {
	def            jsonSchema.TableDef
	methods        []string
	includes       map[string]bool
	consts         map[string]string
	additionalCode []string
}

func (d *DoxygenTemplate) SetTableDef(def jsonSchema.TableDef) {
	d.def = def
}

func (d *DoxygenTemplate) AddMethod(def igenerator.MethodDef) {
	// TODO
}

func (d *DoxygenTemplate) AddInclude(s string) {
	if d.includes == nil {
		d.includes = make(map[string]bool)
	}
	key := fmt.Sprintf(includeFmt, s)
	d.includes[key] = true
}

func (d *DoxygenTemplate) Generate() (string, error) {
	if d.def.ClassName == "" {
		return "", fmt.Errorf("className not set")
	}

	includes := []string{}
	for include := range d.includes {
		includes = append(includes, include)
	}

	// identifier is used to assign the correct c++ type from the columns' tsql.TsqlType
	//identifier := CppIdentifier{}

	fileStr := fmt.Sprintf(modelFileFmt, d.def.ClassName)
	return fmt.Sprintf(partitionModuleFmt, d.def.ClassName, fileStr, strings.Join(includes, "")), nil
}

func (d *DoxygenTemplate) GetFileName() string {
	return fmt.Sprintf(fileNameFmt, d.def.ClassName)
}

func (d *DoxygenTemplate) AddConst(name string, value string) {
	if d.consts == nil {
		d.consts = make(map[string]string)
	}
	d.consts[name] = value
}
