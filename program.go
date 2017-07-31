package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/marstr/collection"
)

var debugLog *log.Logger

type myConst struct {
	Name  string
	Type  ast.Expr
	Value ast.Expr
}

type aliasMaker struct {
	Types  *collection.List
	Consts *collection.List
}

func init() {
	debugWriter := os.Stderr
	//debugWriter := ioutil.Discard
	debugLog = log.New(debugWriter, "[DEBUG] ", 0)
}

func main() {
	exitVal := 1
	defer func() {
		os.Exit(exitVal)
	}()

	var profileName string
	var rootDir string

	goPath := os.Getenv("GOPATH")

	var targetPackages collection.Enumerable

	flag.StringVar(&rootDir, "root", getDefaultRoot(), "The base repository for each ")
	flag.StringVar(&profileName, "profile", "", "The name that should be branded on the generated profile. By default random words are generated.")
	flag.Parse()

	targetPackages = packageFinder{
		root: rootDir,
	}

	if profileName == "" {
		profileName = petname.Generate(3, "-")
		fmt.Println("Profile Name: ", profileName)
	}
	debugLog.Println(profileName)

	for entry := range targetPackages.Enumerate(nil) {
		var err error
		var destination string

		dir, ok := entry.(string)
		if !ok {
			return
		}

		destination, err = getAliasPath(dir, profileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		destination = path.Join(goPath, "src", destination)

		err = createAliasPackage(dir, destination)
		if err != nil {
			debugLog.Print(err)
		}
	}
	exitVal = 0
}

func createAliasPackage(source, destination string) (err error) {
	sourceFiles := &token.FileSet{}

	var packages map[string]*ast.Package
	packages, err = parser.ParseDir(sourceFiles, source, nil, 0)

	if err != nil {
		return
	}
	if packages != nil && len(packages) > 1 {
		err = errors.New("too many packages in target directory")
		return
	}

	os.MkdirAll(destination, os.ModePerm|os.ModeDir)

	outputPath := filepath.Join(destination, "models.go")

	var outputFile *os.File
	outputFile, err = os.Create(outputPath)
	if err != nil {
		return
	}

	// This should only iterate once, but allows us to access this map without
	// knowing the name of the package.
	for _, pkg := range packages {
		createAliasHelper(pkg, source, outputFile)
	}

	return nil
}

func createAliasHelper(original *ast.Package, pkgPath string, output io.Writer) {
	ast.PackageExports(original)

	maker := aliasMaker{
		Types:  collection.NewList(),
		Consts: collection.NewList(),
	}
	ast.Walk(maker, original)

	fmt.Fprintln(output, "package", original.Name)
	fmt.Fprintln(output)

	fmt.Fprintln(output, "import (")
	fmt.Fprintln(output, "\t", "original", "\""+trimGoPath(pkgPath)+"\"")
	fmt.Fprintln(output, ")")
	fmt.Fprintln(output)

	if collection.Any(maker.Types) {
		fmt.Fprintln(output, "type (")
		for t := range maker.Types.Enumerate(nil) {
			name := t.(*ast.TypeSpec).Name.Name
			fmt.Fprintln(output, "\t", name, "=", "original."+name)
		}
		fmt.Fprintln(output, ")")
		fmt.Fprintln(output)
	}

	if collection.Any(maker.Consts) {
		fmt.Fprintln(output, "const (")
		for c := range maker.Consts.Enumerate(nil) {
			name := c.(myConst).Name
			fmt.Fprintln(output, "\t", name, "=", "original."+name)
		}
		fmt.Fprintln(output, ")")
	}
}

func (maker aliasMaker) Visit(node ast.Node) ast.Visitor {
	switch node.(type) {
	case *ast.Package:
		return maker
	case *ast.File:
		return maker
	case *ast.GenDecl:
		cast := node.(*ast.GenDecl)
		if cast.Tok == token.TYPE {
			for _, spec := range cast.Specs {
				maker.Types.Add(spec)
			}
		} else if cast.Tok == token.CONST {
			for _, spec := range cast.Specs {
				valSpec := spec.(*ast.ValueSpec)
				for i, name := range valSpec.Names {
					maker.Consts.Add(myConst{
						Name:  name.Name,
						Type:  valSpec.Type,
						Value: valSpec.Values[i],
					})
				}
			}
		}
	}
	return nil
}

var rawPackagePath = regexp.MustCompile(`github.com/Azure/azure-sdk-for-go/service/(?P<rp>[\w\d\-\.]+)/(?P<management>management/)?(?P<APIVersion>[\w\d\-\.]+)/(?P<resource>[\w\d\-\.]+)`)

// getAliasPath takes an existing API Version path and a package name, and converts the path
// to a path which uses the new profile layout.
func getAliasPath(subject, profile string) (transformed string, err error) {
	subject = strings.TrimSuffix(subject, "/")
	subject = trimGoPath(subject)

	matches := rawPackagePath.FindAllStringSubmatch(subject, -1)
	if matches == nil {
		err = errors.New("path does not resemble a known package path")
		return
	}

	output := []string{
		"github.com",
		"Azure",
		"azure-sdk-for-go",
		"profile",
		profile,
		matches[0][1],
	}

	if matches[0][2] == "management/" {
		output = append(output, "management")
	}

	output = append(output, matches[0][4])

	transformed = strings.Join(output, "/")
	return
}

// trimGoPath removes the prefix defined in the environment variabe GOPATH if it is present in the string provided.
var trimGoPath = func() func(string) string {
	splitGo := strings.Split(os.Getenv("GOPATH"), string(os.PathSeparator))
	splitGo = append(splitGo, "src")

	return func(subject string) string {
		splitPath := strings.Split(subject, string(os.PathSeparator))
		for i, dir := range splitGo {
			if splitPath[i] != dir {
				return subject
			}
		}
		packageIdentifier := splitPath[len(splitGo):]
		return path.Join(packageIdentifier...)
	}
}()

func getDefaultRoot() string {
	return path.Join(os.Getenv("GOPATH"), "src", "github.com", "Azure", "azure-sdk-for-go", "service")
}

type packageFinder struct {
	root string
}

func (finder packageFinder) Enumerate(cancel <-chan struct{}) collection.Enumerator {
	results := make(chan interface{})
	go func() {
		defer close(results)

		filepath.Walk(finder.root, func(localPath string, info os.FileInfo, openErr error) (err error) {
			if !info.IsDir() || openErr != nil {
				return
			}
			if info.Name() == "vendor" {
				err = filepath.SkipDir
				return
			}

			files := &token.FileSet{}

			if pkgs, parseErr := parser.ParseDir(files, localPath, nil, 0); parseErr != nil || len(pkgs) < 1 {
				return
			}

			select {
			case results <- localPath:
				// Intentionally Left Blank
			case <-cancel:
				err = errors.New("enumeration cancelled")
				return
			}

			return
		})
	}()
	return results
}
