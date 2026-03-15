package email

import "testing"

func TestDevSenderStoresMessage(t *testing.T) {
	sender := NewDevSender()
	if err := sender.Send(Message{
		To:      "user@example.com",
		Subject: "Verify email",
		Text:    "Open link",
	}); err != nil {
		t.Fatal(err)
	}
	if len(sender.Messages()) != 1 {
		t.Fatalf("expected 1 message, got %d", len(sender.Messages()))
	}
}
