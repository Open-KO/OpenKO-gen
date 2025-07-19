package cppGen

import (
	"fmt"
	"github.com/Open-KO/kodb-godef/enums/dbType"
	"github.com/Open-KO/kodb-godef/enums/profile"
	"github.com/Open-KO/kodb-godef/enums/tsql"
	"github.com/Open-KO/kodb-godef/jsonSchema"
	"openko-gen/igenerator"
	"openko-gen/utils"
	"regexp"
	"strings"
)

const (
	fileNameFmt string = "%[1]s.ixx"

	// 1. {ns}Model or {ns}Binder
	primaryModuleFileName = "%s.ixx"
	moduleSuffixModelFmt  = "%sModel"
	moduleSuffixBinderFmt = "%sBinder"
)

type CppTemplate struct {
	def     jsonSchema.TableDef
	methods []string
	// we use includes for both #include and import
	// true map val = include
	// false map val = import
	includes       map[string]bool
	consts         map[string]string
	additionalCode []string
	namespace      string
	moduleSuffix   string
	moduleDef      ModuleDef
}

type AggregateUnion struct {
	union         jsonSchema.Union
	firstColName  string
	columnPattern string
	cppType       string
	defs          []jsonSchema.Column
}

func (d *CppTemplate) SetTableDef(def jsonSchema.TableDef) {
	d.def = def
}

func (d *CppTemplate) AddMethod(def igenerator.MethodDef) {
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

	if def.IsThrow {
		pure += " noexcept(false)"
	}

	returnType := ""
	if len(def.ReturnType) > 0 {
		returnType = fmt.Sprintf("%s ", def.ReturnType)
	}

	// 1. description
	// 2. modifiers (static, inline, etc)
	// 3. return type
	// 4. function name
	// 5. params, csv
	// 6. function body
	d.methods = append(d.methods, fmt.Sprintf(methodFmt, def.Description, modifiers, returnType, def.Name, params, def.Body, pure))
}

func (d *CppTemplate) AddInclude(s string) {
	if d.includes == nil {
		d.includes = make(map[string]bool)
	}

	key := ""
	if strings.Contains(s, "\"") || strings.Contains(s, "<") {
		key = fmt.Sprintf(includeFmt, s)
		d.includes[key] = true
	} else {
		key = fmt.Sprintf(importFmt, s)
		d.includes[key] = false
	}
}

func (d *CppTemplate) Generate() (string, error) {
	if d.def.ClassName == "" {
		return "", fmt.Errorf("className not set")
	}

	return d.GenerateModelClass()
}

func (d *CppTemplate) GenerateModelClass() (string, error) {
	// identifier is used to assign the correct c++ type from the columns' tsql.TsqlType
	identifier := CppIdentifier{}

	// fieldStrBuilder will collect all of our generated model class members
	fieldStrBuilder := strings.Builder{}

	unionAggregates := make(map[string]AggregateUnion)
	fieldToUnionPatterns := make(map[string]string)
	for i := range d.def.Unions {
		unionAggregates[d.def.Unions[i].ColumnPattern] = AggregateUnion{
			union: d.def.Unions[i],
			defs:  []jsonSchema.Column{},
		}
	}

	// initial column pass to pre-build aggregate union info
	// so we can just dump the union in place of the first matching member
	// and skip the rest.
	for i := range d.def.Columns {
		field := &d.def.Columns[i]

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

		// does the field column name match any union patterns?
		for pattern, unionAggregate := range unionAggregates {
			matched, err := regexp.Match(pattern, []byte(field.Name))
			if err != nil {
				return "", err
			}
			if matched {
				unionAggregate.defs = append(unionAggregate.defs, *field)
				unionAggregate.cppType = cppType

				if len(unionAggregate.firstColName) == 0 {
					unionAggregate.firstColName = field.Name
				}

				unionAggregate.columnPattern = field.Name
				fieldToUnionPatterns[field.Name] = pattern
			}
			unionAggregates[pattern] = unionAggregate
		}

	}

	for i := range d.def.Columns {
		field := &d.def.Columns[i]

		// skip union member definitions for all but the first column in the group
		unionPattern, isUnionMember := fieldToUnionPatterns[field.Name]
		if isUnionMember && unionAggregates[unionPattern].firstColName != field.Name {
			continue
		}

		if i > 0 {
			fieldStrBuilder.WriteString("\n")
		}

		cppType, err := identifier.GetType(*field)
		if err != nil {
			return "", err
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
				val.WriteString(fmt.Sprintf("\t%s = %s%s", d.def.Columns[i].Enums[j].Name, d.def.Columns[i].Enums[j].Value, comma))
				if d.def.Columns[i].Enums[j].Comment != "" {
					val.WriteString(fmt.Sprintf(" ///< %s", d.def.Columns[i].Enums[j].Comment))
				}
				vals = append(vals, val.String())
			}
			enumName := fmt.Sprintf("Enum%s", d.def.Columns[i].PropertyName)
			enum = fmt.Sprintf(enumFmt, enumName, strings.Join(vals, ""), d.def.Columns[i].Name)
		}

		indentLevel := 2

		// at this point, we're the first union member -- all others are skipped
		// we should include them all here.
		if isUnionMember {
			unionAggregate := unionAggregates[unionPattern]

			colList := strings.Builder{}
			for j := range unionAggregate.defs {
				// subsequent members should be separated by a newline
				if colList.Len() > 0 {
					colList.WriteString("\n\n")
				}

				colList.WriteString(d.GenerateModelMember(*field, unionAggregate.defs[j], unionAggregate.cppType, hasEnums, enum, isUnionMember))
			}

			// create a doxygen block
			doxygen := strings.Builder{}

			doxygen.WriteString(utils.FormatAndIndentLines(indentLevel, unionArrayDoxygenFmt, unionAggregate.firstColName, unionAggregate.columnPattern, unionAggregate.union.PropertyName))
			unionArrayInitializer := getInitializer(unionAggregate.cppType)
			unionArrayDef := fmt.Sprintf(unionArrayDefFmt, unionAggregate.cppType, unionAggregate.union.PropertyName, len(unionAggregate.defs), unionArrayInitializer)

			fieldStrBuilder.WriteString(fmt.Sprintf(
				unionArrayFmt,
				doxygen.String(),
				utils.FormatAndIndentLines(indentLevel+1, "%s", unionArrayDef),
				utils.FormatAndIndentLines(indentLevel+2, "%s", colList.String())))
		} else {
			fieldStrBuilder.WriteString("\n")
			member := d.GenerateModelMember(*field, *field, cppType, hasEnums, enum, isUnionMember)
			fieldStrBuilder.WriteString(utils.FormatAndIndentLines(indentLevel, "%s", member))
		}
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

func (d *CppTemplate) GenerateModelMember(firstField jsonSchema.Column, field jsonSchema.Column, cppType string, hasEnums bool, enum string, isUnionMember bool) string {

	// create a doxygen block
	doxygen := strings.Builder{}
	doxygen.WriteString(fmt.Sprintf("/// \\brief Column [%s]: %s\n", field.Name, field.Description))
	doxygen.WriteString("///\n")

	if hasEnums {
		enumName := fmt.Sprintf("Enum%s", firstField.PropertyName)
		doxygen.WriteString(fmt.Sprintf("/// \\see %s\n", enumName))
	}

	doxygen.WriteString(fmt.Sprintf("/// \\property %s\n", field.PropertyName))

	initializer := ""

	// only the first member of a union can have an initializer.
	// this should be the largest, which in our case is the array.
	// don't assign an initializer for all of the individual members.
	if !isUnionMember {
		initializer = getInitializer(cppType)
	}

	member := doxygen.String()
	member += fmt.Sprintf(memberFmt, cppType, field.PropertyName, initializer, enum)
	return member
}

func (d *CppTemplate) GenerateBinderClass() (string, error) {
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
		} else if strings.Contains(cppType, "std::time_t") {
			propBindBody = fmt.Sprintf(funcPropBindingDateCastFmt, field.PropertyName)
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

func (d *CppTemplate) GetFileName() string {
	return fmt.Sprintf(fileNameFmt, fmt.Sprintf("%s-%s", d.moduleSuffix, d.def.ClassName))
}

func (d *CppTemplate) AddConst(name string, value string) {
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
