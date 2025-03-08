package core

import "fmt"

type FieldI interface {
	TokenI

	IsMeta() bool
	GetType() TypeI
	GetTypeName() string
	DeclType() string
}

type FieldBuilder interface {
	Prepare(tf TypeFactory) error
}

type Field struct {
	Token
	Type         TypeI
	TypeName     string
	CallTypeName string
	decltype     string
	IsArray      bool
}

func (f *Field) Clone() any {
	c := *f
	return &c
}

func (f Field) IsMeta() bool {
	return f.TypeName == "ecs.MetaTag"
}

func (f Field) GetType() TypeI {
	return f.Type
}

func (f Field) GetTypeName() string {
	return f.TypeName
}

func (f Field) DeclType() string {
	return f.decltype
}

func (f *Field) Prepare(tf TypeFactory) error {
	var ok bool
	f.Type, ok = tf.GetType(f.TypeName)
	if !ok {
		return fmt.Errorf("type %s not found", f.TypeName)
	}

	return nil
}
