package gormGen

import (
	"fmt"
	"github.com/Open-KO/kodb-godef/enums/tsql"
	"github.com/Open-KO/kodb-godef/jsonSchema"
)

const optionalFmt = "*%s"

var (
	// TSqlTypeMapping maps a TSqlType key to a Go-type value
	TSqlTypeMapping map[tsql.TSqlType]string
	// NoArrayTypes tracks types that may contain a length but their Go type doesn't need to be an array (think varchar<>string)
	NoArrayTypes map[tsql.TSqlType]bool
)

func init() {
	TSqlTypeMapping = make(map[tsql.TSqlType]string)
	TSqlTypeMapping[tsql.TinyInt] = "uint8"
	TSqlTypeMapping[tsql.SmallInt] = "int16"
	TSqlTypeMapping[tsql.Int] = "int"
	TSqlTypeMapping[tsql.BigInt] = "int64"
	TSqlTypeMapping[tsql.Float] = "float64"
	TSqlTypeMapping[tsql.Real] = "float32"
	TSqlTypeMapping[tsql.Char] = "byte"
	TSqlTypeMapping[tsql.Varchar] = "mssql.VarChar"
	TSqlTypeMapping[tsql.NChar] = "mssql.NChar"
	TSqlTypeMapping[tsql.NVarchar] = "string"
	TSqlTypeMapping[tsql.Binary] = "byte"
	TSqlTypeMapping[tsql.VarBinary] = "byte"
	TSqlTypeMapping[tsql.SmallDateTime] = "time.Time"
	TSqlTypeMapping[tsql.DateTime] = "time.Time"
	// text fields can be tricky - we're mapping it to a byte[]
	// for now, but it can store a very large object (up to 2GB)
	// seems like overkill for KO
	TSqlTypeMapping[tsql.Text] = "byte"
	TSqlTypeMapping[tsql.Image] = "byte"

	// if a type shouldn't be used as an array when Length is specified, add it here
	NoArrayTypes = make(map[tsql.TSqlType]bool)
	NoArrayTypes[tsql.NChar] = true
	NoArrayTypes[tsql.Varchar] = true
	NoArrayTypes[tsql.NVarchar] = true
}

type GoIdentifier struct {
}

// GetType returns the Go type associated with the column tsql.TSqlType as a string
func (this GoIdentifier) GetType(property jsonSchema.Column) (goType string, err error) {
	sqlType := property.Type
	_, IsNoArray := NoArrayTypes[sqlType]
	goType, ok := TSqlTypeMapping[sqlType]
	if !ok {
		return "", fmt.Errorf("goIdentifier.GetType - unsupported type: %s", property.Type)
	}

	if property.Length > 0 && !IsNoArray || (property.Type == tsql.Text || property.Type == tsql.Image) {
		goType = fmt.Sprintf("[]%s", goType)
		//return fmt.Sprintf("[%d]%s", property.Length, goType), nil
	}

	if goType != "" && property.AllowNull {
		goType = fmt.Sprintf(optionalFmt, goType)
	}

	return goType, nil
}

// getPropRefByType is not specific to the identifier interface, but it feels best to group all
// type identification functions in one place
func getPropRefByType(col jsonSchema.Column) (profRef string) {
	// Format a property reference to use as a parameter in the function we're returning
	pName := fmt.Sprintf("this.%s", col.PropertyName)
	if !col.AllowNull {
		pName = fmt.Sprintf("&%s", pName)
	}

	hexProtect := "false"
	if col.ForceBinary {
		hexProtect = "true"
	}

	// return the correct function for the given sqlType
	switch col.Type {
	case tsql.NVarchar:
		return fmt.Sprintf("GetOptionalStringVal(%s, %s)", pName, hexProtect)
	case tsql.Varchar:
		return fmt.Sprintf("GetOptionalVarCharVal(%s, %s)", pName, hexProtect)
	case tsql.NChar:
		return fmt.Sprintf("GetOptionalNCharVal(%s, %s)", pName, hexProtect)
	case tsql.DateTime, tsql.SmallDateTime:
		// TODO: SmallDateTime likely needs its own function/format
		return fmt.Sprintf("GetDateTimeExportFmt(%s)", pName)
	case tsql.Binary, tsql.VarBinary:
		// Binary sql types get output as hex-encoded strings
		return fmt.Sprintf("GetOptionalBinaryVal(%s)", pName)
	case tsql.Char, tsql.Text, tsql.Image:
		// Going to have to work with this carefully when I get a good example to diff against
		return fmt.Sprintf("GetOptionalByteArrayVal(%s, %s)", pName, hexProtect)
	default:
		// pass non-optional properties by reference to re-use the functions
		// Ignore slices; they are always pointers
		return fmt.Sprintf("GetOptionalDecVal(%s)", pName)
	}
}
