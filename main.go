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
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/thinkerou/favicon"
	"github.com/xyproto/permissionbolt"
	// "github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promauto"
	// "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Struc opertator in City and location office
type Location struct {
	ID       int    `storm:"id,increment" form:"id" binding:"required"` //`form:"ID" storm:"id,increment" json:"ID"`
	Office   string `storm:"index" json:"office" form:"office" binding:"required"`
	Operator string `storm:"index" json:"operator" form:"operator" binding:"required"`
}

// Struct service
type Service struct {
	ID             int    `storm:"id,increment" form:"id" binding:"required"`
	Type           string `storm:"index" json:"type" form:"type" binding:"required"`
	NameService    string `storm:"index" json:"nameserv" form:"nameserv" binding:"required"`
	SybNameService string `storm:"index" json:"sybnameserv" form:"nameserv" binding:"required"`
}

// Struc data visitors
type Person struct {
	ID          int `storm:"id,increment" form:"id" binding:"required"` //`form:"ID" storm:"id,increment" json:"ID"`
	User        string
	Name        string `storm:"index" json:"name" form:"name" binding:"required"`               //Заявитель
	SubName     string `storm:"index" json:"subname" form:"subname" binding:"required"`         //Представитель заявитель
	NameService string `storm:"index" json:"nameservice" form:"nameservice" binding:"required"` //Услуга
	DateIn      string `storm:"index" json:"datein" form:"datein" binding:"required"`           //Дата регистрации
	DateSend    string `storm:"index" json:"datesend" form:"datesend" binding:"required"`       //Дата отправки
	DateOut     string `storm:"index" json:"dateout" form:"dateout" binding:"required"`         //Дата получения
	Address     string `storm:"index" json:"address" form:"address" binding:"required"`         //Адрес
	Location    string `storm:"index" json:"address" form:"address" binding:"required"`         //Место оператора
	Number      string `storm:"index" json:"number" form:"number" binding:"required"`           //
	Phone       string `storm:"index" json:"phone" form:"phone" binding:"required"`             //Телефон
	Note        string `storm:"index" json:"note" form:"note" binding:"required"`               //Примечание
}

// var perm, _ = permissionbolt.New()

// open database
func DB() *storm.DB {
	db, err := storm.Open("db/data.db")
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()
	return db
}

var db = DB()

func SetupRouter() *gin.Engine {

	// Set Gin to production mode
	//gin.SetMode(gin.ReleaseMode)

	g := gin.Default()

	g.Static("/web", "./web")
	g.Static("/node_modules", "./node_modules")
	g.LoadHTMLGlob("templates/*.html")

	// g := gin.New()

	// Logging middleware
	g.Use(gin.Logger())
	// Recovery middleware
	g.Use(gin.Recovery())
	g.Use(favicon.New("./assets/favicon.ico"))
	// g.Use(static.Serve("/web", static.LocalFile("/web", false)))
	// v1 := router.Group("api/v1")
	// {
	// 	v1.GET("/instructions", GetInstructions)
	// }

	return g
}

func main() {

	// http.Handle("/metrics", promhttp.Handler())
	// http.ListenAndServe(":2112", nil)

	defer db.Close()

	errdbp := db.Init(&Person{})
	if errdbp != nil {
		log.Fatal(errdbp)
	}
	errdbs := db.Init(&Service{})
	if errdbs != nil {
		log.Fatal(errdbs)
	}
	errdbl := db.Init(&Location{})
	if errdbl != nil {
		log.Fatal(errdbl)
	}

	g := SetupRouter()

	//	Blank slate, no default permissions
	//	perm.Clear()

	filename := "db/bolt.db"
	perm, err := permissionbolt.NewWithConf(filename)
	if err != nil {
		fmt.Println("Could not open database: " + filename)
		return
	}

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

	// Default user/administrator admin
	userstate.AddUser("admin", "admin", "admin@mail.ru")
	userstate.MarkConfirmed("admin")
	userstate.SetAdminStatus("admin")

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

	// Upload user from file
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

	// Logout users
	g.GET("/logout", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		userstate.Logout(usercook)
		http.Redirect(c.Writer, c.Request, "/login", 302)
	})

	// Make user as admin POST
	g.GET("/makeadmin/:user", func(c *gin.Context) {
		user := c.Param("user")
		userstate.SetAdminStatus(user)
		http.Redirect(c.Writer, c.Request, "/adminka", 302)
	})

	// Delete User from Base POST
	g.GET("/delete/:user", func(c *gin.Context) {
		user := c.Param("user")
		userstate.RemoveUser(user)
		http.Redirect(c.Writer, c.Request, "/adminka", 302)
	})

	// Delete Admin status
	g.GET("/adminoff/:user", func(c *gin.Context) {
		user := c.Param("user")
		userstate.IsAdmin(user)
		userstate.RemoveAdminStatus(user)
		http.Redirect(c.Writer, c.Request, "/adminka", 302)
	})

	// Administartort interface
	g.GET("/adminka", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := userstate.IsLoggedIn(usercook)
		isadmin := userstate.IsAdmin(usercook)

		var cheked []bool
		if isloggedin {

			var loc []Location
			err = db.All(&loc)
			fmt.Println(loc)

			if err == storm.ErrNotFound {
				c.Set("Нет данных", loc)
			}

			listusers, _ := userstate.AllUsernames()
			if isadmin {
				for _, i := range listusers {
					cheked = append(cheked, userstate.IsAdmin(i))
				}
			}
			c.HTML(http.StatusOK, "adminka.html", gin.H{"location": loc, "listusers": listusers, "is_logged_in": isloggedin, "isadmin": isadmin})
		} else {
			c.Redirect(301, "/")
		}
	})

	g.POST("/adminka", func(c *gin.Context) {

		isloggedin := isloggedin(c)

		if isloggedin {

			office := c.PostForm("office")
			fmt.Println(office)
			operator := c.PostForm("operator")
			fmt.Println(operator)

			loc := Location{
				// ID:        1,
				Office:   office,
				Operator: operator,
			}

			err := db.Save(&loc)
			if err != nil {
				log.Fatal(err)
			}

			http.Redirect(c.Writer, c.Request, "/adminka", 302)
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}

	})

	// Upload user from file
	g.POST("/uploadService", func(c *gin.Context) {
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

		csvFile, _ := os.Open(url)
		reader := csv.NewReader(bufio.NewReader(csvFile))
		for {
			line, error := reader.Read()
			if error == io.EOF {
				break
			} else if error != nil {
				log.Fatal(error)
			}

			serv := []*Service{
				{
					// ID:          line[0],
					Type:           line[0],
					NameService:    line[1],
					SybNameService: line[2]},
			}

			for _, s := range serv {
				db.Save(s)
				fmt.Println(s)
			}
		}

		http.Redirect(c.Writer, c.Request, "/adminka", 302)
	})

	// Location operators
	g.POST("/addservice", func(c *gin.Context) {

		isloggedin := isloggedin(c)

		if isloggedin {

			sybnameservice := c.PostForm("sybnameserv")
			nameservice := c.PostForm("nameserv")
			servtype := c.PostForm("type")

			service := Service{
				// ID:        1,
				Type:           servtype,
				NameService:    nameservice,
				SybNameService: sybnameservice,
			}

			err := db.Save(&service)
			if err != nil {
				log.Fatal(err)
			}

			http.Redirect(c.Writer, c.Request, "/service", 302)
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}

	})

	// Add value from file
	g.POST("/uploadValue", func(c *gin.Context) {
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
					DateIn:      line[3],
					DateSend:    line[4],
					DateOut:     line[5],
					Address:     line[6],
					Location:    line[7],
					Number:      line[8],
					Phone:       line[9],
					Note:        line[10]},
			}

			for _, p := range peeps {
				db.Save(p)
				fmt.Println(p)
			}
		}

		http.Redirect(c.Writer, c.Request, "/operator", 302)
	})

	// Register users
	g.GET("/operator", func(c *gin.Context) {
		isloggedin := isloggedin(c)
		usercook, _ := userstate.UsernameCookie(c.Request)

		if isloggedin {
			var person []Person

			var loc []Location
			err = db.All(&loc)
			fmt.Println(loc)

			timeNow := time.Now()
			timeNowF := timeNow.Format("2006-01-02T15:04")

			fmt.Println(timeNowF)

			// timeAgo := timeNow.AddDate(0, 0, -1)
			timeAgo := timeNow.Add(-12 * time.Hour)
			timeAgoF := timeAgo.Format("2006-01-02T15:04")

			fmt.Println(timeAgoF)

			err := db.Select(q.Eq("User", usercook), q.And(q.Gte("DateIn", timeAgoF), q.Lte("DateIn", timeNowF))).Find(&person)
			fmt.Println(&person)

			if err == storm.ErrNotFound {
				c.Set("Нет данных", person)
			}

			c.HTML(http.StatusOK, "operator.html", gin.H{"location": loc, "person": person, "is_logged_in": isloggedin, "timeNow": timeNowF})

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	// Register visitors POST
	g.POST("/operator", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := userstate.IsLoggedIn(usercook)

		if isloggedin {

			name := c.PostForm("name")
			subname := c.PostForm("subname")
			nameservice := c.PostForm("nameservice")
			datein := c.PostForm("datein")
			datesend := c.PostForm("datesend")
			dateout := c.PostForm("dateout")
			address := c.PostForm("address")
			loc := c.PostForm("loc")
			number := c.PostForm("number")
			phone := c.PostForm("phone")
			note := c.PostForm("note")

			peeps := []*Person{
				{User: usercook, Name: name, SubName: subname, NameService: nameservice,
					DateIn: datein, DateSend: datesend, DateOut: dateout, Address: address, Location: loc, Number: number, Phone: phone, Note: note},
			}

			datepars, _ := time.Parse(time.RFC3339, datein)
			datef := datepars.Format("2006-01-02T15:04")

			fmt.Println(datef)

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

	// Find value on date
	g.GET("/controller", func(c *gin.Context) {
		isloggedin := isloggedin(c)

		if isloggedin {

			var person []Person

			listusers, err1 := userstate.AllUsernames()
			if err1 != nil {
				log.Fatal(err1)
			}

			users := c.Query("users")
			date := c.Query("date")

			datep, _ := time.Parse("2006-01-02T15:04", date)
			datePF := datep.Format("2006-01-02T15:04")

			dateAdd := datep.Add(-12 * time.Hour)
			dateAF := dateAdd.Format("2006-01-02T15:04")
			fmt.Println(dateAF)
			fmt.Println(datePF)

			err := db.Select(q.Eq("User", users), q.And(q.Gte("DateIn", dateAF), q.Lte("DateIn", datePF))).Find(&person)
			if err == storm.ErrNotFound {
				c.Set("Нет данных", person)
			}

			c.HTML(http.StatusOK, "controller.html", gin.H{"person": person, "is_logged_in": isloggedin, "listusers": listusers})

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	// Find user value on date
	g.GET("/history", func(c *gin.Context) {
		usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := isloggedin(c)

		if isloggedin {

			var person []Person

			date := c.Query("date")

			// timeNow := time.Now()
			// dateAdd := datepars.AddDate(0, 0, -12)

			datep, _ := time.Parse("2006-01-02T15:04", date)
			datePF := datep.Format("2006-01-02T15:04")

			dateAdd := datep.Add(-12 * time.Hour)
			dateAF := dateAdd.Format("2006-01-02T15:04")
			fmt.Println(dateAF)
			fmt.Println(datePF)

			err := db.Select(q.Eq("User", usercook), q.And(q.Gte("DateIn", dateAF), q.Lte("DateIn", datePF))).Find(&person)
			if err == storm.ErrNotFound {
				c.Set("Нет данных", person)
			}

			c.HTML(http.StatusOK, "history.html", gin.H{"person": person, "is_logged_in": isloggedin})

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	// Find service NOT WORK
	g.GET("/serviceSort", func(c *gin.Context) {
		// usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := isloggedin(c)

		if isloggedin {

			var service []Service

			// date := c.Query("nameserv")

			err = db.All(&service)

			// err := db.Select(q.Eq("User", usercook), q.And(q.Gte("DateIn", dateAF), q.Lte("DateIn", datePF))).Find(&person)
			// if err == storm.ErrNotFound {
			// 	c.Set("Нет данных", person)
			// }

			c.HTML(http.StatusOK, "history.html", gin.H{"service": service, "is_logged_in": isloggedin})

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	// Delete value on id
	g.GET("/removeval/:struct/:id", RemVal)

	// Counte all values
	g.GET("/report", func(c *gin.Context) {
		isloggedin := isloggedin(c)
		if isloggedin {
			c.HTML(http.StatusOK, "report.html", gin.H{"is_logged_in": isloggedin})
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	// Select report
	g.POST("/report", func(c *gin.Context) {
		isloggedin := isloggedin(c)
		report := c.PostForm("report")
		fmt.Println(report)
		if isloggedin {

			switch report {
			case "report1":
				query := db.Select()
				count, err := query.Count(new(Person))
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(count)
				c.HTML(http.StatusOK, "report.html", gin.H{"count": count, "is_logged_in": isloggedin})

			case "report2":
				query := db.Select(q.Eq("User", "ren"))
				count, err := query.Count(new(Person))
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(count)
				c.HTML(http.StatusOK, "report.html", gin.H{"count": count, "is_logged_in": isloggedin})

			case "report3":
				query := db.Select(q.Eq("User", "bil"))
				count, err := query.Count(new(Person))
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(count)
				c.HTML(http.StatusOK, "report.html", gin.H{"count": count, "is_logged_in": isloggedin})
			}

		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}

	})

	g.GET("/service", func(c *gin.Context) {
		isloggedin := isloggedin(c)

		var service []Service

		err = db.All(&service)
		fmt.Println(service)

		if isloggedin {
			c.HTML(http.StatusOK, "service.html", gin.H{"is_logged_in": isloggedin, "service": service})
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	// Export report to PDF
	g.POST("/pdfexp", func(c *gin.Context) {
		isloggedin := isloggedin(c)
		if isloggedin {

			pdf := gofpdf.New("P", "mm", "A4", "")
			pdf.SetTopMargin(30)
			pdf.SetHeaderFunc(func() {
				url := "assets/blue-mark_cnzgry.png"
				pdf.Image(url, 10, 6, 30, 0, false, "", 0, "")
				pdf.SetY(5)
				pdf.SetFont("Arial", "B", 15)
				pdf.Cell(80, 0, "")
				pdf.CellFormat(30, 10, "Title", "1", 0, "C", false, 0, "")
				pdf.Ln(20)
			})
			pdf.SetFooterFunc(func() {
				pdf.SetY(-15)
				pdf.SetFont("Arial", "I", 8)
				pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()),
					"", 0, "C", false, 0, "")
			})
			pdf.AliasNbPages("")
			pdf.AddPage()
			pdf.SetFont("Times", "", 12)
			for j := 1; j <= 40; j++ {
				pdf.CellFormat(0, 10, fmt.Sprintf("Printing line number %d", j),
					"", 1, "", false, 0, "")
			}
			err := pdf.OutputFileAndClose("upload/hello1.pdf")
			if err != nil {
				log.Fatal(err)
			}

			c.HTML(http.StatusOK, "report.html", gin.H{"is_logged_in": isloggedin})
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})

	// Edit value
	g.POST("/edit/:id", func(c *gin.Context) {
		id := c.Param("id")
		usercook, _ := userstate.UsernameCookie(c.Request)
		isloggedin := userstate.IsLoggedIn(usercook)
		var person Person

		findVal := db.Select(q.Eq("ID", id))
		err := findVal.First(&person)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(&person)

		if isloggedin {

			// id := c.PostForm("id")
			name := c.PostForm("name")
			subname := c.PostForm("subname")
			nameservice := c.PostForm("nameservice")
			datein := c.PostForm("datein")
			datesend := c.PostForm("datesend")
			dateout := c.PostForm("dateout")
			address := c.PostForm("address")
			loc := c.PostForm("loc")
			number := c.PostForm("number")
			phone := c.PostForm("phone")
			note := c.PostForm("note")

			peeps := &Person{ID: person.ID, User: usercook, Name: name, SubName: subname, NameService: nameservice,
				DateIn: datein, DateSend: datesend, DateOut: dateout, Address: address, Location: loc, Number: number, Phone: phone, Note: note}
			fmt.Println(peeps)

			db.Update(peeps)
			fmt.Println(peeps)

			http.Redirect(c.Writer, c.Request, "/operator", 302)
		} else {
			c.AbortWithStatus(http.StatusForbidden)
			fmt.Fprint(c.Writer, "Permission denied!")
		}
	})
	ginpprof.Wrap(g)

	// 404 page
	g.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "404.html", gin.H{})
	})

	// Start serving the application
	g.Run(":3000")
}

// Delete value on id function
func RemVal(c *gin.Context) {

	param := c.Param("struct")

	switch param {
	case "Person":
		id := c.Param("id")

		query := db.Select(q.Eq("ID", id))
		count, err := query.Count(new(Person))
		if err != nil {
			log.Fatal(err)
		}
		query.Delete(new(Person))

		fmt.Println(count)
		fmt.Println(id)
		http.Redirect(c.Writer, c.Request, "/operator", 302)

	case "Service":

		id := c.Param("id")

		query := db.Select(q.Eq("ID", id))
		count, err := query.Count(new(Service))
		if err != nil {
			log.Fatal(err)
		}
		query.Delete(new(Service))

		fmt.Println(count)
		fmt.Println(id)
		http.Redirect(c.Writer, c.Request, "/service", 302)

	case "Location":

		id := c.Param("id")

		query := db.Select(q.Eq("ID", id))
		count, err := query.Count(new(Location))
		if err != nil {
			log.Fatal(err)
		}
		query.Delete(new(Location))

		fmt.Println(count)
		fmt.Println(id)
		http.Redirect(c.Writer, c.Request, "/adminka", 302)
	}

}

// ----------- chek login func ------------------
func setUserStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		if token, err := c.Cookie("token"); err == nil || token != "" {
			c.Set("is_logged_in", true)
		} else {
			c.Set("is_logged_in", false)
		}
	}
}

func ensureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		loggedInInterface, _ := c.Get("is_logged_in")
		loggedIn := loggedInInterface.(bool)
		if !loggedIn {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func ensureNotLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		loggedInInterface, _ := c.Get("is_logged_in")
		loggedIn := loggedInInterface.(bool)
		if loggedIn {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
