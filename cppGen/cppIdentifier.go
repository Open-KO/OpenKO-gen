package cppGen

import (
	"fmt"
	"github.com/Open-KO/OpenKO-db/jsonSchema"
	"github.com/Open-KO/OpenKO-db/jsonSchema/enums/tsql"
	"strings"
)

const optionalFmt = "std::optional<%s>"

var (
	// TSqlTypeMapping maps a TSqlType key to a Go-type value
	TSqlTypeMapping map[tsql.TSqlType]string
	// NoArrayTypes tracks types that may contain a length but their c++ type doesn't need to be an array (think varchar<>std::string)
	NoArrayTypes map[tsql.TSqlType]bool
)

func init() {
	TSqlTypeMapping = make(map[tsql.TSqlType]string)
	TSqlTypeMapping[tsql.TinyInt] = "uint8_t"
	TSqlTypeMapping[tsql.SmallInt] = "int16_t"
	TSqlTypeMapping[tsql.Int] = "int32_t"
	TSqlTypeMapping[tsql.BigInt] = "int64_t"
	TSqlTypeMapping[tsql.Float] = "double"
	TSqlTypeMapping[tsql.Real] = "float"
	TSqlTypeMapping[tsql.Char] = "std::vector<uint8_t>"
	TSqlTypeMapping[tsql.Varchar] = "std::string"
	// std::vector<uint8_t> feels right (if nanodbc can even bind it), but how would
	// an optional work?  Wouldn't be able to tell the difference between "" and NULL
	TSqlTypeMapping[tsql.Binary] = "std::vector<uint8_t>"
	TSqlTypeMapping[tsql.VarBinary] = "std::vector<uint8_t>"
	TSqlTypeMapping[tsql.SmallDateTime] = "std::time_t"
	TSqlTypeMapping[tsql.DateTime] = "std::time_t"
	// text fields can be tricky - we're mapping it to a byte[]
	// for now, but it can store a very large object (up to 2GB)
	// seems like overkill for KO
	TSqlTypeMapping[tsql.Text] = "std::string"
	TSqlTypeMapping[tsql.Image] = "std::vector<uint8_t>"

	// if a type shouldn't be used as an array when Length is specified, add it here
	NoArrayTypes = make(map[tsql.TSqlType]bool)
	NoArrayTypes[tsql.Varchar] = true
	NoArrayTypes[tsql.Text] = true
}

type CppIdentifier struct {
}

// GetType returns the Go type associated with the column tsql.TSqlType as a string
func (this CppIdentifier) GetType(property jsonSchema.Column) (cppType string, err error) {
	sqlType := property.Type
	//_, IsNoArray := NoArrayTypes[sqlType]
	cppType, ok := TSqlTypeMapping[sqlType]
	if !ok {
		return "", fmt.Errorf("cppIdentifier.GetType - unsupported type: %s", property.Type)
	}

	if property.IsHexProtect {
		cppType = "std::vector<uint8_t>"
	}

	// IsNoArray is a schema type that has a length but uses a data type that doesn't require it to be specified
	// We're using vectors right now, so this whole format isn't needed
	// if we want to implement MaxLength allocators for the vectors at some point in the future
	// this is where we would do it
	//if property.Length > 0 && !IsNoArray {
	//	cppType = fmt.Sprintf("%s[]", cppType)
	//	//return fmt.Sprintf("[%d]%s", property.Length, cppType), nil
	//}

	if cppType != "" && property.AllowNull {
		cppType = fmt.Sprintf(optionalFmt, cppType)
	}

	return cppType, nil
}

func getInitializer(cppType string) string {
	if strings.Contains(cppType, "std::vector") || strings.Contains(cppType, "std::optional") {
		return ""
	}

	switch cppType {
	case "uint8_t", "int16_t", "int32_t", "int64_t", "double", "float":
		return " = {}"
	default:
		return ""
	}
}

func stripOptional(cppType string) string {
	_type := strings.Replace(cppType, "std::optional<", "", 1)
	if len(cppType) == len(_type) {
		return cppType
	} else {
		return strings.Replace(_type, ">", "", 1)
	}
}
