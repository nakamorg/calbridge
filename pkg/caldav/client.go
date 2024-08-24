package caldav

import (
	"context"
	"fmt"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
	"github.com/nakamorg/calbridge/pkg/http"
)

type client struct {
	url string
	c   *caldav.Client
}

func NewClient(username, password, url string) (*client, error) {
	c, err := caldav.NewClient(http.HTTPClientWithDigestAuth(nil, username, password), url)
	if err != nil {
		return nil, err
	}
	return &client{
		url: url,
		c:   c,
	}, nil
}

// GetEvents returns the CalendarObjects from your calendar between the start and end time
func (c *client) GetEvents(ctx context.Context, start, end time.Time) ([]caldav.CalendarObject, error) {
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

// PutEvents puts the CalendarObjects to your calendar
func (c *client) PutEvents(ctx context.Context, calObjects []caldav.CalendarObject) error {
	caldavClient := c.c

	for _, calObject := range calObjects {
		if _, err := caldavClient.PutCalendarObject(ctx, calObject.Path, calObject.Data); err != nil {
			return err
		}
	}

	return nil
}
