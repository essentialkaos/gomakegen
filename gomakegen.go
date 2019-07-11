package main

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2019 ESSENTIAL KAOS                         //
//        Essential Kaos Open Source License <https://essentialkaos.com/ekol>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"go/parser"
	"go/token"

	"pkg.re/essentialkaos/ek.v10/fmtc"
	"pkg.re/essentialkaos/ek.v10/fsutil"
	"pkg.re/essentialkaos/ek.v10/mathutil"
	"pkg.re/essentialkaos/ek.v10/options"
	"pkg.re/essentialkaos/ek.v10/path"
	"pkg.re/essentialkaos/ek.v10/sliceutil"
	"pkg.re/essentialkaos/ek.v10/strutil"
	"pkg.re/essentialkaos/ek.v10/usage"
	"pkg.re/essentialkaos/ek.v10/usage/update"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// App info
const (
	APP  = "gomakegen"
	VER  = "1.1.2"
	DESC = "Utility for generating makefiles for Go applications"
)

// Constants with options names
const (
	OPT_OUTPUT     = "o:output"
	OPT_GLIDE      = "g:glide"
	OPT_DEP        = "d:dep"
	OPT_MOD        = "m:mod"
	OPT_METALINTER = "M:metalinter"
	OPT_STRIP      = "S:strip"
	OPT_BENCHMARK  = "B:benchmark"
	OPT_RACE       = "R:race"
	OPT_VERB_TESTS = "V:verbose"
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

	FuzzPaths []string
	TestPaths []string

	PkgBase string

	MaxTargetNameSize int

	HasTests         bool
	Benchmark        bool
	VerbTests        bool
	Race             bool
	Metalinter       bool
	Strip            bool
	HasSubpackages   bool
	HasStableImports bool

	GlideUsed bool
	DepUsed   bool
	ModUsed   bool
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Options map
var optMap = options.Map{
	OPT_OUTPUT:     {Value: "Makefile"},
	OPT_GLIDE:      {Type: options.BOOL},
	OPT_DEP:        {Type: options.BOOL},
	OPT_MOD:        {Type: options.BOOL},
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
	runtime.GOMAXPROCS(2)

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

// checkDir checks directory with sources
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

// process starts sources processing
func process(dir string) {
	sources := fsutil.ListAllFiles(
		dir, true,
		fsutil.ListingFilter{
			MatchPatterns: []string{"*.go"},
			SizeGreater:   1, // Ignore empty files
		},
	)

	sources = filterSources(sources)
	makefile := generateMakefile(sources, dir)

	exportMakefile(makefile)
}

// filterSources removes sources from vendor directory from sources list
func filterSources(sources []string) []string {
	var result []string

	for _, source := range sources {
		if strings.HasPrefix(source, "vendor/") {
			continue
		}

		result = append(result, source)
	}

	return result
}

// exportMakefile renders makefile and write data to file
func exportMakefile(makefile *Makefile) {
	err := ioutil.WriteFile(options.GetS(OPT_OUTPUT), makefile.Render(), 0644)

	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	fmtc.Printf("{g}Makefile successfully created as {g*}%s{!}\n", options.GetS(OPT_OUTPUT))
}

// generateMakefile collects imports, process options and generate makefile struct
func generateMakefile(sources []string, dir string) *Makefile {
	makefile := collectImports(sources, dir)

	applyOptionsFromMakefile(dir+"/"+options.GetS(OPT_OUTPUT), makefile)

	makefile.Metalinter = makefile.Metalinter || options.GetB(OPT_METALINTER)
	makefile.Benchmark = makefile.Benchmark || options.GetB(OPT_BENCHMARK)
	makefile.VerbTests = makefile.VerbTests || options.GetB(OPT_VERB_TESTS)
	makefile.Race = makefile.Race || options.GetB(OPT_RACE)
	makefile.Strip = makefile.Strip || options.GetB(OPT_STRIP)
	makefile.GlideUsed = makefile.GlideUsed || options.GetB(OPT_GLIDE) || fsutil.IsExist(dir+"/glide.yaml")
	makefile.DepUsed = makefile.DepUsed || options.GetB(OPT_DEP) || fsutil.IsExist(dir+"/Gopkg.toml")
	makefile.ModUsed = makefile.ModUsed || options.GetB(OPT_MOD) || fsutil.IsExist(dir+"/go.mod")

	makefile.HasStableImports = containsStableImports(makefile.BaseImports)
	makefile.HasStableImports = makefile.HasStableImports || containsStableImports(makefile.TestImports)

	makefile.Cleanup(dir)

	return makefile
}

// collectImports collects import from source files and returns imports for
// base sources, test sources and slice with binaries
func collectImports(sources []string, dir string) *Makefile {
	baseSources, testSources := splitSources(sources)

	baseImports, binaries, hasSubPkgs := extractBaseImports(baseSources, dir)
	testImports, testPaths := extractTestImports(testSources, dir)
	fuzzPaths := collectFuzzPaths(baseSources, dir)

	return &Makefile{
		BaseImports:    baseImports,
		TestImports:    testImports,
		FuzzPaths:      fuzzPaths,
		TestPaths:      testPaths,
		PkgBase:        getBasePkgPath(dir),
		Binaries:       binaries,
		HasTests:       hasTests(sources),
		HasSubpackages: hasSubPkgs,
	}
}

// splitSources splits sources to two slices - with base sources and test sources
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

// extractBaseImports extracts base imports from given source files
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

// extractTestImports extracts test imports from given source files
func extractTestImports(sources []string, dir string) ([]string, []string) {
	if len(sources) == 0 {
		return nil, nil
	}

	importsMap := make(map[string]bool)
	testPaths := make(map[string]bool)

	for _, source := range sources {
		imports, _ := extractImports(source, dir)
		basePath := path.Dir(source)

		for _, path := range imports {
			importsMap[path] = true
		}

		testPaths["./"+basePath] = true
	}

	return importMapToSlice(importsMap), importMapToSlice(testPaths)
}

// collectFuzzPaths collects paths with fuzz tests
func collectFuzzPaths(sources []string, dir string) []string {
	var result []string

	for _, source := range sources {
		if hasFuzzTests(source, dir) {
			result = append(result, path.Dir(source))
		}
	}

	return result
}

// cleanupImports removes internal packages and local imports
func cleanupImports(imports []string, dir string) []string {
	if len(imports) == 0 {
		return nil
	}

	result := make(map[string]bool)
	gopath := os.Getenv("GOPATH")
	basePath := getBasePkgPath(dir)

	for _, imp := range imports {
		if !isExternalPackage(imp) {
			continue
		}

		if strings.HasPrefix(imp, basePath) {
			continue
		}

		result[getPackageRoot(imp, gopath)] = true
	}

	return importMapToSlice(result)
}

// cleanupBinaries removes .go suffix from names of binaries
func cleanupBinaries(binaries []string) []string {
	var result []string

	for _, bin := range binaries {
		result = append(result, strings.Replace(bin, ".go", "", -1))
	}

	return result
}

// extractImports returns slice with all imports in source file
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

// hasFuzzTest returns true if given source package
func hasFuzzTests(source, dir string) bool {
	fset := token.NewFileSet()
	file := path.Join(dir, source)
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)

	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	if len(f.Comments) == 0 {
		return false
	}

	return strings.Contains(f.Comments[0].Text(), "+build gofuzz")
}

// hasTests returns true if project has tests
func hasTests(sources []string) bool {
	for _, source := range sources {
		if isTestSource(source) {
			return true
		}
	}

	return false
}

// isTestSource returns true if given file is tests
func isTestSource(source string) bool {
	return strings.HasSuffix(source, "_test.go")
}

// getPackageRoot returns root for package
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

// isPackageRoot returns true if given path is root for package
func isPackageRoot(path string) bool {
	if !fsutil.IsExist(path + "/.git") {
		return false
	}

	files := fsutil.List(path, true, fsutil.ListingFilter{MatchPatterns: []string{"*.go"}})

	return len(files) != 0
}

// isExternalPackage returns true if given package is external
func isExternalPackage(pkg string) bool {
	pkgSlice := strings.Split(pkg, "/")

	if len(pkgSlice) == 0 || !strings.Contains(pkgSlice[0], ".") {
		return false
	}
	return true
}

// importMapToSlice converts map with package names to string slice
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

// containsPackage returns true if imports contains given packages
func containsPackage(imports []string, pkgs []string) bool {
	for _, pkg := range pkgs {
		if sliceutil.Contains(imports, pkg) {
			return true
		}
	}

	return false
}

// getBasePkgPath returns base package path
func getBasePkgPath(dir string) string {
	gopath := os.Getenv("GOPATH")
	absDir, _ := filepath.Abs(dir)

	return strutil.Exclude(absDir, gopath+"/src/")
}

// containsStableImports returns true if imports contains stable import services path
func containsStableImports(imports []string) bool {
	if len(imports) == 0 {
		return false
	}

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

// getGitConfigurationForStableImports returns slice with git configuration commands for
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

// applyOptionsFromFile reads used options from previously generated Makefile
// and applies it to makefile struct
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

// extractOptionsFromMakefile extracts options from previously generated Makefile
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

// Cleanup cleans imports and binaries
func (m *Makefile) Cleanup(dir string) {
	m.BaseImports = cleanupImports(m.BaseImports, dir)
	m.TestImports = cleanupImports(m.TestImports, dir)

	m.Binaries = cleanupBinaries(m.Binaries)

	sort.Strings(m.Binaries)
}

// Render returns makefile data
func (m *Makefile) Render() []byte {
	var result string

	result += m.getHeader()
	result += m.getTargets()

	return []byte(result)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// getHeader returns header data
func (m *Makefile) getHeader() string {
	var result string

	result += getSeparator() + "\n\n"
	result += m.getGenerationComment()
	result += getSeparator() + "\n\n"

	if m.ModUsed {
		result += "export GO111MODULE=on\n\n"
	}

	result += m.getDefaultGoal() + "\n"
	result += m.getPhony() + "\n"
	result += getSeparator() + "\n\n"

	return result
}

// getTargets returns targets data
func (m *Makefile) getTargets() string {
	var result string

	result += m.getBinTarget()
	result += m.getInstallTarget()
	result += m.getUninstallTarget()
	result += m.getGitConfigTarget()
	result += m.getDepsTarget()
	result += m.getTestDepsTarget()
	result += m.getTestTarget()
	result += m.getFuzzTarget()
	result += m.getBenchTarget()
	result += m.getGlideTarget()
	result += m.getDepTarget()
	result += m.getModTarget()
	result += m.getFmtTarget()
	result += m.getMetalinterTarget()
	result += m.getCleanTarget()
	result += m.getHelpTarget()

	result += getSeparator() + "\n"

	return result
}

// codebeat:disable[ABC]

// getPhony returns PHONY part of makefile
func (m *Makefile) getPhony() string {
	phony := []string{"fmt"}

	if len(m.Binaries) != 0 {
		phony = append(phony, "all", "clean")
	}

	if m.HasStableImports {
		phony = append(phony, "git-config")
	}

	if len(m.BaseImports) != 0 {
		phony = append(phony, "deps")
	}

	if len(m.TestImports) != 0 {
		phony = append(phony, "deps-test", "test")
	}

	if len(m.FuzzPaths) != 0 {
		phony = append(phony, "gen-fuzz")
	}

	if m.GlideUsed {
		phony = append(phony, "glide-create", "glide-install", "glide-update")
	}

	if m.DepUsed {
		phony = append(phony, "dep-init", "dep-update")
	}

	if m.ModUsed {
		phony = append(phony, "mod-init", "mod-update")
	}

	if m.Benchmark {
		phony = append(phony, "benchmark")
	}

	if m.Metalinter {
		phony = append(phony, "metalinter")
	}

	phony = append(phony, "help")

	for _, target := range append(phony, m.Binaries...) {
		m.MaxTargetNameSize = mathutil.Max(m.MaxTargetNameSize, len(target)+2)
	}

	return ".PHONY = " + strings.Join(phony, " ") + "\n"
}

// codebeat:enable[ABC]

// getDefaultGoal returns DEFAULT_GOAL part of makefile
func (m *Makefile) getDefaultGoal() string {
	return ".DEFAULT_GOAL := help"
}

// getBinTarget generates target for "all" command and all sub-targets
// for each binary
func (m *Makefile) getBinTarget() string {
	if len(m.Binaries) == 0 {
		return ""
	}

	result := "all: " + strings.Join(m.Binaries, " ") + " ## Build all binaries\n\n"

	for _, bin := range m.Binaries {
		result += bin + ": ## " + fmt.Sprintf("Build %s binary", bin) + "\n"

		if m.Strip {
			result += "\tgo build -ldflags=\"-s -w\" " + bin + ".go\n"
		} else {
			result += "\tgo build " + bin + ".go\n"
		}

		result += "\n"
	}

	return result
}

// getInstallTarget generates target for "install" command
func (m *Makefile) getInstallTarget() string {
	if len(m.Binaries) == 0 {
		return ""
	}

	result := "install: ## Install all binaries\n"

	for _, bin := range m.Binaries {
		result += "\tcp " + bin + " /usr/bin/" + bin + "\n"
	}

	return result + "\n"
}

// getUninstallTarget generates target for "uninstall" command
func (m *Makefile) getUninstallTarget() string {
	if len(m.Binaries) == 0 {
		return ""
	}

	result := "uninstall: ## Uninstall all binaries\n"

	for _, bin := range m.Binaries {
		result += "\trm -f /usr/bin/" + bin + "\n"
	}

	return result + "\n"
}

// getGitConfigTarget generates target for "git-config" command
func (m *Makefile) getGitConfigTarget() string {
	if !m.HasStableImports {
		return ""
	}

	imports := append(m.BaseImports[:0:0], m.BaseImports...)
	imports = append(imports, m.TestImports...)

	result := "git-config: ## Configure git redirects for stable import path services\n"

	for _, gitCommand := range getGitConfigurationForStableImports(imports) {
		result += "\t" + gitCommand + "\n"
	}

	return result + "\n"
}

// getDepsTarget generates target for "deps" command
func (m *Makefile) getDepsTarget() string {
	if len(m.BaseImports) == 0 {
		return ""
	}

	result := "deps: "

	if m.HasStableImports {
		result += "git-config "
	}

	switch {
	case m.GlideUsed:
		result += "glide-install "
	case m.DepUsed:
		result += "dep-update "
	case m.ModUsed:
		result += "mod-download "
	}

	result += "## Download dependencies\n"

	if m.GlideUsed || m.DepUsed || m.ModUsed {
		return result + "\n"
	}

	for _, pkg := range m.BaseImports {
		result += "\tgo get -d -v " + pkg + "\n"
	}

	return result + "\n"
}

// getDepsTarget generates target for "deps-test" command
func (m *Makefile) getTestDepsTarget() string {
	if len(m.TestImports) == 0 {
		return ""
	}

	pkgMngUsed := m.DepUsed || m.GlideUsed || m.ModUsed

	result := "deps-test: "

	if pkgMngUsed {
		result += "deps "
	} else {
		if m.HasStableImports {
			result += "git-config "
		}
	}

	result += "## Download dependencies for tests\n"

	if !pkgMngUsed {
		for _, pkg := range m.TestImports {
			result += "\tgo get -d -v " + pkg + "\n"
		}
	}

	return result + "\n"
}

// getTestTarget generates target for "test" command
func (m *Makefile) getTestTarget() string {
	if !m.HasTests {
		return ""
	}

	targets := "."

	if m.HasSubpackages {
		targets = strings.Join(m.TestPaths, " ")
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

// getFuzzTarget generates target for "fuzz" command
func (m *Makefile) getFuzzTarget() string {
	if len(m.FuzzPaths) == 0 {
		return ""
	}

	result := "gen-fuzz: ## Generate archives for fuzz testing\n"
	result += "\twhich go-fuzz-build &>/dev/null || go get -u -v github.com/dvyukov/go-fuzz/go-fuzz-build\n"

	for _, pkg := range m.FuzzPaths {
		if pkg == "." {
			result += fmt.Sprintf("\tgo-fuzz-build -o fuzz.zip %s\n", m.PkgBase)
		} else {
			pkgName := strings.Replace(pkg, "/", "-", -1)
			binName := pkgName + "-fuzz.zip"

			result += fmt.Sprintf("\tgo-fuzz-build -o %s %s/%s\n", binName, m.PkgBase, pkg)
		}
	}

	return result + "\n"
}

// getBenchTarget generates target for "benchmark" command
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

// getFmtTarget generates target for "fmt" command
func (m *Makefile) getFmtTarget() string {
	result := "fmt: ## Format source code with gofmt\n"
	result += "\tfind . -name \"*.go\" -exec gofmt -s -w {} \\;\n"

	return result + "\n"
}

// getCleanTarget generates target for "clean" command
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

// getGlideTarget generates target for "glide-install" and
// "glide-update" commands
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
	result += "\tglide update\n\n"

	return result
}

// getDepTarget generate target for "dep-init" and "dep-update" commands
func (m *Makefile) getDepTarget() string {
	if !m.DepUsed {
		return ""
	}

	result := "dep-init: ## Initialize dep workspace\n"
	result += "\twhich dep &>/dev/null || go get -u -v github.com/golang/dep/cmd/dep\n"
	result += "\tdep init\n\n"

	result += "dep-update: ## Update packages and dependencies through dep\n"
	result += "\twhich dep &>/dev/null || go get -u -v github.com/golang/dep/cmd/dep\n"
	result += "\ttest -s Gopkg.toml || dep init\n"
	result += "\ttest -s Gopkg.lock && dep ensure -update || dep ensure\n\n"

	return result
}

// getModTarget generates target for "mod-init" and "mod-update" commands
func (m *Makefile) getModTarget() string {
	if !m.ModUsed {
		return ""
	}

	result := "mod-init: ## Initialize new module\n"
	result += "\tgo mod init\n\n"

	result += "mod-update: ## Download modules to local cache\n"
	result += "\tgo mod download\n\n"

	return result
}

// getMetalinterTarget generates target for "metalineter" command
func (m *Makefile) getMetalinterTarget() string {
	if !m.Metalinter {
		return ""
	}

	result := "metalinter: ## Install and run gometalinter\n"
	result += "\ttest -s $(GOPATH)/bin/gometalinter || (go get -u github.com/alecthomas/gometalinter ; $(GOPATH)/bin/gometalinter --install)\n"
	result += "\t$(GOPATH)/bin/gometalinter --deadline 30s\n\n"

	return result
}

// getHelpTarget generates target for "help" command
func (m *Makefile) getHelpTarget() string {
	fmtSize := strconv.Itoa(m.MaxTargetNameSize)

	result := "help: ## Show this info\n"
	result += "\t@echo -e '\\n\\033[1mSupported targets:\\033[0m\\n'\n"
	result += "\t@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \\\n"
	result += "\t\t| awk 'BEGIN {FS = \":.*?## \"}; {printf \"  \\033[33m%-" + fmtSize + "s\\033[0m %s\\n\", $$1, $$2}'\n"
	result += "\t@echo -e ''\n"
	result += "\t@echo -e '\\033[90mGenerated by GoMakeGen " + VER + "\\033[0m\\n'\n\n"

	return result
}

// getGenerationComment returns comment with all used flags
func (m *Makefile) getGenerationComment() string {
	result := "# This Makefile generated by GoMakeGen " + VER + " using next command:\n"
	result += "# gomakegen "

	if m.GlideUsed {
		result += fmt.Sprintf("--%s ", getOptionName(OPT_GLIDE))
	}

	if m.DepUsed {
		result += fmt.Sprintf("--%s ", getOptionName(OPT_DEP))
	}

	if m.ModUsed {
		result += fmt.Sprintf("--%s ", getOptionName(OPT_MOD))
	}

	if m.Metalinter {
		result += fmt.Sprintf("--%s ", getOptionName(OPT_METALINTER))
	}

	if m.Strip {
		result += fmt.Sprintf("--%s ", getOptionName(OPT_STRIP))
	}

	if m.Benchmark {
		result += fmt.Sprintf("--%s ", getOptionName(OPT_BENCHMARK))
	}

	if m.VerbTests {
		result += fmt.Sprintf("--%s ", getOptionName(OPT_VERB_TESTS))
	}

	if m.Race {
		result += fmt.Sprintf("--%s ", getOptionName(OPT_RACE))
	}

	if options.GetS(OPT_OUTPUT) != "Makefile" {
		result += fmt.Sprintf(
			"--%s %s ",
			getOptionName(OPT_OUTPUT),
			options.GetS(OPT_OUTPUT),
		)
	}

	result += ".\n"
	result += "#\n"
	result += "# More info: https://kaos.sh/gomakegen\n\n"

	return result
}

// getOptionName parse option name in options package notation
// and retunr long option name
func getOptionName(opt string) string {
	longOpt, _ := options.ParseOptionName(opt)
	return longOpt
}

// getSeparator returns separator
func getSeparator() string {
	return strings.Repeat("#", SEPARATOR_SIZE)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// showUsage print usage info
func showUsage() {
	info := usage.NewInfo("", "dir")

	info.AddOption(OPT_GLIDE, "Add target to fetching dependecies with glide")
	info.AddOption(OPT_DEP, "Add target to fetching dependecies with dep")
	info.AddOption(OPT_MOD, "Add target to fetching dependecies with go mod")
	info.AddOption(OPT_METALINTER, "Add target with metalinter check")
	info.AddOption(OPT_STRIP, "Strip binaries")
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
