package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// func indexHandler(c *gin.Context) {
//     c.HTML(http.StatusOK, "index.html", gin.H{})
// }

func showLoginPage(c *gin.Context) {
	// Call the render function with the name of the template to render
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

// func performLogin(username string, password string, c *gin.Context) (string, bool) {
// 	username := c.PostForm("username")
// 	// userstate.Login(c.Writer, username)
// 	password := c.PostForm("password")
// 	logintryst := userstate.CorrectPassword(username, password)
// 	// if username, err := userstate.FindUserByConfirmationCode(username); err == nil {
// 	// 	// if password := userstate.CorrectPassword(username, password){
// 	// 	userstate.Login(c.Writer, username)
// 	// 	// }
// 	// 	// c.String(http.StatusOK, "list of all users: "+strings.Join(usernames, ", "))
// 	// }
// 	if logintryst == true {

// 		// if u == username && u.Password == password {
// 		c.HTML(http.StatusOK, "login-successful.html", gin.H{})
// 	} else {
// 		c.HTML(http.StatusOK, "index.html", gin.H{})
// 	}

// 	return "", false
// }

// func logout(c *gin.Context) {
// 	// Clear the cookie
// 	c.SetCookie("token", "", -1, "", "", false, true)

// 	// Redirect to the home page
// 	c.Redirect(http.StatusTemporaryRedirect, "/")
// }

// func showRegistrationPage(c *gin.Context) {
// 	// Call the render function with the name of the template to render
// 	render(c, gin.H{
// 		"title": "Register"}, "register.html")
// }

// func register(c *gin.Context) {
// 	// Obtain the POSTed username and password values
// 	username := c.PostForm("username")
// 	password := c.PostForm("password")

// 	if _, err := registerNewUser(username, password); err == nil {
// 		// If the user is created, set the token in a cookie and log the user in
// 		token := generateSessionToken()
// 		c.SetCookie("token", token, 3600, "", "", false, true)
// 		c.Set("is_logged_in", true)

// 		render(c, gin.H{
// 			"title": "Successful registration & Login"}, "login-successful.html")

// 	} else {
// 		// If the username/password combination is invalid,
// 		// show the error message on the login page
// 		c.HTML(http.StatusBadRequest, "register.html", gin.H{
// 			"ErrorTitle":   "Registration Failed",
// 			"ErrorMessage": err.Error()})

// 	}
// }
