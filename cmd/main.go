package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nakamorg/calbridge/pkg/caldav"
	"github.com/nakamorg/calbridge/pkg/email"
	"github.com/nakamorg/calbridge/pkg/util"
)

func main() {
	ctx := context.Background()
	if err := util.LoadDotEnv(""); err != nil {
		log.Fatal(err)
	}
	client, err := caldav.NewClient(
		os.Getenv("CALDAV_USER"),
		os.Getenv("CALDAV_PASSWORD"),
		os.Getenv("CALDAV_URL"),
	)
	if err != nil {
		log.Fatalf("Failed to create caldav client: %v", err)
	}
	mailClient, err := email.NewSMTPClient(
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASSWORD"),
		os.Getenv("SMTP_HOST"),
		"587",
	)
	if err != nil {
		log.Fatalf("Failed to create smtp client: %v", err)
	}
	defer mailClient.Close()

	events, err := client.GetEvents(ctx, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 1, 0))
	if err != nil {
		log.Fatalf("Failed to read future events: %v", err)
	}

	// Print the retrieved events
	fmt.Printf("Found %d future events\n", len(events))
	for _, event := range events {
		if err := mailClient.SendCalendarInvite(event); err != nil {
			log.Fatalf("Failed to invite: %v", err)
		}
	}
	invites, err := email.ReadCalendarInvites(
		fmt.Sprintf("%s:993", os.Getenv("IMAP_HOST")),
		os.Getenv("IMAP_USER"),
		os.Getenv("IMAP_PASSWORD"),
		3,
	)
	if err != nil {
		log.Fatalf("Failed to read mails: %v", err)
	}
	if err := client.PutEvents(ctx, invites); err != nil {
		fmt.Printf("err adding invites: %v\n", err)
	}
}
