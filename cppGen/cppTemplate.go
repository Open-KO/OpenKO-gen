package cppGen

import (
	"fmt"
	"github.com/Open-KO/OpenKO-db/jsonSchema"
	"github.com/Open-KO/OpenKO-db/jsonSchema/enums/dbType"
	"github.com/Open-KO/OpenKO-db/jsonSchema/enums/profile"
	"github.com/Open-KO/OpenKO-db/jsonSchema/enums/tsql"
	"openko-gen/igenerator"
	"sort"
	"strings"
)

const (
	fileNameFmt string = "%[1]s.ixx"

	// 1. {ns}Model or {ns}Binder
	primaryModuleFileName = "%s.ixx"
	moduleSuffixModelFmt  = "%sModel"
	moduleSuffixBinderFmt = "%sBinder"
)

type DoxygenTemplate struct {
	def            jsonSchema.TableDef
	methods        []string
	includes       map[string]bool
	consts         map[string]string
	additionalCode []string
	namespace      string
	moduleSuffix   string
	moduleDef      ModuleDef
}

func (d *DoxygenTemplate) SetTableDef(def jsonSchema.TableDef) {
	d.def = def
}

func (d *DoxygenTemplate) AddMethod(def igenerator.MethodDef) {
	params := ""
	for i := range def.Params {
		if i > 0 {
			params += ", "
		}
		params += fmt.Sprintf("%s %s", def.Params[i][0], def.Params[i][1])
	}

	modifiers := ""
	if def.IsStatic {
		modifiers = "static "
	}

	pure := ""
	if def.IsPure {
		pure = " const"
	}

	// 1. description
	// 2. modifiers (static, inline, etc)
	// 3. return type
	// 4. function name
	// 5. params, csv
	// 6. function body
	d.methods = append(d.methods, fmt.Sprintf(methodFmt, def.Description, modifiers, def.ReturnType, def.Name, params, def.Body, pure))
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

	fullClassDef, err := d.GenerateModelClass()
	if err != nil {
		return "", err
	}

	var includes []string
	for include := range d.includes {
		includes = append(includes, include)
	}
	// includes is built from an unordered hash map
	// sort the includes alphabetically so they don't cause diffs gen-to-gen
	sort.Strings(includes)

	binderNs := fmt.Sprintf(profile.BinderNsFmt, d.moduleDef.namespace)
	fileStr := fmt.Sprintf(modelFileFmt, d.def.ClassName, fullClassDef, binderNs, d.namespace)
	return fmt.Sprintf(partitionModuleFmt, d.def.ClassName, fileStr, strings.Join(includes, ""), d.moduleSuffix), nil
}

func (d *DoxygenTemplate) GenerateModelClass() (string, error) {

	// identifier is used to assign the correct c++ type from the columns' tsql.TsqlType
	identifier := CppIdentifier{}

	// fieldStrBuilder will collect all of our generated model class members
	fieldStrBuilder := strings.Builder{}
	for i := range d.def.Columns {
		field := &d.def.Columns[i]
		if i > 0 {
			fieldStrBuilder.WriteString("\n")
		}

		// sanity check on length MSSQL only allows 0-8000
		if field.Length < 0 {
			field.Length = 0
		} else if field.Length > 8000 {
			field.Length = 8000
		}

		cppType, err := identifier.GetType(*field)
		if err != nil {
			return "", err
		}

		// add type-specific imports as needed
		// we store includes in a map; duplicates are prevented
		if strings.Contains(cppType, "int") {
			d.AddInclude("<cstdint>")
		} else if strings.Contains(cppType, "time_t") {
			d.AddInclude("<ctime>")
		}
		if strings.Contains(cppType, "std::optional") {
			d.AddInclude("<optional>")
		}

		enum := ""
		hasEnums := len(d.def.Columns[i].Enums) > 0 && isEnumType(cppType)
		if hasEnums {
			var vals []string
			for j := range d.def.Columns[i].Enums {
				val := strings.Builder{}
				if j > 0 {
					val.WriteString("\n")
				}
				comma := ""
				if j < len(d.def.Columns[i].Enums)-1 {
					comma = ","
				}
				val.WriteString(fmt.Sprintf("\t\t\t%s = %s%s", d.def.Columns[i].Enums[j].Name, d.def.Columns[i].Enums[j].Value, comma))
				if d.def.Columns[i].Enums[j].Comment != "" {
					val.WriteString(fmt.Sprintf(" ///< %s", d.def.Columns[i].Enums[j].Comment))
				}
				vals = append(vals, val.String())
			}
			enumName := fmt.Sprintf("Enum%s", d.def.Columns[i].PropertyName)
			enum = fmt.Sprintf(enumFmt, enumName, strings.Join(vals, ""), d.def.Columns[i].Name)
		}

		// create a doxygen block
		doxygen := strings.Builder{}
		doxygen.WriteString(fmt.Sprintf("/// \\brief Column [%s]: %s\n", field.Name, field.Description))
		doxygen.WriteString("\t\t///\n")
		if hasEnums {
			enumName := fmt.Sprintf("Enum%s", d.def.Columns[i].PropertyName)
			doxygen.WriteString(fmt.Sprintf("\t\t/// \\see %s\n", enumName))
		}
		doxygen.WriteString(fmt.Sprintf("\t\t/// \\property %s", field.PropertyName))

		initializer := getInitializer(cppType)

		fieldStrBuilder.WriteString(fmt.Sprintf(memberFmt, doxygen.String(), cppType, field.PropertyName, initializer, enum))
	}

	methods := strings.Builder{}
	for i := range d.methods {
		if i > 0 {
			methods.WriteString("\n")
		}
		methods.WriteString(d.methods[i])
	}

	doxygen := strings.Builder{}
	doxygen.WriteString(fmt.Sprintf("\t/// \\brief [%s] %s\n", d.def.Name, d.def.Description))
	doxygen.WriteString(fmt.Sprintf("\t/// \\class %s\n", d.def.ClassName))
	doxygen.WriteString(fmt.Sprintf(getDbTypeXRefFmt(d.def.Database), d.def.Name, d.def.Description))

	binderNs := fmt.Sprintf(profile.BinderNsFmt, d.moduleDef.namespace)
	return fmt.Sprintf(modelClassFmt, d.def.ClassName, fieldStrBuilder.String(), methods.String(), doxygen.String(), binderNs), nil
}

func (d *DoxygenTemplate) GenerateBinders() (string, error) {
	if d.def.ClassName == "" {
		return "", fmt.Errorf("className not set")
	}

	classStr, err := d.GenerateBinderClass()
	if err != nil {
		return "", err
	}

	var includes []string
	for include := range d.includes {
		includes = append(includes, include)
	}
	// includes is built from an unordered hash map
	// sort the includes alphabetically so they don't cause diffs gen-to-gen
	sort.Strings(includes)

	fileStr := fmt.Sprintf(binderFileFmt, classStr, d.namespace, d.moduleDef.OutDir)
	return fmt.Sprintf(partitionModuleFmt, d.def.ClassName, fileStr, strings.Join(includes, ""), d.moduleSuffix), nil
}

func (d *DoxygenTemplate) GenerateBinderClass() (string, error) {
	// identifier is used to assign the correct c++ type from the columns' tsql.TsqlType
	identifier := CppIdentifier{}
	modelNs := fmt.Sprintf(profile.ModelNsFmt, d.moduleDef.namespace)

	// fieldStrBuilder will collect all of our generated model class members
	fieldStrBuilder := strings.Builder{}
	for i := range d.def.Columns {
		field := &d.def.Columns[i]
		if i > 0 {
			fieldStrBuilder.WriteString("\n")
		}

		cppType, err := identifier.GetType(*field)
		if err != nil {
			return "", err
		}

		// add binding method
		var propBindBody string
		_type := stripOptional(cppType)
		if field.Type == tsql.TinyInt {
			upcast := "int16_t"
			propBindBody = fmt.Sprintf(funcPropBindingUpCastFmt, _type, upcast, field.PropertyName)
		} else if field.AllowNull {
			propBindBody = fmt.Sprintf(funcPropBindingGetFmt, _type, field.PropertyName)
		} else {
			propBindBody = fmt.Sprintf(funcPropBindingFmt, _type, field.PropertyName)
		}
		if field.AllowNull {
			propBindBody = fmt.Sprintf(funcOptionalPropBindingFmt, _type, field.PropertyName, propBindBody)
		}
		propBindDef := igenerator.MethodDef{
			IsStatic:   true,
			ReturnType: "void",
			Params: [][]string{
				{fmt.Sprintf("%s::%s&", modelNs, d.def.ClassName), "m"},
				{"const nanodbc::result&", "result"},
				{"short", "colIndex"},
			},
			Name:        fmt.Sprintf("Bind%s", field.PropertyName),
			Body:        propBindBody,
			Description: fmt.Sprintf("Binds a result's column to %s", field.PropertyName),
		}
		d.AddMethod(propBindDef)
	}

	methods := strings.Builder{}
	for i := range d.methods {
		if i > 0 {
			methods.WriteString("\n")
		}
		methods.WriteString(d.methods[i])
	}

	return fmt.Sprintf(binderClassFmt, d.def.ClassName, methods.String(), modelNs), nil
}

func (d *DoxygenTemplate) GetFileName() string {
	return fmt.Sprintf(fileNameFmt, fmt.Sprintf("%s-%s", d.moduleSuffix, d.def.ClassName))
}

func (d *DoxygenTemplate) AddConst(name string, value string) {
	if d.consts == nil {
		d.consts = make(map[string]string)
	}
	d.consts[name] = value
}

func isEnumType(cppType string) bool {
	_type := strings.Replace(cppType, "std::optional<", "", -1)
	_type = strings.Replace(_type, ">", "", -1)

	switch _type {
	case "uint8_t", "int16_t", "int32_t", "int64_t":
		return true
	}

	return false
}

func getDbTypeXRefFmt(databaseType dbType.DbType) string {
	switch databaseType {
	case dbType.ACCOUNT:
		return "\t/// \\xrefitem acctdb \"Account Database\" \"Account Database\" %s %s"
	case dbType.GAME:
		return "\t/// \\xrefitem gamedb \"Game Database\" \"Game Database\" %s %s"
	case dbType.LOG:
		return "\t/// \\xrefitem logdb \"Log Database\" \"Log Database\" %s %s"
	}

	return ""
}
