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
	"github.com/igadmg/gogen/core"
)

var ()

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of gog:\n")
	fmt.Fprintf(os.Stderr, "\tgog [flags]\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func Execute(fg *flag.FlagSet, generators ...core.Generator) {
	sigChan := make(chan os.Signal, 1)

	// Регистрируем сигналы для Windows
	signal.Notify(sigChan,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // Завершение процесса
		// syscall.SIGBREAK, // Ctrl+Break (раскомментировать если нужно)
	)

	ctx, cancel := context.WithCancel(context.Background())

	s := net.Server{}
	for _, g := range generators {
		s.Register(g)
	}
	go s.Listen(ctx)

	sig := <-sigChan
	log.Printf("Получен сигнал: %v. Завершение...", sig)
	cancel()
}
