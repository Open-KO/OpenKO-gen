package doxygenGen

const (
	// 1: import export list
	// primaryModuleFmt is the template for the primary module file
	primaryModuleFmt = `export module doxygen_model;

%[1]s`

	// 1: partition module to export
	exportImportFmt = "export import :%[1]s;\n"

	// 1: partition file name
	// 2: file contents
	// 3: includes
	// partitionModuleFmt is the template for a module partition
	partitionModuleFmt = `module;

%[3]s
export module doxygen_model:%[1]s;

%[2]s`

	// 1. ClassName
	// 2. Property defs
	modelFileFmt = `//class %[1]sBinder;

export class %[1]s 
{
public:
//	using BinderType = %[1]sBinder;
	
};
`
	// 1: header file to include
	includeFmt = "#include %[1]s\n"
)
