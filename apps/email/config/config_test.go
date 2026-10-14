package config_test

import (
	"os"
	"testing"

	"github.com/crylabsorg/prostack-go/testutils"
	"github.com/stretchr/testify/require"

	"github.com/crylabsorg/prostack-go-showcase/email/config"
)

func TestConfig(t *testing.T) {
	t.Chdir(testutils.ProjectRoot())
	t.Setenv("NATS_URL", "")
	t.Setenv("KAFKA_BROKERS", "")
	c, err := config.New()
	require.NoError(t, err)
	require.Equal(t, "fake", c.Provider)
	require.Equal(t, ":33002", c.HTTPListenAddr)
	require.Equal(t, "nats://localhost:4222", c.NATSURL)
	require.Empty(t, c.KafkaBrokers)
}

func TestKafkaConfig(t *testing.T) {
	t.Chdir(testutils.ProjectRoot())
	t.Setenv("NATS_URL", "")
	t.Setenv("KAFKA_BROKERS", "localhost:9092,localhost:9094")
	c, err := config.New()
	require.NoError(t, err)
	require.Equal(t, "nats://localhost:4222", c.NATSURL)
	require.Equal(t, []string{"localhost:9092", "localhost:9094"}, c.KafkaBrokers)
}

func TestBrokerConfig(t *testing.T) {
	testCases := []struct {
		name         string
		brokerYAML   string
		natsURL      string
		kafkaBrokers []string
		wantError    bool
	}{
		{
			name:       "nats only",
			brokerYAML: "NATS_URL: nats://localhost:4222\n",
			natsURL:    "nats://localhost:4222",
		},
		{
			name:         "nats with empty kafka list",
			brokerYAML:   "NATS_URL: nats://localhost:4222\nKAFKA_BROKERS: []\n",
			natsURL:      "nats://localhost:4222",
			kafkaBrokers: []string{},
		},
		{
			name:         "kafka only",
			brokerYAML:   "NATS_URL: \"\"\nKAFKA_BROKERS: [localhost:9092]\n",
			kafkaBrokers: []string{"localhost:9092"},
		},
		{
			name:         "both brokers",
			brokerYAML:   "NATS_URL: nats://localhost:4222\nKAFKA_BROKERS: [localhost:9092]\n",
			natsURL:      "nats://localhost:4222",
			kafkaBrokers: []string{"localhost:9092"},
		},
		{
			name:      "no brokers",
			wantError: true,
		},
		{
			name:       "empty brokers",
			brokerYAML: "NATS_URL: \"\"\nKAFKA_BROKERS: []\n",
			wantError:  true,
		},
		{
			name:       "empty kafka address",
			brokerYAML: "KAFKA_BROKERS: [\"\"]\n",
			wantError:  true,
		},
		{
			name:       "nats with empty kafka address",
			brokerYAML: "NATS_URL: nats://localhost:4222\nKAFKA_BROKERS: [\"\"]\n",
			wantError:  true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("NATS_URL", "")
			t.Setenv("KAFKA_BROKERS", "")
			configYAML := "LOG_PRESET: development\nSERVICE_NAME: email\nHTTP_LISTEN_ADDR: :33002\nEMAIL_PROVIDER: fake\n" + testCase.brokerYAML
			require.NoError(t, os.WriteFile("config.yaml", []byte(configYAML), 0o600))

			cfg, err := config.New()
			if testCase.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, testCase.natsURL, cfg.NATSURL)
			require.Equal(t, testCase.kafkaBrokers, cfg.KafkaBrokers)
		})
	}
}
