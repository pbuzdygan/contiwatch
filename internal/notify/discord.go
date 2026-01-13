package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type DiscordClient struct {
	WebhookURL string
	Client     *http.Client
}

type discordEmbed struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Color       int    `json:"color,omitempty"`
}

type discordPayload struct {
	Content string         `json:"content,omitempty"`
	Embeds  []discordEmbed `json:"embeds,omitempty"`
}

func NewDiscordClient(webhookURL string) *DiscordClient {
	return &DiscordClient{
		WebhookURL: webhookURL,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *DiscordClient) Send(message string) error {
	if d == nil || d.WebhookURL == "" {
		return errors.New("discord webhook url is empty")
	}
	payload := discordPayload{Content: message}
	return d.post(payload)
}

func (d *DiscordClient) SendEmbed(title, description string, color int) error {
	if d == nil || d.WebhookURL == "" {
		return errors.New("discord webhook url is empty")
	}
	payload := discordPayload{
		Embeds: []discordEmbed{
			{Title: title, Description: description, Color: color},
		},
	}
	return d.post(payload)
}

func (d *DiscordClient) post(payload discordPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, d.WebhookURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("discord notification failed: non-2xx response")
	}
	return nil
}
