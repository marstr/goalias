package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

type packageAliaser ast.Package

type ErrorUnexpectedToken struct {
	Expected token.Token
	Received token.Token
}

func (utoken ErrorUnexpectedToken) Error() string {
	return fmt.Sprintf("Unexpected token %d expecting type: %d", utoken.Received, utoken.Expected)
}

// NewAliasPackage creates
func AliasPackage(original *ast.Package) *ast.Package {
	modelsFile := ast.File{}

	created := &ast.Package{
		Name: original.Name,
		Files: map[string]*ast.File{
			"models.go": &modelsFile,
		},
	}

	return created
}

// AddConst adds a Const block with indiviual aliases for each Spec in `decl`.
func (alias *packageAliaser) AddConst(decl *ast.GenDecl) (err error) {
	if decl.Tok != token.CONST {
		err = ErrorUnexpectedToken{Expected: token.CONST, Received: decl.Tok}
		return
	}

	return
}

// AddType adds a Type delcaration block with individual alias for each Spec handed in `decl`
func (alias *packageAliaser) AddType(decl *ast.GenDecl) (err error) {
	if decl.Tok != token.TYPE {
		err = ErrorUnexpectedToken{Expected: token.TYPE, Received: decl.Tok}
		return
	}

	return
}

func (alias *packageAliaser) AddFunc(decl *ast.FuncDecl) {
	copy := *decl
	copy.Body = &ast.BlockStmt{}
}
