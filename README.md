# Protobuf for Nats.io

This Go module simplifies how you use protobuf with Nats.io:
1. Provide proto files
2. Provide Go implementations
3. Bind the implementation on the server side
4. Make a call from the clients

### Example
Proto:
```
message TestARequest {
    string input = 1; 
}

message TestAResponse {
    string output = 1;
}

service TestService {
    rpc GetTestA(TestARequest) returns (TestAResponse) {}
}
```

Implementation:
```
func (b *TestService) GetTestA(ctx context.Context, req *TestARequest) (*TestAResponse, error) {
	...
    some business
    ...
    if err != nil {
		return nil, err
	}

	result := &TestAResponse{
		Output: "OK",
	}
	return result, nil
}
```

Server side:
```
    service := &TestService{}
    ...
	ctx := context.Background()
	bus, err := NewBus(ctx, ServiceConfiguration{URL: natsURL, ID: "bus1"})
	defer bus.Close()
	bus.BindService(service)
```

Client side:

Generate the code from the proto file with `--protonats_out=` argument to protoc-gogo, then:

```
    var client *Bus
	client, err = NewBus(ctx, ServiceConfiguration{URL: natsURL})
	
	defer client.Close()

	svc := NewTestServiceClient(client)
	resp, err := svc.GetTestA(ctx, &TestARequest{Input: "OK"})

```

### License

This software is licensed under Apache 2 license.

(c) 2019 Citra Digital Lintas

