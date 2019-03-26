package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DeanThompson/ginpprof"
	"github.com/asdine/storm"

	// "github.com/asdine/storm/codec/json"

	"github.com/gin-gonic/gin"

	// "github.com/mongodb/mongo-go-driver/mongo"
	"github.com/thinkerou/favicon"
	"github.com/xyproto/permissionbolt"
	// "github.com/zebresel-com/mongodm"
	// "github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promauto"
	// "github.com/prometheus/client_golang/prometheus/promhttp"
)

// type Subscriber struct {
// 	ID   int    `storm:"id,increment" form:"subid" binding:"required"`
// 	Name string `storm:"index" json:"subname" form:"subname" binding:"required"`
// }

// type SubSubscriber struct {
// 	ID      int    `storm:"id,increment" form:"subsubid" binding:"required"`
// 	SubName string `storm:"index" json:"subname" form:"subname" binding:"required"`
// }

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
	Location    string `storm:"index" json:"location" form:"location" binding:"required"`       //Место оператора
	Number      string `storm:"index" json:"number" form:"number" binding:"required"`           //
	Phone       string `storm:"index" json:"phone" form:"phone" binding:"required"`             //Телефон
	Note        string `storm:"index" json:"note" form:"note" binding:"required"`               //Примечание
}

//
var db = DB()

//middlleware db
var perm, err = perminit("db/bolt.db")

//middlleware var
var userstate = perm.UserState()
var permissionHandler = permHandler()

//open databas
func DB() *storm.DB {
	db, err := storm.Open("db/data.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// func mogoDBinit() {
// 	file, err := ioutil.ReadFile("locals.json")

// 	if err != nil {
// 		fmt.Printf("File error: %v\n", err)
// 		os.Exit(1)
// 	}

// 	var localMap map[string]map[string]string
// 	json.Unmarshal(file, &localMap)

// 	dbConfig := &mongodm.Config{
// 		DialInfo: &mgo.DialInfo{
// 			Addrs:    []string{"127.0.0.1"},
// 			Timeout:  3 * time.Second,
// 			Database: "mongodm_sample",
// 			Username: "admin",
// 			Password: "admin",
// 			Source:   "admin",
// 		},
// 		Locals: localMap["en-US"],
// 	}

// 	connection, err := mongodm.Connect(dbConfig)

// 	if err != nil {
// 		fmt.Println("Database connection error: %v", err)
// 	}
// }

//middlleware init
func perminit(db string) (*permissionbolt.Permissions, error) {
	perm, err := permissionbolt.NewWithConf(db)
	if err != nil {
		log.Fatal(err)
		// fmt.Println("Could not open database: " + db)
		return nil, nil
	}

	return perm, nil
}

//middlleware gin config
func permHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set up a middleware handler for Gin, with a custom "permission denied" message.
		// permissionHandler := func(c *gin.Context) {
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
}

func isloggedin(c *gin.Context) bool {
	usercook, _ := userstate.UsernameCookie(c.Request)
	isloggedin := userstate.IsLoggedIn(usercook)
	return isloggedin
}

//gin init func
func SetupRouter() *gin.Engine {

	// Set Gin to production mode
	//gin.SetMode(gin.ReleaseMode)

	g := gin.Default()

	g.Static("/web", "./web")
	g.Static("/assets", "./assets")
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

	g.Use(permissionHandler)

	return g
}

func main() {

	// client, err := mongo.NewClient("mongodb://localhost:27017")

	//prometheus
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

	// errdbsb := db.Init(&Subscriber{})
	// if errdbsb != nil {
	// 	log.Fatal(errdbsb)
	// }

	// errdbsbsb := db.Init(&SubSubscriber{})
	// if errdbsbsb != nil {
	// 	log.Fatal(errdbsbsb)
	// }

	g := SetupRouter()

	//	Blank slate, no default permissions
	//	perm.Clear()

	// Default user /administrator admin
	userstate.AddUser("admin", "admin", "admin@mail.ru")
	userstate.MarkConfirmed("admin")
	userstate.SetAdminStatus("admin")

	g.GET("/", indexPage)

	// Registaration Users GET
	g.GET("/register", registerGet)

	// Registaration Users OPOST
	g.POST("/register", registerPost)

	// Upload user from file
	g.POST("/uploadUsers", uploadUsers)

	// Loging Users GET
	g.GET("/login", loginGet)

	// Loging Users
	g.POST("/login", loginPost)

	// Logout users
	g.GET("/logout", logout)

	// Make user as admin POST
	g.GET("/makeadmin/:user", makeadmin)

	// Delete User from Base POST
	g.GET("/delete/:user", deleteUser)

	// Delete Admin status
	g.GET("/adminoff/:user", adminoff)

	// Administartort interface
	g.GET("/adminka", adminkaGet)

	g.POST("/adminka", adminkaPost)

	// Upload user from file
	g.POST("/uploadService", uploadService)

	// Location operators
	g.POST("/addservice", addservice)

	// Add value from file
	g.POST("/uploadValue", uploadValue)

	// Register users
	g.GET("/operator", operatorGet)

	// Register visitors POST
	g.POST("/operator", operatorPost)

	// Find value on date
	g.GET("/controller", controller)

	// Find user value on date
	g.GET("/history", history)

	// Service page
	g.GET("/service", service)

	// Find service NOT WORK
	g.GET("/serviceSort", serviceSort)

	// Delete value on id
	g.GET("/removeval/:struct/:id", RemVal)

	// Counte all values
	g.GET("/report", reportGet)

	// Select report
	g.POST("/report", reportPost)

	// Export report to PDF
	g.POST("/pdfexp", pdfexp)

	// Edit value
	g.POST("/edit/:id", editVal)

	// Edit login TODO
	g.POST("/registerEdit", registerEdit)

	ginpprof.Wrap(g)

	// 404 page
	g.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "404.html", gin.H{})
	})

	// Start serving the application
	g.Run(":3000")
}

// JSON
// func (s Service) MarshalJSON() ([]byte, error) {
// 	return []byte(fmt.Sprintf("%v %v", s.NameService, s.SybNameService)), nil
// }

// func (s *Service) UnmarshalJSON(value []byte) error {
// 	parts := strings.Split(string(value), "/")
// 	// m.MonthNumber = strconv.ParseInt(parts[0], 10, 32)
// 	// m.YearNumber = strconv.ParseInt(parts[1], 10, 32)
// 	fmt.Println(parts)

// 	return nil
// }

// ----------- chek login func ------------------
/* func setUserStatus() gin.HandlerFunc {
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
} */
