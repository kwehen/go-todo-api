package auth

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

const (
	key    = "RandomString"
	MaxAge = 86400 * 30
	IsProd = true
)

var store = sessions.NewCookieStore([]byte(key))

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	store.MaxAge(MaxAge)

	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = IsProd

	gothic.Store = store
}

func NewAuth() {
	googleClientId := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	goth.UseProviders(
		google.New(googleClientId, googleClientSecret, "https://test.home.kamaufoundation.com/auth/google/callback", "email", "profile"),
	)
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from the cookies
		cookie, err := c.Request.Cookie("user")
		if err != nil {
			// User not authenticated, redirect to login
			c.Error(err)
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}

		// Set user info in the context if needed
		c.Set("user", cookie.Value)
		c.Next()
	}
}
