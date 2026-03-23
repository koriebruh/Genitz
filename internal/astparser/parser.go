package astparser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

func AddImport(filePath, importPath string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	alias := ""
	parts := strings.SplitN(importPath, " ", 2)
	if len(parts) == 2 {
		alias = parts[0]
		importPath = parts[1]
	}

	// Check if already imported
	for _, imp := range f.Imports {
		if imp.Path.Value == strconv.Quote(importPath) {
			return nil
		}
	}

	newImp := &ast.ImportSpec{
		Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(importPath)},
	}
	if alias != "" {
		newImp.Name = &ast.Ident{Name: alias}
	}

	injected := false
	for _, d := range f.Decls {
		if gen, ok := d.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
			gen.Specs = append(gen.Specs, newImp)
			f.Imports = append(f.Imports, newImp)
			injected = true
			break
		}
	}

	if !injected {
		gen := &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: []ast.Spec{newImp},
		}
		// Insert right after package declaration
		f.Decls = append([]ast.Decl{gen}, f.Decls...)
		f.Imports = append(f.Imports, newImp)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return fmt.Errorf("failed to format node: %w", err)
	}

	return os.WriteFile(filePath, buf.Bytes(), 0644)
}

// InjectToMain parses a raw golang statement snippet and safely appends it 
// into the body sequence of the `func main()` block located in `filePath`.
func InjectToMain(filePath, snippet string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	// Parse snippet into statements by wrapping it in a dummy func
	snippetCode := "package main\nfunc ____() {\n" + snippet + "\n}"
	snippetFile, err := parser.ParseFile(fset, "", snippetCode, 0)
	if err != nil {
		return fmt.Errorf("failed to parse snippet syntax: %w", err)
	}

	var newStmts []ast.Stmt
	for _, d := range snippetFile.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "____" {
			newStmts = fn.Body.List
			break
		}
	}

	if len(newStmts) == 0 {
		return nil // nothing to inject
	}

	modified := false
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
			fn.Body.List = append(fn.Body.List, newStmts...)
			modified = true
			break
		}
	}

	if !modified {
		return fmt.Errorf("func main() not found in %s", filePath)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return fmt.Errorf("failed to format node: %w", err)
	}

	return os.WriteFile(filePath, buf.Bytes(), 0644)
}

// InjectStructField parse code snippet property (e.g `Redis config.RedisConfig`) and safely append it to target struct name globally.
func InjectStructField(filePath, structName, fieldStr string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	// Parse the fieldStr into an *ast.Field by wrapping it in a dummy struct
	snippetCode := "package main\ntype ____ struct {\n\t" + fieldStr + "\n}"
	snippetFile, err := parser.ParseFile(fset, "", snippetCode, 0)
	if err != nil {
		return fmt.Errorf("failed to parse snippet field syntax: %w", err)
	}

	var newField *ast.Field
	for _, d := range snippetFile.Decls {
		if gen, ok := d.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
			if len(gen.Specs) > 0 {
				if tSpec, ok := gen.Specs[0].(*ast.TypeSpec); ok {
					if st, ok := tSpec.Type.(*ast.StructType); ok && len(st.Fields.List) > 0 {
						newField = st.Fields.List[0]
						break
					}
				}
			}
		}
	}

	if newField == nil {
		return fmt.Errorf("could not extract field from snippet")
	}

	modified := false
	for _, d := range f.Decls {
		if gen, ok := d.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
			for _, spec := range gen.Specs {
				if tSpec, ok := spec.(*ast.TypeSpec); ok && tSpec.Name.Name == structName {
					if st, ok := tSpec.Type.(*ast.StructType); ok {
						st.Fields.List = append(st.Fields.List, newField)
						modified = true
					}
				}
			}
		}
	}

	if !modified {
		return fmt.Errorf("struct %s not found in %s", structName, filePath)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return fmt.Errorf("failed to format node: %w", err)
	}

	return os.WriteFile(filePath, buf.Bytes(), 0644)
}

// RemoveImport removes an import path from a Go file (handles alias like `_ "pkg"`).
func RemoveImport(filePath, importPath string) error {
	// Handle `alias pkg` format — extract bare path
	parts := strings.SplitN(importPath, " ", 2)
	if len(parts) == 2 {
		importPath = parts[1]
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	quoted := strconv.Quote(importPath)
	modified := false
	for _, d := range f.Decls {
		gen, ok := d.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			continue
		}
		var kept []ast.Spec
		for _, spec := range gen.Specs {
			imp, ok := spec.(*ast.ImportSpec)
			if ok && imp.Path.Value == quoted {
				modified = true
				continue // drop it
			}
			kept = append(kept, spec)
		}
		gen.Specs = kept
	}

	if !modified {
		return fmt.Errorf("import %q not found in %s", importPath, filePath)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return fmt.Errorf("failed to format node: %w", err)
	}
	return os.WriteFile(filePath, buf.Bytes(), 0644)
}

// RemoveStatementsMatching removes statements in a named func that contain a given substring.
// Used by `genitz remove` to undo injected init code.
func RemoveStatementsMatching(filePath, funcName, keyword string) error {
	fset := token.NewFileSet()
	src, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	f, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	modified := false
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name.Name != funcName || fn.Body == nil {
			continue
		}
		var kept []ast.Stmt
		for _, stmt := range fn.Body.List {
			// Render statement to text and check whether keyword appears
			var sb bytes.Buffer
			format.Node(&sb, fset, stmt) //nolint:errcheck
			if strings.Contains(sb.String(), keyword) {
				modified = true
				continue // drop it
			}
			kept = append(kept, stmt)
		}
		fn.Body.List = kept
	}

	if !modified {
		return fmt.Errorf("no statement matching %q found in func %s", keyword, funcName)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return fmt.Errorf("failed to format node: %w", err)
	}
	return os.WriteFile(filePath, buf.Bytes(), 0644)
}
