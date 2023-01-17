// Package exitlint defines an Analyzer that reports os.Exit use in main functions.
//
// Available at https://github.com/SiberianMonster/go-musthave-devops-tpl/exitlint
package exitlint

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
    Name: "exitcheck",
    Doc:  "check for os.Exit use",
    Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	foundExit := false
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.GoStmt:
				switch fn := n.Call.Fun.(type) {
				case *ast.FuncLit:
					ast.Inspect(fn, func(node ast.Node) bool {
						switch n := node.(type) {
						case *ast.Ident:
							if n.Name == "os.Exit" {
								foundExit = true
							return false
							}
						}
						return true
					})
				}
			}
			if foundExit {
			pass.Reportf(node.Pos(), "os.Exit used in the code")
			}
		return
		})
	}

	return nil, nil
}