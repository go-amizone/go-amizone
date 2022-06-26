package server

import (
	"context"
	"encoding/base64"
	"github.com/ditsuke/go-amizone/amizone"
	v1 "github.com/ditsuke/go-amizone/server/gen/go/v1"
	"github.com/go-logr/logr"
	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"net"
	"net/http"
	"strings"
)

const ContextAmizoneClientKey = "amizone_client"

type Config struct {
	Logger       logr.Logger
	BindAddr     string
	WellKnownDir string
}

func NewConfig() *Config {
	return &Config{
		BindAddr: "127.0.0.1:8081",
		Logger:   logr.Discard(),
	}
}

type ApiServer struct {
	Config     *Config
	Router     http.Handler
	httpServer *http.Server
}

func New(config *Config) *ApiServer {
	return &ApiServer{
		Config: config,
	}
}

func (s *ApiServer) getGrpcServer() *grpc.Server {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpcAuth.UnaryServerInterceptor(authorizeCtx)))
	v1.RegisterAmizoneServiceServer(grpcServer, NewAmizoneServiceServer())
	reflection.Register(grpcServer)
	return grpcServer
}

func (s *ApiServer) getHttpMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve the "well_known" directory for certificate signing.
	if s.Config.WellKnownDir != "" {
		fs := http.FileServer(http.Dir(s.Config.WellKnownDir))
		mux.HandleFunc("/.well_known/", func(writer http.ResponseWriter, request *http.Request) {
			// Rewrite the path to the file to be served.
			request.URL.Path = strings.TrimPrefix(request.URL.Path, "/.well_known/")
			fs.ServeHTTP(writer, request)
		})
		s.Config.Logger.Info("Serving .well-known directory", "dir", s.Config.WellKnownDir)
	} else {
		s.Config.Logger.Info("Not serving .well-known directory")
	}
	// grpc-gateway
	gwMux := runtime.NewServeMux()

	// @todo configurable port here
	err := v1.RegisterAmizoneServiceHandlerFromEndpoint(context.Background(), gwMux, "localhost:8081", []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		s.Config.Logger.Error(err, "Failed to register grpc-gateway")
	}
	mux.HandleFunc("/api/", func(writer http.ResponseWriter, request *http.Request) {
		gwMux.ServeHTTP(writer, request)
	})
	return mux
}

func (s *ApiServer) Run() error {
	// @todo configurable (optional) tls wrapping here

	if s.Router == nil {
		s.Config.Logger.V(1).Info("Configuring router...")
		s.configureRouter()
	}

	if s.httpServer == nil {
		s.Config.Logger.V(1).Info("Initializing server...")
		s.httpServer = &http.Server{
			Addr:    s.Config.BindAddr,
			Handler: h2c.NewHandler(s.Router, &http2.Server{}), // @todo make this configurable
		}
	}

	conn, err := net.Listen("tcp", s.Config.BindAddr)
	if err != nil {
		s.Config.Logger.Error(err, "Failed to listen", "addr", s.Config.BindAddr)
		return err
	}

	s.Config.Logger.Info("Starting server", "bind_addr", s.Config.BindAddr)
	return s.httpServer.Serve(conn)
}

func (s *ApiServer) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *ApiServer) configureRouter() {
	grpcServer := s.getGrpcServer()
	httpMux := s.getHttpMux()
	s.Router = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if isGrpc(request) {
			grpcServer.ServeHTTP(writer, request)
			return
		}
		httpMux.ServeHTTP(writer, request)
	})
}

// isGrpc returns true if the request is for a gRPC endpoint.
func isGrpc(r *http.Request) bool {
	if r.ProtoAtLeast(2, 0) && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
		return true
	}
	return false
}

// authorizeCtx is a grpcAuth.
func authorizeCtx(ctx context.Context) (context.Context, error) {
	credentialsEncoded, err := grpcAuth.AuthFromMD(ctx, "basic")
	if err != nil {
		return ctx, err
	}

	// Base 64 decode
	credentials, err := base64.StdEncoding.DecodeString(credentialsEncoded)
	if err != nil {
		return ctx, err
	}

	index := strings.IndexByte(string(credentials), ':')
	if index == -1 || index == 0 || index == len(credentials)-1 {
		return ctx, status.Errorf(codes.Unauthenticated, "bad auth string")
	}

	user, pass := string(credentials[:index]), string(credentials[index+1:])

	client, err := amizone.NewClient(amizone.Credentials{Username: user, Password: pass}, nil)
	if err != nil {
		return ctx, status.Errorf(codes.Unauthenticated, "amizone: ", err.Error())
	}

	newCtx := context.WithValue(ctx, ContextAmizoneClientKey, client)
	return newCtx, nil
}
