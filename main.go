package main

import (
	"database/sql"
	"log"
	"net/http"

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

type completed struct {
	ID   string `json:"id"`
	Task string `json:"task"`
}

var db *sql.DB

func main() {
	var err error
	// dbUser := os.Getenv("DB_USER")
	// dbPassword := os.Getenv("DB_PASSWORD")
	// dbName := os.Getenv("DB_NAME")
	// dbHost := os.Getenv("DB_HOST")
	// dbPort := os.Getenv("DB_PORT")

	dbConnectionString := "postgres://postgres:postgres@10.0.0.9:5432/postgres?sslmode=disable"

	// Open a connection to the database
	db, err = sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	router.GET("/tasks", getTask)
	router.GET("/tasks/:id", getTaskByID)
	router.DELETE("/delete/:id", deleteTask)
	router.POST("/tasks", addTask)
	router.GET("/completed/:id", completeTask)
	router.POST("/completed/:id", addToCompletedTable)
	router.GET("/completed", getCompletedTasks)

	router.Run("0.0.0.0:8080")
}

func getTask(c *gin.Context) {
	c.Header("Content-Type", "text/html")

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

	c.HTML(http.StatusOK, "tasks.html", tasks)
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

func deleteTask(c *gin.Context) {
	id := c.Param("id")

	stmt, err := db.Prepare("DELETE FROM tasks WHERE task_id = $1")
	if err != nil {
		log.Println("Error preparing SQL statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(id); err != nil {
		log.Println("Error executing SQL statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

func getTaskByID(c *gin.Context) {
	id := c.Param("id")

	rows, err := db.Query("SELECT * FROM tasks WHERE task_id = $1", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var t task
	for rows.Next() {
		if err := rows.Scan(&t.ID, &t.Task, &t.Urgency, &t.Hours, &t.Completed); err != nil {
			log.Fatal(err)
		}
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusOK, t)
}

func completeTask(c *gin.Context) {
	id := c.Param("id")

	stmt, err := db.Prepare("UPDATE tasks SET completed = true WHERE task_id = $1")
	if err != nil {
		log.Println("Error preparing SQL statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(id); err != nil {
		log.Println("Error executing SQL statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Task completed"})
}

func addToCompletedTable(c *gin.Context) {
	id := c.Param("id")

	stmt, err := db.Prepare("INSERT INTO completed(task) SELECT task FROM tasks WHERE task_id = $1")
	if err != nil {
		log.Println("Error preparing SQL statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(id); err != nil {
		log.Println("Error executing SQL statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Task added to completed table"})
}

func getCompletedTasks(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	rows, err := db.Query("SELECT * FROM completed")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tasks []completed
	for rows.Next() {
		var t completed
		if err := rows.Scan(&t.ID, &t.Task); err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusOK, tasks)
}
