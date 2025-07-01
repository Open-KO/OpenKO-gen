package gormGen

import (
	"fmt"
	"github.com/Open-KO/OpenKO-db/jsonSchema"
	"openko-gen/igenerator"
	"sort"
	"strings"
)

const (
	modelDir    string = "kogen"
	fileNameFmt string = "%[1]s.go"
)

type GoTemplate struct {
	def            jsonSchema.TableDef
	methods        []string
	includes       map[string]bool
	consts         map[string]string
	additionalCode []string
}

func (this *GoTemplate) SetTableDef(def jsonSchema.TableDef) {
	this.def = def
}

/** Template interface impl functions **/
// AddInclude adds an include with standard formatting
func (this *GoTemplate) AddInclude(s string) {
	if this.includes == nil {
		this.includes = make(map[string]bool)
	}
	key := fmt.Sprintf(includeFmt, s)
	this.includes[key] = true
}

// AddIncludeAs adds an include without any formatting
func (this *GoTemplate) AddIncludeAs(s string) {
	if this.includes == nil {
		this.includes = make(map[string]bool)
	}
	this.includes[s] = true
}

// AddMethod parses a MethodDef into a model object method
func (this *GoTemplate) AddMethod(def igenerator.MethodDef) {
	params := ""
	for i := range def.Params {
		if i > 0 {
			params += ", "
		}
		params += fmt.Sprintf("%s %s", def.Params[i][0], def.Params[i][1])
	}

	receiverType := this.def.ClassName
	if def.IsPtrReceiver {
		receiverType = fmt.Sprintf("*%s", receiverType)
	}

	this.methods = append(this.methods, fmt.Sprintf(methodFmt, def.ReturnType, def.Name, params, def.Body, def.Description, receiverType))
}

// Generate uses the inputs added via the other interface methods to return a string containing the generated Go code
func (this *GoTemplate) Generate() (string, error) {
	if this.def.ClassName == "" {
		return "", fmt.Errorf("className not set")
	}

	// we store the constants in a hashmap, which means order isn't preserved.
	// we have to sort the consts, otherwise we'll get random/false-diffs between codegens.
	var constSort []string
	for k, v := range this.consts {
		constSort = append(constSort, fmt.Sprintf(constPropFmt, k, v))
	}
	sort.Strings(constSort)
	constStr := strings.Join(constSort, "")
	constBlock := ""
	if constStr != "" {
		constBlock = fmt.Sprintf(constBlockFmt, constStr)
	}

	// identifier is used to assign the correct Go type from the columns' tsql.TsqlType
	identifier := GoIdentifier{}

	pkMap := make(map[string]bool)
	for i := range this.def.Indexes {
		if this.def.Indexes[i].IsPrimaryKey {
			for j := range this.def.Indexes[i].Columns {
				pkMap[this.def.Indexes[i].Columns[j]] = true
			}
		}
	}

	// fieldStrBuilder will collect all of our generated model struct fields
	fieldStrBuilder := strings.Builder{}
	for i := range this.def.Columns {
		field := &this.def.Columns[i]
		if i > 0 {
			fieldStrBuilder.WriteString("\n")
		}

		// sanity check on length MSSQL only allows 0-8000
		if field.Length < 0 {
			field.Length = 0
		} else if field.Length > 8000 {
			field.Length = 8000
		}

		goType, err := identifier.GetType(*field)
		if err != nil {
			return "", err
		}

		// we store includes in a map; duplicates are prevented
		if strings.HasSuffix(goType, "time.Time") {
			this.AddInclude("time")
		} else if strings.Contains(goType, "mssql.") {
			this.AddIncludeAs("\tmssql \"github.com/microsoft/go-mssqldb\"\n")
		}

		gormType := field.GormType()

		// generate json tag
		jsonOpt := []string{field.Name}
		if field.AllowNull {
			jsonOpt = append(jsonOpt, jsonTagOptional)
		}
		jsonTag := fmt.Sprintf(jsonTagFmt, strings.Join(jsonOpt, jsonTagSeparator))

		// generate gorm tag
		gormTags := []string{fmt.Sprintf(gormTagColumnNameFmt, field.Name)}
		collation := ""
		if field.CollationName != nil && *field.CollationName != "" {
			collation = fmt.Sprintf(" COLLATE %s", *field.CollationName)
		}
		gormTags = append(gormTags, fmt.Sprintf(gormTypeTagFmt, gormType, collation))

		if _, ok := pkMap[field.Name]; ok {
			gormTags = append(gormTags, gormTagPrimaryKey)
		}
		if !field.AllowNull {
			gormTags = append(gormTags, gormTagNotNULL)
		}
		if field.DefaultValue != "" {
			gormTags = append(gormTags, fmt.Sprintf(gormTagDefaultValueFmt, field.DefaultValue))
		}
		gormTag := fmt.Sprintf(gormTagFmt, strings.Join(gormTags, gormTagSeparator))

		fieldStrBuilder.WriteString(fmt.Sprintf(propertyFmt, goType, field.PropertyName, field.Description, field.Name, jsonTag, gormTag))
	}

	methodStr := ""
	for i := range this.methods {
		if i > 0 {
			methodStr += "\n\n"
		}
		methodStr += this.methods[i]
	}

	inclStr := ""
	for k := range this.includes {
		inclStr += k
	}
	// extraCode contains raw snippets appended to the end of the file
	extraCode := strings.Join(this.additionalCode, "\n")

	return fmt.Sprintf(fileTemplateFmt, this.def.ClassName, fieldStrBuilder.String(), methodStr, this.def.Description, inclStr, modelDir, constBlock, extraCode), nil
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
