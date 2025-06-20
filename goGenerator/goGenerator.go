package goGenerator

import (
	"fmt"
	cgHelpers "ko-codegen/goGenerator/cgHelpers/kogen"
	"ko-codegen/igenerator"
	"ko-codegen/jsonSchema"
	"ko-codegen/jsonSchema/enums/tsql"
	"ko-codegen/utils"
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

	// 1. Table Name
	// 2. Column List
	// 3. Values list
	insertTemplateFmt = `"INSERT INTO [%s] (%s) \nVALUES (%s)"`

	// 1. Table Name
	// 2. Column Defs
	// 3. Primary Key Def (optional)
	// 4. Database Name
	createTableTemplateFmt = `"CREATE TABLE \"%[1]s\" (\n%[2]s\n%[3]s\n)"`

	// 1. Column Name
	// 2. Column Type
	// 3. Additional Constraints
	createColumnTemplateFmt = `\t\"%[1]s\" %[2]s%[3]s`

	// 1. PK Columns, string wrapped and csv
	primaryKeyFmt = `\tPRIMARY KEY (%[1]s)`
)

// generateGo generates go code files for each schema in utils.SchemaDir, and writes the result to the output dir (utils.OutputDir)
func GenerateGo(clean bool) (err error) {
	// Read and compile all schema files in jsonSchema
	validSchemas, err := utils.LoadSchemas()
	if err != nil {
		return err
	}

	if clean {
		// Go clean needs to be specific to the directories it writes to, as to avoid
		// deleting anything in openko-gorm/* when specified via command line
		err = os.RemoveAll(filepath.Join(utils.OutputDir, kogenPackageOutDir))
		if err != nil {
			fmt.Printf("failed to clean the output directory: %w\n", err)
			return
		}
	}

	err = setupOutDir()
	if err != nil {
		return err
	}

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

		// We need a few constants to save on memory:
		// _(tableName)DatabaseNbr for GetDatabaseName
		// _(tableName)TableName for GetTableName
		// we lead with underscores to avoid potential collisions
		dbNbrConst := fmt.Sprintf(databaseConstFmt, validSchemas[i].ClassName)
		tableNameConst := fmt.Sprintf(tableNameConstFmt, validSchemas[i].ClassName)
		template.AddConst(dbNbrConst, fmt.Sprintf(decFmt, validSchemas[i].Database))
		template.AddConst(tableNameConst, fmt.Sprintf(strWrapFmt, validSchemas[i].Name))

		// Generate a GetDatabaseName() func
		dbNameDef := igenerator.MethodDef{
			ReturnType:  "string",
			Name:        "GetDatabaseName",
			Body:        fmt.Sprintf("\treturn GetDatabaseName(DbType(%s))", dbNbrConst),
			Description: "Returns the table's database name",
		}
		template.AddMethod(dbNameDef)

		// Generate a GetTableName() func
		tblNameDef := igenerator.MethodDef{
			ReturnType:  "string",
			Name:        "GetTableName",
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

		// Generate a GetCreateTableString() func
		createTableDef := igenerator.MethodDef{
			ReturnType:  "string",
			Name:        "GetCreateTableString",
			Body:        generateCreateTableBody(validSchemas[i]),
			Description: "Returns the create table statement for this object",
		}
		template.AddMethod(createTableDef)

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
			fmt.Printf("WARN: failed to run gofmt on %s: %w\n", outFile, fmtErr)
		}
	}

	fmt.Println("Go code generated successfully")
	return nil
}

func setupOutDir() error {
	// create moduleOutDir if it doesn't exist in the utils.OutputDir
	return os.MkdirAll(filepath.Join(utils.OutputDir, kogenPackageOutDir), os.ModePerm)
}

func writeCgHelpers() error {
	outFile := filepath.Join(utils.OutputDir, kogenPackageOutDir, cgHelperFileName)
	if fErr := utils.WriteToFile(outFile, cgHelpers.KogenTemplate); fErr != nil {
		err := fmt.Errorf("failed to write file %s: %w", outFile, fErr)
		return err
	}

	return nil
}

func generateInsertTemplateBody(def jsonSchema.TableDef) string {
	// 1. Table Name
	// 2. Column List
	// 3. Values list

	columnNames := []string{}
	valuesFmt := []string{}
	propRefs := []string{}
	for i := range def.Columns {
		columnNames = append(columnNames, def.Columns[i].Name)
		valuesFmt = append(valuesFmt, "%s")
		propRefs = append(propRefs, getPropRefByType(def.Columns[i]))
	}

	insertFmt := fmt.Sprintf(insertTemplateFmt, def.Name, strings.Join(columnNames, ", "), strings.Join(valuesFmt, ", "))
	return fmt.Sprintf("\treturn fmt.Sprintf(%s,%s)", insertFmt, strings.Join(propRefs, ",\n"))
}

func generateCreateTableBody(def jsonSchema.TableDef) string {
	columnDefs := []string{}
	primaryKeys := []string{}
	for i := range def.Columns {
		opt := ""
		if !def.Columns[i].AllowNull {
			opt = " NOT NULL"
		}
		columnDefs = append(columnDefs, fmt.Sprintf(createColumnTemplateFmt, def.Columns[i].Name, def.Columns[i].GormType(), opt))
		if def.Columns[i].IsPrimaryKey {
			primaryKeys = append(primaryKeys, fmt.Sprintf(`\"%s\"`, def.Columns[i].Name))
		}
	}

	pkDef := ""
	if len(primaryKeys) > 0 {
		pkDef = fmt.Sprintf(primaryKeyFmt, strings.Join(primaryKeys, ", "))
	}

	createTableSql := fmt.Sprintf(createTableTemplateFmt, def.Name, strings.Join(columnDefs, `,\n`), pkDef)
	return wrapQueryWithUseDbFmt(createTableSql)
}

func wrapQueryWithUseDbFmt(query string) string {
	queryVar := fmt.Sprintf("\tquery := %s\n", query)
	returnLn := `return fmt.Sprintf("USE [%[1]s]\nGO\n\n%[2]s", this.GetDatabaseName(), query)`
	return queryVar + returnLn
}

func getPropRefByType(col jsonSchema.Column) string {
	// return GetOptionalStringVal([&]col.PropertyName)
	pName := fmt.Sprintf("this.%s", col.PropertyName)

	switch col.Type {
	case tsql.Char, tsql.NChar, tsql.Varchar, tsql.NVarchar:
		// TODO:  Is string-specific callout needed?  Currently handling all of these as byte arrays
		//return fmt.Sprintf("GetOptionalStringVal(%s)", pName)
		fallthrough
	case tsql.Binary, tsql.VarBinary:
		// Going to have to work with this carefully when I get a good example to diff against
		return fmt.Sprintf("GetOptionalBinaryVal(%s)", pName)
	default:
		// pass non-optional properties by reference to re-use the functions
		// Ignore slices; they are always pointers
		if !col.AllowNull {
			pName = fmt.Sprintf("&%s", pName)
		}
		return fmt.Sprintf("GetOptionalDecVal(%s)", pName)
	}
}
