package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	loader := NewLoader("data/treestore")
	man, err := NewManager(loader)
	if err != nil {
		slog.Error("[main] create manager", "err", err)
		os.Exit(1)
	}

	err = man.Run(ctx)
	if err != nil {
		slog.Error("[main] start manager", "err", err)
		os.Exit(1)
	}

	slog.Info("[main] success")
}
