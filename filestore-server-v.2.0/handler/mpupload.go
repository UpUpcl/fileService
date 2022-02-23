package handler

import (
	"filestore-server/cache/redis"
	"filestore-server/db"
	"filestore-server/util"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// 初始化分块上传
type MultipartUploadInfo struct {
	FileHash string
	FileSize int
	UploadID string
	ChunkSize int
	ChunkCount int
}

func InitialMultipartUploadHandler(c *gin.Context)  {
	//1. 解析用户请求信息
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize, err := strconv.Atoi(c.Request.FormValue("filesize"))
	if err != nil {
		c.JSON(http.StatusOK, util.RespMsg{
			Code: -1,
			Msg:  "params invalid",
			Data: nil,
		})
		return
	}
	//2. 获得redis的连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//3. 生成分块上传的初始化信息
	upInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username+fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5*1024*1024, // 5M
		ChunkCount: int(math.Ceil(float64(filesize)/(5*1024*1024))),
	}
	log.Printf("filesize:%d  chunksize:%d result:%d \n", filesize, upInfo.ChunkSize, upInfo.ChunkCount)
	//4. 将初始化信息写入到redis缓存
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filehash", upInfo.FileHash)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filesize", upInfo.FileSize)

	//5. 将响应初始化数据返回到客户端
	c.JSON(http.StatusOK, util.RespMsg{0, "OK", upInfo})
}

func UploadPartHandler(c *gin.Context)  {
	// 解析用户参数
	//username := r.Form.Get("username")
	uploadID := c.Request.FormValue("uploadid")
	chunkIndex := c.Request.FormValue("index")

	// 获得redis连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 获得文件句柄，用户存储分块内容
	fpath := "/Users/chenlei/filestore_temp/"+uploadID+"/"+chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		c.JSON(http.StatusOK, util.RespMsg{-1, "Upload part failed", nil})
		return
	}
	defer fd.Close()
	buf := make([]byte, 1024*1024)
	for{
		n, err := c.Request.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}
	// 更新redis缓存状态
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	// 返回处理结果到客户端
	c.JSON(http.StatusOK, util.RespMsg{0, "ok", nil})
}

func CompleteUploadHandler(c *gin.Context)  {
	// 解析请求参数

	upid := c.Request.FormValue("uploadid")
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize, _ := strconv.Atoi(c.Request.FormValue("filesize"))
	filename := c.Request.FormValue("filename")

	// 获得redis连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 通过uploadid查询redis并判断是否所有分块上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+upid))
	if err != nil {
		c.JSON(http.StatusOK, util.RespMsg{-1, "complete upload_or error",nil})
		return
	}
	totalCount := 0
	chunkCount := 0
	fmt.Println("从redis中读取到的数据个数为", len(data))
	for i:=0; i<len(data); i+=2{
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount"{
			totalCount, _ = strconv.Atoi(v)
		}else if strings.HasPrefix(k, "chkidx_") && v == "1"{
			chunkCount += 1
		}
	}
	// bug totalCount

	if totalCount != chunkCount {
		fmt.Println("total, chunk", totalCount, chunkCount)
		c.JSON(http.StatusOK, util.RespMsg{-2, "invaild request", nil})
		return
	}
	// 合并分块
	fpath := util.GetCurrentFielParentPath() + "/tmp/" + upid + "/"

	redultFile := fpath + filename

	fil, err := os.OpenFile(redultFile, os.O_CREATE | os.O_WRONLY | os.O_APPEND, os.ModePerm)

	if err != nil {
		panic(err)
		return
	}
	for i := 0; i <= chunkCount; i++ {
		fname := fpath + strconv.Itoa(i)
		f, err := os.OpenFile(fname, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Println("打开文件"+fname+"失败"+err.Error())
		}
		bytes, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("读取文件"+fname+"失败"+err.Error())
		}
		fil.Write(bytes)
		f.Close()
	}

	for i:=1;i<=chunkCount;i++{
		fname := fpath + strconv.Itoa(i)
		err := os.Remove(fname)
		if err != nil {
			log.Printf("分块文件 %s 删除失败 %s \n", fname, err.Error())
		}
	}
	defer fil.Close()
	// 更新唯一文件表及用户文件表
	db.OnFileUploadFinished(filehash, filename, int64(filesize), "")
	db.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))

	// 响应处理结果
	c.JSON(http.StatusOK, util.RespMsg{Code: 0, Msg: "ok"})
}