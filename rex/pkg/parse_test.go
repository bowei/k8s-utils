package pkg

import (
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
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

	result := processEnum(&typeInfo, ident, statusType, docPkg, "test")

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

	result := processEnum(&typeInfo, ident, statusType, docPkg, "test")

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

	result := processEnum(&typeInfo, ident, statusType, docPkg, "test")

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

	result := processEnum(&typeInfo2, ident2, emptyType, docPkg2, "test")

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

	enumValues := findConstantsByType(docPkg, "MyType", "testpkg")

	expected := []EnumInfo{
		{Name: "Val1", DocString: "Doc for Val1"},
		{Name: "Val2", DocString: ""},
		{Name: "Val4", DocString: "Doc for Val4"},
		{Name: "Val5", DocString: ""},
	}

	if len(enumValues) != len(expected) {
		t.Fatalf("expected %d enum values, got %d", len(expected), len(enumValues))
	}

	// Note: The order of constants is not guaranteed, so we check for existence.
	expectedMap := make(map[string]string)
	for _, e := range expected {
		expectedMap[e.Name] = e.DocString
	}

	for _, enum := range enumValues {
		doc, ok := expectedMap[enum.Name]
		if !ok {
			t.Errorf("unexpected enum value: %s", enum.Name)
			continue
		}
		if enum.DocString != doc {
			t.Errorf("for enum %s, expected doc string '%s', got '%s'", enum.Name, doc, enum.DocString)
		}
		delete(expectedMap, enum.Name)
	}

	if len(expectedMap) > 0 {
		for name := range expectedMap {
			t.Errorf("missing expected enum value: %s", name)
		}
	}
}
