package main

import (
	"ax-distiller/lib/ax"
	"context"
	"flag"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"

	"github.com/lmittmann/tint"
)

func main() {
	flag.Parse()
	u := flag.Arg(0)

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level: slog.LevelDebug,
		}),
	))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	navigator, err := ax.NewNavigator(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if u != "" {
		parsed, err := url.Parse(u)
		if err != nil {
			log.Fatal(err)
		}
		_, err = navigator.Navigate(parsed)
		if err != nil {
			log.Fatal("FATAL: ", err)
		}
	}

	<-ctx.Done()
}
