package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "check for os.Exit call in main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				if fn.Recv == nil && pass.Pkg.Name() == "main" {
					for _, stmt := range fn.Body.List {
						if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
							if call, ok := exprStmt.X.(*ast.CallExpr); ok {
								if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
									if pkgIdent, ok := fun.X.(*ast.Ident); ok && pkgIdent.Name == "os" && fun.Sel.Name == "Exit" {
										pass.Reportf(call.Pos(), "os.Exit call in main.main is not allowed")
									}
								}
							}
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
