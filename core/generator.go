package core

import (
	"bytes"
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"deedles.dev/xiter"
	"github.com/igadmg/goex/astex"
	"github.com/igadmg/goex/gx"
	"golang.org/x/tools/go/packages"
	"gonum.org/v1/gonum/graph"
)

type TypeFactory interface {
	GetPackage() *Package
	NewType(pkg *Package, t TypeI, spec *ast.TypeSpec) (TypeI, error)
	NewField(f FieldI, spec *ast.Field) (FieldI, error)
	NewFunc(f FuncI, spec *ast.FuncDecl) (FuncI, error)

	GetType(name string) (t TypeI, ok bool)
	GetFuncs(t TypeI) []FuncI
}

type Generator interface {
	TypeFactory

	Flag() string
	Tags() []string

	Prepare()
	Generate(pkg *Package) bytes.Buffer

	Yaml(fileName string)
	Graph() graph.Graph
}

type GeneratorBase struct {
	Buf  bytes.Buffer // Accumulated output.
	cfg  *packages.Config
	Pkg  *Package
	Pkgs []*Package // Package we are scanning.

	flag string   // cmd line flags
	tags []string // tag names

	//logf func(format string, args ...any) // test logging hook; nil when not testing
}

func (g *GeneratorBase) Flag() string {
	return g.flag
}

func (g *GeneratorBase) Tags() []string {
	return g.tags
}

func (g *GeneratorBase) Print(str string) {
	fmt.Fprint(&g.Buf, str)
}

func (g *GeneratorBase) Printf(format string, args ...any) {
	fmt.Fprintf(&g.Buf, format, args...)
}

func (g *GeneratorBase) Printfln(format string, args ...any) {
	fmt.Fprintf(&g.Buf, format+"\n", args...)
}

func (g *GeneratorBase) GetPackage() *Package {
	return g.Pkg
}

func (g *GeneratorBase) TypeImportName(t TypeI) string {
	if t == nil {
		return "<unknown type>"
	}

	if t.GetPackage() == g.Pkg {
		return ""
	}

	return t.GetPackage().Name + "."
}

func (g *GeneratorBase) LocalTypeName(t TypeI) string {
	if t == nil {
		return "<unknown type>"
	}

	if t.GetPackage() == g.Pkg {
		return t.GetName()
	}

	return t.GetFullName()
}

type GeneratorBaseT struct {
	GeneratorBase

	G Generator

	Types  map[string]TypeI
	Fields []FieldI
	Funcs  map[string][]FuncI
}

func MakeGeneratorB(flag string, tags ...string) GeneratorBaseT {
	return GeneratorBaseT{
		GeneratorBase: GeneratorBase{
			flag: flag,
			tags: tags,
		},
		Types:  map[string]TypeI{},
		Fields: []FieldI{},
		Funcs:  map[string][]FuncI{},
	}
}

func (g *GeneratorBaseT) NewType(pkg *Package, t TypeI, spec *ast.TypeSpec) (TypeI, error) {
	if t == nil {
		t = NewType(pkg)
		defer func() {
			g.Types[t.GetFullName()] = t
		}()
	}

	switch et := t.(type) {
	case *Type:
		ttype, ok := spec.Type.(*ast.StructType)
		if !ok {
			return nil, fmt.Errorf("can parse only *ast.StructType")
		}

		et.Name = spec.Name.Name

		fieldCount := 0
		for _, field := range ttype.Fields.List {
			f, err := g.G.NewField(nil, field)
			if err != nil {
				fieldCount++
				continue
			}

			if fb, ok := f.(FieldBuilder); ok {
				fb.SetOwnerType(t)
				tp := strings.Split(f.GetTypeName(), ".")
				if len(tp) == 1 {
					fb.SetPackagedTypeName(et.Package.Name + "." + f.GetTypeName())
				} else {
					fb.SetPackagedTypeName(f.GetTypeName())
				}
			}

			if len(f.GetName()) == 0 {
				et.BaseFields = append(et.BaseFields, f)

				continue
			}

			if f.IsMeta() {
				et.Tag = f.GetTag()
				continue
			}

			fieldCount++
			et.Fields = append(et.Fields, f)
		}
		et.isZero = fieldCount == 0
	}

	return t, nil
}

func (g *GeneratorBaseT) NewField(f FieldI, spec *ast.Field) (FieldI, error) {
	if f == nil {
		f = &Field{}
		defer func() {
			g.Fields = append(g.Fields, f)
		}()
	}

	switch ef := f.(type) {
	case *Field:
		var ok bool

		if len(spec.Names) > 0 {
			ef.Name = spec.Names[0].Name
		}

		ef.decltype, ok = astex.GetFieldDeclTypeName(spec.Type)
		if !ok {
			return nil, fmt.Errorf("failed to get field decl type name")
		}

		ef.TypeName, ok = astex.ExprGetFullTypeName(spec.Type)
		if !ok {
			return nil, fmt.Errorf("failed to get full type name")
		}

		ef.CallTypeName, ok = astex.ExprGetCallTypeName(spec.Type)
		if !ok {
			return nil, fmt.Errorf("failed to get call type name")
		}

		_, ef.IsArray_ = spec.Type.(*ast.ArrayType)

		ef.Tag, _ = ParseTag(spec.Tag)
		//ef.isComponent = err == nil
	}

	return f, nil
}

func (g *GeneratorBaseT) NewFunc(f FuncI, decl *ast.FuncDecl) (FuncI, error) {
	if f == nil {
		f = NewFunc(g.Pkg)
		defer func() {
			if id := f.GetFullTypeName(); id != "" {
				g.Funcs[id] = append(g.Funcs[id], f)
			}
		}()
	}

	switch ef := f.(type) {
	case *Func:
		ef.Name = decl.Name.Name

		if decl.Doc != nil {
			doc := decl.Doc.Text()
			for _, tag := range g.Tags() {
				if strings.HasPrefix(doc, tag+":") {
					doc = strings.Trim(doc[len(tag)+1:], "\"\n \t")
					if ef.Tag.Data == nil {
						ef.Tag.Data = TagData{}
					}
					ef.Tag.Data[tag] = Tag{gx.Should(UnmarshalTag(doc))}
				}
			}
		}

		rtype, ok := astex.FuncDeclRecvType(decl)
		if ok {
			ef.FType, ok = astex.ExprGetFullTypeName(rtype)
			if !ok {
				return nil, fmt.Errorf("failed to get full type name")
			}

			ef.DeclType, ok = astex.GetFieldDeclTypeName(rtype)
			if !ok {
				return nil, fmt.Errorf("failed to get field type name")
			}
		}

		params, ok := astex.FuncDeclParams(decl)
		if ok {
			ef.Arguments = slices.Collect(
				xiter.Map(slices.Values(params), func(f *ast.Field) Parameter {
					p, _ := MakeParameter(f)
					return p
				}))
		}
	}

	return f, nil
}

func (g *GeneratorBaseT) GetType(name string) (t TypeI, ok bool) {
	name = strings.TrimLeft(name, " *")
	names := strings.Split(name, ".")
	if len(names) > 1 {
		t, ok = g.Types[name]
	} else {
		t, ok = g.Types[g.Pkg.Name+"."+name]
	}
	return
}

func (g *GeneratorBaseT) GetFuncs(t TypeI) []FuncI {
	return g.Funcs[t.GetName()]
}

var reportedTypes map[string]struct{} = map[string]struct{}{}

func (g *GeneratorBaseT) Prepare() {
	for _, f := range g.Fields {
		fb, ok := f.(FieldBuilder)
		if !ok {
			continue
		}

		err := fb.Prepare(g.G)
		if err != nil {
			if _, ok := reportedTypes[f.GetTypeName()]; !ok {
				reportedTypes[f.GetTypeName()] = struct{}{}
				fmt.Printf("Error: %v\n", err)
			}
		}
	}

	for _, t := range g.Types {
		tb, ok := t.(TypeBuilder)
		if !ok {
			continue
		}

		err := tb.Prepare(g.G)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		for base := range t.BasesSeq() {
			tb, ok := base.GetType().(TypeBuilder)
			if !ok {
				continue
			}

			tb.AddSubclass(t)
		}
	}
}
