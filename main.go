package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xyproto/permissionbolt"
)

func main() {
	// Set Gin to production mode
	//gin.SetMode(gin.ReleaseMode)

	// Set the router as the default one provided by Gin
	//router = gin.Default()

	g := gin.New()

	g.LoadHTMLGlob("templates/*.html")

	perm, err := permissionbolt.New()
	if err != nil {
		log.Fatalln(err)
	}

	// Blank slate, no default permissions
	//perm.Clear()

	// Set up a middleware handler for Gin, with a custom "permission denied" message.
	permissionHandler := func(c *gin.Context) {
		// Check if the user has the right admin/user rights
		if perm.Rejected(c.Writer, c.Request) {
			// Deny the request, don't call other middleware handlers
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
			return
		}
		// Call the next middleware handler
		c.Next()
	}

	// Logging middleware
	g.Use(gin.Logger())

	// Enable the permissionbolt middleware, must come before recovery
	g.Use(permissionHandler)

	// Recovery middleware
	g.Use(gin.Recovery())

	// Get the userstate, used in the handlers below
	userstate := perm.UserState()

	g.GET("/", func(c *gin.Context) {
		//msg := ""
		//msg += fmt.Sprintf("Has user bob: %v\n", userstate.HasUser("bob"))
		//msg += fmt.Sprintf("Logged in on server: %v\n", userstate.IsLoggedIn("bob"))
		//msg += fmt.Sprintf("Is confirmed: %v\n", userstate.IsConfirmed("bob"))
		//msg += fmt.Sprintf("Username stored in cookies (or blank): %v\n", userstate.Username(c.Request))
		//msg += fmt.Sprintf("Current user is logged in, has a valid cookie and *user rights*: %v\n", userstate.UserRights(c.Request))
		//msg += fmt.Sprintf("Current user is logged in, has a valid cookie and *admin rights*: %v\n", userstate.AdminRights(c.Request))
		//msg += fmt.Sprintln("\nTry: /register, /confirm, /remove, /login, /logout, /makeadmin, /clear, /data and /admin")
		//c.String(http.StatusOK, msg)
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	g.GET("/register", func(c *gin.Context) {
		//userstate.AddUser("bob", "hunter1", "bob@zombo.com")
		//c.String(http.StatusOK, fmt.Sprintf("User bob was created: %v\n", userstate.HasUser("bob")))

		c.HTML(http.StatusOK, "register.html", gin.H{})
		c.String(http.StatusOK, fmt.Sprintf("User lord was created: %v\n", userstate.HasUser("lord")))
	})

	g.POST("/register", func(c *gin.Context) {
		username := c.PostForm("username")
		pass := c.PostForm("password")
		message := c.PostForm("email")

		userstate.AddUser(username, pass, message)

		c.HTML(http.StatusOK, "register.html", gin.H{})
		c.String(http.StatusOK, fmt.Sprintf("User lord was created: %v\n", userstate.HasUser("lord")))
	})

	g.GET("/confirm", func(c *gin.Context) {
		userstate.MarkConfirmed("bob")
		c.String(http.StatusOK, fmt.Sprintf("User bob was confirmed: %v\n", userstate.IsConfirmed("bob")))
	})

	//g.GET("/remove", func(c *gin.Context) {
	//	userstate.RemoveUser("bob")
	//	userstate.FindUserByConfirmationCode("bob")
	//	c.String(http.StatusOK, fmt.Sprintf("User bob was removed: %v\n", !userstate.HasUser("bob")))
	//})

	g.GET("/listusers", func(c *gin.Context) {
		listusers, _ := userstate.AllUsernames()
		c.HTML(http.StatusOK, "listusers.html", gin.H{"userlist": listusers})
	})

	g.GET("/login", func(c *gin.Context) {
		//test
		userstate.Login(c.Writer, "bob")
		// Headers will be written, for storing a cookie
		//userstate.Login(c.Writer, "bob")
		//c.String(http.StatusOK, fmt.Sprintf("bob is now logged in: %v\n", userstate.IsLoggedIn("bob")))
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})

	g.POST("/login", func(c *gin.Context) {

		username := c.PostForm("username")
		userstate.Login(c.Writer, username)
		// password := c.PostForm("password")
		// if username, err := userstate.FindUserByConfirmationCode(username); err == nil {
		// 	// if password := userstate.CorrectPassword(username, password){
		// 	userstate.Login(c.Writer, username)
		// 	// }
		// 	// c.String(http.StatusOK, "list of all users: "+strings.Join(usernames, ", "))
		// }
		c.HTML(http.StatusOK, "login-successful.html", gin.H{})
		// c.String(http.StatusOK, fmt.Sprintf(username+" is now logged in: %v\n", userstate.IsLoggedIn(username)))
	})

	g.GET("/logout", func(c *gin.Context) {
		// userstate.Logout("bob")
		c.String(http.StatusOK, fmt.Sprintf("bob is now logged out: %v\n", !userstate.IsLoggedIn("bob")))
	})

	g.POST("/logout", func(c *gin.Context) {
		username := c.PostForm("username")
		userstate.Logout(username)
		// userstate.Logout("bob")
		c.String(http.StatusOK, fmt.Sprintf("bob is now logged out: %v\n", !userstate.IsLoggedIn("bob")))
	})

	g.GET("/makeadmin", func(c *gin.Context) {
		userstate.SetAdminStatus("bob")
		c.String(http.StatusOK, fmt.Sprintf("bob is now administrator: %v\n", userstate.IsAdmin("bob")))
	})

	g.GET("/clear", func(c *gin.Context) {
		userstate.ClearCookie(c.Writer)
		c.String(http.StatusOK, "Clearing cookie")
	})

	g.GET("/data", func(c *gin.Context) {
		c.String(http.StatusOK, "user page that only logged in users must see!")
	})

	g.GET("/delete", func(c *gin.Context) {
		c.HTML(http.StatusOK, "delete.html", gin.H{})
	})

	g.POST("/delete", func(c *gin.Context) {
		username := c.PostForm("username")
		userstate.RemoveUser(username)
		c.HTML(http.StatusOK, "delete.html", gin.H{})
	})

	g.GET("/admin", func(c *gin.Context) {
		c.String(http.StatusOK, "super secret information that only logged in administrators must see!\n\n")
		if usernames, err := userstate.AllUsernames(); err == nil {
			c.String(http.StatusOK, "list of all users: "+strings.Join(usernames, ", "))
		}
	})
	// Start serving the application
	g.Run(":3000")
}
