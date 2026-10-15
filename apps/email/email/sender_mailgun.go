package email

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/crylabsorg/prostack-go/logger"
	"go.uber.org/zap"

	"github.com/crylabsorg/prostack-go-showcase/email/config"
)

const mailgunErrorBodyLimit = 1024

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
	return &mailgunSender{
		client:   &http.Client{Timeout: deliveryTimeout},
		endpoint: strings.TrimRight(base, "/") + "/v3/" + url.PathEscape(c.MailgunDomain) + "/messages",
		apiKey:   c.MailgunAPIKey,
		from:     c.From,
	}, nil
}

func (s *mailgunSender) Send(ctx context.Context, w SendWelcomeInput) error {
	greeting := "Welcome!"
	if w.FirstName != "" {
		greeting = fmt.Sprintf("Welcome, %s!", w.FirstName)
	}
	form := url.Values{
		"from":    {s.from},
		"to":      {w.Email},
		"subject": {"Welcome"},
		"text":    {greeting},
	}
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
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.GetLogger(ctx).Warn("close mailgun response body", zap.Error(closeErr))
		}
	}()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, mailgunErrorBodyLimit))
		if readErr != nil {
			return fmt.Errorf("read mailgun status %d response: %w", resp.StatusCode, readErr)
		}
		return fmt.Errorf("mailgun status %d: %s", resp.StatusCode, body)
	}
	return nil
}
