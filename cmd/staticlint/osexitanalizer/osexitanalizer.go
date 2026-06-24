package osexitanalizer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer checks that os.Exit is not called directly in main function of main package
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "errcheck",
	Doc:  "checks that os.Exit is not called directly in main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	callExpr := func(x *ast.CallExpr) {
		// check that this is function call
		selectorExpr, ok := x.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		// check that this is os.Exit
		if ident, ok := selectorExpr.X.(*ast.Ident); ok {
			if ident.Name == "os" && selectorExpr.Sel.Name == "Exit" {
				// check package name
				if obj := pass.TypesInfo.Uses[ident]; obj != nil {
					pkg := obj.Pkg()
					if pkg != nil && pkg.Name() == "main" {
						pass.Reportf(
							x.Pos(),
							"direct call to os.Exit in main function is forbidden",
						)
					}
				}
			}
		}
	}
	for _, file := range pass.Files {
		// go through all AST nodes
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.CallExpr:
				callExpr(x)
			}
			return true
		})
	}
	return nil, nil
}
