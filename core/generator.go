package core

import (
	"bytes"
	"fmt"
	"go/ast"
	"log"
	"path/filepath"
	"slices"
	"strings"

	"deedles.dev/xiter"
	"github.com/igadmg/goex/astex"
	"github.com/igadmg/goex/gx"
	"golang.org/x/tools/go/packages"
)

type TypeFactory interface {
	NewType(t TypeI, spec *ast.TypeSpec) (TypeI, error)
	NewField(f FieldI, spec *ast.Field) (FieldI, error)
	NewFunc(f FuncI, spec *ast.FuncDecl) (FuncI, error)

	GetType(name string) (t TypeI, ok bool)
	GetFuncs(t TypeI) []FuncI
}

type Generator interface {
	Flag() string
	Tags() []string

	ParsePackage(patterns []string /*, tags []string*/)
	Inspect()
	Prepare()
	Generate(pkg string) bytes.Buffer
}

type GeneratorBase struct {
	Buf bytes.Buffer // Accumulated output.
	cfg *packages.Config
	pkg *Package // Package we are scanning.

	flag string   // cmd line flags
	tags []string // tag names

	logf func(format string, args ...any) // test logging hook; nil when not testing
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

func (g *GeneratorBase) ParsePackage(patterns []string /*, tags []string*/) {
	g.cfg = &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		// TODO: Need to think about constants in test files. Maybe write type_string_test.go
		// in a separate pass? For later.
		Tests: false,
		//BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
		Logf: g.logf,
	}
	pkgs, err := packages.Load(g.cfg, patterns...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages matching %v", len(pkgs), strings.Join(patterns, " "))
	}
	g.addPackage(pkgs[0])
}

// addPackage adds a type checked Package and its syntax files to the generator.
func (g *GeneratorBase) addPackage(pkg *packages.Package) {
	g.pkg = &Package{
		pkg:   pkg,
		name:  pkg.Name,
		defs:  pkg.TypesInfo.Defs,
		files: make([]*File, len(pkg.Syntax)),
	}

	for i, file := range pkg.Syntax {
		f := &File{
			file: file,
			pkg:  g.pkg,
		}
		g.pkg.files[i] = f
	}
}

type GeneratorBaseT struct {
	GeneratorBase

	G interface {
		Generator
		TypeFactory
	}

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

func (g *GeneratorBaseT) NewType(t TypeI, spec *ast.TypeSpec) (TypeI, error) {
	if t == nil {
		t = MakeType().New()
		defer func() {
			g.Types[t.GetName()] = t
		}()
	}

	switch et := t.(type) {
	case *Type:
		ttype, ok := spec.Type.(*ast.StructType)
		if !ok {
			return nil, fmt.Errorf("can parse only *ast.StructType")
		}

		et.Name = spec.Name.Name

		for _, field := range ttype.Fields.List {
			f, err := g.G.NewField(nil, field)
			if err != nil {
				continue
			}

			if len(f.GetName()) == 0 {
				et.BaseFields = append(et.BaseFields, f)

				continue
			}

			if f.IsMeta() {
				et.Tag = f.GetTag()
				continue
			}

			et.Fields = append(et.Fields, f)
		}
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

		_, ef.IsArray = spec.Type.(*ast.ArrayType)

		ef.Tag, _ = ParseTag(spec.Tag)
		//ef.isComponent = err == nil
	}

	return f, nil
}

func (g *GeneratorBaseT) NewFunc(f FuncI, decl *ast.FuncDecl) (FuncI, error) {
	if f == nil {
		f = &Func{}
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
	t, ok = g.Types[name]
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

func (g *GeneratorBaseT) Inspect() {
	for _, file := range g.pkg.files {
		// Set the state for this run of the walker.
		//file.values = nil
		if file.file != nil {
			ast.Inspect(file.file, func(n ast.Node) bool {
				switch decl := n.(type) {
				case *ast.File:
					fn := filepath.Base(file.pkg.pkg.Fset.Position(decl.Package).Filename)
					return !strings.HasPrefix(fn, "0.gen")
				default:
					return g.InspectCode(decl)
				}
			})
		}
	}
}

func (g *GeneratorBaseT) InspectCode(node ast.Node) (follow bool) {
	switch decl := node.(type) {
	case *ast.GenDecl:
		for _, spec := range decl.Specs {
			switch tspec := spec.(type) {
			case *ast.TypeSpec:
				_, err := g.G.NewType(nil, tspec)
				if err != nil {
					continue
				}
			}
		}
		return false
	case *ast.FuncDecl:
		_, err := g.G.NewFunc(nil, decl)
		if err != nil {
			return false
		}

		return false
	}

	return true
}
