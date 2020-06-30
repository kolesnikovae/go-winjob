// +build windows

package winjob_test

import (
	"log"
	"os/exec"

	"github.com/kolesnikovae/go-winjob"
)

// The example demonstrates an efficient way to ensure no descendant processes
// will be left after the process exit.
//
// LimitKillOnJobClose acts similar to prctl(PR_SET_PDEATHSIG, SIGKILL) in
// Linux: the job is destroyed when its last handle has been closed and all
// associated processes have been terminated.
func Example_commandStart() {
	cmd := exec.Command("cmd.exe", "/c", "ping localhost")
	limits := []winjob.Limit{
		winjob.WithKillOnJobClose(),
		winjob.WithBreakawayOK(),
	}

	// Start creates a job object with the limits specified and starts the
	// given command within the job. The process is created with suspended
	// threads which are resumed when the process has been added to the job
	// object.
	job, err := winjob.Start(cmd, limits...)
	if err != nil {
		log.Fatal(err)
	}

	// If the job has the LimitKillOnJobClose, closing the last job object
	// handle terminates all associated processes and then destroys the job
	// object itself.
	defer job.Close()

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
