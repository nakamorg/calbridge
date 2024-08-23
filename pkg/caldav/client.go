package caldav

import (
	"context"
	"fmt"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
	"github.com/nakamorg/calbridge/pkg/http"
)

type Client struct {
	URL      string
	Username string
	Password string
	Calendar string
}

func (c *Client) ReadFutureEvents(ctx context.Context) ([]ical.Event, error) {
	var events []ical.Event

	client, err := caldav.NewClient(http.DigestAuthClient{
		Username: c.Username,
		Password: c.Password,
	}, c.URL)
	if err != nil {
		return events, err
	}

	cals, err := client.FindCalendars(ctx, "/calendars/zqau3bsatz/2ixorhsnoh")
	if err != nil {
		return events, err
	}
	for _, cal := range cals {
		fmt.Printf("name: %q, path: %q, com: %q\n", cal.Name, cal.Path, cal.SupportedComponentSet)

		calendarQuery := caldav.CalendarQuery{
			CompRequest: caldav.CalendarCompRequest{
				Name: "VCALENDAR",
				Comps: []caldav.CalendarCompRequest{{
					Name: "VEVENT",
					Props: []string{
						"SUMMARY",
						"UID",
						"DTSTART",
						"DTEND",
						"DURATION",
					},
				}},
			},
			CompFilter: caldav.CompFilter{
				Name: "VCALENDAR",
				Comps: []caldav.CompFilter{{
					Name: "VEVENT",
				}},
			},
		}

		calObjects, err := client.QueryCalendar(ctx, cal.Path, &calendarQuery)
		if err != nil {
			return events, err
		}
		for _, calObject := range calObjects {
			events = append(events, calObject.Data.Events()...)
		}
	}

	return events, nil
}
