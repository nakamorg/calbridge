package email

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/emersion/go-ical"
	sasl "github.com/emersion/go-sasl"
	smtp "github.com/emersion/go-smtp"
	"github.com/emersion/go-webdav/caldav"
)

type client struct {
	username, password, host, smtpPort, imapPort string
}

func NewMailClient(username, password, host, smtpPort, imapPort string) *client {
	return &client{
		username: username,
		password: password,
		host:     host,
		smtpPort: smtpPort,
		imapPort: imapPort,
	}
}

func (c *client) SendCalendarInvite(calObject caldav.CalendarObject) error {
	from := c.username
	to := attendees(calObject)
	if len(to) == 0 {
		return nil
	}

	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(calObject.Data); err != nil {
		return err
	}

	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(to, ",")
	headers["Subject"] = subject(calObject)
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

	smtpServer := fmt.Sprintf("%s:%s", c.host, c.smtpPort)
	auth := sasl.NewLoginClient(c.username, c.password)
	smtpClient, err := smtp.DialStartTLS(smtpServer, nil)
	if err != nil {
		return err
	}
	smtpClient.Auth(auth)
	return smtpClient.SendMail(from, to, strings.NewReader(msg))
}

func attendees(calObject caldav.CalendarObject) []string {
	var attendees []string
	mailPrefix := "mailto:"
	for _, e := range calObject.Data.Events() {
		candidates := e.Props.Values(ical.PropAttendee)
		for _, c := range candidates {
			address := c.Value
			if c.Params.Get(ical.ParamParticipationStatus) == "NEEDS-ACTION" && strings.HasPrefix(address, mailPrefix) {
				attendees = append(attendees, strings.TrimPrefix(address, mailPrefix))
			}
		}
	}
	return attendees
}

func subject(calObject caldav.CalendarObject) string {
	subject := "Invitation"
	for _, e := range calObject.Data.Events() {
		for _, p := range e.Props.Values(ical.PropSummary) {
			subject = subject + ": " + p.Value
		}
	}
	return subject
}

func (c *client) CheckCalendarInvite() error {
	// use imap to check  my emails for calendar invites or use any other techniques that does
	// not increase the load on my mail server.
	// do not fetch messages older than a given time
	return nil
}
