package handler

import (
	"bytes"
	"encoding/json"
	"filestore-server/common"
	"filestore-server/config"
	"filestore-server/db"
	"filestore-server/meta"
	"filestore-server/mq"
	"filestore-server/store/oss"
	"filestore-server/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func UploadHandler(c *gin.Context) {
	//	处理get请求，返回上传html
	data, err := ioutil.ReadFile("/Users/chenlei/GolandProjects/src/filestore-server/static/view/index.html")
	if err != nil {
		c.String(http.StatusNotFound, "页面不存在。。。")
		return
	}
	c.Header("Content-type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(data))
}

// DoUploadHandler 上传的handler w 向用户返回参数  r 接收用户请求
func DoUploadHandler(c *gin.Context) {
	errCode := 0
	errMsg := ""
	defer func() {
		if errCode < 0 {
			c.JSON(http.StatusOK,
				util.RespMsg{
					Msg:  errMsg,
					Code: errCode,
				})
		}
	}()
	// 处理post请求，接受文件流及存储到本地目录
	file, head, err := c.Request.FormFile("file")
	util.SimplePrint(err, util.FailedGetData)
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		errCode = -1
		errMsg = "Fail to get data, err:" + err.Error()
		log.Println(errMsg)
		return
	}

	// 3. 构建文件元信息
	fileMeta := meta.FileMeta{
		FileName: head.Filename,
		FileSha1: util.Sha1(buf.Bytes()),
		FileSize: int64(len(buf.Bytes())),
		Location: "/Users/chenlei/GoProjects/transfertmp/" + head.Filename,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	fmt.Println(fileMeta.FileName, ":", fileMeta.FileSha1)
	// 创建本地文件来接收
	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		log.Printf("Create file failed, err:%s\n", err.Error())
		return
	}
	defer newFile.Close()

	nByte, err := newFile.Write(buf.Bytes())
	if int64(nByte) != fileMeta.FileSize || err != nil{
		errCode = -2
		errMsg = "Failed to save data into file, writtenSize:" + string(int64(nByte)) + "fileSize" + string(fileMeta.FileSize)
		log.Println(errMsg)
		if err != nil {
			log.Printf("err:%s\n", err.Error())
			errMsg += err.Error()
		}
		return
	}


	newFile.Seek(0, 0)
	//TODO： 写入ceph
	if config.CurrentStoreType == common.StoreCeph {

		//	写入oss
	} else if config.CurrentStoreType == common.StoreOSS {

		//data, _ := ioutil.ReadAll(newFile)
		ossPath := "oss/" + fileMeta.FileSha1

		// 是否开启异步，config.AsyncTransferEnable为true表示开启异步传输
		if !config.AsyncTransferEnable {

			err = oss.Bucket().PutObject(ossPath, newFile)
			if err != nil {
				errCode = -3
				errMsg = "oss upload_or failed! err :" + err.Error()
				log.Println(errMsg)
				return
			}
			fileMeta.Location = ossPath
		} else {
			//将 文件元信息赋值给 mq队列的消息体
			data := mq.TransferData{
				FileHash:      fileMeta.FileSha1,
				CurLocation:   fileMeta.Location,
				DestLocation:  ossPath,
				DestStoreType: common.StoreOSS,
			}
			// json序列化
			pubData, err := json.Marshal(data)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			// mq upload为生产者 向broker发布消息 等待mq消费
			sus := mq.Publish(
				config.TransExchangeName,  // exchangeName交换者姓名，生产只能向exchange发送数据不能，他们不知道数据被发送到哪里 uploadserver.trans
				config.TransOSSRoutingKey, // routing的关键词 oss
				pubData,                   //消息体
			)
			if !sus {
				// TODO: 加入重拾发送消息逻辑
			}
		}

	}

	//meta.UploadFileMeta(fileMeta)
	meta.UpdateFileMetaDB(fileMeta)
	//TODO: 更新用户文件表记录

	username := c.Request.FormValue("username")
	sus := db.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
	if sus {
		c.Redirect(http.StatusFound,"/static/view/home.html")

	} else {
		errCode = -4
		errMsg = "upload_or failed : 更新用户文件表记录失败"
	}
	//http.Redirect(w, r, "/file/upload_or/suc", http.StatusFound)
}


func UploadSucHandler(c *gin.Context) {
	c.JSON(http.StatusOK, util.RespMsg{Code: 0, Msg: "Upload finished"})
}

// GetFileMetaHandler 获取元数据
func GetFileMetaHandler(c *gin.Context) {

	filehash := c.Request.FormValue("filehash")
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError, util.RespMsg{
				Code: -1,
				Msg:  "Upload failed",
				Data: nil,
			},
		)
	}
	if fMeta != nil{
		data, err := json.Marshal(fMeta)
		if err != nil {
			c.JSON(http.StatusInternalServerError, util.RespMsg{
				Code: -2,
				Msg:  "Upload failed",
				Data: nil,
			})
			return
		}
		c.Data(http.StatusOK, "application/json", data)
	}else{
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -3,
			Msg:  "No sub file",
			Data: nil,
		})
	}

}

func FileQueryHandler(c *gin.Context) {

	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	username := c.Request.FormValue("username")

	userFiles, err := db.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, util.RespMsg{
			Code: -1,
			Msg:  "Query Failed",
			Data: nil,
		})
		return
	}
	c.JSON(http.StatusOK, util.RespMsg{
		Code: 0,
		Msg: "获取成功",
		Data: userFiles,
	})
}

// DownloadHandler 下载文件
func DownloadHandler(c *gin.Context) {
	//将url中的参数解析到form中
	fsha1 := c.Request.FormValue("filehash")
	fileMeta, _ := meta.GetFileMetaDB(fsha1)

	c.FileAttachment(fileMeta.Location, fileMeta.FileName)
}

// FileMetaUpdateHandler 更新元信息
func FileNameUpdateHandler(c *gin.Context) {
	opType := c.Request.FormValue("op")
	fileSha1 := c.Request.FormValue("filehash")
	newFileName := c.Request.FormValue("filename")
	username := c.Request.FormValue("username")

	if opType != "0" {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -1,
			Msg:  "操作错误",
			Data: nil,
		})
		return
	}

	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -2,
			Msg:  "该接口不支持这种http method：" + c.Request.Method,
			Data: nil,
		})
		return
	}
	curFileMate, err := meta.GetFileMetaDB(fileSha1)
	if err != nil {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -2,
			Msg: "获取文件信息失败，filehash：" + fileSha1,
		})
		return
	}
	tblUserFileSuc := meta.UpdateUserFileNameDB(username, newFileName, *curFileMate)
	if !tblUserFileSuc{
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -3,
			Msg:  "更新用户文件失败",
			Data: nil,
		})
		return
	}
	c.JSON(http.StatusOK, util.RespMsg{
		Code: 0,
		Msg:  "文件名修改成功",
		Data: nil,
	})
}

// FileDeleteHandler 删除文件以及元信息
func FileDeleteHandler(c *gin.Context) {
	filesha1 := c.Request.FormValue("filehash")

	fMeta, err := meta.GetFileMetaDB(filesha1)
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}
	os.Remove(fMeta.Location)

	meta.RemoveFileMeta(filesha1)

	c.Status(http.StatusOK)
}

func TryFastUploadHandle(c *gin.Context) {
	// 1. 解析请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize := c.Request.FormValue("filesize")
	// 2. 从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	// 3. 查询不到记录则返回秒传失败

	if fileMeta.FileSha1 == "" {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传",
			Data: nil,
		}
		c.Data(http.StatusOK, "application/json", resp.JsonToBytes())
		return
	}

	// 4. 上传过则将文件信息写入用户文件表，返回成功
	parsefileSize,  _ := strconv.ParseInt(filesize, 10, 64)
	sus := db.OnUserFileUploadFinished(username, filehash, filename, parsefileSize)
	if sus {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
			Data: nil,
		}
		c.Data(http.StatusOK, "application/json", resp.JsonToBytes())
		return
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败",
			Data: nil,
		}
		c.Data(http.StatusOK, "application/json", resp.JsonToBytes())
		return
	}
}

// DownloadURLHandler 生成地址接口
func DownloadURLHandler(c *gin.Context) {

	filehash := c.Request.FormValue("filehash")
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		log.Println(err.Error())
	}
	//本地
	if strings.HasPrefix(fileMeta.Location, "/Users/chenlei/GoProjects/transfertmp/"){
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		temUrl := fmt.Sprintf("http://%s/download?filehash=%s&username=%s&token=%s",
			c.Request.Host, filehash, username, token)
		c.Data(http.StatusOK, "octet-stream", []byte(temUrl))
	}else if strings.HasPrefix(fileMeta.Location, "oss/"){
		//oss 生成url
		signedURL := oss.DownloadUrl(fileMeta.Location)
		c.Data(http.StatusOK, "octest-stream", []byte(signedURL))
	}else{
		c.Data(http.StatusOK, "octet-stream", []byte("无法识别下载路径："+fileMeta.Location))
	}


}
