package main

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"log"
	"task/internal/handler"
	"task/internal/model"
	"task/pkg/userpb"
)

func main() {
	db, err := model.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	// gRPC подключение к user-сервису
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to user service: %v", err)
	}
	defer conn.Close()
	userClient := userpb.NewUserServiceClient(conn)

	r := gin.Default()
	taskHandler := handler.NewTaskHandler(db, userClient)

	r.GET("/tasks", taskHandler.GetTasks)
	r.POST("/tasks", taskHandler.AddTask)
	r.DELETE("/tasks", taskHandler.DeleteTask)

	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
