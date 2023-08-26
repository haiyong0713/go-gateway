package model

import (
	"database/sql/driver"
	"time"
)

type StrTime string

// Scan scan time.
func (jt *StrTime) Scan(src interface{}) (err error) {
	switch sc := src.(type) {
	case time.Time:
		*jt = StrTime(sc.Format("2006-01-02 15:04:05"))
	case string:
		*jt = StrTime(sc)
	}
	return
}

// Value get time value.
func (jt StrTime) Value() (driver.Value, error) {
	return time.Parse("2006-01-02 15:04:05", string(jt))
}
