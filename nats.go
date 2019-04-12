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
	"reflect"
	"time"

	"github.com/gogo/protobuf/proto"

	nats "github.com/nats-io/go-nats"
)

type BusInterface interface {
}

type ServiceConfiguration struct {
	URL string
	ID  string
}

type Bus struct {
	Connection     *nats.Conn
	Configuration  ServiceConfiguration
	Context        context.Context
	Service        BusInterface
	serviceType    reflect.Type
	serviceMap     map[string]reflect.Method
	serviceNameMap map[string]string
}

func NewBus(ctx context.Context, config ServiceConfiguration) (*Bus, error) {
	k := string("BusID")

	busID := config.ID
	if busID == "" {
		busID = "BUS"
	}

	s := &Bus{
		Configuration: config,
		Context:       context.WithValue(ctx, k, busID),
	}

	err := s.initConnection()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (bus *Bus) BindService(service BusInterface) {
	bus.Service = service
	v := reflect.ValueOf(service)

	bus.serviceMap = make(map[string]reflect.Method)
	for i := 0; i < v.NumMethod(); i++ {
		method := v.Type().Method(i)
		valid := method.Func.IsNil() == false &&
			method.Func.IsValid() == true &&
			method.Type.NumIn() == 3

		if valid {
			bus.subscribe(method.Name)
			bus.serviceMap[method.Name] = method
		}
	}
}

func (bus *Bus) BindClient(service BusInterface) {
	bus.Service = service
	v := reflect.ValueOf(service)

	bus.serviceNameMap = make(map[string]string)
	for i := 0; i < v.NumMethod(); i++ {
		method := v.Type().Method(i)
		valid := method.Func.IsNil() == false &&
			method.Func.IsValid() == true &&
			method.Type.NumIn() == 3

		if valid {
			id := ""
			for j := 1; j < 2; j++ {
				id = id + method.Type.In(j).String()
			}
			bus.serviceNameMap[id] = method.Name
		}
	}
}

func (bus *Bus) initConnection() error {
	nc, err := nats.Connect(bus.Configuration.URL)
	if err != nil {
		return err
	}
	bus.Connection = nc

	return nil
}

func (bus *Bus) HandleError(replySubject string, err error) {
	if replySubject == "" {
		return
	}

	now := time.Now()
	data, errx := proto.Marshal(&ErrorMessage{
		ErrorMessage: err.Error(),
		Timestamp:    now.UnixNano(),
		BusID:        bus.Configuration.ID,
	})

	if errx == nil {
		one := []byte{1}
		bus.Connection.Publish(replySubject, append(one, data...))
	}
}

func (bus *Bus) subscribe(subject string) error {
	_, err := bus.Connection.QueueSubscribe(subject, "q", func(m *nats.Msg) {
		if bus.Service == nil {
			// error
			return
		}

		method := bus.serviceMap[subject]
		args := make([]reflect.Value, 3)
		args[0] = reflect.ValueOf(bus.Service)
		args[1] = reflect.ValueOf(bus.Context)

		data := reflect.New(method.Type.In(2).Elem())
		err := proto.Unmarshal(m.Data, data.Interface().(proto.Message))

		if err != nil {
			bus.HandleError(m.Reply, err)
			return
		}
		args[2] = data
		result := method.Func.Call(args)

		if m.Reply != "" {
			if len(result) == 2 {
				if result[1].IsNil() == false {
					err := result[1].Interface().(error)

					bus.HandleError(m.Reply, err)
				} else {
					data := result[0].Interface().(proto.Message)
					raw, err := proto.Marshal(data)
					if err != nil {
						bus.HandleError(m.Reply, err)
					} else {
						zero := []byte{0}
						bus.Connection.Publish(m.Reply, append(zero, raw...))
					}
				}
			}
		}
	})

	if err != nil {
		return err
	}

	return nil
}

func (bus *Bus) Close() {
	if bus.Connection != nil {
		bus.Connection.Close()
	}
}

func (bus *Bus) Call(ctx context.Context, fn interface{}, req proto.Message, resp interface{}) error {

	t := reflect.ValueOf(fn).Type()
	valid := t.NumIn() == 2 && t.NumOut() == 2 && bus.Service != nil

	if valid == false {
		return errors.New("invalid-function")
	}

	id := ""
	for i := 0; i < 1; i++ {
		id = id + t.In(i).String()
	}
	functionName, ok := bus.serviceNameMap[id]

	if ok == false {
		return errors.New("function-not-found")
	}

	reqRaw, err := proto.Marshal(req)

	result, err := bus.Connection.RequestWithContext(ctx, functionName, reqRaw)
	if err != nil {
		return err
	}

	if result.Data[0] == 0 {
		p, ok := resp.(proto.Message)
		if ok == false {
			return errors.New("invalid-response-type")
		}
		err = proto.Unmarshal(result.Data[1:], p)
	} else {
		var pErr ErrorMessage
		err = proto.Unmarshal(result.Data[1:], &pErr)

		if err == nil {
			return errors.New(pErr.ErrorMessage)
		}
	}

	return err

}
