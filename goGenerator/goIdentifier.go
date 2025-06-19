package goGenerator

import (
	"fmt"
	"ko-codegen/enums/genType"
	"ko-codegen/jsonSchema"
)

// Right now we're just doing using pointers; that's what Gorm seems to expect.
const optionalFmt = "*%s"

type GoIdentifier struct {
}

func (this GoIdentifier) GetType(property jsonSchema.Column) (_type string, err error) {
	defer func() {
		// If valid, check if the return type is optional.  If it is, wrap it
		// Don't do pointers to Binary types; they're a type alias for a byte slice and slices are already pointers
		if err == nil && _type != "" && property.AllowNull {
			_type = fmt.Sprintf(optionalFmt, _type)
		}
	}()

	// genTypes are mostly based on Go Types; they directly translate with few exceptions
	switch property.Type {
	case genType.BINARY:
		return "Binary", nil
	case genType.FLOAT:
		return "float32", nil
	default:
		return string(property.Type), nil
	}
}
