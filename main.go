package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
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
	UserID    string  `json:"user_id"`
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

// var sessionStore = make(map[string]string)

func main() {
	var err error
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	// dbPort := os.Getenv("DB_PORT")

	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=disable", dbUser, dbPassword, dbName, dbHost))
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error: Could not establish a connection with the database")
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
		// authorized.POST("/completed/:id", addToCompletedTable)
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
	email, err := c.Cookie("email")
	if err != nil {
		// Handle error
		c.JSON(http.StatusBadRequest, gin.H{"error": "No email cookie found"})
		return
	}
	// decrypt email
	decryptedEmail, err := auth.Decrypt(email, os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Printf("Error decrypting email: %v\n", err)
	}

	rows, err := db.Query("SELECT * FROM tasks WHERE user_id = $1", decryptedEmail)
	if err != nil {
		log.Printf("Error querying database: %v\n", err)
	}
	defer rows.Close()

	// decrypt task before displaying

	var tasks []task
	for rows.Next() {
		var t task
		if err := rows.Scan(&t.ID, &t.Task, &t.Urgency, &t.Hours, &t.Completed, &t.UserID); err != nil {
			log.Printf("Error scanning rows: %v\n", err)
		}

		decryptedTask, err := auth.Decrypt(t.Task, os.Getenv("SECRET_KEY"))
		if err != nil {
			log.Printf("Error decrypting task: %v\n", err)
		}
		t.Task = decryptedTask
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v\n", err)
	}

	c.HTML(http.StatusOK, "tasks.html", tasks)
}

func addTask(c *gin.Context) {
	var newTask task
	if err := c.BindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Payload"})
		return
	}

	email, err := c.Cookie("email")
	if err != nil {
		// Handle error
		c.JSON(http.StatusBadRequest, gin.H{"error": "No email cookie found"})
		return
	}

	newTask.UserID = email

	stmt, err := db.Prepare("INSERT INTO tasks(task, urgency, hours, completed, user_id) VALUES($1, $2, $3, $4, $5)")
	if err != nil {
		log.Println("Error preparing SQL statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer stmt.Close()

	// encrypt task before inserting into database
	encryptedTask, err := auth.Encrypt(newTask.Task, os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Println("Error encrypting task:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// decrypt email
	decryptedEmail, err := auth.Decrypt(email, os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Printf("Error decrypting email: %v\n", err)
	}

	if _, err := stmt.Exec(encryptedTask, newTask.Urgency, newTask.Hours, newTask.Completed, decryptedEmail); err != nil {
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
		log.Printf("Error querying database: %v\n", err)
	}
	defer rows.Close()

	var t task
	for rows.Next() {
		if err := rows.Scan(&t.ID, &t.Task, &t.Urgency, &t.Hours, &t.Completed); err != nil {
			log.Printf("Error scanning rows: %v\n", err)
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v\n", err)
	}
	c.HTML(http.StatusOK, "gettaskid.html", t)
}

func completeTaskDeleteFromTasks(c *gin.Context) {
	id := c.Param("id")
	email, err := c.Cookie("email")
	if err != nil {
		// Handle error
		c.JSON(http.StatusBadRequest, gin.H{"error": "No email cookie found"})
		return
	}
	// decrypt email
	decryptedEmail, err := auth.Decrypt(email, os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Printf("Error decrypting email: %v\n", err)
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// First SQL command: Update
	if _, err := tx.Exec("UPDATE tasks SET completed = true WHERE task_id = $1 AND user_id = $2", id, decryptedEmail); err != nil {
		tx.Rollback()
		log.Println("Error executing update statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Second SQL command: Insert
	if _, err := tx.Exec("INSERT INTO completed(task, user_id) SELECT task, user_id FROM tasks WHERE task_id = $1 AND user_id = $2", id, decryptedEmail); err != nil {
		tx.Rollback()
		log.Println("Error executing insert statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Third SQL command: Delete
	if _, err := tx.Exec("DELETE FROM tasks WHERE task_id = $1 AND user_id = $2", id, decryptedEmail); err != nil {
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
	email, err := c.Cookie("email")
	if err != nil {
		// Handle error
		c.JSON(http.StatusBadRequest, gin.H{"error": "No email cookie found"})
		return
	}
	// decrypt email
	decryptedEmail, err := auth.Decrypt(email, os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Printf("Error decrypting email: %v\n", err)
	}

	stmt, err := db.Prepare("UPDATE tasks SET completed = true WHERE task_id = $1 AND user_id = $2")
	if err != nil {
		log.Println("Error preparing SQL statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(id, decryptedEmail); err != nil {
		log.Println("Error executing SQL statement:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Task completed"})
}

// func addToCompletedTable(c *gin.Context) {
// 	id := c.Param("id")

// 	stmt, err := db.Prepare("INSERT INTO completed(task) SELECT task FROM tasks WHERE task_id = $1")
// 	if err != nil {
// 		log.Println("Error preparing SQL statement:", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
// 		return
// 	}
// 	defer stmt.Close()

// 	if _, err := stmt.Exec(id); err != nil {
// 		log.Println("Error executing SQL statement:", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
// 		return
// 	}

// 	c.IndentedJSON(http.StatusOK, gin.H{"message": "Task added to completed table"})
// }

func getCompletedTasks(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	email, err := c.Cookie("email")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No email cookie found"})
		return
	}
	decryptedEmail, err := auth.Decrypt(email, os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Printf("Error decrypting email: %v\n", err)
	}

	rows, err := db.Query("SELECT * FROM completed WHERE user_id = $1", decryptedEmail)
	if err != nil {
		log.Printf("Error querying database: %v\n", err)
	}
	defer rows.Close()

	var tasks []completed
	for rows.Next() {
		var t completed
		if err := rows.Scan(&t.ID, &t.Task); err != nil {
			log.Printf("Error scanning rows: %v\n", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v\n", err)
	}
	c.HTML(http.StatusOK, "completed.html", tasks)
}

func handleGoogleAuth(c *gin.Context) {
	provider := c.Param("provider")

	r := c.Request.WithContext(context.WithValue(c.Request.Context(), gothic.ProviderParamKey, provider))
	gothic.BeginAuthHandler(c.Writer, r)
}

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
	encryptedEmail, err := auth.Encrypt(user.Email, os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Println("Error encrypting email:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	// // Store the session token and OAuth token in your database
	// storeSession(sessionToken, user.AccessToken)

	// Set a secure cookie with the session token
	c.SetCookie("user", sessionToken, 3600, "/", ".kamaufoundation.com", true, true)
	c.SetCookie("email", encryptedEmail, 3600, "/", ".kamaufoundation.com", true, true)
	// c.SetCookie("user", sessionToken, 3600, "/", "localhost", true, true)
	// c.SetCookie("email", encryptedEmail, 3600, "/", "localhost", true, true)

	log.Println("Logged in as:", user.Name)
	log.Println("Email:", user.Email)
	c.Redirect(http.StatusMovedPermanently, "/home")
}

func googleLogout(c *gin.Context) {
	gothic.Logout(c.Writer, c.Request)

	c.SetCookie("user", "", -1, "/", ".kamaufoundation.com", true, true)
	c.SetCookie("email", "", -1, "/", ".kamaufoundation.com", true, true)
	// c.SetCookie("user", "", -1, "/", "localhost", true, true)
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

// func storeSession(sessionToken string, oauthToken string) {
// 	sessionStore[sessionToken] = oauthToken
// }

// func deleteSession(sessionToken string) {
// 	delete(sessionStore, sessionToken)
// }

// func getOAuthToken(sessionToken string) (string, bool) {
// 	oauthToken, exists := sessionStore[sessionToken]
// 	return oauthToken, exists
// }
