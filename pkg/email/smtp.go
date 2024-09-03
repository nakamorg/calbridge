package email

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/emersion/go-ical"
	sasl "github.com/emersion/go-sasl"
	smtp "github.com/emersion/go-smtp"
	"github.com/nakamorg/calbridge/pkg/util"
)

type SMTPClient struct {
	from string
	c    *smtp.Client
}

func NewSMTPClient(username, password, host, port string) (*SMTPClient, error) {
	smtpServer := fmt.Sprintf("%s:%s", host, port)
	c, err := smtp.DialStartTLS(smtpServer, nil)
	if err != nil {
		return nil, err
	}
	c.Auth(sasl.NewLoginClient(username, password))
	return &SMTPClient{
		from: username,
		c:    c,
	}, nil
}

func (c *SMTPClient) Close() {
	c.c.Close()
}

// SendCalendarInvite sends calendar invite to all the attendees using email. Invites are not sent
// if the calendar organizer and the email sender does not match
func (c *SMTPClient) SendCalendarInvite(cal *ical.Calendar) error {
	from := c.from
	if !isOrganizer(cal, from) {
		return nil
	}
	to := attendees(cal)
	if len(to) == 0 {
		return nil
	}

	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(cal); err != nil {
		return err
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(to, ",")
	headers["Subject"] = subject(cal)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "multipart/mixed; boundary=\"boundary\""

	var msg string
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n--boundary\r\n"
	msg += "Content-Type: text/plain; charset=utf-8\r\n\r\n"
	msg += "Please find the attached calendar invite.\r\n"
	msg += "\r\n--boundary\r\n"
	msg += "Content-Type: text/calendar; method=REQUEST; charset=utf-8\r\n"
	msg += "Content-Disposition: attachment; filename=\"invite.ics\"\r\n\r\n"
	msg += buf.String()
	msg += "\r\n--boundary--\r\n"

	return c.c.SendMail(from, to, strings.NewReader(msg))
}

func attendees(cal *ical.Calendar) []string {
	var attendees []string
	attendeeToParticipationStatus := util.EventAttendees(cal)
	for attendee, status := range attendeeToParticipationStatus {
		if status == "NEEDS-ACTION" {
			attendees = append(attendees, attendee)
		}
	}
	return attendees
}

func isOrganizer(cal *ical.Calendar, user string) bool {
	organizerToParticipationStatus := util.EventOrganizers(cal)
	_, ok := organizerToParticipationStatus[user]
	return ok
}

func subject(cal *ical.Calendar) string {
	subject := "Invitation"
	summary := util.EventSummary(cal)
	return subject + ": " + summary
}
