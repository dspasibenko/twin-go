// Copyright 2023 The acquirecloud Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package chans

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func BenchmarkOpened(b *testing.B) {
	ch := make(chan struct{})
	for n := 0; n < b.N; n++ {
		IsOpened(ch)
	}
}

func BenchmarkClosed(b *testing.B) {
	ch := make(chan struct{})
	close(ch)
	for n := 0; n < b.N; n++ {
		IsOpened(ch)
	}
}

func TestIsOpened(t *testing.T) {
	var ch chan int
	assert.False(t, IsOpened(ch))
	ch = make(chan int)
	assert.True(t, IsOpened(ch))
	close(ch)
	assert.False(t, IsOpened(ch))
}
