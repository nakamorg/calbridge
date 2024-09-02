package email

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/emersion/go-ical"
	sasl "github.com/emersion/go-sasl"
	smtp "github.com/emersion/go-smtp"
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

func (c *SMTPClient) SendCalendarInvite(cal *ical.Calendar) error {
	from := c.from
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
	mailPrefix := "mailto:"
	for _, e := range cal.Events() {
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

func subject(cal *ical.Calendar) string {
	subject := "Invitation"
	for _, e := range cal.Events() {
		for _, p := range e.Props.Values(ical.PropSummary) {
			subject = subject + ": " + p.Value
		}
	}
	return subject
}
