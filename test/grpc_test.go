package test

import (
	"context"
	io "io"
	"log"
	"net"
	"testing"
	"time"

	"github.com/citradigital/toldata"
	"github.com/stretchr/testify/assert"
	grpc "google.golang.org/grpc"
	status "google.golang.org/grpc/status"
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
	api, err := NewTestServiceGRPC(ctx, toldata.ServiceConfiguration{URL: natsURL})
	if err != nil {
		log.Fatalln("Failed to create Toldata service")
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

	res, err := grpcClient.GetTestA(ctx, req)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, res)
	assert.Equal(t, "OKGRPC", res.Output)
}

func TestGRPCStreamDataHappy(t *testing.T) {
	ctx := context.Background()
	req := StreamDataRequest{Id: 2}
	stream, err := grpcClient.StreamData(ctx, &req)

	data := [10]int64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	assert.Equal(t, nil, err)
	i := 0
	for {
		ret, err := stream.Recv()
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				if st.Err() == io.EOF {
					break
				}
			}
			break
		}
		assert.Equal(t, nil, err)
		assert.Equal(t, data[i]*req.Id, ret.Data)
		i = i + 1
	}

	assert.Equal(t, 10, i)
}

func TestGRPCFeedDataHappy(t *testing.T) {
	ctx := context.Background()
	stream, err := grpcClient.FeedData(ctx)
	assert.Equal(t, nil, err)

	for i := 0; i < 10; i++ {
		_ = stream.Send(&FeedDataRequest{
			Data: int64(i),
		})
	}

	resp, err := stream.CloseAndRecv()

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, resp)
	assert.Equal(t, int64(45), resp.Sum)

}
