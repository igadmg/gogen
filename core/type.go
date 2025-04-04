package core

import (
	"iter"
	"maps"
	"slices"
)

type TypeI interface {
	TokenI

	BasesSeq() iter.Seq[FieldI]
	FieldsSeq() iter.Seq[FieldI]
	FuncsSeq() iter.Seq[FuncI]

	CanCall(name string) bool
	HasFunction(name string) bool
}

type TypeBuilder interface {
	Prepare(tf TypeFactory) error

	AddSubclass(s TypeI)
}

type Type struct {
	Token
	Subclasses []TypeI
	BaseFields []FieldI // base types go lang way (deprecated for archetype)
	Extends    []TypeI  // extends for archetypes
	Fields     []FieldI
	Funcs      map[string]FuncI
}

var _ TypeI = (*Type)(nil)
var _ TypeBuilder = (*Type)(nil)

func MakeType(pkg *Package) Type {
	return Type{
		Token: Token{
			Package: pkg,
		},
		Funcs: map[string]FuncI{},
	}
}

func NewType(pkg *Package) *Type {
	t := MakeType(pkg)
	return t.New()
}

func (t *Type) New() *Type {
	return t
}

func (t Type) GetName() string {
	return t.Name
}

func (t Type) BasesSeq() iter.Seq[FieldI] {
	return slices.Values(t.BaseFields)
}

func (t Type) FieldsSeq() iter.Seq[FieldI] {
	return slices.Values(t.Fields)
}

func (t Type) FuncsSeq() iter.Seq[FuncI] {
	return maps.Values(t.Funcs)
}

func (t Type) CanCall(name string) bool {
	return t.HasFunction(name)
}

func (t Type) HasFunction(name string) bool {
	_, ok := t.Funcs[name]
	if ok {
		return true
	}

	for _, base := range t.BaseFields {
		if bt := base.GetType(); bt != nil && bt.HasFunction(name) {
			return true
		}
	}

	return false
}

func (t *Type) Prepare(tf TypeFactory) error {
	t.Funcs = map[string]FuncI{}
	for _, f := range tf.GetFuncs(t) {
		t.Funcs[f.GetName()] = f
	}
	return nil
}

func (t *Type) AddSubclass(s TypeI) {
	t.Subclasses = append(t.Subclasses, s)
}
