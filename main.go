// Command appwrite-go-kit-template is a runnable Gin service showing appwrite-base
// end to end: env config, capability wiring, guarded routes, database CRUD and
// storage uploads.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/E3TU/appwrite-go-kit-template/internal/config"
	"github.com/E3TU/appwrite-go-kit-template/internal/router"
	"github.com/e3tu/appwrite-base"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "appwrite-go-kit-template:", err)
		os.Exit(1)
	}
}

func run() error {
	env, err := config.Load()
	if err != nil {
		return err
	}

	b, err := base.New(env.Base)
	if err != nil {
		return err
	}

	sessionHeader := ""
	if env.Base.Auth != nil {
		sessionHeader = env.Base.Auth.SessionHeader
	}

	srv := &http.Server{
		Addr:              env.Listen,
		Handler:           router.New(b, env.TodosTable, sessionHeader),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errc := make(chan error, 1)
	go func() {
		fmt.Printf("listening on %s\n", env.Listen)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errc <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errc:
		return err
	case sig := <-stop:
		fmt.Fprintf(os.Stderr, "received %s, shutting down\n", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}
