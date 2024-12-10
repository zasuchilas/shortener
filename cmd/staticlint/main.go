// The staticlint application is a custom multichecker build
// for checking the shortener project code.
//
// It contains all the standard static analyzers of the package
// golang.org/x/tools/go/analysis/passes.
// See the detailed description of the analyzers in the documentation
// https://staticcheck.dev/docs/checks/.
//
// It also contains all the analyzers of the package staticcheck.io.
// See documentation https://staticcheck.dev/docs/checks/.
//
// Added custom analyzer osexitcheck.
//
// To use:
//
//	go run ./cmd/staticlint/ ./...
//	cd ./cmd/staticlint && go build -o staticlint
//	./cmd/staticlint/staticlint ./...
//	./cmd/staticlint/staticlint --help
package main

import (
	"github.com/timakin/bodyclose/passes/bodyclose"
	"github.com/zasuchilas/shortener/cmd/staticlint/osexitcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
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
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
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
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
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
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	// collecting analyzers
	var analyzers []*analysis.Analyzer
	analyzers = append(analyzers, analysisPassesAnalyzers()...)
	analyzers = append(analyzers, staticcheckAnalyzers()...)
	analyzers = append(analyzers, simpleAnalyzers()...)
	analyzers = append(analyzers, stylecheckAnalyzers()...)
	analyzers = append(analyzers, quickfixAnalyzers()...)
	analyzers = append(analyzers, osexitcheck.OsExitAnalyzer)
	analyzers = append(analyzers, customAnalyzers()...)

	// linting
	multichecker.Main(
		analyzers...,
	)
}

// All standard analyzers from golang.org/x/tools/go/analysis/passes.
func analysisPassesAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		appends.Analyzer,
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
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		//fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpmux.Analyzer,
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
		slog.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stdversion.Analyzer,
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
		//waitgroup.Analyzer,
	}
}

// All SA staticcheck analyzers from honnef.co/go/tools/staticcheck.
func staticcheckAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}
	return analyzers
}

// All S (simple) staticcheck analyzers from honnef.co/go/tools/simple.
func simpleAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer
	for _, v := range simple.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}
	return analyzers
}

// All ST (stylecheck) staticcheck analyzers from honnef.co/go/tools/stylecheck.
func stylecheckAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer
	for _, v := range stylecheck.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}
	return analyzers
}

// All QF (quickfix) staticcheck analyzers from honnef.co/go/tools/quickfix.
func quickfixAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer
	for _, v := range quickfix.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}
	return analyzers
}

// Some custom public analyzers.
func customAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		// checks whether res.Body is correctly closed https://github.com/timakin/bodyclose
		bodyclose.Analyzer,
	}
}
