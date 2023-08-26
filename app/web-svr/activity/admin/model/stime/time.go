package stime

import (
	"database/sql/driver"
	xtime "go-common/library/time"
	"strconv"
	"time"
)

// Time be used to MySql timestamp converting.
type Time time.Time

// FromString ...
func FromString(src string) Time {
	now, err := time.ParseInLocation(`2006-01-02 15:04:05`, src, time.Local)
	if err != nil {
		return Time(time.Unix(0, 0))
	}
	return Time(now)
}

// Scan scan time.
func (jt *Time) Scan(src interface{}) (err error) {
	switch sc := src.(type) {
	case time.Time:
		*jt = Time(sc)
	case string:
		var i int64
		i, err = strconv.ParseInt(sc, 10, 64)
		*jt = Time(time.Unix(i, 0))
	}
	return
}

// Value get time value.
func (jt Time) Value() (driver.Value, error) {
	return time.Time(jt), nil
}

// Time get time.
func (jt Time) Time() time.Time {
	return time.Time(jt)
}

// Time get time.
func (jt Time) XTime() xtime.Time {
	return xtime.Time(jt.Time().Unix())
}

// MarshalJSON ...
func (jt *Time) MarshalJSON() ([]byte, error) {
	t := jt.Time()
	if t.IsZero() {
		return []byte(`"0000-00-00 00:00:00"`), nil
	}
	return []byte(`"` + jt.Time().Local().Format("2006-01-02 15:04:05") + `"`), nil
}

// UnmarshalJSON ...
func (jt *Time) UnmarshalJSON(data []byte) error {
	now, err := time.ParseInLocation(`2006-01-02 15:04:05`, string(data), time.Local)
	*jt = Time(now)
	return err
}
