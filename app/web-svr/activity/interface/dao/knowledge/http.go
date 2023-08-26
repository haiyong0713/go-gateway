package knowledge

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/log"
	knowledge "go-gateway/app/web-svr/activity/interface/model/knowledge"
)

// BvidList 获取关注mid
func (d *Dao) BvidList(ctx context.Context, uri string, timeStamp int64) (res *knowledge.Period, err error) {
	params := url.Values{}
	params.Set("t", strconv.FormatInt(timeStamp, 10))
	res = new(knowledge.Period)
	if err = d.singleClient.Get(ctx, uri, "", params, &res); err != nil {
		log.Errorc(ctx, "BvidList d.client.Get(%s) error(%+v)", uri+"?"+params.Encode(), err)
		return
	}
	log.Infoc(ctx, "BvidList d.client.Get url(%s) params (%v) res(%v)", uri, params.Encode(), res)
	return
}
