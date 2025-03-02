package core

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type File struct {
	pkg  *Package  // Package to which this file belongs.
	file *ast.File // Parsed AST.
}

type Package struct {
	pkg   *packages.Package
	name  string
	defs  map[*ast.Ident]types.Object
	files []*File
}

type TokenI interface {
	GetName() string
	GetTag() Tag
}

type TokenM interface {
	SetTag(tag Tag)
}

type Token struct {
	Name string
	Tag  Tag
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

func (t Token) GetTag() Tag {
	return t.Tag
}

func (t *Token) SetTag(tag Tag) {
	t.Tag = tag
}
