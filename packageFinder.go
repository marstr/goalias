package main

import (
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/marstr/collection"
)

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
