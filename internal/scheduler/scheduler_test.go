package scheduler

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockJob struct {
	runFunc func() error
}

func (m *mockJob) Run() error {
	if m.runFunc != nil {
		return m.runFunc()
	}
	return nil
}

func TestNewScheduler(t *testing.T) {
	s := NewScheduler(4, 20)

	require.NotNil(t, s)
	assert.Equal(t, 4, s.workers)
	assert.NotNil(t, s.jobs)
	assert.Equal(t, 20, cap(s.jobs))
	assert.NotNil(t, s.stopChan)
}

func TestScheduler_StartStop(t *testing.T) {
	s := NewScheduler(2, 10)
	s.Start()

	// Stop should complete without deadlock
	done := make(chan struct{})
	go func() {
		s.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success: Stop returned
	case <-time.After(2 * time.Second):
		t.Fatal("Stop deadlocked")
	}
}

func TestScheduler_EnqueueAndProcess(t *testing.T) {
	s := NewScheduler(1, 10)
	s.Start()
	defer s.Stop()

	var called int32

	s.Enqueue(&mockJob{
		runFunc: func() error {
			atomic.AddInt32(&called, 1)
			return nil
		},
	})

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&called), "job's Run method should have been called")
}

func TestScheduler_EnqueueMultipleJobs(t *testing.T) {
	s := NewScheduler(2, 10)
	s.Start()
	defer s.Stop()

	var count int32

	for i := 0; i < 5; i++ {
		s.Enqueue(&mockJob{
			runFunc: func() error {
				atomic.AddInt32(&count, 1)
				return nil
			},
		})
	}

	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, int32(5), atomic.LoadInt32(&count), "all 5 jobs should have run")
}

func TestScheduler_EnqueueDropsWhenFull(t *testing.T) {
	s := NewScheduler(1, 1)

	enqueued := make(chan struct{}, 2)

	// First job: goes into the queue buffer (size 1)
	s.Enqueue(&mockJob{
		runFunc: func() error {
			return nil
		},
	})
	enqueued <- struct{}{}

	// Second job: queue is full (buffer taken by first job), should be dropped
	s.Enqueue(&mockJob{
		runFunc: func() error {
			return nil
		},
	})
	// If we reach here, Enqueue did not block (non-blocking send)

	// Start and stop to clean up
	s.Start()
	s.Stop()
}

func TestScheduler_JobError(t *testing.T) {
	s := NewScheduler(1, 10)
	s.Start()
	defer s.Stop()

	var errorJobRan int32
	var successJobRan int32

	// Enqueue a job that returns an error
	s.Enqueue(&mockJob{
		runFunc: func() error {
			atomic.AddInt32(&errorJobRan, 1)
			return errors.New("something went wrong")
		},
	})

	// Enqueue a second job that succeeds
	s.Enqueue(&mockJob{
		runFunc: func() error {
			atomic.AddInt32(&successJobRan, 1)
			return nil
		},
	})

	time.Sleep(200 * time.Millisecond)

	// Scheduler should continue processing after an error
	assert.Equal(t, int32(1), atomic.LoadInt32(&errorJobRan), "error job should have run")
	assert.Equal(t, int32(1), atomic.LoadInt32(&successJobRan), "scheduler should continue processing after error")
}

func TestScheduler_ConcurrentEnqueue(t *testing.T) {
	s := NewScheduler(4, 20)
	s.Start()
	defer s.Stop()

	var count int32

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Enqueue(&mockJob{
				runFunc: func() error {
					atomic.AddInt32(&count, 1)
					return nil
				},
			})
		}()
	}

	wg.Wait()

	time.Sleep(300 * time.Millisecond)

	assert.Equal(t, int32(10), atomic.LoadInt32(&count), "all 10 concurrently enqueued jobs should have run")
}

func TestScheduler_StopWithoutStart(t *testing.T) {
	s := NewScheduler(2, 10)

	// Calling Stop without Start should not panic
	// stopChan is closed and wg.Wait returns immediately since nothing was Added
	assert.NotPanics(t, func() {
		s.Stop()
	})
}
