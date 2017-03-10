package main

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2017 ESSENTIAL KAOS                         //
//        Essential Kaos Open Source License <https://essentialkaos.com/ekol>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"go/parser"
	"go/token"

	"pkg.re/essentialkaos/ek.v7/arg"
	"pkg.re/essentialkaos/ek.v7/env"
	"pkg.re/essentialkaos/ek.v7/fmtc"
	"pkg.re/essentialkaos/ek.v7/fsutil"
	"pkg.re/essentialkaos/ek.v7/path"
	"pkg.re/essentialkaos/ek.v7/sliceutil"
	"pkg.re/essentialkaos/ek.v7/usage"
	"pkg.re/essentialkaos/ek.v7/usage/update"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	APP  = "gomakegen"
	VER  = "0.2.0"
	DESC = "Utility for generating makefiles for Golang applications"
)

const (
	ARG_GLIDE      = "g:glide"
	ARG_METALINTER = "m:metalinter"
	ARG_BENCHMARK  = "b:benchmark"
	ARG_VERB_TESTS = "V:verbose"
	ARG_OUTPUT     = "o:output"
	ARG_NO_COLOR   = "nc:no-color"
	ARG_HELP       = "h:help"
	ARG_VER        = "v:version"
)

const SEPARATOR = "########################################################################################"

// ////////////////////////////////////////////////////////////////////////////////// //

type Makefile struct {
	BaseImports []string
	TestImports []string
	Binaries    []string

	HasTests   bool
	Benchmark  bool
	VerbTests  bool
	Glide      bool
	Metalinter bool
}

// ////////////////////////////////////////////////////////////////////////////////// //

var argMap = arg.Map{
	ARG_OUTPUT:     {Value: "Makefile"},
	ARG_GLIDE:      {Type: arg.BOOL},
	ARG_METALINTER: {Type: arg.BOOL},
	ARG_BENCHMARK:  {Type: arg.BOOL},
	ARG_VERB_TESTS: {Type: arg.BOOL},
	ARG_NO_COLOR:   {Type: arg.BOOL},
	ARG_HELP:       {Type: arg.BOOL, Alias: "u:usage"},
	ARG_VER:        {Type: arg.BOOL, Alias: "ver"},
}

var checkPackageImports = []string{
	"github.com/go-check/check",
	"gopkg.in/check.v1",
	"pkg.re/check.v1",
}

// ////////////////////////////////////////////////////////////////////////////////// //

func main() {
	runtime.GOMAXPROCS(1)

	args, errs := arg.Parse(argMap)

	if len(errs) != 0 {
		printError("Arguments parsing errors:")

		for _, err := range errs {
			printError("  %v", err)
		}

		os.Exit(1)
	}

	if arg.GetB(ARG_NO_COLOR) {
		fmtc.DisableColors = true
	}

	if arg.GetB(ARG_VER) {
		showAbout()
		return
	}

	if arg.GetB(ARG_HELP) || len(args) == 0 {
		showUsage()
		return
	}

	dir := args[0]

	checkDir(dir)
	process(dir)
}

// checkDir check directory with sources
func checkDir(dir string) {
	if !fsutil.IsExist(dir) {
		printWarn("Directory %s does not exist", dir)
		os.Exit(1)
	}

	if !fsutil.IsReadable(dir) {
		printWarn("Directory %s is not readable", dir)
		os.Exit(1)
	}

	if !fsutil.IsExecutable(dir) {
		printWarn("Directory %s is not executable", dir)
		os.Exit(1)
	}
}

// process start sources processing
func process(dir string) {
	sources := fsutil.ListAllFiles(
		dir, true,
		fsutil.ListingFilter{
			MatchPatterns: []string{"*.go"},
			SizeGreater:   1,
		},
	)

	makefile := collectImports(sources, dir)

	makefile.Glide = arg.GetB(ARG_GLIDE)
	makefile.Metalinter = arg.GetB(ARG_METALINTER)
	makefile.Benchmark = arg.GetB(ARG_BENCHMARK)
	makefile.VerbTests = arg.GetB(ARG_VERB_TESTS)

	makefile.Cleanup(dir)

	exportMakefile(makefile)
}

// exportMakefile render makefile and write data to file
func exportMakefile(makefile *Makefile) {
	err := ioutil.WriteFile(arg.GetS(ARG_OUTPUT), makefile.Render(), 0644)

	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	fmtc.Printf("{g}Makefile successfully created as {g*}%s{!}\n", arg.GetS(ARG_OUTPUT))
}

// collectImports collect import from source files and return imports for
// base sources, test sources and slice with binaries
func collectImports(sources []string, dir string) *Makefile {
	var (
		bImports map[string]bool
		tImports map[string]bool
		binaries []string
		imports  []string
		isBinary bool
		source   string
		path     string
	)

	bSources, tSources := splitSources(sources)

	bImports = make(map[string]bool)

	for _, source = range bSources {
		imports, isBinary = extractImports(source, dir)

		for _, path = range imports {
			bImports[path] = true
		}

		// Append to slice only binaries in root directory
		if isBinary && !strings.Contains(source, "/") {
			binaries = append(binaries, source)
		}
	}

	if len(tSources) != 0 {
		tImports = make(map[string]bool)

		for _, source = range tSources {
			imports, _ = extractImports(source, dir)

			for _, path = range imports {
				tImports[path] = true
			}
		}
	}

	return &Makefile{
		BaseImports: importMapToSlice(bImports),
		TestImports: importMapToSlice(tImports),
		Binaries:    binaries,
		HasTests:    hasTests(sources),
	}
}

// splitSources split sources to two slices - with base sources and test sources
func splitSources(sources []string) ([]string, []string) {
	if !hasTests(sources) {
		return sources, nil
	}

	var bSources, tSources []string

	for _, source := range sources {
		if isTestSource(source) {
			tSources = append(tSources, source)
		} else {
			bSources = append(bSources, source)
		}
	}

	return bSources, tSources
}

// cleanupImports remove internal packages and local imports
func cleanupImports(imports []string, dir string) []string {
	if len(imports) == 0 {
		return nil
	}

	result := make(map[string]bool)

	gopath := env.Get().GetS("GOPATH")
	absDir, _ := filepath.Abs(dir)
	baseSrc := strings.Replace(absDir, gopath+"/src/", "", -1)

	for _, imp := range imports {
		if !isExternalPackage(imp) {
			continue
		}

		if strings.HasPrefix(imp, baseSrc) {
			continue
		}

		result[getPackageRoot(imp, gopath)] = true
	}

	return importMapToSlice(result)
}

func cleanupBinaries(binaries []string) []string {
	var result []string

	for _, bin := range binaries {
		result = append(result, strings.Replace(bin, ".go", "", -1))
	}

	return result
}

// extractImports return slice with all imports in source file
func extractImports(source, dir string) ([]string, bool) {
	fset := token.NewFileSet()
	file := path.Join(dir, source)
	f, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)

	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	var result []string
	var isBinary bool

	for _, imp := range f.Imports {
		if f.Name.String() == "main" {
			isBinary = true
		}

		result = append(result, strings.Trim(imp.Path.Value, "\""))
	}

	return result, isBinary
}

// hasTests return true if project has tests
func hasTests(sources []string) bool {
	for _, source := range sources {
		if isTestSource(source) {
			return true
		}
	}

	return false
}

// isTestSource return true if given file is tests
func isTestSource(source string) bool {
	return strings.HasSuffix(source, "_test.go")
}

// getPackageRoot return root for package
func getPackageRoot(pkg, gopath string) string {
	if isPackageRoot(gopath + "/src/" + pkg) {
		return pkg
	}

	pkgSlice := strings.Split(pkg, "/")

	for i := 2; i < len(pkgSlice); i++ {
		path := strings.Join(pkgSlice[:i], "/")

		if isPackageRoot(gopath + "/src/" + path) {
			return path
		}
	}

	return pkg
}

// isPackageRoot return true if given path is root for package
func isPackageRoot(path string) bool {
	return fsutil.IsExist(path + "/.git")
}

// isExternalPackage return true if given package is external
func isExternalPackage(pkg string) bool {
	pkgSlice := strings.Split(pkg, "/")

	if len(pkgSlice) == 0 || !strings.Contains(pkgSlice[0], ".") {
		return false
	}
	return true
}

// importMapToSlice convert map with package names to string slice
func importMapToSlice(imports map[string]bool) []string {
	if len(imports) == 0 {
		return nil
	}

	var result []string

	for path := range imports {
		result = append(result, path)
	}

	sort.Strings(result)

	return result
}

// containsPackage return true if imports contains given packages
func containsPackage(imports []string, pkgs []string) bool {
	for _, pkg := range pkgs {
		if sliceutil.Contains(imports, pkg) {
			return true
		}
	}

	return false
}

// printError prints error message to console
func printError(f string, a ...interface{}) {
	fmtc.Printf("{r}"+f+"{!}\n", a...)
}

// printWarn prints warning message to console
func printWarn(f string, a ...interface{}) {
	fmtc.Printf("{y}"+f+"{!}\n", a...)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Cleanup clean imports and binaries
func (m *Makefile) Cleanup(dir string) {
	m.BaseImports = cleanupImports(m.BaseImports, dir)
	m.TestImports = cleanupImports(m.TestImports, dir)

	m.Binaries = cleanupBinaries(m.Binaries)

	sort.Strings(m.Binaries)
}

// Render return makefile data
func (m *Makefile) Render() []byte {
	var result string

	result += SEPARATOR + "\n\n"
	result += m.getPhony() + "\n"
	result += SEPARATOR + "\n\n"

	result += m.getBinTarget()
	result += m.getDepsTarget()
	result += m.getTestTarget()
	result += m.getGlideTarget()
	result += m.getFmtTarget()
	result += m.getMetalinterTarget()
	result += m.getCleanTarget()

	result += SEPARATOR + "\n\n"

	return []byte(result)
}

// getPhony return PHONY part of makefile
func (m *Makefile) getPhony() string {
	phony := []string{"fmt"}

	if len(m.Binaries) != 0 {
		phony = append(phony, "all", "clean")
	}

	if len(m.BaseImports) != 0 {
		phony = append(phony, "deps")
	}

	if len(m.TestImports) != 0 {
		phony = append(phony, "deps-test", "test")
	}

	if m.Glide {
		phony = append(phony, "glide-up", "glide-install")
	}

	if m.Benchmark {
		phony = append(phony, "benchmark")
	}

	return ".PHONY = " + strings.Join(phony, " ") + "\n"
}

// getMetalinterTarget generate target for "all" command and all sub targets
// for each binary
func (m *Makefile) getBinTarget() string {
	if len(m.Binaries) == 0 {
		return ""
	}

	var result string

	result += "all: " + strings.Join(m.Binaries, " ") + "\n\n"

	for _, bin := range m.Binaries {
		result += bin + ":\n"
		result += "\tgo build " + bin + ".go\n\n"
	}

	return result
}

// getMetalinterTarget generate target for "deps" command
func (m *Makefile) getDepsTarget() string {
	if len(m.BaseImports) == 0 {
		return ""
	}

	var result string

	result += "deps:\n"

	for _, pkg := range m.BaseImports {
		result += "\tgo get -v " + pkg + "\n"
	}

	result += "\n"

	return result
}

// getMetalinterTarget generate target for "test", "deps-test"
// and "benchmark" commands
func (m *Makefile) getTestTarget() string {
	if !m.HasTests {
		return ""
	}

	var result string

	if len(m.TestImports) != 0 {
		result += "deps-test:\n"

		for _, pkg := range m.TestImports {
			result += "\tgo get -v " + pkg + "\n"
		}

		result += "\n"
	}

	result += "test:\n"

	if m.VerbTests {
		result += "\tgo test -v -covermode=count .\n"
	} else {
		result += "\tgo test -covermode=count .\n"
	}

	result += "\n"

	if m.Benchmark {
		result += "benchmark:\n"

		if containsPackage(m.TestImports, checkPackageImports) {
			result += "\tgo test -check.v -check.b -check.bmem\n"
		} else {
			result += "\tgo test -bench=.\n"
		}

		result += "\n"
	}

	return result
}

// getMetalinterTarget generate target for "fmt" command
func (m *Makefile) getFmtTarget() string {
	var result string

	result += "fmt:\n"
	result += "\tfind . -name \"*.go\" -exec gofmt -s -w {} \\;\n"
	result += "\n"

	return result
}

// getMetalinterTarget generate target for "clean" command
func (m *Makefile) getCleanTarget() string {
	if len(m.Binaries) == 0 {
		return ""
	}

	var result string

	result += "clean:\n"

	for _, bin := range m.Binaries {
		result += "\trm -f " + bin + "\n"
	}

	result += "\n"

	return result
}

// getMetalinterTarget generate target for "glide-up" and
// "glide-install" commands
func (m *Makefile) getGlideTarget() string {
	if !m.Glide {
		return ""
	}

	var result string

	result += "glide-up:\n"
	result += "\tglide up\n"
	result += "\n"

	result += "glide-install:\n"
	result += "\tglide install\n"
	result += "\n"

	return result
}

// getMetalinterTarget generate target for "metalineter" command
func (m *Makefile) getMetalinterTarget() string {
	if !m.Metalinter {
		return ""
	}

	var result string

	result += "metalinter:\n"
	result += "\ttest -s $(GOPATH)/bin/gometalinter || go get -u github.com/alecthomas/gometalinter ; $(GOPATH)/bin/gometalinter --install\n"
	result += "\t$(GOPATH)/bin/gometalinter --deadline 30s\n"
	result += "\n"

	return result
}

// ////////////////////////////////////////////////////////////////////////////////// //

//
func showUsage() {
	info := usage.NewInfo("", "dir")

	info.AddOption(ARG_GLIDE, "Add target to fetching dependecies with glide")
	info.AddOption(ARG_METALINTER, "Add target with metalinter check")
	info.AddOption(ARG_BENCHMARK, "Add target to run benchmarks")
	info.AddOption(ARG_VERB_TESTS, "Enable verbose output for tests")
	info.AddOption(ARG_OUTPUT, "Output file {s-}(Makefile by default){!}")
	info.AddOption(ARG_NO_COLOR, "Disable colors in output")
	info.AddOption(ARG_HELP, "Show this help message")
	info.AddOption(ARG_VER, "Show version")

	info.AddExample(
		"$GOPATH/src/github.com/profile/project",
		"Generate makefile for github.com/profile/project and save as Makefile",
	)

	info.AddExample(
		"$GOPATH/src/github.com/profile/project -o project.make",
		"Generate makefile for github.com/profile/project and save as project.make",
	)

	info.Render()
}

func showAbout() {
	about := &usage.About{
		App:           APP,
		Version:       VER,
		Desc:          DESC,
		Year:          2009,
		Owner:         "Essential Kaos",
		License:       "Essential Kaos Open Source License <https://essentialkaos.com/ekol>",
		UpdateChecker: usage.UpdateChecker{"essentialkaos/gomakegen", update.GitHubChecker},
	}

	about.Render()
}
