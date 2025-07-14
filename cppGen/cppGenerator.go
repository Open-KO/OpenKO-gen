package cppGen

import (
	"fmt"
	"github.com/Open-KO/kodb-godef"
	"github.com/Open-KO/kodb-godef/enums/profile"
	"openko-gen/igenerator"
	"openko-gen/utils"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

const (
	modelPackageOutDir  = "model"
	binderPackageOutDir = "binder"
	procPackageOutDir   = "Procedures"
	procCgHelperDir     = "cppGen/cgHelpers/Procedures"
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
// and writes the result to the output dir (default: ./OpenKO-db-modules/)
// It then generates additional profiles from profile.Profiles[]
func Generate(clean bool) (err error) {
	// Read and bind all *.json files in jsonSchema
	validSchemas, err := utils.LoadSchemas()
	if err != nil {
		return err
	}
	validProcs, err := utils.LoadProcs()
	if err != nil {
		return err
	}

	for i := range moduleDefs {
		err = generateModule(clean, validSchemas, moduleDefs[i])
		if err != nil {
			return err
		}
	}

	err = generateProcModule(clean, validProcs)
	if err != nil {
		return err
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

	// identifier is used to assign the correct c++ type from the columns' tsql.TsqlType
	identifier := CppIdentifier{}

	moduleSuffixModel := fmt.Sprintf(moduleSuffixModelFmt, moduleDef.OutDir)
	moduleSuffixBinder := fmt.Sprintf(moduleSuffixBinderFmt, moduleDef.OutDir)

	partitionModules := []string{}
	for i := range validSchemas {
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
		fmt.Println(fmt.Sprintf("generating c++ for: %s", validSchemas[i].Name))

		pk := jsonSchema.IndexDef{}
		for x := range validSchemas[i].Indexes {
			if validSchemas[i].Indexes[x].IsPrimaryKey {
				pk = validSchemas[i].Indexes[x]
				break
			}
		}

		filterDef := validSchemas[i]
		// trim down to the export columns for profile-specific exports
		if moduleDef.namespace != profile.All && len(exportDef.Columns) > 0 {
			filterDef.Columns = []jsonSchema.Column{}

			for x := range validSchemas[i].Columns {
				// include a column if it's part of the export or part of the PK
				if slices.Contains(exportDef.Columns, validSchemas[i].Columns[x].Name) ||
					slices.Contains(pk.Columns, validSchemas[i].Columns[x].Name) {
					filterDef.Columns = append(filterDef.Columns, validSchemas[i].Columns[x])
				}
			}
		}

		partitionModules = append(partitionModules, filterDef.ClassName)

		// the template is an interface implementation that allows us to
		// structure and generate a code file
		template := CppTemplate{}
		template.def = filterDef
		template.namespace = fmt.Sprintf(profile.ModelNsFmt, moduleDef.namespace)
		template.moduleSuffix = moduleSuffixModel
		template.AddInclude("<unordered_set>")
		template.AddInclude("<string>")
		template.moduleDef = moduleDef

		bindingTemplate := CppTemplate{}
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
			Body:        fmt.Sprintf(funcColumnNamesFmt, strings.Join(colNames, ", "), "columnNames"),
			Description: "Returns a set of column names for the table",
		}
		template.AddMethod(colNameDef)

		// Generate a BlobColumns() func
		blobCols := []string{}
		for j := range filterDef.Columns {
			if filterDef.Columns[j].IsBlobType() {
				blobCols = append(blobCols, fmt.Sprintf(`"%s"`, filterDef.Columns[j].Name))
			}
		}
		blobColNameDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  "const std::unordered_set<std::string>&",
			Name:        "BlobColumns",
			Body:        fmt.Sprintf(funcColumnNamesFmt, strings.Join(blobCols, ", "), "blobColumns"),
			Description: "Returns a set of blob column names for the table",
		}
		template.AddMethod(blobColNameDef)

		// Generate a DbType func
		dbTypeDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  "const modelUtil::DbType",
			Name:        "DbType",
			Body:        fmt.Sprintf(funcDbTypeFmt, filterDef.Database),
			Description: "Returns the associated database type for the table",
		}
		template.AddMethod(dbTypeDef)

		if len(pk.Columns) > 0 {
			// Generate a PrimaryKeyColumns func
			pkNames := strings.Builder{}
			pkPropDefs := []jsonSchema.Column{}
			for j := range pk.Columns {
				if j > 0 {
					pkNames.WriteString(", ")
				}
				pkNames.WriteString(fmt.Sprintf(`"%s"`, pk.Columns[j]))

				// add associated property name to list
				for x := range filterDef.Columns {
					if filterDef.Columns[x].Name == pk.Columns[j] {
						pkPropDefs = append(pkPropDefs, filterDef.Columns[x])
					}
				}
			}
			pkColumnsDef := igenerator.MethodDef{
				IsStatic:    true,
				ReturnType:  "const std::vector<std::string>&",
				Name:        "PrimaryKey",
				Body:        fmt.Sprintf(funcPrimaryKeyFmt, pkNames.String()),
				Description: "Returns the columns associated with the table's Primary Key",
			}
			template.AddMethod(pkColumnsDef)

			// Generate a MapKey() func
			retType := ""
			body := ""
			if len(pkPropDefs) == 1 {
				cppType, err := identifier.GetType(pkPropDefs[0])
				if err != nil {
					return err
				}

				// A PK should never be optional, but for the sake of being safe
				retType = fmt.Sprintf(constRefFmt, stripOptional(cppType))
				body = fmt.Sprintf(funcMapKeySingleFmt, pkPropDefs[0].PropertyName)
			} else if len(pkPropDefs) > 1 {
				template.AddInclude("<tuple>")
				retType = "std::tuple<"
				vals := ""
				for j := range pkPropDefs {
					if j > 0 {
						retType += ", "
						vals += ", "
					}
					cppType, err := identifier.GetType(pkPropDefs[j])
					if err != nil {
						return err
					}
					_type := fmt.Sprintf(constRefFmt, stripOptional(cppType))
					retType += _type
					vals += pkPropDefs[j].PropertyName
				}
				retType += ">"
				body = fmt.Sprintf(funcMapKeyMultiFmt, retType, vals)
			}
			mapKeyDef := igenerator.MethodDef{
				IsPure:      true,
				ReturnType:  retType,
				Name:        "MapKey",
				Body:        body,
				Description: "Returns a value for use in map keys based on the table's primary key",
			}
			template.AddMethod(mapKeyDef)
		}

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
		fmt.Println(fmt.Sprintf("... written to: %s", outFile))

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

func generateProcModule(clean bool, validProcs []jsonSchema.ProcDef) (err error) {
	procOut := filepath.Join(utils.OutputDir, procPackageOutDir)
	if clean {
		// clean needs to be specific to the directories it writes to
		err = os.RemoveAll(procOut)
		if err != nil {
			fmt.Printf("failed to clean the procedure output directory: %v\n", err)
			return
		}
	}

	// create the output directories if they don't exist
	err = utils.SetupOutDir(procOut)
	if err != nil {
		return err
	}

	partitionModules := []string{}
	// Read any manually coded files:
	manFiles, err := filepath.Glob(filepath.Join(procCgHelperDir, "*.ixx"))
	if err != nil {
		return err
	}
	for i := range manFiles {
		// add the file to the partition list
		partName := strings.TrimSuffix(filepath.Base(manFiles[i]), filepath.Ext(manFiles[i]))
		partName = strings.Replace(partName, "Procedures-", "", 1)
		partitionModules = append(partitionModules, partName)

		// read the file from cgHelpers
		bytes, err := os.ReadFile(manFiles[i])
		if err != nil {
			err = fmt.Errorf("failed to read cgHelper file: %w", err)
			return err
		}

		// write the file
		outFile := filepath.Join(procOut, filepath.Base(manFiles[i]))
		if fErr := utils.WriteToFile(outFile, string(bytes)); fErr != nil {
			err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
			return err
		}
	}

	for i := range validProcs {
		fmt.Println(fmt.Sprintf("generating c++ for: %s", validProcs[i].Name))
		partitionModules = append(partitionModules, validProcs[i].ClassName)
		template := CppTemplate{}
		funcParamList := [][]string{}
		paramBindings := strings.Builder{}
		for _, param := range validProcs[i].Params {
			cppType, ok := TSqlTypeMapping[param.Type]
			if !ok {
				return fmt.Errorf("unimplemented T-SQL type: %s", param.Type)
			}

			if strings.Contains(cppType, "int") {
				template.AddInclude("<cstdint>")
			} else if strings.Contains(cppType, "time_t") {
				template.AddInclude("<ctime>")
			}
			if strings.Contains(cppType, "std::string") {
				template.AddInclude("<string>")
			}

			_type := ""
			if param.IsOutput {
				_type = fmt.Sprintf("%s&", cppType)
			} else {
				_type = fmt.Sprintf(constRefFmt, cppType)
			}
			funcParamList = append(funcParamList, []string{_type, param.ParamName})

			bindFmt := procBindFmt
			if param.IsOutput {
				bindFmt = procBindRetFmt
			}
			// TODO: likely need to add modifiers like .c_str()
			binding := fmt.Sprintf(bindFmt, param.ParamIndex-1, param.ParamName)
			paramBindings.WriteString(binding)
		}

		executeDef := igenerator.MethodDef{
			ReturnType:  "nanodbc::result*",
			Name:        "execute",
			Params:      funcParamList,
			Body:        fmt.Sprintf(procExecuteFmt, paramBindings.String()),
			Description: "Executes the stored procedure",
		}
		template.AddMethod(executeDef)

		var includes []string
		for include := range template.includes {
			includes = append(includes, include)
		}
		// includes is built from an unordered hash map
		// sort the includes alphabetically so they don't cause diffs gen-to-gen
		sort.Strings(includes)

		pList := strings.Repeat("?,", len(validProcs[i].Params))
		// drop off trailing ','
		if len(pList) > 1 {
			pList = pList[:len(pList)-1]
		}
		procCallStr := fmt.Sprintf(procCallFmt, validProcs[i].Name, pList)

		methods := strings.Join(template.methods, "")

		// file contents:
		// 1. includes
		// 2. Class Name
		// 3. Procedure Call prepared statement i.e., {? = CALL LOAD_ACCOUNT_CHARID(?)}
		// 4. Methods
		// 5. Proc description
		partFileStr := fmt.Sprintf(procPartitionFmt, strings.Join(includes, ""),
			validProcs[i].ClassName, procCallStr, methods, validProcs[i].Description)

		// write file
		partFileName := fmt.Sprintf("%s-%s.ixx", procPackageOutDir, validProcs[i].ClassName)
		partOut := filepath.Join(procOut, partFileName)
		if fErr := utils.WriteToFile(partOut, partFileStr); fErr != nil {
			err = fmt.Errorf("failed to write file %s: %w", partOut, fErr)
			return err
		}
	}

	// create primary module file
	exportImports := strings.Builder{}
	for i := range partitionModules {
		exportImports.WriteString(fmt.Sprintf(exportImportFmt, partitionModules[i]))
	}
	procModuleStr := fmt.Sprintf(primaryModuleFmt, exportImports.String(), procPackageOutDir)
	outFile := filepath.Join(procOut, fmt.Sprintf(primaryModuleFileName, procPackageOutDir))
	if fErr := utils.WriteToFile(outFile, procModuleStr); fErr != nil {
		err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	return nil
}
