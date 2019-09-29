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

package toldata

import (
	"context"
	"time"

	"github.com/gogo/protobuf/proto"

	nats "github.com/nats-io/go-nats"
)

type ServiceConfiguration struct {
	URL string
	ID  string
}

type Bus struct {
	Connection    *nats.Conn
	Configuration ServiceConfiguration
	Context       context.Context
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

func (bus *Bus) Close() {
	if bus.Connection != nil {
		bus.Connection.Close()
	}
}
