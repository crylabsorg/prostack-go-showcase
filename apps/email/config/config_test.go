package config

import (
	"testing"

	"github.com/crylabsorg/prostack-go/testutils"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Chdir(testutils.ProjectRoot())
	c, err := New()
	require.NoError(t, err)
	require.Equal(t, "fake", c.Provider)
	require.Equal(t, ":33002", c.HTTPListenAddr)
	require.Equal(t, "nats", c.EventBroker)
}

func TestKafkaConfig(t *testing.T) {
	t.Chdir(testutils.ProjectRoot())
	t.Setenv("EVENT_BROKER", "kafka")
	t.Setenv("KAFKA_BROKERS", "localhost:9092,localhost:9094")
	c, err := New()
	require.NoError(t, err)
	require.Equal(t, "kafka", c.EventBroker)
	require.Equal(t, []string{"localhost:9092", "localhost:9094"}, c.KafkaBrokers)
	t.Setenv("EVENT_BROKER", "unknown")
	_, err = New()
	require.Error(t, err)
}
