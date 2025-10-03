package pkg

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestProcessEnum_Constants(t *testing.T) {
	// Test enum detection with constants only
	src := `
package test

// Status represents the status of something
type Status string

const (
	// StatusActive means it's active
	StatusActive Status = "active"
	// StatusInactive means it's inactive
	StatusInactive Status = "inactive"
)
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	pkg := &ast.Package{
		Name:  "test",
		Files: map[string]*ast.File{"test.go": f},
	}

	docPkg := doc.New(pkg, "test", 0)

	// Find the Status type
	var statusType *doc.Type
	for _, t := range docPkg.Types {
		if t.Name == "Status" {
			statusType = t
			break
		}
	}

	if statusType == nil {
		t.Fatal("Status type not found")
	}

	typeSpec := statusType.Decl.Specs[0].(*ast.TypeSpec)
	ident := typeSpec.Type.(*ast.Ident)

	typeInfo := TypeInfo{
		Package:  "test",
		TypeName: "Status",
	}

	result := processEnum(&typeInfo, ident, docPkg)

	if !result {
		t.Error("processEnum should return true for valid enum")
	}

	expectedEnums := []EnumInfo{
		{
			Name:      "StatusActive",
			DocString: "StatusActive means it's active",
		},
		{
			Name:      "StatusInactive",
			DocString: "StatusInactive means it's inactive",
		},
	}

	if len(typeInfo.EnumValues) != len(expectedEnums) {
		t.Errorf("expected %d enum values, got %d", len(expectedEnums), len(typeInfo.EnumValues))
	}

	for i, expected := range expectedEnums {
		if i >= len(typeInfo.EnumValues) {
			break
		}
		actual := typeInfo.EnumValues[i]
		if actual.Name != expected.Name {
			t.Errorf("enum[%d].Name: expected %s, got %s", i, expected.Name, actual.Name)
		}
		if actual.DocString != expected.DocString {
			t.Errorf("enum[%d].DocString: expected %s, got %s", i, expected.DocString, actual.DocString)
		}
	}
}

func TestProcessEnum_Variables(t *testing.T) {
	// Test enum detection with variables only
	src := `
package test

// Status represents the status of something
type Status string

var (
	// StatusActive means it's active
	StatusActive Status = "active"
	// StatusInactive means it's inactive
	StatusInactive Status = "inactive"
)
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	pkg := &ast.Package{
		Name:  "test",
		Files: map[string]*ast.File{"test.go": f},
	}

	docPkg := doc.New(pkg, "test", 0)

	// Find the Status type
	var statusType *doc.Type
	for _, t := range docPkg.Types {
		if t.Name == "Status" {
			statusType = t
			break
		}
	}

	if statusType == nil {
		t.Fatal("Status type not found")
	}

	typeSpec := statusType.Decl.Specs[0].(*ast.TypeSpec)
	ident := typeSpec.Type.(*ast.Ident)

	typeInfo := TypeInfo{
		Package:  "test",
		TypeName: "Status",
	}

	result := processEnum(&typeInfo, ident, docPkg)

	if !result {
		t.Error("processEnum should return true for valid enum with variables")
	}

	expectedEnums := []EnumInfo{
		{
			Name:      "StatusActive",
			DocString: "StatusActive means it's active",
		},
		{
			Name:      "StatusInactive",
			DocString: "StatusInactive means it's inactive",
		},
	}

	if len(typeInfo.EnumValues) != len(expectedEnums) {
		t.Errorf("expected %d enum values, got %d", len(expectedEnums), len(typeInfo.EnumValues))
	}

	for i, expected := range expectedEnums {
		if i >= len(typeInfo.EnumValues) {
			break
		}
		actual := typeInfo.EnumValues[i]
		if actual.Name != expected.Name {
			t.Errorf("enum[%d].Name: expected %s, got %s", i, expected.Name, actual.Name)
		}
		if actual.DocString != expected.DocString {
			t.Errorf("enum[%d].DocString: expected %s, got %s", i, expected.DocString, actual.DocString)
		}
	}
}

func TestProcessEnum_Mixed(t *testing.T) {
	// Test enum detection with both constants and variables
	src := `
package test

// Status represents the status of something
type Status string

const (
	// StatusActive means it's active
	StatusActive Status = "active"
)

var (
	// StatusInactive means it's inactive
	StatusInactive Status = "inactive"
)
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	pkg := &ast.Package{
		Name:  "test",
		Files: map[string]*ast.File{"test.go": f},
	}

	docPkg := doc.New(pkg, "test", 0)

	// Find the Status type
	var statusType *doc.Type
	for _, t := range docPkg.Types {
		if t.Name == "Status" {
			statusType = t
			break
		}
	}

	if statusType == nil {
		t.Fatal("Status type not found")
	}

	typeSpec := statusType.Decl.Specs[0].(*ast.TypeSpec)
	ident := typeSpec.Type.(*ast.Ident)

	typeInfo := TypeInfo{
		Package:  "test",
		TypeName: "Status",
	}

	result := processEnum(&typeInfo, ident, docPkg)

	if !result {
		t.Error("processEnum should return true for valid enum with mixed constants and variables")
	}

	// Should have both constant and variable enum values
	if len(typeInfo.EnumValues) != 2 {
		t.Errorf("expected 2 enum values, got %d", len(typeInfo.EnumValues))
	}

	// Check that we have both values (order may vary)
	foundActive := false
	foundInactive := false
	for _, enumVal := range typeInfo.EnumValues {
		if enumVal.Name == "StatusActive" {
			foundActive = true
		}
		if enumVal.Name == "StatusInactive" {
			foundInactive = true
		}
	}

	if !foundActive {
		t.Error("StatusActive enum value not found")
	}
	if !foundInactive {
		t.Error("StatusInactive enum value not found")
	}
}

func TestProcessEnum_NotEnum(t *testing.T) {
	// Test that non-enum types are not detected as enums
	src := `
package test

// MyStruct is just a struct
type MyStruct struct {
	Field string
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	pkg := &ast.Package{
		Name:  "test",
		Files: map[string]*ast.File{"test.go": f},
	}

	docPkg := doc.New(pkg, "test", 0)

	// Find the MyStruct type
	var structType *doc.Type
	for _, t := range docPkg.Types {
		if t.Name == "MyStruct" {
			structType = t
			break
		}
	}

	if structType == nil {
		t.Fatal("MyStruct type not found")
	}

	typeSpec := structType.Decl.Specs[0].(*ast.TypeSpec)
	// This will be a *ast.StructType, not an *ast.Ident
	if _, ok := typeSpec.Type.(*ast.Ident); ok {
		t.Fatal("expected struct type, got ident")
	}

	// processEnum expects an ident, so this test shows it won't be called for structs
	// But let's test with a string type that has no constants/variables
	src2 := `
package test

// EmptyStatus has no enum values
type EmptyStatus string
`

	f2, err := parser.ParseFile(fset, "test2.go", src2, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	pkg2 := &ast.Package{
		Name:  "test",
		Files: map[string]*ast.File{"test2.go": f2},
	}

	docPkg2 := doc.New(pkg2, "test", 0)

	var emptyType *doc.Type
	for _, t := range docPkg2.Types {
		if t.Name == "EmptyStatus" {
			emptyType = t
			break
		}
	}

	if emptyType == nil {
		t.Fatal("EmptyStatus type not found")
	}

	typeSpec2 := emptyType.Decl.Specs[0].(*ast.TypeSpec)
	ident2 := typeSpec2.Type.(*ast.Ident)

	typeInfo2 := TypeInfo{
		Package:  "test",
		TypeName: "EmptyStatus",
	}

	result := processEnum(&typeInfo2, ident2, docPkg2)

	if result {
		t.Error("processEnum should return false for type with no constants or variables")
	}

	if len(typeInfo2.EnumValues) != 0 {
		t.Errorf("expected 0 enum values, got %d", len(typeInfo2.EnumValues))
	}
}

func TestFindConstantsByType(t *testing.T) {
	src := `
package testpkg

// MyType is a test type.
type MyType string

const (
	// Doc for Val1
	Val1 MyType = "Value1"  // explicit type
	Val2 = MyType("Value2") // other way to declare the const.
	// unexported
	val3 MyType = "Value3"
)

const (
	// Doc for Val4
	Val4 = MyType("Value4") // type conversion
	Val5 = MyType("Value5")
)

const (
	OtherVal = "other"
)
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	pkg := &ast.Package{
		Name:  "testpkg",
		Files: map[string]*ast.File{"test.go": f},
	}

	docPkg := doc.New(pkg, "testpkg", 0)

	got := findConstantsByType(docPkg, "MyType")

	sort.Slice(got, func(i, j int) bool {
		return got[i].Name < got[j].Name
	})
	expected := []EnumInfo{
		{
			Name:      "Val1",
			DocString: "Doc for Val1",
			ParsedDocString: GoDocString{
				Elements: []GoDocElem{
					{Type: "p", Content: []string{"Doc for Val1"}},
				},
			},
		},
		{Name: "Val2", DocString: ""},
		{
			Name:      "Val4",
			DocString: "Doc for Val4",
			ParsedDocString: GoDocString{
				Elements: []GoDocElem{
					{Type: "p", Content: []string{"Doc for Val4"}},
				},
			},
		},
		{Name: "Val5", DocString: ""},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

func TestProcessStruct(t *testing.T) {
	src := `
package test

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MyStruct is a test struct.
type MyStruct struct {
	// Field1 doc string
	Field1 string
	// Field2 is an integer pointer
	Field2 *int
	// Field3 is a list of strings
	Field3 []string
	// Field4 is a map
	Field4 map[string]bool
	// Field5 is a nested struct
	Field5 AnotherStruct
	// Field6 is from an external package
	Field6 v1.Time
	// Embedded field
	AnotherStruct
	unexportedField string
}

// AnotherStruct is used in MyStruct
type AnotherStruct struct {
	NestedField string
}

// K8sResource is a root type
type K8sResource struct {
	v1.TypeMeta
	v1.ObjectMeta
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	pkg := &ast.Package{
		Name:  "test",
		Files: map[string]*ast.File{"test.go": f},
	}

	docPkg := doc.New(pkg, "test", 0)

	// --- Test MyStruct ---
	var myStructType *doc.Type
	for _, typ := range docPkg.Types {
		if typ.Name == "MyStruct" {
			myStructType = typ
			break
		}
	}
	if myStructType == nil {
		t.Fatal("MyStruct type not found")
	}

	typeSpec := myStructType.Decl.Specs[0].(*ast.TypeSpec)
	structType := typeSpec.Type.(*ast.StructType)
	externalPkgs := make(map[string]bool)

	typeInfo := TypeInfo{
		Package:  "test",
		TypeName: "MyStruct",
	}

	err = processStruct(&typeInfo, typeSpec, structType, pkg, "test", externalPkgs)
	if err != nil {
		t.Fatalf("processStruct failed: %v", err)
	}

	expectedFields := []FieldInfo{
		{
			FieldName: "Field1",
			TypeName:  "string",
			Package:   "",
			DocString: "Field1 doc string",
			ParsedDocString: GoDocString{
				Elements: []GoDocElem{
					{Type: "p", Content: []string{"Field1 doc string"}},
				},
			},
		},
		{
			FieldName:      "Field2",
			TypeName:       "int",
			Package:        "",
			TypeDecorators: []string{"Ptr"},
			DocString:      "Field2 is an integer pointer",
			ParsedDocString: GoDocString{
				Elements: []GoDocElem{
					{Type: "p", Content: []string{"Field2 is an integer pointer"}},
				},
			},
		},
		{
			FieldName:      "Field3",
			TypeName:       "string",
			Package:        "",
			TypeDecorators: []string{"List"},
			DocString:      "Field3 is a list of strings",
			ParsedDocString: GoDocString{
				Elements: []GoDocElem{
					{Type: "p", Content: []string{"Field3 is a list of strings"}},
				},
			},
		},
		{
			FieldName:      "Field4",
			TypeName:       "bool",
			Package:        "",
			TypeDecorators: []string{"Map[string]"},
			DocString:      "Field4 is a map",
			ParsedDocString: GoDocString{
				Elements: []GoDocElem{
					{Type: "p", Content: []string{"Field4 is a map"}},
				},
			},
		},
		{
			FieldName: "Field5",
			TypeName:  "test.AnotherStruct",
			Package:   "test",
			DocString: "Field5 is a nested struct",
			ParsedDocString: GoDocString{
				Elements: []GoDocElem{
					{Type: "p", Content: []string{"Field5 is a nested struct"}},
				},
			},
		},
		{
			FieldName: "Field6",
			TypeName:  "k8s.io/apimachinery/pkg/apis/meta/v1.Time",
			Package:   "k8s.io/apimachinery/pkg/apis/meta/v1",
			DocString: "Field6 is from an external package",
			ParsedDocString: GoDocString{
				Elements: []GoDocElem{
					{Type: "p", Content: []string{"Field6 is from an external package"}},
				},
			},
		},
		{
			FieldName: "AnotherStruct",
			TypeName:  "test.AnotherStruct",
			Package:   "test",
			DocString: "Embedded field",
			ParsedDocString: GoDocString{
				Elements: []GoDocElem{
					{Type: "p", Content: []string{"Embedded field"}},
				},
			},
		},
	}

	if !reflect.DeepEqual(typeInfo.Fields, expectedFields) {
		t.Errorf("For MyStruct, got fields %#v, want %#v", typeInfo.Fields, expectedFields)
	}

	expectedExternalPkgs := map[string]bool{
		"k8s.io/apimachinery/pkg/apis/meta/v1": true,
	}
	if !reflect.DeepEqual(externalPkgs, expectedExternalPkgs) {
		t.Errorf("For MyStruct, got external packages %#v, want %#v", externalPkgs, expectedExternalPkgs)
	}

	if typeInfo.IsRoot {
		t.Error("MyStruct should not be a root type")
	}

	// --- Test K8sResource ---
	var k8sResourceType *doc.Type
	for _, typ := range docPkg.Types {
		if typ.Name == "K8sResource" {
			k8sResourceType = typ
			break
		}
	}
	if k8sResourceType == nil {
		t.Fatal("K8sResource type not found")
	}

	typeSpec = k8sResourceType.Decl.Specs[0].(*ast.TypeSpec)
	structType = typeSpec.Type.(*ast.StructType)
	externalPkgs = make(map[string]bool)

	typeInfo = TypeInfo{
		Package:  "test",
		TypeName: "K8sResource",
	}

	err = processStruct(&typeInfo, typeSpec, structType, pkg, "test", externalPkgs)
	if err != nil {
		t.Fatalf("processStruct for K8sResource failed: %v", err)
	}

	// Simplified check for root type, as field processing is tested above
	if !typeInfo.IsRoot {
		t.Error("K8sResource should be a root type")
	}
}

func TestParsePackage(t *testing.T) {
	// Create a temporary directory for the test package
	tempDir, err := os.MkdirTemp("", "testpkg")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a subdirectory for the package
	pkgDir := filepath.Join(tempDir, "testpkg")
	err = os.Mkdir(pkgDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create pkg dir: %v", err)
	}

	// Create Go source files in the temporary directory
	mainGo := `
package testpkg

import "fmt"

// MyStruct is a test struct.
type MyStruct struct {
	Field1 string
	Field2 AnotherType
}

// MyEnum is a test enum.
type MyEnum string

const (
	EnumVal1 MyEnum = "val1"
)
`
	typesGo := `
package testpkg

// AnotherType is defined in a separate file.
type AnotherType struct {
	NestedField bool
}
`
	mainTestGo := `
package testpkg

import "testing"

// This type should be ignored
type TestStruct struct {
	Field string
}
`
	goMod := `
module example.com/testpkg
go 1.18
`

	err = os.WriteFile(filepath.Join(pkgDir, "main.go"), []byte(mainGo), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}
	err = os.WriteFile(filepath.Join(pkgDir, "types.go"), []byte(typesGo), 0644)
	if err != nil {
		t.Fatalf("Failed to write types.go: %v", err)
	}
	err = os.WriteFile(filepath.Join(pkgDir, "main_test.go"), []byte(mainTestGo), 0644)
	if err != nil {
		t.Fatalf("Failed to write main_test.go: %v", err)
	}
	err = os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	allTypes := make(map[string]TypeInfo)
	externalPkgs, err := parsePackage(pkgDir, allTypes)
	if err != nil {
		t.Fatalf("parsePackage failed: %v", err)
	}

	// 1. Check that all expected types were processed
	if _, ok := allTypes["example.com/testpkg.MyStruct"]; !ok {
		t.Error("MyStruct was not processed")
	}
	if _, ok := allTypes["example.com/testpkg.MyEnum"]; !ok {
		t.Error("MyEnum was not processed")
	}
	if _, ok := allTypes["example.com/testpkg.AnotherType"]; !ok {
		t.Error("AnotherType was not processed")
	}

	// 2. Check that test types were skipped
	if _, ok := allTypes["example.com/testpkg.TestStruct"]; ok {
		t.Error("TestStruct from test file should not have been processed")
	}

	if len(allTypes) != 3 {
		t.Errorf("Expected 3 processed types, got %d", len(allTypes))
	}

	// 3. Check that external packages were identified
	// "fmt" is a standard library package and should be ignored by the caller,
	// but parsePackage should still report it.
	if _, ok := externalPkgs["fmt"]; !ok {
		t.Error("Expected 'fmt' in external packages")
	}
}

func TestProcessType(t *testing.T) {
	src := `
package test

// MyStruct is a test struct.
type MyStruct struct {
	Field1 string
}

// MyEnum is a test enum.
type MyEnum string

const (
	// EnumVal1 is a value.
	EnumVal1 MyEnum = "val1"
)

// unexportedType should be skipped.
type unexportedType struct{}

// SimpleAlias is not a struct or enum.
type SimpleAlias int
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse source: %v", err)
	}

	pkg := &ast.Package{
		Name:  "test",
		Files: map[string]*ast.File{"test.go": f},
	}

	docPkg := doc.New(pkg, "test", 0)

	allTypes := make(map[string]TypeInfo)
	externalPkgs := make(map[string]bool)
	pkgImportPath := "example.com/test"

	for _, typ := range docPkg.Types {
		processType(typ, pkgImportPath, allTypes, pkg, externalPkgs, docPkg)
	}

	// 1. Check that struct and enum were processed
	if _, ok := allTypes["example.com/test.MyStruct"]; !ok {
		t.Error("MyStruct was not processed")
	} else {
		if len(allTypes["example.com/test.MyStruct"].Fields) != 1 {
			t.Errorf("Expected 1 field for MyStruct, got %d", len(allTypes["example.com/test.MyStruct"].Fields))
		}
	}

	if _, ok := allTypes["example.com/test.MyEnum"]; !ok {
		t.Error("MyEnum was not processed")
	} else {
		if len(allTypes["example.com/test.MyEnum"].EnumValues) != 1 {
			t.Errorf("Expected 1 enum value for MyEnum, got %d", len(allTypes["example.com/test.MyEnum"].EnumValues))
		}
	}

	// 2. Check that unexported type and simple alias were skipped
	if _, ok := allTypes["example.com/test.unexportedType"]; ok {
		t.Error("unexportedType should not have been processed")
	}
	if _, ok := allTypes["example.com/test.SimpleAlias"]; ok {
		t.Error("SimpleAlias should not have been processed")
	}

	if len(allTypes) != 2 {
		t.Errorf("Expected 2 processed types, got %d", len(allTypes))
	}

	// 3. Check that reprocessing is skipped
	// Find MyStruct again and process it
	var myStructType *doc.Type
	for _, typ := range docPkg.Types {
		if typ.Name == "MyStruct" {
			myStructType = typ
			break
		}
	}
	if myStructType == nil {
		t.Fatal("Could not find MyStruct type to test reprocessing")
	}

	// Modify the existing entry to see if it gets overwritten
	originalTypeInfo := allTypes["example.com/test.MyStruct"]
	modifiedTypeInfo := originalTypeInfo
	modifiedTypeInfo.DocString = "modified"
	allTypes["example.com/test.MyStruct"] = modifiedTypeInfo

	processType(myStructType, pkgImportPath, allTypes, pkg, externalPkgs, docPkg)

	if allTypes["example.com/test.MyStruct"].DocString != "modified" {
		t.Error("processType should have skipped reprocessing an existing type")
	}
}

func TestParsePackages(t *testing.T) {
	// 1. Create temp directory structure
	tempDir, err := os.MkdirTemp("", "test-parse-packages")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pkg1Dir := filepath.Join(tempDir, "pkg1")
	pkg2Dir := filepath.Join(tempDir, "pkg2")
	os.Mkdir(pkg1Dir, 0755)
	os.Mkdir(pkg2Dir, 0755)

	// 2. Create files for pkg1
	pkg1GoMod := `
module example.com/pkg1
go 1.18
`
	pkg1GoFile := `
package pkg1

// Pkg1Struct is a struct in pkg1.
type Pkg1Struct struct {
    Field string
}
`
	if err := os.WriteFile(filepath.Join(pkg1Dir, "go.mod"), []byte(pkg1GoMod), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkg1Dir, "main.go"), []byte(pkg1GoFile), 0644); err != nil {
		t.Fatal(err)
	}

	// 3. Create files for pkg2
	pkg2GoMod := `
module example.com/pkg2
go 1.18

replace example.com/pkg1 => ../pkg1
`
	pkg2GoFile := `
package pkg2

import "example.com/pkg1"

// Pkg2Struct is a struct in pkg2.
type Pkg2Struct struct {
    Other pkg1.Pkg1Struct
}
`
	if err := os.WriteFile(filepath.Join(pkg2Dir, "go.mod"), []byte(pkg2GoMod), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkg2Dir, "main.go"), []byte(pkg2GoFile), 0644); err != nil {
		t.Fatal(err)
	}

	// 4. Call ParsePackages
	allTypes, err := ParsePackages([]string{pkg1Dir, pkg2Dir})
	if err != nil {
		t.Fatalf("ParsePackages failed: %v", err)
	}

	// 5. Assertions
	if len(allTypes) != 2 {
		t.Errorf("Expected 2 types to be parsed, got %d", len(allTypes))
	}

	if _, ok := allTypes["example.com/pkg1.Pkg1Struct"]; !ok {
		t.Error("Pkg1Struct was not processed")
	}
	if _, ok := allTypes["example.com/pkg2.Pkg2Struct"]; !ok {
		t.Error("Pkg2Struct was not processed")
	}

	// Check field type in Pkg2Struct
	pkg2Struct := allTypes["example.com/pkg2.Pkg2Struct"]
	if len(pkg2Struct.Fields) != 1 {
		t.Fatalf("Expected 1 field in Pkg2Struct, got %d", len(pkg2Struct.Fields))
	}
	field := pkg2Struct.Fields[0]
	expectedTypeName := "example.com/pkg1.Pkg1Struct"
	if field.TypeName != expectedTypeName {
		t.Errorf("Expected field type %s, got %s", expectedTypeName, field.TypeName)
	}
}