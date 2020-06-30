// +build windows

package winjob_test

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/kolesnikovae/go-winjob"
)

const notificationsTestLimit = time.Second * 3

func TestNotifications(t *testing.T) {
	runTestWithTestJobObjectWithProcess(t, func(job *winjob.JobObject, p *os.Process) {
		c := make(chan winjob.Notification, 1)
		s, err := winjob.Notify(c, job)
		defer func() {
			requireNoError(t, s.Close())
			requireNoError(t, s.Err())
		}()
		requireNoError(t, err)
		requireNoError(t, p.Kill())
		select {
		case n, ok := <-c:
			// We expect at least NewProcess/ActiveProcessZero
			// notification, whichever occurs first.
			if !ok {
				t.Fatal("Notification channel is closed")
			}
			t.Logf("Notification: %#v", n)
		case <-time.After(notificationsTestLimit):
			t.Fatal("No notifications received")
		}
	})
}

// The test ensures that the notification channel is closed
// with close of the subscription created.
func TestNotifications_Interruption(t *testing.T) {
	runTestWithTestJobObjectWithProcess(t, func(job *winjob.JobObject, p *os.Process) {
		c := make(chan winjob.Notification, 1)
		s, err := winjob.Notify(c, job)
		requireNoError(t, err)
		requireNoError(t, s.Close())
		requireNoError(t, s.Err())
		select {
		case n, ok := <-c:
			if ok {
				t.Logf("Notification: %#v", n)
			}
		case <-time.After(notificationsTestLimit):
			t.Fatal("No notifications received")
		}
	})
}

// The test ensures that the notification channel is closed on completion
// port error and the error can be retrieved by Err call.
func TestNotifications_Error(t *testing.T) {
	runTestWithTestJobObjectWithProcess(t, func(job *winjob.JobObject, p *os.Process) {
		c := make(chan winjob.Notification, 1)
		s, err := winjob.Notify(c, job)
		requireNoError(t, err)
		requireNoError(t, s.Port.Close())
		select {
		case n, ok := <-c:
			if ok {
				t.Logf("Notification: %#v", n)
			}
		case <-time.After(notificationsTestLimit):
			t.Fatal("No notifications received")
		}
		expectedError := syscall.Errno(0x6) // The handle is invalid.
		if !errors.Is(s.Err(), expectedError) {
			t.Fatalf("Expected %#v, got %#v", expectedError, s.Err())
		}
	})
}
