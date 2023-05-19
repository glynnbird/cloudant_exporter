package utils

import "time"

// FailBox centralises a way to fail after continuous
// failures for a given amount of time.
type FailBox struct {
	lastSuccess time.Time
	failAfter   time.Duration
	tripped     bool
}

// NewFailBox returns a new FailBox that will trip after
// continuous failures for failAfter time.
func NewFailBox(failAfter time.Duration) *FailBox {
	return &FailBox{
		lastSuccess: time.Now(),
		failAfter:   failAfter,
	}
}

func (fb *FailBox) Success() {
	fb.lastSuccess = time.Now()
}

func (fb *FailBox) Failure() {
	if time.Since(fb.lastSuccess) > fb.failAfter {
		fb.tripped = true
	}
}

func (fb *FailBox) ShouldExit() bool {
	return fb.tripped
}

func (fb *FailBox) LastSuccess() time.Time {
	return fb.lastSuccess
}
