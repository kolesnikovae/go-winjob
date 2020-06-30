// +build windows

package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"golang.org/x/sys/windows"

	"github.com/kolesnikovae/go-winjob"
)

var limits = []winjob.Limit{
	winjob.WithBreakawayOK(),
	winjob.WithKillOnJobClose(),
	winjob.WithActiveProcessLimit(3),
	winjob.WithProcessTimeLimit(10 * time.Second),
	winjob.WithCPUHardCapLimit(500),         // 5%
	winjob.WithProcessMemoryLimit(10 << 20), // 10MB
	winjob.WithWriteClipboardLimit(),
}

const defaultCommand = "notepad.exe"

func createCommand() *exec.Cmd {
	switch len(os.Args) {
	case 1:
		return exec.Command(defaultCommand)
	case 2:
		return exec.Command(os.Args[1])
	default:
		return exec.Command(os.Args[1], os.Args[1:]...)
	}
}

func main() {
	job, err := winjob.Create("", limits...)
	if err != nil {
		log.Fatalf("Create: %v", err)
	}

	cmd := createCommand()
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.CREATE_SUSPENDED,
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("Start: %v", err)
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)

	c := make(chan winjob.Notification)
	subscription, err := winjob.Notify(c, job)
	if err != nil {
		log.Fatalf("Notify: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		var counters winjob.Counters
		for {
			select {
			case <-s:
				log.Println("Closing job object")
				if err := job.Close(); err != nil {
					log.Fatal(err)
				}
				log.Println("Closing subscription")
				if err := subscription.Close(); err != nil {
					log.Fatal(err)
				}
				return

			case n, ok := <-c:
				if ok {
					log.Printf("Notification: %#v\n", n)
				} else if err := subscription.Err(); err != nil {
					log.Fatalf("Subscription: %v", err)
				}

			case <-ticker.C:
				if err := job.QueryCounters(&counters); err != nil {
					log.Fatalf("QueryCounters: %v", err)
				}
				b, err := json.MarshalIndent(counters, "", "\t")
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Counters: \n%s\n", b)
			}
		}
	}()

	if err := job.Assign(cmd.Process); err != nil {
		log.Fatalf("Assign: %v", err)
	}

	if err := winjob.Resume(cmd); err != nil {
		log.Fatalf("Resume: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalf("Wait: %v", err)
	}

	// Wait for a signal.
	<-done
}
