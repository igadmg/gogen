package core

import (
	"iter"
	"slices"
)

type TypeI interface {
	TokenI

	BasesSeq() iter.Seq[FieldI]
	FieldsSeq() iter.Seq[FieldI]

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
	Bases      []FieldI // base types
	Fields     []FieldI
	Funcs      map[string]FuncI
}

var _ TypeI = (*Type)(nil)
var _ TypeBuilder = (*Type)(nil)

func MakeType() Type {
	return Type{
		Funcs: map[string]FuncI{},
	}
}

func NewType() *Type {
	t := MakeType()
	return t.New()
}

func (t Type) New() *Type {
	return &t
}

func (t Type) GetName() string {
	return t.Name
}

func (t Type) BasesSeq() iter.Seq[FieldI] {
	return slices.Values(t.Bases)
}

func (t Type) FieldsSeq() iter.Seq[FieldI] {
	return slices.Values(t.Fields)
}

func (t Type) CanCall(name string) bool {
	return t.HasFunction(name)
}

func (t Type) HasFunction(name string) bool {
	_, ok := t.Funcs[name]
	if ok {
		return true
	}

	for _, base := range t.Bases {
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
