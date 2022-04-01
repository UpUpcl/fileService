package handler

import (
	"apigw/common"
	User "apigw/proto/account"
	Upload "apigw/proto/upload"
	"apigw/util"
	"context"
	"apigw/proto/download"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/v3"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

var (
	userCli User.UserService
	upCli Upload.UploadService
	dlCli download.DownloadService
)

func init()  {
	service := micro.NewService(
		micro.Name("go.micro.service.apigw"),
		micro.Registry(consul.NewRegistry()))

	service.Init()
	cli := service.Client()
	userCli = User.NewUserService("go.micro.service.user", cli)
	upCli = Upload.NewUploadService("go.micro.service.upload", cli)
	dlCli = download.NewDownloadService("go.micro.service.download", cli)
}

func SignupHandler(c *gin.Context)  {
	c.Redirect(http.StatusFound, "http://"+c.Request.Host+"/static/view/signup.html")
}

// DoSignupHandler 处理注册请求post
func DoSignupHandler(c *gin.Context)  {

	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("password")
	resp, err := userCli.Signup(context.TODO(), &User.ReqSignup{
		Username: username,
		Password: passwd,
	})
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":resp.Code,
		"message":resp.Message,
	})
}

func SignInHandler(c *gin.Context)  {

	c.Redirect(http.StatusFound, "http://"+c.Request.Host+"/static/view/signin.html")

}

// DoSignInHandler 响应登陆
func DoSignInHandler(c *gin.Context)  {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")

	// 1. 校验用户名以及密码
	resp, err := userCli.Signin(context.TODO(), &User.ReqSignin{
		Username: username,
		Password: password,
	})

	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	if resp.Code != common.StatusOK{
		c.JSON(200, gin.H{
			"msg": "登陆失败",
			"code": resp.Code,
		})
		return
	}
	// 动态获取上传地址
	upEntryResp, err := upCli.UploadEntry(context.TODO(), &Upload.ReqEntry{})
	if err != nil {
		log.Println(err.Error())
	}else if upEntryResp.Code != common.StatusOK{
		log.Println(upEntryResp.Message)
	}

	// 动态获取下载入口
	dlEntryResp, err :=  dlCli.DownloadEntry(context.TODO(), &download.ReqEntry{})
	if err != nil {
		log.Println(err.Error())
	}else if dlEntryResp.Code != common.StatusOK{
		log.Println(dlEntryResp.Message)
	}
	cliResp := util.RespMsg{
		Code: int(common.StatusOK),
		Msg:  "登陆成功",
		Data: struct {
			Location string
			Username string
			Token string
			UploadEntry string
			DownloadEntry string
		}{
			Location: "http://"+c.Request.Host+"/static/view/home.html",
			Username: username,
			Token: resp.Token,
			//UploadEntry: "http://upload.fileserver.com",
			//DownloadEntry: "http://download.fileserver.com",
			UploadEntry: upEntryResp.Entry,
			DownloadEntry: dlEntryResp.Entry,
		},
	}
	c.Data(http.StatusOK, "application/json", cliResp.JsonToBytes())
}

func UserInfoHandler(c *gin.Context)  {
	// 1. 解析请求参数

	username := c.Request.FormValue("username")

	resp, err := userCli.UserInfo(context.TODO(), &User.ReqUserInfo{
		Username: username,
	})
	// 3. 查询用户信息
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	// 4. 组装并切相应用户数据
	cliResp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: gin.H{
			"Username":username,
			"SignupAt":resp.SignupAt,
			"LastActive":resp.LastActiveAt,
		},
	}
	c.Data(http.StatusOK, "application/json", cliResp.JsonToBytes())
}

// FileQueryHandler 文件查询
func FileQueryHandler(c *gin.Context) {

	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	username := c.Request.FormValue("username")

	rpcResp, err :=  userCli.UserFiles(context.TODO(), &User.ReqUserFile{
		Username: username,
		Limit: int32(limitCnt),
	})
	if err != nil {
		log.Println("FileQueryHandler :", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if len(rpcResp.FileDate) <= 0{
		rpcResp.FileDate = []byte("[]")
	}
	c.Data(http.StatusOK, "application/json", rpcResp.FileDate)
}

// FileMetaUpdateHandler 更新元信息
func FileNameUpdateHandler(c *gin.Context) {
	opType := c.Request.FormValue("op")
	fileSha1 := c.Request.FormValue("filehash")
	newFileName := c.Request.FormValue("filename")
	username := c.Request.FormValue("username")
	if opType != "0" && len(newFileName)< 1{
		c.Status(http.StatusForbidden)
		return
	}

	rpcResp, err := userCli.UserFileRename(context.TODO(), &User.ReqUserFileRename{
		Username:    username,
		Filehash:    fileSha1,
		NewFileName: newFileName,
	})
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(rpcResp.FileData) <= 0{
		rpcResp.FileData = []byte("[]")
	}
	c.Data(http.StatusOK, "application/json", rpcResp.FileData)
}

