package main

import (
	"context"
	"flag"
	"github.com/ditsuke/go-amizone/server"
	v1 "github.com/ditsuke/go-amizone/server/gen/go/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultAddress = "0.0.0.0:8081"
	AddressEnvVar  = "AMIZONE_API_ADDRESS"
)

func main() {
	logger := klog.NewKlogr()
	ctxBg := context.Background()
	_ = godotenv.Load(".env")

	config := struct{ address string }{}

	flagSet := flag.NewFlagSet("server config", flag.ExitOnError)
	flagSet.StringVar(&config.address, "address", EnvOrDefault(AddressEnvVar, DefaultAddress), "Address to listen on")
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		logger.Error(err, "failed to parse flags")
		os.Exit(1)
	}

	grpcMux := grpc.NewServer()
	v1.RegisterAmizoneServiceServer(grpcMux, server.NewAmizoneServiceServer([]byte("nnnnnnnnnnnnnnnnnnnnnnnnnnnnnnnn")))
	reflection.Register(grpcMux)

	gwMux := runtime.NewServeMux()
	err := v1.RegisterAmizoneServiceHandlerFromEndpoint(ctxBg, gwMux, "localhost:8081", []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		logger.Error(err, "failed to register grpc gateway")
		os.Exit(1)
	}

	// Get a tcp network listener
	conn, err := net.Listen("tcp", config.address)
	if err != nil {
		panic(err)
	}

	s := &http.Server{
		Addr: config.address,
		Handler: h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isGrpc(r) { // Forward to the gRPC server
				grpcMux.ServeHTTP(w, r)
				return
			}
			gwMux.ServeHTTP(w, r)
		}), &http2.Server{}),
	}

	// Start the server on a new go-thread
	go func() {
		logger.Info("Starting server", "address", config.address)
		if err := s.Serve(conn); err != nil && err != http.ErrServerClosed {
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

	cancelCtx, cancelFunc := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelFunc()

	if err := s.Shutdown(cancelCtx); err != nil {
		logger.Error(err, "failed to gracefully shut down serer", err)
	}

	logger.Info("server gracefully shut down")
}

// EnvOrDefault is a generic implementation that returns either the environment variable accessed by `key`
// or the default value.
func EnvOrDefault[T string | int | bool](key string, def T) T {
	env, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	var ret T
	switch p := any(&ret).(type) {
	case *string:
		*p = env
	case *int:
		*p, _ = strconv.Atoi(env)
	case *bool:
		*p, _ = strconv.ParseBool(env)
	}
	return ret
}

func isGrpc(r *http.Request) bool {
	//if r.ProtoAtLeast(2, 0) && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
	//	return true
	//}
	if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
		return true
	}
	return false
}
