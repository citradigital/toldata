package test

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/citradigital/protonats"
	"github.com/stretchr/testify/assert"
	grpc "google.golang.org/grpc"
)

var grpcServer *grpc.Server
var grpcClient TestServiceClient

const serverAddr = "localhost:21001"

func startTestServer(grpcServer *grpc.Server) {
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	ctx := context.Background()
	api, err := NewTestServiceGRPC(ctx, protonats.ServiceConfiguration{URL: natsURL})
	if err != nil {
		log.Fatalln("Failed to create Protonats service")
	}

	RegisterTestServiceServer(grpcServer, api)
	log.Println("Starting GRPC server...")
	grpcServer.Serve(lis)
}

func TestGRPCInit(t *testing.T) {
	grpcServer = grpc.NewServer()

	go startTestServer(grpcServer)

	time.Sleep(time.Second * 2)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithTimeout(time.Second))

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		t.Fatal(err)
	}

	grpcClient = NewTestServiceClient(conn)
	log.Println("GRPC connected")
}

func TestGRPC1(t *testing.T) {
	req := &TestARequest{
		Input: "GRPC",
	}

	ctx := context.Background()

	bus, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)
	defer bus.Close()

	d := createTestService()

	svr := NewTestServiceProtonatsServer(bus, d)
	_, err = svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	res, err := grpcClient.GetTestA(ctx, req)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, res)
	assert.Equal(t, "OKGRPC", res.Output)
}
