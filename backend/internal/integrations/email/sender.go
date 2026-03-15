package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Message struct {
	To      string
	Subject string
	Text    string
}

type Sender interface {
	Send(message Message) error
}

type DevSender struct {
	mu       sync.Mutex
	messages []Message
}

func NewDevSender() *DevSender {
	return &DevSender{}
}

func (s *DevSender) Send(message Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, message)
	return nil
}

func (s *DevSender) Messages() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Message, len(s.messages))
	copy(out, s.messages)
	return out
}

type MailjetSender struct {
	APIKey    string
	SecretKey string
	From      string
	BaseURL   string
	Client    *http.Client
}

func (s MailjetSender) Send(message Message) error {
	if s.Client == nil {
		s.Client = http.DefaultClient
	}
	if s.BaseURL == "" {
		s.BaseURL = "https://api.mailjet.com/v3.1/send"
	}
	payload := map[string]any{
		"Messages": []map[string]any{
			{
				"From": map[string]string{
					"Email": s.From,
				},
				"To": []map[string]string{
					{"Email": message.To},
				},
				"Subject":  message.Subject,
				"TextPart": message.Text,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, s.BaseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.APIKey, s.SecretKey)
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("mailjet send failed: %d", resp.StatusCode)
	}
	return nil
}
