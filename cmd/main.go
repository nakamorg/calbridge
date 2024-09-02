package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/emersion/go-ical"
	"github.com/nakamorg/calbridge/pkg/backend"
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

	backend := backend.NewFileBackend("/tmp/calbridge.csv")

	if err := sendInvites(ctx, calClient, smtpClient, backend); err != nil {
		log.Fatal(err)
	}

	if err := addInvites(ctx, calClient, imapClient, backend); err != nil {
		log.Fatal(err)
	}

}

func sendInvites(ctx context.Context, calClient *caldav.Client, smtpClient *email.SMTPClient, storage backend.Backend) error {
	var events []*ical.Calendar
	var err error
	var data backend.Data

	if events, err = calClient.GetEvents(ctx, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 1, 0)); err != nil {
		return fmt.Errorf("failed reading future events: %v", err)
	}

	for _, event := range events {
		if data, err = eventBackendData(event); err != nil {
			return fmt.Errorf("failed creating event backend data: %v", err)
		}
		if data, err = storage.Get(ctx, data); err != nil {
			return fmt.Errorf("failed getting event backend data: %v", err)
		}
		if data.Synced {
			continue
		}
		if err = smtpClient.SendCalendarInvite(event); err != nil {
			return fmt.Errorf("failed sending invitation: %v", err)
		}
		data.Direction = backend.DirectionOut
		data.Synced = true
		data.SyncedTime = time.Now()
		if err = storage.Put(ctx, data); err != nil {
			return fmt.Errorf("invitations were already sent but failed setting event backend data: %v", err)
		}
	}
	return nil
}

func addInvites(ctx context.Context, calClient *caldav.Client, imapClient *email.IMAPClient, storage backend.Backend) error {
	var events []*ical.Calendar
	var err error
	var data backend.Data

	if events, err = imapClient.ReadCalendarInvites(3); err != nil {
		return fmt.Errorf("failed reading emails: %v", err)
	}

	for _, event := range events {
		if data, err = eventBackendData(event); err != nil {
			return fmt.Errorf("failed creating event backend data: %v", err)
		}
		if data, err = storage.Get(ctx, data); err != nil {
			return fmt.Errorf("failed getting event backend data: %v", err)
		}
		if data.Synced {
			continue
		}
		if err := calClient.PutEvent(ctx, event); err != nil {
			return fmt.Errorf("failed adding event: %v", err)
		}
		data.Direction = backend.DirectionIn
		data.Synced = true
		data.SyncedTime = time.Now()
		if err = storage.Put(ctx, data); err != nil {
			return fmt.Errorf("invitations were already added but failed setting event backend data: %v", err)
		}
	}
	return nil
}

func eventBackendData(cal *ical.Calendar) (backend.Data, error) {
	data := backend.Data{Synced: false}
	for _, event := range cal.Events() {
		for _, p := range event.Props.Values(ical.PropUID) {
			data.UID = p.Value
		}
		if len(data.UID) == 0 {
			return data, fmt.Errorf("could not find event uid")
		}
	}
	hash, err := eventHash(cal)
	if err != nil {
		return data, fmt.Errorf("could not find event hash: %v", err)
	}
	data.Hash = hash
	return data, nil
}

func eventHash(cal *ical.Calendar) (string, error) {
	hash := sha256.New()

	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(cal); err != nil {
		return "", err
	}
	hash.Write(buf.Bytes())
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
