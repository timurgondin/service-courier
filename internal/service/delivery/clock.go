package delivery

import "time"

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

type FixedClock struct {
	t time.Time
}

func (c FixedClock) Now() time.Time {
	return c.t
}

func NewFixedClock(t time.Time) FixedClock {
	return FixedClock{t: t}
}
