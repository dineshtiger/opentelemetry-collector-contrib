// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kafkaexporter

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	tests := []struct {
		id       config.ComponentID
		expected config.Exporter
	}{
		{
			id: config.NewComponentIDWithName(typeStr, ""),
			expected: &Config{
				ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
				TimeoutSettings: exporterhelper.TimeoutSettings{
					Timeout: 10 * time.Second,
				},
				RetrySettings: exporterhelper.RetrySettings{
					Enabled:         true,
					InitialInterval: 10 * time.Second,
					MaxInterval:     1 * time.Minute,
					MaxElapsedTime:  10 * time.Minute,
				},
				QueueSettings: exporterhelper.QueueSettings{
					Enabled:      true,
					NumConsumers: 2,
					QueueSize:    10,
				},
				Topic:    "spans",
				Encoding: "otlp_proto",
				Brokers:  []string{"foo:123", "bar:456"},
				Authentication: Authentication{
					PlainText: &PlainTextConfig{
						Username: "jdoe",
						Password: "pass",
					},
				},
				Metadata: Metadata{
					Full: false,
					Retry: MetadataRetry{
						Max:     15,
						Backoff: defaultMetadataRetryBackoff,
					},
				},
				Producer: Producer{
					MaxMessageBytes: 10000000,
					RequiredAcks:    sarama.WaitForAll,
					Compression:     "none",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, config.UnmarshalExporter(sub, cfg))

			assert.NoError(t, cfg.Validate())
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestValidate_err_compression(t *testing.T) {
	config := &Config{
		Producer: Producer{
			Compression: "idk",
		},
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "producer.compression should be one of 'none', 'gzip', 'snappy', 'lz4', or 'zstd'. configured value idk")
}

func Test_saramaProducerCompressionCodec(t *testing.T) {
	tests := map[string]struct {
		compression         string
		expectedCompression sarama.CompressionCodec
		expectedError       error
	}{
		"none": {
			compression:         "none",
			expectedCompression: sarama.CompressionNone,
			expectedError:       nil,
		},
		"gzip": {
			compression:         "gzip",
			expectedCompression: sarama.CompressionGZIP,
			expectedError:       nil,
		},
		"snappy": {
			compression:         "snappy",
			expectedCompression: sarama.CompressionSnappy,
			expectedError:       nil,
		},
		"lz4": {
			compression:         "lz4",
			expectedCompression: sarama.CompressionLZ4,
			expectedError:       nil,
		},
		"zstd": {
			compression:         "zstd",
			expectedCompression: sarama.CompressionZSTD,
			expectedError:       nil,
		},
		"unknown": {
			compression:         "unknown",
			expectedCompression: sarama.CompressionNone,
			expectedError:       fmt.Errorf("producer.compression should be one of 'none', 'gzip', 'snappy', 'lz4', or 'zstd'. configured value unknown"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			c, err := saramaProducerCompressionCodec(test.compression)
			assert.Equal(t, c, test.expectedCompression)
			assert.Equal(t, err, test.expectedError)
		})
	}
}
