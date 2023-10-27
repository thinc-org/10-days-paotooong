package main

import (
	"errors"
	"fmt"
	"log"

	"context"
	"net"

	_ "github.com/lib/pq"
	"github.com/thinc-org/10-days-paotooong/config"
	"github.com/thinc-org/10-days-paotooong/gen/ent"
	genauth "github.com/thinc-org/10-days-paotooong/gen/proto/auth/v1"
	genwallet "github.com/thinc-org/10-days-paotooong/gen/proto/wallet/v1"
	"github.com/thinc-org/10-days-paotooong/internal/auth"
	"github.com/thinc-org/10-days-paotooong/internal/interceptor"
	"github.com/thinc-org/10-days-paotooong/internal/token"
	"github.com/thinc-org/10-days-paotooong/internal/user"
	"github.com/thinc-org/10-days-paotooong/internal/wallet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func run() error {
	config, err := config.LoadGrpcConfig()
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", config.Port))
	if err != nil {
		return errors.New(fmt.Sprintf("unable to listen to the port: %v", err))
	}
	tokenSvc := token.NewService(([]byte)("5555"), 3600)
	authInterceptor := interceptor.NewAuthInterceptor(tokenSvc)
	server := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor.Unary()))
	ctx := context.Background()
	dbClient, err := ent.Open("postgres", config.DbConnectionString)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to connect to database: %v", err))
	}
	defer dbClient.Close()

	userRepo := user.NewRepository(dbClient)

	authSvc := auth.NewService(dbClient, tokenSvc, userRepo)
	walletSvc := wallet.NewService(dbClient, userRepo)

	genauth.RegisterAuthServiceServer(server, authSvc)
	genwallet.RegisterWalletServiceServer(server, walletSvc)
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	if err = dbClient.Schema.Create(ctx); err != nil {
		return errors.New(fmt.Sprintf("unable to migrate: %v", err))
	}

	log.Printf("start listening grpc service on port %v", config.Port)
	if err := server.Serve(lis); err != nil {
		return errors.New(fmt.Sprintf("server unexpectedly failed: %v", err))
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
