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

package dynatraceexporter

import (
	"path/filepath"
	"testing"

	"github.com/dynatrace-oss/dynatrace-metric-utils-go/metric/apiconstants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtest"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	dtconfig "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/dynatraceexporter/config"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry"
)

// Test that the factory creates the default configuration
func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	assert.Equal(t, &dtconfig.Config{
		ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
		RetrySettings:    exporterhelper.NewDefaultRetrySettings(),
		QueueSettings:    exporterhelper.NewDefaultQueueSettings(),
		ResourceToTelemetrySettings: resourcetotelemetry.Settings{
			Enabled: false,
		},

		Tags:              []string{},
		DefaultDimensions: make(map[string]string),
	}, cfg, "failed to create default config")

	assert.NoError(t, configtest.CheckConfigStruct(cfg))
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yml"))
	require.NoError(t, err)

	tests := []struct {
		id           config.ComponentID
		expected     config.Exporter
		errorMessage string
	}{
		{
			id: config.NewComponentIDWithName(typeStr, "defaults"),
			expected: &dtconfig.Config{
				ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
				RetrySettings:    exporterhelper.NewDefaultRetrySettings(),
				QueueSettings:    exporterhelper.NewDefaultQueueSettings(),

				HTTPClientSettings: confighttp.HTTPClientSettings{
					Endpoint: apiconstants.GetDefaultOneAgentEndpoint(),
					Headers: map[string]string{
						"Content-Type": "text/plain; charset=UTF-8",
						"User-Agent":   "opentelemetry-collector"},
				},
				Tags:              []string{},
				DefaultDimensions: make(map[string]string),
			},
		},
		{
			id: config.NewComponentIDWithName(typeStr, "valid"),
			expected: &dtconfig.Config{
				ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
				RetrySettings:    exporterhelper.NewDefaultRetrySettings(),
				QueueSettings:    exporterhelper.NewDefaultQueueSettings(),

				HTTPClientSettings: confighttp.HTTPClientSettings{
					Endpoint: "http://example.com/api/v2/metrics/ingest",
					Headers: map[string]string{
						"Authorization": "Api-Token token",
						"Content-Type":  "text/plain; charset=UTF-8",
						"User-Agent":    "opentelemetry-collector"},
				},
				APIToken: "token",

				Prefix: "myprefix",

				Tags: []string{},
				DefaultDimensions: map[string]string{
					"dimension_example": "dimension_value",
				},
			},
		},
		{
			id: config.NewComponentIDWithName(typeStr, "valid_tags"),
			expected: &dtconfig.Config{
				ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
				RetrySettings:    exporterhelper.NewDefaultRetrySettings(),
				QueueSettings:    exporterhelper.NewDefaultQueueSettings(),

				HTTPClientSettings: confighttp.HTTPClientSettings{
					Endpoint: "http://example.com/api/v2/metrics/ingest",
					Headers: map[string]string{
						"Authorization": "Api-Token token",
						"Content-Type":  "text/plain; charset=UTF-8",
						"User-Agent":    "opentelemetry-collector"},
				},
				APIToken: "token",

				Prefix: "myprefix",

				Tags:              []string{"tag_example=tag_value"},
				DefaultDimensions: make(map[string]string),
			},
		},
		{
			id:           config.NewComponentIDWithName(typeStr, "bad_endpoint"),
			errorMessage: "endpoint must start with https:// or http://",
		},
		{
			id:           config.NewComponentIDWithName(typeStr, "missing_token"),
			errorMessage: "api_token is required if Endpoint is provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, config.UnmarshalExporter(sub, cfg))

			if tt.expected == nil {
				assert.EqualError(t, cfg.Validate(), tt.errorMessage)
				return
			}

			assert.NoError(t, cfg.Validate())
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
