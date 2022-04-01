package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"upload/api"
)

func Router() *gin.Engine {
	router := gin.Default()

	router.Static("/static/", "/Users/chenlei/GolandProjects/src/filestore-server/static")

	//router.Use()

	router.Use(Cors())
	// 文件上传
	router.POST("/file/upload", api.DoUploadHandler)

	// 秒传
	router.POST("/file/fastupload", api.TryFastUploadHandle)
	// 分块上传
	router.POST("/file/mpupload/init", api.InitialMultipartUploadHandler)
	router.POST("/file/mpupload/uppart", api.UploadPartHandler)
	router.POST("/file/mpupload/complete", api.CompleteUploadHandler)
	return router
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")      //服务器支持的所有跨域请求的方
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}