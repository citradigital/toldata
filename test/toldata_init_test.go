package test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/darmawan01/toldata"
)

var d *TestToldataService

func TestMain(m *testing.M) {
	natsURL = os.Getenv("NATS_URL")
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	d = createTestService()
	ctx, cancel := context.WithCancel(context.Background())

	log.Println("init")
	bus, err := toldata.NewBus(ctx, toldata.ServiceConfiguration{URL: natsURL})

	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()
	getab := NewTestServiceToldataServer(bus, d)
	done, err := getab.SubscribeTestService()

	code := m.Run()

	cancel()
	<-done
	os.Exit(code)
}
