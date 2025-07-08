package doxygenGen

import (
	"fmt"
	"openko-gen/igenerator"
	"openko-gen/utils"
	"os"
	"path/filepath"
	"strings"
)

const (
	modelPackageOutDir = "model"
)

type BindingEntry struct {
	Name    string
	Binding string
}

// Generate generates c++ code files for each schema in OpenKO-db/jsonSchema,
// and writes the result to the output dir (default: ./doxygen-db/)
func Generate(clean bool) (err error) {
	// Read and bind all *.json files in jsonSchema
	validSchemas, err := utils.LoadSchemas()
	if err != nil {
		return err
	}

	if clean {
		// clean needs to be specific to the directories it writes to
		err = os.RemoveAll(filepath.Join(utils.OutputDir, modelPackageOutDir))
		if err != nil {
			fmt.Printf("failed to clean the output directory: %v\n", err)
			return
		}
	}

	// create the output directory if it doesn't exist
	err = utils.SetupOutDir(modelPackageOutDir)
	if err != nil {
		return err
	}

	partitionModules := []string{}
	for i := range validSchemas {
		fmt.Print(fmt.Sprintf("generating doxygen c++ for: %s", validSchemas[i].Name))
		partitionModules = append(partitionModules, validSchemas[i].ClassName)

		// the template is an interface implementation that allows us to
		// structure and generate a code file
		template := DoxygenTemplate{}
		bindingTemplate := DoxygenTemplate{}
		template.def = validSchemas[i]
		bindingTemplate.def = validSchemas[i]

		template.AddInclude("<unordered_set>")
		template.AddInclude("<string>")
		bindingTemplate.AddInclude("<string>")
		bindingTemplate.AddInclude("<unordered_map>")
		bindingTemplate.AddInclude("<nanodbc/nanodbc.h>")

		// function defs
		// Generate a TableName() func
		tblNameDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  "const std::string&",
			Name:        "TableName",
			Body:        fmt.Sprintf(funcTableNameFmt, validSchemas[i].Name),
			Description: "Returns the table name",
		}
		template.AddMethod(tblNameDef)

		// Generate a ColumnNames() func
		colNames := []string{}
		for j := range validSchemas[i].Columns {
			colNames = append(colNames, fmt.Sprintf(`"%s"`, validSchemas[i].Columns[j].Name))
		}
		colNameDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  "std::unordered_set<std::string>&",
			Name:        "ColumnNames",
			Body:        fmt.Sprintf(funcColumnNamesFmt, strings.Join(colNames, ", ")),
			Description: "Returns a set of column names for the table",
		}
		template.AddMethod(colNameDef)

		// Generate a DbType func
		dbTypeDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  "std::string&",
			Name:        "DbType",
			Body:        fmt.Sprintf(funcDbTypeFmt, validSchemas[i].Database),
			Description: "Returns the associated database type for the table",
		}
		template.AddMethod(dbTypeDef)

		// generate template
		templateStr, tErr := template.Generate()
		if tErr != nil {
			err = fmt.Errorf("%s failed to generate c++ source: %w", validSchemas[i].Name, tErr)
			return err
		}

		outFile := filepath.Join(utils.OutputDir, modelPackageOutDir, template.GetFileName())
		if fErr := utils.WriteToFile(outFile, templateStr); fErr != nil {
			err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
			return err
		}

		// binder functions
		// Generate a GetColumnBindings func
		bindings := strings.Builder{}
		for j := range validSchemas[i].Columns {
			if j > 0 {
				bindings.WriteString(",")
			}
			bindings.WriteString(fmt.Sprintf(bindingFmt, validSchemas[i].Columns[j].Name, validSchemas[i].ClassName, validSchemas[i].Columns[j].PropertyName))
		}
		colBindDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  "const BindingsMapType&",
			Name:        "GetColumnBindings",
			Body:        fmt.Sprintf(funcColumnBindingsFmt, bindings.String()),
			Description: "Returns the binding function associated with the column name",
		}
		bindingTemplate.AddMethod(colBindDef)

		// generate binding template
		bindingTemplateStr, tErr := bindingTemplate.GenerateBinders()
		if tErr != nil {
			err = fmt.Errorf("%s failed to generate c++ bindings source: %w", validSchemas[i].Name, tErr)
			return err
		}

		outFile = filepath.Join(utils.OutputDir, modelPackageOutDir, bindingTemplate.GetBindingFileName())
		if fErr := utils.WriteToFile(outFile, bindingTemplateStr); fErr != nil {
			err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
			return err
		}

		fmt.Println(fmt.Sprintf("... written to: %s", outFile))
	}

	// create primary module file
	exportImports := []string{}
	for i := range partitionModules {
		exportImports = append(exportImports, fmt.Sprintf(exportImportFmt, partitionModules[i]))
	}
	primaryModuleStr := fmt.Sprintf(primaryModuleFmt, strings.Join(exportImports, ""))
	outFile := filepath.Join(utils.OutputDir, modelPackageOutDir, primaryModuleFileName)
	if fErr := utils.WriteToFile(outFile, primaryModuleStr); fErr != nil {
		err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	return nil
}
