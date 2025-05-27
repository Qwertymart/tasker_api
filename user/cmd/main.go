package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"user/internal/model"
	"user/pkg/auth_user_pb"
	"user/pkg/userpb"
	"user/transport"
)

func main() {
	db, err := model.ConnectDB()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	grpcServer := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcServer, transport.NewUserServiceServer(db))
	auth_user_pb.RegisterAuthServiceServer(grpcServer, transport.NewUserAuthServer(db))

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Starting gRPC server on :50051")
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
