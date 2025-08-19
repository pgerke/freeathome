package freeathome

import "time"

// clock provides an interface for time operations that can be mocked in tests
type clock interface {
	After(d time.Duration) <-chan time.Time
	Now() time.Time
}

// realClock implements clock using the real time package
type realClock struct{}

func (rt *realClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (rt *realClock) Now() time.Time {
	return time.Now()
}
