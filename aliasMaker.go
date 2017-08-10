package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
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

func NewAliasPackage(original *ast.Package) (alias *AliasPackage, err error) {
	alias = &AliasPackage{
		Name: original.Name,
		Files: map[string]*ast.File{
			modelFile: &ast.File{
				Name: &ast.Ident{
					Name: "models",
				},
			},
		},
	}

	return
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

	targetFile, ok := alias.Files[modelFile]
	if !ok {
		targetFile = &ast.File{
			Name: &ast.Ident{
				Name: "models",
			},
		}
		alias.Files[modelFile] = targetFile
	}
	targetFile.Decls = append(targetFile.Decls, decl)

	return
}

// AddType adds a Type delcaration block with individual alias for each Spec handed in `decl`
func (alias *AliasPackage) AddType(decl *ast.GenDecl) (err error) {
	if decl.Tok != token.TYPE {
		err = ErrorUnexpectedToken{Expected: token.TYPE, Received: decl.Tok}
		return
	}

	return
}

func (alias *AliasPackage) AddFunc(decl *ast.FuncDecl) {
	copy := *decl
	copy.Body = &ast.BlockStmt{}
}
