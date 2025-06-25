package cgHelpers

const (
	// static helper file.  We keep it in string format here to keep the tool portable.
	// can take it out of string form to make edits
	KogenTemplate = `package kogen

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"golang.org/x/exp/constraints"
	mssql "github.com/microsoft/go-mssqldb"
)

/** Enums **/
type DbType uint8

const (
	LOGIN DbType = 0
	GAME  DbType = 1
	LOG   DbType = 2
)

/** End Enums **/

var (
	// These can be changed by the importing program for customized output at runtime
	_loginDbName = "ACCOUNT"
	_gameDbName  = "GAME"
	_logDbName   = "LOG"

	// Model list holds the list of all generated models;
	// Models add themselves to this list in their init() function
	ModelList = []Model{}
)

func SetDbNames(loginDbName string, gameDbName string, logDbName string) {
	_loginDbName = loginDbName
	_gameDbName = gameDbName
	_logDbName = logDbName
}

func GetDatabaseName(dbType DbType) string {
	switch dbType {
	case LOGIN:
		return _loginDbName
	case GAME:
		return _gameDbName
	case LOG:
		return _logDbName
	default:
		return ""
	}
}

// Model is an interface that implements all expected helper functions
type Model interface {
	GetDatabaseName() string
	TableName() string
	GetInsertString() string
	GetInsertHeader() string
	GetInsertData() string
	GetCreateTableString() string
	GetAllTableData(db *gorm.DB) (results []Model, err error)
	SelectClause() (selectClause clause.Select)
}

// GetOptionalStringVal Returns the sql format for an optional string
func GetOptionalStringVal(val *string, isHexProtect bool) string {
	if val == nil {
		return "NULL"
	}
	if isHexProtect {
		return "0x" + hex.EncodeToString([]byte(*val))
	}
	return fmt.Sprintf("N'%s'", treatSqlVal(*val))
}

func GetOptionalVarCharVal(val *mssql.VarChar, isHexProtect bool) string {
	if val == nil {
		return "NULL"
	}
	cast := string(*val)
	if isHexProtect {
		return "0x" + hex.EncodeToString([]byte(cast))
	}
	// read the generic type into a string for use; generics don't play well with dereferencing
	return fmt.Sprintf("N'%s'", treatSqlVal(cast))
}

func GetOptionalNCharVal(val *mssql.NChar, isHexProtect bool) string {
	if val == nil {
		return "NULL"
	}
	cast := string(*val)
	if isHexProtect {
		return "0x" + hex.EncodeToString([]byte(cast))
	}
	// read the generic type into a string for use; generics don't play well with dereferencing
	return fmt.Sprintf("N'%s'", treatSqlVal(cast))
}

func GetOptionalHexStringVal(val *string) string {
	if val == nil {
		return "NULL"
	}
	return "0x" + hex.EncodeToString([]byte(*val))
}

// GetOptionalDecVal Returns the sql format for an optional numeric value
func GetOptionalDecVal[T constraints.Integer | constraints.Float](val *T) string {
	if val == nil {
		return "NULL"
	}
	return fmt.Sprintf("%v", *val)
}

// GetOptionalBinaryVal Returns the sql format for an optional binary array
func GetOptionalBinaryVal(val *[]byte) string {
	if val == nil {
		return "NULL"
	}
	return "0x" + hex.EncodeToString(*val)
}

func GetOptionalByteArrayVal(val *[]byte, isHexProtect bool) string {
	if val == nil {
		return "NULL"
	}
	if isHexProtect {
		return "0x" + hex.EncodeToString(*val)
	}
	return fmt.Sprintf("N'%s'", treatSqlVal(string(*val)))
}

// GetDateTimeExportFmt Returns an INSERT-friendly value from time.Time
func GetDateTimeExportFmt(t *time.Time) string {
	if t == nil {
		return "NULL"
	}
	// CAST(N'2012-11-11T06:59:06.643' AS DateTime)
	return fmt.Sprintf("CAST(N'%s' AS DateTime)", t.Format("2006-01-02T15:04:05.999"))
}

// treatSqlVal handles escaping/modifying values for SQL insertion
func treatSqlVal(val string) string {
	return strings.ReplaceAll(val, "'", "''")
}

`
)
