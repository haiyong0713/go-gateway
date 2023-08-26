package region

import "time"

type Region struct {
	ID            int64     `json:"-"`
	DefiniteState int8      `json:"-"`
	DefiniteTime  time.Time `json:"-"`
}
