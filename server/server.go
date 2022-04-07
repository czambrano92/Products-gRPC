package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"product/productpb"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

type server struct{}

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Connecting to Mongo DB")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://foo:bar@localhost:27017"))
	if err != nil {
		log.Fatalf("Error creating the client DB %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Error creating the client DB %v", err)
	}

	fmt.Println("Product service is running")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	s := grpc.NewServer()

	productpb.RegisterProductServiceServer(s, &server{})

	go func() {
		fmt.Println("Starting product server")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Falied to serve %v", err)
		}
	}()

	//wait for ctrl + x to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	//Block until we get the signal

	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("Goodbye :D ...")
}
