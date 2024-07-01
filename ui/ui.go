package ui

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/trvita/caldav-client-yandex/caldav"
	"github.com/trvita/caldav-client-yandex/mycal"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("\u001b[31m%s: %s\u001b[0m\n", msg, err)
	}
}

func BlueLine(str string) {
	fmt.Printf("\u001b[34m%s\u001b[0m", str)
}

func RedLine(err error) {
	fmt.Printf("\u001b[31m%s\u001b[0m\n", err)
}

func ExtractUsername(path string) string {
	startMarker := "/principals/users/"
	startPos := strings.Index(path, startMarker)
	startPos += len(startMarker)
	username := path[startPos:]

	return username
}

func GetString(message string) string {
	var str string
	fmt.Print(message)
	fmt.Scan(&str)
	return str
}

func GetEvent() (string, string, time.Time, time.Time) {
	var summary, startDate, startTime, endDate, endTime string
	var startDateTime, endDateTime time.Time
	var err error
	summary = GetString("Enter event summary: ")
	uid, err := uuid.NewUUID()
	if err != nil {
		log.Fatalf("could not generate UUID: %v", err)
	}
	for {
		startDate = GetString("Enter event start date (YYYY.MM.DD): ")
		startTime = GetString("Enter event start time (HH.MM.SS): ")

		startDateTime, err = time.Parse("2006.01.02 15.04.05", startDate+" "+startTime)
		if err != nil {
			fmt.Println("invalid start date/time format")
			continue
		}
		break
	}
	for {
		endDate = GetString("Enter event end date (YYYY.MM.DD): ")
		endTime = GetString("Enter event end time (HH.MM.SS): ")

		endDateTime, err = time.Parse("2006.01.02 15.04.05", endDate+" "+endTime)
		if err != nil {
			fmt.Println("invalid end date/time format")
			continue
		}
		break
	}
	return summary, uid.String(), startDateTime, endDateTime
}

func StartMenu(url string) {
	BlueLine("Main menu:\n")
	for {
		fmt.Println("1. Log in")
		fmt.Println("0. Exit")
		var answer int
		fmt.Scan(&answer)
		switch answer {
		case 1:
			var client *caldav.Client
			var principal string
			var ctx context.Context
			var err error
			for {
				client, principal, ctx, err = mycal.CreateClient(url, os.Stdin)
				if err == nil {
					break
				}
				fmt.Println("Wrong username or password, try again? ([y/n])")
				var ans string
				fmt.Scan(&ans)
				ans = strings.ToLower(ans)
				if ans == "y" {
					continue
				} else if ans == "n" {
					BlueLine("Shutting down...\n")
					return
				}

			}
			err = CalendarMenu(client, principal, ctx)
			if err != nil {
				RedLine(err)
				return
			}
		case 0:
			BlueLine("Shutting down...\n")
			return
		}
	}
}

func CalendarMenu(client *caldav.Client, principal string, ctx context.Context) error {
	homeset, err := client.FindCalendarHomeSet(ctx, principal)
	if err != nil {
		RedLine(err)
		return err
	}

	username := ExtractUsername(principal)
	BlueLine("Current user: " + username + "\n")
	for {
		fmt.Println("1. List calendars")
		fmt.Println("2. Goto calendar")
		fmt.Println("3. Create calendar")
		fmt.Println("0. Log out")
		var answer int
		fmt.Scan(&answer)
		switch answer {
		case 1:
			err := mycal.ListCalendars(ctx, client, homeset)
			if err != nil {
				RedLine(err)
			}
		case 2:
			calendarName := GetString("Enter calendar name: ")
			calendar, err := mycal.FindCalendar(ctx, client, homeset, calendarName)
			if err != nil {
				RedLine(err)
				break
			}
			EventMenu(ctx, client, homeset, calendar)
		case 3:
			calendarName := GetString("Enter new calendar name: ")
			summary, uid, startDateTime, endDateTime := GetEvent()
			err := mycal.CreateCalendar(ctx, client, homeset, calendarName, summary, uid, startDateTime, endDateTime)
			if err != nil {
				RedLine(err)
			}
		case 0:
			BlueLine("Logging out...\n")
			return nil
		}
	}
}

func EventMenu(ctx context.Context, client *caldav.Client, homeset string, calendar caldav.Calendar) {
	BlueLine("Current calendar:" + calendar.Name + " " + calendar.Path + " " + homeset + "\n")
	for {
		fmt.Println("1. List events")
		fmt.Println("2. Create event")
		fmt.Println("3. Delete event")
		fmt.Println("0. Back to calendar menu")
		var answer int
		fmt.Scan(&answer)
		switch answer {
		case 1:
			err := mycal.ListEvents(ctx, client, calendar)
			if err != nil {
				RedLine(err)
			}
		case 2:
			summary, uid, startDateTime, endDateTime := GetEvent()
			event := mycal.GetEvent(summary, uid, startDateTime, endDateTime)
			err := mycal.CreateEvent(ctx, client, calendar, event)
			if err != nil {
				RedLine(err)
			}

		case 3:
			eventUID := GetString("Enter event UID: ")
			err := mycal.DeleteEvent(ctx, client, calendar, eventUID)
			if err != nil {
				RedLine(err)
			}

		case 0:
			BlueLine("Returning to calendar menu...\n")
			return
		}
	}
}
