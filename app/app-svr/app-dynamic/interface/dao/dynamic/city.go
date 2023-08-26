package dynamic

import (
	"context"
	"encoding/json"
	"go-common/library/ecode"
	"go-common/library/log"
	"net/url"
	"strconv"

	dymdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"

	"github.com/pkg/errors"
)

const (
	_drawDetailsV2 = "/link_draw/v0/Doc/dynamicDetailsV2"
)

func (d *Dao) DrawDetails(c context.Context, drawIds []int64) (map[int64]*dymdl.DrawDetailRes, error) {
	params := url.Values{}
	for _, id := range drawIds {
		params.Add("ids[]", strconv.FormatInt(id, 10))
	}
	drawDetails := d.drawDetailsV2
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		// Data []*dymdl.DrawDetailRes `json:"data"`
		Data struct {
			DocItems []struct {
				DocID string `json:"doc_id"`
				Item  string `json:"item"`
			} `json:"doc_items"`
		} `json:"data"`
	}
	if err := d.client.Get(c, drawDetails, "", params, &ret); err != nil {
		log.Errorc(c, "DrawDetails http GET(%s) failed, params:(%s), error(%+v)", drawDetails, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "DrawDetails http GET(%s) failed, params:(%s), code: %v, msg: %v", drawDetails, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "DrawDetails url(%v) code(%v) msg(%v)", drawDetails, ret.Code, ret.Msg)
		return nil, err
	}
	res := make(map[int64]*dymdl.DrawDetailRes)
	for _, d := range ret.Data.DocItems {
		rid, err := strconv.ParseInt(d.DocID, 10, 64)
		if err != nil {
			log.Error("%v", err)
			continue
		}
		var item *dymdl.DrawDetailRes
		err = json.Unmarshal([]byte(d.Item), &item)
		if err != nil {
			log.Error("%v", err)
			continue
		}
		res[rid] = item
	}
	return res, nil
}
