package main

import (
	"fmt"
	crobat "github.com/Cgboal/SonarSearch/proto"
	"go.elastic.co/apm/module/apmhttp"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
)

func main() {
	fmt.Println("Staring Crobat API... You get it? Because Sonar? Sonar, Bats?")
	server := NewServer()
	go http.ListenAndServe(":1338", apmhttp.Wrap(server.Router))

	lis, err := net.Listen("tcp", ":1997")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	crobatServer := NewRPCServer()
	crobat.RegisterCrobatServer(grpcServer, &crobatServer)
	grpcServer.Serve(lis)

}
