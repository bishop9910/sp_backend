package frontend

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func WebInit(router *gin.Engine) {
	router.Static("/assets", "./web/assets")

	router.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./web/favicon.ico")
	})

	router.GET("/:path", func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			c.Next()
			return
		}
		c.File("./web/index.html")
	})
	router.GET("/:path/*any", func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/swagger") {
			c.Next()
			return
		}
		c.File("./web/index.html")
	})
}
