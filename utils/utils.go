package utils

import (
	"encoding/json"
	"fmt"
	"github.com/Open-KO/kodb-godef"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultOutputDir = "out"
	DefaultSchemaDir = "./OpenKO-db/jsonSchema"
	schemaExtPattern = "*.json"
	GormLibOut       = "OpenKO-gorm/"
	CppLibOut        = "OpenKO-db-modules/"
	procedureDir     = "procedures"
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

// LoadProcs reads all *.json files from the given schemas directory and marshals them into ProcDefs
func LoadProcs() (validProcs []jsonSchema.ProcDef, err error) {
	procDir := filepath.Join(SchemaDir, procedureDir)
	fmt.Println("reading procedure json file names from: " + procDir)
	fileNames, err := GetSchemaFileNames(procDir)
	if err != nil {
		err = fmt.Errorf("failed to read procedure json file names: %w", err)
		return validProcs, err
	}
	fmt.Println(fmt.Sprintf("found %d procedure json files", len(fileNames)))

	for i := range fileNames {
		fmt.Print(fmt.Sprintf("loading schema file: %s", fileNames[i]))
		bytes, err := os.ReadFile(fileNames[i])
		if err != nil {
			err = fmt.Errorf("failed to read schema file: %w", err)
			return validProcs, err
		}

		def := jsonSchema.ProcDef{}
		err = json.Unmarshal(bytes, &def)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal schema file: %w", err)
			return validProcs, err
		}

		fmt.Println(" ...done")
		validProcs = append(validProcs, def)
	}

	return validProcs, nil
}

func WriteToFile(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

func GetSchemaFileNames(schemaDir string) (fileNames []string, err error) {
	return filepath.Glob(filepath.Join(schemaDir, schemaExtPattern))
}

// SetupOutputDir creates an output directory if it doesn't exist
func SetupOutDir(packageDir string) error {
	return os.MkdirAll(packageDir, os.ModePerm)
}

func GetIndentation(level int) string {
	return strings.Repeat("\t", level)
}

func FormatAndIndentLines(level int, format string, args ...any) string {
	indent := GetIndentation(level)
	formattedLine := fmt.Sprintf(format, args...)
	lines := strings.Split(formattedLine, "\n")
	for i := range lines {
		if len(lines[i]) > 0 {
			lines[i] = indent + lines[i]
		}
	}

	return strings.Join(lines, "\n")
}
