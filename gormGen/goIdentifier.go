package gormGen

import (
	"fmt"
	"github.com/kenner2/OpenKO-db/jsonSchema"
	"github.com/kenner2/OpenKO-db/jsonSchema/enums/tsql"
)

// Right now we're just doing using pointers; that's what Gorm seems to expect.
const optionalFmt = "*%s"

var (
	TSqlTypeMapping map[tsql.TSqlType]string
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
	TSqlTypeMapping[tsql.Varchar] = "byte"
	TSqlTypeMapping[tsql.NChar] = "byte"
	TSqlTypeMapping[tsql.NVarchar] = "byte"
	TSqlTypeMapping[tsql.Binary] = "byte"
	TSqlTypeMapping[tsql.VarBinary] = "byte"
	TSqlTypeMapping[tsql.SmallDateTime] = "int"
	TSqlTypeMapping[tsql.DateTime] = "time.Time"
	// text fields can be tricky - we're mapping it to a byte[]
	// for now, but it can store any character set and can store a
	// very large object (up to 2GB)
	TSqlTypeMapping[tsql.Text] = "byte"
	TSqlTypeMapping[tsql.Image] = "byte"
}

type GoIdentifier struct {
}

func (this GoIdentifier) GetType(property jsonSchema.Column) (_type string, err error) {
	_type, ok := TSqlTypeMapping[property.Type]
	if !ok {
		return "", fmt.Errorf("goIdentifier.GetType - unsupported type: %s", property.Type)
	}

	if property.Type == tsql.Text || property.Type == tsql.Image {
		// don't set the capacity like other types; some of the defaults are 2GB
		return fmt.Sprintf("[]%s", _type), nil
	}

	if property.Length > 0 {
		return fmt.Sprintf("[%d]%s", property.Length, _type), nil
	} else if _type != "" && property.AllowNull {
		_type = fmt.Sprintf(optionalFmt, _type)
	}

	return _type, nil
}
