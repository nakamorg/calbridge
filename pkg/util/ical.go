package util

import (
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-ical"
)

// EventUid returns the UID of calendar event. If the cal object doesn't have exactly
// one event and one UID prop an error is returned.
func EventUid(cal *ical.Calendar) (string, error) {
	if cal == nil {
		return "", fmt.Errorf("event is nil")
	}
	// get the uid from first event, not sure if this might cause issues
	events := cal.Events()
	if len(events) != 1 {
		return "", fmt.Errorf("calendar has %d events, expected 1", len(events))
	}
	propUids := events[0].Props.Values(ical.PropUID)
	if len(propUids) != 1 {
		return "", fmt.Errorf("length of UID prop is %d, expected 1", len(propUids))
	}
	return propUids[0].Value, nil
}

// EventAttendees returns a map of email addresses of attendees to their participation status
func EventAttendees(cal *ical.Calendar) map[string]string {
	var attendees = map[string]string{}
	mailPrefix := "mailto:"
	for _, e := range cal.Events() {
		candidates := e.Props.Values(ical.PropAttendee)
		for _, c := range candidates {
			address := c.Value
			if strings.HasPrefix(address, mailPrefix) {
				attendee := strings.TrimPrefix(address, mailPrefix)
				participationStatus := c.Params.Get(ical.ParamParticipationStatus)
				attendees[attendee] = participationStatus
			}
		}
	}
	return attendees
}

// EventOrganizers returns a map of email addresses of organizers to their participation status
func EventOrganizers(cal *ical.Calendar) map[string]string {
	var organizers = map[string]string{}
	mailPrefix := "mailto:"
	for _, e := range cal.Events() {
		candidates := e.Props.Values(ical.PropOrganizer)
		for _, c := range candidates {
			address := c.Value
			if strings.HasPrefix(address, mailPrefix) {
				organizer := strings.TrimPrefix(address, mailPrefix)
				participationStatus := c.Params.Get(ical.ParamParticipationStatus)
				organizers[organizer] = participationStatus
			}
		}
	}
	return organizers
}

// EventDescription returns the concatenation of descriptions of all the events in the cal object
func EventDescription(cal *ical.Calendar) string {
	var desc string
	for _, e := range cal.Events() {
		for _, p := range e.Props.Values(ical.PropDescription) {
			desc = desc + p.Value
		}
	}
	return desc
}

// EventSummary returns the concatenation of summaries of all the events in the cal object
func EventSummary(cal *ical.Calendar) string {
	var summary string
	for _, e := range cal.Events() {
		for _, p := range e.Props.Values(ical.PropSummary) {
			summary = summary + p.Value
		}
	}
	return summary
}

// EventDTStart returns the earliest start time from all the events in the cal object
func EventDTStart(cal *ical.Calendar) (time.Time, error) {
	var start time.Time
	for _, e := range cal.Events() {
		dtstart, err := e.DateTimeStart(time.UTC)
		if err != nil {
			return start, err
		}
		if start.IsZero() {
			start = dtstart
		}
		if start.After(dtstart) {
			start = dtstart
		}
	}
	return start, nil
}

// EventDTEnd returns the latest end time from all the events in the cal object
func EventDTEnd(cal *ical.Calendar) (time.Time, error) {
	var end time.Time
	for _, e := range cal.Events() {
		dtend, err := e.DateTimeEnd(time.UTC)
		if err != nil {
			return end, err
		}
		if end.IsZero() {
			end = dtend
		}
		if end.Before(dtend) {
			end = dtend
		}
	}
	return end, nil
}
