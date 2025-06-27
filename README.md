# OpenKO Codegen

## Project Overview
The goal of this project is to:
* Create a full json-schema representation of the OpenKO dbo schema (see `OpenKO-db/jsonSchema` TODO Repo URL)
* Generate a Go Object Relational Mapping (GORM) Model library from the jsonSchema with additional helper functions
* Implement everything via interfaces, such that the interfaces can be implemented for other languages/use-cases as desired

## Dependencies
The following commands assume that you have a terminal open in the root folder of the project.

The `OpenKO-db` and `openko-gorm` projects are submodules:
* `OpenKO-db`: Schema definitions in OpenKO-db/jsonSchema are used to generate code.
* `openko-gorm`: The `gorm` language option generates the gorm model library.  This library is imported to `kodb-util` to perform import/export tasks


To fetch submodule updates:
```shell
git submodule update --recursive --remote
```

This utility is programmed with Go 1.24+.  You'll need to install the language if you want to build locally. See https://go.dev/doc/install

If Go is correctly installed on your path, you should be able to run `go version` in your terminal and get version
information output:
```
PS C:\> go version
go version go1.24.1 windows/amd64
```

To download Go dependencies, run:
```shell
go mod download
```

To run the application, run:
```shell
go run openko-gen.go
```

## CLI Arguments

CLI Usage (-usage arg):
```
------------------------------------------------------------------------------------------------------------------------
                                                 OpenKO Code Generator
------------------------------------------------------------------------------------------------------------------------
Usage of openko-gen.exe:
  -clean
    	Cleans the output directory
  -l string
    	Language/library to generate code for.  Valid options are: [gorm] (default "gorm")
  -list
    	Lists supported language/library information
  -o string
    	Path to the directory where the generated code will be written. If unspecified uses the language default (see -list) (default "out")
  -openkodb string
    	Path to the openko-db project directory (default "./OpenKO-db/jsonSchema")
  -usage
    	Prints program usage information - will ignore all other arguments
```

## Output
-l gorm
Description: Go Object Relational Mapping (gorm) model library; built for use in the kodb-util project
Default Output: openko-gorm/
Artifact Produced: openko-gorm


## Building the utility program
To build `openko-gen.exe`, run the following command in this directory:
```shell
go build
```
