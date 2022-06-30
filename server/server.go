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
	"sync"
)

const ContextAmizoneClientKey = "amizone_client"

// Config is the configuration entity for ApiServer.
type Config struct {
	Logger       logr.Logger
	BindAddr     string
	WellKnownDir string
}

// NewConfig returns a Config with sensible defaults and a logr.Discard logger.
func NewConfig(bindAddress string) *Config {
	return &Config{
		BindAddr:     bindAddress,
		Logger:       logr.Discard(),
		WellKnownDir: "",
	}
}

// ApiServer implements an HTTP + gRPC API interface for the go-amizone SDK.
type ApiServer struct {
	router http.Handler
	muInit struct {
		done bool
		sync.Mutex
	}
	config     *Config
	httpServer *http.Server
}

func New(config *Config) *ApiServer {
	return &ApiServer{
		config: config,
	}
}

// Init initialises the server. It is usually called internally by ListenAndServe or ServeHTTP.
func (s *ApiServer) Init() {
	s.muInit.Lock()
	defer s.muInit.Unlock()
	if s.muInit.done {
		return
	}
	s.config.Logger.V(1).Info("Configuring server and router...")
	s.router = h2c.NewHandler(s.newRouter(), &http2.Server{})
	s.httpServer = &http.Server{
		Addr:    s.config.BindAddr,
		Handler: s.router,
	}
	s.muInit.done = true
	return
}

// ServeHTTP implements the http.Handler interface for ApiServer.
func (s *ApiServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if !s.muInit.done {
		s.Init()
	}
	s.router.ServeHTTP(writer, request)
}

// ListenAndServe starts the server on Config.BindAddr and blocks until it is stopped. The error returned is consistent with the
// error returned by http.Server.ListenAndServe.
func (s *ApiServer) ListenAndServe() error {
	if !s.muInit.done {
		s.Init()
	}
	// @todo optional tls?
	s.config.Logger.Info("Starting server", "bind_addr", s.config.BindAddr)
	return s.httpServer.ListenAndServe()
}

// Stop stops the server.
func (s *ApiServer) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// newRouter creates a new router for the ApiServer that routes gRPC and HTTP requests to
// routers configured by the newGrpcServer and newHttpMux functions.
func (s *ApiServer) newRouter() http.Handler {
	grpcServer := s.newGrpcServer()
	httpMux := s.newHttpMux()

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if isGrpc(request) {
			grpcServer.ServeHTTP(writer, request)
			return
		}
		httpMux.ServeHTTP(writer, request)
	})
}

func (s *ApiServer) newGrpcServer() *grpc.Server {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpcAuth.UnaryServerInterceptor(authorizeCtx)))
	v1.RegisterAmizoneServiceServer(grpcServer, NewAmizoneServiceServer())
	reflection.Register(grpcServer)
	return grpcServer
}

// newHttpMux creates a new multiplexer for the ApiServer that routes gRPC and HTTP requests based on the server config.
func (s *ApiServer) newHttpMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve the "well_known" directory for certificate signing.
	if s.config.WellKnownDir != "" {
		fs := http.FileServer(http.Dir(s.config.WellKnownDir))
		mux.HandleFunc("/.well_known/", func(writer http.ResponseWriter, request *http.Request) {
			// Rewrite the path to the file to be served.
			request.URL.Path = strings.TrimPrefix(request.URL.Path, "/.well_known/")
			fs.ServeHTTP(writer, request)
		})
		s.config.Logger.Info("Serving .well-known directory", "dir", s.config.WellKnownDir)
	} else {
		s.config.Logger.Info("Not serving .well-known directory")
	}
	// grpc-gateway
	gwMux := runtime.NewServeMux()

	_, port, err := net.SplitHostPort(s.config.BindAddr)
	if err != nil {
		s.config.Logger.Error(err, "Failed to parse bind port", "addr", s.config.BindAddr)
		// @todo check if caller accommodates for the nil return
		return nil
	}
	err = v1.RegisterAmizoneServiceHandlerFromEndpoint(context.Background(), gwMux, "localhost:"+port, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	if err != nil {
		s.config.Logger.Error(err, "Failed to register grpc-gateway")
	}
	mux.HandleFunc("/api/", func(rw http.ResponseWriter, req *http.Request) {
		gwMux.ServeHTTP(rw, req)
	})
	return mux
}

// isGrpc returns true if the request is a gRPC request.
func isGrpc(r *http.Request) bool {
	if r.ProtoAtLeast(2, 0) && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
		return true
	}
	return false
}

// authorizeCtx is a grpc_auth.AuthFunc. It authorizes the request by checking for
// the (currently) supported Basic auth header and then validating the credentials by
// getting a logged-in instance of amizone.Client.
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
		return ctx, status.Error(codes.Unauthenticated, "amizone: "+err.Error())
	}
	return context.WithValue(ctx, ContextAmizoneClientKey, client), nil
}
