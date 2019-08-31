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
	"time"
)

type TestFixtures struct {
	Value   string
	Time    time.Time
	Counter map[string]int
	data    int64
}

func (f *TestFixtures) SetData(data int64) {
	f.data = data
}

func (f *TestFixtures) GetData() int64 {
	return f.data
}

func (f *TestFixtures) SetCounter(s string) {
	f.Counter[s] = f.Counter[s] + 1
}

func (f *TestFixtures) GetCounter(s string) int {
	return f.Counter[s]
}

func (f *TestFixtures) SetTime(t time.Time) {
	f.Time = t
}

func (f *TestFixtures) GetTime() time.Time {
	return f.Time
}

func (f *TestFixtures) SetValue(s string) {
	f.Value = s
}

func (f *TestFixtures) GetValue() string {
	return f.Value
}

var fixtures Fixtures

func CreateFixtures() *TestFixtures {
	return &TestFixtures{Value: "",
		Counter: make(map[string]int)}
}
