package oss

import (
	"transfer/config"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var ossCli *oss.Client

// Client 创建oss client对象
func Client() *oss.Client {
	if ossCli != nil{
		return ossCli
	}

	ossCli, err := oss.New(config.OSSEndpoint, config.OSSAccessKeyID, config.OSSAccessKeySecret)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return ossCli
}

// Bucket 获取存储空间bucket
func Bucket () *oss.Bucket{
	cli := Client()
	if cli != nil {
		bucket, err := cli.Bucket(config.OSSBucket)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		return bucket
	}
	return nil
}

// DownloadUrl 临时授权下载文件
func DownloadUrl(objName string) string{
	signedUrl, err := Bucket().SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return signedUrl
}