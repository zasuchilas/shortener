// Package osexitcheck is a custom analyzer that checks if there is no os.Exit call in func main().
package osexitcheck

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"strings"
)

// OsExitAnalyzer describes an analysis function and its options.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check for the use of a direct os.Exit call in the main function of the main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Name.Name != "main" {
				continue
			}
			pos := pass.Fset.Position(fn.Pos())
			if strings.Contains(pos.Filename, ".cache") {
				continue
			}

			check(pass, fn)
		}
	}

	return nil, nil
}

func check(pass *analysis.Pass, fn *ast.FuncDecl) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Name == "os" && sel.Sel.Name == "Exit" {
			pass.Reportf(call.Pos(), "it is forbidden to use a direct os.Exit call in the main function of the main package")
		}
		return true
	})
}
