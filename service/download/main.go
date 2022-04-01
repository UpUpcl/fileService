package main

import (
	"download/config"
	db "download/db/client"
	"download/handler"
	pb "download/proto"
	"download/router"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/v3"
	"log"
)

func StartRPCService()  {

	// Create service
	reg := consul.NewRegistry()
	service:= micro.NewService(
		micro.Name("go.micro.service.download"),
		micro.Registry(reg),
	)
	service.Init()
	db.Init(service)
	pb.RegisterDownloadHandler(service.Server(), new(handler.Download1))
	if err := service.Run(); err != nil{
		log.Println(err)
	}
}

func startAPIService()  {
	router := router.Router()
	router.Run(config.DownloadServiceHost)
}

func main() {
	go startAPIService()
	StartRPCService()
}
