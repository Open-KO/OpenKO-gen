package cppGen

import (
	"fmt"
	"github.com/Open-KO/kodb-godef"
	"github.com/Open-KO/kodb-godef/enums/profile"
	"github.com/Open-KO/kodb-godef/enums/tsql"
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

	modelModuleName := fmt.Sprintf(moduleSuffixModelFmt, moduleDef.OutDir)
	binderModuleName := fmt.Sprintf(moduleSuffixBinderFmt, moduleDef.OutDir)

	modelTemplate := CppTemplate{}
	modelTemplate.AddInclude("<unordered_set>")
	modelTemplate.AddInclude("<string>")
	modelTemplate.AddInclude("<string>")
	modelTemplate.AddInclude("ModelUtil")
	binderTemplate := CppTemplate{}
	binderTemplate.AddInclude("<string>")
	binderTemplate.AddInclude("<unordered_map>")
	binderTemplate.AddInclude("<nanodbc/nanodbc.h>")
	binderTemplate.AddInclude(modelModuleName)
	binderTemplate.AddInclude("BinderUtil")

	// identifier is used to assign the correct c++ type from the columns' tsql.TsqlType
	identifier := CppIdentifier{}

	// memory is tasty
	modelFileContents := strings.Builder{}
	modelFwdDeclares := strings.Builder{}
	binderFileContents := strings.Builder{}
	modelNs := fmt.Sprintf(profile.ModelNsFmt, moduleDef.namespace)
	binderNs := fmt.Sprintf(profile.BinderNsFmt, moduleDef.namespace)
	modelFileContents.WriteString(fmt.Sprintf(namespaceOpen, modelNs))
	binderFileContents.WriteString(fmt.Sprintf(namespaceOpen, binderNs))
	modelFwdDeclares.WriteString(fmt.Sprintf(namespaceOpen, binderNs))
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

		// model forward declare of binder
		modelFwdDeclares.WriteString(fmt.Sprintf(modelFwdDeclareFmt, validSchemas[i].ClassName))

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

		// the template is an interface implementation that allows us to
		// structure and generate a code file
		template := CppTemplate{}
		template.def = filterDef
		template.namespace = fmt.Sprintf(profile.ModelNsFmt, moduleDef.namespace)
		template.moduleSuffix = modelModuleName
		template.moduleDef = moduleDef

		bindingTemplate := CppTemplate{}
		bindingTemplate.def = filterDef
		bindingTemplate.namespace = fmt.Sprintf(profile.BinderNsFmt, moduleDef.namespace)
		bindingTemplate.moduleSuffix = binderModuleName
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
		modelClassDef, tErr := template.Generate()
		if tErr != nil {
			err = fmt.Errorf("%s failed to generate c++ source: %w", filterDef.Name, tErr)
			return err
		}
		modelFileContents.WriteString(modelClassDef)

		// copy includes to main template
		for k, v := range template.includes {
			modelTemplate.includes[k] = v
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
		binderClassDef, tErr := bindingTemplate.GenerateBinderClass()
		if tErr != nil {
			err = fmt.Errorf("%s failed to generate c++ bindings source: %w", filterDef.Name, tErr)
			return err
		}
		binderFileContents.WriteString(binderClassDef)
	}

	var includes []string
	var imports []string
	for include, isInclude := range modelTemplate.includes {
		if isInclude {
			includes = append(includes, include)
		} else {
			imports = append(imports, include)
		}
	}
	// includes is built from an unordered hash map
	// sort the includes alphabetically so they don't cause diffs gen-to-gen
	sort.Strings(includes)
	sort.Strings(imports)

	// close the namespace
	modelFwdDeclares.WriteString("\n}\n\n")
	modelFileContents.WriteString("}")

	// combine model content
	modelFwdDeclares.WriteString(modelFileContents.String())

	// 1: Module name
	// 2. includes
	// 3. imports
	// 4: file contents
	modelModuleStr := fmt.Sprintf(primaryModuleFmt, modelModuleName, strings.Join(includes, ""), strings.Join(imports, ""), modelFwdDeclares.String())
	outFile := filepath.Join(modelOut, fmt.Sprintf(primaryModuleFileName, modelModuleName))
	if fErr := utils.WriteToFile(outFile, modelModuleStr); fErr != nil {
		err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	includes = []string{}
	imports = []string{}
	for include, isInclude := range binderTemplate.includes {
		if isInclude {
			includes = append(includes, include)
		} else {
			imports = append(imports, include)
		}
	}
	// includes is built from an unordered hash map
	// sort the includes alphabetically so they don't cause diffs gen-to-gen
	sort.Strings(includes)
	sort.Strings(imports)

	// close the namespace
	binderFileContents.WriteString("}")

	bindingModuleStr := fmt.Sprintf(primaryModuleFmt, binderModuleName, strings.Join(includes, ""), strings.Join(imports, ""), binderFileContents.String())
	outFile = filepath.Join(binderOut, fmt.Sprintf(primaryModuleFileName, binderModuleName))
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

	template := CppTemplate{}
	template.AddInclude("<nanodbc/nanodbc.h>")
	template.AddInclude("<memory>")

	procFileContents := strings.Builder{}
	procFileContents.WriteString(fmt.Sprintf(namespaceOpen, "procedures"))
	procFileContents.WriteString("\n")
	// Read any manually coded files:
	manFiles, err := filepath.Glob(filepath.Join(procCgHelperDir, "*.ixx"))
	if err != nil {
		return err
	}
	for i := range manFiles {
		// read the file from cgHelpers
		bytes, err := os.ReadFile(manFiles[i])
		if err != nil {
			err = fmt.Errorf("failed to read cgHelper file: %w", err)
			return err
		}
		procFileContents.Write(bytes)
	}

	for i := range validProcs {
		fmt.Println(fmt.Sprintf("generating c++ for: %s", validProcs[i].Name))

		classTemplate := CppTemplate{}
		paramBindings := strings.Builder{}
		funcParamList := [][]string{}
		pList := strings.Repeat("?,", len(validProcs[i].Params))
		// drop off trailing ','
		if len(pList) > 1 {
			pList = pList[:len(pList)-1]
		}

		hasOutputOrReturn := false
		procCallStr := ""
		// posMod is used to shift the binding position depending on if we bind a return
		posMod := 0
		if validProcs[i].HasReturn != nil && *validProcs[i].HasReturn {
			hasOutputOrReturn = true
			procCallStr = fmt.Sprintf(procCallWithRetFmt, validProcs[i].Name, pList)
			paramBindings.WriteString(fmt.Sprintf(procBindRetFmt, 0, "returnValue"))
			funcParamList = append(funcParamList, []string{"int*", "returnValue"})
		} else {
			posMod = -1
			procCallStr = fmt.Sprintf(procCallFmt, validProcs[i].Name, pList)
		}

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

			_type := ""
			if cppType == "std::string" || param.Type == tsql.Text {
				cppType = "char*"
			} else if strings.Contains(cppType, "vector") {
				// TODO: Figure out best binding type
				cppType = "char*"
			} else if cppType == "uint8_t" {
				// upcast since nanodbc can't handle tinyint right
				cppType = "int16_t"
			}
			if param.IsOutput {
				_type = cppType
				hasOutputOrReturn = true

				if !strings.HasSuffix(_type, "*") {
					_type = fmt.Sprintf(ptrFmt, _type)
				}
			} else {
				_type = fmt.Sprintf(constFmt, cppType)
			}
			isPtr := strings.HasSuffix(_type, "*")
			funcParamList = append(funcParamList, []string{_type, param.ParamName})

			bindFmt := procBindFmt
			if param.IsOutput {
				bindFmt = procBindRetFmt
			}
			// TODO: likely need to add modifiers like .c_str()
			p := param.ParamName
			if !isPtr {
				p = "&" + p
			}
			binding := fmt.Sprintf(bindFmt, param.ParamIndex+posMod, p)
			paramBindings.WriteString(binding)
		}

		executeBody := procExecuteNoParam
		if paramBindings.Len() > 0 {
			executeBody = fmt.Sprintf(procExecuteFmt, paramBindings.String())
		}

		classTemplate.AddInclude("<memory>")
		executeDef := igenerator.MethodDef{
			ReturnType:  "std::weak_ptr<nanodbc::result>",
			Name:        "execute",
			Params:      funcParamList,
			Body:        executeBody,
			Description: "Executes the stored procedure",
		}
		classTemplate.AddMethod(executeDef)

		if hasOutputOrReturn {
			destructorDef := igenerator.MethodDef{
				ReturnType:  "",
				Name:        fmt.Sprintf("~%s", validProcs[i].ClassName),
				Body:        procDestructorWithFlushDef,
				Description: "Flushes any output variables or return values on destruction",
			}
			classTemplate.AddMethod(destructorDef)
		}

		methods := strings.Join(classTemplate.methods, "\n")

		// file contents:
		// 1. Class Name
		// 2. Procedure Call prepared statement i.e., {? = CALL LOAD_ACCOUNT_CHARID(?)}
		// 3. Methods
		// 4. Proc description
		partFileStr := fmt.Sprintf(procClassFmt, validProcs[i].ClassName,
			procCallStr, methods, validProcs[i].Description)
		procFileContents.WriteString(partFileStr)
	}

	var includes []string
	var imports []string
	for include, isInclude := range template.includes {
		if isInclude {
			includes = append(includes, include)
		} else {
			imports = append(imports, include)
		}
	}
	// includes is built from an unordered hash map
	// sort the includes alphabetically so they don't cause diffs gen-to-gen
	sort.Strings(includes)
	sort.Strings(imports)

	// close the namespace
	procFileContents.WriteString("}")

	// 1: Module name
	// 2. includes
	// 3. imports
	// 4: file contents
	procModuleStr := fmt.Sprintf(primaryModuleFmt, procPackageOutDir, strings.Join(includes, ""), strings.Join(imports, ""), procFileContents.String())
	outFile := filepath.Join(procOut, fmt.Sprintf(primaryModuleFileName, procPackageOutDir))
	if fErr := utils.WriteToFile(outFile, procModuleStr); fErr != nil {
		err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	return nil
}
