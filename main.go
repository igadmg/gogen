package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/igadmg/gogen/core"
	"github.com/igadmg/goex/gx"
)

var (
	ecs_f     = flag.Bool("ecs", false, "generate ecs code")
	gog_f     = flag.Bool("gog", false, "generate gog code")
	profile_f = flag.String("profile", "", "write cpu profile to `file`")
	generator core.Generator
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of gog:\n")
	fmt.Fprintf(os.Stderr, "\tgog [flags]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("gog: ")
	flag.Usage = Usage
	flag.Parse()

	if len(*profile_f) > 0 {
		f, err := os.Create(*profile_f)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	var dir string

	// TODO(suzmue): accept other patterns for packages (directories, list of files, import paths, etc).
	if len(args) == 1 && gog.IsDirectory(args[0]) {
		dir = args[0]
	} else {
		dir = gx.Must(os.Getwd())
	}

	if *ecs_f {
		generator = ecs.NewGeneratorEcs(gx.Must(os.Getwd()))
	} else {
		// Parse the package once.
		generator = gog.NewGeneratorGog()
	}

	Run(generator, dir)
}

func Run(g core.Generator, dir string) {
	g.ParsePackage([]string{dir})
	g.Inspect()
	g.Prepare()
	g.Generate(dir)
}
