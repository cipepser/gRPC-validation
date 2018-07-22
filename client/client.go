package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/cipepser/gRPC-validation/user"
	"google.golang.org/grpc"
)

const (
	address = "localhost"
	port    = ":8080"
)

var (
	empty = new(pb.Empty)
)

type client struct {
}

func main() {
	conn, err := grpc.Dial(address+port, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewUserServiceClient(conn)

	u := pb.User{
		Name:  "Bob",
		Age:   24,
		Phone: "",
		Mail:  "",
	}

	_, err = c.AddUser(context.Background(), &u)
	if err != nil {
		log.Fatalf("failed to add user: %v", err)
	}

	resp, err := c.GetUsers(context.Background(), empty)
	if err != nil {
		log.Fatalf("failed to get users: %v", err)
	}
	log.Printf("users:")
	for _, u := range resp.Users {
		fmt.Println(u)
	}
}
