package handler

import (
	"account/common"
	"account/config"
	dbcli "account/db/client"
	User "account/proto"
	"account/util"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type User1 struct{}

func GenToken(username string) string {
	// md5(username + timestamp + token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username+ts+"token_salt"))
	return tokenPrefix + ts[:8]
}

// Call is a single request handler called via client.Call or the generated client code
func (u *User1)Signup(cxt context.Context, req *User.ReqSignup, resp *User.RespSignup) error{
	username := req.Username
	passwd := req.Password

	if len(username) < 3 || len(passwd) < 5 {
		resp.Code = common.StatusParamInvalid
		resp.Message = "注册参数无效"
		return nil
	}
	enc_passwd := util.Sha1([]byte(passwd+config.PassWordSalt))
	dbResp, err := dbcli.UserSignup(username, enc_passwd)
	//suc := db.UserSignup(username, enc_passwd)
	if err == nil && dbResp.Suc{
		resp.Code = common.StatusOK
		resp.Message = "注册成功"
	}else{
		resp.Code = common.StatusRegisterFailed
		resp.Message = "注册失败"
	}
	return nil
}

// Signin 登陆
func (u *User1)Signin(ctx context.Context, req *User.ReqSignin, resp *User.RespSignin) error{
	username := req.Username
	password := req.Password
	encPasswd := util.Sha1([]byte(password+config.PassWordSalt))
	// 1. 校验用户名以及密码
	dbResp, err := dbcli.UserSignin(username, encPasswd)
	//pwdChecked := db.UserSignIn(username, encPasswd)
	if err != nil || !dbResp.Suc{
		resp.Code = common.StatusLoginFailed
		resp.Message = "密码错误"
		return nil
	}

	// 2. 生成访问凭证token
	token := GenToken(username)
	log.Println("User token is :", token)
	upRes, err := dbcli.UpdateToken(username, token)
	if !upRes.Suc || err != nil{
		resp.Code = common.StatusServerError
		return nil
	}
	// 3. 登陆成功后重定向到主页
	//http.Redirect(w, r, "http://localhost:8080/static/view/home.html", http.StatusFound)
	resp.Code = common.StatusOK
	resp.Token = token
	return nil
}

// UserInfo 获取用户信息
func (u *User1)UserInfo(ctx context.Context, req *User.ReqUserInfo, resp *User.RespUserInfo) error{
	username := req.Username

	// 3. 查询用户信息
	dbResp, err := dbcli.GetUserInfo(username)
	if err != nil {
		log.Println(err.Error())
		resp.Code = common.StatusServerError
		resp.Message = "服务错误"
		return nil
	}

	if !dbResp.Suc{
		resp.Code = common.StatusUserNotExists
		resp.Message = "用户名不存在"
		return nil
	}

	user := dbcli.ToTableUser(dbResp.Data)

	// 4. 组装并切相应用户数据
	resp.Code = common.StatusOK
	resp.Username = user.Username
	resp.SignupAt = user.SignupAt
	resp.LastActiveAt = user.LastActiveAt
	resp.Status = int32(user.Status)
	log.Println("user signupat :", resp.SignupAt)
	return nil
}

// UserFiles 获取用户文件
func (u *User1)UserFiles(ctx context.Context, req *User.ReqUserFile, resp *User.RespUserFile) error{
	limitCnt := int(req.Limit)
	username := req.Username

	dbResp, err := dbcli.QueryUserFileMetas(username, limitCnt)

	if err != nil || !dbResp.Suc{
		log.Println(err.Error())
		resp.Code = common.StatusServerError
		return err
	}
	userFiles := dbcli.ToTableUserFiles(dbResp.Data)
	data, err := json.Marshal(userFiles)
	if err != nil {
		resp.Code = common.StatusServerError
		return nil
	}
	resp.FileDate = data
	return nil
}

// UserFileRename 获取用户文件（重命名）
func (u *User1)UserFileRename(cxt context.Context, req *User.ReqUserFileRename, resp *User.RespUserFileRename) error{
	dbResp, err := dbcli.RenameFileName(req.Username, req.Filehash, req.NewFileName)
	if err != nil || !dbResp.Suc{
		resp.Code = common.StatusServerError
		return err
	}

	userFiles := dbcli.ToTableUserFiles(dbResp.Data)
	data, err := json.Marshal(userFiles)
	if err != nil {
		resp.Code = common.StatusServerError
		return err
	}

	resp.FileData = data

	return nil
}
