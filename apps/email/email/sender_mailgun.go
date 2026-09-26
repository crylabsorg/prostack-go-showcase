package email

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/crylabsorg/prostack-go-showcase/email/config"
)

type mailgunSender struct {
	client                 *http.Client
	endpoint, apiKey, from string
}

func NewSender(c *config.Config) (Sender, error) {
	if c.Provider == "fake" {
		return fakeSender{}, nil
	}
	base := c.MailgunBaseURL
	if base == "" {
		base = "https://api.mailgun.net"
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return nil, fmt.Errorf("invalid MAILGUN_BASE_URL: %q", base)
	}
	return &mailgunSender{client: &http.Client{Timeout: 20 * time.Second}, endpoint: strings.TrimRight(base, "/") + "/v3/" + url.PathEscape(c.MailgunDomain) + "/messages", apiKey: c.MailgunAPIKey, from: c.From}, nil
}

func (s *mailgunSender) Send(ctx context.Context, w Welcome) error {
	form := url.Values{"from": {s.from}, "to": {w.Email}, "subject": {"Welcome"}, "text": {"Welcome, " + w.Name + "!"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth("api", s.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("mailgun status %d: %s", resp.StatusCode, body)
	}
	return nil
}
