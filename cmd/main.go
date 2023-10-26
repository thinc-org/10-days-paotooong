package main

import (
	"context"
	"log"
	"net"

	"github.com/thinc-org/10-days-paotooong/gen/ent"
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
	dbClient, err := ent.Open("postgres", "root@localhost:5432/paotooong")
	defer dbClient.Close()
	
	authSvc := auth.NewService(dbClient)

	genauth.RegisterAuthServiceServer(server, authSvc)

	err = dbClient.Schema.Create(
		ctx,
	)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("server unexpectedly failed: %v", err)
	}
}
