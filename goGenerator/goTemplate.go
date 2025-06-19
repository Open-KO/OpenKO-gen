package goGenerator

import (
	"fmt"
	"ko-codegen/enums/genType"
	"ko-codegen/igenerator"
	"ko-codegen/jsonSchema"
	"strings"
)

const (
	// FileTemplate args:
	// 1. Class Name
	// 2. Properties
	// 3. Methods
	// 4. Class Description
	// 5. Includes
	// 6. Package Name
	// 7. Constants
	fileTemplateFmt string = `package %[6]s

import (
	"fmt"
%[5]s)

%[7]s

// %[1]s: %[4]s
type %[1]s struct
{
%[2]s
}

/* Helper Functions */

%[3]s
`

	modelDir string = "kogen"

	// MemberDefinition args:
	// 1. Type
	// 2. PropertyName
	// 3. Description
	// 4. DbName
	// 5. JsonTag
	// 6. GormTag
	propertyFmt string = "\t%[2]s %[1]s `%[6]s %[5]s`"

	// JsonTag args:
	// 1. Tag Content
	jsonTagFmt       string = "json:\"%[1]s\""
	jsonTagSeparator string = ","
	jsonTagOptional  string = "omitempty"

	// GormTag
	gormTagFmt             string = "gorm:\"%[1]s\""
	gormTagSeparator       string = ";"
	gormTagColumnNameFmt   string = "column:%[1]s"
	gormTagAutoIncrement   string = "autoIncrement"
	gormTagPrimaryKey      string = "primaryKey"
	gormTagDefaultValueFmt string = "default:%[1]s"
	gormTagNotNULL         string = "not null"
	gormBinaryTagFmt       string = "type:binary(%d)"

	// gormTagUnique string = "unique"
	// gormTagIndex string = "index"
	// gormTagForeignKey string = "foreignKey"
	//gormTagUniqueIndex string = "uniqueIndex"

	// HelperFunctionDefinition args:
	// 1. Return Type
	// 2. FuncName
	// 3. Parameters
	// 4. Function Body
	// 5. Description
	// 6. ClassName
	methodFmt string = `// %[2]s %[5]s
func (this *%[6]s) %[2]s(%[3]s) (%[1]s) {
%[4]s
}`

	fileNameFmt string = "%[1]s.go"

	includeFmt string = "\t\"%[1]s\"\n"

	// 1: formatted concat of constant variables
	constBlockFmt string = "const (\n%[1]s)"

	// 1: prop name
	// 2: prop value
	constPropFmt string = "\t%[1]s = %[2]s\n"
)

type GoTemplate struct {
	def      jsonSchema.TableDef
	methods  []string
	includes map[string]bool
	consts   map[string]string
}

func (this *GoTemplate) SetTableDef(def jsonSchema.TableDef) {
	this.def = def
}

/** Template interface impl functions **/
func (this *GoTemplate) AddInclude(s string) {
	if this.includes == nil {
		this.includes = make(map[string]bool)
	}
	this.includes[s] = true
}

func (this *GoTemplate) AddMethod(def igenerator.MethodDef) {
	params := ""
	for i := range def.Params {
		params += fmt.Sprintf("%s %s", def.Params[i][0], def.Params[i][1])
	}

	this.methods = append(this.methods, fmt.Sprintf(methodFmt, def.ReturnType, def.Name, params, def.Body, def.Description, this.def.ClassName))
}

func (this *GoTemplate) Generate() (string, error) {
	if this.def.ClassName == "" {
		return "", fmt.Errorf("Class name not set")
	}

	inclStr := ""
	for k, _ := range this.includes {
		inclStr += fmt.Sprintf(includeFmt, k)
	}

	constStr := ""
	for k, v := range this.consts {
		constStr += fmt.Sprintf(constPropFmt, k, v)
	}
	constBlock := ""
	if constStr != "" {
		constBlock = fmt.Sprintf(constBlockFmt, constStr)
	}

	identifier := GoIdentifier{}

	propStr := ""
	for i := range this.def.Columns {
		prop := &this.def.Columns[i]
		if i > 0 {
			propStr += "\n"
		}

		// generate json tag
		jsonOpt := []string{prop.Name}
		if prop.AllowNull {
			jsonOpt = append(jsonOpt, jsonTagOptional)
		}
		jsonTag := fmt.Sprintf(jsonTagFmt, strings.Join(jsonOpt, jsonTagSeparator))

		// generate gorm tag
		gormTags := []string{fmt.Sprintf(gormTagColumnNameFmt, prop.Name)}
		if prop.Type == genType.BINARY {
			gormTags = append(gormTags, fmt.Sprintf(gormBinaryTagFmt, prop.MaxLength))
		}
		if prop.IsPrimaryKey {
			gormTags = append(gormTags, gormTagPrimaryKey)
		}
		if !prop.AllowNull {
			gormTags = append(gormTags, gormTagNotNULL)
		}
		if prop.DefaultValue != "" {
			gormTags = append(gormTags, fmt.Sprintf(gormTagDefaultValueFmt, prop.DefaultValue))
		}
		gormTag := fmt.Sprintf(gormTagFmt, strings.Join(gormTags, gormTagSeparator))

		_type, err := identifier.GetType(*prop)
		if err != nil {
			return "", err
		}
		propStr += fmt.Sprintf(propertyFmt, _type, prop.PropertyName, prop.Description, prop.Name, jsonTag, gormTag)
	}

	methodStr := ""
	for i := range this.methods {
		if i > 0 {
			methodStr += "\n\n"
		}
		methodStr += this.methods[i]
	}

	return fmt.Sprintf(fileTemplateFmt, this.def.ClassName, propStr, methodStr, this.def.Description, inclStr, modelDir, constBlock), nil
}

func (this *GoTemplate) GetFileName() string {
	return fmt.Sprintf(fileNameFmt, this.def.ClassName)
}

func (this *GoTemplate) AddConst(name string, value string) {
	if this.consts == nil {
		this.consts = make(map[string]string)
	}
	this.consts[name] = value
}
