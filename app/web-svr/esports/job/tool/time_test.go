package tool

import (
	"testing"
	"time"
)

func TestDateSub(t *testing.T) {
	t.Run("date sub in the same day", testSameDay)
	t.Run("date sub in one week", testOneWeekSub)
}

func testSameDay(t *testing.T) {
	now := time.Now()
	t1 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, now.Nanosecond(), now.Location())

	if d := CalculateDateSub(t1, now); d != 0 {
		t.Errorf("date sub between %v and %v should as %v, but now %v", t1, now, 0, d)
	}
}

func testOneWeekSub(t *testing.T) {
	now := time.Now()
	t1 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, now.Nanosecond(), now.Location())
	t2 := t1.Add(-time.Hour * 24 * 7).Add(time.Hour * 1).Add(time.Minute * 120)

	if d := CalculateDateSub(t2, t1); d != 7 {
		t.Errorf("date sub between %v and %v should as %v, but now %v", t1, t2, 7, d)
	}
}
