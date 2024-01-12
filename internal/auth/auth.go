package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
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
		// google.New(googleClientId, googleClientSecret, "https://test.home.kamaufoundation.com/auth/google/callback", "email", "profile"),
		google.New(googleClientId, googleClientSecret, "http://localhost:8080/auth/google/callback", "email", "profile"),
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

func Encrypt(plainText string, key string) (string, error) {
	// Convert the key to bytes
	keyBytes := []byte(key)

	// Create a new cipher block
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// Create a new GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create a new nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)

	// Base64 encode the ciphertext
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(encryptedString string, keyString string) (string, error) {
	// Convert the key to bytes directly, without base64 decoding
	key := []byte(keyString)

	enc, err := base64.StdEncoding.DecodeString(encryptedString)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(enc) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
