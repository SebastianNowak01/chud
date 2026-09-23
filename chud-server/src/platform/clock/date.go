package clock

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Date struct {
	t time.Time
}

func NewDate(year int, month time.Month, day int) Date {
	return Date{t: time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

func ParseDate(value string) (Date, error) {
	t, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return Date{}, fmt.Errorf("invalid date %q: %w", value, err)
	}
	return NewDate(t.Date()), nil
}

func DateOf(t time.Time) Date {
	return NewDate(t.In(location).Date())
}

func (d Date) Start() time.Time {
	y, m, day := d.t.Date()
	return time.Date(y, m, day, 0, 0, 0, 0, location)
}

func (d Date) String() string {
	return d.t.Format(time.DateOnly)
}

func (d Date) IsZero() bool {
	return d.t.IsZero()
}

func (d Date) Before(other Date) bool {
	return d.t.Before(other.t)
}

func (d Date) After(other Date) bool {
	return d.t.After(other.t)
}

func (d Date) AddDays(days int) Date {
	return Date{t: d.t.AddDate(0, 0, days)}
}

func (d Date) DaysUntil(other Date) int {
	return int(other.t.Sub(d.t).Hours() / 24)
}

func (d Date) Weekday() time.Weekday {
	return d.t.Weekday()
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	parsed, err := ParseDate(value)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func (d Date) Value() (driver.Value, error) {
	return d.String(), nil
}

func (d *Date) Scan(src any) error {
	switch v := src.(type) {
	case time.Time:
		*d = NewDate(v.Date())
		return nil
	case string:
		parsed, err := ParseDate(v)
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	case []byte:
		return d.Scan(string(v))
	default:
		return errors.New("clock.Date: unsupported scan type")
	}
}
