// +build windows

package winjob_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/windows"

	"github.com/kolesnikovae/go-winjob"
)

const (
	commandName    = "notepad.exe"
	jobTestTimeout = 10 * time.Second
)

func newTestJobObject() (*winjob.JobObject, error) {
	return winjob.Create(fmt.Sprintf("go-winjob-testing-%d", time.Now().UnixNano()))
}

// runTestWithTestJobObjectWithProcess creates a new test job object with
// a process and passes them to the provided jobTestFn.
func runTestWithTestJobObjectWithProcess(t *testing.T, jobTestFn func(*winjob.JobObject, *os.Process)) {
	ctx, cancel := context.WithTimeout(context.Background(), jobTestTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, commandName)
	cmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.CREATE_SUSPENDED,
	}

	requireNoError(t, cmd.Start(), "Starting process")
	job, err := newTestJobObject()
	requireNoError(t, err, "Creating job object")
	defer func() {
		requireNoError(t, job.Close(), "Closing job object")
	}()

	requireNoError(t, job.Assign(cmd.Process), "Assigning process to job object")
	requireNoError(t, winjob.Resume(cmd), "Resuming")
	defer func() {
		requireNoError(t, job.Terminate(), "Terminate test job object")
	}()
	jobTestFn(job, cmd.Process)
}

func runTestWithEmptyJobObject(t *testing.T, jobTestFn func(*winjob.JobObject)) {
	job, err := newTestJobObject()
	requireNoError(t, err, "Creating job object")
	defer func() {
		requireNoError(t, job.Close(), "Closing job object")
	}()
	jobTestFn(job)
}

func requireNoError(t *testing.T, err error, msg ...string) {
	if err != nil {
		if len(msg) > 0 {
			t.Fatalf("%s: %v", msg, err)
		}
		t.Fatalf("%v", err)
	}
}

func TestInvalidJobObjectHandle(t *testing.T) {
	var (
		job      = &winjob.JobObject{Handle: syscall.InvalidHandle}
		process  = new(os.Process)
		counters winjob.Counters
	)
	requireError := func(t *testing.T, err error) {
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	}

	requireError(t, job.Assign(process))
	requireError(t, job.Terminate())
	_, err := job.Contains(process)
	requireError(t, err)
	requireError(t, job.QueryLimits())
	requireError(t, job.QueryCounters(&counters))
	_, err = job.HasLimits()
	requireError(t, err)
	requireError(t, job.ResetLimits())
	requireError(t, job.ResetLimit(winjob.LimitBreakawayOK))
	requireError(t, job.SetLimit(winjob.LimitCPU))
}

func TestCreateWithLimits(t *testing.T) {
	job, err := winjob.Create("", winjob.WithBreakawayOK())
	requireNoError(t, err)
	defer func() {
		requireNoError(t, job.Close())
	}()
	requireNoError(t, job.QueryLimits())
	if !winjob.LimitBreakawayOK.IsSet(job) {
		t.Fatal("Job object limit is not set")
	}
}

func TestTerminate(t *testing.T) {
	runTestWithTestJobObjectWithProcess(t, func(job *winjob.JobObject, p *os.Process) {
		const exitCode = 3
		requireNoError(t, job.TerminateWithExitCode(exitCode))
		s, err := p.Wait()
		requireNoError(t, err)
		if s.ExitCode() != exitCode {
			t.Fatalf("Espected exit code %d, got %d", exitCode, s.ExitCode())
		}
	})
}

func TestContainsProcess(t *testing.T) {
	runTestWithTestJobObjectWithProcess(t, func(job *winjob.JobObject, p *os.Process) {
		contains, err := job.Contains(p)
		requireNoError(t, err)
		if !contains {
			t.Fatal("Job does not contain the process specified")
		}
	})
}

func TestOpenJobObject(t *testing.T) {
	runTestWithTestJobObjectWithProcess(t, func(job *winjob.JobObject, _ *os.Process) {
		_, err := winjob.Open(job.Name)
		requireNoError(t, err)
	})
}

func TestOpenNonexistentJobObject(t *testing.T) {
	if _, err := winjob.Open(time.Now().String()); err == nil {
		t.Fatal("Open: expected error, got nil")
	}
}

func TestCounters(t *testing.T) {
	runTestWithTestJobObjectWithProcess(t, func(job *winjob.JobObject, _ *os.Process) {
		counters, err := job.Counters()
		requireNoError(t, err)
		requireNoError(t, job.QueryCounters(counters))
		if counters.ActiveProcesses == 0 {
			t.Fatal("Empty counters")
		}
	})
}
