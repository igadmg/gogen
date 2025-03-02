package core

import (
	"fmt"
	"go/ast"
	"iter"

	"github.com/igadmg/goex/astex"
)

type Parameter struct {
	Name     string
	Type     string
	DeclType string
}

func AstFieldName(decl *ast.Field) (string, error) {
	if len(decl.Names) > 0 {
		return decl.Names[0].Name, nil
	}

	return "", fmt.Errorf("'Names' not found")
}

func MakeParameter(decl *ast.Field) (Parameter, error) {
	name, err := AstFieldName(decl)
	if err != nil {
		return Parameter{}, err
	}

	ptype, _ := astex.GetFieldDeclTypeName(decl.Type)

	p := Parameter{
		Name: name,
		Type: ptype,
	}
	return p, nil
}

type FuncI interface {
	TokenI

	GetFullTypeName() string
}

type Func struct {
	Token
	FType     string
	Arguments []Parameter
	Return    Type
}

func (f Func) GetFullTypeName() string {
	return f.FType
}

func CastFunc(f FuncI) (t *Func, ok bool) {
	t, ok = f.(*Func)
	return
}

func EnumFuncs(x []FuncI) iter.Seq[*Func] {
	return func(yield func(*Func) bool) {
		for _, i := range x {
			t, ok := CastFunc(i)
			if !ok {
				continue
			}

			if !yield(t) {
				return
			}
		}
	}
}

func EnumFuncsSeq(x iter.Seq[FuncI]) iter.Seq[*Func] {
	return func(yield func(*Func) bool) {
		for i := range x {
			t, ok := CastFunc(i)
			if !ok {
				continue
			}

			if !yield(t) {
				return
			}
		}
	}
}
