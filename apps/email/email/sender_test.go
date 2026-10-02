package email

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crylabsorg/prostack-go-showcase/email/config"
	"github.com/stretchr/testify/require"
)

func TestMailgunSender(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v3/example.com/messages", r.URL.Path)
		user, password, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, "api", user)
		require.Equal(t, "key", password)
		require.NoError(t, r.ParseForm())
		require.Equal(t, "ada@example.com", r.Form.Get("to"))
		require.Equal(t, "Welcome, Ada!", r.Form.Get("text"))
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	sender, err := NewSender(&config.Config{Provider: "mailgun", MailgunDomain: "example.com", MailgunAPIKey: "key", From: "welcome@example.com", MailgunBaseURL: server.URL})
	require.NoError(t, err)
	sender.(*mailgunSender).client = server.Client()
	require.NoError(t, sender.Send(context.Background(), SendWelcomeInput{ID: "1", Name: "Ada", Email: "ada@example.com"}))
}
