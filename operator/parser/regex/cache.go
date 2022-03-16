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

import (
	"sync"
	"sync/atomic"
	"time"
)

// CacheConfig is a configuration for caching
type CacheConfig struct {
	CacheType    string `json:"type" yaml:"type"`
	CacheMaxSize uint16 `json:"size" yaml:"size"`
}

// cache allows operators to cache a value and look it up later
type cache interface {
	get(key string) (interface{}, bool)
	add(key string, data interface{})
	copy() map[string]interface{}
	maxSize() uint16
}

// Default max size for Memory cache
const defaultMemoryCacheMaxSize uint16 = 100

// newMemoryCache returns a new memory backed cache
func newMemoryCache(maxSize uint16) *memoryCache {
	if maxSize < 1 {
		maxSize = defaultMemoryCacheMaxSize
	}

	// rate limiter will throttle when cache is above
	// 100% turnover within the limit inteval
	limitCount := uint64(maxSize) + 1
	limitInterval := time.Second * 5
	limiter := newAtomicLimiter(limitCount, limitInterval)

	m := &memoryCache{
		cache:   make(map[string]interface{}),
		keys:    make(chan string, maxSize),
		limiter: limiter,
	}

	// start the rate limiter and return the cache
	m.limiter.start()
	return m
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

	// Limiter rate limits the cache
	limiter limiter
}

var _ cache = (&memoryCache{})

// get returns an item from the cache and should be treated
// the same as indexing a map
func (m *memoryCache) get(key string) (interface{}, bool) {
	// Read and unlock as fast as possible
	m.mutex.RLock()
	data, ok := m.cache[key]
	m.mutex.RUnlock()

	return data, ok
}

// add inserts an item into the cache, if the cache is full, the
// oldest item is removed
func (m *memoryCache) add(key string, data interface{}) {
	if m.limiter.throttled() {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.keys) == cap(m.keys) {
		// Pop the oldest key from the channel
		// and remove it from the cache
		delete(m.cache, <-m.keys)

		// notify the rate limiter that an entry
		// was evicted
		m.limiter.increment()
	}

	// Write the cached entry and add the key
	// to the channel
	m.cache[key] = data
	m.keys <- key
}

// copy returns a deep copy of the cache
func (m *memoryCache) copy() map[string]interface{} {
	copy := make(map[string]interface{}, cap(m.keys))

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for k, v := range m.cache {
		copy[k] = v
	}
	return copy
}

// maxSize returns the max size of the cache
func (m *memoryCache) maxSize() uint16 {
	return uint16(cap(m.keys))
}

type limiter interface {
	start()
	increment()
	throttled() bool
}

func newAtomicLimiter(max uint64, interval time.Duration) *atomicLimiter {
	return &atomicLimiter{
		count:    0,
		max:      max,
		interval: interval,
	}
}

// atomicLimiter enables rate limiting using an atomic
// counter
type atomicLimiter struct {
	count    uint64
	max      uint64
	interval time.Duration
}

var _ limiter = (&atomicLimiter{})

// start runs the reset go routine
func (l *atomicLimiter) start() {
	go func() {
		ticker := time.NewTicker(l.interval)
		for _ = range ticker.C {
			atomic.SwapUint64(&l.count, 0)
		}
	}()
}

// increment resets the limiter on an interval and starts
// the cleanup go routine on the first run
func (l *atomicLimiter) increment() {
	atomic.AddUint64(&l.count, 1)
}

// Returns true if the cache is currently throttled, meaning a high
// number of evictions have recently happened due to the cache being
// full. When the cache is contantly being locked for writes, reads
// are blocked, causing the regex parser to be slower than if it was
// not caching at all.
func (l *atomicLimiter) throttled() bool {
	return l.count >= l.max
}
