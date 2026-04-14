package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/sergehall/lavoval/backend/api/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("bootstrap app: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Lavoval API listening on %s", application.Config.HTTPAddr)
		if err := application.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.Server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown server: %v", err)
	}

	application.Store.Close()
}
