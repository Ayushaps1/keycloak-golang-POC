package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/labstack/echo/v4"
)

func setupGoCloakClient() *gocloak.GoCloak {
	client := gocloak.NewClient("https://keycloak-url/")

	return client
}

func login(c echo.Context) error {
	client := setupGoCloakClient()

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

	return c.String(http.StatusOK, token.AccessToken)
}

func getUserData(c echo.Context) error {
	client := setupGoCloakClient()

	accessToken := c.Request().Header.Get("Authorization")

	userID := "admin-user-id" // Replace with the actual admin user ID

	user, err := client.GetUserByID(context.Background(), accessToken, "realm-name", userID)
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
	e.GET("/users/admin", getUserData)

	port := ":8888" // Replace with the desired port number
	log.Fatal(e.Start(port))
}
