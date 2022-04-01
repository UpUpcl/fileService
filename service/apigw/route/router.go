package route

import (
	"apigw/assets"
	"apigw/handler"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	assetfs "github.com/moxiaomomo/go-bindata-assetfs"
	"net/http"
	"strings"
)

type binaryFileSystem struct {
	fs http.FileSystem
}

func (b *binaryFileSystem) Open(name string) (http.File, error) {
	return b.fs.Open(name)
}

func (b *binaryFileSystem)Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath){
		if _, err := b.fs.Open(p);err !=nil{
			return false
		}
		return true
	}
	return false
}

func BinaryFileSystem(root string) *binaryFileSystem {
	fs := &assetfs.AssetFS{
		Asset:     assets.Asset,
		AssetDir:  assets.AssetDir,
		AssetInfo: assets.AssetInfo,
		Prefix:    root,
	}
	return &binaryFileSystem{fs}
}


func Router() *gin.Engine {
	router := gin.Default()

	router.Use(static.Serve("/static/", BinaryFileSystem("static")))
	//router.Static("/static/", "/Users/chenlei/GolandProjects/src/filestore-server/service/apigw/static")

	router.GET("/user/signup", handler.SignupHandler)
	router.POST("/user/signup", handler.DoSignupHandler)

	router.GET("/user/signin", handler.SignInHandler)
	router.POST("/user/signin", handler.DoSignInHandler)

	router.Use(handler.HTTPInterceptor())

	// 用户查询
	router.POST("/user/info", handler.UserInfoHandler)

	router.POST("/file/query", handler.FileQueryHandler)

	router.POST("/file/update", handler.FileNameUpdateHandler)
	return router
}
