package caldav

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
	"github.com/nakamorg/calbridge/pkg/http"
	"github.com/nakamorg/calbridge/pkg/util"
)

type Client struct {
	url string
	c   *caldav.Client
}

func NewClient(username, password, url string) (*Client, error) {
	c, err := caldav.NewClient(http.HTTPClientWithDigestAuth(nil, username, password), url)
	if err != nil {
		return nil, err
	}
	return &Client{
		url: url,
		c:   c,
	}, nil
}

// GetCalendarObject returns the CalendarObjects from your calendar between the start and end time
func (c *Client) GetCalendarObject(ctx context.Context, start, end time.Time) ([]caldav.CalendarObject, error) {
	var calObjects []caldav.CalendarObject
	caldavClient := c.c

	calendars, err := caldavClient.FindCalendars(ctx, "")
	if err != nil {
		return calObjects, err
	}
	for _, calendar := range calendars {
		calendarQuery := caldav.CalendarQuery{
			CompFilter: caldav.CompFilter{
				Name: ical.CompCalendar,
				Comps: []caldav.CompFilter{{
					Name:  ical.CompEvent,
					Start: start,
					End:   end,
				}},
			},
		}
		return caldavClient.QueryCalendar(ctx, calendar.Path, &calendarQuery)
	}

	return calObjects, fmt.Errorf("no calendars found")
}

// GetEvents returns the CalendarObjects from your calendar between the start and end time
func (c *Client) GetEvents(ctx context.Context, start, end time.Time) ([]*ical.Calendar, error) {
	calObjects, err := c.GetCalendarObject(ctx, start, end)
	if err != nil {
		return nil, err
	}
	var events []*ical.Calendar
	for _, calcalObject := range calObjects {
		events = append(events, calcalObject.Data)
	}

	return events, nil
}

// PutEvent puts the Calendar event in your calendar. It removes the METHOD property from the event.
// If the METHOD property value was CANCEL, it'll try to remove the event from the server.
func (c *Client) PutEvent(ctx context.Context, cal *ical.Calendar) error {
	caldavClient := c.c

	uid, err := util.EventUid(cal)
	if err != nil {
		return fmt.Errorf("could not calculate path to save the event: %v", err)
	}
	path := fmt.Sprintf("%s.%s", uid, ical.Extension)
	if strings.HasPrefix(string(ical.EventCancelled), methodProp(cal)) {
		caldavClient.RemoveAll(ctx, path) // ignore any errors here
		return nil
	}
	cal.Props.Del(ical.PropMethod)
	if _, err := caldavClient.PutCalendarObject(ctx, path, cal); err != nil {
		return err
	}
	return nil
}

func methodProp(cal *ical.Calendar) string {
	if cal == nil {
		return ""
	}
	values := cal.Props.Values(ical.PropMethod)
	if len(values) != 0 {
		return values[0].Value
	}
	return ""
}
