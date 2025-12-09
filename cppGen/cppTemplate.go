package cppGen

import (
	"fmt"
	"openko-gen/igenerator"
	"openko-gen/utils"
	"regexp"
	"strings"

	"github.com/Open-KO/kodb-godef/enums/dbType"
	"github.com/Open-KO/kodb-godef/enums/profile"
	"github.com/Open-KO/kodb-godef/jsonSchema"
)

const (
	// 1. {ns}Model or {ns}Binder
	primaryHeaderFileName = "%s.h"
	primarySourceFileName = "%s.cpp"
	classSuffixModelFmt   = "%sModel"
	classSuffixBinderFmt  = "%sBinder"
)

type AggregateArray struct {
	array         jsonSchema.Union // TODO: This should be changed to reflect that it's purely an array everywhere
	firstColName  string
	columnPattern string
	cppType       string
	defs          []jsonSchema.Column
}

type ArrayElement struct {
	pattern string
	index   int
}

type CppTemplate struct {
	def     jsonSchema.TableDef
	decls   []string
	methods []string
	// we use includes for both #include and import
	// true map val = include
	// false map val = import
	includes            map[string]bool
	consts              map[string]string
	additionalCode      []string
	namespace           string
	moduleDef           ModuleDef
	arrayAggregates     map[string]AggregateArray
	fieldToArrayElement map[string]ArrayElement
}

func (d *CppTemplate) SetTableDef(def jsonSchema.TableDef) {
	d.def = def
}

func (d *CppTemplate) AddMethod(def igenerator.MethodDef) {
	params := ""
	for i := range def.Params {
		if i > 0 {
			params += ", "
			if i%3 == 0 {
				params += "\n\t\t\t"
			}
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

	implReturnType := returnType
	if len(def.ImplReturnType) > 0 {
		implReturnType = fmt.Sprintf("%s ", def.ImplReturnType)
	}

	// 1. description
	// 2. modifiers (static, inline, etc)
	// 3. return type
	// 4. function name
	// 5. params, csv
	// 6. pure
	d.decls = append(d.decls, fmt.Sprintf(methodDeclFmt, def.Description, modifiers, returnType, def.Name, params, pure))

	// 1. description
	// 2. return type
	// 3. class name
	// 4. function name
	// 5. params, csv
	// 6. pure
	// 7. function body
	d.methods = append(d.methods, fmt.Sprintf(methodImplFmt, def.Description, implReturnType, def.ClassName, def.Name, params, pure, def.Body))
}

func (d *CppTemplate) AddInclude(s string) {
	if d.includes == nil {
		d.includes = make(map[string]bool)
	}

	key := fmt.Sprintf(includeFmt, s)
	d.includes[key] = true
}

func (d *CppTemplate) Generate() (string, string, error) {
	if d.def.ClassName == "" {
		return "", "", fmt.Errorf("className not set")
	}

	return d.GenerateModelClass()
}

func (d *CppTemplate) PrepareClassForGeneration() error {
	d.arrayAggregates = make(map[string]AggregateArray)
	d.fieldToArrayElement = make(map[string]ArrayElement)
	for i := range d.def.Unions {
		d.arrayAggregates[d.def.Unions[i].ColumnPattern] = AggregateArray{
			array: d.def.Unions[i],
			defs:  []jsonSchema.Column{},
		}
	}

	// identifier is used to assign the correct c++ type from the columns' tsql.TsqlType
	identifier := CppIdentifier{}

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
			return err
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

		// does the field column name match any array patterns?
		for pattern, arrayAggregate := range d.arrayAggregates {
			matched, err := regexp.Match(pattern, []byte(field.Name))
			if err != nil {
				return err
			}

			if matched {
				arrayIndex := len(arrayAggregate.defs)
				arrayAggregate.defs = append(arrayAggregate.defs, *field)
				arrayAggregate.cppType = cppType

				if len(arrayAggregate.firstColName) == 0 {
					arrayAggregate.firstColName = field.Name
				}

				arrayAggregate.columnPattern = field.Name

				arrayElement := ArrayElement{}
				arrayElement.pattern = pattern
				arrayElement.index = arrayIndex

				d.fieldToArrayElement[field.Name] = arrayElement
			}

			d.arrayAggregates[pattern] = arrayAggregate
		}

	}
	return nil
}

func (d *CppTemplate) GenerateModelClass() (string, string, error) {
	err := d.PrepareClassForGeneration()
	if err != nil {
		return "", "", err
	}

	// identifier is used to assign the correct c++ type from the columns' tsql.TsqlType
	identifier := CppIdentifier{}

	// fieldStrBuilder will collect all of our generated model class members
	fieldStrBuilder := strings.Builder{}

	for i := range d.def.Columns {
		field := &d.def.Columns[i]

		// skip definitions for all but the first column in the group of 'array' columns
		arrayElement, isPartOfArray := d.fieldToArrayElement[field.Name]
		if isPartOfArray && d.arrayAggregates[arrayElement.pattern].firstColName != field.Name {
			continue
		}

		if i > 0 {
			fieldStrBuilder.WriteString("\n")
		}

		cppType, err := identifier.GetType(*field)
		if err != nil {
			return "", "", err
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

		// at this point, we're the first column in an array -- all others are skipped
		// we should include them all here.
		if isPartOfArray {
			arrayAggregate := d.arrayAggregates[arrayElement.pattern]

			colDocList := strings.Builder{}
			for j := range arrayAggregate.defs {
				colDocList.WriteString(fmt.Sprintf("/// Column [%s]: %s", arrayAggregate.defs[j].Name, arrayAggregate.defs[j].Description))

				if j != len(arrayAggregate.defs)-1 {
					colDocList.WriteString("\n")
				}
			}

			// create a doxygen block
			doxygen := strings.Builder{}

			doxygen.WriteString(utils.FormatAndIndentLines(indentLevel, arrayDoxygenFmt, arrayAggregate.firstColName, arrayAggregate.columnPattern, colDocList.String(), arrayAggregate.array.PropertyName))
			unionArrayInitializer := getInitializer(arrayAggregate.cppType)
			unionArrayDef := fmt.Sprintf(arrayDefFmt, arrayAggregate.cppType, arrayAggregate.array.PropertyName, len(arrayAggregate.defs), unionArrayInitializer)

			fieldStrBuilder.WriteString(fmt.Sprintf(
				arrayFmt,
				doxygen.String(),
				utils.FormatAndIndentLines(indentLevel, "%s", unionArrayDef)))
		} else {
			fieldStrBuilder.WriteString("\n")
			member := d.GenerateModelMember(*field, *field, cppType, hasEnums, enum)
			fieldStrBuilder.WriteString(utils.FormatAndIndentLines(indentLevel, "%s", member))
		}
	}

	decls := strings.Builder{}
	for i := range d.decls {
		if i > 0 {
			decls.WriteString("\n")
		}
		decls.WriteString(d.decls[i])
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
	doxygen.WriteString(fmt.Sprintf(getDbTypeXRefFmt(d.def.Database, d.moduleDef.OutDir), d.def.Name, d.def.Description))

	binderNs := fmt.Sprintf(profile.BinderNsFmt, d.moduleDef.namespace)
	headerStr := fmt.Sprintf(modelClassHeaderFmt, d.def.ClassName, fieldStrBuilder.String(), decls.String(), doxygen.String(), binderNs)
	sourceStr := fmt.Sprintf(modelClassSourceFmt, methods.String())
	return headerStr, sourceStr, nil
}

func (d *CppTemplate) GenerateModelMember(firstField jsonSchema.Column, field jsonSchema.Column, cppType string, hasEnums bool, enum string) string {

	// create a doxygen block
	doxygen := strings.Builder{}
	doxygen.WriteString(fmt.Sprintf("/// \\brief Column [%s]: %s\n", field.Name, field.Description))
	doxygen.WriteString("///\n")

	if hasEnums {
		enumName := fmt.Sprintf("Enum%s", firstField.PropertyName)
		doxygen.WriteString(fmt.Sprintf("/// \\see %s\n", enumName))
	}

	doxygen.WriteString(fmt.Sprintf("/// \\property %s\n", field.PropertyName))

	initializer := getInitializer(cppType)

	member := doxygen.String()
	member += fmt.Sprintf(memberFmt, cppType, field.PropertyName, initializer, enum)
	return member
}

func (d *CppTemplate) GenerateBinderClass() (string, string, error) {
	d.PrepareClassForGeneration()

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
			return "", "", err
		}

		propertyName := field.PropertyName

		arrayElement, isPartOfArray := d.fieldToArrayElement[field.Name]
		if isPartOfArray {
			arrayAggregate := d.arrayAggregates[arrayElement.pattern]
			propertyName = fmt.Sprintf("%s[%d]", arrayAggregate.array.PropertyName, arrayElement.index)
		}

		// add binding method
		var propBindBody string
		_type := stripOptional(cppType)

		if strings.Contains(cppType, "std::time_t") {
			castCppType := "nanodbc::timestamp"
			castFunc := "binderUtil::CTimeFromDbTime"

			if field.AllowNull {
				propBindBody = fmt.Sprintf(funcOptionalPropBindingCastFmt, castCppType, propertyName, castFunc)
			} else {
				propBindBody = fmt.Sprintf(funcPropBindingCastFmt, castCppType, propertyName, castFunc)
			}
		} else if field.AllowNull {
			propBindBody = fmt.Sprintf(funcPropBindingFmt, cppType, propertyName)
		} else {
			propBindBody = fmt.Sprintf(funcPropBindingFmt, _type, propertyName)
		}

		propBindDef := igenerator.MethodDef{
			IsStatic:   true,
			ReturnType: "void",
			Params: [][]string{
				{fmt.Sprintf("%s::%s&", modelNs, d.def.ClassName), "m"},
				{"const nanodbc::result&", "result"},
				{"short", "colIndex"},
			},
			ClassName:   d.def.ClassName,
			Name:        fmt.Sprintf("Bind%s", field.PropertyName),
			Body:        propBindBody,
			Description: fmt.Sprintf("Binds a result's column to %s", field.PropertyName),
		}
		d.AddMethod(propBindDef)
	}

	decls := strings.Builder{}
	for i := range d.decls {
		if i > 0 {
			decls.WriteString("\n")
		}
		decls.WriteString(d.decls[i])
	}

	methods := strings.Builder{}
	for i := range d.methods {
		if i > 0 {
			methods.WriteString("\n")
		}
		methods.WriteString(d.methods[i])
	}

	headerStr := fmt.Sprintf(binderClassHeaderFmt, d.def.ClassName, decls.String(), modelNs)
	sourceStr := fmt.Sprintf(binderClassSourceFmt, methods.String(), modelNs)
	return headerStr, sourceStr, nil
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

func getDbTypeXRefFmt(databaseType dbType.DbType, extractName string) string {
	dbId := fmt.Sprintf("db_%s", databaseType)
	dbName := fmt.Sprintf("%s Database", databaseType)
	if extractName != "" {
		dbId += fmt.Sprintf("_%s", extractName)
		dbName += fmt.Sprintf(" - %s Library", extractName)
	}
	retStr := fmt.Sprintf("\t/// \\xrefitem %[1]s \"%[2]s\" \"%[2]s\"", dbId, dbName)
	return retStr + " %s %s"
}

func getProcXRefFmt(databaseType dbType.DbType) string {
	dbId := fmt.Sprintf("dbproc_%s", databaseType)
	dbName := fmt.Sprintf("%s Database Stored Procedures", databaseType)
	retStr := fmt.Sprintf("\t/// \\xrefitem %[1]s \"%[2]s\" \"%[2]s\"", dbId, dbName)
	return retStr + " %s %s"
}

func generateIncludeGuard(path string) string {
	path = strings.ReplaceAll(path, ".", "_")
	path = strings.ReplaceAll(path, "/", "_")
	path = strings.ReplaceAll(path, "\\", "_")
	path = strings.ToUpper(path)
	return path
}
