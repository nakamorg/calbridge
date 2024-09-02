package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/emersion/go-ical"
	"github.com/nakamorg/calbridge/pkg/caldav"
	"github.com/nakamorg/calbridge/pkg/email"
	"github.com/nakamorg/calbridge/pkg/util"
)

func main() {
	ctx := context.Background()
	if err := util.LoadDotEnv(""); err != nil {
		log.Fatal(err)
	}
	calClient, err := caldav.NewClient(
		os.Getenv("CALDAV_USER"),
		os.Getenv("CALDAV_PASSWORD"),
		os.Getenv("CALDAV_URL"),
	)
	if err != nil {
		log.Fatalf("Failed to create caldav client: %v", err)
	}
	smtpClient, err := email.NewSMTPClient(
		os.Getenv("SMTP_USER"),
		os.Getenv("SMTP_PASSWORD"),
		os.Getenv("SMTP_HOST"),
		"587",
	)
	if err != nil {
		log.Fatalf("Failed to create smtp client: %v", err)
	}
	defer smtpClient.Close()

	imapClient, err := email.NewIMAPClient(
		os.Getenv("IMAP_USER"),
		os.Getenv("IMAP_PASSWORD"),
		os.Getenv("IMAP_HOST"),
		"993",
	)
	if err != nil {
		log.Fatalf("Failed to create imap client: %v", err)
	}
	defer imapClient.Close()

	if err := sendInvites(ctx, calClient, smtpClient); err != nil {
		log.Fatal(err)
	}

	if err := addInvites(ctx, calClient, imapClient); err != nil {
		log.Fatal(err)
	}

}

func sendInvites(ctx context.Context, calClient *caldav.Client, smtpClient *email.SMTPClient) error {
	var events []*ical.Calendar
	var err error

	if events, err = calClient.GetEvents(ctx, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 1, 0)); err != nil {
		return fmt.Errorf("failed reading future events: %v", err)
	}

	for _, event := range events {
		if err = smtpClient.SendCalendarInvite(event); err != nil {
			return fmt.Errorf("failed sending invitation: %v", err)
		}
	}
	return nil
}

func addInvites(ctx context.Context, calClient *caldav.Client, imapClient *email.IMAPClient) error {
	var events []*ical.Calendar
	var err error

	if events, err = imapClient.ReadCalendarInvites(3); err != nil {
		return fmt.Errorf("failed reading emails: %v", err)
	}

	for _, event := range events {
		if err := calClient.PutEvent(ctx, event); err != nil {
			return fmt.Errorf("failed adding event: %v", err)
		}
	}
	return nil
}
