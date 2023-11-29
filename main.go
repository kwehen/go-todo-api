package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type task struct {
	ID      string  `json:"id"`
	ToDo    string  `json:"toDO"`
	Urgency string  `json:"urgency"`
	EstTime float64 `json:"hours"`
}

var tasks = []task{
	{ID: "1", ToDo: "Add web frontend to GPT API", Urgency: "Low", EstTime: 4.0},
	{ID: "2", ToDo: "Consume API with GO", Urgency: "Medium", EstTime: 5.0},
	{ID: "3", ToDo: "Reconfigure k3s cluster", Urgency: "Medium", EstTime: 2.5},
}

func main() {
	router := gin.Default()
	router.GET("/tasks", getTask)
	router.GET("/tasks/:id", getTaskByID)
	router.DELETE("/tasks/:id", deleteTask)
	router.POST("/tasks", addTask)

	router.Run("localhost:8080")
}

func getTask(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, tasks)
}

func addTask(c *gin.Context) {
	var newTask task

	if err := c.BindJSON(&newTask); err != nil {
		return
	}

	tasks = append(tasks, newTask)
	c.IndentedJSON(http.StatusCreated, newTask)
}

func getTaskByID(c *gin.Context) {
	id := c.Param("id")

	for _, task := range tasks {
		if task.ID == id {
			c.IndentedJSON(http.StatusOK, task)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Task not found"})
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")

	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			c.IndentedJSON(http.StatusOK, gin.H{"message": "Task deleted"})
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Task not found"})
}
