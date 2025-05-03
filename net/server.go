package net

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	sigChan := make(chan os.Signal, 1)

	// Регистрируем сигналы для Windows
	signal.Notify(sigChan,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // Завершение процесса
		// syscall.SIGBREAK, // Ctrl+Break (раскомментировать если нужно)
	)

	ctx, cancel := context.WithCancel(context.Background())

	s := net.Server{}
	server := 0
	s.Listen(server, ctx)

	sig := <-sigChan
	log.Printf("Получен сигнал: %v. Завершение...", sig)
	cancel()
}
