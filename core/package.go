package core

import (
	"go/ast"
	"go/types"
	"path"
	"strings"
	"time"

	"golang.org/x/tools/go/packages"
)

type Import struct {
	Name string
	Path string
}

type Package struct {
	Pkg           *packages.Package
	Name          string `hash:""`
	Defs          map[*ast.Ident]types.Object
	Files         []*File
	ModTime       time.Time
	Imports       []Import
	ImportedPkgs  map[string]*Package // Package imported by Pkg
	ImportsByName map[string]int
	ImportsByPath map[string]int

	Types  map[string]TypeI
	Fields []FieldI
	Funcs  map[string][]FuncI
}

func MakePackage(pkg *packages.Package) Package {
	return Package{
		Pkg:           pkg,
		Name:          pkg.Name,
		Defs:          pkg.TypesInfo.Defs,
		Files:         make([]*File, 0, len(pkg.Syntax)),
		Types:         map[string]TypeI{},
		Funcs:         map[string][]FuncI{},
		ImportedPkgs:  map[string]*Package{},
		ImportsByName: map[string]int{},
		ImportsByPath: map[string]int{},
	}
}

func NewPackage(pkg *packages.Package) *Package {
	p := MakePackage(pkg)
	return &p
}

func (p *Package) Above(pkg *Package) bool {
	return strings.HasPrefix(pkg.Pkg.PkgPath, p.Pkg.PkgPath)
}

func (p *Package) AddImport(decl *ast.ImportSpec) {
	pkg_path := strings.Trim(decl.Path.Value, "\"")
	_, pkg_name := path.Split(pkg_path)
	if decl.Name != nil {
		pkg_name = decl.Name.Name
	}

	if _, ok := p.ImportsByPath[pkg_path]; !ok {
		p.Imports = append(p.Imports, Import{
			Name: pkg_name,
			Path: pkg_path,
		})
		p.ImportsByName[pkg_name] = len(p.Imports)
		p.ImportsByPath[pkg_path] = len(p.Imports)
	}
}

func (p *Package) HasImport(pkg *Package) bool {
	_, ok := p.ImportsByPath[pkg.Pkg.PkgPath]
	return ok
}
