package apns

import (
	"container/list"
	"sync"
)

// circular buffer of sent messages
// to retry if connection is dropped
type buffer struct {
	size int
	m    sync.Mutex
	*list.List
}

func newBuffer(size int) *buffer {
	return &buffer{size, sync.Mutex{}, list.New()}
}

func (b *buffer) Add(v interface{}) *list.Element {
	b.m.Lock()
	defer b.m.Unlock()

	e := b.PushBack(v)

	if b.Len() > b.size {
		b.Remove(b.Front())
	}

	return e
}

// NotificationResult associates an error from Apple to a notification.
type NotificationResult struct {
	Notif Notification
	Err   Error
}

func (s NotificationResult) Error() string {
	return s.Err.Error()
}

func (b *buffer) FindFailedNotification(e Error) NotificationResult {
	for cursor := b.Back(); cursor != nil; cursor = cursor.Prev() {
		// Get serialized notification
		n, _ := cursor.Value.(Notification)

		// If the notification, move cursor after the trouble notification
		if n.Identifier == e.Identifier {
			return NotificationResult{n, e}
		}
	}
	return NotificationResult{Notification{}, e}
}

// RequeueableNotifications returns good notifications sent after an error.
func (b *buffer) RequeueableNotifications(identifier uint32) []Notification {
	notifs := []Notification{}

	// Walk back to last known good notification and return the slice
	var e *list.Element
	for e = b.Front(); e != nil; e = e.Next() {
		if n, ok := e.Value.(Notification); ok && n.Identifier == identifier {
			break
		}
	}

	// Start right after errored ID and get the rest of the list
	for e = e.Next(); e != nil; e = e.Next() {
		n, ok := e.Value.(Notification)
		if !ok {
			continue
		}

		notifs = append(notifs, n)
	}

	return notifs
}
