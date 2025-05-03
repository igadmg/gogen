package gogen

import (
	"flag"
	"fmt"
	"go/format"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"deedles.dev/xiter"
	"github.com/igadmg/goex/gx"
	"github.com/igadmg/goex/pprofex"
	"github.com/igadmg/gogen/core"
	"golang.org/x/tools/go/packages"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gopkg.in/yaml.v3"
)

var (
	profile_f       = flag.Bool("profile", false, "write cpu profile to `file`")
	no_store_dot_f  = flag.Bool("no_store_dot", false, "don't store dot file with class diagram")
	no_store_yaml_f = flag.Bool("no_store_yaml", false, "don't store yaml file with metadata")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of gog:\n")
	fmt.Fprintf(os.Stderr, "\tgog [flags]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func Execute(generators ...core.Generator) {
	flags := map[string]*bool{}
	tags := map[string]struct{}{}
	for _, generator := range generators {
		for _, tag := range generator.Tags() {
			tags[tag] = struct{}{}
		}
		flags[generator.Flag()] = flag.Bool(generator.Flag(), false, "generate "+generator.Flag()+" code")
	}

	core.Tags = slices.Collect(maps.Keys(tags))

	log.SetFlags(0)
	log.SetPrefix("gogen: ")
	flag.Usage = Usage
	flag.Parse()

	var dir []string
	args := flag.Args()
	if len(args) > 0 {
		dir = args
	} else {
		dir = []string{gx.Must(os.Getwd())}
	}
	/*
		// TODO(suzmue): accept other patterns for packages (directories, list of files, import paths, etc).
		if len(args) == 1 && gog.IsDirectory(args[0]) {
			dir = args[0]
		} else {
			dir = gx.Must(os.Getwd())
		}
	*/

	for _, generator := range generators {
		if f, ok := flags[generator.Flag()]; !ok || !*f {
			continue
		}

		fmt.Printf("Runnug generator %s in %s\n", generator.Flag(), dir)
		Run(generator, dir)
	}
}

func Run(g core.Generator, pkgNames []string) {
	baseName := "0.gen_" + g.Flag() + ".go"

	if *profile_f {
		defer gx.Must(pprofex.WriteCPUProfile(baseName + ".prof"))()
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

	ppkg := slices.Collect(
		xiter.Map(slices.Values(pkgs),
			func(pkg *packages.Package) *core.Package {
				lpkg := &core.Package{
					Pkg:   pkg,
					Name:  pkg.Name,
					Defs:  pkg.TypesInfo.Defs,
					Files: make([]*core.File, len(pkg.Syntax)),
					Types: map[string]core.TypeI{},
					Funcs: map[string][]core.FuncI{},
				}

				for i, file := range pkg.Syntax {
					f := &core.File{
						File: file,
						Pkg:  lpkg,
					}
					lpkg.Files[i] = f
				}

				return lpkg
			}))

	g.Inspect(ppkg)

	var wg sync.WaitGroup

	for _, pkg := range ppkg {
		func(pkg *core.Package) {
			outputName := filepath.Join(pkg.Pkg.Dir, strings.ToLower(baseName))
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
					log.Printf("Writing file %s", dotName)
					file, err := os.Create(dotName)
					if err != nil {
						log.Fatalf("writing output: %s", err)
					}
					defer file.Close()

					yaml.NewEncoder(file).Encode(g)

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
		}(pkg)
	}

	wg.Wait()
}
