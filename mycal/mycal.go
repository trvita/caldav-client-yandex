package mycal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	webdav "github.com/trvita/caldav-client-yandex"
	"github.com/trvita/caldav-client-yandex/caldav"
	"github.com/trvita/go-ical"
	"golang.org/x/term"
)

func GetCredentials(r io.Reader) (string, string, error) {
	reader := bufio.NewReader(r)
	fmt.Print("username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	username = strings.TrimSpace(username)

	fmt.Print("password: ")
	var password string
	if r == os.Stdin {
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", "", err
		}
		password = string(bytePassword)
		fmt.Println()
	} else {
		password, err = reader.ReadString('\n')
		if err != nil {
			return "", "", err
		}
		password = strings.TrimSpace(password)
	}
	return username, password, nil
}

func CreateClient(url string, r io.Reader) (*caldav.Client, string, context.Context, error) {
	username, password, err := GetCredentials(r)
	if err != nil {
		return nil, "", nil, err
	}
	httpClient := webdav.HTTPClientWithBasicAuth(&http.Client{}, username, password)
	client, err := caldav.NewClient(httpClient, url)
	if err != nil {
		return nil, "", nil, err
	}

	ctx := context.Background()
	principal, err := client.FindCurrentUserPrincipal(ctx)
	if err != nil {
		return nil, "", nil, err
	}
	return client, principal, ctx, nil
}

func ListCalendars(ctx context.Context, client *caldav.Client, homeset string) error {
	calendars, err := client.FindCalendars(ctx, homeset)
	if err != nil {
		return err
	}
	for _, calendar := range calendars {
		fmt.Printf("Calendar: %s\n", calendar.Name)
	}
	return nil
}

func CreateCalendar(ctx context.Context, client *caldav.Client, homeset string, calendarName string, summary string, startDateTime time.Time, endDateTime time.Time) error {
	err := client.Mkdir(ctx, calendarName)
	if err != nil {
		return err
	}
	// calendar := ical.NewCalendar()
	// calendar.Props.SetText(ical.PropVersion, "2.0")
	// calendar.Props.SetText(ical.PropProductID, "-//Yandex LLC//Yandex Calendar//EN")
	// calendar.Props.SetText(ical.PropCalendarScale, "GREGORIAN")

	// event, err := GetEvent(summary, startDateTime, endDateTime)
	// if err != nil {
	// 	return err
	// }

	// uid, err := event.Props.Text("UID")
	// if err != nil {
	// 	return err
	// }

	// calendar.Children = append(calendar.Children, event.Component)
	// calendarURL := homeset + calendarName + "/"
	// fmt.Println(calendarURL)
	// _, err = client.PutCalendarObject(ctx, calendarURL+uid+".ics", calendar)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func FindCalendar(ctx context.Context, client *caldav.Client, homeset string, calendarName string) (caldav.Calendar, error) {
	var calendar caldav.Calendar

	calendars, err := client.FindCalendars(ctx, homeset)
	if err != nil {
		return calendar, err
	}
	for _, calendar = range calendars {
		if calendar.Name == calendarName {
			return calendar, nil
		}
	}
	return calendar, fmt.Errorf("calendar with name %s not found", calendarName)
}

func ListEvents(ctx context.Context, client *caldav.Client, calendar caldav.Calendar) error {
	query := &caldav.CalendarQuery{
		CompRequest: caldav.CalendarCompRequest{
			Name:     "VCALENDAR",
			AllProps: true,
			Comps: []caldav.CalendarCompRequest{{
				Name:     "VEVENT",
				AllProps: true,
			}},
		},
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []caldav.CompFilter{{
				Name: "VEVENT",
			}},
		},
	}
	cal, err := client.QueryCalendar(
		ctx,
		calendar.Path,
		query,
	)
	if err != nil {
		return err
	}
	for _, calendarObject := range cal {
		for _, event := range calendarObject.Data.Events() {
			for _, prop := range event.Props {
				for _, p := range prop {
					fmt.Printf("%s: %s\n", p.Name, p.Value)
				}
			}
			fmt.Println()
		}
	}
	return nil
}

func GetEvent(summary string, startDateTime time.Time, endDateTime time.Time) (*ical.Event, error) {
	event := ical.NewEvent()
	uid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	event.Props.SetText(ical.PropUID, uid.String())
	event.Props.SetText(ical.PropSummary, summary)
	event.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())
	event.Props.SetDateTime(ical.PropDateTimeStart, startDateTime)
	event.Props.SetDateTime(ical.PropDateTimeEnd, endDateTime)
	return event, nil
}

func CreateEvent(ctx context.Context, client *caldav.Client, homeset, calendarName string, event *ical.Event) error {
	calendar := ical.NewCalendar()
	calendar.Props.SetText(ical.PropVersion, "2.0")
	calendar.Props.SetText(ical.PropProductID, "-//trvita//EN")
	calendar.Props.SetText(ical.PropCalendarScale, "GREGORIAN")

	calendar.Children = append(calendar.Children, event.Component)
	eventUID, err := event.Props.Text(ical.PropUID)
	if err != nil {
		return err
	}
	eventURL := homeset + calendarName + "/" + eventUID + ".ics"
	_, err = client.PutCalendarObject(ctx, eventURL, calendar)
	if err != nil {
		return err
	}
	return nil
}

func Delete(ctx context.Context, client *caldav.Client, path string) error {
	err := client.RemoveAll(ctx, path)
	if err != nil {
		return err
	}
	return nil
}
