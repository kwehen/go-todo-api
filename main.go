package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type task struct {
	ID        string  `json:"id"`
	Task      string  `json:"task"`
	Urgency   string  `json:"urgency"`
	Hours     float64 `json:"hours"`
	Completed bool    `json:"completed"`
}

// var tasks = []task{
// 	{ID: "1", ToDo: "Add web frontend to GPT API", Urgency: "Low", EstTime: 4.0, Completed: false},
// 	{ID: "2", ToDo: "Consume API with GO", Urgency: "Medium", EstTime: 5.0, Completed: false},
// 	{ID: "3", ToDo: "Reconfigure k3s cluster", Urgency: "Medium", EstTime: 2.5, Completed: false},
// }

var db *sql.DB

func main() {
	var err error
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	dbConnectionString := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"

	// Open a connection to the database
	db, err = sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.GET("/tasks", getTask)
	// router.GET("/tasks/:id", getTaskByID)
	// router.DELETE("/tasks/:id", deleteTask)
	router.POST("/tasks", addTask)

	router.Run("0.0.0.0:8080")
}

func getTask(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	rows, err := db.Query("SELECT * FROM tasks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tasks []task
	for rows.Next() {
		var t task
		if err := rows.Scan(&t.ID, &t.Task, &t.Urgency, &t.Hours, &t.Completed); err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusOK, tasks)
}

func addTask(c *gin.Context) {
	var newTask task
	if err := c.BindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Payload"})
		return
	}

	stmt, err := db.Prepare("INSERT INTO tasks(task, urgency, hours, completed) VALUES($1, $2, $3, $4)")
	if err != nil {
		log.Println("Error preparing SQL statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(newTask.Task, newTask.Urgency, newTask.Hours, newTask.Completed); err != nil {
		log.Println("Error executing SQL statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusCreated, newTask)
}

// func getTaskByID(c *gin.Context) {
// 	id := c.Param("id")

// 	for _, task := range tasks {
// 		if task.ID == id {
// 			c.IndentedJSON(http.StatusOK, task)
// 			return
// 		}
// 	}
// 	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Task not found"})
// }

// func deleteTask(c *gin.Context) {
// 	id := c.Param("id")

// 	for i, task := range tasks {
// 		if task.ID == id {
// 			tasks = append(tasks[:i], tasks[i+1:]...)
// 			c.IndentedJSON(http.StatusOK, gin.H{"message": "Task deleted"})
// 			return
// 		}
// 	}
// 	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Task not found"})
// }
