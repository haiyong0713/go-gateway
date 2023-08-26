package bnj

import (
	"context"
	"time"

	"go-common/library/database/elastic"
	"go-gateway/app/web-svr/activity/admin/model/bnj"
)

const _samplePS = 1

// LiveGift check mid has send live gift.
func (d *Dao) LiveGift(c context.Context, mid, roomID int64, indexes []string, timeFrom, timeTo time.Time) (result *bnj.Result, err error) {
	r := d.esClient.NewRequest("log_user_action").Index(indexes...)
	r.Fields("ctime", "int_0", "ip", "mid", "oid")
	r.WhereEq("mid", mid).WhereEq("int_0", roomID)
	r.WhereRange("ctime", timeFrom.Format("2006-01-02 15:04:05"), timeTo.Format("2006-01-02 15:04:05"), elastic.RangeScopeLoRo)
	r.Ps(_samplePS)
	err = r.Scan(c, &result)
	return
}
