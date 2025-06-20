package cgHelpers

const (
	// static helper file.  We keep it in string format here to keep the tool portable.
	// can take it out of string form to make edits
	KogenTemplate = `package kogen

/** Enums **/
type DbType uint8

const (
	LOGIN DbType = 0
	GAME  DbType = 1
	LOG   DbType = 2
)

type TSqlType string

// doc: https://learn.microsoft.com/en-us/sql/t-sql/data-types/data-types-transact-sql?view=sql-server-ver17#binary-strings
const (
	TinyInt   TSqlType = "tinyint"   //uint8
	SmallInt  TSqlType = "smallint"  //int16
	Int       TSqlType = "int"       // int32
	BigInt    TSqlType = "bigint"    // int64
	Float     TSqlType = "float"     // float64 but value interpretation depends on (n)
	Real      TSqlType = "real"      // float32
	Char      TSqlType = "char"      // byte, fixed length
	Varchar   TSqlType = "varchar"   // byte, variable length
	NChar     TSqlType = "nchar"     // unicode byte, fixed length
	NVarchar  TSqlType = "nvarchar"  // unicode byte, variable length
	Binary    TSqlType = "binary"    // fixed length byte array
	Varbinary TSqlType = "varbinary" // variable-length byte array
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
	GetTableName() string
	GetInsertString() string
	GetCreateTableString() string
}

// GetOptionalStringVal Returns the sql format for an optional string
func GetOptionalStringVal(val *string) string {
	if val == nil {
		return "NULL"
	}
	return "N'%s'"
}

// GetOptionalDecVal Returns the sql format for an optional numeric value
func GetOptionalDecVal(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	return "%d"
}

// GetOptionalBinaryVal Returns the sql format for an optional binary array
// we use interface as the type to support both fixed-length and variable-length byte arrays
func GetOptionalBinaryVal(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	return "N'%s'"
}
`
)
