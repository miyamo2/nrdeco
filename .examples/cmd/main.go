package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/miyamo2/nrdeco/examples/di"
)

func main() {
	mux := di.ServeMux()
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return
	}
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	errChan := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			errChan <- err
			return
		}
	}()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	select {
	case err := <-errChan:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error(err.Error())
		}
	case <-ctx.Done():
		slog.Info("shutdown server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Shutdown(ctx)
		switch {
		case err == nil:
			return
		case !errors.Is(err, context.Canceled), !errors.Is(err, http.ErrServerClosed):
			slog.Error("server shutdown failed", "error", err)
		}
		return
	}
}
