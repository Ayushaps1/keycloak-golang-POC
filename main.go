package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/labstack/echo/v4"
)

var client = gocloak.NewClient("http://keycloak_svr:8080/")

func login(c echo.Context) error {
	var loginCredentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.Bind(&loginCredentials); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	token, err := client.LoginAdmin(context.Background(), loginCredentials.Username, loginCredentials.Password, "master")
	if err != nil {
		log.Println("Failed to login as admin:", err)
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	response := map[string]interface{}{
		"message":      "Admin login successful",
		"access_token": token.AccessToken,
	}

	return c.JSON(http.StatusOK, response)
}

func getUserData(c echo.Context) error {
	userID := c.Param("userid")
	realm := c.Param("realm")

	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	// Extract the JWT token from the header
	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := client.GetUserByID(context.Background(), token, realm, userID)
	if err != nil {
		log.Println("Failed to get user data:", err)
		return c.String(http.StatusInternalServerError, "Failed to get user data")
	}

	return c.JSON(http.StatusOK, user)
}

func main() {
	e := echo.New()

	// Existing routes
	e.POST("/login", login)

	// New route
	e.GET("/users/:userid/:realm", getUserData)

	port := ":8888" // Replace with the desired port number
	log.Fatal(e.Start(port))
}
