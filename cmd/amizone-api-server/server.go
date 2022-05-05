package main

import (
	"context"
	"github.com/ditsuke/go-amizone/server"
	"github.com/joho/godotenv"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	logger := klog.NewKlogr()

	err := godotenv.Load(".env")
	if err != nil {
		logger.Error(err, "Failed to load .env file")
	}
	address := os.Getenv("AMIZONE_API_ADDRESS")

	s := server.ApiServer{
		Config: &server.Config{
			Logger:   logger,
			BindAddr: address,
		},
		Router: http.NewServeMux(),
	}

	// Start the server on a new go-thread
	go func() {
		logger.Info("Starting server", "address", address)
		if err := s.Run(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// Main thread -- we use for interrupting OS signals
	osChannel := make(chan os.Signal, 1)
	signal.Notify(osChannel, os.Interrupt)
	signal.Notify(osChannel, os.Kill)

	// Block until a signal is received
	sig := <-osChannel

	// Log the signal
	logger.Info("os signal received", "signal", sig)

	ctx, cancelFunc := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelFunc()

	err = s.Stop(ctx)
	if err != nil {
		logger.Error(err, "failed to gracefully shut down serer", err)
	}

	logger.Info("server gracefully shut down")
}
