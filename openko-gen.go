package main

import (
	"flag"
	"fmt"
	"openko-gen/arg"
	"openko-gen/cppGen"
	"openko-gen/gormGen"
	"strings"
)

const (
	appTitle    = "OpenKO Code Generator"
	outputWidth = 120
)

func printHeaderRow() {
	fmt.Println(fmt.Sprintf("%s", strings.Repeat("-", outputWidth)))
}

func main() {
	printHeaderRow()
	titlePad := (outputWidth - len(appTitle)) / 2
	fmt.Println(fmt.Sprintf("%[2]s%[1]s%[2]s", appTitle, strings.Repeat(" ", titlePad)))
	printHeaderRow()

	args := arg.GetArgs()
	// if -usage was specified, print the args doc and exit
	if args.Usage {
		flag.Usage()
		return
	}

	if args.List {
		// print language information table
		printLanguageList()
		return
	}

	var genErr error
	switch args.Lang {
	case arg.GormLibrary:
		// generate Go source for all the schemas
		genErr = gormGen.GenerateGo(args.Clean)
	case arg.CppLibrary:
		// generate doxygen-complaint c++ for all schemas
		genErr = cppGen.Generate(args.Clean)
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

func printLanguageList() {
	fmt.Print("\nSupported Language Information\n\n")
	for langName := range arg.LangInfoMap {
		fmt.Printf(" %s\n", langName)
		fmt.Printf("\tDescription: %s\n", arg.LangInfoMap[langName].Description)
		fmt.Printf("\tDefault Output: %s\n", arg.LangInfoMap[langName].DefaultOut)
		fmt.Printf("\tArtifact Produced: %s\n", arg.LangInfoMap[langName].ArtifactProduced)
	}
}
