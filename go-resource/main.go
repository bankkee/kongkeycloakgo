package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Nerzal/gocloak/v11"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func main() {
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
	if err != nil {
		log.Printf("Login failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": token.AccessToken})
}
