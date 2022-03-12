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

// item is a cache entry. The incoming data type is
// intended to be the input and output type passed
// to parser operators, which happens to be interface{}
type item interface{}

// Default max size for Memory cache
const DefaultMemoryMaxSize uint16 = 100

// NewMemory returns a new Memory cache with a max size.
// Uses DefaultMemoryMaxSize if maxSize is 0
func NewMemory(maxSize uint16) *Memory {
	if maxSize < 1 {
		maxSize = DefaultMemoryMaxSize
	}

	return &Memory{
		cache: make(map[string]item),
		keys:  make(chan string, maxSize),
	}
}

// Memory is an in memory cache of items with a pre defined
// max size. Memory's underlying storage is a map[string]item
// and does not perform any manipulation of the data. Memory
// is designed to be as fast as possible while being thread safe.
// When the cache is full, new items will evict the oldest
// item using a FIFO style queue.
type Memory struct {
	// Key / Value pairs of cached items
	cache map[string]item

	// When the cache is full, the oldest entry's key is
	// read from the channel and used to index into the
	// cache during cleanup
	keys chan string

	// All read options will trigger a read lock while all
	// write options will trigger a lock
	mutex sync.RWMutex
}

var _ Cache = (&Memory{})

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

	if len(m.keys) == cap(m.keys) {
		// Pop the oldest key from the channel
		// and remove it from the cache
		delete(m.cache, <-m.keys)
	}

	// Write the cached entry and push it's
	// key to the channel
	m.cache[key] = data
	m.keys <- key
}

// Copy returns a deep copy of the cache
func (m *Memory) Copy() map[string]interface{} {
	copy := make(map[string]interface{}, cap(m.keys))

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for k, v := range m.cache {
		copy[k] = v
	}
	return copy
}

// MaxSize returns the max size of the cache
func (m *Memory) MaxSize() uint16 {
	return uint16(cap(m.keys))
}
