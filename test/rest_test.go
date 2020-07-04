package test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/citradigital/toldata"
	"github.com/stretchr/testify/assert"
)

const serverAddrREST = "localhost:21002"

func startRESTTestServer(s *http.Server) {

	log.Println("Starting REST server...")
	log.Fatal(http.ListenAndServe(serverAddrREST, nil))
}

func TestRESTInit(t *testing.T) {
	ctx := context.Background()
	api, err := NewTestServiceREST(ctx, toldata.ServiceConfiguration{URL: natsURL})
	if err != nil {
		log.Fatalln("Failed to create Toldata service")
	}

	mux := http.NewServeMux()
	api.InstallTestServiceMux(mux)
	s := &http.Server{
		Addr:           serverAddrREST,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	local := &TestToldataService{}
	api.Service.SetBuslessObject(local)

	go s.ListenAndServe()

	time.Sleep(time.Second * 2)

	log.Println("REST connected")
}

func TestREST1(t *testing.T) {
	req := &TestARequest{
		Input: "REST",
	}

	jsonPayload, err := json.Marshal(req)
	assert.Equal(t, nil, err)

	url := "http://" + serverAddrREST + "/api/test/cdl.toldatatest/TestService/GetTestA"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	assert.Equal(t, nil, err)
	defer httpResp.Body.Close()

	var resp TestAResponse
	err = json.NewDecoder(httpResp.Body).Decode(&resp)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, resp)
	assert.Equal(t, "OKREST", resp.Output)
}

func TestREST2(t *testing.T) {
	req := &TestARequest{
		Input: "REST",
		Id:    199,
	}

	jsonPayload, err := json.Marshal(req)
	assert.Equal(t, nil, err)

	url := "http://" + serverAddrREST + "/api/test/cdl.toldatatest/TestService/GetTestAB"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	assert.Equal(t, nil, err)
	defer httpResp.Body.Close()

	var resp TestAResponse
	err = json.NewDecoder(httpResp.Body).Decode(&resp)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, resp)
	assert.Equal(t, "ABREST", resp.Output)
	assert.Equal(t, int64(199), resp.Id)
}

func TestRESTError(t *testing.T) {
	req := &TestARequest{
		Input: "123456",
		Id:    999,
	}

	jsonPayload, err := json.Marshal(req)
	assert.Equal(t, nil, err)

	url := "http://" + serverAddrREST + "/api/test/cdl.toldatatest/TestService/GetTestAB"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	httpResp, err := client.Do(httpReq)
	assert.Equal(t, nil, err)
	defer httpResp.Body.Close()

	assert.Equal(t, 500, httpResp.StatusCode)

	var errResp toldata.ErrorMessage
	err = json.NewDecoder(httpResp.Body).Decode(&errResp)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, errResp)
	assert.Equal(t, "test-error-1", errResp.ErrorMessage)
}
