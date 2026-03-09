package handler

import (
	"os"
	"sync"
	"time"
)

const (
	DefaultInactivityMinutes = 5
)

var (
	lastActivity   time.Time
	inactivityMu   sync.Mutex
	inactivityStop chan struct{}
)

func StartInactivityTimer(minutes int) {
	if minutes <= 0 {
		minutes = DefaultInactivityMinutes
	}
	inactivityMu.Lock()
	if inactivityStop != nil {
		close(inactivityStop)
	}
	inactivityStop = make(chan struct{})
	lastActivity = time.Now()
	d := time.Duration(minutes) * time.Minute
	stop := inactivityStop
	inactivityMu.Unlock()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				inactivityMu.Lock()
				elapsed := time.Since(lastActivity)
				inactivityMu.Unlock()
				if elapsed >= d {
					os.Exit(0)
				}
			}
		}
	}()
}

func ResetInactivity() {
	inactivityMu.Lock()
	lastActivity = time.Now()
	inactivityMu.Unlock()
}
