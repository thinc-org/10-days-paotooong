package main

import (
  "context"
  "flag"
  "net/http"

  "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials/insecure"
  "google.golang.org/grpc/grpclog"

  auth "github.com/thinc-org/10-days-paotooong/gen/auth/v1"
  wallet "github.com/thinc-org/10-days-paotooong/gen/wallet/v1"
)

var (
  // command-line options:
  // gRPC server endpoint
  grpcServerEndpoint = flag.String("grpc-server-endpoint",  "localhost:8181", "gRPC server endpoint")
)

func run() error {
  ctx := context.Background()
  ctx, cancel := context.WithCancel(ctx)
  defer cancel()

  // Register gRPC server endpoint
  // Note: Make sure the gRPC server is running properly and accessible
  mux := runtime.NewServeMux()
  opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
  err := auth.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
  if err != nil {
    return err
  }

  err = wallet.RegisterWalletServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
  if err != nil {
    return err
  }
  // Start HTTP server (and proxy calls to gRPC server endpoint)
  return http.ListenAndServe(":8080", mux)
}

func main() {
  flag.Parse()

  if err := run(); err != nil {
    grpclog.Fatal(err)
  }
}
