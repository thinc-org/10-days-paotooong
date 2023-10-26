package main

import (
	"context"
	"log"
	"net"

	"github.com/thinc-org/10-days-paotooong/gen/ent"
	_ "github.com/lib/pq"
	genauth "github.com/thinc-org/10-days-paotooong/gen/proto/auth/v1"
	"github.com/thinc-org/10-days-paotooong/internal/auth"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8181")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	ctx := context.Background()
	dbClient, err := ent.Open("postgres", "postgres://postgres:123456@localhost:5432/paotooong?sslmode=disable")
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer dbClient.Close()
	
	authSvc := auth.NewService(dbClient)

	genauth.RegisterAuthServiceServer(server, authSvc)

	if err = dbClient.Schema.Create(ctx); err != nil {
		log.Fatalf("unable to migrate: %v", err)
	}

	if err := server.Serve(lis); err != nil {
		log.Fatalf("server unexpectedly failed: %v", err)
	}
}
