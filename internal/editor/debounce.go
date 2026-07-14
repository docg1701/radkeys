package editor

import "time"

// debouncer coalesces rapid callbacks into a single delayed execution.
// Each call to Add resets the timer, so only the last callback in the
// burst fires after the delay. It is safe for use on a single goroutine
// (the Fyne UI goroutine in this project).
type debouncer struct {
	delay  time.Duration
	timer  *time.Timer
	latest func()
}

// newDebouncer creates a debouncer with the given delay.
func newDebouncer(delay time.Duration) *debouncer {
	return &debouncer{delay: delay}
}

// Add schedules fn to run after the debouncer delay, replacing any
// pending callback from a previous Add call.
func (d *debouncer) Add(fn func()) {
	if d == nil {
		return
	}
	if d.timer != nil {
		d.timer.Stop()
	}
	d.latest = fn
	d.timer = time.AfterFunc(d.delay, func() {
		if d.latest != nil {
			d.latest()
		}
	})
}
