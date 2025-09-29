package pkg

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

// ParsePackages from the files in the given directories.
func ParsePackages(pkgDirs []string) (map[string]TypeInfo, error) {
	allTypes := make(map[string]TypeInfo)
	parsedPkgs := make(map[string]bool)

	var errs []error
	queue := []string{}

	for _, pkgDir := range pkgDirs {
		absPath, err := filepath.Abs(pkgDir)
		if err != nil {
			err2 := fmt.Errorf("error getting file path for %s (error was %w). Possible fix: `go get %s`",
				pkgDir, err, pkgDir)
			log.Printf("%v", err2)
			errs = append(errs, err2)
			continue
		}

		pkgImportPath, err := getPkgPathFromDir(absPath)
		if err != nil {
			log.Printf("Skipping directory %s, not a Go package: %v", absPath, err)
			continue
		}

		if _, parsed := parsedPkgs[pkgImportPath]; !parsed {
			queue = append(queue, absPath)
			parsedPkgs[pkgImportPath] = true
		}
	}

	for len(queue) > 0 {
		pkgDir := queue[0]
		queue = queue[1:]

		externalPkgs, err := parsePackage(pkgDir, allTypes)
		if err != nil {
			log.Printf("Error parsing package %s: %v", pkgDir, err)
			continue
		}

		for pkgPath := range externalPkgs {
			// As per design, ignore packages in the Go
			// standard library. A common heuristic is
			// that standard library packages do not have
			// a dot in their first path component.
			firstPart := strings.Split(pkgPath, "/")[0]
			if !strings.Contains(firstPart, ".") {
				continue
			}

			if _, parsed := parsedPkgs[pkgPath]; !parsed {
				dir, err := resolvePkgDir(pkgPath)
				if err != nil {
					err2 := fmt.Errorf("error getting file path for %s (error was %w). Possible fix: `go get %s`",
						pkgPath, err, pkgPath)
					log.Printf("%v", err2)
					errs = append(errs, err2)
					continue
				}
				queue = append(queue, dir)
				parsedPkgs[pkgPath] = true
			}
		}
	}

	if len(errs) > 0 {
		log.Printf("Errors:")
		for _, e := range errs {
			log.Printf("- %v", e)
		}
	}

	return allTypes, nil
}

// getPkgPathFromDir uses `go list` to find the import path of a package in a directory.
func getPkgPathFromDir(pkgDir string) (string, error) {
	args := []string{"list", "-f", "{{.ImportPath}}", "."}
	log.Printf("getPkgPathFromDir %v", args)

	cmd := exec.Command("go", args...)
	cmd.Dir = pkgDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not run 'go list' in dir %s: %w", pkgDir, err)
	}

	ret := strings.TrimSpace(string(out))
	log.Printf("getPkgPathFromDir %v = %q", args, ret)

	return ret, nil
}

// resolvePkgDir uses `go list` to find the directory of a package.
func resolvePkgDir(pkgPath string) (string, error) {
	args := []string{"list", "-f", "{{.Dir}}", pkgPath}
	log.Printf("resolvePkgDir %v", args)

	cmd := exec.Command("go", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not run 'go list' for package %s: %w", pkgPath, err)
	}

	ret := strings.TrimSpace(string(out))
	log.Printf("resolvePkgDir %v = %q", args, ret)

	return ret, nil
}

// resolveType resolves an ast.Expr to a type name and decorators.
func resolveType(
	expr ast.Expr,
	pkgImportPath string,
	importMap map[string]string,
	externalPkgs map[string]bool,
) (string, []string) {
	return resolveTypeRecursive(expr, pkgImportPath, importMap, externalPkgs, 0)
}

// resolveTypeRecursive is the recursive implementation for resolveType with logging depth.
func resolveTypeRecursive(
	expr ast.Expr,
	pkgImportPath string,
	importMap map[string]string,
	externalPkgs map[string]bool,
	depth int,
) (string, []string) {
	switch t := expr.(type) {
	case *ast.StarExpr:
		log.Printf("resolving StarExpr (pointer)")
		baseType, decorators := resolveTypeRecursive(t.X, pkgImportPath, importMap, externalPkgs, depth+1)
		return baseType, append(decorators, "Ptr")
	case *ast.ArrayType:
		log.Printf("resolving ArrayType (list/slice)")
		baseType, decorators := resolveTypeRecursive(t.Elt, pkgImportPath, importMap, externalPkgs, depth+1)
		return baseType, append(decorators, "List")
	case *ast.MapType:
		log.Printf("resolving MapType")
		keyType, _ := resolveTypeRecursive(t.Key, pkgImportPath, importMap, externalPkgs, depth+1)
		valueType, decorators := resolveTypeRecursive(t.Value, pkgImportPath, importMap, externalPkgs, depth+1)
		return valueType, append(decorators, fmt.Sprintf("Map[%s]", keyType))
	case *ast.Ident:
		if t.Obj == nil {
			log.Printf("resolving Ident: built-in type %s", t.Name)
			return t.Name, nil // Built-in type
		}
		log.Printf("resolving Ident: same-package type %s", t.Name)
		return pkgImportPath + "." + t.Name, nil // Type in the same package
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			pkgName := x.Name
			if pkgPath, ok := importMap[pkgName]; ok {
				externalPkgs[pkgPath] = true
				log.Printf("resolving SelectorExpr: external type %s.%s (from %s)", pkgName, t.Sel.Name, pkgPath)
				return pkgPath + "." + t.Sel.Name, nil
			}
			log.Printf("resolving SelectorExpr: unknown type %s.%s", pkgName, t.Sel.Name)
			return pkgName + "." + t.Sel.Name, nil
		}
	}
	log.Printf("resolving unknown type %T", expr)
	return "", nil
}

func processStruct(
	typeInfo *TypeInfo,
	typeSpec *ast.TypeSpec,
	structType *ast.StructType,
	pkg *ast.Package,
	pkgImportPath string,
	externalPkgs map[string]bool,
) error {
	log.Printf("processing struct %s", typeInfo.TypeName)
	// Find the file that contains this type spec
	var file *ast.File
	for _, f := range pkg.Files {
		if f.Pos() <= typeSpec.Pos() && typeSpec.Pos() < f.End() {
			file = f
			break
		}
	}
	if file == nil {
		return fmt.Errorf("file not found for type %s", typeInfo.TypeName)
	}

	importMap := make(map[string]string)
	for _, i := range file.Imports {
		path := strings.Trim(i.Path.Value, `"`)
		if i.Name != nil {
			importMap[i.Name.Name] = path
		} else {
			parts := strings.Split(path, "/")
			importMap[parts[len(parts)-1]] = path
		}
	}

	if structType.Fields != nil {
		for _, field := range structType.Fields.List {
			if len(field.Names) > 0 {
				var fieldNamesForLog []string
				for _, name := range field.Names {
					fieldNamesForLog = append(fieldNamesForLog, name.Name)
				}
				log.Printf("processing field(s): %s", strings.Join(fieldNamesForLog, ", "))
			} else {
				log.Printf("processing embedded field")
			}

			fieldType, decorators := resolveType(field.Type, pkgImportPath, importMap, externalPkgs)
			if fieldType == "" {
				log.Printf("skipping field, could not resolve type")
				continue
			}

			// Reverse decorators to get the correct order (e.g., Ptr to List)
			for i, j := 0, len(decorators)-1; i < j; i, j = i+1, j-1 {
				decorators[i], decorators[j] = decorators[j], decorators[i]
			}

			fieldDoc := ""
			if field.Doc != nil {
				fieldDoc = strings.TrimSpace(field.Doc.Text())
			}

			if len(field.Names) > 0 {
				for _, name := range field.Names {
					if !ast.IsExported(name.Name) {
						log.Printf("skipping unexported field: %s", name.Name)
						continue
					}
					log.Printf("            found exported field: %s %s", name.Name, fieldType)
					fieldInfo := FieldInfo{
						FieldName:       name.Name,
						TypeName:        fieldType,
						TypeDecorators:  decorators,
						DocString:       fieldDoc,
						ParsedDocString: *parseGoDocString(fieldDoc),
					}
					typeInfo.Fields = append(typeInfo.Fields, fieldInfo)
				}
			} else { // Embedded field
				log.Printf("found embedded field of type: %s", fieldType)
				parts := strings.Split(fieldType, ".")
				fieldName := parts[len(parts)-1]
				fieldInfo := FieldInfo{
					FieldName:       fieldName,
					TypeName:        fieldType,
					TypeDecorators:  decorators,
					DocString:       fieldDoc,
					ParsedDocString: *parseGoDocString(fieldDoc),
				}
				typeInfo.Fields = append(typeInfo.Fields, fieldInfo)
			}
		}
	}

	// Check if this is a root type (e.g. k8s resource or compute RPC)
	// TODO: split this mode and make it a command line flag.
	hasTypeMeta := false
	hasObjectMeta := false

	for _, field := range typeInfo.Fields {
		if field.FieldName == "TypeMeta" {
			hasTypeMeta = true
		}
		if field.FieldName == "ObjectMeta" {
			hasObjectMeta = true
		}
	}

	requestSuffix := strings.HasSuffix(typeInfo.TypeName, "Request")
	responseSuffix := strings.HasSuffix(typeInfo.TypeName, "Response")

	if hasTypeMeta && hasObjectMeta || requestSuffix || responseSuffix {
		typeInfo.IsRoot = true
	}
	return nil
}

// findConstantsByType finds constants that have an explicit type matching the target type
func findConstantsByType(docPkg *doc.Package, targetTypeName string, pkgImportPath string) []EnumInfo {
	var enumValues []EnumInfo

	for _, c := range docPkg.Consts {
		if c.Decl != nil {
			for _, spec := range c.Decl.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					// Check if this constant has an explicit type that matches our target
					if vs.Type != nil {
						if ident, ok := vs.Type.(*ast.Ident); ok && ident.Name == targetTypeName {
							// Found a constant with explicit type matching our target
							for _, name := range vs.Names {
								if ast.IsExported(name.Name) {
									docString := ""
									if vs.Doc != nil {
										docString = vs.Doc.Text()
									}
									log.Printf("found exported enum const value with explicit type: %s", name.Name)
									enumValues = append(enumValues, EnumInfo{
										Name:            name.Name,
										DocString:       strings.TrimSpace(docString),
										ParsedDocString: *parseGoDocString(docString),
									})
								}
							}
						}
					} else if vs.Values != nil {
						// Check if the value is a type conversion like MyType("value")
						for i, name := range vs.Names {
							if ast.IsExported(name.Name) && i < len(vs.Values) {
								if callExpr, ok := vs.Values[i].(*ast.CallExpr); ok {
									if ident, ok := callExpr.Fun.(*ast.Ident); ok && ident.Name == targetTypeName {
										docString := ""
										if vs.Doc != nil {
											docString = vs.Doc.Text()
										}
										log.Printf("found exported enum const value with type conversion: %s", name.Name)
										enumValues = append(enumValues, EnumInfo{
											Name:            name.Name,
											DocString:       strings.TrimSpace(docString),
											ParsedDocString: *parseGoDocString(docString),
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return enumValues
}

func processEnum(typeInfo *TypeInfo, ident *ast.Ident, t *doc.Type, docPkg *doc.Package, pkgImportPath string) bool {
	// Handle potential enums
	isEnum := false

	// First check for constants with explicit types
	explicitlyTypedConsts := findConstantsByType(docPkg, typeInfo.TypeName, pkgImportPath)

	switch ident.Name {
	case "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64", "byte", "rune":
		if len(t.Consts) > 0 || len(t.Vars) > 0 || len(explicitlyTypedConsts) > 0 {
			isEnum = true
		}
	}

	if !isEnum {
		log.Printf("type %s is not an enum (base type %s, %d consts, %d vars, %d explicitly typed consts)", typeInfo.TypeName, ident.Name, len(t.Consts), len(t.Vars), len(explicitlyTypedConsts))
		return false
	}

	log.Printf("type %s is an enum (base type %s)", typeInfo.TypeName, ident.Name)

	// Process constants
	for _, c := range t.Consts {
		for _, name := range c.Names {
			if !ast.IsExported(name) {
				log.Printf("skipping unexported enum const value: %s", name)
				continue
			}
			log.Printf("found exported enum const value: %s", name)

			docString := c.Doc
			// Find the specific doc for this const value
			if c.Decl != nil {
				for _, spec := range c.Decl.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, n := range vs.Names {
							if n.Name == name {
								if vs.Doc != nil {
									docString = vs.Doc.Text()
								}
								break
							}
						}
					}
				}
			}

			typeInfo.EnumValues = append(typeInfo.EnumValues, EnumInfo{
				Name:            name,
				DocString:       strings.TrimSpace(docString),
				ParsedDocString: *parseGoDocString(docString),
			})
		}
	}

	// Process variables
	for _, v := range t.Vars {
		for _, name := range v.Names {
			if !ast.IsExported(name) {
				log.Printf("skipping unexported enum var value: %s", name)
				continue
			}

			// Check if this variable has the correct type annotation
			if v.Decl != nil {
				for _, spec := range v.Decl.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						// Check if this variable spec contains our variable name
						hasOurVar := false
						for _, n := range vs.Names {
							if n.Name == name {
								hasOurVar = true
								break
							}
						}
						if !hasOurVar {
							continue
						}

						// Check if the type matches our enum's base type
						if vs.Type != nil {
							if typeIdent, ok := vs.Type.(*ast.Ident); ok {
								if typeIdent.Name == typeInfo.TypeName {
									log.Printf("found exported enum var value: %s", name)

									docString := v.Doc
									if vs.Doc != nil {
										docString = vs.Doc.Text()
									}

									typeInfo.EnumValues = append(typeInfo.EnumValues, EnumInfo{
										Name:            name,
										DocString:       strings.TrimSpace(docString),
										ParsedDocString: *parseGoDocString(docString),
									})
								}
							}
						}
					}
				}
			}
		}
	}

	// Add explicitly typed constants that weren't already included
	typeInfo.EnumValues = append(typeInfo.EnumValues, explicitlyTypedConsts...)

	return true
}

func parsePackage(pkgDir string, allTypes map[string]TypeInfo) (map[string]bool, error) {
	pkgImportPath, err := getPkgPathFromDir(pkgDir)
	if err != nil {
		log.Printf("Skipping directory %s, not a Go package: %v", pkgDir, err)
		return nil, nil
	}
	log.Printf("parsing package: %s", pkgImportPath)

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgDir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	externalPkgs := make(map[string]bool)

	for _, pkg := range pkgs {
		if strings.HasSuffix(pkg.Name, "_test") {
			log.Printf("skipping test package: %s", pkg.Name)
			continue
		}
		log.Printf("processing package AST: %s", pkg.Name)
		docPkg := doc.New(pkg, pkgImportPath, 0)

		for _, c := range docPkg.Consts {
			log.Printf("Const: %v", c.Names)
		}

		for _, v := range docPkg.Vars {
			log.Printf("Vars: %v", v.Names)
		}

		for _, t := range docPkg.Types {
			processType(t, pkgImportPath, allTypes, pkg, externalPkgs, docPkg)
		}
	}

	return externalPkgs, nil
}

func processType(
	t *doc.Type,
	pkgImportPath string,
	allTypes map[string]TypeInfo,
	pkg *ast.Package,
	externalPkgs map[string]bool,
	docPkg *doc.Package,
) {
	log.Printf("found type: %s", t.Name)
	typeSpec, ok := t.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return
	}

	typeName := typeSpec.Name.Name
	if !ast.IsExported(typeName) {
		log.Printf("skipping unexported type: %s", typeName)
		return
	}

	qualifiedTypeName := pkgImportPath + "." + typeName
	if _, exists := allTypes[qualifiedTypeName]; exists {
		log.Printf("skipping already processed type: %s", qualifiedTypeName)
		return
	}

	typeInfo := TypeInfo{
		Package:         pkgImportPath,
		TypeName:        typeName,
		Fields:          []FieldInfo{},
		EnumValues:      []EnumInfo{},
		DocString:       strings.TrimSpace(t.Doc),
		ParsedDocString: *parseGoDocString(t.Doc),
	}

	isProcessed := false

	switch spec := typeSpec.Type.(type) {
	case *ast.StructType:
		log.Printf("type %s is a struct", qualifiedTypeName)
		if err := processStruct(&typeInfo, typeSpec, spec, pkg, pkgImportPath, externalPkgs); err != nil {
			log.Printf("Error processing struct %s: %v", qualifiedTypeName, err)
			return
		}
		isProcessed = true
	case *ast.Ident:
		log.Printf("type %s is an ident, checking for enum", qualifiedTypeName)
		if processEnum(&typeInfo, spec, t, docPkg, pkgImportPath) {
			isProcessed = true
		}
	}

	if isProcessed {
		log.Printf("successfully processed type: %s", qualifiedTypeName)
		allTypes[qualifiedTypeName] = typeInfo
	} else {
		log.Printf("type %s was not processed (not a struct or enum)", qualifiedTypeName)
	}
}
