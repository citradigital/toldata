package test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/darmawan01/toldata"
	"github.com/joho/godotenv"
)

var d *TestToldataService

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

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
	done, _ := getab.SubscribeTestService()

	code := m.Run()

	cancel()
	<-done
	os.Exit(code)
}
