package main

import (
	"embed"
	_ "embed"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
)

//go:embed public
var public embed.FS

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		panic(err)
	}
	slog.Info("serving", "addr", "127.0.0.1:8000")

	sub, err := fs.Sub(public, "public")
	if err != nil {
		panic(err)
	}
	err = http.Serve(listener, http.FileServerFS(sub))
	if err != nil {
		panic(err)
	}
}
