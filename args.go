package main

import (
	"flag"
	"fmt"
	"openko-gen/utils"
)

const (
	gormLibrary = "gorm"

	// a C++ generator was prototyped but is not officially supported at this time; no valid usecase has been identified yet
	// would need updating from prototype as the jsonSchema approach has been updated to simplify type identification
	cppLang = "c++"
)

var (
	supportedLangs = []string{gormLibrary /*, cppLang*/}
	langInfo       map[string]LangInfo
)

func init() {
	langInfo = make(map[string]LangInfo)
	langInfo[gormLibrary] = LangInfo{
		Name:             gormLibrary,
		Description:      "Go Object Relational Mapping (gorm) model library; built for use in the kodb-util project",
		DefaultOut:       utils.GormLibOut,
		ArtifactProduced: "openko-gorm",
	}
}

type LangInfo struct {
	Name             string
	Description      string
	DefaultOut       string
	ArtifactProduced string
}

type Args struct {
	Clean      bool
	SchemaPath string
	OutputPath string
	Lang       string
	Usage      bool
	List       bool
}

func getArgs() (a Args) {
	clean := flag.Bool("clean", false, "Cleans the output directory")
	schemaPath := flag.String("openkodb", utils.DefaultSchemaDir, "Path to the openko-db project directory")
	outputPath := flag.String("o", utils.DefaultOutputDir, "Path to the directory where the generated code will be written. If unspecified uses the language default (see -list)")
	lang := flag.String("l", gormLibrary, fmt.Sprintf("Language/library to generate code for.  Valid options are: %v", supportedLangs))
	list := flag.Bool("list", false, "Lists supported language/library information")
	usage := flag.Bool("usage", false, "Prints program usage information - will ignore all other arguments")

	flag.Parse()

	if clean != nil {
		a.Clean = *clean
	}

	if schemaPath != nil {
		utils.SchemaDir = *schemaPath
		a.SchemaPath = *schemaPath
	}

	if lang != nil {
		a.Lang = *lang
	}

	if outputPath != nil {
		a.OutputPath = *outputPath
		if a.OutputPath == utils.DefaultOutputDir {
			langOut, ok := langInfo[a.Lang]
			if ok {
				a.OutputPath = langOut.DefaultOut
			}
		}
		utils.OutputDir = a.OutputPath
	}

	if usage != nil {
		a.Usage = *usage
	}

	if list != nil {
		a.List = *list
	}

	return a
}
