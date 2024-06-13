package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

var rdb *redis.Client

func main() {
	rdb = initRedis()
	router := gin.Default()

	// API สำหรับ /hello ที่ไม่ตรวจสอบโทเค็น
	router.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, this is the /hello endpoint!",
		})
	})

	router.GET("/hello-secure", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, this is the /hello endpoint!",
		})
	})

	router.POST("/login", loginHandler)

	// ticker := time.NewTicker(1 * time.Second)
	// defer ticker.Stop()

	// go func() {
	// 	for range ticker.C {
	// 		log.Println("Logging keys in Redis:")
	// 		iter := rdb.Scan(0, "*", 0).Iterator()
	// 		for iter.Next() {
	// 			key := iter.Val()
	// 			value, _ := rdb.Get(key).Result()
	// 			log.Printf("- Key: %s value: %s\n", key, value)
	// 		}
	// 		if err := iter.Err(); err != nil {
	// 			log.Printf("Error scanning keys: %v", err)
	// 		}
	// 		log.Println("End of keys list")
	// 	}
	// }()

	// เริ่มเซิร์ฟเวอร์
	if err := router.Run(":9999"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func loginHandler(c *gin.Context) {
	var loginReq LoginRequest
	var realm = "my-demo"
	var clientID = "test-secret-auth"
	var clientSecret = "lq5FbPOisF1mpdCcmKQ3J4PbbhV2HJdy"
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	client := gocloak.NewClient("http://192.168.0.128:8080/", gocloak.SetAuthAdminRealms("admin/realms"), gocloak.SetAuthRealms("realms"))

	token, err := client.Login(context.Background(), clientID, clientSecret, realm, loginReq.Username, loginReq.Password)
	log.Printf("Token: %v", token)
	if err != nil {
		log.Printf("Login failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	log.Printf("Access token: %v", token.AccessToken)

	if err := storeAccessToken(loginReq.Username, token.AccessToken); err != nil {
		log.Printf("Failed to store access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": token.AccessToken})
}

func storeAccessToken(key string, token string) error {
	err := rdb.Set(key, token, 30*time.Minute).Err()
	log.Printf("storeAccessToken.error: %v", err)
	return err
}

func initRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "redis:6379",
		// Password: "1234",
	})
}
