package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Alerter sends alert notifications.
type Alerter interface {
	Send(title, content string) error
}

// LogAlerter logs alerts instead of sending them (dev mode).
type LogAlerter struct{}

func NewLogAlerter() *LogAlerter {
	return &LogAlerter{}
}

func (a *LogAlerter) Send(title, content string) error {
	log.Printf("[ALERT-DEV] %s: %s", title, content)
	return nil
}

// DingTalkAlerter sends alerts to a DingTalk robot webhook.
type DingTalkAlerter struct {
	webhookURL string
}

func NewDingTalkAlerter(webhookURL string) *DingTalkAlerter {
	return &DingTalkAlerter{webhookURL: webhookURL}
}

func (a *DingTalkAlerter) Send(title, content string) error {
	msg := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  fmt.Sprintf("### %s\n\n%s\n\n> %s", title, content, time.Now().Format("2006-01-02 15:04:05")),
		},
	}
	return a.sendJSON(msg)
}

func (a *DingTalkAlerter) sendJSON(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(a.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("dingtalk webhook error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("dingtalk webhook status=%d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// SlackAlerter sends alerts to a Slack incoming webhook.
type SlackAlerter struct {
	webhookURL string
}

func NewSlackAlerter(webhookURL string) *SlackAlerter {
	return &SlackAlerter{webhookURL: webhookURL}
}

func (a *SlackAlerter) Send(title, content string) error {
	msg := map[string]interface{}{
		"text": fmt.Sprintf("*%s*\n%s\n_%s_", title, content, time.Now().Format("2006-01-02 15:04:05")),
	}
	data, _ := json.Marshal(msg)
	resp, err := http.Post(a.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("slack webhook error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack webhook status=%d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GenericWebhookAlerter sends alerts to any HTTP endpoint via POST JSON.
type GenericWebhookAlerter struct {
	webhookURL string
}

func NewGenericWebhookAlerter(webhookURL string) *GenericWebhookAlerter {
	return &GenericWebhookAlerter{webhookURL: webhookURL}
}

func (a *GenericWebhookAlerter) Send(title, content string) error {
	msg := map[string]interface{}{
		"title":     title,
		"content":   content,
		"timestamp": time.Now().Unix(),
	}
	data, _ := json.Marshal(msg)
	resp, err := http.Post(a.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("webhook error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("webhook status=%d", resp.StatusCode)
	}
	return nil
}

// NewAlerter creates an alerter based on environment configuration.
// ALERT_WEBHOOK_URL env var is used. If empty, uses log mode.
// ALERT_TYPE can be: "log" (default), "dingtalk", "slack", "webhook"
func NewAlerter() Alerter {
	webhookURL := os.Getenv("ALERT_WEBHOOK_URL")
	if webhookURL == "" {
		return NewLogAlerter()
	}

	alertType := os.Getenv("ALERT_TYPE")
	switch alertType {
	case "dingtalk":
		return NewDingTalkAlerter(webhookURL)
	case "slack":
		return NewSlackAlerter(webhookURL)
	case "webhook":
		return NewGenericWebhookAlerter(webhookURL)
	default:
		return NewLogAlerter()
	}
}
