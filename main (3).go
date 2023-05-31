package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"

)

var client = gocloak.NewClient("http://localhost:8080/")

func get_token() *gocloak.JWT {
	var token, err = client.LoginAdmin(context.Background(), "admin", "admin", "master")
	if err != nil {
		panic("something is wrong with loginning in")
	}
	return token
}

func createUser(c *gin.Context) {
	type values struct {
		User     gocloak.User `json:"user"`
		Password string       `json:"password"`
	}
	var token = get_token()
	var value values
	if err := c.BindJSON(&value); err != nil {
		return
	}
	Realm := c.Param("realm")
	res, err := client.CreateUser(context.Background(), token.AccessToken, Realm, value.User)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	if err := client.SetPassword(context.Background(), token.AccessToken, res, Realm, value.Password, false); err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}

	c.IndentedJSON(http.StatusCreated, res)
}

func adminLogin(c *gin.Context) {
	var loginCredentials struct {
		Username string `json:"username"`
		Password string `json:"password"`  
	}
	if err := c.BindJSON(&loginCredentials); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	token, err := client.LoginAdmin(context.Background(), loginCredentials.Username, loginCredentials.Password, "master")
	if err != nil {
		c.JSON(http.StatusUnauthorized, err)
		return
	}

	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    token.RefreshToken,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(c.Writer, &refreshCookie)

	expiration := time.Now().Add(24 * time.Hour)

	accessCookie := http.Cookie{
		Name:     "access_token",
		Value:    token.AccessToken,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
	}
	http.SetCookie(c.Writer, &accessCookie)

	// Return the access token in the response body
	c.JSON(http.StatusOK, gin.H{"access_token": token.AccessToken})
}

func loginUser(c *gin.Context) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var token = get_token()

	if err := c.BindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	Realm := c.Param("realm")
	token, err := client.Login(context.Background(), "demo-cli", "gHP1RRCz6aj0Gpuwt9TLMs7M2TcjzTy6", Realm, loginData.Username, loginData.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	cookie := http.Cookie{
		Name:     "refresh_token",
		Value:    token.RefreshToken,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(c.Writer, &cookie)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	})
}

func Logout_User(c *gin.Context) {
	cookie, err := c.Request.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not provided"})
		return
	}
	refreshToken := cookie.Value

	Realm := c.Param("realm")
	err = client.Logout(context.Background(), "demo-cli", "gHP1RRCz6aj0Gpuwt9TLMs7M2TcjzTy6", Realm, refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout user"})
		return
	}

	// Clear the refresh token cookie
	cookie = &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(c.Writer, cookie)

	// Return a success response
	c.JSON(http.StatusOK, gin.H{"message": "user logged out successfully"})
}

func adminLogout(c *gin.Context) {
	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(c.Writer, cookie)
	c.JSON(http.StatusOK, gin.H{"message": "admin logged out successfully"})
}

func createRealm(c *gin.Context) {
	accessCookie, err := c.Request.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}

	var Realm gocloak.RealmRepresentation
	if err := c.BindJSON(&Realm); err != nil {
		return
	}
	res, err := client.CreateRealm(context.Background(), accessCookie.Value, Realm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusCreated, res)

}

func createGroup(c *gin.Context) {
	accessCookie, err := c.Request.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}
	var group gocloak.Group
	if err := c.BindJSON(&group); err != nil {
		return
	}
	Realm := c.Param("realm")
	res, err := client.CreateGroup(context.Background(), accessCookie.Value, Realm, group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusCreated, res)
}

func createRealmRole(c *gin.Context) {
	accessCookie, err := c.Request.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}
	var newRole gocloak.Role
	if err := c.ShouldBindJSON(&newRole); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	Realm := c.Param("realm")
	roleID, err := client.CreateRealmRole(context.Background(), accessCookie.Value, Realm, newRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create realm role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"role_id": roleID})
}

func getUsers(c *gin.Context) {
	accessCookie, err := c.Request.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}

	Realm := c.Param("realm")

	numUsers := c.Query("numUsers")
	if numUsers == "" {
		numUsers = "1000"
	}
	num, err := strconv.Atoi(numUsers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userRoles parameter"})
	}

	users, err := client.GetUsers(context.Background(), accessCookie.Value, Realm, gocloak.GetUsersParams{Max: &num})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve users"})
		return
	}

	// Create a JSON array of users
	var userArray []map[string]interface{}
	for _, user := range users {
		userMap := map[string]interface{}{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
		}
		userArray = append(userArray, userMap)
	}

	// Return the JSON array in the response body
	c.JSON(http.StatusOK, userArray)
}

func getGroups(c *gin.Context) {
	accessCookie, err := c.Request.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}
	Realm := c.Param("realm")

	numGroups := c.Query("numGroups")
	if numGroups == "" {
		numGroups = "1000"
	}
	num, err := strconv.Atoi(numGroups)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid numGroups parameter"})
		return
	}

	groups, err := client.GetGroups(context.Background(), accessCookie.Value, Realm, gocloak.GetGroupsParams{Max: &num})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve groups"})
		return
	}
	var groupArray []map[string]interface{}
	for _, group := range groups {
		groupMap := map[string]interface{}{
			"id":         group.ID,
			"name":       group.Name,
			"path":       group.Path,
			"RealmRoles": group.RealmRoles,
		}
		groupArray = append(groupArray, groupMap)
	}

	// Return the JSON array in the response body
	c.JSON(http.StatusOK, groupArray[:])
}

func getRoles(c *gin.Context) {
	accessCookie, err := c.Request.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}
	Realm := c.Param("realm")

	numRoles := c.Query("numRoles")
	if numRoles == "" {
		numRoles = "1000"
	}
	num, err := strconv.Atoi(numRoles)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid numRoles parameter"})
		return
	}

	roles, err := client.GetRealmRoles(context.Background(), accessCookie.Value, Realm, gocloak.GetRoleParams{Max: &num})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve Roles"})
		return
	}

	// Create a JSON array of realm roles
	var roleArray []map[string]interface{}
	for _, role := range roles {
		roleMap := map[string]interface{}{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description,
		}
		roleArray = append(roleArray, roleMap)
	}

	// Return the JSON array in the response body
	c.JSON(http.StatusOK, roleArray[:])

}

func getUserByID(c *gin.Context) {
	// Fetch the access token from cookies
	accessCookie, err := c.Request.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}
	accessToken := accessCookie.Value

	Realm := c.Param("realm")

	userID := c.Query("userID")

	// Get user by ID
	userByID, err := client.GetUserByID(context.Background(), accessToken, Realm, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	// Create a JSON object of the user
	userMap := map[string]interface{}{
		"id":        userByID.ID,
		"username":  userByID.Username,
		"email":     userByID.Email,
		"firstName": userByID.FirstName,
		"lastName":  userByID.LastName,
	}

	// Return the JSON object in the response body
	c.JSON(http.StatusOK, userMap)
}

func getGroupByID(c *gin.Context) {
	// Fetch the access token from cookies
	accessCookie, err := c.Request.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}
	accessToken := accessCookie.Value

	Realm := c.Param("realm")

	groupID := c.Param("groupID")

	// Get group by ID
	groupByID, err := client.GetGroup(context.Background(), accessToken, Realm, groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve group by ID"})
		return
	}

	// Create a JSON object of the group
	groupMap := map[string]interface{}{
		"id":   groupByID.ID,
		"name": groupByID.Name,
	}

	// Return the JSON object in the response body
	c.JSON(http.StatusOK, groupMap)
}

func getRealmRoleByName(c *gin.Context) {
	// Fetch the access token from cookies
	accessCookie, err := c.Request.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}
	accessToken := accessCookie.Value

	Realm := c.Param("realm")

	roleName := c.Param("roleName")

	// Get realm role by name
	roleByName, err := client.GetRealmRole(context.Background(), accessToken, Realm, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve realm role by name"})
		return
	}

	// Create a JSON object of the realm role
	roleMap := map[string]interface{}{
		"id":   roleByName.ID,
		"name": roleByName.Name,
	}

	// Return the JSON object in the response body
	c.JSON(http.StatusOK, roleMap)
}

func deleteUser(c *gin.Context) {
	accessCookie, errr := c.Request.Cookie("access_token")
	if errr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}
	userID := c.Param("userID")

	Realm := c.Param("realm")
	err := client.DeleteUser(context.Background(), accessCookie.Value, Realm, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

func deleteGroup(c *gin.Context) {
	accessCookie, errr := c.Request.Cookie("access_token")
	if errr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}
	groupID := c.Param("groupID")
	Realm := c.Param("realm")

	err := client.DeleteGroup(context.Background(), accessCookie.Value, Realm, groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "group deleted successfully"})
}

func deleteRealmRole(c *gin.Context) {
	accessCookie, errr := c.Request.Cookie("access_token")
	if errr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access token not found"})
		return
	}
	roleName := c.Param("roleName")
	Realm := c.Param("realm")

	err := client.DeleteRealmRole(context.Background(), accessCookie.Value, Realm, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete realm"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "realmrole deleted successfully"})
}

func main() {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowCredentials = true
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowHeaders = []string{"Content-Type"}
	router.Use(cors.New(config))

	// Api to Register User
	router.POST("/create/user/:realm", createUser)

	// API to Login admin
	router.POST("/login/admin/", adminLogin)

	// API to login user
	router.POST("/login/user/:realm", loginUser)

	// API to Create Realm
	router.POST("/create/realm", createRealm)

	// API to Logout User
	router.POST("/logout/user/", Logout_User)

	// API to logout admin
	router.POST("/logout/admin/:realm", adminLogout)

	// API to Create Group
	router.POST("/create/group/:realm", createGroup)

	//API to create a realm role
	router.POST("/create/role/:realm", createRealmRole)

	//API
	router.GET("/get/users/:realm", getUsers)

	router.GET("/get/groups/:realm", getGroups)

	router.GET("/get/roles/:realm", getRoles)

	// API to get users
	router.GET("/get/user/:realm", getUserByID)

	// API to get froups
	router.GET("/get/group/:realm", getGroupByID)

	// API to get realms
	router.GET("/get/role/:realm", getRealmRoleByName)

	// API to delete user
	router.DELETE("/delete/user/:realm/:userID/", deleteUser)

	router.DELETE("/delete/group/:realm/:groupID/", deleteGroup)

	router.DELETE("/delete/role/:realm/:roleName/", deleteRealmRole)

	router.Run("localhost:8888")
}
