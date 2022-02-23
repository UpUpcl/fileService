package main

import (
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/logger"
	"os"
	Uploadconfig "upload/config"
	dbcli "upload/db/client"
	"upload/handler"
	Upload "upload/proto"
	"upload/router"
)

func startRPCService()  {
	srv := micro.NewService(
		micro.Name("go.micro.service.upload"),
		micro.Registry(consul.NewRegistry()),
	)

	srv.Init()
	// Register handler
	dbcli.Init(srv)
	Upload.RegisterUploadServiceHandler(srv.Server(), new(handler.Upload1))
	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}

func main() {
	os.MkdirAll(Uploadconfig.TempLocalRootDir,0777)
	os.MkdirAll(Uploadconfig.TempPartRootDir, 0777)
	// Create service
	go startRPCService()

	router := router.Router()
	router.Run(Uploadconfig.UploadServiceHost)
}
