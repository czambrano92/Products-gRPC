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

	//Creating Product
	fmt.Println("-------Creating Product------")
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

	//Getting Product
	fmt.Println("-------Getting Product------")

	productID := createdProduct.GetProduct().GetId()

	getProductReq := &productpb.GetProductRequest{
		ProductId: productID,
	}

	getProductRes, getProductErr := c.GetProduct(context.Background(), getProductReq)

	if getProductErr != nil {
		log.Fatalf("Failed to getting product %v", getProductErr)
	}

	fmt.Printf("Product gotten_: %v", getProductRes)

	//Updating product
	fmt.Println("-------Updating Product------")
	newProduct := &productpb.Product{
		Id:    productID,
		Name:  "New name: Smartphone XV",
		Price: 40500,
	}

	updateResponse, updateErr := c.UpdateProduct(context.Background(), &productpb.UpdateProductRequest{Product: newProduct})
	if updateErr != nil {
		fmt.Printf("Error happened while updating product %v ", updateErr)
	}

	fmt.Printf("Product was updated %v", updateResponse)

	//Delete product
	deleteRes, deleteErr := c.DeleteProduct(context.Background(), &productpb.DeleteProductRequest{
		ProductId: productID,
	})

	if deleteErr != nil {
		fmt.Printf("Error deleting the product %v", deleteErr)
	}

	fmt.Printf("product deleted %v: \n", deleteRes.GetProductId())

}
