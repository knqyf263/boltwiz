package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4/middleware"

	"github.com/knqyf263/boltwiz/modules/database/repository"
	"github.com/knqyf263/boltwiz/server/handlers"
	"github.com/knqyf263/boltwiz/server/routes"

	"github.com/labstack/echo/v4"
)

type Options struct {
	DBPath     string
	Port       int
	ProtoFiles []string
	ProtoType  string
}

func StartServer(opts Options) error {
	repo, err := repository.NewRepository(opts.DBPath, opts.ProtoType, opts.ProtoFiles)
	if err != nil {
		return err
	}
	defer repo.Close()
	h := handlers.NewHandlers(repo)

	// Echo instance
	e := echo.New()

	e.Use(middleware.CORS())
	// Routes
	routes.RegisterStaticRoutes(e)
	routes.RegisterV1Routes(e, h)

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", opts.Port),
	}

	// Start server
	go func() {
		if err := e.StartServer(server); !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatalf("Failed to start server, %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	slog.Info("Gracefully shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = e.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}
