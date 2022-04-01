package main

import (
	mysql "dbproxy/conn"
	dbproxy "dbproxy/proto"
	"dbproxy/rpc"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/logger"
)

func main() {
	// Create service
	rsg := consul.NewRegistry()
	srv := micro.NewService(
		micro.Name("go.micro.service.dbproxy"),
		micro.Registry(rsg),
	)

	// Register handler
	srv.Init()
	mysql.InitDBConn()

	dbproxy.RegisterDBProxyServiceHandler(srv.Server(), new(rpc.DBProxy1))

	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
