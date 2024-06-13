package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/go-redis/redis"
)

type Config struct {
	Endpoint string
}

type VerifyRequest struct {
	Token string `json:"token" binding:"required"`
}

const Version = "0.1"
const Priority = 1

var rdb *redis.Client

func New() interface{} {
	return &Config{}
}

func (conf *Config) Access(kong *pdk.PDK) {
	auth, err := kong.Request.GetHeader("Authorization")
	if err != nil || auth == "" {
		kong.Response.Exit(http.StatusUnauthorized, []byte("Authorization required"), nil)
		return
	}

	// Call the real auth server
	statusCode, body, _ := callAuthServer(auth)

	if statusCode != http.StatusOK {
		kong.Response.Exit(statusCode, body, nil)
		return
	}
}

// func verifyHandlerKeycloak(token string) (statusCode int, body []byte, err error) {
// 	var realm = "my-demo"
// 	var clientID = "test-secret-auth"
// 	var clientSecret = "lq5FbPOisF1mpdCcmKQ3J4PbbhV2HJdy"

// 	client := gocloak.NewClient("http://192.168.0.128:8080/", gocloak.SetAuthAdminRealms("admin/realms"), gocloak.SetAuthRealms("realms"))

// 	// ตรวจสอบ token และรับผลลัพธ์จากการตรวจสอบ

// 	result, err := client.RetrospectToken(context.Background(), token, clientID, clientSecret, realm)
// 	log.Printf("Token validation result: %v", result)
// 	if err != nil || !*result.Active {
// 		log.Printf("Token validation failed: %v", err)
// 		return http.StatusUnauthorized, []byte(`{"error": "Invalid token"}`), err
// 	}

// 	// สร้างข้อมูล JSON กลับเพื่อตอบกลับ
// 	respJSON := map[string]string{"message": "Token is valid"}
// 	response, err := json.Marshal(respJSON)
// 	if err != nil {
// 		return http.StatusInternalServerError, nil, err
// 	}

// 	return http.StatusOK, response, nil
// }

func verifyHandlerRedis(token string) (statusCode int, body []byte, err error) {
	tokenExists, err := checkTokenInRedis(token)
	if err != nil {
		log.Printf("Failed to check token in Redis: %v", err)
		return http.StatusInternalServerError, nil, err
	}
	if tokenExists {
		// Token found in Redis, return success response immediately
		respJSON := map[string]string{"message": "Token is valid"}
		response, err := json.Marshal(respJSON)
		if err != nil {
			return http.StatusInternalServerError, nil, err
		}
		return http.StatusOK, response, nil
	}

	// Token not found in Redis, return unauthorized error
	return http.StatusUnauthorized, []byte(`{"error": "Invalid token"}`), nil
}

func checkTokenInRedis(token string) (bool, error) {
	result, err := rdb.LRange("access_token", 0, -1).Result()
	if err != nil {
		return false, err
	}
	for _, t := range result {
		if t == token {
			return true, nil
		}
	}
	return false, nil
}

func callAuthServer(token string) (statusCode int, body []byte, err error) {
	statusCode, body, err = verifyHandlerRedis(token)
	if err != nil {
		return statusCode, body, err
	}

	return statusCode, body, nil
}
func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "192.168.0.128:6379",
	})
	server.StartServer(New, Version, Priority)
}
