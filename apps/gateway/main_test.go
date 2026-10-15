package main_test

import (
	"encoding/json/v2"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crylabsorg/prostack-go-showcase/pkg/events"
)

func TestUserCreatedContract(t *testing.T) {
	t.Parallel()

	specificationJSON, err := os.ReadFile("asyncapi.json")
	require.NoError(t, err)

	var specification struct {
		Channels map[string]struct {
			Address string `json:"address"`
		} `json:"channels"`
		Components struct {
			Schemas map[string]struct {
				Required   []string `json:"required"`
				Properties map[string]struct {
					Type string `json:"type"`
				} `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	require.NoError(t, json.Unmarshal(specificationJSON, &specification))
	require.Len(t, specification.Channels, 2)
	for _, channel := range specification.Channels {
		require.Equal(t, events.UserCreatedSubject, channel.Address)
	}

	payloadSchema := specification.Components.Schemas["UserCreated"]
	require.ElementsMatch(t, []string{"id", "email"}, payloadSchema.Required)
	for _, property := range payloadSchema.Properties {
		require.Equal(t, "string", property.Type)
	}

	for _, test := range []struct {
		name  string
		event events.UserCreatedEvent
		keys  []string
	}{
		{
			name:  "absent names",
			event: events.UserCreatedEvent{ID: "user-id", Email: "user@example.com"},
			keys:  []string{"id", "email"},
		},
		{
			name: "explicit empty names",
			event: events.UserCreatedEvent{
				ID: "user-id", Email: "user@example.com", FirstName: new(string), LastName: new(string),
			},
			keys: []string{"id", "email", "first_name", "last_name"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			payload, marshalErr := json.Marshal(test.event)
			require.NoError(t, marshalErr)
			var values map[string]string
			require.NoError(t, json.Unmarshal(payload, &values))
			keys := make([]string, 0, len(values))
			for key := range values {
				keys = append(keys, key)
				require.Contains(t, payloadSchema.Properties, key)
			}
			require.ElementsMatch(t, test.keys, keys)
			for _, required := range payloadSchema.Required {
				require.Contains(t, values, required)
			}
			if test.event.FirstName != nil {
				require.Empty(t, values["first_name"])
				require.Empty(t, values["last_name"])
				require.Len(t, payloadSchema.Properties, len(keys))
			}
		})
	}
}
