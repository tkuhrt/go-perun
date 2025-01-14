// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"runtime"
	"strconv"
	"sync"
	stdatomic "sync/atomic"

	"github.com/stretchr/testify/require"

	"perun.network/go-perun/pkg/sync/atomic"
)

// stage is a single stage of execution in a concurrent test.
type stage struct {
	name      string      // The stage's name.
	failed    atomic.Bool // Whether a stage failed.
	spawnOnce sync.Once

	wg    sync.WaitGroup // Stage barrier.
	wgN   int            // Used to detect spawn() calls with wrong N.
	count int32          // The number of instances.

	require.TestingT // The stage's test object.

	ct *ConcurrentT // The concurrent testing object.
}

// wait waits until a stage is terminated and then returns whether it succeeded.
func (s *stage) wait() bool {
	s.wg.Wait()
	return !s.failed.IsSet()
}

// spawn sets up a stage when it is spawned.
// Checks that the stage is not spawned multiple times.
func (s *stage) spawn(n int) {
	s.spawnOnce.Do(func() {
		s.wg.Add(n - 1)
		s.wgN = n
	})

	if n != s.wgN {
		panic("spawned stage '" + s.name + "' with inconsistent N: " +
			strconv.Itoa(n) + " vs. " + strconv.Itoa(s.wgN))
	}
	if int(stdatomic.AddInt32(&s.count, 1)) > s.wgN {
		panic("spawned stage '" + s.name + "' too often")
	}
}

// pass marks the stage as passed and waits until it is complete.
func (s *stage) pass() {
	s.wg.Done()
}

// FailNow marks the stage as failed and terminates the goroutine.
func (s *stage) FailNow() {
	s.failed.Set()
	s.wg.Done()
	s.ct.FailNow()
}

// ConcurrentT is a testing object that can be used in multiple goroutines.
// Specifically, using the helper objects created by the Stage/StageN calls,
// FailNow can be called by any goroutine (however, the helper objects must not
// be used in multiple goroutines).
type ConcurrentT struct {
	failNowMutex sync.Mutex
	t            require.TestingT
	failed       bool

	mutex  sync.Mutex
	stages map[string]*stage
}

// NewConcurrent creates a new concurrent testing object.
func NewConcurrent(t require.TestingT) *ConcurrentT {
	return &ConcurrentT{
		t:      t,
		stages: make(map[string]*stage),
	}
}

// spawnStage retrieves/creates a stage and sets it up.
func (t *ConcurrentT) spawnStage(name string, n int) *stage {
	stage := t.getStage(name)
	stage.spawn(n)
	return stage
}

// getStage retrieves and existing stage or creates a new one, if it does not
// exist yet.
func (t *ConcurrentT) getStage(name string) *stage {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if s, ok := t.stages[name]; ok {
		return s
	}

	s := &stage{name: name, TestingT: t.t, ct: t}
	s.wg.Add(1)
	t.stages[name] = s
	return s
}

// Wait waits until the stages with the requested names terminate.
// If any stage fails, terminates the current goroutine.
func (t *ConcurrentT) Wait(names ...string) {
	if len(names) == 0 {
		panic("Wait(): called with 0 names")
	}

	for _, name := range names {
		if !t.getStage(name).wait() {
			runtime.Goexit()
		}
	}
}

// FailNow fails and aborts the test.
func (t *ConcurrentT) FailNow() {
	t.failNowMutex.Lock()
	defer t.failNowMutex.Unlock()
	if !t.failed {
		t.failed = true
		t.t.FailNow()
	} else {
		runtime.Goexit()
	}
}

// StageN creates a named execution stage.
// The parameter goroutines specifies the number of goroutines that share the
// stage. This number must be consistent across all StageN calls with the same
// stage name and exactly match the number of times StageN is called for that
// name.
// Executes fn. If fn calls FailNow on the supplied T object, the stage fails.
// fn must not spawn any goroutines or pass along the T object to goroutines
// that call T.Fatal. To achieve this, make other goroutines call
// ConcurrentT.StageN() instead.
func (t *ConcurrentT) StageN(name string, goroutines int, fn func(require.TestingT)) {
	stage := t.spawnStage(name, goroutines)
	aborted := true
	// Detect whether fn terminates via panic() or direct calls to
	// runtime.Goexit(). Mark the stage as done if it is not yet marked. Call
	// FailNow() on the test.
	defer func() {
		if aborted {
			if stage.failed.TrySet() {
				defer stage.wg.Done()
			}
			t.FailNow()
		}
	}()

	fn(stage)

	aborted = false

	stage.pass()
	t.Wait(name)
}

// Stage creates a named execution stage.
// It is a shorthand notation for StageN(name, 1, fn).
func (t *ConcurrentT) Stage(name string, fn func(require.TestingT)) {
	t.StageN(name, 1, fn)
}
