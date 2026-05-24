package main

import (
	"context"
	"log"
	"os"
	"strings"

	"example.com/inga/internal/config"
	"example.com/inga/internal/notification/emails"
	"github.com/sethvargo/go-envconfig"
)

func main() {
	ctx := context.Background()

	cfg := config.Config{}
	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatal(err)
	}

	to := os.Getenv("EMAIL_TO")
	if to == "" {
		log.Fatal("EMAIL_TO is required")
	}

	emailer, err := emails.New(cfg.Email)
	if err != nil {
		log.Fatal(err)
	}

	recipients := splitRecipients(to)
	if len(recipients) == 0 {
		log.Fatal("EMAIL_TO does not contain any valid recipient addresses")
	}

	if err := emailer.SendTemplate(ctx, emails.Message{
		To:       recipients,
		Subject:  "Welcome to Inga",
		Template: "welcome.html",
		Data: map[string]any{
			"Email":   recipients[0],
			"AppName": "Inga",
		},
	}); err != nil {
		log.Fatal(err)
	}

	log.Println("email sent successfully")
}

func splitRecipients(raw string) []string {
	parts := strings.Split(raw, ",")
	recipients := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			recipients = append(recipients, part)
		}
	}
	return recipients
}
