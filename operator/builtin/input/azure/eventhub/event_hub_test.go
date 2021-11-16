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

package eventhub

import (
	"testing"

	"github.com/open-telemetry/opentelemetry-log-collection/operator/builtin/input/azure"
	"github.com/open-telemetry/opentelemetry-log-collection/testutil"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	cases := []struct {
		name      string
		input     EventHubInputConfig
		expectErr bool
	}{
		{
			"default",
			EventHubInputConfig{
				AzureConfig: azure.AzureConfig{
					Namespace:        "test",
					Name:             "test",
					Group:            "test",
					ConnectionString: "test",
					PrefetchCount:    1000,
				},
			},
			false,
		},
		{
			"prefetch",
			EventHubInputConfig{
				AzureConfig: azure.AzureConfig{
					Namespace:        "test",
					Name:             "test",
					Group:            "test",
					ConnectionString: "test",
					PrefetchCount:    100,
				},
			},
			false,
		},
		{
			"startat-end",
			EventHubInputConfig{
				AzureConfig: azure.AzureConfig{
					Namespace:        "test",
					Name:             "test",
					Group:            "test",
					ConnectionString: "test",
					StartAt:          "end",
					PrefetchCount:    1000,
				},
			},
			false,
		},
		{
			"startat-beginning",
			EventHubInputConfig{
				AzureConfig: azure.AzureConfig{
					Namespace:        "test",
					Name:             "test",
					Group:            "test",
					ConnectionString: "test",
					StartAt:          "beginning",
					PrefetchCount:    1000,
				},
			},
			false,
		},
		{
			"prefetch-invalid",
			EventHubInputConfig{
				AzureConfig: azure.AzureConfig{
					Namespace:        "test",
					Name:             "test",
					Group:            "test",
					ConnectionString: "test",
					PrefetchCount:    0,
				},
			},
			true,
		},
		{
			"default-required-startat-invalid",
			EventHubInputConfig{
				AzureConfig: azure.AzureConfig{
					Namespace:        "test",
					Name:             "test",
					Group:            "test",
					ConnectionString: "test",
					StartAt:          "invalid",
					PrefetchCount:    1000,
				},
			},
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewEventHubConfig("test_id")
			cfg.Namespace = tc.input.Namespace
			cfg.Name = tc.input.Name
			cfg.Group = tc.input.Group
			cfg.ConnectionString = tc.input.ConnectionString

			if tc.input.PrefetchCount != NewEventHubConfig("").PrefetchCount {
				cfg.PrefetchCount = tc.input.PrefetchCount
			}

			if tc.input.StartAt != "" {
				cfg.StartAt = tc.input.StartAt
			}

			_, err := cfg.Build(testutil.NewBuildContext(t))
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}