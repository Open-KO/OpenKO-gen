package cppGenerator

// This was done as a prototype using jsonSchema; it would need to be refactored using
// jsonSchema/tableDef.go, if a use case for it arises.
//
//import "fmt"
//
//// FileTemplate args:
//// 1. Class Name
//// 2. Properties
//// 3. Methods
//// 4. Parent Class
//// 5. Class Description
//// 6. Includes
//// 7. Using statements
//const fileTemplateFmt string = `#pragma once
//%[6]s
//%[7]s
//// %[1]s: %[5]s
//class %[1]s %[4]s
//{
//public:
//	/* Variables */
//%[2]s
//
//	/* Helper Functions */
//%[3]s
//};`
//
//// MemberDefinition args:
//// 1. Type
//// 2. Name
//// 3. Description
//const propertyFmt string = `	// %[2]s: %[3]s
//	%[1]s %[2]s;`
//
//// HelperFunctionDefinition args:
//// 1. Return Type
//// 2. Name
//// 3. Parameters
//// 4. Function Body
//// 5. Description
//const methodFmt string = `	// %[2]s: %[5]s
//	%[1]s %[2]s(%[3]s)
//	{
//		%[4]s
//	}`
//
//const parentClassFmt string = ": public %[1]s"
//
//const fileNameFmt string = "%[1]s.h"
//
//const includeFmt string = "#include \"%[1]s\"\n"
//
//const usingFmt string = "using %[1]s;\n"
//
//type CppTemplate struct {
//	className   string
//	parentClass string
//	igenerator  []string
//	methods     []string
//	includes    map[string]bool
//	classDesc   string
//
//	// C++ specific impls
//	usingStatements map[string]bool
//}
//
///** c++ specific functions **/
//func (this *CppTemplate) AddUsing(s string) {
//	if this.usingStatements == nil {
//		this.usingStatements = make(map[string]bool)
//	}
//	this.usingStatements[s] = true
//}
//
///** Template interface impl functions **/
//func (this *CppTemplate) AddInclude(s string) {
//	if this.includes == nil {
//		this.includes = make(map[string]bool)
//	}
//	this.includes[s] = true
//}
//
//func (this *CppTemplate) SetClassName(s string) {
//	this.className = s
//}
//
//func (this *CppTemplate) SetClassDesc(s string) {
//	this.classDesc = s
//}
//
//func (this *CppTemplate) SetParentClass(s string) {
//	this.parentClass = fmt.Sprintf(parentClassFmt, s)
//}
//
//func (this *CppTemplate) AddProperty(def PropertyDefinition) {
//	this.igenerator = append(this.igenerator, fmt.Sprintf(propertyFmt, def.Type, def.Name, def.Description))
//}
//
//func (this *CppTemplate) AddMethod(def MethodDefinition) {
//	params := ""
//	for i := range def.Params {
//		params += fmt.Sprintf("%s %s", def.Params[i][0], def.Params[i][1])
//	}
//	this.methods = append(this.methods, fmt.Sprintf(methodFmt, def.ReturnType, def.Name, params, def.Body, def.Description))
//}
//
//func (this CppTemplate) Generate() (string, error) {
//	if this.className == "" {
//		return "", fmt.Errorf("Class name not set")
//	}
//
//	inclStr := ""
//	for k, _ := range this.includes {
//		inclStr += fmt.Sprintf(includeFmt, k)
//	}
//
//	usingStr := ""
//	for k, _ := range this.usingStatements {
//		usingStr += fmt.Sprintf(usingFmt, k)
//	}
//
//	propStr := ""
//	for i := range this.igenerator {
//		if i > 0 {
//			propStr += "\n\n"
//		}
//		propStr += this.igenerator[i]
//	}
//
//	methodStr := ""
//	for i := range this.methods {
//		if i > 0 {
//			methodStr += "\n\n"
//		}
//		methodStr += this.methods[i]
//	}
//
//	return fmt.Sprintf(fileTemplateFmt, this.className, propStr, methodStr, this.parentClass, this.classDesc, inclStr, usingStr), nil
//}
//
//func (this CppTemplate) GetFileName() string {
//	return fmt.Sprintf(fileNameFmt, this.className)
//}
