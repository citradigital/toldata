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
	"log"
	"os"
	"testing"
	"time"

	"github.com/citradigital/protonats"
	"github.com/stretchr/testify/assert"
)

type TestService struct {
	Fixtures Fixtures
}

func (b *TestService) GetTestA(ctx context.Context, req *TestARequest) (*TestAResponse, error) {
	if req.Input == "123456" {
		return nil, errors.New("test-error-1")
	}

	id := ctx.Value(string("BusID"))

	if id != nil {
		b.Fixtures.SetCounter(id.(string))
	}
	result := &TestAResponse{
		Output: "OK",
		Id:     req.Id,
	}
	return result, nil
}

func createTestService() *TestService {
	test := TestService{
		Fixtures: CreateFixtures(),
	}

	return &test
}

var natsURL string

func TestInit(t *testing.T) {
	natsURL = os.Getenv("NATS_URL")
	log.SetFlags(log.Lshortfile)
}

func TestError1(t *testing.T) {
	d := createTestService()

	ctx := context.Background()
	bus, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)
	defer bus.Close()

	svr := NewTestServiceProtonatsServer(bus, d)
	_, err = svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	client, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()
	svc := NewTestServiceProtonatsClient(client)
	_, err = svc.GetTestA(ctx, &TestARequest{Input: "123456"})

	assert.NotEqual(t, nil, err)
	assert.Equal(t, "test-error-1", err.Error())
}

func TestOK1(t *testing.T) {
	d := createTestService()

	ctx, cancel := context.WithCancel(context.Background())
	bus, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)
	defer bus.Close()
	svr := NewTestServiceProtonatsServer(bus, d)
	done, err := svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	var client *protonats.Bus
	client, err = protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	svc := NewTestServiceProtonatsClient(client)
	resp, err := svc.GetTestA(ctx, &TestARequest{Input: "OK"})

	log.Println(err)
	assert.Equal(t, nil, err)
	assert.Equal(t, "OK", resp.Output)

	cancel()
	<-done
}

func TestOKLoop(t *testing.T) {
	d := createTestService()

	ctx := context.Background()
	bus, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL, ID: "bus1"})
	assert.Equal(t, nil, err)
	defer bus.Close()
	svr := NewTestServiceProtonatsServer(bus, d)
	_, err = svr.SubscribeTestService()
	assert.Equal(t, nil, err)

	bus2, err := protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL, ID: "bus2"})
	assert.Equal(t, nil, err)
	defer bus2.Close()
	svr2 := NewTestServiceProtonatsServer(bus2, d)
	_, err = svr2.SubscribeTestService()
	assert.Equal(t, nil, err)

	var client *protonats.Bus
	client, err = protonats.NewBus(ctx, protonats.ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()

	max := 100000
	svc := NewTestServiceProtonatsClient(client)

	t1 := time.Now()
	for i := 0; i < max; i++ {
		resp, err := svc.GetTestA(ctx, &TestARequest{Input: "OK", Id: int64(i)})

		if err != nil {
			t.Fail()
		}
		if resp.Output != "OK" {
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
