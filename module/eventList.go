package module

import (
	"strings"
)

func appendEvent(e Event) bool {
	e = Event(strings.ToUpper(string(e)))

	eventsMut.Lock()
	defer eventsMut.Unlock()

	for _, ev := range Events {
		if e == ev {
			return false
		}
	}

	Events = append(Events, e)
	return true
}

func RegisteredEvents() []Event {
	eventCopy := make([]Event, len(Events))

	eventsMut.RLock()
	defer eventsMut.RUnlock()
	copy(eventCopy, Events)

	return eventCopy
}
