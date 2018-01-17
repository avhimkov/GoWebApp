package main

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/drivers/:id", GetDriverHandler)

	r.Run(":3000")
}

func GetDriverHandler(c *gin.Context) {
	//fmt.Println(c.Request.RequestURI)
	if strings.EqualFold(c.Request.RequestURI, "/drivers/price") {
		fmt.Fprintf(c.Writer, "GetDriversPrice")
	} else {
		id := c.Param("id")
		fmt.Fprintf(c.Writer, "GetDriversId "+id)
	}
}
