package cppGenerator

/*
// This was done as a prototype using jsonSchema; it would need to be refactored using
// jsonSchema/tableDef.go, if a use case for it arises.

const optionalFmt = "Nullable<%s>"

type CppIdentifier struct {
}

func (this CppIdentifier) GetType(name string, property jsonschema.Schema, isOptional bool) (_type string, err error) {
	defer func() {
		// If valid, check if the return type is optional.  If it is, wrap it
		if err == nil && _type != "" && isOptional {
			_type = fmt.Sprintf(optionalFmt, _type)
		}
	}()

	// Multiple types for a property is some typescript nonsense.  Ignore anything past the first
	if len(property.Types) > 0 {
		_type = property.Types[0]
	} else {
		return "", fmt.Errorf("No type passed into CppIdentifier.GetType for property: %v", name)
	}

	switch _type {
	case "string":
		// TODO:  If optional, should this be a pointer to a string?  Likely going to need a null package impl
		// to check if a value is valid/set without using pointers.  This will allow us to leverage stack memory
		// over heap, which will be faster.
		return "std::string", nil
	case "number":
		// check constraints to identify correct type
		if property.MultipleOf != nil && property.MultipleOf.Num().Int64() < 1 {
			// TODO: Double vs float?  Would have to be differentiated by property.Maximum if there's a good use case for it
			return "double", nil
		}

		// whole numbers
		if property.Minimum != nil && property.Maximum != nil {
			_min := property.Minimum.Num().Int64()
			_max := property.Maximum.Num().Int64()
			if _min == 0 && _max == 255 {
				return "uint8_t", nil
			} else if _min == -127 && _max == 128 {
				return "int8_t", nil
			} else if _min == 0 && _max == 65535 {
				return "uint16_t", nil
			} else if _min == -32768 && _max == 32767 {
				return "int16_t", nil
			} else if _min == 0 && _max == 4294967295 {
				return "uint32_t", nil
			} else if _min == -2147483647 && _max == 2147483648 {
				return "int32_t", nil
			} else if _min == 0 {
				// TODO:  Checking the max on this against big.Rat is something I'd have to look into the syntax of
				return "uint64_t", nil
			}
		}

		// if no constraints, default to int64_t
		return "int64_t", nil
	case "object", "array":
		// TODO - I'm not sure if our schema has anything other than simple types
		return "", fmt.Errorf("Object/Array type unimplemented in CppIdentifier.GetType")
	case "boolean":
		return "bool", nil
	case "null":
		// why?
		return "", fmt.Errorf("Null type unimplemented in CppIdentifier.GetType")

	}

	return "", fmt.Errorf("Failed to parse type in CppIdentifier.GetType for property: %v", name)
}
*/
