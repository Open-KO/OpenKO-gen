package cppGen

import (
	"fmt"
	"github.com/Open-KO/OpenKO-db/jsonSchema"
	"github.com/Open-KO/OpenKO-db/jsonSchema/enums/profile"
	"openko-gen/igenerator"
	"openko-gen/utils"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	modelPackageOutDir  = "model"
	binderPackageOutDir = "binder"
)

type ModuleDef struct {
	OutDir    string
	namespace profile.ExportName
}

var (
	moduleDefs = []ModuleDef{
		{
			OutDir:    "Full",
			namespace: profile.All,
		},
		{
			OutDir:    "VersionManager",
			namespace: profile.VersionManager,
		},
		{
			OutDir:    "Ebenezer",
			namespace: profile.Ebenezer,
		},
		{
			OutDir:    "AIServer",
			namespace: profile.AIServer,
		},
		{
			OutDir:    "Aujard",
			namespace: profile.Aujard,
		},
	}
)

// Generate generates c++ code files for each schema in OpenKO-db/jsonSchema,
// and writes the result to the output dir (default: ./doxygen-db/)
// It then generates additional profiles from profile.Profiles[]
func Generate(clean bool) (err error) {
	// Read and bind all *.json files in jsonSchema
	validSchemas, err := utils.LoadSchemas()
	if err != nil {
		return err
	}

	for i := range moduleDefs {
		err = generateModule(clean, validSchemas, moduleDefs[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func generateModule(clean bool, validSchemas []jsonSchema.TableDef, moduleDef ModuleDef) (err error) {
	modelOut := filepath.Join(utils.OutputDir, moduleDef.OutDir, modelPackageOutDir)
	binderOut := filepath.Join(utils.OutputDir, moduleDef.OutDir, binderPackageOutDir)
	if clean {
		// clean needs to be specific to the directories it writes to
		err = os.RemoveAll(modelOut)
		if err != nil {
			fmt.Printf("failed to clean the model output directory: %v\n", err)
			return
		}

		err = os.RemoveAll(binderOut)
		if err != nil {
			fmt.Printf("failed to clean the binder output directory: %v\n", err)
			return
		}
	}

	// create the output directories if they don't exist
	err = utils.SetupOutDir(modelOut)
	if err != nil {
		return err
	}
	err = utils.SetupOutDir(binderOut)
	if err != nil {
		return err
	}

	partitionModules := []string{}
	for i := range validSchemas {
		fmt.Print(fmt.Sprintf("generating doxygen c++ for: %s", validSchemas[i].Name))
		isIncluded := moduleDef.namespace == profile.All
		exportDef := jsonSchema.Export{}
		for j := 0; !isIncluded && j < len(validSchemas[i].Exports); j++ {
			if validSchemas[i].Exports[j].Namespace == moduleDef.namespace {
				isIncluded = true
				exportDef = validSchemas[i].Exports[j]
			}
		}
		if !isIncluded {
			// the export isn't defined for this schema, skip processing it
			continue
		}

		filterDef := validSchemas[i]
		// trim down to the export columns for profile-specific exports
		if moduleDef.namespace != profile.All && len(exportDef.Columns) > 0 {
			filterDef.Columns = []jsonSchema.Column{}
			for x := range validSchemas[i].Columns {
				if slices.Contains(exportDef.Columns, validSchemas[i].Columns[x].Name) {
					filterDef.Columns = append(filterDef.Columns, validSchemas[i].Columns[x])
				}
			}
		}

		partitionModules = append(partitionModules, filterDef.ClassName)

		// the template is an interface implementation that allows us to
		// structure and generate a code file
		template := DoxygenTemplate{}
		template.def = filterDef
		template.namespace = string(moduleDef.namespace)
		template.moduleSuffix = moduleSuffixModel
		template.AddInclude("<unordered_set>")
		template.AddInclude("<string>")
		template.moduleDef = moduleDef

		bindingTemplate := DoxygenTemplate{}
		bindingTemplate.def = filterDef
		bindingTemplate.namespace = fmt.Sprintf(profile.BinderNsFmt, moduleDef.namespace)
		bindingTemplate.moduleSuffix = moduleSuffixBinder
		bindingTemplate.AddInclude("<string>")
		bindingTemplate.AddInclude("<unordered_map>")
		bindingTemplate.AddInclude("<nanodbc/nanodbc.h>")
		bindingTemplate.moduleDef = moduleDef

		// function defs
		// Generate a TableName() func
		tblNameDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  "const std::string&",
			Name:        "TableName",
			Body:        fmt.Sprintf(funcTableNameFmt, filterDef.Name),
			Description: "Returns the table name",
		}
		template.AddMethod(tblNameDef)

		// Generate a ColumnNames() func
		colNames := []string{}
		for j := range filterDef.Columns {
			colNames = append(colNames, fmt.Sprintf(`"%s"`, filterDef.Columns[j].Name))
		}
		colNameDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  "const std::unordered_set<std::string>&",
			Name:        "ColumnNames",
			Body:        fmt.Sprintf(funcColumnNamesFmt, strings.Join(colNames, ", ")),
			Description: "Returns a set of column names for the table",
		}
		template.AddMethod(colNameDef)

		// Generate a DbType func
		dbTypeDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  "const std::string&",
			Name:        "DbType",
			Body:        fmt.Sprintf(funcDbTypeFmt, filterDef.Database),
			Description: "Returns the associated database type for the table",
		}
		template.AddMethod(dbTypeDef)

		// generate template
		templateStr, tErr := template.Generate()
		if tErr != nil {
			err = fmt.Errorf("%s failed to generate c++ source: %w", filterDef.Name, tErr)
			return err
		}

		outFile := filepath.Join(modelOut, template.GetFileName())
		if fErr := utils.WriteToFile(outFile, templateStr); fErr != nil {
			err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
			return err
		}

		// binder functions
		// Generate a GetColumnBindings func
		bindings := strings.Builder{}
		for j := range filterDef.Columns {
			if j > 0 {
				bindings.WriteString(",")
			}
			bindings.WriteString(fmt.Sprintf(bindingFmt, filterDef.Columns[j].Name, filterDef.ClassName, filterDef.Columns[j].PropertyName))
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
			err = fmt.Errorf("%s failed to generate c++ bindings source: %w", filterDef.Name, tErr)
			return err
		}

		outFile = filepath.Join(binderOut, bindingTemplate.GetFileName())
		if fErr := utils.WriteToFile(outFile, bindingTemplateStr); fErr != nil {
			err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
			return err
		}

		fmt.Println(fmt.Sprintf("... written to: %s", outFile))
	}

	// create primary module file
	exportImports := strings.Builder{}
	for i := range partitionModules {
		exportImports.WriteString(fmt.Sprintf(exportImportFmt, partitionModules[i]))
	}
	modelModuleStr := fmt.Sprintf(primaryModuleFmt, exportImports.String(), moduleSuffixModel)
	outFile := filepath.Join(modelOut, fmt.Sprintf(primaryModuleFileName, moduleSuffixModel))
	if fErr := utils.WriteToFile(outFile, modelModuleStr); fErr != nil {
		err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}
	bindingModuleStr := fmt.Sprintf(primaryModuleFmt, exportImports.String(), moduleSuffixBinder)
	outFile = filepath.Join(binderOut, fmt.Sprintf(primaryModuleFileName, moduleSuffixBinder))
	if fErr := utils.WriteToFile(outFile, bindingModuleStr); fErr != nil {
		err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	return nil
}
