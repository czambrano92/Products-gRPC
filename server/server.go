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

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type product struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Name  string             `bson:"name"`
	Price float64            `bson:"price"`
}

type server struct{}

func (*server) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	//parse content and save to mongo
	prod := req.GetProduct()
	data := product{
		Name:  prod.GetName(),
		Price: prod.GetPrice(),
	}

	res, err := collection.InsertOne(context.Background(), data)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal Error %v", err),
		)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert OID :%v", err),
		)
	}

	return &productpb.CreateProductResponse{
		Product: &productpb.Product{
			Id:    oid.Hex(),
			Name:  prod.GetName(),
			Price: prod.GetPrice(),
		},
	}, nil
}

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Connecting to Mongo DB")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Error creating the client DB %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Error creating the client DB %v", err)
	}

	collection = client.Database("productdb").Collection("products")

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
	client.Disconnect(context.TODO())
	lis.Close()
	fmt.Println("Goodbye :D ...")
}
