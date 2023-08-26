package util

import (
	"time"

	"go-common/library/queue/databus/report"
)

const (
	Business = 204
	Type     = 10031
	Action   = "publish"
)

// AddLogs add action logs
func AddLogs(uname string, obj interface{}) (err error) {
	_ = report.Manager(&report.ManagerInfo{
		Uname:    uname,
		UID:      0,
		Business: Business,
		Type:     Type,
		Oid:      0,
		Action:   Action,
		Ctime:    time.Now(),
		// extra
		Index: []interface{}{},
		Content: map[string]interface{}{
			"json": obj,
		},
	})
	return
}
