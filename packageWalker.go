package main

import (
	"go/ast"

	"github.com/marstr/collection"
)

type PackageWalker struct {
	target ast.Node
}

func (pw PackageWalker) Enumerate(cancel <-chan struct{}) collection.Enumerator {
	results := make(chan interface{})
	go func() {
		defer close(results)
		ast.Inspect(pw.target, func(current ast.Node) bool {
			select {
			case results <- current:
			case <-cancel:
				return false
			}
			return true
		})
	}()
	return results
}
