package doxygenGen

import (
	"fmt"
	"openko-gen/utils"
	"os"
	"path/filepath"
	"strings"
)

const (
	modelPackageOutDir = "model"
)

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
		template.def = validSchemas[i]

		// TODO: full impl

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
