// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"sync"
)

// Default max size for Memory cache
const DefaultMemoryMaxSize uint16 = 100

// NewMemory returns a new Memory cache with a max size.
// Uses DefaultMemoryMaxSize if maxSize is 0
func NewMemory(maxSize uint16) *Memory {
	if maxSize < 1 {
		maxSize = DefaultMemoryMaxSize
	}

	return &Memory{
		cache:   make(map[string]interface{}),
		keys:    make([]string, 0, maxSize),
		maxSize: int(maxSize),
	}
}

// Memory is an in memory cache of items with a pre defined
// max size. Memory's underlying storage is a map[string]interface{}
// and does not perform any manipulation of the data. Memory
// is designed to be as fast as possible while being thread safe.
type Memory struct {
	// Key / Value pairs of cached items
	cache map[string]interface{}

	// Tracks a list of keys added to the cache, when the
	// cache is full, the first item in the slice is used
	// to determine which key / value pair to remove from
	// the cache because keys[0] will always be the oldest
	// item in the cache
	keys []string

	// The max number of items the cache will hold before
	// it starts to remove items. High load systems should
	// avoid setting this value too low because performance
	// degrades significantly when items are being removed
	// for every addition
	maxSize int

	// All read options will trigger a read lock while all
	// write options will trigger a lock
	mutex sync.RWMutex
}

// Get returns an item from the cache and should be treated
// the same as indexing a map
func (m *Memory) Get(key string) (interface{}, bool) {
	// Read and unlock as fast as possible
	m.mutex.RLock()
	data, ok := m.cache[key]
	m.mutex.RUnlock()

	return data, ok
}

// Add inserts an item into the cache, if the cache is full, the
// oldest item is removed
func (m *Memory) Add(key string, data interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.keys) == m.maxSize {
		delete(m.cache, m.keys[0])
		m.keys = m.keys[1:]
	}

	m.cache[key] = data
	m.keys = append(m.keys, key)
}