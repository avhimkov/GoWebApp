package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

// var tmpServiceList []Service

// This function is used for setup before executing the test functions
func TestMain(m *testing.M) {
	//Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Run the other tests
	os.Exit(m.Run())
}

// Helper function to create a router during testing
func getRouter(withTemplates bool) *gin.Engine {
	r := gin.Default()
	if withTemplates {
		r.LoadHTMLGlob("templates/*")
	}
	return r
}

// This function is used to store the main lists into the temporary one
// for testing
/* func saveLists() {
	tmpServiceList = serviceList
} */

// This function is used to restore the main lists from the temporary one
/* func restoreLists() {
	serviceList = tmpServiceList
} */

// For this demo, we're storing the article list in memory
// In a real application, this list will most likely be fetched
// from a database or from static files
/* var serviceList = []Service{
	{ID: 3, Type: "Федеральная", NameService: "Получение паспорта", SybNameService: "Загран паспорта"},
	{ID: 12, Type: "Региональная", NameService: "Получение ИНН", SybNameService: ""},
	{ID: 15, Type: "Муниципальная", NameService: "Выдача разрешение на вылов рыбы", SybNameService: ""},
	// article{ID: 1, Title: "Article 1", Content: "Article 1 body"},
	// article{ID: 2, Title: "Article 2", Content: "Article 2 body"},
} */

// Return a list of all the articles
/* func getAllService() []Service {
	return serviceList
} */

// Helper function to process a request and test its response
func testHTTPResponse(t *testing.T, r *gin.Engine, req *http.Request, f func(w *httptest.ResponseRecorder) bool) {

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create the service and process the above request.
	r.ServeHTTP(w, req)

	if !f(w) {
		t.Fail()
	}
}

func TestShowIndexPageUnauthenticated(t *testing.T) {
	r := getRouter(true)

	r.GET("/login", loginGet)

	// Create a request to send to the above route
	req, _ := http.NewRequest("GET", "/login", nil)

	testHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {

		// Test that the http status code is 200
		statusOK := w.Code == http.StatusOK

		// Test that the page title is "Home Page"
		// You can carry out a lot more detailed tests using libraries that can
		// parse and process HTML pages
		// p, err := ioutil.ReadAll(w.Body)
		// pageOK := err == nil && strings.Index(string(p), "<h1>Вход</h1>") > 0

		return statusOK
	})
}

func TestShowRegisterPageUnauthenticated(t *testing.T) {

	r := getRouter(true)

	r.GET("/register", registerGet)

	// Create a request to send to the above route
	req, _ := http.NewRequest("GET", "/register", nil)

	testHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {

		// Test that the http status code is 200
		statusOK := w.Code == http.StatusOK

		// Test that the page title is "Home Page"
		// You can carry out a lot more detailed tests using libraries that can
		// parse and process HTML pages
		// p, err := ioutil.ReadAll(w.Body)
		// pageOK := err == nil && strings.Index(string(p), "<h1>Вход</h1>") > 0

		return statusOK
	})
}

func TestShowAdminkaPageAuthenticated(t *testing.T) {
	r := getRouter(true)

	r.GET("/adminka", adminkaGet)

	// Create a request to send to the above route
	req, _ := http.NewRequest("GET", "/adminka", nil)

	testHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {

		r.GET("/adminka", func(c *gin.Context) {
			usercook, _ := userstate.UsernameCookie(c.Request)
			isloggedin := isloggedin(c)
			// logintryst := userstate.CorrectPassword("admin", "admin")
			userstate.Login(c.Writer, "admin")
			userstate.IsAdmin("admin")
			isadmin := userstate.IsAdmin(usercook)

			c.HTML(http.StatusOK, "adminka.html", gin.H{"is_logged_in": isloggedin, "isadmin": isadmin})
			// c.String(200, "OK")
		})

		// Test that the http status code is 200
		statusOK := w.Code == http.StatusOK

		// Test that the page title is "Home Page"
		// You can carry out a lot more detailed tests using libraries that can
		// parse and process HTML pages
		// p, err := ioutil.ReadAll(w.Body)
		// pageOK := err == nil && strings.Index(string(p), "<h1>Вход</h1>") > 0

		return statusOK
	})
}
