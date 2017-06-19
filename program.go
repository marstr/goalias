package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"

	"github.com/marstr/collection"
)

var debugLog *log.Logger

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

	maker := aliasMaker{
		Types:  collection.NewList(),
		Consts: collection.NewList(),
	}

	for _, pkg := range packages {
		ast.Walk(maker, pkg)
	}

	typeNames := maker.Types.Enumerate().Select(func(x interface{}) interface{} {
		cast, ok := x.(*ast.TypeSpec)
		if !ok {
			return nil
		}
		return cast.Name.Name
	}).Where(func(x interface{}) bool {
		_, ok := x.(string)
		return ok
	})

	fmt.Println("Types:")
	for entry := range typeNames {
		fmt.Println("\t", entry)
	}

	constNames := maker.Consts.Enumerate().Select(func(x interface{}) interface{} {
		cast, ok := x.(*ast.Const)
	})

	return nil
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
		}
	}
	return nil
}
