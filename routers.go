package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

func indexPage(c *gin.Context) {
	isloggedin := isloggedin(c)
	if isloggedin {
		c.HTML(http.StatusOK, "operator.html", gin.H{"is_logged_in": isloggedin})
	} else {
		http.Redirect(c.Writer, c.Request, "/login", 302)
	}
}

func registerGet(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{})
}

func registerPost(c *gin.Context) {

	username := c.PostForm("username")
	pass := c.PostForm("password")
	mail := c.PostForm("email")

	userstate.AddUser(username, pass, mail)
	userstate.Login(c.Writer, username)
	userstate.MarkConfirmed(username)

	http.Redirect(c.Writer, c.Request, "/", 302)
}

func loginGet(c *gin.Context) {

	isloggedin := isloggedin(c)
	c.HTML(http.StatusOK, "login.html", gin.H{"title": "Login Page", "is_logged_in": isloggedin})
}

func loginPost(c *gin.Context) {
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
}

func logout(c *gin.Context) {
	usercook, _ := userstate.UsernameCookie(c.Request)
	userstate.Logout(usercook)
	http.Redirect(c.Writer, c.Request, "/login", 302)
}

func makeadmin(c *gin.Context) {
	user := c.Param("user")
	userstate.SetAdminStatus(user)
	http.Redirect(c.Writer, c.Request, "/adminka", 302)
}

func deleteUser(c *gin.Context) {
	user := c.Param("user")
	userstate.RemoveUser(user)
	http.Redirect(c.Writer, c.Request, "/adminka", 302)
}

func adminoff(c *gin.Context) {
	user := c.Param("user")
	userstate.IsAdmin(user)
	userstate.RemoveAdminStatus(user)
	http.Redirect(c.Writer, c.Request, "/adminka", 302)
}

func uploadUsers(c *gin.Context) {

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

	// peopleJSON, _ := json.Marshal(people)
	// fmt.Println(string(peopleJSON))

}

func adminkaGet(c *gin.Context) {
	usercook, _ := userstate.UsernameCookie(c.Request)
	isloggedin := userstate.IsLoggedIn(usercook)
	isadmin := userstate.IsAdmin(usercook)

	// var cheked []bool
	if isloggedin {

		var loc []Location
		err = db.All(&loc)
		fmt.Println(loc)

		if err == storm.ErrNotFound {
			c.Set("Нет данных", loc)
		}

		listusers, _ := userstate.AllUsernames()
		// if isadmin {
		// 	for _, i := range listusers {
		// 		cheked = append(cheked, userstate.IsAdmin(i))
		// 	}
		// }
		// fmt.Println(cheked)

		c.HTML(http.StatusOK, "adminka.html", gin.H{"location": loc, "listusers": listusers, "is_logged_in": isloggedin, "isadmin": isadmin})
	} else {
		c.Redirect(301, "/")
	}
}

/* func (ctrl TestController) TestConfig(c *gin.Context){
	var testForm form.TestForm
	if c.ShouldBindWith(&testForm, binding.FormPost) != nil {
		c.Redirect(302, "/v1/console/config/request_failed")
	}
	TestModel.TestConfig(testForm)
	fmt.Println(testForm.age)
	c.Redirect(302, "/v1/console/config/success")
} */

func adminkaPost(c *gin.Context) {

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

}

// Upload user from file
func uploadService(c *gin.Context) {
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
}

// Location operators
func addservice(c *gin.Context) {

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

}

// Add value from file
func uploadValue(c *gin.Context) {
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
}

// Register users
func operatorGet(c *gin.Context) {
	isloggedin := isloggedin(c)
	usercook, _ := userstate.UsernameCookie(c.Request)

	if isloggedin {
		var person []Person

		var loc []Location
		err = db.All(&loc)
		// fmt.Println(loc)

		var service []Service
		err = db.All(&service)
		// fmt.Println(service)

		var servicename []Service
		err = db.All(&servicename)
		// fmt.Println(servicename)

		// var arrayServ []string

		// for i := range servicename {

		// 	parts := ("\"" + servicename[i].NameService + "\",")

		// 	arrayServ = append(arrayServ, parts)
		// }

		// fmt.Println(arrayServ)

		timeNow := time.Now()
		timeNowF := timeNow.Format("2006-01-02T15:04")

		// fmt.Println(timeNowF)

		// timeAgo := timeNow.AddDate(0, 0, -1)
		timeAgo := timeNow.Add(-12 * time.Hour)
		timeAgoF := timeAgo.Format("2006-01-02T15:04")

		// fmt.Println(timeAgoF)

		err := db.Select(q.Eq("User", usercook), q.And(q.Gte("DateIn", timeAgoF), q.Lte("DateIn", timeNowF))).Find(&person)
		// fmt.Println(&person)

		if err == storm.ErrNotFound {
			c.Set("Нет данных", person)
		}

		c.HTML(http.StatusOK, "operator.html", gin.H{"location": loc, "service": service, "servicename": servicename, "person": person, "is_logged_in": isloggedin, "timeNow": timeNowF})

	} else {
		c.AbortWithStatus(http.StatusForbidden)
		fmt.Fprint(c.Writer, "Permission denied!")
	}
}

// Register visitors POST
func operatorPost(c *gin.Context) {
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
}

// Find value on date
func controller(c *gin.Context) {
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
}

// Edit value
func editVal(c *gin.Context) {
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
}

// Find user value on date
func history(c *gin.Context) {
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
}

// Service page
func service(c *gin.Context) {
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
}

// :TODO
// Find service NOT WORK
func serviceSort(c *gin.Context) {
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
}

// Counte all values
func reportPost(c *gin.Context) {
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

}

// Select report
func reportGet(c *gin.Context) {
	isloggedin := isloggedin(c)
	if isloggedin {
		c.HTML(http.StatusOK, "report.html", gin.H{"is_logged_in": isloggedin})
	} else {
		c.AbortWithStatus(http.StatusForbidden)
		fmt.Fprint(c.Writer, "Permission denied!")
	}
}

// Export report to PDF
func pdfexp(c *gin.Context) {
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
