package gormGen

import (
	"fmt"
	"github.com/kenner2/OpenKO-db/jsonSchema"
	cgHelpers "openko-gen/gormGen/cgHelpers/kogen"
	"openko-gen/igenerator"
	"openko-gen/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	kogenPackageOutDir = "kogen" // kogen package will be flat until we introduce multi-db support
	cgHelperFileName   = "kogen.go"
	databaseConstFmt   = "_%sDatabaseNbr"
	strWrapFmt         = "\"%s\""
	decFmt             = "%d"
	tableNameConstFmt  = "_%sTableName"
)

// GenerateGo generates go code files for each schema in OpenKO-db/jsonSchema,
// and writes the result to the output dir (default: ./openko-gorm/)
func GenerateGo(clean bool) (err error) {
	// Read and bind all *.json files in jsonSchema
	validSchemas, err := utils.LoadSchemas()
	if err != nil {
		return err
	}

	if clean {
		// Go clean needs to be specific to the directories it writes to, as to avoid
		// deleting anything in openko-gorm/* when specified via command line
		err = os.RemoveAll(filepath.Join(utils.OutputDir, kogenPackageOutDir))
		if err != nil {
			fmt.Printf("failed to clean the output directory: %v\n", err)
			return
		}
	}

	// setupOutputDir creates the output directory if it doesn't exist
	err = setupOutDir()
	if err != nil {
		return err
	}

	// writeCgHelpers writes the hand-coded helper file from cgHelpers/kogen to the output folder
	err = writeCgHelpers()
	if err != nil {
		return err
	}

	for i := range validSchemas {
		fmt.Print(fmt.Sprintf("generating Go for: %s", validSchemas[i].Name))

		// the template is an interface implementation that allows us to
		// structure and generate a code file
		template := GoTemplate{}
		template.def = validSchemas[i]
		template.AddInclude("gorm.io/gorm")
		template.AddInclude("gorm.io/gorm/clause")

		// We need a few constants to save on memory:
		// _(tableName)DatabaseNbr for GetDatabaseName
		// _(tableName)TableName for GetTableName
		// we lead with underscores to avoid potential collisions
		dbNbrConst := fmt.Sprintf(databaseConstFmt, validSchemas[i].ClassName)
		tableNameConst := fmt.Sprintf(tableNameConstFmt, validSchemas[i].ClassName)
		template.AddConst(dbNbrConst, fmt.Sprintf(strWrapFmt, validSchemas[i].Database))
		template.AddConst(tableNameConst, fmt.Sprintf(strWrapFmt, validSchemas[i].Name))

		// Generate a GetDatabaseName() func
		dbNameDef := igenerator.MethodDef{
			ReturnType:  "string",
			Name:        "GetDatabaseName",
			Body:        fmt.Sprintf("\treturn GetDatabaseName(%s)", dbNbrConst),
			Description: "Returns the table's database name",
		}
		template.AddMethod(dbNameDef)

		// Generate a GetTableName() func
		tblNameDef := igenerator.MethodDef{
			ReturnType:  "string",
			Name:        "TableName",
			Body:        fmt.Sprintf("\treturn %s", tableNameConst),
			Description: "Returns the table name",
		}
		template.AddMethod(tblNameDef)

		// Generate a GetInsertString() func
		insertStrDef := igenerator.MethodDef{
			ReturnType:  "string",
			Name:        "GetInsertString",
			Body:        generateInsertTemplateBody(validSchemas[i]),
			Description: "Returns the insert statement for the table populated with record from the object",
		}
		template.AddMethod(insertStrDef)

		// Generate a GetInsertHeader() func
		insertHeaderDef := igenerator.MethodDef{
			ReturnType:  "string",
			Name:        "GetInsertHeader",
			Body:        generateInsertHeaderBody(validSchemas[i]),
			Description: "Returns the header for the table insert dump (insert into table (cols) values",
		}
		template.AddMethod(insertHeaderDef)

		// Generate a GetInsertData() func
		insertDataDef := igenerator.MethodDef{
			ReturnType:  "string",
			Name:        "GetInsertData",
			Body:        generateInsertDataBody(validSchemas[i]),
			Description: "Returns the record data for the table insert dump",
		}
		template.AddMethod(insertDataDef)

		// Generate a GetCreateTableString() func
		createTableDef := igenerator.MethodDef{
			ReturnType:  "string",
			Name:        "GetCreateTableString",
			Body:        generateCreateTableBody(validSchemas[i]),
			Description: "Returns the create table statement for this object",
		}
		template.AddMethod(createTableDef)

		// generate a SelectClause() func
		selectClauseDef := igenerator.MethodDef{
			ReturnType:  "selectClause clause.Select",
			Name:        "SelectClause",
			Body:        fmt.Sprintf(`return %s`, fmt.Sprintf(selectVarNameFmt, validSchemas[i].ClassName)),
			Description: "Returns a safe select clause for the model",
		}
		template.AddMethod(selectClauseDef)

		// additional code snippets
		// selectVar is used to generate a _SelectClause package variable
		selectVar, colNames := GenerateSelectVar(validSchemas[i])
		template.additionalCode = append(template.additionalCode, selectVar)

		// Generate a GetAllTableData method
		getTableDataDef := igenerator.MethodDef{
			ReturnType:  "results []Model, err error",
			Name:        "GetAllTableData",
			Params:      [][]string{{"db", "*gorm.DB"}},
			Body:        fmt.Sprintf(getTableDataBodyFmt, validSchemas[i].ClassName, strings.Join(colNames, ", "), validSchemas[i].Name),
			Description: "Returns a list of all table data",
		}
		template.AddMethod(getTableDataDef)

		// generate template
		templateStr, tErr := template.Generate()
		if tErr != nil {
			err = fmt.Errorf("%s failed to generate Go source: %w", validSchemas[i].Name, tErr)
			return err
		}

		// write the template to a file
		outFile := filepath.Join(utils.OutputDir, kogenPackageOutDir, template.GetFileName())
		if fErr := utils.WriteToFile(outFile, templateStr); fErr != nil {
			err = fmt.Errorf("failed to write file %s: %w", outFile, fErr)
			return err
		}
		fmt.Println(fmt.Sprintf("... written to: %s", outFile))

		// attempt gofmt - will only work if Go is installed on the machine
		// we'll log a warn on error
		fmtErr := exec.Command("gofmt", "-w", outFile).Run()
		if fmtErr != nil {
			fmt.Printf("WARN: failed to run gofmt on %s: %v\n", outFile, fmtErr)
		}
	}

	fmt.Println("Go code generated successfully")
	return nil
}

// setupOutputDir creates the output directory if it doesn't exist
func setupOutDir() error {
	return os.MkdirAll(filepath.Join(utils.OutputDir, kogenPackageOutDir), os.ModePerm)
}

// writeCgHelpers writes the hand-coded helper file from cgHelpers/kogen to the output folder
func writeCgHelpers() error {
	outFile := filepath.Join(utils.OutputDir, kogenPackageOutDir, cgHelperFileName)
	if fErr := utils.WriteToFile(outFile, cgHelpers.KogenTemplate); fErr != nil {
		err := fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	return nil
}

// generateInsertTemplateBody generates the function body of GetInsertString()
func generateInsertTemplateBody(def jsonSchema.TableDef) string {
	columnNames := []string{}
	valuesFmt := []string{}
	propRefs := []string{}
	for i := range def.Columns {
		columnNames = append(columnNames, fmt.Sprintf("[%s]", def.Columns[i].Name))
		values := "%s"
		if def.Columns[i].IsHexProtect {
			ln := "MAX"
			if def.Columns[i].Length > 0 {
				ln = fmt.Sprintf("%d", def.Columns[i].Length)
			}
			values = fmt.Sprintf("CONVERT(%[1]s(%[2]s), %[3]s)", def.Columns[i].Type, ln, values)
		}
		valuesFmt = append(valuesFmt, values)
		propRefs = append(propRefs, getPropRefByType(def.Columns[i]))
	}

	// 1. Table Name
	// 2. Column List
	// 3. Values list
	insertFmt := fmt.Sprintf(insertTemplateFmt, def.Name, strings.Join(columnNames, ", "), strings.Join(valuesFmt, ", "))
	return fmt.Sprintf("\treturn fmt.Sprintf(%s,%s)", insertFmt, strings.Join(propRefs, ",\n"))
}

// generateInsertHeaderBody generates the function body of GetInsertHeader()
func generateInsertHeaderBody(def jsonSchema.TableDef) string {
	columnNames := []string{}
	for i := range def.Columns {
		columnNames = append(columnNames, fmt.Sprintf("[%s]", def.Columns[i].Name))
	}

	// 1. Table Name
	// 2. Column List
	header := fmt.Sprintf(insertHeaderFmt, def.Name, strings.Join(columnNames, ", "))
	return fmt.Sprintf("\treturn %s", header)
}

// generateInsertDataBody generates the function body of GetInsertData()
func generateInsertDataBody(def jsonSchema.TableDef) string {
	valuesFmt := []string{}
	propRefs := []string{}
	for i := range def.Columns {
		values := "%s"
		if def.Columns[i].IsHexProtect {
			ln := "MAX"
			if def.Columns[i].Length > 0 {
				ln = fmt.Sprintf("%d", def.Columns[i].Length)
			}
			values = fmt.Sprintf("CONVERT(%[1]s(%[2]s), %[3]s)", def.Columns[i].Type, ln, values)
		}
		valuesFmt = append(valuesFmt, values)
		propRefs = append(propRefs, getPropRefByType(def.Columns[i]))
	}

	dataFmt := fmt.Sprintf(`"(%s)"`, strings.Join(valuesFmt, ", "))
	return fmt.Sprintf("\treturn fmt.Sprintf(%s,%s)", dataFmt, strings.Join(propRefs, ",\n"))
}

// generateCreateTableBody generates the function body of GetCreateTableString()
func generateCreateTableBody(def jsonSchema.TableDef) string {
	columnDefs := []string{}
	constraints := []string{}
	for i := range def.Columns {
		opt := ""
		if !def.Columns[i].AllowNull {
			opt = " NOT NULL"
		}
		columnDefs = append(columnDefs, fmt.Sprintf(createColumnTemplateFmt, def.Columns[i].Name, def.Columns[i].GormType(), opt))
		if def.Columns[i].DefaultValue != "" {
			constraints = append(constraints, fmt.Sprintf(defaultValFmt, def.Name, def.Columns[i].Name, def.Columns[i].DefaultValue))
		}
	}

	pkDef := ""
	indexes := []string{}
	for i := range def.Indexes {
		// format columns
		cols := []string{}
		for j := range def.Indexes[i].Columns {
			cols = append(cols, fmt.Sprintf("[%s]", def.Indexes[i].Columns[j]))
		}
		colList := strings.Join(cols, ", ")
		if def.Indexes[i].IsPrimaryKey {
			pkDef = fmt.Sprintf(primaryKeyFmt, def.Indexes[i].Name, def.Indexes[i].Type, colList)
		} else {
			ixType := def.Indexes[i].Type
			if def.Indexes[i].IsUnique {
				ixType = "UNIQUE " + ixType
			}
			indexes = append(indexes, fmt.Sprintf(indexFormat, ixType, def.Indexes[i].Name, def.Name, colList))
		}
	}
	indexOut := strings.Join(indexes, "")

	constr := ""
	if len(constraints) > 0 {
		constr = strings.Join(constraints, "")
	}

	createTableSql := fmt.Sprintf(createTableTemplateFmt, def.Name, strings.Join(columnDefs, `,\n`), pkDef, indexOut, constr)
	return wrapQueryWithUseDbFmt(createTableSql)
}

// wrapQueryWithUseDbFmt wraps an input query with a USE [db] statement
func wrapQueryWithUseDbFmt(query string) string {
	queryVar := fmt.Sprintf("\tquery := %s\n", query)
	returnLn := `return fmt.Sprintf("USE [%[1]s]\nGO\n\n%[2]s", this.GetDatabaseName(), query)`
	return queryVar + returnLn
}

// GenerateSelectVar generates the package _SelectClause variable
func GenerateSelectVar(def jsonSchema.TableDef) (selectVar string, colNames []string) {
	cols := []string{}
	for i := range def.Columns {
		colName := def.Columns[i].Name
		if def.Columns[i].IsHexProtect {
			ln := "MAX"
			if def.Columns[i].Length > 0 {
				ln = fmt.Sprintf("%d", def.Columns[i].Length)
			}
			colName = fmt.Sprintf("CONVERT(VARBINARY(%[1]s), [%[2]s]) as [%[2]s]", ln, colName)
		} else {
			colName = fmt.Sprintf("[%s]", colName)
		}
		colNames = append(colNames, colName)
		cols = append(cols, fmt.Sprintf(colFmt, colName, def.Name))
	}

	varName := fmt.Sprintf(selectVarNameFmt, def.ClassName)
	return fmt.Sprintf(selectVarFmt, strings.Join(cols, ""), varName), colNames
}
