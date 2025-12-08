package cppGen

const (
	// 1. include guard name (uppercase relative header path)
	// 2. includes
	// 3. file contents
	// TODO: \mainpage doc?
	// primaryHeaderFmt is the template for the primary header file
	primaryHeaderFmt = `#ifndef %[1]s
#define %[1]s

#pragma once

%[2]s
%[3]s

#endif // %[1]s`

	// 1. includes
	// 2. file contents
	// TODO: \mainpage doc?
	// primarySourceFmt is the template for the primary source file
	primarySourceFmt = `%[1]s
%[2]s`

	// 1. include guard name (uppercase relative header path)
	// 2. includes
	// 3. model class forward declarations
	// 4. file contents
	// TODO: \mainpage doc?
	// primaryHeaderFmt is the template for the primary header file
	binderHeaderFmt = `#ifndef %[1]s
#define %[1]s

#pragma once

%[2]s
namespace nanodbc
{
	class result;
}

%[3]s
%[4]s

#endif // %[1]s`

	// 1. binder filename
	// 2. model filename
	// 3. file contents
	// TODO: \mainpage doc?
	// binderSourceFmt is the template for the binder source file
	binderSourceFmt = `#include "%[1]s"
#include <%[2]s>
#include <BinderUtil/BinderUtil.h>
#include <nanodbc/nanodbc.h>

%[3]s`

	// 1. Class Name
	modelFwdDeclareFmt = "\n\tclass %[1]s;"

	// 1. Class Name
	binderFwdDeclareFmt = "\n\tclass %[1]s;"

	// 1. ClassName
	// 2. Member defs
	// 3. Method defs
	// 4. Class-level Doxygen block
	// 5. binder namespace
	modelClassHeaderFmt = `
%[4]s
	class %[1]s 
	{
	/// \publicsection
	public:
		using BinderType = %[5]s::%[1]s;
%[2]s
%[3]s

	};
`

	// 1. Method implementations
	modelClassSourceFmt = `%[1]s`

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
	binderClassHeaderFmt = `
	/// \brief generated nanodbc column binder for %[3]s::%[1]s
	class %[1]s
	{
	/// \publicsection
	public:
		typedef void (*BindColumnFunction_t)(%[3]s::%[1]s& m, const nanodbc::result& result, short colIndex);

		using BindingsMapType = std::unordered_map<std::string, BindColumnFunction_t>;
%[2]s

	};
`

	// 1. method implementations
	binderClassSourceFmt = `%[1]s
`

	// 1: header file to include
	includeFmt = "#include %[1]s\n"

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
	methodDeclFmt = `
		/// \brief %[1]s
		%[2]s%[3]s%[4]s(%[5]s)%[6]s;`

	// 1. description
	// 2. return type
	// 3. class name
	// 4. function name
	// 5. params, csv
	// 6. pure
	// 7. function body
	methodImplFmt = `
	/// \brief %[1]s
	%[2]s%[3]s::%[4]s(%[5]s)%[6]s
	{%[7]s
	}`

	// 1: table name
	funcTableNameFmt = `
		static const std::string tableName = "%[1]s";
		return tableName;`

	// 1: list of column names, string wrapped and CSV
	// 2: Static const name
	// 3: return type
	funcColumnNamesFmt = `
		static %[3]s %[2]s =
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
	// 3. cast function
	funcPropBindingCastFmt = `
		%[1]s tmpValue = {};
		result.get_ref<%[1]s>(colIndex, tmpValue);
		m.%[2]s = %[3]s(tmpValue);`

	// 1. cppType to cast to
	// 2. PropertyName
	// 3. cast function
	funcOptionalPropBindingCastFmt = `
		std::optional<%[1]s> tmpValue;
		result.get_ref<std::optional<%[1]s>>(colIndex, tmpValue);

		if (tmpValue.has_value())
			m.%[2]s = %[3]s(*tmpValue);
		else
			m.%[2]s.reset();`

	// 1. Type
	constRefFmt = "const %s&"
	ptrFmt      = "%s*"
	constPtrFmt = "const %s*"
	constFmt    = "const %s"

	// 1. Class name
	// 2. Method declarations
	// 3. Class description
	// 4. additional doyxgen
	procClassHeaderFmt = `
	/// \brief %[3]s
	/// \class %[1]s
%[4]s
	class %[1]s : public detail::StoredProcedure
	{
	public:
		%[1]s();
		%[1]s(std::shared_ptr<nanodbc::connection> conn);
%[2]s
	};
`

	// 1. Class name
	// 2. Methods
	procClassImplFmt = `
	%[1]s::%[1]s()
		: StoredProcedure()
	{
	}

	%[1]s::%[1]s(std::shared_ptr<nanodbc::connection> conn) 
		: StoredProcedure(conn)
	{
	}
%[2]s
`

	// 1. proc name
	// 2. "?" list, csv for len of param
	procCallFmt        = "{CALL %[1]s(%[2]s)}"
	procCallWithRetFmt = "{? = CALL %[1]s(%[2]s)}"

	// 1. binding list
	procExecuteFmt = `
		prepare(Query());
		_stmt.reset_parameters();
%[1]s
	
		return StoredProcedure::execute();`

	procExecuteNoParam = `
		prepare(Query());
		return StoredProcedure::execute();`

	// 1. bind function (bind/bind_binary)
	// 2. paramIndex
	// 3. param
	procBindInputFmt = `
		_stmt.%[1]s(%[2]d, %[3]s);`

	// 1. bind function (bind/bind_binary)
	// 2. paramIndex
	// 3. param
	// 4. param type
	procBindFmt = `
		_stmt.%[1]s(%[2]d, %[3]s, %[4]s);`

	procDestructorWithFlushDef = `
		flush_on_destruct();`

	// 1: query
	procFuncQueryFmt = `
		static const std::string query = "%[1]s";
		return query;`
)
