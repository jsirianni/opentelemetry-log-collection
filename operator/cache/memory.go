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

// NewMemory returns a new Memory cache with a max size
func NewMemory(maxSize uint) *Memory {
	m := Memory{}
	m.cache = make(map[string]interface{})
	m.keys = make([]string, 0)
	m.maxSize = int(maxSize)
	return &m
}

// Memory is an in memory cache of items with a pre defined
// max size
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

// Get returns an item from the cache
func (m *Memory) Get(key string) (interface{}, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	data, ok := m.cache[key]
	return data, ok
}

// Add inserts an item into the cache, if the cache is full, the
// oldest item is removed
func (m *Memory) Add(key string, data interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.maxSize > 0 && len(m.cache) == m.maxSize {
		delete(m.cache, m.keys[0])
		m.keys = m.keys[1:]
	}

	m.cache[key] = data
	m.keys = append(m.keys, key)
}
