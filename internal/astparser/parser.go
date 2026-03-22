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
)

// AddImport safely injects a new package import to a go file if it doesn't already exist.
func AddImport(filePath, importPath string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filePath, err)
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
