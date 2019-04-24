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

package protonats

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestService struct {
	Fixtures Fixtures
}

func (b *TestService) BusNameSpace() string {
	return "TestService"
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
	bus, err := NewBus(ctx, ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)
	defer bus.Close()
	bus.BindService(d)

	var client *Bus
	client, err = NewBus(ctx, ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()
	client.BindClient(d)

	var resp TestAResponse
	err = client.Call(ctx, d.GetTestA, &TestARequest{Input: "123456"}, &resp)

	assert.NotEqual(t, nil, err)
	assert.Equal(t, "test-error-1", err.Error())
}

func TestOK1(t *testing.T) {
	d := createTestService()

	ctx := context.Background()
	bus, err := NewBus(ctx, ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)
	defer bus.Close()
	bus.BindService(d)

	var client *Bus
	client, err = NewBus(ctx, ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()
	client.BindClient(d)

	var resp TestAResponse
	err = client.Call(ctx, d.GetTestA, &TestARequest{Input: "OK"}, &resp)

	log.Println(err)
	assert.Equal(t, nil, err)
	assert.Equal(t, "OK", resp.Output)
}

func TestOKLoop(t *testing.T) {
	d := createTestService()

	ctx := context.Background()
	bus, err := NewBus(ctx, ServiceConfiguration{URL: natsURL, ID: "bus1"})
	assert.Equal(t, nil, err)
	defer bus.Close()
	bus.BindService(d)

	bus2, err := NewBus(ctx, ServiceConfiguration{URL: natsURL, ID: "bus2"})
	assert.Equal(t, nil, err)
	defer bus2.Close()
	bus2.BindService(d)

	var client *Bus
	client, err = NewBus(ctx, ServiceConfiguration{URL: natsURL})
	assert.Equal(t, nil, err)

	defer client.Close()
	client.BindClient(d)

	t1 := time.Now()
	max := 100000
	for i := 0; i < max; i++ {
		var resp TestAResponse
		err = client.Call(ctx, d.GetTestA, &TestARequest{Input: "OK"}, &resp)

		assert.Equal(t, nil, err)
		assert.Equal(t, "OK", resp.Output)
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
