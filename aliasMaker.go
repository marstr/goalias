package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"

	"github.com/marstr/collection"
)

type AliasPackage ast.Package

type ErrorUnexpectedToken struct {
	Expected token.Token
	Received token.Token
}

func (utoken ErrorUnexpectedToken) Error() string {
	return fmt.Sprintf("Unexpected token %d expecting type: %d", utoken.Received, utoken.Expected)
}

const modelFile = "models.go"
const origImportAlias = "original"

// ModelFile is a getter for the file accumulating aliased content.
func (alias AliasPackage) ModelFile() (result *ast.File) {
	if alias.Files != nil {
		result = alias.Files[modelFile]
	}
	return
}

// SetModelFile is a setter for the file accumulating aliased content.
func (alias *AliasPackage) SetModelFile(val *ast.File) {
	if alias.Files == nil {
		alias.Files = make(map[string]*ast.File)
	}

	alias.Files[modelFile] = val
}

// NewAliasPackage stuff and things
func NewAliasPackage(original *ast.Package) (alias *AliasPackage, err error) {
	const buildTag = "// +build go1.9"
	models := &ast.File{
		Name: &ast.Ident{
			Name: original.Name,
		},
		Package: token.Pos(len(buildTag) + 1),
	}

	alias = &AliasPackage{
		Name: original.Name,
		Files: map[string]*ast.File{
			modelFile: models,
		},
	}

	models.Comments = append(models.Comments, &ast.CommentGroup{
		List: []*ast.Comment{
			&ast.Comment{
				Text: buildTag,
			},
		},
	})

	models.Decls = append(models.Decls, &ast.GenDecl{
		Tok: token.IMPORT,
		Specs: []ast.Spec{
			&ast.ImportSpec{
				Name: &ast.Ident{
					Name: origImportAlias,
				},
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("\"%s\"", original.Name),
				},
			},
		},
	})

	walker := PackageWalker{target: original}

	generalDecls := collection.Where(walker, func(x interface{}) (ok bool) {
		_, ok = x.(*ast.GenDecl)
		return
	})

	for item := range generalDecls.Enumerate(nil) {
		alias.AddGeneral(item.(*ast.GenDecl))
	}

	funcDecls := collection.Where(walker, func(x interface{}) (ok bool) {
		_, ok = x.(*ast.FuncDecl)
		return
	})

	for item := range funcDecls.Enumerate(nil) {
		alias.AddFunc(item.(*ast.FuncDecl))
	}

	return
}

func (alias *AliasPackage) AddGeneral(decl *ast.GenDecl) error {
	var adder func(*ast.GenDecl) error

	switch decl.Tok {
	case token.CONST:
		adder = alias.AddConst
	case token.TYPE:
		adder = alias.AddType
	}

	return adder(decl)
}

// AddConst adds a Const block with indiviual aliases for each Spec in `decl`.
func (alias *AliasPackage) AddConst(decl *ast.GenDecl) (err error) {
	if decl == nil {
		err = errors.New("unexpected nil")
		return
	} else if decl.Tok != token.CONST {
		err = ErrorUnexpectedToken{Expected: token.CONST, Received: decl.Tok}
		return
	}

	targetFile := alias.ModelFile()

	for _, spec := range decl.Specs {
		cast := spec.(*ast.ValueSpec)
		for j, name := range cast.Names {
			cast.Values[j] = &ast.SelectorExpr{
				X: &ast.Ident{
					Name: origImportAlias,
				},
				Sel: &ast.Ident{
					Name: name.Name,
				},
			}
		}
	}

	targetFile.Decls = append(targetFile.Decls, decl)

	return
}

// AddType adds a Type delcaration block with individual alias for each Spec handed in `decl`
func (alias *AliasPackage) AddType(decl *ast.GenDecl) (err error) {
	if decl == nil {
		err = errors.New("unexpected nil")
		return
	} else if decl.Tok != token.TYPE {
		err = ErrorUnexpectedToken{Expected: token.TYPE, Received: decl.Tok}
		return
	}

	targetFile := alias.ModelFile()

	for _, spec := range decl.Specs {
		cast := spec.(*ast.TypeSpec)
		cast.Assign = 0
		cast.Type = &ast.SelectorExpr{
			X: &ast.Ident{
				Name: origImportAlias,
			},
			Sel: &ast.Ident{
				Name: cast.Name.Name,
			},
		}
	}

	targetFile.Decls = append(targetFile.Decls, decl)
	return
}

func (alias *AliasPackage) AddFunc(decl *ast.FuncDecl) {
	copy := *decl
	copy.Body = &ast.BlockStmt{}
}
