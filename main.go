package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/asdine/storm"
	"github.com/gin-gonic/gin"
	"github.com/xyproto/permissionbolt"
)

//Struc data visitors
type Person struct {
	ID          int `storm:"id,increment"` //`form:"ID" storm:"id,increment" json:"ID"`
	User        string
	Name        string `storm:"index"` //Заявитель
	SubName     string `storm:"index"` //Представитель заявитель
	NameService string `storm:"index"` //Услуга
	Date        string `storm:"index"` //Дата
	Address     string `storm:"index"` //Адрес
	Location    string `storm:"index"` //Место оператора
	Number      string `storm:"index"` //
	Phone       string `storm:"index"` //Телефон
	Note        string `storm:"index"` //Примечание
}

var perm, _ = permissionbolt.New()

//open database
func DB() *storm.DB {
	db, err := storm.Open("db/data.db")
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()
	return db
}

func SetupRouter() *gin.Engine {

	//ADD EXAMPLE BOLTDB
	// Set Gin to production mode
	//gin.SetMode(gin.ReleaseMode)

	// Set the router as the default one provided by Gin
	//router = gin.Default()

	g := gin.Default()
	g.Static("/assets", "./assets")
	g.LoadHTMLGlob("templates/*.html")

	// g := gin.New()

	// Logging middleware
	g.Use(gin.Logger())
	// Recovery middleware
	g.Use(gin.Recovery())

	// g.Use(static.Serve("/assets", static.LocalFile("/assets", false)))
	// v1 := router.Group("api/v1")
	// {
	// 	v1.GET("/instructions", GetInstructions)
	// }

	return g
}

func main() {

	db := DB()
	defer db.Close()

	// db.Update(func(tx *bolt.Tx) error {
	// 	b := tx.Bucket([]byte("bucketName"))
	// 	err := b.Delete([]byte("keyToDelete"))
	// 	return err
	// })

	g := SetupRouter()

	// perm, err := permissionbolt.New()
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	//	Blank slate, no default permissions
	//	perm.Clear()

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

	// Enable the permissionbolt middleware, must come before recovery
	g.Use(permissionHandler)

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
			c.HTML(http.StatusOK, "operator.html", gin.H{"is_logged_in": isloggedin})
		} else {
			http.Redirect(c.Writer, c.Request, "/login", 302)
		}
	})

	// Registaration Users GET
	g.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{})
	})

	// Registaration Users OPOST
	g.POST("/register", func(c *gin.Context) {

		username := c.PostForm("username")
		pass := c.PostForm("password")
		message := c.PostForm("email")

		userstate.AddUser(username, pass, message)
		userstate.Login(c.Writer, username)
		userstate.MarkConfirmed(username)

		http.Redirect(c.Writer, c.Request, "/", 302)
	})

	// Loging Users GET
	g.GET("/login", func(c *gin.Context) {

		isloggedin := isloggedin(c)
		c.HTML(http.StatusOK, "login.html", gin.H{"title": "Login Page", "is_logged_in": isloggedin})
	})

	// Loging Users
	g.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		logintryst := userstate.CorrectPassword(username, password)

		if logintryst == true {
			userstate.Login(c.Writer, username)
			// c.HTML(http.StatusOK, "index.html", gin.H{"title": "Successful Login"})
			http.Redirect(c.Writer, c.Request, "/operator", 302)
		} else {

			// c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"ErrorTitle":   "Login Failed",
				"ErrorMessage": "Invalid credentials provided"})
		}
	})

	//Logout users
	g.GET("/logout", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		userstate.Logout(usercook)
		http.Redirect(c.Writer, c.Request, "/login", 302)
	})

	//Make user as admin POST
	g.GET("/makeadmin/:user", func(c *gin.Context) {
		user := c.Param("user")
		// username := c.PostForm(user)
		userstate.SetAdminStatus(user)
		http.Redirect(c.Writer, c.Request, "/adminka", 302)
	})

	//Delete User from Base POST
	g.GET("/delete/:user", func(c *gin.Context) {
		user := c.Param("user")
		// username := c.PostForm(user)
		userstate.RemoveUser(user)
		http.Redirect(c.Writer, c.Request, "/adminka", 302)
	})

	//Delete Admin status
	g.GET("/adminoff/:user", func(c *gin.Context) {
		user := c.Param("user")
		userstate.IsAdmin(user)
		userstate.RemoveAdminStatus(user)
		http.Redirect(c.Writer, c.Request, "/adminka", 302)
	})

	//Administartort interface
	g.GET("/adminka", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := userstate.IsLoggedIn(usercook)
		isadmin := userstate.IsAdmin(usercook)

		var cheked []bool
		if isloggedin {
			listusers, _ := userstate.AllUsernames()
			fmt.Println(isadmin)
			if isadmin {
				for _, i := range listusers {
					fmt.Println(i)
					cheked = append(cheked, userstate.IsAdmin(i))
				}
				fmt.Println(cheked)
				fmt.Println(isadmin)
			}
			c.HTML(http.StatusOK, "adminka.html", gin.H{"listusers": listusers, "is_logged_in": isloggedin})
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//operator register users
	g.GET("/operator", func(c *gin.Context) {
		isloggedin := isloggedin(c)

		if isloggedin {

			var person []Person
			err := db.All(&person)
			if err != nil {
				log.Fatal(err)
			}

			c.HTML(http.StatusOK, "operator.html", gin.H{"person": person, "is_logged_in": isloggedin})

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//Register visitors POST
	g.POST("/operator", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := userstate.IsLoggedIn(usercook)

		if isloggedin {

			name := c.PostForm("name")
			nameservice := c.PostForm("nameservice")
			date := c.PostForm("date")
			number := c.PostForm("number")

			peeps := []*Person{
				{User: usercook, Name: name, NameService: nameservice, Date: date, Number: number},
			}

			for _, p := range peeps {
				fmt.Println(p)
				db.Save(p)
			}

			http.Redirect(c.Writer, c.Request, "/operator", 302)
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//operator register users
	g.GET("/kontroler", func(c *gin.Context) {
		isloggedin := isloggedin(c)

		if isloggedin {

			var person []Person
			err := db.All(&person)
			if err != nil {
				log.Fatal(err)
			}

			c.HTML(http.StatusOK, "kontroler.html", gin.H{"person": person, "is_logged_in": isloggedin})

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//Register visitors POST
	g.POST("/kontroler", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)

		// id := c.PostForm("id")
		name := c.PostForm("name")
		nameservice := c.PostForm("nameservice")
		date := c.PostForm("date")
		number := c.PostForm("number")

		peeps := []*Person{
			{User: usercook, Name: name, NameService: nameservice, Date: date, Number: number},
		}

		for _, p := range peeps {
			db.Save(p)
		}

		http.Redirect(c.Writer, c.Request, "/kontroler", 302)
	})

	//konsult register users
	g.GET("/konsult", func(c *gin.Context) {
		isloggedin := isloggedin(c)

		if isloggedin {

			var person []Person
			err := db.All(&person)
			if err != nil {
				log.Fatal(err)
			}

			c.HTML(http.StatusOK, "konsult.html", gin.H{"person": person, "is_logged_in": isloggedin})

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	//konsult visitors POST
	g.POST("/konsult", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)

		// id := c.PostForm("id")
		name := c.PostForm("name")
		nameservice := c.PostForm("nameservice")
		date := c.PostForm("date")
		number := c.PostForm("number")

		peeps := []*Person{
			{User: usercook, Name: name, NameService: nameservice, Date: date, Number: number},
		}

		for _, p := range peeps {
			db.Save(p)
		}

		http.Redirect(c.Writer, c.Request, "/konsult", 302)
	})

	//Delete value on id
	g.GET("/removeval/:id", Remove)

	//Edit data
	g.GET("/edit/:uid", EditValue)

	//Edit
	// g.GET("/edit", func(c *gin.Context) {
	// 	usercook, _ := userstate.UsernameCookie(c.Request)
	// 	isloggedin := userstate.IsLoggedIn(usercook)

	// 	if isloggedin {

	// 		name := c.PostForm("name")
	// 		nameservice := c.PostForm("nameservice")
	// 		date := c.PostForm("date")
	// 		number := c.PostForm("number")

	// 		peeps := []*Person{
	// 			{User: usercook, Name: name, NameService: nameservice, Date: date, Number: number},
	// 		}

	// 		for _, p := range peeps {
	// 			// fmt.Println(p)
	// 			db.Update(p)
	// 		}
	// 		c.HTML(http.StatusOK, "editTable.html", gin.H{"peeps": peeps, "is_logged_in": isloggedin})
	// 		// http.Redirect(c.Writer, c.Request, "/edit", 302)
	// 	} else {
	// 		c.AbortWithStatus(http.StatusForbidden)
	// 		fmt.Fprint(c.Writer, "Permission denied!")
	// 	}
	// })

	// Start serving the application
	g.Run(":3000")
}

func EditValue(c *gin.Context) {
	// db := DB()

	uid := c.Param("uid")
	fmt.Println(uid)
	qid := c.Query(uid)
	fmt.Println(qid)

	// var person []Person
	// err := db.Find(uid, qid, &person)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("Found", len(person))

	http.Redirect(c.Writer, c.Request, "/edit", 302)
}

func Remove(c *gin.Context) {
	id := c.Param("id")
	fmt.Println(id)
	// username := c.PostForm(user)
	// userstate.RemoveUser(user)
	http.Redirect(c.Writer, c.Request, "/operator", 302)
}

// func Logger() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		t := time.Now()
// 		// Set example variable
// 		c.Set("example", "12345")
// 		// before request
// 		c.Next()
// 		// after request
// 		latency := time.Since(t)
// 		log.Print(latency)
// 		// access the status we are sending
// 		status := c.Writer.Status()
// 		log.Println(status)
// 	}
// }
