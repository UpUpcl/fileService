package main

import (
	dbcli "account/db/client"
	"account/handler"
	pb "account/proto"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/logger"
)

func main() {
	// Create service
	reg := consul.NewRegistry()
	service := micro.NewService(
		micro.Name("go.micro.service.user"),
		micro.Registry(reg),
		)
	service.Init()

	dbcli.Init(service)
	err := pb.RegisterUserServiceHandler(service.Server(), new(handler.User1))
	if err != nil {
		return 
	}

	if err := service.Run(); err != nil{
		logger.Fatal(err)
	}
}
