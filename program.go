package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
)

var (
	output  io.Writer
	subject *ast.Package
)

func main() {
	var err error
	exitStatus := 1
	defer func() {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(exitStatus)
		}
	}()

	aliased, err := NewAliasPackage(subject)
	if err != nil {
		return
	}

	var files token.FileSet

	//err = format.Node(output, &files, aliased.ModelFile())
	err = printer.Fprint(output, &files, aliased.ModelFile())
	if err != nil {
		return
	}
}

func init() {
	var outputLocation string
	var inputLocation string
	flag.StringVar(&outputLocation, "o", "", "The name of the output file that should be generated.")
	flag.StringVar(&inputLocation, "i", "", "The file or directory containing Go source to be aliased.")
	flag.Parse()

	if outputLocation == "" {
		output = os.Stdout
	} else {
		var err error
		output, err = os.Create(outputLocation)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	var files token.FileSet

	selectedMode := parser.ParseComments

	var fauxPackage ast.Package
	fauxPackage.Name = "faux"
	fauxPackage.Files = make(map[string]*ast.File)

	if inputLocation == "" {
		const filename = "source.go"
		source, err := parser.ParseFile(&files, filename, os.Stdin, selectedMode)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fauxPackage.Files[filename] = source
		subject = &fauxPackage
	} else if inputInfo, err := os.Stat(inputLocation); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if inputInfo.IsDir() {
		packages, err := parser.ParseDir(&files, inputLocation, nil, selectedMode)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		for _, v := range packages {
			subject = v
			break
		}
	} else {
		source, err := parser.ParseFile(&files, inputLocation, nil, selectedMode)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fauxPackage.Files[inputLocation] = source
		subject = &fauxPackage
	}

}
