package utils

import (
	"encoding/json"
	"fmt"
	"github.com/Open-KO/OpenKO-db/jsonSchema"
	"os"
	"path/filepath"
)

const (
	DefaultOutputDir = "out"
	DefaultSchemaDir = "./OpenKO-db/jsonSchema"
	schemaExtPattern = "*.json"
	GormLibOut       = "openko-gorm/"
)

var (
	SchemaDir = DefaultSchemaDir
	OutputDir = DefaultOutputDir
)

// LoadSchemas reads all *.json files from the given schemas directory and marshals them into TableDefs
func LoadSchemas() (validSchemas []jsonSchema.TableDef, err error) {
	fmt.Println("reading schema file names from: " + SchemaDir)
	fileNames, err := GetSchemaFileNames(SchemaDir)
	if err != nil {
		err = fmt.Errorf("failed to read schema file names: %w", err)
		return validSchemas, err
	}
	fmt.Println(fmt.Sprintf("found %d schema files", len(fileNames)))

	for i := range fileNames {
		fmt.Print(fmt.Sprintf("loading schema file: %s", fileNames[i]))
		bytes, err := os.ReadFile(fileNames[i])
		if err != nil {
			err = fmt.Errorf("failed to read schema file: %w", err)
			return validSchemas, err
		}

		def := jsonSchema.TableDef{}
		err = json.Unmarshal(bytes, &def)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal schema file: %w", err)
			return validSchemas, err
		}

		fmt.Println(" ...done")
		validSchemas = append(validSchemas, def)
	}

	return validSchemas, nil
}

func WriteToFile(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

func GetSchemaFileNames(schemaDir string) (fileNames []string, err error) {
	return filepath.Glob(filepath.Join(schemaDir, schemaExtPattern))
}
