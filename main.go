package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xyproto/permissionbolt"
)

func main() {

	//ADD EXAMPLE BOLTDB
	// Set Gin to production mode
	//gin.SetMode(gin.ReleaseMode)

	// Set the router as the default one provided by Gin
	//router = gin.Default()

	Open()
	defer Close()

	// A Person struct consists of ID, Name, Age, Job.
	/* peeps := []*Person{
		{"100", "Bill Joy", "60", "Programmer"},
		{"101", "Peter Norvig", "58", "Programmer"},
		{"102", "Donald Knuth", "77", "Programmer"},
		{"103", "Jeff Dean", "47", "Programmer"},
		{"104", "Rob Pike", "59", "Programmer"},
		{"200", "Brian Kernighan", "73", "Programmer"},
		{"201", "Ken Thompson", "72", "Programmer"},
	}
	// Persist people in the database.
	for _, p := range peeps {
		p.save()
	}
	// Get a person from the database by their ID.
	for _, id := range []string{"100", "101"} {
		p, err := GetPerson(id)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(p)
	} */

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

	isloggedin := func(c *gin.Context) bool {
		usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := userstate.IsLoggedIn(usercook)
		return isloggedin
	}

	g.GET("/", func(c *gin.Context) {
		isloggedin := isloggedin(c)
		if isloggedin {
			c.HTML(http.StatusOK, "index.html", gin.H{"is_logged_in": isloggedin})
		} else {
			c.Redirect(307, "/login")
		}
	})

	g.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{})
	})

	g.POST("/register", func(c *gin.Context) {

		username := c.PostForm("username")
		pass := c.PostForm("password")
		message := c.PostForm("email")

		userstate.AddUser(username, pass, message)
		userstate.Login(c.Writer, username)
		userstate.MarkConfirmed(username)

		http.Redirect(c.Writer, c.Request, "/", 302)
	})

	g.GET("/login", func(c *gin.Context) {
		isloggedin := isloggedin(c)
		c.HTML(http.StatusOK, "login.html", gin.H{"title": "Login Page",
			"is_logged_in": isloggedin})
	})

	g.POST("/login", func(c *gin.Context) {

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

	g.GET("/logout", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		userstate.Logout(usercook)
		http.Redirect(c.Writer, c.Request, "/", 302)
	})

	g.GET("/listusers", func(c *gin.Context) {
		isloggedin := isloggedin(c)

		if isloggedin {
			listusers, _ := userstate.AllUsernames()
			ListPrefix("people", "20") // ... with key prefix `20`
			c.HTML(http.StatusOK, "listusers.html", gin.H{"userlist": listusers, "is_logged_in": isloggedin})
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}

	})

	//List register users
	g.GET("/base", func(c *gin.Context) {
		isloggedin := isloggedin(c)

		z := []string{}
		if isloggedin {
			for _, id := range []string{"100", "101", "103", "104", "105", "106"} {
				p, err := GetPerson(id)
				list := p.Name
				z = append(z, list)
				if err != nil {
					log.Fatal(err)
				}
			}

			// List("people")

			listx, _ := ListX("people")
			fmt.Println(listx)

			c.HTML(http.StatusOK, "base.html", gin.H{"z": z, "is_logged_in": isloggedin})
			// c.HTML(http.StatusOK, "base.html", gin.H{"listx": listx, "is_logged_in": isloggedin})

			// list := List("people") // each key/val in people bucket
			// ListPrefix("people", "20") // ... with key prefix `20`
			// ListRange("people", "101", "103") // ... within range `101` to `103`

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//Register visitors GET
	g.GET("/visitors", func(c *gin.Context) {
		isloggedin := isloggedin(c)

		if isloggedin {
			c.HTML(http.StatusOK, "visitors.html", gin.H{"is_logged_in": isloggedin})
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//Register visitors POST
	g.POST("/visitors", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := userstate.IsLoggedIn(usercook)

		if isloggedin {
			id := c.PostForm("ID")
			name := c.PostForm("Name")
			age := c.PostForm("Age")
			job := c.PostForm("Job")

			peeps := []*Person{
				{id, usercook, name, age, job},
			}

			for _, p := range peeps {
				// fmt.Println(p)
				p.create()
			}

			// List("people")
			http.Redirect(c.Writer, c.Request, "/visitors", 302)

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//Make user as admin GET
	g.GET("/makeadmin", func(c *gin.Context) {
		isloggedin := isloggedin(c)
		c.HTML(http.StatusOK, "makeadmin.html", gin.H{"is_logged_in": isloggedin})
	})

	//Make user as admin POST
	g.POST("/makeadmin", func(c *gin.Context) {
		username := c.PostForm("username")
		userstate.SetAdminStatus(username)
		c.HTML(http.StatusOK, "makeadmin.html", gin.H{})
	})

	/* 	g.GET("/clear", func(c *gin.Context) {
		userstate.ClearCookie(c.Writer)
		c.String(http.StatusOK, "Clearing cookie")
	}) */

	//Delete User from Base GET
	g.GET("/delete", func(c *gin.Context) {
		isloggedin := isloggedin(c)
		if isloggedin {
			c.HTML(http.StatusOK, "delete.html", gin.H{"is_logged_in": isloggedin})
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//Delete User from Base POST
	g.POST("/delete", func(c *gin.Context) {
		username := c.PostForm("username")
		userstate.RemoveUser(username)
		c.HTML(http.StatusOK, "delete.html", gin.H{})
	})

	/* 	g.GET("/admin", func(c *gin.Context) {
		c.String(http.StatusOK, "super secret information that only logged in administrators must see!\n\n")
		if usernames, err := userstate.AllUsernames(); err == nil {
			c.String(http.StatusOK, "list of all users: "+strings.Join(usernames, ", "))
		}
	}) */

	// Start serving the application
	g.Run(":3000")
}
