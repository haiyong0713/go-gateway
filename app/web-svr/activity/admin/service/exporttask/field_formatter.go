package exporttask

import "time"

func formatTimeString(s string) string {
	t, e := time.Parse(time.RFC3339, s)
	if e != nil {
		return s
	}
	return t.Format("2006-01-02 15:04:05")
}
