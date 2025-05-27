package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"tasker_api/auth/internal/handler"
	"tasker_api/auth/pkg/auth_user_pb"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		log.Fatalf("failed to connect to the user service: %v", err)
	}
	defer conn.Close()

	authClient := auth_user_pb.NewAuthServiceClient(conn)

	authHandler := handler.NewAuthHandler(authClient)

	router := gin.Default()

	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to launch the server: %v", err)
	}
}
