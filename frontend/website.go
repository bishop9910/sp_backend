package frontend

import (
	"github.com/gin-gonic/gin"
)

func WebInit(router *gin.Engine) {
	router.Static("/assets", "./web/assets")

	router.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	router.GET("/:path", func(c *gin.Context) {
		c.File("./web/index.html")
	})
	router.GET("/:path/*any", func(c *gin.Context) {
		c.File("./web/index.html")
	})
}
