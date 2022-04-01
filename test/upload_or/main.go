package main

import (
	"filestore-server/config"
	"filestore-server/route"
	"fmt"
)

func main() {

	router := route.Router()
	router.Run(config.UploadServiceHost)

	// 开启监听127.0.0.1:8080端口
	fmt.Printf("上传服务启动中，开始监听[%s]...\n", config.UploadServiceHost)
}
