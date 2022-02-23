package route

import (
	"filestore-server/handler"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	// gin framework
	router := gin.Default()

	// 处理静态资源
	router.Static("/static", "/Users/chenlei/GolandProjects/src/filestore-server/static")


	// 用户注册
	router.GET("/user/signup", handler.SignupHandler)
	router.POST("/user/signup", handler.DoSignupHandler)
	// 用户登陆
	router.GET("/user/signin", handler.SignInHandler)
	router.POST("/user/signin", handler.DoSignInHandler)

	// 加入中间价 校验token
	router.Use(handler.HTTPInterceptor())

	//User之后要经过token验证

	// 用户信息接口
	router.POST("/user/info", handler.UserInfoHandler)

	// 文件操作
	// 普通上传
	// 上传static
	router.GET("/file/upload_or", handler.UploadHandler)
	router.POST("/file/upload_or", handler.DoUploadHandler)

	// 上传成功
	router.GET("/file/upload_or/suc", handler.UploadSucHandler)
	// 获取file元信息
	router.GET("/file/meta", handler.GetFileMetaHandler)
	// 查询
	router.POST("/file/query", handler.FileQueryHandler)
	// 下载文件
	router.GET("/file/download", handler.DownloadHandler)
	router.POST("/file/download", handler.DownloadHandler)
	// 更新文件元信息
	router.POST("/file/update", handler.FileNameUpdateHandler)
	// 删除源文件
	router.POST("/file/delete", handler.FileDeleteHandler)
	// 秒传
	router.POST("/file/fastupload", handler.TryFastUploadHandle)
	// oss下载
	router.POST("/file/downloadurl", handler.DownloadURLHandler)
	// 分块传输
	router.POST("/file/mpupload/init", handler.InitialMultipartUploadHandler)
	router.POST("/file/mpupload/uppart", handler.UploadPartHandler)
	router.POST("/file/mpupload/complete", handler.CompleteUploadHandler)

	return router
}