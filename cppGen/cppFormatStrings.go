package cppGen

const (
	// 1: Module name
	// 2. includes
	// 3. imports
	// 4: file contents
	// TODO: \mainpage doc?
	// primaryModuleFmt is the template for the primary module file
	primaryModuleFmt = `module;

%[2]s
export module %[1]s;

%[3]s
%[4]s`

	// 1: partition module to export
	//exportImportFmt = "export import :%[1]s;\n"

	// 1: ClassName
	// 2: file contents
	// 3: includes
	// 4: Model or Binder
	// partitionModuleFmt is the template for a module partition
	//	partitionModuleFmt = `module;
	//
	//%[3]s
	//export module %[4]s:%[1]s;
	//
	//%[2]s`

	// 1. ClassName
	// 2. Class contents
	// 3. binder namespace
	// 4. model namespace
	//	modelFileFmt = `namespace %[3]s
	//{
	//	export class %[1]s;
	//}
	//
	//namespace %[4]s
	//{
	//%[2]s
	//}
	//`

	// 1. Class Name
	modelFwdDeclareFmt = "\n\texport class %[1]s;"

	// 1. ClassName
	// 2. Member defs
	// 3. Method defs
	// 4. Class-level Doxygen block
	// 5. binder namespace
	modelClassFmt = `
%[4]s
	export class %[1]s 
	{
	/// \publicsection
	public:
		using BinderType = %[5]s::%[1]s;
%[2]s
%[3]s

	};
`

	// 1. doxygen block
	// 2. union array def
	// 3. column list
	unionArrayFmt = `
		union
		{%[1]s
%[2]s

			struct
			{
%[3]s
			};
		};`

	// 1. cppType
	// 2. Union Array Name
	// 3. Union Array Len
	// 4. initialized value
	unionArrayDefFmt = "%[1]s %[2]s[%[3]d]%[4]s;"

	// 1. first column name in array group
	// 2. last column name in array group
	// 3. property name
	unionArrayDoxygenFmt = `
	/// \brief Union array grouping for columns [%[1]s] to [%[2]s]
	///
	/// \property %[3]s`

	// 1. name
	// 2. contents
	namespaceFmt = `namespace %[1]s
{%[2]s
}
`
	// 1. namespace name
	namespaceOpen = `namespace %[1]s
{`

	// 1. ClassName
	// 2. Method defs
	// 3. model namespace
	binderClassFmt = `
	/// \brief generated nanodbc column binder for %[3]s::%[1]s
	export class %[1]s
	{
	/// \publicsection
	public:
		typedef void (*BindColumnFunction_t)(%[3]s::%[1]s& m, const nanodbc::result& result, short colIndex);

		using BindingsMapType = std::unordered_map<std::string, BindColumnFunction_t>;
%[2]s

	};
`

	// 1: header file to include
	includeFmt = "#include %[1]s\n"
	importFmt  = "import %[1]s;\n"

	// 1: cppType
	// 2: PropertyName
	// 3: initialized value
	// 4: associated enum
	memberFmt = "%[1]s %[2]s%[3]s;%[4]s"

	// 1. enumName
	// 2. Value list
	// 3. Column Name
	enumFmt = `

/// \enum %[1]s
/// \brief Known valid values for %[3]s
enum class %[1]s
{
%[2]s
};`

	// 1. description
	// 2. modifiers (static, inline, etc)
	// 3. return type
	// 4. function name
	// 5. params, csv
	// 6. function body
	// 7. pure
	methodFmt = `
		/// \brief %[1]s
		%[2]s%[3]s %[4]s(%[5]s)%[7]s
		{%[6]s
		}`

	// 1: table name
	funcTableNameFmt = `
			static const std::string tableName = "%[1]s";
			return tableName;`

	// 1: list of column names, string wrapped and CSV
	// 2: Static const name
	funcColumnNamesFmt = `
			static const std::unordered_set<std::string> %[2]s =
			{
				%[1]s
			};
			return %[2]s;`

	funcDbTypeFmt = `
			return modelUtil::DbType::%[1]s;`

	// 1: list of column names in the pk, string wrapped and CSV
	funcPrimaryKeyFmt = `
			static const std::vector<std::string> primaryKey =
			{
				%[1]s
			};
			return primaryKey;`

	// 1 Binding map entries
	funcColumnBindingsFmt = `
			static const BindingsMapType bindingsMap =
			{%[1]s
			};
			return bindingsMap;`

	// 1. PK Property Name
	funcMapKeySingleFmt = `
			return %[1]s;`

	// 1. tuple def
	// 2. tuple values, csv
	funcMapKeyMultiFmt = `
			return %[1]s{%[2]s};`

	// 1. Column Name
	// 2. Class Name
	// 3. Property Name
	bindingFmt = `
				{"%[1]s", &%[2]s::Bind%[3]s}`

	// 1. cppType
	// 2. PropertyName
	funcPropBindingFmt = `
			result.get_ref<%[1]s>(colIndex, m.%[2]s);`

	// 1. cppType
	// 2. PropertyName
	funcPropBindingGetFmt = `
				m.%[2]s = result.get<%[1]s>(colIndex);`

	// 1. original type
	// 2. upcast type
	// 3. Property name
	funcPropBindingUpCastFmt = `
			m.%[3]s = static_cast<%[1]s>(result.get<%[2]s>(colIndex));`

	// 1. Property Name
	funcPropBindingDateCastFmt = `
			m.%[1]s = binderUtil::CTimeFromDbTime(result.get<nanodbc::timestamp>(colIndex));`

	// 1. cppType, optional stripped.  upscale and static cast back Tinyint to SmallInt
	// 2. PropertyName
	// 3. proper binding func
	funcOptionalPropBindingFmt = `
			if (result.is_null(colIndex))
			{
				m.%[2]s.reset();
			}
			else
			{%[3]s
			}`

	// 1. Type
	constRefFmt = "const %s&"
	ptrFmt      = "%s*"
	constPtrFmt = "const %s*"
	constFmt    = "const %s"

	// 1. Class Name
	// 2. Procedure Call prepared statement i.e., {? = CALL LOAD_ACCOUNT_CHARID(?)}
	// 3. Methods
	// 4. Class description
	procClassFmt = `
	/// \brief %[4]s
	/// \class %[1]s
	export class %[1]s : public StoredProcedure
	{
	public:
		%[1]s(nanodbc::connection& conn) 
			: StoredProcedure(conn)
		{
			_stmt.prepare("%[2]s");
		}
%[3]s
	};
`
	// 1. proc name
	// 2. "?" list, csv for len of param
	procCallFmt        = "{CALL %[1]s(%[2]s)}"
	procCallWithRetFmt = "{? = CALL %[1]s(%[2]s)}"

	// 1. binding list
	procExecuteFmt = `
			_stmt.reset_parameters();
%[1]s
	
			return StoredProcedure::execute();`

	procExecuteNoParam = `
			return StoredProcedure::execute();`

	// 1. paramIndex
	// 2. param
	procBindFmt = `
			_stmt.bind(%[1]d, %[2]s);`

	// 1. paramIndex
	// 2. param
	procBindRetFmt = `
			_stmt.bind(%[1]d, %[2]s, nanodbc::statement::PARAM_RETURN);`
)
