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
		// google.New(googleClientId, googleClientSecret, "http://localhost:8080/auth/google/callback", "email", "profile"),
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

// func Encrypt(stringToEncrypt string, keyString string) (encryptedString string) {
// 	// Since the key is in string, we need to convert decode it to bytes
// 	key, _ := base64.StdEncoding.DecodeString(keyString)
// 	plaintext := []byte(stringToEncrypt)

// 	// Create a new Cipher Block from the key
// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	// Create a new GCM
// 	aesGCM, err := cipher.NewGCM(block)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	// Create a nonce. Nonce should be from GCM
// 	nonce := make([]byte, aesGCM.NonceSize())
// 	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
// 		panic(err.Error())
// 	}

// 	// Encrypt the data using aesGCM.Seal
// 	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
// 	return base64.StdEncoding.EncodeToString(ciphertext)
// }

// func Decrypt(encryptedString string, keyString string) (decryptedString string) {
// 	key, _ := base64.StdEncoding.DecodeString(keyString)
// 	enc, _ := base64.StdEncoding.DecodeString(encryptedString)

// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	aesGCM, err := cipher.NewGCM(block)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	nonceSize := aesGCM.NonceSize()
// 	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]
// 	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	return string(plaintext)
// }
