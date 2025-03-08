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

	"github.com/igadmg/goex/gx"
	"github.com/igadmg/goex/pprofex"
	"github.com/igadmg/gogen/core"
)

var (
	pkg_f     = flag.String("pkg", "game", "define package name used during generation")
	profile_f = flag.Bool("profile", false, "write cpu profile to `file`")
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

	//args := flag.Args()
	//if len(args) == 0 {
	//	// Default: process whole package in current directory.
	//	args = []string{"."}
	//}
	/*
		// TODO(suzmue): accept other patterns for packages (directories, list of files, import paths, etc).
		if len(args) == 1 && gog.IsDirectory(args[0]) {
			dir = args[0]
		} else {
			dir = gx.Must(os.Getwd())
		}
	*/

	var dir string = gx.Must(os.Getwd())

	for _, generator := range generators {
		if f, ok := flags[generator.Flag()]; !ok || !*f {
			continue
		}

		fmt.Printf("Runnug generator %s in %s\n", generator.Flag(), dir)
		Run(generator, dir)
	}
}

func Run(g core.Generator, dir string) {
	baseName := "0.gen_" + g.Flag() + ".go"
	outputName := filepath.Join(dir, strings.ToLower(baseName))

	if *profile_f {
		defer gx.Must(pprofex.WriteCPUProfile(outputName + ".prof"))()
	}

	g.ParsePackage([]string{dir})
	g.Inspect()
	g.Prepare()

	//func() {
	code := g.Generate(*pkg_f)

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
	//}()
}
