package model

import (
	"database/sql/driver"
	"time"
)

type DateTime string

// Scan scan time.
func (jt *DateTime) Scan(src interface{}) (err error) {
	switch sc := src.(type) {
	case time.Time:
		*jt = DateTime(sc.Format("2006-01-02 15:04:05"))
	case string:
		*jt = DateTime(sc)
	}
	return
}

// Value get time value.
func (jt DateTime) Value() (driver.Value, error) {
	return time.Parse("2006-01-02 15:04:05", string(jt))
}
