package clock

import (
	"fmt"
	"time"
)

var location = time.UTC

func Init(name string) error {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return fmt.Errorf("unknown time zone %q: %w", name, err)
	}
	location = loc
	return nil
}

func Location() *time.Location {
	return location
}

type Clock struct {
	now func() time.Time
}

func Fixed(t time.Time) Clock {
	return Clock{now: func() time.Time { return t }}
}

func (c Clock) Now() time.Time {
	if c.now == nil {
		return time.Now()
	}
	return c.now()
}

func (c Clock) Today() Date {
	return DateOf(c.Now())
}
