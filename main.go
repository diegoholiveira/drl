package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/diegoholiveira/drl/limiter"
	"github.com/diegoholiveira/drl/server"
)

const (
	// HTTP server defaults
	DefaultHTTPAddr        = ":8080"
	DefaultIdleConnTimeout = 30 * time.Second
	DefaultReadTimeout     = 10 * time.Second
	DefaultWriteTimeout    = 10 * time.Second

	// Limiter defaults
	DefaultRateLimit   = 3
	DefaultTimeWindow  = 1 * time.Minute
	DefaultGranularity = 10 * time.Second
)

func main() {
	shutdownSignal := make(chan os.Signal, 1)

	signal.Notify(
		shutdownSignal,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: ":6379",
	})
	_ = rdb.FlushDB(ctx).Err()

	proxy := server.NewReverseProxy(
		// limiter.NewSimpleCounterLimit(DefaultRateLimit, DefaultTimeWindow, DefaultGranularity),
		limiter.NewRedisLimiter(rdb, DefaultRateLimit, DefaultTimeWindow, DefaultGranularity),
	)

	srv := &http.Server{
		Addr:         DefaultHTTPAddr,
		Handler:      proxy,
		IdleTimeout:  DefaultIdleConnTimeout,
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
	}

	go func() {
		log.Printf("http server at %s\n", DefaultHTTPAddr)

		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server error: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for sig := range shutdownSignal {
		log.Printf("received signal: %s\n", sig)

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("http server shutdown error: %v", err)
		} else {
			log.Printf("http server shutdown completed\n")
			break
		}
	}
}
