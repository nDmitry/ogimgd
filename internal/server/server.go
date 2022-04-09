package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Run starts the HTTP server
func Run(port int, d drawer) {
	ctx, cancel := context.WithCancel(context.Background())
	startedAt := time.Now().UTC()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/preview", getPreview(d))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(
			w, "Server is on since %s. Online: %s \n",
			startedAt.Format(time.RFC3339), time.Now().UTC().Sub(startedAt),
		)
	})

	server := &http.Server{
		Addr:        fmt.Sprintf(":%d", port),
		Handler:     r,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	server.RegisterOnShutdown(cancel)

	go func() {
		log.Printf("HTTP server started on port %d\n", port)

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	<-signalChan
	log.Println("os.Interrupt - shutting down...")

	go func() {
		<-signalChan
		log.Fatalln("os.Kill - terminating...")
	}()

	gracefullCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(gracefullCtx); err != nil {
		log.Printf("shutdown error: %v\n", err)
		defer os.Exit(1)
		return
	}

	log.Println("gracefully stopped")
}
