package core

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type File struct {
	Pkg  *Package  // Package to which this file belongs.
	File *ast.File // Parsed AST.
}

type Package struct {
	Pkg   *packages.Package
	Name  string
	Defs  map[*ast.Ident]types.Object
	Files []*File

	Types map[string]TypeI
	Funcs map[string][]FuncI
}

type TokenI interface {
	GetName() string
	GetFullName() string
	GetTag() Tag
	GetPackage() *Package
	SetPackage(pkg *Package)
}

type TokenM interface {
	SetTag(tag Tag)
}

type Token struct {
	Name    string
	Tag     Tag
	Package *Package
}

func MakeToken(name string, tag Tag) Token {
	return Token{
		Name: name,
		Tag:  tag,
	}
}

func (t Token) GetName() string {
	return t.Name
}

func (t Token) GetFullName() string {
	return t.Package.Name + "." + t.Name
}

func (t Token) GetTag() Tag {
	return t.Tag
}

func (t *Token) SetTag(tag Tag) {
	t.Tag = tag
}

func (t Token) GetPackage() *Package {
	return t.Package
}

func (t *Token) SetPackage(pkg *Package) {
	t.Package = pkg
}
