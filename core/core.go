package core

import (
	"go/ast"
	"time"
)

type File struct {
	Pkg     *Package  // Package to which this file belongs.
	File    *ast.File // Parsed AST.
	ModTime time.Time
}
