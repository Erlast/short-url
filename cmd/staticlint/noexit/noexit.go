package noexit

import (
	"errors"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const noExitAnalyzerName = "noexit"

// Analyzer структура analysis.Analyzer для noosexit.
var Analyzer = &analysis.Analyzer{
	Name: noExitAnalyzerName,
	Doc:  "disallow direct calls to os.Exit in main function",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

// run функция для запукска анализатора.
func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, errors.New("not a main package")
	}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != "main" {
				continue
			}
			ast.Inspect(funcDecl, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
							pass.Reportf(call.Pos(), "direct call to os.Exit is not allowed")
						}
					}
				}
				return true
			})
		}
	}

	return "", nil
}
