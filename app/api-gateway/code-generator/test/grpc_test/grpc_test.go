package grpc_test

import (
	"context"
	pb "go-gateway/app/api-gateway/code-generator/test/grpc_test/api"
	"google.golang.org/grpc"
	"log"
	"testing"
	"time"
)

func TestGrpc(t *testing.T) {
	address := "localhost:9000"
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewDemoClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloReq{Name: "test"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("%s", r.Content)
}
