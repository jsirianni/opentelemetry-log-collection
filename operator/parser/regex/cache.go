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

package regex

import "sync"

// CacheConfig is a configuration for caching
type CacheConfig struct {
	CacheType    string `json:"type" yaml:"type"`
	CacheMaxSize uint16 `json:"size" yaml:"size"`
}

// cache allows operators to cache a value and look it up later
type cache interface {
	Get(key string) (interface{}, bool)
	Add(key string, data interface{})
	Copy() map[string]interface{}
	MaxSize() uint16
}

// Default max size for Memory cache
const defaultMemoryCacheMaxSize uint16 = 100

// newMemoryCache returns a new Memory cache with a max size.
// Uses defaultMemoryCacheMaxSize if maxSize is 0
func newMemoryCache(maxSize uint16) *memoryCache {
	if maxSize < 1 {
		maxSize = defaultMemoryCacheMaxSize
	}

	return &memoryCache{
		cache: make(map[string]interface{}),
		keys:  make(chan string, maxSize),
	}
}

// memoryCache is an in memory cache of items with a pre defined
// max size. Memory's underlying storage is a map[string]item
// and does not perform any manipulation of the data. Memory
// is designed to be as fast as possible while being thread safe.
// When the cache is full, new items will evict the oldest
// item using a FIFO style queue.
type memoryCache struct {
	// Key / Value pairs of cached items
	cache map[string]interface{}

	// When the cache is full, the oldest entry's key is
	// read from the channel and used to index into the
	// cache during cleanup
	keys chan string

	// All read options will trigger a read lock while all
	// write options will trigger a lock
	mutex sync.RWMutex
}

var _ cache = (&memoryCache{})

// Get returns an item from the cache and should be treated
// the same as indexing a map
func (m *memoryCache) Get(key string) (interface{}, bool) {
	// Read and unlock as fast as possible
	m.mutex.RLock()
	data, ok := m.cache[key]
	m.mutex.RUnlock()

	return data, ok
}

// Add inserts an item into the cache, if the cache is full, the
// oldest item is removed
func (m *memoryCache) Add(key string, data interface{}) {
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
func (m *memoryCache) Copy() map[string]interface{} {
	copy := make(map[string]interface{}, cap(m.keys))

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for k, v := range m.cache {
		copy[k] = v
	}
	return copy
}

// MaxSize returns the max size of the cache
func (m *memoryCache) MaxSize() uint16 {
	return uint16(cap(m.keys))
}