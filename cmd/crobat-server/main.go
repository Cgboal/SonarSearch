package main

import (
	"github.com/spf13/viper"
	"github.com/cgboal/sonarsearch/cmd/crobat-server/rest"
	"net"
	crobat "github.com/cgboal/sonarsearch/proto"
	"log"
	"google.golang.org/grpc"
	cgrpc "github.com/cgboal/sonarsearch/cmd/crobat-server/grpc"
)

func init() {
	viper.SetEnvPrefix("crobat")
	viper.AutomaticEnv()
}


func main() {
	restRouter := rest.NewRouter()

	go restRouter.Run(":1998")

	lis, err := net.Listen("tcp", ":1997")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	crobatServer := cgrpc.CrobatServer{}
	crobat.RegisterCrobatServer(grpcServer, &crobatServer)
	grpcServer.Serve(lis)
}

