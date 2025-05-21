package cmd

import (
	"github.com/gin-gonic/gin"
	"log"
	"tasker_api/task/internal/handler"
	"tasker_api/task/internal/model"
)

func main() {
	db, err := model.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	// создаём обработчик с доступом к БД
	taskHandler := handler.NewTaskHandler(db)

	r.GET("/tasks", taskHandler.GetTasks)
	r.POST("/tasks", taskHandler.AddTask)
	r.DELETE("/tasks", taskHandler.DeleteTask)

	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
