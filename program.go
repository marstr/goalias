package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

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
	//debugWriter := os.Stderr
	debugWriter := ioutil.Discard
	debugLog = log.New(debugWriter, "[DEBUG] ", 0)
}

func main() {
	for _, dir := range os.Args[1:] {
		err := createAliasPackage(dir, "")
		if err != nil {
			debugLog.Print(err)
		}
	}
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

	var outputFile *os.File
	outputFile, err = os.Create(filepath.Join(destination, "models.go"))
	if err != nil {
		return
	}

	// This should only iterate once, but allows us to separate
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

	fmt.Fprintln(output, "import (")
	fmt.Fprintln(output, "\t", "original", "\""+trimGoPath(pkgPath)+"\"")
	fmt.Fprintln(output, ")")

	if collection.Any(maker.Types) {
		fmt.Fprintln(output, "type (")
		for t := range maker.Types.Enumerate(nil) {
			name := t.(*ast.TypeSpec).Name.Name
			fmt.Fprintln(output, "\t", name, "=", "original."+name)
		}
		fmt.Fprintln(output, ")")
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
				debugLog.Printf("Names: %d Values: %d\n", len(valSpec.Names), len(valSpec.Values))
			}
		}
	}
	return nil
}

// trimGoPath removes the prefix defined in the environment variabe GOPATH if it is present in the string provided.
var trimGoPath = func() func(string) string {
	splitGo := strings.Split(os.Getenv("GOPATH"), string(os.PathSeparator))
	splitGo = append(splitGo, "src")

	return func(subject string) string {
		splitPath := strings.Split(subject, string(os.PathSeparator))
		debugLog.Println(len(splitPath))
		for i, dir := range splitGo {
			if splitPath[i] != dir {
				debugLog.Println(splitPath[i], "!=", dir)
				return subject
			}
		}
		packageIdentifier := splitPath[len(splitGo):]
		debugLog.Println("Joining this: ", packageIdentifier)
		return path.Join(packageIdentifier...)
	}
}()

// getAliasPath takes an existing API Version path and a package name, and converts the path
// to a path which uses the new profile layout.
var getAliasPath = func() func(string, string) (string, error) {

	rawPathPattern := regexp.MustCompile(`github.com/Azure/azure-sdk-for-go/arm/([\w_\-/\\\.]+)/(?:\d{4}-\d{2}-\d{2})(?:-[\w\.\d]+)?/([\w_\-/\\\.]+)`)

	return func(subject, profile string) (string, error) {
		var err error
		subject = strings.TrimSuffix(subject, "/")
		matches := rawPathPattern.FindAllStringSubmatch(subject, -1)
		if matches == nil {
			err = errors.New("path does not resemble a known package path")
			return "", err
		}

		return fmt.Sprint("github.com/Azure/azure-sdk-for-go/arm/profile/", profile, "/", matches[0][1], "/", matches[0][2]), nil
	}
}()
