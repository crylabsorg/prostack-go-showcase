package email //nolint:testpackage // Tests replace the private sender HTTP client with the TLS test client.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/crylabsorg/prostack-go-showcase/email/config"
)

func TestMailgunSender(t *testing.T) {
	t.Parallel()

	type receivedRequest struct {
		path, username, password, recipient, text string
		authenticated                             bool
		parseErr                                  error
	}
	requests := make(chan receivedRequest, 1)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, authenticated := r.BasicAuth()
		parseErr := r.ParseForm()
		requests <- receivedRequest{
			path:          r.URL.Path,
			username:      username,
			password:      password,
			authenticated: authenticated,
			parseErr:      parseErr,
			recipient:     r.Form.Get("to"),
			text:          r.Form.Get("text"),
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	sender, err := NewSender(
		&config.Config{
			Provider:       "mailgun",
			MailgunDomain:  "example.com",
			MailgunAPIKey:  "key",
			From:           "welcome@example.com",
			MailgunBaseURL: server.URL,
		},
	)
	require.NoError(t, err)
	sender.(*mailgunSender).client = server.Client()
	require.NoError(
		t,
		sender.Send(
			context.Background(),
			SendWelcomeInput{ID: "1", FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com"},
		),
	)

	request := <-requests
	require.Equal(t, "/v3/example.com/messages", request.path)
	require.True(t, request.authenticated)
	require.Equal(t, "api", request.username)
	require.Equal(t, "key", request.password)
	require.NoError(t, request.parseErr)
	require.Equal(t, "ada@example.com", request.recipient)
	require.Equal(t, "Welcome, Ada!", request.text)
}
