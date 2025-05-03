package net

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/igadmg/goecs/ecs/net"
)

var (
	profile_f      = flag.Bool("profile", false, "write cpu profile to `file`")
	no_store_dot_f = flag.Bool("no_store_dot", false, "don't store dot file with class diagram")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of gog:\n")
	fmt.Fprintf(os.Stderr, "\tgog [flags]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

type Server int

func Execute() {
	ctx, cancel := context.WithCancel(context.Background())

	s := net.Server{}
	server := 0
	s.Listen(server, ctx)

	/*
		for {
			select {
			case <-ctx.Done():

			}
		}
	*/
	cancel()
}
