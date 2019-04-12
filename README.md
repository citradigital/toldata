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
```
    var client *Bus
	client, err = NewBus(ctx, ServiceConfiguration{URL: natsURL})
	
	defer client.Close()
	client.BindClient(d)

	var resp TestAResponse
	err = client.Call(ctx, d.GetTestA, &TestARequest{Input: "OK"}, &resp)

```

### License

This software is licensed under Apache 2 license.

(c) 2019 Citra Digital Lintas

