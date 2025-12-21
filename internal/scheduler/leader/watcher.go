package leader

import (
	"context"
	"time"
)

type Watcher struct {
	election    *Election
	checkPeriod time.Duration
	onAcquire   func()
	onLose      func()
}

func NewWatcher(election *Election, checkPeriod time.Duration) *Watcher {
	if checkPeriod <= 0 {
		checkPeriod = 10 * time.Second
	}
	return &Watcher{
		election:    election,
		checkPeriod: checkPeriod,
	}
}

func (w *Watcher) OnAcquire(fn func()) *Watcher {
	w.onAcquire = fn
	return w
}

func (w *Watcher) OnLose(fn func()) *Watcher {
	w.onLose = fn
	return w
}

func (w *Watcher) Watch(ctx context.Context) {
	ticker := time.NewTicker(w.checkPeriod)
	defer ticker.Stop()

	wasLeader := false

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			isLeader := w.election.IsLeader()

			if isLeader && !wasLeader && w.onAcquire != nil {
				w.onAcquire()
			} else if !isLeader && wasLeader && w.onLose != nil {
				w.onLose()
			}

			wasLeader = isLeader
		}
	}
}
