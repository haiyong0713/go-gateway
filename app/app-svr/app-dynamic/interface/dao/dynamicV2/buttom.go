package dynamicV2

import (
	"context"
	"encoding/json"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	"github.com/pkg/errors"
)

const (
	_buttomFeedInfo = "/bottom_svr/v0/bottom_svr/feed_info"
)

func (d *Dao) BottomFeedInfo(c context.Context, bus []*dynamicV2.BottomBusiness) (map[int64]*dynamicV2.BottomDetail, error) {
	buttomURL := d.c.Hosts.VcCo + _buttomFeedInfo
	params := url.Values{}
	busm := map[string]interface{}{
		"business": bus,
	}
	bjson, _ := json.Marshal(busm)
	params.Set("input_extend", string(bjson))
	var res struct {
		Code int `json:"code"`
		Data struct {
			OutputExtend string `json:"output_extend"`
		} `json:"data"`
	}
	if err := d.client.Get(c, buttomURL, "", params, &res); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if res.Code != 0 {
		err := errors.Wrapf(ecode.Int(res.Code), "bottomFeedInfo url(%s) code(%d)", buttomURL, res.Code)
		return nil, err
	}
	buttom := &dynamicV2.ButtomFeedInfo{}
	if err := json.Unmarshal([]byte(res.Data.OutputExtend), &buttom); err != nil {
		return nil, err
	}
	data := map[int64]*dynamicV2.BottomDetail{}
	for _, v := range buttom.BottomDetails {
		if v != nil {
			data[v.Rid] = v
		}
	}
	return data, nil
}
