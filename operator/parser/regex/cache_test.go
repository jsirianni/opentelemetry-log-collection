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
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestnewMemoryCache(t *testing.T) {
	cases := []struct {
		name       string
		maxSize    uint16
		expect     *memoryCache
		expectSize int
	}{
		{
			"size-50",
			50,
			&memoryCache{
				cache: make(map[string]interface{}),
				keys:  make(chan string, 50),
			},
			50,
		},
	}

	for _, tc := range cases {
		output := newMemoryCache(tc.maxSize)
		require.Equal(t, tc.expect.cache, output.cache)
		require.Len(t, output.cache, 0, "new memory should always be empty")
		require.Len(t, output.keys, 0, "new memory should always be empty")
		require.Equal(t, tc.expectSize, cap(output.keys), "keys channel should have cap of expected size")
	}
}

func TestMemory(t *testing.T) {
	cases := []struct {
		name   string
		cache  *memoryCache
		input  map[string]interface{}
		expect *memoryCache
	}{
		{
			"basic",
			func() *memoryCache {
				return newMemoryCache(3)
			}(),
			map[string]interface{}{
				"key": "value",
				"map-value": map[string]string{
					"x":   "y",
					"dev": "stanza",
				},
			},
			&memoryCache{
				cache: map[string]interface{}{
					"key": "value",
					"map-value": map[string]string{
						"x":   "y",
						"dev": "stanza",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for key, value := range tc.input {
				tc.cache.add(key, value)
				out := tc.cache.get(key)
				require.NotNil(t, out, "expected to get value from cache immediately after adding it")
				require.Equal(t, value, out, "expected value to equal the value that was added to the cache")
			}

			require.Equal(t, len(tc.expect.cache), len(tc.cache.cache))

			for expectKey, expectItem := range tc.expect.cache {
				actual := tc.cache.get(expectKey)
				require.NotNil(t, actual)
				require.Equal(t, expectItem, actual)
			}
		})
	}
}

// A full cache should replace the oldest element with the new element
func TestCleanupLast(t *testing.T) {
	maxSize := 10

	m := newMemoryCache(uint16(maxSize))

	// Add to cache until it is full
	for i := 0; i <= cap(m.keys); i++ {
		str := strconv.Itoa(i)
		m.add(str, i)
	}

	// make sure the cache looks the way we expect
	expectCache := map[string]interface{}{
		"1":  1, // oldest key, will be removed when 11 is added
		"2":  2,
		"3":  3,
		"4":  4,
		"5":  5,
		"6":  6,
		"7":  7,
		"8":  8,
		"9":  9,
		"10": 10, // youngest key, will be removed when 20 is added
	}
	require.Equal(t, expectCache, m.cache)
	require.Len(t, m.cache, maxSize)
	require.Len(t, m.keys, maxSize)

	// for every additional key, the oldest should be removed
	// 1, 2, 3 and so on.
	for i := 11; i <= 20; i++ {
		str := strconv.Itoa(i)
		m.add(str, i)

		removedKey := strconv.Itoa(i - 10)
		x := m.get(removedKey)
		require.Nil(t, x, "expected key %s to have been removed", removedKey)
		require.Len(t, m.cache, maxSize)
	}

	// All entries should have been replaced by now
	expectCache = map[string]interface{}{
		"11": 11,
		"12": 12,
		"13": 13,
		"14": 14,
		"15": 15,
		"16": 16,
		"17": 17,
		"18": 18,
		"19": 19,
		"20": 20,
	}
	require.Equal(t, expectCache, m.cache)
	require.Len(t, m.cache, maxSize)
}
