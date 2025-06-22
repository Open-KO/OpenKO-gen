# OpenKO Codegen

## Project Overview
The goal of this project is to:
* Create a full json-schema representation of the OpenKO dbo schema (see `OpenKO-db/jsonSchema` TODO Repo URL)
* Generate a Go Object Relational Mapping (GORM) Model library from the jsonSchema with additional helper functions
* Implement everything via interfaces, such that the interfaces can be implemented for other languages/use-cases as desired

The project is currently in prototyping and subject to refactoring.

## Build/Run Project

The project is coded in Go 1.24; You'll need to install the language and add it to your PATH. See https://go.dev/doc/install

If Go is correctly installed on your path, you should be able to run `go version` in your terminal and get version
information output:
```
PS C:\> go version
go version go1.24.1 windows/amd64
```
The following commands assume that you have a terminal open in the root folder of the project.

To download project dependencies, run:
```shell
go mod download
```

To build/run the application, run:
```shell
go run main.go
```

## CLI Arguments

CLI Usage (-usage arg):
```
Usage of openko-gen.exe:
  -clean
    	Cleans the output directory
  -lang string
    	Language to generate code for.  Valid options are: [Go] (default "Go")
  -outputPath string
    	Path to the directory where the generated code will be written (default "out")
  -schemaPath string
    	Path to the directory containing the schema files (default "jsonSchema")
  -usage
    	Prints program usage information - will ignore all other arguments
```

## Output
Currently, the only interface implemented is for the GORM library (openko-gorm). 

## TODOs
Things that would be good to idea in the near term:
1. implement extending OdbcRecordSet and associated functions.  Maybe.  Current OdbcRecordSet behavior seems to be for specific queries rather than generalized.  Would require thought-out proposal and discussion with larger group

## Out of Scope
1. C++ file formatter: Looked into this a bit. There doesn't appear to be any stand-alone utilities to run .editorconfig against a file, and .editorconfig properties are largely IDE-dependent.
