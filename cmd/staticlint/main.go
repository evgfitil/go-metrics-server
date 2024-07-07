// Package main provides a multichecker that combines various static analysis tools.
//
// To run the multichecker, execute the following command:
//
//	go run ./cmd/staticlint ./...
//
// The multichecker includes:
// - Standard static analysis passes from golang.org/x/tools/go/analysis/passes
// - All SA class analyzers from staticcheck.io
// - One analyzer from other classes of staticcheck.io
// - Two public analyzers: ineffassign and errcheck
// - A custom analyzer that forbids the use of os.Exit in the main function of the main package
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers, []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}...)

	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[0:2] == "SA" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[0:2] != "SA" {
			analyzers = append(analyzers, v.Analyzer)
			break
		}
	}

	analyzers = append(analyzers, ineffassign.Analyzer)
	analyzers = append(analyzers, errcheck.Analyzer)
	analyzers = append(analyzers, NoOsExitAnalyzer)

	multichecker.Main(analyzers...)
}
