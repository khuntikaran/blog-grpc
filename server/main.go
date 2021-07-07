package main

import (
	"bloga/database"
	"bloga/proto/proto"
	"fmt"
	"log"
	"net"

	//	"project-1/proto/proto"
	//"github.com/golang/protobuf/proto"

	"google.golang.org/grpc"
)

//type Server struct{}

func main() {
	listener, err := net.Listen("tcp", ":5050")

	if err != nil {
		log.Fatal(err)
	}
	opts := []grpc.ServerOption{}
	//	s := blog.Server{}
	grpcServer := grpc.NewServer(opts...)

	proto.RegisterBlogServiceServer(grpcServer, &proto.UnimplementedBlogServiceServer{})
	fmt.Println("connecting to mongodb")

	database.ConnectDB()

	fmt.Println("connected to mongodb")

	go func() {

		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("error while serving the grpc server")
		}
	}()

	fmt.Println("Server successfully started on port 5050")

}
