package tool

import (
	"time"
)

// Calculate the sub between two dates
func CalculateDateSub(t1, t2 time.Time) int {
	if t2.After(t1) {
		tmp := t1
		t1 = t2
		t2 = tmp
	}

	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t2.Location())

	return int(t1.Sub(t2).Hours() / 24)
}
