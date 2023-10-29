package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/thinc-org/10-days-paotooong/config"
	"github.com/thinc-org/10-days-paotooong/docs"
	auth "github.com/thinc-org/10-days-paotooong/gen/proto/auth/v1"
	wallet "github.com/thinc-org/10-days-paotooong/gen/proto/wallet/v1"
	"github.com/thinc-org/10-days-paotooong/static"
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	config, err := config.LoadProxyConfig()
	if err != nil {
		return err
	}

	grpcConn, err := grpc.Dial(config.GrpcUrl, grpc.WithInsecure())
	if err != nil {
		return err
	}

	docsHandlerFunc := docs.GetDocHandler()
	staticHandlerFunc := static.GetDocHandler()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	healthClient := grpc_health_v1.NewHealthClient(grpcConn)
	mux := runtime.NewServeMux(runtime.WithHealthEndpointAt(healthClient, "/health"))
	mux.HandlePath("GET", "/v1/docs", docsHandlerFunc)
	mux.HandlePath("GET", "/static/*", staticHandlerFunc)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err = auth.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, config.GrpcUrl, opts)
	if err != nil {
		return err
	}

	err = wallet.RegisterWalletServiceHandlerFromEndpoint(ctx, mux, config.GrpcUrl, opts)
	if err != nil {
		return err
	}
	// Start HTTP server (and proxy calls to gRPC server endpoint)

	withCors := cors.New(cors.Options{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"ACCEPT", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler(mux)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%v", config.Port),
		Handler: withCors,
	}

	log.Printf("start listening http proxy on port %v", config.Port)
	return srv.ListenAndServe()
}

func main() {
	if err := run(); err != nil {
		grpclog.Fatal(err)
	}
}
