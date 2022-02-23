package handler

import (
	"filestore-server/db"
	"filestore-server/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)


const  (
	pwd_salt = "*#890"
)

// SignupHandler 处理用户注册请求get
func SignupHandler(c *gin.Context)  {
	c.Redirect(http.StatusFound, "http://"+c.Request.Host+"/static/view/signup.html")
}

// DoSignupHandler 处理注册请求post
func DoSignupHandler(c *gin.Context)  {

	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("password")

	if len(username) < 3 || len(passwd) < 5 {
		c.JSON(http.StatusOK, gin.H{
			"msg" : "Invalid parameter",
			"code": -1,
		})
		return
	}
	enc_passwd := util.Sha1([]byte(passwd+pwd_salt))
	suc := db.UserSignup(username, enc_passwd)
	if suc {
		c.JSON(http.StatusOK, gin.H{
			"msg" : "Signup succeeded",
			"code": 0,
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"msg" : "Signup failed",
			"code": -2,
		})
	}
}

// SignInHandler 用户登陆按钮
func SignInHandler(c *gin.Context)  {

	c.Redirect(http.StatusFound, "http://"+c.Request.Host+"/static/view/signin.html")

}

// DoSignInHandler 响应登陆
func DoSignInHandler(c *gin.Context)  {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	encPasswd := util.Sha1([]byte(password+pwd_salt))
	// 1. 校验用户名以及密码
	pwdChecked := db.UserSignIn(username, encPasswd)
	if ! pwdChecked{
		c.JSON(http.StatusOK, gin.H{
			"msg" : "Login Failed",
			"code": -1,
		})
		return
	}
	// 2. 生成访问凭证token
	token := GenToken(username)
	fmt.Println("User token is :", token)
	upRes := db.UpdateToken(username, token)
	if !upRes {
		c.JSON(http.StatusOK, util.RespMsg{
			Msg : "Login Failed token",
			Code : -2,
		})
		return
	}
	// 3. 登陆成功后重定向到主页
	//http.Redirect(w, r, "http://localhost:8080/static/view/home.html", http.StatusFound)
	resp := util.RespMsg{
		Code: 0,
		Msg: "OK",
		Data: struct {
			Location string
			Username string
			Token string
		}{
			Location: "http://"+c.Request.Host+"/static/view/home.html",
			Username: username,
			Token: token,
		},
	}
	c.JSON(http.StatusOK, resp)
}

func UserInfoHandler(c *gin.Context)  {
	// 1. 解析请求参数
	username := c.Request.FormValue("username")
	//token := r.Form.Get("token")

	// 2. 验证token是否有效
	//isVaildToken := IsTokenVaild(token)
	//if !isVaildToken {
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}
	// 3. 查询用户信息
	user, err := db.GetUserInfo(username)
	if err != nil {
		log.Println(err.Error())
		c.JSON(
			http.StatusForbidden,
			gin.H{},
			)
		return
	}
	// 4. 组装并切相应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	c.Data(http.StatusOK, "octet-stream", resp.JsonToBytes())
}


func GenToken(username string) string {
	// md5(username + timestamp + token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username+ts+"token_salt"))
	return tokenPrefix + ts[:8]
}
