package manager

import (
	"context"
	"fmt"
	"sync"
)

// Add adds a runnable to closest group of runnable that they belong to.
//
// Add should be able to be called before and after Start, but not after StopAndWait.
// Add should return an error when called during StopAndWait.
// The runnables added before Start are started when Start is called.
// The runnables added after Start are started directly.
func (r *runnables) Add(fn Runnable) error {
	switch runnable := fn.(type) {
	case *httpServer:
		return r.HTTPServers.Add(fn, nil)
	case *grpcServer:
		return r.GRPCServers.Add(fn, nil)
	default:
		fmt.Println("runnable group add ", runnable)
		//return r.LeaderElection.Add(fn, nil)
		return nil

	}
}

// Add should be able to be called before and after Start, but not after StopAndWait.
// Add should return an error when called during StopAndWait.
func (r *runnableGroup) Add(rn Runnable, ready runnableCheck) error {
	r.stop.RLock()
	if r.stopped {
		r.stop.RUnlock()
		return errRunnableGroupStopped
	}
	r.stop.RUnlock()

	if ready == nil {
		ready = func(_ context.Context) bool { return true }
	}

	readyRunnable := &readyRunnable{
		Runnable: rn,
		Check:    ready,
	}

	// Handle start.
	// If the overall runnable group isn't started yet
	// we want to buffer the runnables and let Start()
	// queue them up again later.
	{
		r.start.Lock()

		// Check if we're already started.
		if !r.started {
			// Store the runnable in the internal if not.
			r.startQueue = append(r.startQueue, readyRunnable)
			//r.stopQueue = append(r.stopQueue, readyRunnable)
			r.start.Unlock()
			return nil
		}
		r.start.Unlock()
	}

	// Enqueue the runnable.
	r.ch <- readyRunnable
	return nil
}

// newRunnables creates a new runnables object.
func newRunnables(baseContext BaseContextFunc, errChan chan error) *runnables {
	return &runnables{
		HTTPServers: newRunnableGroup(baseContext, errChan),
		GRPCServers: newRunnableGroup(baseContext, errChan),
	}
}

func newRunnableGroup(baseContext BaseContextFunc, errChan chan error) *runnableGroup {
	r := &runnableGroup{
		startReadyCh: make(chan *readyRunnable),
		errChan:      errChan,
		ch:           make(chan *readyRunnable),
		wg:           new(sync.WaitGroup),
	}

	r.ctx, r.cancel = context.WithCancel(baseContext())
	return r
}
