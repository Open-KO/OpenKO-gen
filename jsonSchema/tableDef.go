package jsonSchema

import (
	"ko-codegen/enums/dbType"
	"ko-codegen/enums/genType"
)

type TableDef struct {
	Database    dbType.DbType `json:"database"`    // Which database this table should be created inside of
	Name        string        `json:"name"`        // Table name in the database
	ClassName   string        `json:"className"`   // Code-friendly class/struct name
	Description string        `json:"description"` // Table description
	Columns     []Column      `json:"columns"`     // Columns belonging to the table
}

type Column struct {
	Name         string          `json:"name"`                   // Column name in the database
	PropertyName string          `json:"propertyName"`           // Code-friendly property name
	Description  string          `json:"description"`            // Property description
	Type         genType.GenType `json:"type"`                   // Property Type (uint8, string, etc)
	IsPrimaryKey bool            `json:"isPrimaryKey,omitempty"` // Is this column part of the table's Primary Key?
	DefaultValue string          `json:"defaultValue,omitempty"` // Default value that should be assigned to the property
	AllowNull    bool            `json:"allowNull,omitempty"`    // Can the column's value be null?
	MaxLength    int             `json:"maxLength,omitempty"`    // only applies to Strings (varchars)
}
