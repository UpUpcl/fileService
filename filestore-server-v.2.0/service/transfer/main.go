package main

import (
	"transfer/config"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/v3"
	"log"
	"transfer/db/client"
	"transfer/mq"
	"transfer/process"
)

func startRPCService()  {
	rsg := consul.NewRegistry()
	service := micro.NewService(
		micro.Name("go.micro.service.transfer"),
		micro.Registry(rsg),
		)
	service.Init()
	client.Init(service)
	if err := service.Run(); err != nil{
		log.Println(err.Error())
	}
}

func startTransferService()  {

	if !config.AsyncTransferEnable{
		log.Println("异步转移文件功能目前被禁用，请检查相关配置")
	}
	log.Println("文件转移服务启动中，开始监听转移任务队列")
	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer_oss",
		process.Transfer,
	)
}

func main() {
	go startTransferService()

	startRPCService()
}
