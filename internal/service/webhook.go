package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type WebhookPayload struct {
	Event   string `json:"event"`
	Message string `json:"message"`
	SentAt  string `json:"sent_at"`
}

func SendWebhook(payload WebhookPayload) error {
	payload.SentAt = time.Now().Format(time.RFC3339)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", os.Getenv("WEBHOOK_URL"), bytes.NewBuffer(jsonData))
	if err != nil {
		logrus.Info("1")
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Auth-service")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		logrus.Info("2")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook return status %d", resp.StatusCode)
	}

	return nil
}
