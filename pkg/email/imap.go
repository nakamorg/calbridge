package email

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-sasl"
)

func ReadCalendarInvites(host, username, password string, hours int) ([]*ical.Calendar, error) {
	// Connect to the IMAP server
	c, err := client.DialTLS(host, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %v", err)
	}
	defer c.Logout()

	if err := c.Authenticate(sasl.NewPlainClient("", username, password)); err != nil {
		return nil, fmt.Errorf("failed to login to IMAP server: %v", err)
	}

	// Select the INBOX mailbox
	if _, err := c.Select("INBOX", false); err != nil {
		return nil, fmt.Errorf("failed to select INBOX: %v", err)
	}

	// Calculate the time range for the last n hours
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	criteria := imap.NewSearchCriteria()
	criteria.SentSince = since

	// Search for emails within the time range
	seqNums, err := c.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search emails: %v", err)
	}
	if len(seqNums) == 0 {
		return nil, nil
	}

	// We want to fetch `BODY.PEEK[]`, peek to prevent marking the emails as `Seen`.
	section := &imap.BodySectionName{Peek: true}
	items := []imap.FetchItem{section.FetchItem()}
	msgs := make(chan *imap.Message, len(seqNums))
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)
	if err := c.Fetch(seqSet, items, msgs); err != nil {
		return nil, fmt.Errorf("failed to fetch email: %v", err)
	}

	var invites []*ical.Calendar
	for msg := range msgs {
		if msg == nil {
			continue
		}
		for _, bodySection := range msg.Body {
			msgInvites, err := extractCalendarInvites(bodySection)
			if err != nil {
				return invites, err
			}
			invites = append(invites, msgInvites...)

		}
	}
	return invites, nil
}

func extractCalendarInvites(bodySection imap.Literal) ([]*ical.Calendar, error) {
	var invites []*ical.Calendar

	mr, err := mail.CreateReader(bodySection)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail reader: %v", err)
	}
	defer mr.Close()
	// Iterate over the email parts
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to read email part: %v", err)
		}
		// Check if the part is a calendar invite
		if strings.HasPrefix(p.Header.Get("Content-Type"), "text/calendar") {
			invite, err := ical.NewDecoder(p.Body).Decode() // io.ReadAll(p.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read calendar invite: %v", err)
			}

			invites = append(invites, invite)
		}
	}
	return invites, nil
}
