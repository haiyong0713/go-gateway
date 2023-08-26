package selected

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
)

// DurationString duration to string
func DurationString(second int64) (s string) {
	var hour, min, sec int
	if second < 1 {
		return
	}
	d, err := time.ParseDuration(strconv.FormatInt(second, 10) + "s")

	if err != nil {
		log.Error("%+v", err)
		return
	}
	r := strings.NewReplacer("h", ":", "m", ":", "s", ":")
	ts := strings.Split(strings.TrimSuffix(r.Replace(d.String()), ":"), ":")
	//nolint:gomnd
	if len(ts) == 1 {
		sec, _ = strconv.Atoi(ts[0])
	} else if len(ts) == 2 {
		min, _ = strconv.Atoi(ts[0])
		sec, _ = strconv.Atoi(ts[1])
	} else if len(ts) == 3 {
		hour, _ = strconv.Atoi(ts[0])
		min, _ = strconv.Atoi(ts[1])
		sec, _ = strconv.Atoi(ts[2])
	}
	if hour == 0 {
		s = fmt.Sprintf("%d:%02d", min, sec)
		return
	}
	s = fmt.Sprintf("%d:%02d:%02d", hour, min, sec)
	return
}

// ArchiveViewString ArchiveView to string
func ArchiveViewString(number int32) string {
	const _suffix = "观看"
	return StatString(number, _suffix)
}

// StatString Stat to string
func StatString(number int32, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	//nolint:gomnd
	if number < 10000 {
		s = strconv.FormatInt(int64(number), 10) + suffix
		return
	}
	//nolint:gomnd
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

// PubDataString is.
func PubDataString(t time.Time) (s string) {
	if t.IsZero() {
		return
	}
	now := time.Now()
	sub := now.Sub(t)
	if sub < time.Minute {
		s = "刚刚"
		return
	}
	if sub < time.Hour {
		s = strconv.FormatFloat(sub.Minutes(), 'f', 0, 64) + "分钟前"
		return
	}
	if sub < 24*time.Hour {
		s = strconv.FormatFloat(sub.Hours(), 'f', 0, 64) + "小时前"
		return
	}
	if now.Year() == t.Year() {
		if now.YearDay()-t.YearDay() == 1 {
			s = "昨天"
			return
		}
		s = t.Format("01-02")
		return
	}
	s = t.Format("2006-01-02")
	return
}
