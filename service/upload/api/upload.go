package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"upload/common"
	config "upload/config"
	dbcli "upload/db/client"
	"upload/db/orm"
	"upload/mq"
	"upload/oss"
	"upload/util"
)

func DoUploadHandler(c *gin.Context) {
	errCode := 0
	defer func() {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GEt")
		if errCode < 0 {
			c.JSON(http.StatusOK,
				gin.H{
					"msg":  "上传失败",
					"code": errCode,
				})
		}else{
			c.JSON(http.StatusOK, gin.H{
				"msg":"上传成功",
				"code":errCode,
			})
		}
	}()
	// 处理post请求，接受文件流及存储到本地目录
	file, head, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Failed to get form data, err:%s\n", err.Error())
		errCode = -1
		return
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		errCode = -2
		log.Println(err.Error())
		return
	}

	// 3. 构建文件元信息
	fileMeta := dbcli.FileMeta{
		FileName: head.Filename,
		FileSha1: util.Sha1(buf.Bytes()),
		FileSize: int64(len(buf.Bytes())),
		Location: config.TempLocalRootDir + head.Filename,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	log.Println(fileMeta.FileName, ":", fileMeta.FileSha1)
	// 创建本地文件来接收
	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		log.Printf("Create file failed, err:%s\n", err.Error())
		errCode = -3
		return
	}
	defer newFile.Close()

	nByte, err := newFile.Write(buf.Bytes())
	if int64(nByte) != fileMeta.FileSize || err != nil{
		errCode = -4
		log.Println(err.Error())
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
				errCode = -5
				log.Println(err.Error())
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

	_, err = dbcli.OnFileUploadFinished(fileMeta)
	if err != nil {
		errCode = -6
		return
	}
	//TODO: 更新用户文件表记录

	username := c.Request.FormValue("username")
	upResp, err := dbcli.OnUserFileUploadFinished(username, fileMeta)
	if err == nil && upResp.Suc {
		errCode = 0
	} else {
		errCode = -6
	}
}

func TryFastUploadHandle(c *gin.Context) {
	// 1. 解析请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	//filesize := c.Request.FormValue("filesize")
	// 2. 从文件表中查询相同hash的文件记录
	fileMetaResp, err := dbcli.GetFileMeta(filehash)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	// 3. 查询不到记录则返回秒传失败

	if !fileMetaResp.Suc {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传",
			Data: nil,
		}
		c.Data(http.StatusOK, "application/json", resp.JsonToBytes())
		return
	}

	// 4. 上传过则将文件信息写入用户文件表，返回成功
	fmeta := dbcli.TableFileToFileMeta(fileMetaResp.Data.(orm.TableFile))
	fmeta.FileName = filename
	upResp, err := dbcli.OnUserFileUploadFinished(username, fmeta)
	if err == nil && upResp.Suc {
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
