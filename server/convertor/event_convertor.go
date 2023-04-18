package convertor

import (
	"github.com/emersion/go-ical"
	"github.com/lugamuga/go-webdav/caldav"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/dto"
	"github.com/lugamuga/mattermost-yandex-calendar-plugin/server/util"
	"github.com/pkg/errors"
	"time"
)

func CalendarObjectToEventArray(calendarObjects []caldav.CalendarObject, userTimeZone string) ([]dto.Event, error) {
	eventById := make(map[string]dto.Event)
	for _, calendarObject := range calendarObjects {
		for _, e := range calendarObject.Data.Events() {
			eventId := util.GetPropertyValue(e.Props.Get("UID"))
			if _, ok := eventById[eventId]; ok {
				continue
			}
			timezone, err := GetTimezone(calendarObject)
			location, _ := time.LoadLocation(timezone)

			userLocation, err := time.LoadLocation(userTimeZone)
			tzName, tzOffset := time.Now().In(userLocation).Zone()
			changeTZ := time.FixedZone(tzName, tzOffset)

			eventName, _ := e.Props.Text("SUMMARY")
			eventDescription, _ := e.Props.Text("DESCRIPTION")
			eventUrl := util.GetPropertyValue(e.Props.Get("URL"))

			getStartTime, err := e.Props.DateTime("DTSTART", location)
			startTime := getStartTime.In(changeTZ)
			if err != nil {
				return nil, errors.Wrap(err, "Can't parse DTSTART for event "+eventName)
			}

			getEndTime, err := e.Props.DateTime("DTEND", location)
			endTime := getEndTime.In(changeTZ)
			if err != nil {
				return nil, errors.Wrap(err, "Can't parse DTEND for event "+eventName)
			}
			getLastModifiedTime, err := e.Props.DateTime("LAST-MODIFIED", location)
			lastModifiedTime := getLastModifiedTime.In(changeTZ)
			if err != nil {
				return nil, errors.Wrap(err, "Can't parse LAST-MODIFIED for event "+eventName)
			}
			eventById[eventId] = *dto.NewEvent(
				eventId,
				eventName,
				eventDescription,
				eventUrl,
				timezone,
				startTime,
				endTime,
				lastModifiedTime,
			)
		}
	}
	events := make([]dto.Event, 0, len(eventById))
	for _, event := range eventById {
		events = append(events, event)
	}
	return events, nil
}

func GetTimezone(calendarObject caldav.CalendarObject) (string, error) {
	for _, child := range calendarObject.Data.Children {
		if child.Name == ical.CompTimezone {
			return child.Props.Text("TZID")
		}
	}
	return "Etc/UTC", errors.New("Timezone not found")
}

func SliceEventToMapById(events []dto.Event) map[string]dto.Event {
	eventsById := make(map[string]dto.Event, len(events))
	for _, e := range events {
		eventsById[e.Id] = e
	}
	return eventsById
}
