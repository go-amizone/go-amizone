package main

import (
	"amizone/amizone_api/handlers"
	"context"
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
	address := os.Getenv("AMIZONE_API_PORT")

	a := handlers.NewHandlerCfg(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/attendance", a.AttendanceHandler)
	mux.HandleFunc("/schedule", a.ClassScheduleHandler)

	server := http.Server{
		Addr:    address,
		Handler: mux,
	}

	// Start the server on a new go-thread
	go func() {
		logger.Info("Starting server", "address", address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

	err = server.Shutdown(ctx)
	if err != nil {
		logger.Error(err, "failed to gracefully shut down serer", err)
	}

	logger.Info("server gracefully shut down")
}
