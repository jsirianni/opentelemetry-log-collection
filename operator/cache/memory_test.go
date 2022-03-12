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
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMemory(t *testing.T) {
	m := NewMemory(11)
	require.NotNil(t, m.cache)
	require.NotNil(t, m.keys)
	require.Equal(t, 11, m.maxSize)
}

func TestMemory(t *testing.T) {
	cases := []struct {
		name   string
		cache  *Memory
		input  map[string]interface{}
		expect *Memory
	}{
		{
			"basic",
			func() *Memory {
				return NewMemory(3)
			}(),
			map[string]interface{}{
				"key": "value",
				"map-value": map[string]string{
					"x":   "y",
					"dev": "stanza",
				},
			},
			&Memory{
				cache: map[string]interface{}{
					"key": "value",
					"map-value": map[string]string{
						"x":   "y",
						"dev": "stanza",
					},
				},
				maxSize: 3,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for key, value := range tc.input {
				tc.cache.Add(key, value)
				out, ok := tc.cache.Get(key)
				require.True(t, ok, "expected to get value from cache immediately after adding it")
				require.Equal(t, value, out, "expected value to equal the value that was added to the cache")
			}

			require.Equal(t, tc.expect.maxSize, tc.cache.maxSize)
			require.Equal(t, len(tc.expect.cache), len(tc.cache.cache))

			for expectKey, expectItem := range tc.expect.cache {
				actual, ok := tc.cache.Get(expectKey)
				require.True(t, ok)
				require.Equal(t, expectItem, actual)
			}
		})
	}
}

// A full cache should replace the oldest element with the new element
func TestCleanupLast(t *testing.T) {
	maxSize := 10

	m := NewMemory(uint(maxSize))

	// Add to cache until it is full
	for i := 0; i <= m.maxSize; i++ {
		str := strconv.Itoa(i)
		m.Add(str, i)
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
	expectKeys := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	require.Equal(t, expectCache, m.cache)
	require.Equal(t, expectKeys, m.keys)
	require.Len(t, m.cache, maxSize)
	require.Len(t, m.keys, maxSize)

	// for every additional key, the oldest should be removed
	// 1, 2, 3 and so on.
	for i := 11; i <= 20; i++ {
		str := strconv.Itoa(i)
		m.Add(str, i)

		removedKey := strconv.Itoa(i - 10)
		_, ok := m.Get(removedKey)
		require.False(t, ok, "expected key %s to have been removed", removedKey)
		require.Len(t, m.cache, maxSize)
		require.Len(t, m.keys, maxSize)
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
	expectKeys = []string{"11", "12", "13", "14", "15", "16", "17", "18", "19", "20"}
	require.Equal(t, expectCache, m.cache)
	require.Equal(t, expectKeys, m.keys)
	require.Len(t, m.cache, maxSize)
	require.Len(t, m.keys, maxSize)
}
