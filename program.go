package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"

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

func main() {
	debugLog = log.New(os.Stderr, "[DEBUG] ", 0)

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

	for _, pkg := range packages {
		fmt.Println(createAliasHelper(pkg))
	}

	return nil
}

func createAliasHelper(original *ast.Package) string {

	maker := aliasMaker{
		Types:  collection.NewList(),
		Consts: collection.NewList(),
	}
	ast.Walk(maker, original)

	output := &bytes.Buffer{}
	publicTypes := collection.Where(maker.Types, func(x interface{}) bool {

	})
	for t := range maker.Types.Enumerate(nil) {
		name := t.(*ast.TypeSpec).Name.Name
		fmt.Fprintln(output, "type", name, "=", "original."+name)
	}

	return output.String()
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
