// +build windows

package winjob_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/kolesnikovae/go-winjob"
	"github.com/kolesnikovae/go-winjob/jobapi"
)

var (
	errLimitNotSet   = errors.New("limit not set")
	errLimitNotReset = errors.New("limit not reset")
	errJobHasLimits  = errors.New("job has limits")
)

type limitCase struct {
	limit    winjob.Limit
	expected interface{}
}

// There are special cases for:
//   - TestLimits_WorkingSetSizeLimit
//   - TestLimits_PreserveJobTimeLimit
//   - TestLimits_AffinityLimit
//   - TestLimits_CPULimit
var limitCases = []limitCase{
	{winjob.WithBreakawayOK(), true},
	{winjob.WithSilentBreakawayOK(), true},
	{winjob.WithDieOnUnhandledException(), true},
	{winjob.WithKillOnJobClose(), true},

	{winjob.WithDesktopLimit(), true},
	{winjob.WithDisplaySettingsLimit(), true},
	{winjob.WithExitWindowsLimit(), true},
	{winjob.WithGlobalAtomsLimit(), true},
	{winjob.WithHandlesLimit(), true},
	{winjob.WithReadClipboardLimit(), true},
	{winjob.WithSystemParametersLimit(), true},
	{winjob.WithWriteClipboardLimit(), true},

	{
		winjob.WithAffinity(1),
		uintptr(1),
	},
	{
		winjob.WithJobMemoryLimit(8192 << 10),
		uintptr(8192 << 10),
	},
	{
		winjob.WithJobTimeLimit(time.Second * 10),
		time.Second * 10, // May be flaky
	},
	{
		winjob.WithProcessMemoryLimit(8192 << 10),
		uintptr(8192 << 10),
	},
	{
		winjob.WithProcessTimeLimit(time.Second * 10),
		time.Second * 10, // May be flaky
	},
	{
		winjob.WithActiveProcessLimit(2),
		uint32(2),
	},
	{
		winjob.WithWorkingSetLimit(uintptr(8192<<10), uintptr(8<<20)),
		true, // Refer to TestLimits_WorkingSetSizeLimit
	},
	{
		winjob.WithSchedulingClassLimit(4),
		uint32(4),
	},
	{
		winjob.WithPriorityClassLimit(jobapi.ABOVE_NORMAL_PRIORITY_CLASS),
		jobapi.ABOVE_NORMAL_PRIORITY_CLASS,
	},

	{
		winjob.WithCPUHardCapLimit(5000),
		winjob.CPURate{HardCap: 5000},
	},
	{
		winjob.WithOutgoingBandwidthLimit(1 << 20),
		uint64(1 << 20),
	},
	{
		winjob.WithDSCPTag(0x4),
		byte(0x4),
	},
}

func (c *limitCase) print(t *testing.T, msg string) {
	t.Logf("%s %[2]T: %+[2]v", msg, c.limit)
}

func (c *limitCase) set(t *testing.T, job *winjob.JobObject) {
	c.print(t, "Applying")
	requireNoError(t, job.SetLimit(c.limit))
}

func (c *limitCase) requireSet(t *testing.T, job *winjob.JobObject) {
	c.print(t, "Require set")
	if !c.limit.IsSet(job) {
		t.Fatal(errLimitNotSet)
	}
	actual := c.limit.Value(job)
	if !reflect.DeepEqual(actual, c.expected) {
		t.Fatalf("Value missmatch: got %#[1]v [%[1]T], expected: %#[2]v [%[2]T]", actual, c.expected)
	}
}

func (c *limitCase) reset(t *testing.T, job *winjob.JobObject) {
	c.print(t, "Resetting")
	requireNoError(t, job.ResetLimit(c.limit))
	if c.limit.IsSet(job) {
		t.Fatal(errLimitNotReset)
	}
}

func (c *limitCase) requireReset(t *testing.T, job *winjob.JobObject) {
	c.print(t, "Require reset")
	requireNoError(t, job.ResetLimit(c.limit))
	if c.limit.IsSet(job) {
		t.Fatal(errLimitNotReset)
	}
}

// TestLimits_Single validates limits independently each other:
// a new job object is created for every limit.
func TestLimits_Single(t *testing.T) {
	for _, x := range limitCases {
		x := x // var pinning for scopelint false-positive
		runTestWithEmptyJobObject(t, func(job *winjob.JobObject) {
			x.set(t, job)
			requireNoError(t, job.QueryLimits())
			x.requireSet(t, job)
			x.reset(t, job)
		})
	}
}

// TestLimits_Multiple validates limits all together:
// the same job object is used.
func TestLimits_Multiple(t *testing.T) {
	runTestWithEmptyJobObject(t, func(job *winjob.JobObject) {
		for _, x := range limitCases {
			x.set(t, job)
		}
		requireNoError(t, job.QueryLimits())
		for _, x := range limitCases {
			x.requireSet(t, job)
			x.reset(t, job)
		}
	})
}

func limitPreset(limitCases []limitCase) []winjob.Limit {
	preset := make([]winjob.Limit, len(limitCases))
	for i, x := range limitCases {
		preset[i] = x.limit
	}
	return preset
}

// TestLimits_MultiplePreset validates limit setting all together at a time.
func TestLimits_MultiplePreset(t *testing.T) {
	runTestWithEmptyJobObject(t, func(job *winjob.JobObject) {
		requireNoError(t, job.SetLimit(limitPreset(limitCases)...))
		requireNoError(t, job.QueryLimits())
		for _, x := range limitCases {
			x.requireSet(t, job)
			x.reset(t, job)
		}
	})
}

func jobHasLimitSubTest(t *testing.T, job *winjob.JobObject, expected bool) {
	limited, err := job.HasLimits()
	requireNoError(t, err)
	if limited && !expected {
		t.Fatal(errJobHasLimits)
	}
}

func TestLimits_HasLimits(t *testing.T) {
	runTestWithEmptyJobObject(t, func(job *winjob.JobObject) {
		jobHasLimitSubTest(t, job, false)
		requireNoError(t, job.SetLimit(limitCases[0].limit))
		requireNoError(t, job.QueryLimits())
		jobHasLimitSubTest(t, job, true)
	})
}

func TestLimits_ResetLimits(t *testing.T) {
	runTestWithEmptyJobObject(t, func(job *winjob.JobObject) {
		requireNoError(t, job.ResetLimits())
		for _, x := range limitCases {
			x.set(t, job)
		}
		requireNoError(t, job.ResetLimits())
		requireNoError(t, job.QueryLimits())
		jobHasLimitSubTest(t, job, false)
	})
}

// LimitWorkingSet modifies two values at once.
func TestLimits_WorkingSetSizeLimit(t *testing.T) {
	runTestWithEmptyJobObject(t, func(job *winjob.JobObject) {
		wsLimit := winjob.LimitWorkingSet.WithValue(1<<20, 8<<20)
		x := limitCase{limit: wsLimit, expected: true}
		x.set(t, job)
		requireNoError(t, job.QueryLimits())
		x.requireSet(t, job)
		if wsLimit.MinWorkingSetSize(job) != 1<<20 {
			t.Fatalf("Incorrect MinWorkingSetSize")
		}
		if wsLimit.MaxWorkingSetSize(job) != 8<<20 {
			t.Fatalf("Incorrect MaxWorkingSetSize")
		}
		requireNoError(t, job.QueryLimits())
		x.reset(t, job)
	})
}

// JOB_OBJECT_LIMIT_PRESERVE_JOB_TIME cannot be used with JOB_OBJECT_LIMIT_JOB_TIME.
func TestLimits_PreserveJobTimeLimit(t *testing.T) {
	runTestWithEmptyJobObject(t, func(job *winjob.JobObject) {
		requireNoError(t, job.SetLimit(winjob.WithJobTimeLimit(time.Second*10)))
		requireNoError(t, job.SetLimit(
			winjob.WithPreserveJobTime(),
			winjob.WithAffinity(1)))
		requireNoError(t, job.QueryLimits())
	})
}

// JOB_OBJECT_LIMIT_SUBSET_AFFINITY depends on JOB_OBJECT_LIMIT_AFFINITY.
func TestLimits_AffinityLimit(t *testing.T) {
	runTestWithEmptyJobObject(t, func(job *winjob.JobObject) {
		affinityLimitCases := []limitCase{
			{winjob.WithSubsetAffinity(), true},
			{winjob.WithAffinity(1), uintptr(1)},
		}
		requireNoError(t, job.SetLimit(limitPreset(affinityLimitCases)...))
		for _, x := range affinityLimitCases {
			x.requireSet(t, job)
			x.reset(t, job)
		}
	})
}

// Only one CPU limit can be applied to a job object at a time.
func TestLimits_CPULimit(t *testing.T) {
	runTestWithEmptyJobObject(t, func(job *winjob.JobObject) {
		testCases := []limitCase{
			{winjob.WithCPUHardCapLimit(500), winjob.CPURate{HardCap: 500}},
			{winjob.WithCPUWeightedLimit(7), winjob.CPURate{Weight: 7}},
			{winjob.WithCPUMinMaxLimit(500, 1000), winjob.CPURate{Min: 500, Max: 1000}},
		}
		for _, x := range testCases {
			x.set(t, job)
			requireNoError(t, job.QueryLimits())
			x.requireSet(t, job)
			x.reset(t, job)
			requireNoError(t, job.QueryLimits())
			x.requireReset(t, job)
		}
	})
}
