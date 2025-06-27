package cppGenerator

// This was done as a prototype using jsonSchema; it would need to be refactored using
// jsonSchema/tableDef.go, if a use case for it arises.

// generateCpp generates c++ code files for each schema in utils.SchemaDir, and writes the result to the output dir (utils.OutputDir)
/*func GenerateCpp() (err error) {
	// Read and compile all schema files in jsonSchema
	validSchemas, err := utils.LoadSchemas()
	if err != nil {
		fmt.Println(err)
		return
	}

	// the identifier is used to correctly assign types to each property based on constraints
	// and other igenerator
	identifier := igenerator.CppIdentifier{}

	for i := range validSchemas {
		fmt.Print(fmt.Sprintf("generating c++ for: %s", validSchemas[i].Title))

		// the template is an interface implementation that allows us to
		// structure and generate a code file
		template := templates.CppTemplate{}
		tableName, err := utils.GetTableNameFromId(validSchemas[i].Location)
		template.SetClassName(tableName)
		template.SetClassDesc(validSchemas[i].Description)

		// always include string since we're always generating a GetTableName func
		template.AddInclude("string")

		// Generate a static GetTableName() func
		tblNameDef := templates.MethodDefinition{
			ReturnType:  "static std::string",
			Name:        "GetTableName",
			Body:        fmt.Sprintf("return \"%s\";", strings.ToUpper(tableName)),
			Description: "Returns the database table name",
		}
		template.AddMethod(tblNameDef)

		// Setup igenerator and property-based funcs
		for key, _ := range validSchemas[i].Properties {
			// translate the jsonSchema type to a C++ type
			// i.e., a number type with a min:max of 0:255 will return uint8_t
			isOptional := !slices.Contains(validSchemas[i].Required, key)
			_type, tErr := identifier.GetType(key, *validSchemas[i].Properties[key], isOptional)
			if tErr != nil {
				err = fmt.Errorf("%s:%s failed to get type for property: %w", tableName, key, tErr)
				return err
			}

			if strings.HasSuffix(_type, "_t") {
				template.AddInclude("cstdint")
			}

			if strings.HasPrefix(_type, "Nullable") {
				template.AddInclude("cgHelpers/nullable.h")
				template.AddUsing("cgHelpers::Nullable")
			}

			// Setup property definition and add to template
			propDef := templates.PropertyDefinition{
				Type:        _type,
				Name:        key,
				Description: validSchemas[i].Properties[key].Description,
			}
			template.AddProperty(propDef)

			// Setup static column-name funcs
			colNameDef := templates.MethodDefinition{
				ReturnType:  "static std::string",
				Name:        "CN" + "_" + key,
				Body:        fmt.Sprintf("return \"%s\";", key),
				Description: "Returns the database column name",
			}
			template.AddMethod(colNameDef)
		}

		// generate template
		templateStr, tErr := template.Generate()
		if tErr != nil {
			err = fmt.Errorf("%s failed to generate c++ source: %w", tableName, tErr)
			return err
		}

		// write the template to a file
		outFile := filepath.Join(utils.OutputDir, template.GetFileName())
		if fErr := utils.WriteToFile(outFile, templateStr); fErr != nil {
			err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
			return err
		}
		fmt.Println(fmt.Sprintf("... written to: %s", outFile))
	}

	fmt.Println("c++ code generated successfully")
	return nil
}*/
