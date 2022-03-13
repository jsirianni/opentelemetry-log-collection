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

package helper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

func TestNewEncodingConfig(t *testing.T) {
	config := NewEncodingConfig()
	expect := EncodingConfig{
		Encoding: "utf-8",
	}
	require.Equal(t, expect, config)
}

func TestBuild(t *testing.T) {
	cases := []struct {
		name      string
		config    EncodingConfig
		expect    Encoding
		expectErr string
	}{
		{
			"default",
			EncodingConfig{""},
			Encoding{
				Encoding: unicode.UTF8,
			},
			"",
		},
		{
			"utf-8",
			EncodingConfig{"utf-8"},
			Encoding{
				Encoding: unicode.UTF8,
			},
			"",
		},
		{
			"utf8",
			EncodingConfig{"utf8"},
			Encoding{
				Encoding: unicode.UTF8,
			},
			"",
		},
		{
			"ascii",
			EncodingConfig{"ascii"},
			Encoding{
				Encoding: unicode.UTF8,
			},
			"",
		},
		{
			"us-ascii",
			EncodingConfig{"us-ascii"},
			Encoding{
				Encoding: unicode.UTF8,
			},
			"",
		},
		{
			"nop",
			EncodingConfig{"nop"},
			Encoding{
				Encoding: encoding.Nop,
			},
			"",
		},
		{
			"utf-16",
			EncodingConfig{"utf-16"},
			Encoding{
				Encoding: unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
			},
			"",
		},
		{
			"utf16",
			EncodingConfig{"utf16"},
			Encoding{
				Encoding: unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
			},
			"",
		},
		{
			// ISO-8859-1 is a valid iana charset that is not explicitly mapped
			// in this package
			"custom-ISO-8859-1",
			EncodingConfig{"ISO-8859-1"},
			Encoding{
				Encoding: unicode.UTF8BOM,
			},
			"",
		},
		{
			// ISO-8859-2 is a valid iana charset that is not explicitly mapped
			// in this package
			"custom-ISO-8859-2",
			EncodingConfig{"ISO-8859-2"},
			Encoding{
				Encoding: unicode.UTF8BOM,
			},
			"",
		},
		{
			// Invalid is not an iana charset
			"invalid",
			EncodingConfig{"invalid"},
			Encoding{
				Encoding: nil,
			},
			"unsupported encoding 'invalid'",
		},
		{
			// KS_C_5601-1987 is a valid iana charset but is not supported
			// by std lib encoding package.
			"unsupported",
			EncodingConfig{"KS_C_5601-1987"},
			Encoding{
				Encoding: nil,
			},
			"no charmap defined for encoding 'KS_C_5601-1987'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := tc.config.Build()
			if tc.expectErr != "" {
				require.Error(t, err)
				require.Equal(t, err.Error(), tc.expectErr)
				return
			}

			// Test mapped encoding configs
			for k := range encodingOverrides {
				if k == tc.name || tc.name == "default" {
					require.Equal(t, tc.expect, output)
					return
				}
			}
			// Test unmapped encoding configs, which are returned by
			// ianana as a charmap.Charmap
			require.IsType(t, &charmap.Charmap{}, output.Encoding)
		})
	}
}
