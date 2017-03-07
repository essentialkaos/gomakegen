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
	"pkg.re/essentialkaos/ek.v7/usage"
	"pkg.re/essentialkaos/ek.v7/usage/update"
)

// ////////////////////////////////////////////////////////////////////////////////// //

const (
	APP  = "gomakegen"
	VER  = "0.0.1"
	DESC = "Utility for generating makefiles for golang applications"
)

const (
	ARG_OUTPUT   = "o:output"
	ARG_NO_COLOR = "nc:no-color"
	ARG_HELP     = "h:help"
	ARG_VER      = "v:version"
)

const SEPARATOR = "########################################################################################"

// ////////////////////////////////////////////////////////////////////////////////// //

var argMap = arg.Map{
	ARG_OUTPUT:   {Value: "Makefile"},
	ARG_NO_COLOR: {Type: arg.BOOL},
	ARG_HELP:     {Type: arg.BOOL, Alias: "u:usage"},
	ARG_VER:      {Type: arg.BOOL, Alias: "ver"},
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

	bImports, tImports, binaries := collectImports(sources, dir)

	bImports = cleanupImports(bImports, dir)
	tImports = cleanupImports(tImports, dir)
	binaries = cleanupBinaries(binaries)

	data := generateMakefile(bImports, tImports, binaries)

	err := ioutil.WriteFile(arg.GetS(ARG_OUTPUT), []byte(data), 0644)

	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	fmtc.Printf("{g}Makefile successfully created as {g*}%s{!}\n", arg.GetS(ARG_OUTPUT))
}

// generateMakefile generate makefile data
func generateMakefile(bImports, tImports, binaries []string) string {
	var data string

	data += SEPARATOR + "\n\n"
	data += ".PHONY = " + generatePhonyPart(bImports, tImports, binaries) + "\n\n"
	data += SEPARATOR + "\n\n"

	// build section
	if len(binaries) != 0 {
		data += "all: " + strings.Join(binaries, " ") + "\n\n"

		for _, bin := range binaries {
			data += bin + ":\n"
			data += "\tgo build " + bin + ".go\n\n"
		}
	}

	// deps section
	if len(bImports) != 0 {
		data += "deps:\n"

		for _, pkg := range bImports {
			data += "\tgo get -v " + pkg + "\n"
		}

		data += "\n"
	}

	// deps-test and test sections
	if len(tImports) != 0 {
		data += "deps-test:\n"

		for _, pkg := range tImports {
			data += "\tgo get -v " + pkg + "\n"
		}

		data += "\n"
		data += "test:\n"
		data += "\tgo test -covermode=count .\n"
		data += "\n"
	}

	// fmt section
	data += "fmt:\n"
	data += "\tfind . -name \"*.go\" -exec gofmt -s -w {} \\;\n"
	data += "\n"

	// clean section
	if len(binaries) != 0 {
		data += "clean:\n"

		for _, bin := range binaries {
			data += "\trm -f " + bin + "\n"
		}

		data += "\n"
	}

	data += SEPARATOR + "\n\n"

	return data
}

// generatePhonyPart return phony part for Makefile
func generatePhonyPart(bImports, tImports, binaries []string) string {
	phony := []string{"fmt"}

	if len(binaries) != 0 {
		phony = append(phony, "all", "clean")
	}

	if len(bImports) != 0 {
		phony = append(phony, "deps")
	}

	if len(tImports) != 0 {
		phony = append(phony, "deps-test", "test")
	}

	return strings.Join(phony, " ")
}

// collectImports collect import from source files and return imports for
// base sources, test sources and slice with binaries
func collectImports(sources []string, dir string) ([]string, []string, []string) {
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

	return importMapToSlice(bImports), importMapToSlice(tImports), binaries
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

// printError prints error message to console
func printError(f string, a ...interface{}) {
	fmtc.Printf("{r}"+f+"{!}\n", a...)
}

// printWarn prints warning message to console
func printWarn(f string, a ...interface{}) {
	fmtc.Printf("{y}"+f+"{!}\n", a...)
}

// ////////////////////////////////////////////////////////////////////////////////// //

//
func showUsage() {
	info := usage.NewInfo("", "dir")

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
