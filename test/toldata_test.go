// Copyright 2019 Citra Digital Lintas
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"context"
	"errors"
	fmt "fmt"
	io "io"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/citradigital/toldata"
	"github.com/stretchr/testify/assert"
)

type TestToldataService struct {
	Fixtures Fixtures
}

func (b *TestToldataService) ToldataHealthCheck(ctx context.Context, req *toldata.Empty) (*toldata.ToldataHealthCheckInfo, error) {
	ret := &toldata.ToldataHealthCheckInfo{Data: ""}
	return ret, nil
}

func (b *TestToldataService) GetTestA(ctx context.Context, req *TestARequest) (*TestAResponse, error) {
	if req.Input == "123456" {
		return nil, errors.New("test-error-1")
	}

	id := ctx.Value(string("BusID"))

	if id != nil {
		b.Fixtures.SetCounter(id.(string))
	}
	result := &TestAResponse{
		Output: "OK" + req.Input,
		Id:     req.Id,
	}
	return result, nil
}

func (b *TestToldataService) GetTestAB(ctx context.Context, req *TestARequest) (*TestAResponse, error) {
	if req.Input == "123456" {
		return nil, errors.New("test-error-1")
	}

	rand.Seed(time.Now().UnixNano())
	tx := rand.Intn(200) + 1

	time.Sleep(time.Duration(tx) * time.Millisecond)

	result := &TestAResponse{
		Output: "AB" + req.Input,
		Id:     req.Id,
	}
	return result, nil
}

func (b *TestToldataService) FeedData(stream TestService_FeedDataToldataServer) {
	var sum int64

	var data *FeedDataRequest
	var err error
	for {
		data, err = stream.Receive()
		if b.Fixtures != nil && b.Fixtures.GetValue() == "crash" {
			err = errors.New("crash")
		}

		if err != nil {
			break
		}

		sum = sum + data.Data
	}

	if b.Fixtures != nil && b.Fixtures.GetValue() == "crash2" {
		err = errors.New("crash2")
	}

	if err == io.EOF {
		err := stream.Done(&FeedDataResponse{Sum: sum})

		if err != nil {
			stream.Error(err)
		}
	} else if err != nil {
		stream.Error(err)
	}

}

func (b *TestToldataService) StreamData(req *StreamDataRequest, stream TestService_StreamDataToldataServer) error {
	// We have a set of data which will be multiplied by the req
	// and stream those numbers down to the client
	data := [10]int64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	if b.Fixtures != nil && b.Fixtures.GetData() != 0 {
		start := b.Fixtures.GetData()
		data := make([]int64, start)
		for i := range data {
			data[i] = start - int64(i)
		}
	}
	for i := range data {
		if b.Fixtures != nil && b.Fixtures.GetValue() == "crash" {
			return errors.New("crash")
		}
		err := stream.Send(&StreamDataResponse{Data: data[i] * req.Id})

		if err != nil {
			return err
		}
	}
	return nil
}

func (b *TestToldataService) StreamDataAlt1(req *StreamDataRequest, stream TestService_StreamDataAlt1ToldataServer) error {
	start := int64(req.Id)
	data := make([]int64, start)
	for i := range data {
		data[i] = start - int64(i)
	}

	for i := range data {
		err := stream.Send(&StreamDataResponse{Data: data[i]})
		if err != nil {
			return err
		}
	}
	return nil
}

func createTestService() *TestToldataService {
	test := TestToldataService{
		Fixtures: CreateFixtures(),
	}

	return &test
}

var natsURL string

func TestError1(t *testing.T) {
	ctx := context.Background()
	client, err := toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceToldataClient(client)

	_, err = svc.GetTestA(ctx, &TestARequest{Input: "123456"})

	assert.NotEqual(t, nil, err)
	assert.Equal(t, "test-error-1", err.Error())
}

func testOK1(t *testing.T, title string) {
	log.Println(title)

	ctx, cancel := context.WithCancel(context.Background())

	var client *toldata.Bus
	client, err := toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceToldataClient(client)
	resp, err := svc.GetTestAB(ctx, &TestARequest{Input: title})

	assert.Equal(t, nil, err)
	assert.Equal(t, "AB"+title, resp.Output)

	cancel()
}

func TestOK1(t *testing.T) {
	testOK1(t, "t-ok1")
}

func TestOKParallel1(t *testing.T) {
	t.Parallel()
	log.Println("Parallel 1 -----------")
	rand.Seed(time.Now().UnixNano())

	v := (rand.Intn(19) + 5)
	sum := 0
	for j := 0; j < v; j++ {
		testOK1(t, fmt.Sprintf("t-ok1-%d/%d", j, v))
		log.Println(fmt.Sprintf("t-ok1-%d/%d done", j, v))
		sum++
	}

	assert.Equal(t, v, sum)

}
func TestOKParallel2(t *testing.T) {
	t.Parallel()
	log.Println("Parallel 2 -----------")

	rand.Seed(time.Now().UnixNano())

	sum := 0
	v := (rand.Intn(19) + 5)
	for j := 0; j < v; j++ {
		testOK1(t, fmt.Sprintf("t-ok2-%d/%d", j, v))
		log.Println(fmt.Sprintf("t-ok2-%d/%d done", j, v))
		sum++
	}
	assert.Equal(t, v, sum)

}
func TestOKParallel3(t *testing.T) {
	t.Parallel()
	log.Println("Parallel 3 -----------")

	rand.Seed(time.Now().UnixNano())
	sum := 0
	v := (rand.Intn(19) + 5)
	for j := 0; j < v; j++ {
		testOK1(t, fmt.Sprintf("t-ok3-%d/%d", j, v))
		log.Println(fmt.Sprintf("t-ok3-%d/%d done", j, v))
		sum++
	}
	assert.Equal(t, v, sum)

}

/*
func TestOKLoop(t *testing.T) {
	d := createTestService()

	ctx := context.Background()
	bus, err := toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL, ID: "bus1"})
	assert.Equal(t, nil, err)
	defer bus.Close()
	svr := NewTestServiceToldataServer(bus, d)
	_, err = svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	bus2, err := toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL, ID: "bus2"})
	assert.Equal(t, nil, err)
	defer bus2.Close()
	svr2 := NewTestServiceToldataServer(bus2, d)
	_, err = svr2.SubscribeTestService()
	assert.Equal(t, nil, err)

	var client *toldata.Bus
	client, err = toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	max := 100000
	svc := NewTestServiceToldataClient(client)

	t1 := time.Now()
	for i := 0; i < max; i++ {
		resp, err := svc.GetTestA(ctx, &TestARequest{Input: "OK", Id: int64(i)})

		if err != nil {
			t.Fail()
		}
		if resp.Output != "OKOK" {
			t.Fail()
		}
		if i%10000 == 0 {
			log.Println(i)
		}
	}
	t2 := time.Now()

	dur := t2.Sub(t1).Seconds()
	log.Printf("%f reqs/sec\n", float64(max)/dur)
	assert.Equal(t, true, (d.Fixtures.GetCounter("bus1") < max))
	assert.Equal(t, true, (d.Fixtures.GetCounter("bus2") < max))

	log.Println(d.Fixtures)
}
*/

func TestClientStreamHappy(t *testing.T) {
	log.Println("ClientStreamHappy")
	ctx, cancel := context.WithCancel(context.Background())

	var client *toldata.Bus
	client, err := toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceToldataClient(client)
	stream, err := svc.FeedData(ctx)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, stream)

	for i := 0; i < 10; i++ {
		_ = stream.Send(&FeedDataRequest{
			Data: int64(i),
		})
	}

	resp, err := stream.Done()

	assert.Equal(t, int64(45), resp.Sum)
	cancel()
}

func TestClientStreamSad1(t *testing.T) {
	log.Println("ClietnStreamSad1")
	ctx, cancel := context.WithCancel(context.Background())

	var client *toldata.Bus
	client, err := toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceToldataClient(client)
	stream, err := svc.FeedData(ctx)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, stream)

	for i := 0; i < 10; i++ {
		if i == 7 {
			// simulate crash on 7th iteration
			d.Fixtures.SetValue("crash")
		}
		err = stream.Send(&FeedDataRequest{
			Data: int64(i),
		})

		if err != nil {
			assert.NotEqual(t, nil, err)
			break
		}
	}
	resp, err := stream.Done()

	assert.NotEqual(t, nil, err)

	assert.Equal(t, true, resp == nil)
	cancel()
}

func TestClientStreamSad2(t *testing.T) {
	log.Println("ClietnStreamSad2")
	ctx, cancel := context.WithCancel(context.Background())

	var client *toldata.Bus
	client, err := toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceToldataClient(client)
	stream, err := svc.FeedData(ctx)

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, stream)

	// simulate crash
	d.Fixtures.SetValue("crash2")

	for i := 0; i < 10; i++ {
		err = stream.Send(&FeedDataRequest{
			Data: int64(i),
		})

		if err != nil {
			assert.NotEqual(t, nil, err)
			break
		}
	}

	resp, err := stream.Done()

	assert.NotEqual(t, nil, err)

	assert.Equal(t, true, resp == nil)
	cancel()

}

func TestServerStreamHappy(t *testing.T) {
	log.Println("ServerStreamHappy")

	ctx, cancel := context.WithCancel(context.Background())

	var client *toldata.Bus
	client, err := toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceToldataClient(client)
	stream, err := svc.StreamData(ctx, &StreamDataRequest{
		Id: 2,
	})

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, stream)

	count := 0
	var sum int64
	for {
		// Wait for the data to be available from the stream

		data, err := stream.Receive()
		if count == 10 {
			assert.Equal(t, io.EOF, err)
		} else {
			assert.Equal(t, nil, err)
		}
		if err != nil {
			break
		}
		sum = sum + data.Data
		count++
	}

	assert.Equal(t, int64(110), sum)
	assert.Equal(t, 10, count)

	cancel()

}

func TestServerStreamSad1(t *testing.T) {
	log.Println("ServerStreamSad")

	ctx, cancel := context.WithCancel(context.Background())

	var client *toldata.Bus
	client, err := toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceToldataClient(client)
	stream, err := svc.StreamData(ctx, &StreamDataRequest{
		Id: 2,
	})

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, stream)

	count := 0
	var sum int64
	for {
		// Wait for the data to be available from the stream
		data, err := stream.Receive()
		if count == 8 {
			assert.Equal(t, errors.New("crash"), err)
		} else {
			assert.Equal(t, nil, err)
		}
		if err != nil {
			break
		}
		sum = sum + data.Data
		count++
		if count == 7 {
			d.Fixtures.SetValue("crash")
		}
	}

	assert.Equal(t, int64(104), sum)
	assert.Equal(t, 8, count)

	cancel()
}

func getsum(start int64) int64 {
	data := make([]int64, start)
	var sum int64
	for i := range data {
		data[i] = start - int64(i)
		sum = sum + data[i]
	}
	return sum
}

func testServerStreamP(t *testing.T, title string, value int) {
	log.Println(title)

	ctx, cancel := context.WithCancel(context.Background())

	var client *toldata.Bus
	client, err := toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceToldataClient(client)
	stream, err := svc.StreamDataAlt1(ctx, &StreamDataRequest{
		Id: int64(value),
	})

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, stream)

	count := 0
	var sum int64
	for {
		// Wait for the data to be available from the stream
		data, err := stream.Receive()

		if err != nil {
			if err.Error() != "EOF" {
				assert.Equal(t, nil, err)
			}
			break
		}
		sum = sum + data.Data
		count++

	}

	assert.Equal(t, getsum(int64(value)), sum)
	assert.Equal(t, value, count)

	cancel()
}

func TestServerStreamParallel1(t *testing.T) {

	t.Parallel()
	rand.Seed(time.Now().UnixNano())

	num := 10
	sum := 0
	for i := 0; i < num; i++ {
		v := (rand.Intn(50) + 20)
		testServerStreamP(t, fmt.Sprintf("t1 %d-a [%d]", i, v), v)
		sum++
	}

	assert.Equal(t, num, sum)
}

func TestServerStreamParallel2(t *testing.T) {

	t.Parallel()
	rand.Seed(time.Now().UnixNano())

	num := 10
	sum := 0
	for i := 0; i < num; i++ {
		v := (rand.Intn(50) + 20)
		testServerStreamP(t, fmt.Sprintf("t2 %d-a [%d]", i, v), v)
		sum++
	}

	assert.Equal(t, num, sum)
}

func TestServerStreamParallel3(t *testing.T) {

	t.Parallel()
	rand.Seed(time.Now().UnixNano())

	num := 10
	sum := 0
	for i := 0; i < num; i++ {
		v := (rand.Intn(50) + 20)
		testServerStreamP(t, fmt.Sprintf("t3 %d-a [%d]", i, v), v)
		sum++
	}

	assert.Equal(t, num, sum)
}

func TestServerStreamP(t *testing.T) {
	p1 := time.Now().UnixNano()
	testServerStreamP(t, "Sending 1 million records", 1000*1000)
	p2 := time.Now().UnixNano()
	diff := (p2 - p1) / 1000000000
	log.Println(fmt.Sprintf("%d secs, %d records/sec", diff, 1000000/diff))
}
