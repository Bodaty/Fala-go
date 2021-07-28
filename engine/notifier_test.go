package engine

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

// TestNotifier_PassByValue verifies that passing Notifier by value is safe
func TestNotifier_PassByValue(t *testing.T) {
	t.Parallel()
	notifier := NewNotifier()

	var sent sync.WaitGroup
	sent.Add(1)
	go func(n Notifier) {
		notifier.Notify()
		sent.Done()
	}(notifier)
	sent.Wait()

	select {
	case <-notifier.Channel(): // expected
	default:
		t.Fail()
	}
}

// TestNotifier_NoNotificationsAtStartup verifies that Notifier is initialized
// without notifications
func TestNotifier_NoNotificationsInitialization(t *testing.T) {
	t.Parallel()
	notifier := NewNotifier()
	select {
	case <-notifier.Channel():
		t.Fail()
	default: //expected
	}
}

// TestNotifier_ManyNotifications sends many notifications to the Notifier
// and verifies that:
//  * the notifier accepts them all without a notification being consumed
//  * only one notification is internally stored and subsequent attempts to
//    read a notification would block
func TestNotifier_ManyNotifications(t *testing.T) {
	t.Parallel()
	notifier := NewNotifier()

	var counter sync.WaitGroup
	for i := 0; i < 10; i++ {
		counter.Add(1)
		go func() {
			notifier.Notify()
			counter.Done()
		}()
	}
	counter.Wait()

	// attempt to consume first notification:
	// expect that one notification should be available
	c := notifier.Channel()
	select {
	case <-c: // expected
	default:
		t.Fail()
	}

	// attempt to consume first notification
	// expect that no notification is available
	select {
	case <-c:
		t.Fail()
	default: //expected
	}
}

// TestNotifier_ManyConsumers spans many worker routines and
// sends just as many notifications with small delays. We require that
// all workers eventually get a notification.
func TestNotifier_ManyConsumers(t *testing.T) {
	t.Parallel()
	notifier := NewNotifier()
	c := notifier.Channel()

	// spawn 100 worker routines to each wait for a notification
	var startingWorkers sync.WaitGroup
	pendingWorkers := atomic.NewInt32(100)
	for i := 0; i < 100; i++ {
		startingWorkers.Add(1)
		go func() {
			startingWorkers.Done()
			<-c
			pendingWorkers.Dec()
		}()
	}
	startingWorkers.Wait()
	time.Sleep(1 * time.Millisecond)

	// send 100 notifications, with small delays
	for i := 0; i < 100; i++ {
		notifier.Notify()
		time.Sleep(10 * time.Microsecond)
	}

	nPendingWorkers := pendingWorkers.Load()
	// require that all workers got a notification
	require.Eventuallyf(t,
		func() bool {
			nPendingWorkers = pendingWorkers.Load()
			return nPendingWorkers == 0
		},
		3*time.Second, 100*time.Millisecond,
		"still awaiting %d workers to get notification", nPendingWorkers,
	)
}

// TestNotifier_AllWorkProcessed spans many worker routines and
// sends just as many notifications with small delays. We require that
// all workers eventually get a notification.
func TestNotifier_AllWorkProcessed(t *testing.T) {
	singleTestRun := func(t *testing.T) {
		t.Parallel()
		notifier := NewNotifier()

		totalWork := int32(100)
		pendingWorkQueue := make(chan struct{}, totalWork)
		scheduledWork := atomic.NewInt32(0)
		consumedWork := atomic.NewInt32(0)

		// starts the consuming first, because if we starts the production first instead, then
		// we might finish pushing all jobs, before any of our consumer has started listening
		// to the queue.
		// 5 routines consuming work
		var consumersAllReady sync.WaitGroup
		consumersAllReady.Add(5)
		for i := 0; i < 5; i++ {
			go func() {
				consumersAllReady.Done()
				for consumedWork.Load() < totalWork {
					<-notifier.Channel()
					for {
						select {
						case <-pendingWorkQueue:
							consumedWork.Inc()
						default:
							break
						}
					}
				}
			}()
		}

		// wait long enough for all consumer to be ready for new notification.
		consumersAllReady.Wait()

		var workersAllReady sync.WaitGroup
		workersAllReady.Add(10)

		// 10 routines pushing work
		for i := 0; i < 10; i++ {
			go func() {
				workersAllReady.Done()
				for scheduledWork.Inc() <= totalWork {
					pendingWorkQueue <- struct{}{}
					notifier.Notify()
				}
			}()
		}

		// wait long enough for all workers to be started.
		workersAllReady.Wait()

		nConsumedWork := consumedWork.Load()
		// require that all work is eventually consumed
		require.Eventuallyf(t,
			func() bool {
				nConsumedWork = consumedWork.Load()
				return nConsumedWork == totalWork
			},
			3*time.Second, 100*time.Millisecond,
			"only consumed %d units of work but expecting %d", nConsumedWork, totalWork,
		)
	}

	for r := 0; r < 100; r++ {
		t.Run(fmt.Sprintf("run %d", r), singleTestRun)
	}
}
