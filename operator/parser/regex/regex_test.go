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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/open-telemetry/opentelemetry-log-collection/entry"
	"github.com/open-telemetry/opentelemetry-log-collection/operator"
	"github.com/open-telemetry/opentelemetry-log-collection/operator/helper"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
)

func newTestParser(t *testing.T, regex string) *RegexParser {
	cfg := NewRegexParserConfig("test")
	cfg.Regex = regex
	op, err := cfg.Build(testutil.Logger(t))
	require.NoError(t, err)
	return op.(*RegexParser)
}

func TestRegexParserBuildFailure(t *testing.T) {
	cfg := NewRegexParserConfig("test")
	cfg.OnError = "invalid_on_error"
	_, err := cfg.Build(testutil.Logger(t))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid `on_error` field")
}

func TestRegexParserByteFailure(t *testing.T) {
	parser := newTestParser(t, "^(?P<key>test)")
	_, err := parser.parse([]byte("invalid"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "type '[]uint8' cannot be parsed as regex")
}

func TestRegexParserStringFailure(t *testing.T) {
	parser := newTestParser(t, "^(?P<key>test)")
	_, err := parser.parse("invalid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "regex pattern does not match")
}

func TestRegexParserInvalidType(t *testing.T) {
	parser := newTestParser(t, "^(?P<key>test)")
	_, err := parser.parse([]int{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "type '[]int' cannot be parsed as regex")
}

func TestParserRegex(t *testing.T) {
	cases := []struct {
		name       string
		configure  func(*RegexParserConfig)
		inputBody  interface{}
		outputBody interface{}
	}{
		{
			"RootString",
			func(p *RegexParserConfig) {
				p.Regex = "a=(?P<a>.*)"
			},
			"a=b",
			map[string]interface{}{
				"a": "b",
			},
		},
		{
			"MemeoryCache",
			func(p *RegexParserConfig) {
				p.Regex = "a=(?P<a>.*)"
				p.CacheType = "memory"
			},
			"a=b",
			map[string]interface{}{
				"a": "b",
			},
		},
		{
			"K8sFileCache",
			func(p *RegexParserConfig) {
				p.Regex = `^(?P<pod_name>[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*)_(?P<namespace>[^_]+)_(?P<container_name>.+)-(?P<container_id>[a-z0-9]{64})\.log$`
				p.CacheType = "memory"
			},
			"coredns-5644d7b6d9-mzngq_kube-system_coredns-901f7510281180a402936c92f5bc0f3557f5a21ccb5a4591c5bf98f3ddbffdd6.log",
			map[string]interface{}{
				"container_id":   "901f7510281180a402936c92f5bc0f3557f5a21ccb5a4591c5bf98f3ddbffdd6",
				"container_name": "coredns",
				"namespace":      "kube-system",
				"pod_name":       "coredns-5644d7b6d9-mzngq",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewRegexParserConfig("test")
			cfg.OutputIDs = []string{"fake"}
			tc.configure(cfg)

			op, err := cfg.Build(testutil.Logger(t))
			require.NoError(t, err)

			fake := testutil.NewFakeOutput(t)
			op.SetOutputs([]operator.Operator{fake})

			entry := entry.New()
			entry.Body = tc.inputBody
			err = op.Process(context.Background(), entry)
			require.NoError(t, err)

			fake.ExpectBody(t, tc.outputBody)

			regexOp := op.(*RegexParser)
			cacheKey := tc.inputBody.(string)
			cacheOut, ok := regexOp.Get(cacheKey)
			require.True(t, ok, "expected cache key to exist")
			require.Equal(t, entry, cacheOut)
		})
	}
}

func TestBuildParserRegex(t *testing.T) {
	newBasicRegexParser := func() *RegexParserConfig {
		cfg := NewRegexParserConfig("test")
		cfg.OutputIDs = []string{"test"}
		cfg.Regex = "(?P<all>.*)"
		return cfg
	}

	t.Run("BasicConfig", func(t *testing.T) {
		c := newBasicRegexParser()
		_, err := c.Build(testutil.Logger(t))
		require.NoError(t, err)
	})

	t.Run("MissingRegexField", func(t *testing.T) {
		c := newBasicRegexParser()
		c.Regex = ""
		_, err := c.Build(testutil.Logger(t))
		require.Error(t, err)
	})

	t.Run("InvalidRegexField", func(t *testing.T) {
		c := newBasicRegexParser()
		c.Regex = "())()"
		_, err := c.Build(testutil.Logger(t))
		require.Error(t, err)
	})

	t.Run("NoNamedGroups", func(t *testing.T) {
		c := newBasicRegexParser()
		c.Regex = ".*"
		_, err := c.Build(testutil.Logger(t))
		require.Error(t, err)
		require.Contains(t, err.Error(), "no named capture groups")
	})

	t.Run("NoNamedGroups", func(t *testing.T) {
		c := newBasicRegexParser()
		c.Regex = "(.*)"
		_, err := c.Build(testutil.Logger(t))
		require.Error(t, err)
		require.Contains(t, err.Error(), "no named capture groups")
	})
}

func TestRegexParserConfig(t *testing.T) {
	expect := NewRegexParserConfig("test")
	expect.Regex = "test123"
	expect.ParseFrom = entry.NewBodyField("from")
	expect.ParseTo = entry.NewBodyField("to")

	t.Run("mapstructure", func(t *testing.T) {
		input := map[string]interface{}{
			"id":         "test",
			"type":       "regex_parser",
			"regex":      "test123",
			"parse_from": "$.from",
			"parse_to":   "$.to",
			"on_error":   "send",
		}
		var actual RegexParserConfig
		err := helper.UnmarshalMapstructure(input, &actual)
		require.NoError(t, err)
		require.Equal(t, expect, &actual)
	})

	t.Run("yaml", func(t *testing.T) {
		input := `
type: regex_parser
id: test
on_error: "send"
regex: "test123"
parse_from: $.from
parse_to: $.to`
		var actual RegexParserConfig
		err := yaml.Unmarshal([]byte(input), &actual)
		require.NoError(t, err)
		require.Equal(t, expect, &actual)
	})
}
