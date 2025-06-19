package main

import (
	"flag"
	"fmt"
	"ko-codegen/goGenerator"
	"ko-codegen/utils"
)

const (
	goLang = "Go"
	// a C++ generator was prototyped but is not officially supported at this time; no valid usecase has been identified yet
	// would need updating from prototype as the jsonSchema approach has been updated to simplify type identification
	cppLang = "C++"
)

var (
	supportedLangs = []string{goLang /*, cppLang*/}
)

type Args struct {
	Clean      bool
	SchemaPath string
	OutputPath string
	Lang       string
	Usage      bool
}

func getArgs() (a Args) {
	clean := flag.Bool("clean", false, "Cleans the output directory")
	schemaPath := flag.String("schemaPath", utils.DefaultSchemaDir, "Path to the directory containing the schema files")
	outputPath := flag.String("outputPath", utils.DefaultOutputDir, "Path to the directory where the generated code will be written")
	lang := flag.String("lang", "Go", fmt.Sprintf("Language to generate code for.  Valid options are: %v", supportedLangs))
	usage := flag.Bool("usage", false, "Prints program usage information - will ignore all other arguments")

	flag.Parse()

	if clean != nil {
		a.Clean = *clean
	}

	if schemaPath != nil {
		utils.SchemaDir = *schemaPath
		a.SchemaPath = *schemaPath
	}

	if outputPath != nil {
		utils.OutputDir = *outputPath
		a.OutputPath = *outputPath
	}

	if lang != nil {
		a.Lang = *lang
	}

	if usage != nil {
		a.Usage = *usage
	}

	return a
}

func main() {
	fmt.Println("|-----------------------|")
	fmt.Println("| OpenKO Code Generator |")
	fmt.Println("|-----------------------|")

	args := getArgs()
	// if -usage was specified, print the args doc and exit
	if args.Usage {
		flag.Usage()
		return
	}

	var genErr error
	switch args.Lang {
	case goLang:
		// generate Go source for all the schemas
		genErr = goGenerator.GenerateGo(args.Clean)
	/*case cppLang:
	// generate c++ source for all the schemas
	genErr = cppGenerator.GenerateCpp()*/
	default:
		fmt.Printf("Unsupported language: %s\n", args.Lang)
		return
	}

	if genErr != nil {
		fmt.Println(genErr)
		return
	}

	fmt.Println("OpenKO Code Generator completed successfully")
}
