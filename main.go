package main

import (
	"net/http"
	"log"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/xyproto/permissionbolt"
)

var router *gin.Engine

func main() {
	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	// Set the router as the default one provided by Gin
	//router = gin.Default()

	g := gin.New()
	perm, err := permissionbolt.New()
	if err != nil {
		log.Fatalln(err)
	}

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

	// Process the templates at the start so that they don't have to be loaded
	// from the disk again. This makes serving HTML pages very fast.
	 router.LoadHTMLGlob("templates/*")

	// Initialize the routes
	initializeRoutes()

	// Start serving the application
	router.Run()
}

//func render(c *gin.Context, data gin.H, templateName string) {
//	loggedInInterface, _ := c.Get("is_logged_in")
//	data["is_logged_in"] = loggedInInterface.(bool)
//
//	switch c.Request.Header.Get("Accept") {
//	case "application/json":
//		// Respond with JSON
//		c.JSON(http.StatusOK, data["payload"])
//	case "application/xml":
//		// Respond with XML
//		c.XML(http.StatusOK, data["payload"])
//	default:
//		// Respond with HTML
//		c.HTML(http.StatusOK, templateName, data)
//	}
//}
