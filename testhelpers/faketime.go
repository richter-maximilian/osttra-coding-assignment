package testhelpers

import "time"

type FakeTime struct {
	now         time.Time
	transformFn func(now time.Time) time.Time
}

func NewFakeTime(now time.Time) *FakeTime {
	return &FakeTime{
		now:         now,
		transformFn: func(now time.Time) time.Time { return now },
	}
}

func (t *FakeTime) Now() time.Time {
	return t.transformFn(t.now)
}
