package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/gin-gonic/gin"
	"github.com/xyproto/permissionbolt"
)

//Struc data visitors
type Person struct {
	ID          int `storm:"id,increment"` //`form:"ID" storm:"id,increment" json:"ID"`
	User        string
	Name        string `storm:"index" json:"name" form:"name" binding:"required"`               //Заявитель
	SubName     string `storm:"index" json:"subname" form:"subname" binding:"required"`         //Представитель заявитель
	NameService string `storm:"index" json:"nameservice" form:"nameservice" binding:"required"` //Услуга
	Date        string `storm:"index" json:"date" form:"date" binding:"required"`               //Дата
	// Address   *Address `json:"address,omitempty"`
	Address  string `storm:"index" json:"address" form:"address" binding:"required"`   //Адрес
	Location string `storm:"index" json:"location" form:"location" binding:"required"` //Место оператора
	Number   string `storm:"index" json:"number" form:"number" binding:"required"`     //
	Phone    string `storm:"index" json:"phone" form:"phone" binding:"required"`       //Телефон
	Note     string `storm:"index" json:"note" form:"note" binding:"required"`         //Примечание
}

// type Address struct {
//     City  string `json:"city"`
//     State string `json:"state"`
// }

var perm, _ = permissionbolt.New()

// perm, err := permissionbolt.New()
// if err != nil {
// 	log.Fatalln(err)
// }

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

	// Set Gin to production mode
	//gin.SetMode(gin.ReleaseMode)

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

// func uploadValue(c *gin.Context) {
// 	uid := c.Request.FormValue("uid")
// 	file, header, err := c.Request.FormFile("uploadFile")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	filename := header.Filename
// 	fmt.Println(filename)
// 	err = os.Mkdir("./upload/"+uid, 777)
// 	out, err := os.Create("./upload/" + uid + "/" + filename)
// 	_, err = io.Copy(out, file)
// }

func main() {

	db := DB()
	defer db.Close()

	g := SetupRouter()

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
		mail := c.PostForm("email")

		userstate.AddUser(username, pass, mail)
		userstate.Login(c.Writer, username)
		userstate.MarkConfirmed(username)

		http.Redirect(c.Writer, c.Request, "/", 302)
	})

	g.POST("/uploadUsers", func(c *gin.Context) {

		type Users struct {
			User    string
			Address string
			Pass    string
		}

		csvFile, _ := os.Open("people.csv")
		reader := csv.NewReader(bufio.NewReader(csvFile))
		var people []Users
		for {
			line, error := reader.Read()
			if error == io.EOF {
				break
			} else if error != nil {
				log.Fatal(error)
			}
			people = append(people, Users{
				User:    line[0],
				Pass:    line[1],
				Address: line[2],
			})
		}

		for _, p := range people {
			userstate.AddUser(p.User, p.Pass, p.Address)
			userstate.Login(c.Writer, p.User)
			userstate.MarkConfirmed(p.User)
		}

		peopleJson, _ := json.Marshal(people)
		fmt.Println(string(peopleJson))

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
			http.Redirect(c.Writer, c.Request, "/operator", 302)
		} else {
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
		userstate.SetAdminStatus(user)
		http.Redirect(c.Writer, c.Request, "/adminka", 302)
	})

	//Delete User from Base POST
	g.GET("/delete/:user", func(c *gin.Context) {
		user := c.Param("user")
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
			fmt.Println(person)

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
			subname := c.PostForm("subname")
			nameservice := c.PostForm("nameservice")
			date := c.PostForm("date")
			address := c.PostForm("address")
			loc := c.PostForm("loc")
			number := c.PostForm("number")
			phone := c.PostForm("phone")
			note := c.PostForm("note")

			peeps := []*Person{
				{User: usercook, Name: name, SubName: subname, NameService: nameservice,
					Date: date, Address: address, Location: loc, Number: number, Phone: phone, Note: note},
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

	//API Maxima
	g.GET("/maxima", func(c *gin.Context) {
		isloggedin := isloggedin(c)

		if isloggedin {

			var person []Person
			err := db.All(&person)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(person)

			c.HTML(http.StatusOK, "operator.html", gin.H{"person": person, "is_logged_in": isloggedin})

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	// g.POST("/uploadValue", uploadValue)

	g.POST("/uploadValue", func(c *gin.Context) {
		// uid, _ := userstate.GenerateUniqueConfirmationCode()
		usercook, _ := userstate.UsernameCookie(c.Request)

		path := filepath.Clean("./upload/")

		file, header, err := c.Request.FormFile("uploadFile")
		if err != nil {
			log.Fatal(err)
		}
		filename := header.Filename
		fmt.Println(filename)
		err = os.MkdirAll(path, 0777)
		if err != nil {
			log.Fatal(err)
		}
		out, err := os.Create(path + "/" + filename)
		_, err = io.Copy(out, file)

		url := path + "/" + filename
		fmt.Println(url)

		//Work
		csvFile, _ := os.Open(url)
		reader := csv.NewReader(bufio.NewReader(csvFile))
		for {
			line, error := reader.Read()
			if error == io.EOF {
				break
			} else if error != nil {
				log.Fatal(error)
			}

			peeps := []*Person{
				{User: usercook,
					Name:        line[0],
					SubName:     line[1],
					NameService: line[2],
					Date:        line[3],
					Address:     line[4],
					Location:    line[5],
					Number:      line[6],
					Phone:       line[7],
					Note:        line[8]},
			}

			for _, p := range peeps {
				db.Save(p)
				fmt.Println(p)
			}
		}

		http.Redirect(c.Writer, c.Request, "/operator", 302)
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
		subname := c.PostForm("subname")
		nameservice := c.PostForm("nameservice")
		date := c.PostForm("date")
		address := c.PostForm("address")
		loc := c.PostForm("loc")
		number := c.PostForm("number")
		phone := c.PostForm("phone")
		note := c.PostForm("note")

		peeps := []*Person{
			{User: usercook, Name: name, SubName: subname, NameService: nameservice,
				Date: date, Address: address, Location: loc, Number: number, Phone: phone, Note: note},
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
		subname := c.PostForm("subname")
		nameservice := c.PostForm("nameservice")
		date := c.PostForm("date")
		address := c.PostForm("address")
		loc := c.PostForm("loc")
		number := c.PostForm("number")
		phone := c.PostForm("phone")
		note := c.PostForm("note")

		peeps := []*Person{
			{User: usercook, Name: name, SubName: subname, NameService: nameservice,
				Date: date, Address: address, Location: loc, Number: number, Phone: phone, Note: note},
		}

		for _, p := range peeps {
			db.Save(p)
		}

		http.Redirect(c.Writer, c.Request, "/konsult", 302)
	})

	//Delete value on id
	g.GET("/removeval/:id", func(c *gin.Context) {
		id := c.Param("id")

		query := db.Select(q.Eq("ID", id))
		query.Delete(new(Person))

		fmt.Println(id)
		http.Redirect(c.Writer, c.Request, "/operator", 302)
	})

	g.GET("/edit/:id", func(c *gin.Context) {
		id := c.Param("id")
		isloggedin := isloggedin(c)

		if isloggedin {
			var person Person
			findVal := db.Select(q.Eq("ID", id))
			err := findVal.First(&person)
			if err != nil {
				log.Fatal(err)
			}
			c.Bind(&person)
			fmt.Println(person)

			c.HTML(http.StatusOK, "edittable.html", gin.H{"person": person, "is_logged_in": isloggedin})
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	g.POST("/edit/:id", func(c *gin.Context) {
		isloggedin := isloggedin(c)
		id := c.Param("id")
		fmt.Println(id)

		if isloggedin {
			name := c.PostForm("name")
			subname := c.PostForm("subname")
			nameservice := c.PostForm("nameservice")
			date := c.PostForm("date")
			address := c.PostForm("address")
			loc := c.PostForm("loc")
			number := c.PostForm("number")
			phone := c.PostForm("phone")
			note := c.PostForm("note")

			person := []*Person{
				{Name: name, SubName: subname, NameService: nameservice,
					Date: date, Address: address, Location: loc, Number: number, Phone: phone, Note: note},
			}
			err := db.Update(&person)
			if err != nil {
				log.Fatal(err)
			}

			http.Redirect(c.Writer, c.Request, "/operator", 302)
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	// Start serving the application
	g.Run(":3000")
}
