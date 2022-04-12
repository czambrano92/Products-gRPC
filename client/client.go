package main

import (
	"context"
	"fmt"
	"log"
	"product/productpb"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Go Client is running")

	opts := grpc.WithInsecure()

	cc, err := grpc.Dial("localhost:50051", opts)

	if err != nil {
		log.Fatal("Failed to connect %v", err)
	}

	defer cc.Close()

	c := productpb.NewProductServiceClient(cc)

	product := &productpb.Product{
		Name:  "Smartphone YY",
		Price: 25000.05,
	}

	createdProduct, err := c.CreateProduct(context.Background(), &productpb.CreateProductRequest{
		Product: product,
	})

	if err != nil {
		log.Fatalf("Failed to create product %v", err)
	}
	fmt.Printf("Product created %v", createdProduct)
}
