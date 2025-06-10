package gogen

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"deedles.dev/xiter"
	"github.com/igadmg/goex/gx"
	"github.com/igadmg/goex/pprofex"
	"github.com/igadmg/goex/timeex"
	"github.com/igadmg/gogen/core"
	"golang.org/x/tools/go/packages"
	"gonum.org/v1/gonum/graph/encoding/dot"
)

var (
	profile_f       *bool
	no_store_dot_f  *bool
	no_store_yaml_f *bool
	appModTime      time.Time
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of gog:\n")
	fmt.Fprintf(os.Stderr, "\tgog [flags]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func Execute(fg *flag.FlagSet, generators ...core.Generator) {
	profile_f = fg.Bool("profile", false, "write cpu profile to `file`")
	no_store_dot_f = fg.Bool("no_store_dot", true, "don't store dot file with class diagram")
	no_store_yaml_f = fg.Bool("no_store_yaml", true, "don't store yaml file with metadata")

	flags := map[string]*bool{}
	tags := map[string]struct{}{}
	for _, generator := range generators {
		for _, tag := range generator.Tags() {
			tags[tag] = struct{}{}
		}
		flags[generator.Flag()] = fg.Bool(generator.Flag(), false, "generate "+generator.Flag()+" code")
	}

	core.Tags = slices.Collect(maps.Keys(tags))

	log.SetFlags(0)
	log.SetPrefix("gogen: ")
	fg.Usage = Usage
	fg.Parse(os.Args[1:])

	var dir []string
	args := fg.Args()
	if len(args) > 0 {
		dir = args
	} else {
		dir = []string{gx.Must(os.Getwd())}
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	pfs, err := os.Stat(ex)
	if err != nil {
		panic(err)
	}
	appModTime = pfs.ModTime()

	/*
		// TODO(suzmue): accept other patterns for packages (directories, list of files, import paths, etc).
		if len(args) == 1 && gog.IsDirectory(args[0]) {
			dir = args[0]
		} else {
			dir = gx.Must(os.Getwd())
		}
	*/

	generators = slices.Collect(xiter.Filter(slices.Values(generators), func(g core.Generator) bool {
		if f, ok := flags[g.Flag()]; !ok || !*f {
			return false
		}

		return true
	}))

	Run(dir, generators...)
}

func Run(pkgNames []string, generators ...core.Generator) {
	if *profile_f {
		defer gx.Must(pprofex.WriteCPUProfile("gogen"))()
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		// TODO: Need to think about constants in test files. Maybe write type_string_test.go
		// in a separate pass? For later.
		Tests: false,
		//BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
		//Logf: g.logf,
	}
	pkgs, err := packages.Load(cfg, pkgNames...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) == 0 {
		log.Fatalf("error: %d packages matching %v", len(pkgs), strings.Join(pkgNames, " "))
	}

	ppkgs := map[string]*core.Package{}
	for _, pkg := range pkgs {
		ppkgs[pkg.PkgPath] = func() *core.Package {
			lpkg := core.NewPackage(pkg)

			lpkg.ModTime = time.Time{}
			for _, file := range pkg.Syntax {
				fileName := pkg.Fset.Position(file.Package).Filename
				if strings.HasPrefix(filepath.Base(fileName), "0.gen_") {
					continue
				}

				f := &core.File{
					File:    file,
					Pkg:     lpkg,
					ModTime: timeex.EndOfTime,
				}

				fileInfo, err := os.Stat(fileName)
				if err == nil {
					f.ModTime = fileInfo.ModTime()
				}

				if f.ModTime.Compare(lpkg.ModTime) > 0 {
					lpkg.ModTime = f.ModTime
				}
				lpkg.Files = append(lpkg.Files, f)
			}

			return lpkg
		}()
	}

	Inspect(ppkgs, generators...)

	for _, pkg := range ppkgs {
		pkg.ImportedPkgs = map[string]*core.Package{}
		for _, imp := range pkg.Imports {
			if ipkg, ok := ppkgs[imp.Path]; ok {
				pkg.ImportedPkgs[imp.Path] = ipkg

				if ipkg.ModTime.Compare(pkg.ModTime) > 0 {
					pkg.ModTime = ipkg.ModTime
				}
			}
		}

		if appModTime.Compare(pkg.ModTime) > 0 {
			pkg.ModTime = appModTime
		}
	}

	var wg sync.WaitGroup

	for _, pkg := range ppkgs {
		for _, g := range generators {
			func(g core.Generator, pkg *core.Package) {
				baseName := "0.gen_" + g.Flag() + ".go"
				outputName := filepath.Join(pkg.Pkg.Dir, strings.ToLower(baseName))
				if fs, err := os.Stat(outputName); !errors.Is(err, os.ErrNotExist) {
					if fs.ModTime().Compare(pkg.ModTime) > 0 {
						return
					}
				}

				code := g.Generate(pkg)

				if !*no_store_dot_f {
					wg.Add(1)
					go func() {
						defer wg.Done()
						dg := g.Graph()

						// Write the graph to DOT format
						data, err := dot.Marshal(dg, "", "", "  ")
						if err != nil {
							log.Fatal(err)
						}

						baseName := "0.gen_" + g.Flag() + ".dot"
						dotName := filepath.Join(pkg.Pkg.Dir, strings.ToLower(baseName))
						log.Printf("Writing file %s", dotName)
						err = os.WriteFile(dotName, data, 0644)
						if err != nil {
							log.Fatalf("writing output: %s", err)
						}
						log.Printf("Done file %s", dotName)
					}()
				}

				if !*no_store_yaml_f {
					wg.Add(1)
					go func() {
						defer wg.Done()
						baseName := "0.gen_" + g.Flag() + ".yaml"
						dotName := filepath.Join(pkg.Pkg.Dir, strings.ToLower(baseName))

						// TODO: network comm here
						//g.Yaml(dotName)

						log.Printf("Done file %s", dotName)
					}()
				}

				log.Printf("Formatting file %s", outputName)
				src, err := format.Source(code.Bytes())
				if err != nil {
					// Should never happen, but can arise when developing this code.
					// The user can compile the output to see the error.
					log.Printf("warning: internal error: invalid Go generated: %s", err)
					log.Printf("warning: compile the package to analyze the error")

					os.WriteFile(outputName, code.Bytes(), 0644)
					return
				}

				// Write to file.
				log.Printf("Writing file %s", outputName)
				err = os.WriteFile(outputName, src, 0644)
				if err != nil {
					log.Fatalf("writing output: %s", err)
				}
				log.Printf("Formatting file %s", outputName)

				lint := exec.Command("go", "run", "golang.org/x/tools/cmd/goimports", "-w", outputName)
				if err := lint.Run(); err != nil {
					log.Fatal(err)
				}

				log.Printf("Done file %s", outputName)
			}(g, pkg)
		}
	}

	wg.Wait()
}

func Inspect(pkgs map[string]*core.Package, generators ...core.Generator) {
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			if file.File != nil {
				ast.Inspect(file.File, func(n ast.Node) bool {
					switch decl := n.(type) {
					case *ast.File:
						fn := filepath.Base(file.Pkg.Pkg.Fset.Position(decl.Package).Filename)
						return !strings.HasPrefix(fn, "0.gen")
					default:
						return InspectCode(pkg, decl, generators...)
					}
				})
			}
		}
	}
}

func InspectCode(pkg *core.Package, node ast.Node, generators ...core.Generator) (follow bool) {
	switch decl := node.(type) {
	case *ast.GenDecl:
		for _, spec := range decl.Specs {
			switch tspec := spec.(type) {
			case *ast.TypeSpec:
				for _, g := range generators {
					g.NewType(pkg, nil, tspec)
				}
			case *ast.ImportSpec:
				pkg.AddImport(tspec)
			}
		}
		return false
	case *ast.FuncDecl:
		for _, g := range generators {
			g.NewFunc(nil, decl)
		}
		return false
	}

	return true
}
