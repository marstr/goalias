package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"testing"

	"github.com/marstr/collection"
)

func TestAliasMaker_AddConst(t *testing.T) {
	parsedFiles := &token.FileSet{}
	fileContents, err := parser.ParseFile(parsedFiles, "testdata/publicconst.go", nil, 0)

	if err != nil {
		t.Error(err)
	}

	wrapper := &ast.Package{
		Name: "wrapper",
		Files: map[string]*ast.File{
			"testdata/publicconst.go": fileContents,
		},
	}

	subject, err := NewAliasPackage(wrapper)
	if err != nil {
		t.Error(err)
	}
	var constDecls collection.Enumerable = PackageWalker{target: fileContents}

	constDecls = collection.Where(constDecls, func(x interface{}) bool {
		cast, ok := x.(*ast.GenDecl)
		return ok && cast.Tok == token.CONST
	})
	constDecls = collection.Reverse(constDecls)

	for entry := range constDecls.Enumerate(nil) {
		if err := subject.AddConst(entry.(*ast.GenDecl)); err == nil {
			t.Log("Logging Block")
			for _, namedItem := range entry.(*ast.GenDecl).Specs {
				t.Log("Added Const: ", namedItem)
			}
		} else {
			t.Error(err)
		}
	}

	result := &bytes.Buffer{}
	t.Log(*subject.Files["models.go"])
	for key, f := range subject.Files {
		t.Log("Printing File Contents for:", key)
		err = printer.Fprint(result, parsedFiles, f)
		if err != nil {
			t.Error(err)
		}
	}
	//ast.Fprint(result, parsedFiles, subject.Package.Files[subject.aliasFile], nil)
	t.Log(result.String())
}
