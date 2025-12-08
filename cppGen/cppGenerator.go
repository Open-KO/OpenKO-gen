package cppGen

import (
	"fmt"
	"openko-gen/igenerator"
	"openko-gen/utils"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/Open-KO/kodb-godef/enums/profile"
	"github.com/Open-KO/kodb-godef/enums/tsql"
	"github.com/Open-KO/kodb-godef/jsonSchema"
)

const (
	modelPackageOutDir    = "model"
	binderPackageOutDir   = "binder"
	procPackageOutDir     = "StoredProc"
	nanodbcParamTypeOut   = "nanodbc::statement::PARAM_OUT"
	nanodbcParamTypeRet   = "nanodbc::statement::PARAM_RETURN"
	nanodbcBindFunc       = "bind"
	nanodbcBindBinaryFunc = "bind_binary"
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
		err = generateTableClasses(clean, validSchemas, moduleDefs[i])
		if err != nil {
			return err
		}
	}

	err = generateProcClasses(clean, validProcs)
	if err != nil {
		return err
	}

	return nil
}

func generateTableClasses(clean bool, validSchemas []jsonSchema.TableDef, moduleDef ModuleDef) (err error) {
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

	modelClassName := fmt.Sprintf(classSuffixModelFmt, moduleDef.OutDir)
	binderClassName := fmt.Sprintf(classSuffixBinderFmt, moduleDef.OutDir)

	modelHeaderTemplate := CppTemplate{}
	modelHeaderTemplate.AddInclude("<unordered_set>")
	modelHeaderTemplate.AddInclude("<string>")
	modelHeaderTemplate.AddInclude("<vector>")
	modelHeaderTemplate.AddInclude("<ModelUtil/ModelUtil.h>")

	binderHeaderTemplate := CppTemplate{}
	binderHeaderTemplate.AddInclude("<string>")
	binderHeaderTemplate.AddInclude("<unordered_map>")
	binderHeaderTemplate.AddInclude("<ModelUtil/ModelUtil.h>")

	// identifier is used to assign the correct c++ type from the columns' tsql.TsqlType
	identifier := CppIdentifier{}

	// memory is tasty
	modelHeaderFileContents := strings.Builder{}
	modelSourceFileContents := strings.Builder{}
	modelHeaderFwdDeclares := strings.Builder{}
	binderHeaderFileContents := strings.Builder{}
	binderSourceFileContents := strings.Builder{}
	binderHeaderFwdDeclares := strings.Builder{}
	modelNs := fmt.Sprintf(profile.ModelNsFmt, moduleDef.namespace)
	binderNs := fmt.Sprintf(profile.BinderNsFmt, moduleDef.namespace)
	modelHeaderFileContents.WriteString(fmt.Sprintf(namespaceOpen, modelNs))
	modelSourceFileContents.WriteString(fmt.Sprintf(namespaceOpen, modelNs))
	binderHeaderFileContents.WriteString(fmt.Sprintf(namespaceOpen, binderNs))
	binderSourceFileContents.WriteString(fmt.Sprintf(namespaceOpen, binderNs))
	modelHeaderFwdDeclares.WriteString(fmt.Sprintf(namespaceOpen, binderNs))
	binderHeaderFwdDeclares.WriteString(fmt.Sprintf(namespaceOpen, modelNs))
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
		modelHeaderFwdDeclares.WriteString(fmt.Sprintf(modelFwdDeclareFmt, validSchemas[i].ClassName))

		// binder forward declare of model
		binderHeaderFwdDeclares.WriteString(fmt.Sprintf(binderFwdDeclareFmt, validSchemas[i].ClassName))

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
		template.moduleDef = moduleDef

		bindingTemplate := CppTemplate{}
		bindingTemplate.def = filterDef
		bindingTemplate.namespace = fmt.Sprintf(profile.BinderNsFmt, moduleDef.namespace)
		bindingTemplate.moduleDef = moduleDef

		// function defs
		// Generate a TableName() func
		tblNameDef := igenerator.MethodDef{
			IsStatic:    true,
			ClassName:   filterDef.ClassName,
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
		retType := "const std::unordered_set<std::string>"
		colNameDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  retType + "&",
			ClassName:   filterDef.ClassName,
			Name:        "ColumnNames",
			Body:        fmt.Sprintf(funcColumnNamesFmt, strings.Join(colNames, ", "), "columnNames", retType),
			Description: "Returns a set of column names for the table",
		}
		template.AddMethod(colNameDef)

		retType = "const std::vector<std::string>"
		orderedColNameDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  retType + "&",
			ClassName:   filterDef.ClassName,
			Name:        "OrderedColumnNames",
			Body:        fmt.Sprintf(funcColumnNamesFmt, strings.Join(colNames, ", "), "orderedColumnNames", retType),
			Description: "Returns an ordered vector of column names for the table",
		}
		template.AddMethod(orderedColNameDef)

		// Generate a BlobColumns() func
		blobCols := []string{}
		for j := range filterDef.Columns {
			if filterDef.Columns[j].IsBlobType() {
				blobCols = append(blobCols, fmt.Sprintf(`"%s"`, filterDef.Columns[j].Name))
			}
		}
		retType = "const std::unordered_set<std::string>"
		blobColNameDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  retType + "&",
			ClassName:   filterDef.ClassName,
			Name:        "BlobColumns",
			Body:        fmt.Sprintf(funcColumnNamesFmt, strings.Join(blobCols, ", "), "blobColumns", retType),
			Description: "Returns a set of blob column names for the table",
		}
		template.AddMethod(blobColNameDef)

		// Generate a DbType func
		dbTypeDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  "const modelUtil::DbType",
			ClassName:   filterDef.ClassName,
			Name:        "DbType",
			Body:        fmt.Sprintf(funcDbTypeFmt, filterDef.Database),
			Description: "Returns the associated database type for the table",
		}
		template.AddMethod(dbTypeDef)

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
			ClassName:   filterDef.ClassName,
			Name:        "PrimaryKey",
			Body:        fmt.Sprintf(funcPrimaryKeyFmt, pkNames.String()),
			Description: "Returns the columns associated with the table's Primary Key",
		}
		template.AddMethod(pkColumnsDef)

		if len(pk.Columns) > 0 {
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
				ClassName:   filterDef.ClassName,
				Name:        "MapKey",
				Body:        body,
				Description: "Returns a value for use in map keys based on the table's primary key",
			}
			template.AddMethod(mapKeyDef)
		}

		// generate template
		modelClassHeaderDef, modelClassSourceDef, tErr := template.Generate()
		if tErr != nil {
			err = fmt.Errorf("%s failed to generate c++ source: %w", filterDef.Name, tErr)
			return err
		}
		modelHeaderFileContents.WriteString(modelClassHeaderDef)
		modelSourceFileContents.WriteString(modelClassSourceDef)

		// copy includes to main template
		for k, v := range template.includes {
			modelHeaderTemplate.includes[k] = v
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
			IsStatic:       true,
			ReturnType:     "const BindingsMapType&",
			ImplReturnType: fmt.Sprintf("const %s::BindingsMapType&", filterDef.ClassName),
			ClassName:      filterDef.ClassName,
			Name:           "GetColumnBindings",
			Body:           fmt.Sprintf(funcColumnBindingsFmt, bindings.String()),
			Description:    "Returns the binding function associated with the column name",
		}
		bindingTemplate.AddMethod(colBindDef)

		// generate binding template
		binderClassHeaderDef, binderClassSourceDef, tErr := bindingTemplate.GenerateBinderClass()
		if tErr != nil {
			err = fmt.Errorf("%s failed to generate c++ bindings source: %w", filterDef.Name, tErr)
			return err
		}
		binderHeaderFileContents.WriteString(binderClassHeaderDef)
		binderSourceFileContents.WriteString(binderClassSourceDef)
	}

	var headerIncludes []string
	for include := range modelHeaderTemplate.includes {
		headerIncludes = append(headerIncludes, include)
	}

	// includes are built from unordered hash maps
	// sort the includes alphabetically so they don't cause diffs gen-to-gen
	sort.Strings(headerIncludes)

	// close the namespace
	modelHeaderFwdDeclares.WriteString("\n}\n\n")
	binderHeaderFwdDeclares.WriteString("\n}\n")
	modelHeaderFileContents.WriteString("}")
	modelSourceFileContents.WriteString("\n}")

	// combine model content
	modelHeaderFwdDeclares.WriteString(modelHeaderFileContents.String())

	modelHeaderFilename := fmt.Sprintf(primaryHeaderFileName, modelClassName)

	// 1. includes
	// 2. file contents
	modelHeaderStr := fmt.Sprintf(primaryHeaderFmt, strings.Join(headerIncludes, ""), modelHeaderFwdDeclares.String())
	outFile := filepath.Join(modelOut, modelHeaderFilename)
	if fErr := utils.WriteToFile(outFile, modelHeaderStr); fErr != nil {
		err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	modelSourceFileIncludes := fmt.Sprintf(includeFmt, fmt.Sprintf("\"%s\"", modelHeaderFilename))

	// 1. includes
	// 2. file contents
	modelSourceStr := fmt.Sprintf(primarySourceFmt, modelSourceFileIncludes, modelSourceFileContents.String())
	outFile = filepath.Join(modelOut, fmt.Sprintf(primarySourceFileName, modelClassName))
	if fErr := utils.WriteToFile(outFile, modelSourceStr); fErr != nil {
		err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	headerIncludes = []string{}
	for include := range binderHeaderTemplate.includes {
		headerIncludes = append(headerIncludes, include)
	}

	// includes are built from unordered hash maps
	// sort the includes alphabetically so they don't cause diffs gen-to-gen
	sort.Strings(headerIncludes)

	// close the namespace
	binderHeaderFileContents.WriteString("}")
	binderSourceFileContents.WriteString("}")

	bindingHeaderStr := fmt.Sprintf(binderHeaderFmt, strings.Join(headerIncludes, ""), binderHeaderFwdDeclares.String(), binderHeaderFileContents.String())
	outFile = filepath.Join(binderOut, fmt.Sprintf(primaryHeaderFileName, binderClassName))
	if fErr := utils.WriteToFile(outFile, bindingHeaderStr); fErr != nil {
		err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	relativeModelHeaderPath := filepath.Join(moduleDef.OutDir, modelPackageOutDir, fmt.Sprintf(primaryHeaderFileName, modelClassName))
	relativeModelHeaderPath = filepath.ToSlash(relativeModelHeaderPath)

	bindingSourceStr := fmt.Sprintf(binderSourceFmt, fmt.Sprintf(primaryHeaderFileName, binderClassName), relativeModelHeaderPath, binderSourceFileContents.String())
	outFile = filepath.Join(binderOut, fmt.Sprintf(primarySourceFileName, binderClassName))
	if fErr := utils.WriteToFile(outFile, bindingSourceStr); fErr != nil {
		err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	return nil
}

func generateProcClasses(clean bool, validProcs []jsonSchema.ProcDef) (err error) {
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

	headerTemplate := CppTemplate{}
	headerTemplate.AddInclude("<nanodbc/nanodbc.h>")
	headerTemplate.AddInclude("<memory>")
	headerTemplate.AddInclude("<string>")
	headerTemplate.AddInclude("\"detail/StoredProcedure.h\"")
	headerTemplate.AddInclude("<ModelUtil/ModelUtil.h>")

	sourceTemplate := CppTemplate{}
	sourceTemplate.AddInclude("\"StoredProc.h\"")

	procHeaderFileContents := strings.Builder{}
	procHeaderFileContents.WriteString(fmt.Sprintf(namespaceOpen, "storedProc"))
	procHeaderFileContents.WriteString("\n")

	procSourceFileContents := strings.Builder{}
	procSourceFileContents.WriteString(fmt.Sprintf(namespaceOpen, "storedProc"))
	procSourceFileContents.WriteString("\n")

	for i := range validProcs {
		fmt.Println(fmt.Sprintf("generating c++ for: %s", validProcs[i].Name))

		className := validProcs[i].ClassName

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
			paramBindings.WriteString(fmt.Sprintf(procBindFmt, nanodbcBindFunc, 0, "returnValue", nanodbcParamTypeRet))
			funcParamList = append(funcParamList, []string{"int*", "returnValue"})
		} else {
			posMod = -1
			procCallStr = fmt.Sprintf(procCallFmt, validProcs[i].Name, pList)
		}

		for j, param := range validProcs[i].Params {
			cppType, ok := TSqlTypeMapping[param.Type]
			if !ok {
				return fmt.Errorf("unimplemented T-SQL type: %s", param.Type)
			}

			// binary override flag
			if param.ForceBinary {
				cppType = "std::vector<uint8_t>"
			}

			if strings.Contains(cppType, "int") {
				headerTemplate.AddInclude("<cstdint>")
			} else if strings.Contains(cppType, "time_t") {
				headerTemplate.AddInclude("<ctime>")
			}

			bindFunc := nanodbcBindFunc
			if cppType == "std::string" || param.Type == tsql.Text {
				cppType = "char*"
			} else if strings.Contains(cppType, "vector") {
				bindFunc = nanodbcBindBinaryFunc
				headerTemplate.AddInclude("<vector>")
			}

			_type := cppType

			if param.IsOutput {
				hasOutputOrReturn = true

				if !strings.HasSuffix(_type, "*") {
					_type = fmt.Sprintf(ptrFmt, _type)
				}
			} else {
				_type = fmt.Sprintf(constFmt, _type)

				// pass binary fields (std::vector<uint8_t>) by reference
				if bindFunc == nanodbcBindBinaryFunc {
					_type += "&"
				}
			}
			isPtr := strings.HasSuffix(_type, "*")

			// do try do to make twostars happy
			if j == 0 {
				_type = "\n\t\t\t" + _type
			}
			funcParamList = append(funcParamList, []string{_type, param.ParamName})

			// TODO: likely need to add modifiers like .c_str()
			p := param.ParamName
			if !isPtr && bindFunc != nanodbcBindBinaryFunc {
				p = "&" + p
			}

			if param.IsOutput {
				paramBindings.WriteString(fmt.Sprintf(procBindFmt, bindFunc, param.ParamIndex+posMod, p, nanodbcParamTypeOut))
			} else {
				paramBindings.WriteString(fmt.Sprintf(procBindInputFmt, bindFunc, param.ParamIndex+posMod, p))
			}
		}

		// Generate a Query() func
		funcQueryDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  "const std::string&",
			ClassName:   className,
			Name:        "Query",
			Body:        fmt.Sprintf(procFuncQueryFmt, procCallStr),
			Description: "Returns the query associated with preparing this statement",
		}
		classTemplate.AddMethod(funcQueryDef)

		// Generate a DbType func
		dbTypeDef := igenerator.MethodDef{
			IsStatic:    true,
			ReturnType:  "const modelUtil::DbType",
			ClassName:   className,
			Name:        "DbType",
			Body:        fmt.Sprintf(funcDbTypeFmt, validProcs[i].Database),
			Description: "Returns the associated database type for the table",
		}
		classTemplate.AddMethod(dbTypeDef)

		executeBody := procExecuteNoParam
		if paramBindings.Len() > 0 {
			executeBody = fmt.Sprintf(procExecuteFmt, paramBindings.String())
		}

		classTemplate.AddInclude("<memory>")
		executeDef := igenerator.MethodDef{
			IsThrow:    true,
			ReturnType: "std::weak_ptr<nanodbc::result>",
			ClassName:  className,
			Name:       "execute",
			Params:     funcParamList,
			Body:       executeBody,
			Description: `Executes the stored procedure
		/// \throws nanodbc::database_error`,
		}
		classTemplate.AddMethod(executeDef)

		if hasOutputOrReturn {
			destructorDef := igenerator.MethodDef{
				ReturnType:  "",
				ClassName:   className,
				Name:        fmt.Sprintf("~%s", validProcs[i].ClassName),
				Body:        procDestructorWithFlushDef,
				Description: "Flushes any output variables or return values on destruction",
			}
			classTemplate.AddMethod(destructorDef)
		}

		decls := strings.Join(classTemplate.decls, "\n")
		methods := strings.Join(classTemplate.methods, "\n")

		addlDoxygen := fmt.Sprintf(getProcXRefFmt(validProcs[i].Database), validProcs[i].Name, validProcs[i].Description)

		// file contents:
		// 1. Class Name
		// 2. Methods
		// 3. Proc description
		headerFileStr := fmt.Sprintf(procClassHeaderFmt, validProcs[i].ClassName,
			decls, validProcs[i].Description, addlDoxygen)
		procHeaderFileContents.WriteString(headerFileStr)

		// file contents:
		// 1. Class Name
		// 2. Methods
		// 3. Proc description
		sourceFileStr := fmt.Sprintf(procClassImplFmt, validProcs[i].ClassName,
			methods, validProcs[i].Description)
		procSourceFileContents.WriteString(sourceFileStr)
	}

	var headerIncludes []string
	var sourceIncludes []string

	for include := range headerTemplate.includes {
		headerIncludes = append(headerIncludes, include)
	}

	for include := range sourceTemplate.includes {
		sourceIncludes = append(sourceIncludes, include)
	}

	// includes are built from unordered hash maps
	// sort the includes alphabetically so they don't cause diffs gen-to-gen
	sort.Strings(headerIncludes)
	sort.Strings(sourceIncludes)

	// close the namespace
	procHeaderFileContents.WriteString("}")
	procSourceFileContents.WriteString("}")

	// 1. includes
	// 2. file contents
	procHeaderFileStr := fmt.Sprintf(primaryHeaderFmt, strings.Join(headerIncludes, ""), procHeaderFileContents.String())
	outFile := filepath.Join(procOut, fmt.Sprintf(primaryHeaderFileName, procPackageOutDir))
	if fErr := utils.WriteToFile(outFile, procHeaderFileStr); fErr != nil {
		err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	procSourceFileStr := fmt.Sprintf(primarySourceFmt, strings.Join(sourceIncludes, ""), procSourceFileContents.String())
	outFile = filepath.Join(procOut, fmt.Sprintf(primarySourceFileName, procPackageOutDir))
	if fErr := utils.WriteToFile(outFile, procSourceFileStr); fErr != nil {
		err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	return nil
}
