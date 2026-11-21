// Package notify posts alert messages to Slack and/or DingTalk webhooks.
// Webhook URLs are read from the environment; if neither is set, Send is a
// no-op (the caller still prints findings to stdout).
package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Send posts title+text to every configured webhook (Slack and/or DingTalk).
// It returns the last error encountered; it is not an error if no webhook is
// configured.
func Send(title, text string) error {
	var lastErr error

	if url := os.Getenv("SLACK_WEBHOOK_URL"); url != "" {
		if err := postSlack(url, title, text); err != nil {
			lastErr = err
		}
	}

	if url := os.Getenv("DINGTALK_WEBHOOK_URL"); url != "" {
		if err := postDingTalk(url, title, text); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

func postSlack(url, title, text string) error {
	payload := map[string]any{
		"text": fmt.Sprintf("*%s*\n%s", title, text),
	}
	return postJSON(url, payload)
}

func postDingTalk(url, title, text string) error {
	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  fmt.Sprintf("### %s\n%s", title, text),
		},
	}
	return postJSON(url, payload)
}

func postJSON(url string, payload any) error {
	buf, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned %s", resp.Status)
	}
	return nil
}
