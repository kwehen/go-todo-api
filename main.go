package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kwehen/go-todo-api/internal/auth"
	_ "github.com/lib/pq"
	"github.com/markbates/goth/gothic"
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

type User struct {
	RawData           map[string]interface{}
	Provider          string
	Email             string
	Name              string
	FirstName         string
	LastName          string
	NickName          string
	Description       string
	UserID            string
	AvatarURL         string
	Location          string
	AccessToken       string
	AccessTokenSecret string
	RefreshToken      string
	ExpiresAt         time.Time
	IDToken           string
}

var db *sql.DB

var sessionStore = make(map[string]string)

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

	auth.NewAuth()
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	authorized := router.Group("/")
	authorized.Use(auth.AuthMiddleware())
	{
		authorized.GET("/home", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", nil)
		})
		authorized.GET("/tasks", getTask)
		authorized.GET("/tasks/:id", getTaskByID)
		authorized.DELETE("/delete/:id", deleteTask)
		authorized.POST("/tasks", addTask)
		authorized.POST("/completeTask/:id", completeTask)
		authorized.GET("/completed/:id", completeTaskDeleteFromTasks)
		authorized.POST("/completed/:id", addToCompletedTable)
		authorized.GET("/completed", getCompletedTasks)
		router.GET("/auth/:provider", handleGoogleAuth)
		router.GET("/auth/:provider/callback", handleGoogleCallback)
		router.GET("/logout", googleLogout)
		router.GET("/login", googleLogin)
	}
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

	c.IndentedJSON(http.StatusCreated, newTask)
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
	c.HTML(http.StatusOK, "gettaskid.html", t)
}

func completeTaskDeleteFromTasks(c *gin.Context) {
	id := c.Param("id")

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// First SQL command: Update
	if _, err := tx.Exec("UPDATE tasks SET completed = true WHERE task_id = $1", id); err != nil {
		tx.Rollback()
		log.Println("Error executing update statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Second SQL command: Insert
	if _, err := tx.Exec("INSERT INTO completed(task) SELECT task FROM tasks WHERE task_id = $1", id); err != nil {
		tx.Rollback()
		log.Println("Error executing insert statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Third SQL command: Delete
	if _, err := tx.Exec("DELETE FROM tasks WHERE task_id = $1", id); err != nil {
		tx.Rollback()
		log.Println("Error executing delete statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Task completed, deleted from tasks table, and added to completed table"})
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
	c.Header("Content-Type", "text/html")

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
	c.HTML(http.StatusOK, "completed.html", tasks)
}

// func handleGoogleAuth(c *gin.Context) {
// 	gothic.BeginAuthHandler(c.Writer, c.Request)
// }

func handleGoogleAuth(c *gin.Context) {
	provider := c.Param("provider")

	r := c.Request.WithContext(context.WithValue(c.Request.Context(), gothic.ProviderParamKey, provider))
	gothic.BeginAuthHandler(c.Writer, r) // Use the new request r here
}

// func googleAuth(c *gin.Context) {
// 	err := gothic.BeginAuthHandler(c.Writer, c.Request)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// }

func handleGoogleCallback(c *gin.Context) {
	provider := c.Param("provider")
	w := c.Writer
	r := c.Request

	r = r.WithContext(context.WithValue(r.Context(), gothic.ProviderParamKey, provider))
	// Set the provider in the request context

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Println("Error during CompleteUserAuth:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate a secure random string for the session token
	b := make([]byte, 32)
	_, err = rand.Read(b)
	if err != nil {
		c.Error(err)
		return
	}

	sessionToken := base64.StdEncoding.EncodeToString(b)

	// // Store the session token and OAuth token in your database
	// storeSession(sessionToken, user.AccessToken)

	// Set a secure cookie with the session token
	c.SetCookie("user", sessionToken, 3600, "/", ".kamaufoundation.com", true, true)

	log.Println("Logged in as:", user.Name)
	log.Println("Email:", user.Email)
	c.Redirect(http.StatusMovedPermanently, "/home")
}

func googleLogout(c *gin.Context) {
	gothic.Logout(c.Writer, c.Request)

	// Get the session token from the cookie
	// sessionToken, err := c.Cookie("user")
	// if err != nil {
	// 	c.Error(err)
	// 	return
	// }

	// // Delete the session from your database
	// // Assuming you have a function deleteSession that does this
	// deleteSession(sessionToken)

	c.SetCookie("user", "", -1, "/", ".kamaufoundation.com", true, true)
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func googleLogin(c *gin.Context) {
	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request); err == nil {
		c.JSON(200, gin.H{"user": gothUser})
	} else {
		gothic.BeginAuthHandler(c.Writer, c.Request)
	}
}

func storeSession(sessionToken string, oauthToken string) {
	sessionStore[sessionToken] = oauthToken
}

func deleteSession(sessionToken string) {
	delete(sessionStore, sessionToken)
}

func getOAuthToken(sessionToken string) (string, bool) {
	oauthToken, exists := sessionStore[sessionToken]
	return oauthToken, exists
}
