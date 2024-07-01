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

	"github.com/emersion/go-ical"
	webdav "github.com/trvita/caldav-client-yandex"
	"github.com/trvita/caldav-client-yandex/caldav"
	"golang.org/x/term"
)

func GetCredentials(r io.Reader) (string, string, error) {
	reader := bufio.NewReader(r)
	fmt.Print("username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("%s\n", err)
		return "", "", err
	}
	username = strings.TrimSpace(username)

	fmt.Print("password: ")
	var password string
	if r == os.Stdin {
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Printf("%s\n", err)
			return "", "", err
		}
		password = string(bytePassword)
		fmt.Println()
	} else {
		password, err = reader.ReadString('\n')
		if err != nil {
			fmt.Printf("%s\n", err)
			return "", "", err
		}
		password = strings.TrimSpace(password)
	}
	return username, password, nil
}

func CreateClient(url string, r io.Reader) (*caldav.Client, string, context.Context, error) {
	username, password, err := GetCredentials(r)
	if err != nil {
		fmt.Printf("%s\n", err)
		return nil, "", nil, err
	}
	httpClient := webdav.HTTPClientWithBasicAuth(&http.Client{}, username, password)
	client, err := caldav.NewClient(httpClient, url)
	if err != nil {
		fmt.Printf("%s\n", err)
		return nil, "", nil, err
	}

	ctx := context.Background()
	principal, err := client.FindCurrentUserPrincipal(ctx)
	if err != nil {
		fmt.Printf("%s\n", err)
		return nil, "", nil, err
	}
	return client, principal, ctx, nil
}

func ListCalendars(ctx context.Context, client *caldav.Client, homeset string) error {
	calendars, err := client.FindCalendars(ctx, homeset)
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}
	for _, calendar := range calendars {
		fmt.Printf("Calendar: %s\n", calendar.Name)
	}
	return nil
}

func CreateCalendar(ctx context.Context, client *caldav.Client, homeset string, calendarName string, summary string, uid string, startDateTime time.Time, endDateTime time.Time) error {
	calendar := ical.NewCalendar()
	calendar.Props.SetText(ical.PropVersion, "2.0")
	calendar.Props.SetText(ical.PropProductID, "-//trvita//EN")
	calendar.Props.SetText(ical.PropCalendarScale, "GREGORIAN")

	event := GetEvent(summary, uid, startDateTime, endDateTime)

	calendar.Children = append(calendar.Children, event.Component)
	var buf strings.Builder
	encoder := ical.NewEncoder(&buf)
	err := encoder.Encode(calendar)
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}
	calendarURL := homeset + calendarName + "/"
	_, err = client.PutCalendarObject(ctx, calendarURL, calendar)
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}
	fmt.Println("Calendar created")
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
			Name:  "VCALENDAR",
			Props: []string{"VERSION"},
			Comps: []caldav.CalendarCompRequest{{
				Name: "VEVENT",
				Props: []string{
					"SUMMARY",
					"UID",
					"DTSTART",
					"DTEND",
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
	cal, err := client.QueryCalendar(
		ctx,
		calendar.Path,
		query,
	)
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}
	for _, calendarObject := range cal {
		for _, event := range calendarObject.Data.Events() {
			summary, err := event.Props.Text("SUMMARY")
			if err != nil {
				fmt.Printf("%s\n", err)
				return err
			}
			uid, err := event.Props.Text("UID")
			if err != nil {
				fmt.Printf("%s\n", err)
				return err
			}
			dtstart, err := event.Props.DateTime("DTSTART", time.Local)
			if err != nil {
				fmt.Printf("%s\n", err)
				return err
			}
			dtend, err := event.Props.DateTime("DTEND", time.Local)
			if err != nil {
				fmt.Printf("%s\n", err)
				return err
			}
			fmt.Printf("Summary: %s,\tUID: %s,\tStart: %s,\tEnd: %s\n", summary, uid, dtstart, dtend)
		}
	}
	return nil
}

func GetEvent(summary string, uid string, startDateTime time.Time, endDateTime time.Time) *ical.Event {
	event := ical.NewEvent()
	event.Props.SetText(ical.PropUID, uid)
	event.Props.SetText(ical.PropSummary, summary)
	event.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())
	event.Props.SetDateTime(ical.PropDateTimeStart, startDateTime)
	event.Props.SetDateTime(ical.PropDateTimeEnd, endDateTime)
	return event
}

func CreateEvent(ctx context.Context, client *caldav.Client, calendar caldav.Calendar, event *ical.Event) error {
	// uid, err := event.Props.Text("UID")
	// if err != nil {
	// 	fmt.Printf("%s\n", err)
	// 	return err
	// }
	// eventURL := calendar.Path + uid + ".ics"

	query := &caldav.CalendarQuery{
		CompRequest: caldav.CalendarCompRequest{
			Name:  "VCALENDAR",
			Props: []string{"VERSION"},
			Comps: []caldav.CalendarCompRequest{{
				Name: "VEVENT",
				Props: []string{
					"SUMMARY",
					"UID",
					"DTSTART",
					"DTEND",
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
	cal, err := client.QueryCalendar(
		ctx,
		calendar.Path,
		query,
	)
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}

	for _, calendarObject := range cal {
		if calendarObject.Data.Name == calendar.Name {
			calendarObject.Data.Component.Children = append(calendarObject.Data.Component.Children, event.Component)
			_, err := client.PutCalendarObject(ctx, calendarObject.Data.Name, calendarObject.Data)
			if err != nil {
				fmt.Printf("%s\n", err)
				return err
			}
			break
		}
	}
	return nil
}

func DeleteEvent(ctx context.Context, client *caldav.Client, calendar caldav.Calendar, eventUID string) error {
	calendarObject, err := client.GetCalendarObject(ctx, calendar.Name)
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}
	var updatedEvents []*ical.Component
	for _, component := range calendarObject.Data.Component.Children {
		if component.Name == ical.CompEvent {
			uid, err := component.Props.Text(ical.PropUID)
			if err != nil {
				fmt.Printf("%s\n", err)
				return err
			}
			if uid == eventUID {
				continue
			}
		}
		updatedEvents = append(updatedEvents, component)
	}
	if len(updatedEvents) == 0 {
		fmt.Println("Cannot delete the event as it would leave the calendar empty.") // add delete calendar call if implemented
		return nil
	}

	calendarObject.Data.Component.Children = updatedEvents

	var buf strings.Builder
	encoder := ical.NewEncoder(&buf)
	err = encoder.Encode(calendarObject.Data)
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}
	_, err = client.PutCalendarObject(ctx, calendarObject.Data.Name, calendarObject.Data)
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}
	fmt.Println("Event deleted")
	return nil
}
