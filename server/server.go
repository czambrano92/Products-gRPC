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

	"go.mongodb.org/mongo-driver/bson"
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

func (*server) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductResponse, error) {

	productId := req.GetProductId()
	oid, err := primitive.ObjectIDFromHex(productId)
	if err != nil {

		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("cannot parse ID %v", productId),
		)
	}

	//create empty struct
	data := &product{}
	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)

	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find the product: %v", err),
		)
	}

	return &productpb.GetProductResponse{
		Product: dbToProductPb(data),
	}, nil
}

func dbToProductPb(data *product) *productpb.Product {
	return &productpb.Product{
		Id:    data.ID.Hex(),
		Name:  data.Name,
		Price: data.Price,
	}
}

func (*server) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.UpdateProductResponse, error) {
	fmt.Println("Update product request")

	prod := req.GetProduct()

	oid, err := primitive.ObjectIDFromHex(prod.GetId())

	if err != nil {

		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("cannot parse ID %v", prod),
		)
	}

	//create empty struc

	data := &product{}
	filter := bson.M{"_id": oid}

	//search the product in db
	res := collection.FindOne(context.Background(), filter)
	if res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find the product with the id %v", err),
		)
	}
	//update the internal struct product
	data.Name = prod.GetName()
	data.Price = prod.GetPrice()
	//update in db
	_, updateError := collection.ReplaceOne(context.Background(), filter, data)
	if updateError != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot update the product %v", updateError),
		)
	}

	return &productpb.UpdateProductResponse{
		Product: dbToProductPb(data),
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
