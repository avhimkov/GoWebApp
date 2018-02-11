package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func initializeRoutes() {
	g := gin.New()

	// Use the setUserStatus middleware for every route to set a flag
	// indicating whether the request was from an authenticated user or not

	// Handle the index route
	g.GET("/", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := userstate.IsLoggedIn(usercook)

		if isloggedin {
			c.HTML(http.StatusOK, "index.html", gin.H{"is_logged_in": isloggedin})
		} else {
			c.Redirect(307, "/login")
		}
	})
	// Group user related routes together
	userRoutes := g.Group("/u")
	{
		// Handle the GET requests at /u/login
		// Show the login page
		// Ensure that the user is not logged in by using the middleware
		userRoutes.GET("/login", ensureNotLoggedIn(), showLoginPage)

		// Handle POST requests at /u/login
		// Ensure that the user is not logged in by using the middleware
		userRoutes.POST("/login", func(c *gin.Context) {

			username := c.PostForm("username")
			password := c.PostForm("password")
			logintryst := userstate.CorrectPassword(username, password)

			if logintryst == true {
				userstate.Login(c.Writer, username)
				// c.HTML(http.StatusOK, "index.html", gin.H{"title": "Successful Login"})
				http.Redirect(c.Writer, c.Request, "/", 302)
			} else {

				// c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
				c.HTML(http.StatusBadRequest, "login.html", gin.H{
					"ErrorTitle":   "Login Failed",
					"ErrorMessage": "Invalid credentials provided"})
			}
		})

		// Handle GET requests at /u/logout
		// Ensure that the user is logged in by using the middleware
		userRoutes.GET("/logout", func(c *gin.Context) {
			usercook, _ := userstate.UsernameCookie(c.Request)
			isloggedin := userstate.IsLoggedIn(usercook)
			c.HTML(http.StatusOK, "login.html", gin.H{"title": "Login Page",
				"is_logged_in": isloggedin})
		})

		// Handle the GET requests at /u/register
		// Show the registration page
		// Ensure that the user is not logged in by using the middleware
		userRoutes.GET("/register", func(c *gin.Context) {
			c.HTML(http.StatusOK, "register.html", gin.H{})
		})

		// Handle POST requests at /u/register
		// Ensure that the user is not logged in by using the middleware
		userRoutes.POST("/register", func(c *gin.Context) {

			username := c.PostForm("username")
			pass := c.PostForm("password")
			message := c.PostForm("email")

			userstate.AddUser(username, pass, message)
			userstate.Login(c.Writer, username)
			userstate.MarkConfirmed(username)

			http.Redirect(c.Writer, c.Request, "/", 302)
		})
	}

}
