// Multichecker module is a tool for static analysis of Go programs.
//
// Available at https://github.com/SiberianMonster/go-musthave-devops-tpl/cmd/staticlint
package main

import (
	"encoding/json"
    "os"
    "path/filepath"

	"golang.org/x/tools/go/analysis"
	"go/ast"
    "golang.org/x/tools/go/analysis/multichecker"
    "golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"honnef.co/go/tools/staticcheck"

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
		return true
		})
	}

	return nil, nil
}


// Config — имя файла конфигурации.
const Config = `/config.json`

// ConfigData описывает структуру файла конфигурации.
type ConfigData struct {
    Staticcheck []string
}

func main() {
	appfile, err := os.Executable()
    
    data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
    if err != nil {
        panic(err)
    }
    var cfg ConfigData
    if err = json.Unmarshal(data, &cfg); err != nil {
        panic(err)
    }
    mychecks := []*analysis.Analyzer{
        ExitCheckAnalyzer,
        asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
        tests.Analyzer,
        timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
        unsafeptr.Analyzer,
        unusedresult.Analyzer,
		unusedwrite.Analyzer,
        usesgenerics.Analyzer,
    }
    checks := make(map[string]bool)
    for _, v := range cfg.Staticcheck {
        checks[v] = true
    }
    // добавляем анализаторы из staticcheck, которые указаны в файле конфигурации
    for _, v := range staticcheck.Analyzers {
        if checks[v.Analyzer.Name] {
            mychecks = append(mychecks, v.Analyzer)
        }
    }
    multichecker.Main(
        mychecks...,
    )
} 