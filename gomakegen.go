package main

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2018 ESSENTIAL KAOS                         //
//        Essential Kaos Open Source License <https://essentialkaos.com/ekol>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"go/parser"
	"go/token"

	"pkg.re/essentialkaos/ek.v9/env"
	"pkg.re/essentialkaos/ek.v9/fmtc"
	"pkg.re/essentialkaos/ek.v9/fsutil"
	"pkg.re/essentialkaos/ek.v9/options"
	"pkg.re/essentialkaos/ek.v9/path"
	"pkg.re/essentialkaos/ek.v9/sliceutil"
	"pkg.re/essentialkaos/ek.v9/strutil"
	"pkg.re/essentialkaos/ek.v9/usage"
	"pkg.re/essentialkaos/ek.v9/usage/update"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// App info
const (
	APP  = "gomakegen"
	VER  = "0.8.1"
	DESC = "Utility for generating makefiles for Go applications"
)

// Constants with options names
const (
	OPT_GLIDE      = "g:glide"
	OPT_DEP        = "d:dep"
	OPT_METALINTER = "m:metalinter"
	OPT_STRIP      = "s:strip"
	OPT_BENCHMARK  = "b:benchmark"
	OPT_RACE       = "r:race"
	OPT_VERB_TESTS = "V:verbose"
	OPT_OUTPUT     = "o:output"
	OPT_NO_COLOR   = "nc:no-color"
	OPT_HELP       = "h:help"
	OPT_VER        = "v:version"
)

// SEPARATOR_SIZE is default separator size
const SEPARATOR_SIZE = 80

// ////////////////////////////////////////////////////////////////////////////////// //

// Makefile contains full info for makefile generation
type Makefile struct {
	BaseImports []string
	TestImports []string
	Binaries    []string

	HasTests       bool
	Benchmark      bool
	VerbTests      bool
	Race           bool
	Metalinter     bool
	Strip          bool
	HasSubpackages bool

	GlideUsed bool
	DepUsed   bool
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Options map
var optMap = options.Map{
	OPT_OUTPUT:     {Value: "Makefile"},
	OPT_GLIDE:      {Type: options.BOOL},
	OPT_DEP:        {Type: options.BOOL},
	OPT_METALINTER: {Type: options.BOOL},
	OPT_STRIP:      {Type: options.BOOL},
	OPT_BENCHMARK:  {Type: options.BOOL},
	OPT_VERB_TESTS: {Type: options.BOOL},
	OPT_RACE:       {Type: options.BOOL},
	OPT_NO_COLOR:   {Type: options.BOOL},
	OPT_HELP:       {Type: options.BOOL, Alias: "u:usage"},
	OPT_VER:        {Type: options.BOOL, Alias: "ver"},
}

// Paths for check package
var checkPackageImports = []string{
	"github.com/go-check/check",
	"gopkg.in/check.v1",
	"pkg.re/check.v1",
}

// ////////////////////////////////////////////////////////////////////////////////// //

func main() {
	runtime.GOMAXPROCS(1)

	args, errs := options.Parse(optMap)

	if len(errs) != 0 {
		printError("Options parsing errors:")

		for _, err := range errs {
			printError("  %v", err)
		}

		os.Exit(1)
	}

	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
	}

	if options.GetB(OPT_VER) {
		showAbout()
		return
	}

	if options.GetB(OPT_HELP) || len(args) == 0 {
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

	makefile := generateMakefile(sources, dir)

	exportMakefile(makefile)
}

// exportMakefile render makefile and write data to file
func exportMakefile(makefile *Makefile) {
	err := ioutil.WriteFile(options.GetS(OPT_OUTPUT), makefile.Render(), 0644)

	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	fmtc.Printf("{g}Makefile successfully created as {g*}%s{!}\n", options.GetS(OPT_OUTPUT))
}

// generateMakefile collect imports, process options and generate makefile struct
func generateMakefile(sources []string, dir string) *Makefile {
	makefile := collectImports(sources, dir)

	applyOptionsFromMakefile(dir+"/"+options.GetS(OPT_OUTPUT), makefile)

	makefile.Metalinter = makefile.Metalinter || options.GetB(OPT_METALINTER)
	makefile.Benchmark = makefile.Benchmark || options.GetB(OPT_BENCHMARK)
	makefile.VerbTests = makefile.VerbTests || options.GetB(OPT_VERB_TESTS)
	makefile.Race = makefile.Race || options.GetB(OPT_RACE)
	makefile.Strip = makefile.Strip || options.GetB(OPT_STRIP)
	makefile.GlideUsed = makefile.GlideUsed && fsutil.IsExist(dir+"/glide.yaml")
	makefile.DepUsed = makefile.DepUsed && fsutil.IsExist(dir+"/Gopkg.toml")

	if options.GetB(OPT_GLIDE) {
		makefile.GlideUsed = true
	}

	if options.GetB(OPT_DEP) {
		makefile.DepUsed = true
	}

	makefile.Cleanup(dir)

	return makefile
}

// collectImports collect import from source files and return imports for
// base sources, test sources and slice with binaries
func collectImports(sources []string, dir string) *Makefile {
	baseSources, testSources := splitSources(sources)

	baseImports, binaries, hasSubPkgs := extractBaseImports(baseSources, dir)
	testImports := extractTestImports(testSources, dir)

	return &Makefile{
		BaseImports:    baseImports,
		TestImports:    testImports,
		Binaries:       binaries,
		HasTests:       hasTests(sources),
		HasSubpackages: hasSubPkgs,
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

// extractBaseImports extract base imports from given source files
func extractBaseImports(sources []string, dir string) ([]string, []string, bool) {
	importsMap := make(map[string]bool)
	binaries := make([]string, 0)
	hasSubPkgs := false

	for _, source := range sources {
		imports, isBinary := extractImports(source, dir)

		for _, path := range imports {
			importsMap[path] = true
		}

		// Append to slice only binaries in root directory
		if isBinary && !strings.Contains(source, "/") {
			binaries = append(binaries, source)
		}

		if !hasSubPkgs && strings.Contains(source, "/") {
			hasSubPkgs = true
		}
	}

	return importMapToSlice(importsMap), binaries, hasSubPkgs
}

// extractTestImports extract test imports from given source files
func extractTestImports(sources []string, dir string) []string {
	if len(sources) == 0 {
		return nil
	}

	importsMap := make(map[string]bool)

	for _, source := range sources {
		imports, _ := extractImports(source, dir)

		for _, path := range imports {
			importsMap[path] = true
		}
	}

	return importMapToSlice(importsMap)
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

// cleanupBinaries remove .go suffix from names of binaries
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
	if !fsutil.IsExist(path + "/.git") {
		return false
	}

	files := fsutil.List(path, true, fsutil.ListingFilter{MatchPatterns: []string{"*.go"}})

	return len(files) != 0
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

// containsStablePathImports return true if imports contains stable import services path
func containsStablePathImports(imports []string) bool {
	for _, pkg := range imports {
		if strings.HasPrefix(pkg, "pkg.re") {
			return true
		}

		if strings.HasPrefix(pkg, "gopkg.in") {
			return true
		}
	}

	return false
}

// getGitConfigurationForStableImports return slice with git configuration commands for
// stable import services
func getGitConfigurationForStableImports(imports []string) []string {
	var hasGopkg, hasPkgre bool
	var result []string

	for _, pkg := range imports {
		if strings.HasPrefix(pkg, "pkg.re") && !hasPkgre {
			result = append(result, "git config --global http.https://pkg.re.followRedirects true")
			hasPkgre = true
		}

		if strings.HasPrefix(pkg, "gopkg.in") && !hasGopkg {
			result = append(result, "git config --global http.https://gopkg.in.followRedirects true")
			hasGopkg = true
		}
	}

	return result
}

// applyOptionsFromFile read used options from previously generated Makefile
// and apply it to makefile struct
func applyOptionsFromMakefile(file string, m *Makefile) {
	if !fsutil.IsExist(file) {
		return
	}

	opts := extractOptionsFromMakefile(file)

	if opts == "" {
		return
	}

	for _, opt := range strutil.Fields(opts) {
		switch strings.TrimLeft(opt, "-") {
		case getOptionName(OPT_GLIDE):
			m.GlideUsed = true
		case getOptionName(OPT_DEP):
			m.DepUsed = true
		case getOptionName(OPT_METALINTER):
			m.Metalinter = true
		case getOptionName(OPT_STRIP):
			m.Strip = true
		case getOptionName(OPT_BENCHMARK):
			m.Benchmark = true
		case getOptionName(OPT_VERB_TESTS):
			m.VerbTests = true
		case getOptionName(OPT_RACE):
			m.Race = true
		}
	}
}

// extractOptionsFromMakefile extract options from previously generated Makefile
func extractOptionsFromMakefile(file string) string {
	fd, err := os.OpenFile(file, os.O_RDONLY, 0)

	if err != nil {
		return ""
	}

	defer fd.Close()

	r := bufio.NewReader(fd)
	s := bufio.NewScanner(r)

	for s.Scan() {
		text := s.Text()

		if !strings.HasPrefix(text, "# gomakegen ") {
			continue
		}

		return strings.Replace(text, "# gomakegen ", "", -1)
	}

	return ""
}

// printError prints error message to console
func printError(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{r}"+f+"{!}\n", a...)
}

// printError prints warning message to console
func printWarn(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{y}"+f+"{!}\n", a...)
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

	result += m.getHeader()
	result += m.getTargets()

	return []byte(result)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// getHeader return header data
func (m *Makefile) getHeader() string {
	var result string

	result += getSeparator() + "\n\n"
	result += m.getGenerationComment()
	result += getSeparator() + "\n\n"
	result += m.getDefaultGoal() + "\n"
	result += m.getPhony() + "\n"
	result += getSeparator() + "\n\n"

	return result
}

// getTargets return targets data
func (m *Makefile) getTargets() string {
	var result string

	result += m.getBinTarget()
	result += m.getInstallTarget()
	result += m.getUninstallTarget()
	result += m.getDepsTarget()
	result += m.getTestDepsTarget()
	result += m.getTestTarget()
	result += m.getBenchTarget()
	result += m.getGlideTarget()
	result += m.getDepTarget()
	result += m.getFmtTarget()
	result += m.getMetalinterTarget()
	result += m.getCleanTarget()
	result += m.getHelpTarget()

	result += getSeparator() + "\n"

	return result
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

	if m.GlideUsed {
		phony = append(phony, "glide-create", "glide-install", "glide-update")
	}

	if m.DepUsed {
		phony = append(phony, "dep-init", "dep-update")
	}

	if m.Benchmark {
		phony = append(phony, "benchmark")
	}

	if m.Metalinter {
		phony = append(phony, "metalinter")
	}

	phony = append(phony, "help")

	return ".PHONY = " + strings.Join(phony, " ") + "\n"
}

// getDefaultGoal return DEFAULT_GOAL part of makefile
func (m *Makefile) getDefaultGoal() string {
	return ".DEFAULT_GOAL := help"
}

// getBinTarget generate target for "all" command and all sub targets
// for each binary
func (m *Makefile) getBinTarget() string {
	if len(m.Binaries) == 0 {
		return ""
	}

	result := "all: " + strings.Join(m.Binaries, " ") + " ## Build all binaries\n\n"

	for _, bin := range m.Binaries {
		result += bin + ": ## " + fmtc.Sprintf("Build %s binary", bin) + "\n"

		if m.Strip {
			result += "\tgo build -ldflags=\"-s -w\" " + bin + ".go\n"
		} else {
			result += "\tgo build " + bin + ".go\n"
		}

		result += "\n"
	}

	return result
}

// getInstallTarget generate target for "install" command
func (m *Makefile) getInstallTarget() string {
	if len(m.Binaries) == 0 {
		return ""
	}

	result := "install: ## Install binaries\n"

	for _, bin := range m.Binaries {
		result += "\tcp " + bin + " /usr/bin/" + bin + "\n"
	}

	return result + "\n"
}

// getUninstallTarget generate target for "uninstall" command
func (m *Makefile) getUninstallTarget() string {
	if len(m.Binaries) == 0 {
		return ""
	}

	result := "uninstall: ## Uninstall binaries\n"

	for _, bin := range m.Binaries {
		result += "\trm -f /usr/bin/" + bin + "\n"
	}

	return result + "\n"
}

// getDepsTarget generate target for "deps" command
func (m *Makefile) getDepsTarget() string {
	if len(m.BaseImports) == 0 {
		return ""
	}

	result := "deps: ## Download dependencies\n"

	if containsStablePathImports(m.BaseImports) {
		for _, gitCommand := range getGitConfigurationForStableImports(m.BaseImports) {
			result += "\t" + gitCommand + "\n"
		}
	}

	for _, pkg := range m.BaseImports {
		result += "\tgo get -d -v " + pkg + "\n"
	}

	return result + "\n"
}

// getDepsTarget generate target for "deps-test" command
func (m *Makefile) getTestDepsTarget() string {
	if len(m.TestImports) == 0 {
		return ""
	}

	result := "deps-test: ## Download dependencies for tests\n"

	if containsStablePathImports(m.TestImports) {
		for _, gitCommand := range getGitConfigurationForStableImports(m.TestImports) {
			result += "\t" + gitCommand + "\n"
		}
	}

	for _, pkg := range m.TestImports {
		result += "\tgo get -d -v " + pkg + "\n"
	}

	return result + "\n"
}

// getTestTarget generate target for "test" command
func (m *Makefile) getTestTarget() string {
	if !m.HasTests {
		return ""
	}

	targets := "."

	if m.HasSubpackages {
		targets = "./..."
	}

	result := "test: ## Run tests\n"

	if m.Race {
		if m.VerbTests {
			result += "\tgo test -v -race " + targets + "\n"
		} else {
			result += "\tgo test -race " + targets + "\n"
		}
	}

	if m.VerbTests {
		result += "\tgo test -v -covermode=count " + targets + "\n"
	} else {
		result += "\tgo test -covermode=count " + targets + "\n"
	}

	return result + "\n"
}

// getBenchTarget generate target for "benchmark" command
func (m *Makefile) getBenchTarget() string {
	if !m.Benchmark {
		return ""
	}

	result := "benchmark: ## Run benchmarks\n"

	if containsPackage(m.TestImports, checkPackageImports) {
		result += "\tgo test -check.v -check.b -check.bmem\n"
	} else {
		result += "\tgo test -bench=.\n"
	}

	return result + "\n"
}

// getFmtTarget generate target for "fmt" command
func (m *Makefile) getFmtTarget() string {
	result := "fmt: ## Format source code with gofmt\n"
	result += "\tfind . -name \"*.go\" -exec gofmt -s -w {} \\;\n"

	return result + "\n"
}

// getCleanTarget generate target for "clean" command
func (m *Makefile) getCleanTarget() string {
	if len(m.Binaries) == 0 {
		return ""
	}

	result := "clean: ## Remove generated files\n"

	for _, bin := range m.Binaries {
		result += "\trm -f " + bin + "\n"
	}

	return result + "\n"
}

// getGlideTarget generate target for "glide-up" and
// "glide-install" commands
func (m *Makefile) getGlideTarget() string {
	if !m.GlideUsed {
		return ""
	}

	result := "glide-create: ## Initialize glide workspace\n"
	result += "\twhich glide &>/dev/null || (echo -e '\\e[31mGlide is not installed\\e[0m' ; exit 1)\n"
	result += "\tglide init\n"
	result += "\n"

	result += "glide-install: ## Install packages and dependencies through glide\n"
	result += "\twhich glide &>/dev/null || (echo -e '\\e[31mGlide is not installed\\e[0m' ; exit 1)\n"
	result += "\ttest -s glide.yaml || glide init\n"
	result += "\tglide install\n"
	result += "\n"

	result += "glide-update: ## Update packages and dependencies through glide\n"
	result += "\twhich glide &>/dev/null || (echo -e '\\e[31mGlide is not installed\\e[0m' ; exit 1)\n"
	result += "\ttest -s glide.yaml || glide init\n"
	result += "\tglide update\n"
	result += "\n"

	return result
}

// getDepTarget generate target for "dep-init" and
// "dep-update" commands
func (m *Makefile) getDepTarget() string {
	if !m.DepUsed {
		return ""
	}

	result := "dep-init: ## Initialize dep workspace\n"
	result += "\twhich dep &>/dev/null || (echo -e '\\e[31mDep is not installed\\e[0m' ; exit 1)\n"
	result += "\tdep init\n\n"

	result += "dep-update: ## Update packages and dependencies through dep\n"
	result += "\twhich dep &>/dev/null || (echo -e '\\e[31mDep is not installed\\e[0m' ; exit 1)\n"
	result += "\ttest -s Gopkg.toml || dep init\n"
	result += "\tdep ensure -update\n\n"

	return result
}

// getMetalinterTarget generate target for "metalineter" command
func (m *Makefile) getMetalinterTarget() string {
	if !m.Metalinter {
		return ""
	}

	result := "metalinter: ## Install and run gometalinter\n"
	result += "\ttest -s $(GOPATH)/bin/gometalinter || (go get -u github.com/alecthomas/gometalinter ; $(GOPATH)/bin/gometalinter --install)\n"
	result += "\t$(GOPATH)/bin/gometalinter --deadline 30s\n\n"

	return result
}

// getHelpTarget generate target for "help" command
func (m *Makefile) getHelpTarget() string {
	result := "help: ## Show this info\n"
	result += "\t@echo -e '\\nSupported targets:\\n'\n"
	result += "\t@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \\\n"
	result += "\t\t| awk 'BEGIN {FS = \":.*?## \"}; {printf \"  \\033[33m%-12s\\033[0m %s\\n\", $$1, $$2}'\n"
	result += "\t@echo -e ''\n\n"

	return result
}

// getGenerationComment return comment with all used flags
func (m *Makefile) getGenerationComment() string {
	result := "# This Makefile generated by GoMakeGen " + VER + " using next command:\n"
	result += "# gomakegen "

	if m.GlideUsed {
		result += fmtc.Sprintf("--%s ", getOptionName(OPT_GLIDE))
	}

	if m.DepUsed {
		result += fmtc.Sprintf("--%s ", getOptionName(OPT_DEP))
	}

	if m.Metalinter {
		result += fmtc.Sprintf("--%s ", getOptionName(OPT_METALINTER))
	}

	if m.Strip {
		result += fmtc.Sprintf("--%s ", getOptionName(OPT_STRIP))
	}

	if m.Benchmark {
		result += fmtc.Sprintf("--%s ", getOptionName(OPT_BENCHMARK))
	}

	if m.VerbTests {
		result += fmtc.Sprintf("--%s ", getOptionName(OPT_VERB_TESTS))
	}

	if m.Race {
		result += fmtc.Sprintf("--%s ", getOptionName(OPT_RACE))
	}

	if options.GetS(OPT_OUTPUT) != "Makefile" {
		result += fmtc.Sprintf(
			"--%s %s ",
			getOptionName(OPT_OUTPUT),
			options.GetS(OPT_OUTPUT),
		)
	}

	result += ".\n\n"

	return result
}

// getOptionName parse option name in options package notation
// and retunr long option name
func getOptionName(opt string) string {
	longOpt, _ := options.ParseOptionName(opt)
	return longOpt
}

// getSeparator return separator
func getSeparator() string {
	return strings.Repeat("#", SEPARATOR_SIZE)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// showUsage print usage info
func showUsage() {
	info := usage.NewInfo("", "dir")

	info.AddOption(OPT_GLIDE, "Add target to fetching dependecies with glide")
	info.AddOption(OPT_DEP, "Add target to fetching dependecies with dep")
	info.AddOption(OPT_METALINTER, "Add target with metalinter check")
	info.AddOption(OPT_STRIP, "Strip binary")
	info.AddOption(OPT_BENCHMARK, "Add target to run benchmarks")
	info.AddOption(OPT_VERB_TESTS, "Enable verbose output for tests")
	info.AddOption(OPT_RACE, "Add target to test race conditions")
	info.AddOption(OPT_OUTPUT, "Output file {s-}(Makefile by default){!}", "file")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VER, "Show version")

	info.AddExample(
		".", "Generate makefile for project in current directory and save as Makefile",
	)

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

// showAbout print info about version
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
