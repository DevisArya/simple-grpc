package main

import (
	"context"
	"log"
	"net"
	"simple_mongo_grpc/cmd/config"
	"simple_mongo_grpc/cmd/service"
	productPb "simple_mongo_grpc/pb/product"

	"google.golang.org/grpc"
)

func main() {

	netListen, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen %v", err.Error())
	}

	db := config.ConnectMongoDB()
	defer func() {
		if err := db.Client().Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	grpcServer := grpc.NewServer()
	productService := service.ProductService{DB: db}
	productPb.RegisterProductServiceServer(grpcServer, &productService)

	log.Printf("Server started at %v", netListen.Addr())

	if err := grpcServer.Serve(netListen); err != nil {
		log.Fatalf("Failed to serve %v", err.Error())
	}

}
