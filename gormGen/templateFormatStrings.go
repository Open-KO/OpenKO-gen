package gormGen

// file contains all Sprintf pattern strings used in generation

const (
	// FileTemplate args:
	// 1. Class Name
	// 2. Properties
	// 3. Methods
	// 4. Class Description
	// 5. Includes
	// 6. Package Name
	// 7. Constants
	// 8. Additional code
	// fileTemplateFmt is the template for the entire .Go file
	fileTemplateFmt string = `package %[6]s

import (
	"fmt"
%[5]s)

%[7]s

func init() {
	ModelList = append(ModelList, &%[1]s{})
}

// %[1]s %[4]s
type %[1]s struct
{
%[2]s
}


%[3]s

%[8]s
`

	// 1. Type
	// 2. PropertyName
	// 3. Description
	// 4. DbName
	// 5. JsonTag
	// 6. GormTag
	// propertyFmt is used to generate a model struct property
	propertyFmt string = "\t%[2]s %[1]s `%[6]s %[5]s`"

	/*
		json binding
	*/
	jsonTagFmt       string = "json:\"%[1]s\""
	jsonTagSeparator string = ","
	jsonTagOptional  string = "omitempty"

	/*
		Gorm binding
	*/
	gormTagFmt             string = "gorm:\"%[1]s\""
	gormTagSeparator       string = ";"
	gormTagColumnNameFmt   string = "column:%[1]s"
	gormTagPrimaryKey      string = "primaryKey"
	gormTagDefaultValueFmt string = "default:%[1]s"
	gormTagNotNULL         string = "not null"
	gormTypeTagFmt         string = "type:%[1]s"
	// unimplemented
	// gormTagAutoIncrement   string = "autoIncrement"
	// gormTagUnique string = "unique"
	// gormTagIndex string = "index"
	// gormTagForeignKey string = "foreignKey"
	// gormTagUniqueIndex string = "uniqueIndex"

	// 1. Return Type
	// 2. FuncName
	// 3. Parameters
	// 4. Function Body
	// 5. Description
	// 6. ClassName
	// methodFmt is used to generate a model object function
	methodFmt string = `// %[2]s %[5]s
func (this %[6]s) %[2]s(%[3]s) (%[1]s) {
%[4]s
}`

	// 1: import name
	// includeFmt is used to generate a code import
	includeFmt string = "\t\"%[1]s\"\n"

	// 1: formatted concat of constant variables
	// constBlockFmt is used to generate the constant section wrapper
	constBlockFmt string = "const (\n%[1]s)"

	// 1: prop name
	// 2: prop value
	// constPropFmt is used to generate a single constant declaration
	constPropFmt string = "\t%[1]s = %[2]s\n"

	// 1. Table Name
	// 2. Column List
	// 3. Populated values list
	// insertHeaderFmt generates the header line for the insert dump files
	insertHeaderFmt = `"INSERT INTO [%[1]s] (%[2]s) VALUES\n"`
	// insertTemplateFmt generates the full insert statement for a model object
	insertTemplateFmt = `"INSERT INTO [%[1]s] (%[2]s) VALUES\n(%[3]s)"`

	// 1. Table Name
	// 2. Column Defs
	// 3. Primary Key Def (optional)
	// 4. Indexes
	// 5. Constraints
	// createTableTemplateFmt generates the SQL to create model object table
	createTableTemplateFmt = `"CREATE TABLE [%[1]s] (\n%[2]s%[3]s\n)\nGO\n%[4]s%[5]s"`

	// 1. Column Name
	// 2. Column Type
	// 3. Additional Constraints
	// createColumnTemplateFmt generates a column definition for the create table query
	createColumnTemplateFmt = `\t[%[1]s] %[2]s%[3]s`

	// 1. PK Name
	// 2. Clustering
	// 3. PK Columns, string wrapped and csv
	// primaryKeyFmt generates the primary key constraint
	primaryKeyFmt = `\n\tCONSTRAINT [%[1]s] PRIMARY KEY %[2]s (%[3]s)`

	// 1. Table name
	// 2. Column name
	// 3. Default val
	// defaultValFmt generates a default value constraint
	defaultValFmt = `ALTER TABLE [%[1]s] ADD CONSTRAINT [DF_%[1]s_%[2]s] DEFAULT %[3]s FOR [%[2]s]\nGO\n`

	// 1. [Unique] [(NON)CLUSTERED]
	// 2. Index name
	// 3. Table name
	// 4. Columns, string wrapped and csv
	// indexFormat generates a unique index constraint
	indexFormat = `CREATE %[1]s INDEX [%[2]s] ON [%[3]s] (%[4]s)\nGO\n`

	// 1. ClassName
	// 2. SelectQuery.Columns.Name list, csv
	// 3. Table name
	// getTableDataBodyFmt generates the gorm model code needed to select all table data for export
	getTableDataBodyFmt = `res := []%[1]s{}
	rawSql := "SELECT %[2]s FROM [%[3]s]"
	err = db.Raw(rawSql).Find(&res).Error
	if err != nil {
		return nil, err
	}
	for i := range res {
		results = append(results, &res[i])
	}
	return results, nil`

	// package SelectClause format
	// 1. ClassName
	// selectVarNameFmt generates the gorm SelectClause variable name
	selectVarNameFmt = "_%[1]sSelectClause"

	// 1. Columns formatted with colFmt
	// 2. SelectClause package variable name
	// selectVarFmt generates the gorm SelectCause variable
	selectVarFmt = `var %[2]s = clause.Select{
	Columns: []clause.Column{%[1]s
	},
}`

	// 1. Column name from definition, with any modifiers
	// 2. Table name
	// colFmt generates a SelectClause column definition
	colFmt = `
		clause.Column{
			Name: "%[1]s",
		},`
)
