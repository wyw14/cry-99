package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/railvolt/internal/api"
)

func main() {
	address := flag.String("listen", env("RAILVOLT_LISTEN", "127.0.0.1:19699"), "HTTP listen address")
	dataDir := flag.String("data", env("RAILVOLT_DATA", "data"), "persistent data directory")
	webDir := flag.String("web", env("RAILVOLT_WEB", "web"), "web asset directory")
	flag.Parse()
	app, err := api.NewApplication(*dataDir, *webDir)
	if err != nil {
		log.Fatal(err)
	}
	server := api.NewServer(*address, app)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server stopped: %v", err)
			done <- syscall.SIGTERM
		}
	}()
	log.Printf("RailVolt listening on http://%s", *address)
	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
