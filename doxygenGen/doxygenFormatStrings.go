package doxygenGen

const (
	// 1: import export list
	// TODO: \mainpage doc?
	// primaryModuleFmt is the template for the primary module file
	primaryModuleFmt = `export module FullModel;

%[1]s`

	// 1: partition module to export
	exportImportFmt = "export import :%[1]s;\n"

	// 1: partition file name
	// 2: file contents
	// 3: includes
	// partitionModuleFmt is the template for a module partition
	partitionModuleFmt = `module;

%[3]s
export module FullModel:%[1]s;

%[2]s`

	// 1. ClassName
	// 2. Member defs
	// 3. Method defs
	// 4. Class-level doxygen
	modelFileFmt = `namespace model
{
	class %[1]sBinder;
	
%[4]s
	export class %[1]s 
	{
	/// \publicsection
	public:
		using BinderType = %[1]sBinder;
%[2]s
%[3]s

	};
}
`

	// 1. ClassName
	// 2. Method defs
	binderFileFmt = `namespace model
{
	/// \brief generated column binder for the %[1]s model, using nanodbc
	export class %[1]sBinder
	{
	/// \publicsection
	public:
		typedef void (*BindColumnFunction_t)(%[1]s& m, const nanodbc::result& result, short colIndex);

		using BindingsMapType = std::unordered_map<std::string, BindColumnFunction_t>;
%[2]s

	};
}
`

	// 1: header file to include
	includeFmt = "#include %[1]s\n"

	// 1: doxygen comment block
	// 2: cppType
	// 3: PropertyName
	// 4: initialized value
	// 5: associated enum
	memberFmt = `
		%[1]s
		%[2]s %[3]s%[4]s;%[5]s`

	// 1. enumName
	// 2. Value list
	// 3. Column Name
	enumFmt = `
	
		/// \enum %[1]s
		/// \brief Known valid values for %[3]s
		enum class %[1]s
		{
%[2]s
		}`

	// 1. description
	// 2. modifiers (static, inline, etc)
	// 3. return type
	// 4. function name
	// 5. params, csv
	// 6. function body
	methodFmt = `
		/// \brief %[1]s
		%[2]s%[3]s %[4]s(%[5]s)
		{%[6]s
		}`

	// 1: table name
	funcTableNameFmt = `
			static const std::string tableName = "%[1]s";
			return tableName;`

	// 1: list of column names, string wrapped and CSV
	funcColumnNamesFmt = `
			static const std::unordered_set<std::string> columnNames =
			{
				%[1]s
			};
			return columnNames;`

	funcDbTypeFmt = `
			static const std::string dbType = "%[1]s";
			return dbType;`

	// 1 Binding map entries
	funcColumnBindingsFmt = `
			static const BindingsMapType bindingsMap =
			{%[1]s
			};
			return bindingsMap;`

	// 1. Column Name
	// 2. Class Name
	// 3. Property Name
	bindingFmt = `
				{"%[1]s", &%[2]sBinder::Bind%[3]s}`

	// 1. cppType
	// 2. PropertyName
	funcPropBindingFmt = `
			result.get_ref<%[1]s>(colIndex, m.%[2]s);`
)
