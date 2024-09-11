package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/emersion/go-ical"
	"github.com/nakamorg/calbridge/pkg/backend"
	"github.com/nakamorg/calbridge/pkg/caldav"
	"github.com/nakamorg/calbridge/pkg/config"
	"github.com/nakamorg/calbridge/pkg/email"
	"github.com/nakamorg/calbridge/pkg/util"
)

func main() {
	ctx := context.Background()
	var err error
	var users []config.User
	var storage backend.Backend
	var configFolder string
	configFolder, err = configFolderPath()
	if err != nil {
		log.Fatal(err)
	}

	configFilePath := filepath.Join(configFolder, "config.json")
	if users, err = config.LoadUsersFromConfig(configFilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("could not find the config file. A sample config file will be created for you at %s\n", configFilePath)
			if err := config.CreateSampleConfig(configFilePath); err != nil {
				log.Fatalf("failed creating sample config file: %v", err)
			}
			log.Fatal("update the sample file with real configuration data")
		} else {
			log.Fatal(err)
		}
	}

	if storage, err = backend.NewBoltBackend(filepath.Join(configFolder, "bolt.db")); err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	for _, user := range users {
		if err := handleUser(ctx, user, storage); err != nil {
			log.Print(err)
		}
	}
}

func configFolderPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".calbridge"), nil
}

func handleUser(ctx context.Context, user config.User, storage backend.Backend) error {
	var err error
	var calClient *caldav.Client
	var smtpClient *email.SMTPClient
	var imapClient *email.IMAPClient

	if calClient, err = caldav.NewClient(user.CalDAV.Username, user.CalDAV.Password, user.CalDAV.URL); err != nil {
		return fmt.Errorf("failed to create caldav client: %w", err)
	}

	if smtpClient, err = email.NewSMTPClient(user.SMTP.Username, user.SMTP.Password, user.SMTP.Host, "587"); err != nil {
		return fmt.Errorf("failed to create smtp client: %w", err)
	}
	defer smtpClient.Close()

	if imapClient, err = email.NewIMAPClient(user.IMAP.Username, user.IMAP.Password, user.IMAP.Host, "993"); err != nil {
		return fmt.Errorf("failed to create imap client: %w", err)
	}
	defer imapClient.Close()

	if err = sendInvites(ctx, user.Name, user.CalDAV.EventDays, calClient, smtpClient, storage); err != nil {
		log.Fatal(err)
	}

	if err = addInvites(ctx, user.Name, user.IMAP.EmailHours, calClient, imapClient, storage); err != nil {
		log.Fatal(err)
	}
	return nil
}

func sendInvites(ctx context.Context, username string, eventDays int, calClient *caldav.Client, smtpClient *email.SMTPClient, storage backend.Backend) error {
	var events []*ical.Calendar
	var err error
	var data backend.Data

	if events, err = calClient.GetEvents(ctx, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, eventDays)); err != nil {
		return fmt.Errorf("failed reading future events: %v", err)
	}
	for _, event := range events {
		if data, err = eventBackendData(ctx, username, event, backend.DirectionOut, storage); err != nil {
			return fmt.Errorf("failed creating event backend data: %v", err)
		}
		if data.Synced || data.Direction != backend.DirectionOut {
			continue
		}
		fmt.Println("sending")
		fmt.Println(event.Events()[0].Props)
		if err = smtpClient.SendCalendarInvite(event); err != nil {
			return fmt.Errorf("failed sending invitation: %v", err)
		}
		data.Synced = true
		data.SyncedTime = time.Now()
		if err = storage.Put(ctx, data); err != nil {
			return fmt.Errorf("invitations were already sent but failed setting event backend data: %v", err)
		}
	}
	return nil
}

func addInvites(ctx context.Context, username string, emailHours int, calClient *caldav.Client, imapClient *email.IMAPClient, storage backend.Backend) error {
	var events []*ical.Calendar
	var err error
	var data backend.Data

	if events, err = imapClient.ReadCalendarInvites(emailHours); err != nil {
		return fmt.Errorf("failed reading emails: %v", err)
	}

	for _, event := range events {
		if data, err = eventBackendData(ctx, username, event, backend.DirectionIn, storage); err != nil {
			return fmt.Errorf("failed creating event backend data: %v", err)
		}
		if data.Synced || data.Direction != backend.DirectionIn {
			continue
		}
		if err := calClient.PutEvent(ctx, event); err != nil {
			return fmt.Errorf("failed adding event: %v", err)
		}
		fmt.Println("adding")
		fmt.Println(event.Events()[0].Props)
		data.Synced = true
		data.SyncedTime = time.Now()
		if err = storage.Put(ctx, data); err != nil {
			return fmt.Errorf("invitations were already added but failed setting event backend data: %v", err)
		}
	}
	return nil
}

func eventBackendData(ctx context.Context, username string, cal *ical.Calendar, direction backend.Direction, storage backend.Backend) (backend.Data, error) {
	var err error
	var uid, hash string

	data := backend.Data{
		User:      username,
		Synced:    false,
		Direction: direction,
	}

	if uid, err = util.EventUid(cal); err != nil {
		return data, err
	}
	data.UID = uid

	if hash, err = eventHash(cal); err != nil {
		return data, fmt.Errorf("could not find event hash: %v", err)
	}
	data.Hash = hash

	if data, err = storage.Get(ctx, data); err != nil {
		return data, fmt.Errorf("failed getting event backend data: %v", err)
	}
	return data, nil
}

func eventHash(cal *ical.Calendar) (string, error) {
	// We need smaller (in size) hashes and md5 should be secure enough for our case
	hash := md5.New()
	var buf bytes.Buffer
	uid, _ := util.EventUid(cal)
	attendees := util.EventAttendees(cal)
	organizers := util.EventOrganizers(cal)
	desc := util.EventDescription(cal)
	summary := util.EventSummary(cal)
	dtstart, _ := util.EventDTStart(cal)
	dtend, _ := util.EventDTEnd(cal)

	buf.WriteString(uid)
	for attendee, status := range attendees {
		buf.WriteString(attendee)
		buf.WriteString(status)
	}
	// TODO: I've noticed that when an event is added to my caldav server, the status for organizer becomes `NEEDS_ACTION`
	// even though it was ACCEPTED in the received event
	for organizer, status := range organizers {
		buf.WriteString(organizer)
		buf.WriteString(status)
	}
	buf.WriteString(desc)
	buf.WriteString(summary)
	buf.WriteString(dtstart.String())
	buf.WriteString(dtend.String())

	hash.Write(buf.Bytes())
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
