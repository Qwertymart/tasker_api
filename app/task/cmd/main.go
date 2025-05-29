package main

import (
	"log"
	"os"
	"task/internal/handler"
	"task/internal/middleware"
	"task/internal/model"
	"task/pkg/userpb"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	db, err := model.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	userClient := userpb.NewUserServiceClient(conn)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	r := gin.Default()

	authMiddleware := middleware.AuthMiddleware(secret, userClient)

	taskHandler := handler.NewTaskHandler(db, userClient)

	authorized := r.Group("/")
	authorized.Use(authMiddleware)
	{
		authorized.GET("/tasks", taskHandler.GetTasks)
		authorized.POST("/tasks", taskHandler.AddTask)
		authorized.DELETE("/tasks", taskHandler.DeleteTask)
	}

	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
