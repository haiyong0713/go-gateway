package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/dynamic/service/model"

	"github.com/pkg/errors"
)

// RegionArcs get new dynamic by bigData API.
func (d *Dao) RegionArcs(c context.Context, rid int32, remoteIP string) (aids []int64, total int, err error) {
	params := url.Values{}
	params.Set("rid", strconv.FormatInt(int64(rid), 10))
	var res struct {
		Code int            `json:"code"`
		Data model.AvidData `json:"data"`
	}
	if err = d.httpR.Get(c, d.regionURI, remoteIP, params, &res); err != nil {
		PromError("大数据分区接口", "d.httpR.Get(%s) error(%+v)", d.regionURI+"?"+params.Encode(), err)
		return
	}
	if res.Code != ecode.OK.Code() {
		PromError("大数据分区接口", "dynamic url(%s) res code(%d) or res.data(%v)", d.regionURI+"?"+params.Encode(), res.Code, res.Data)
		err = ecode.Int(res.Code)
		return
	}
	if len(res.Data.Avid) == 0 {
		log.Error("dynamic url(%s) res(%v)", d.regionURI+"?"+params.Encode(), res)
	}
	aids = res.Data.Avid
	total = res.Data.Count
	return
}

func (d *Dao) LpRegionArcs(c context.Context, rid int32, tagID []int64) ([]int64, int, error) {
	params := url.Values{}
	params.Set("rid", strconv.FormatInt(int64(rid), 10))
	params.Set("tag_id", xstr.JoinInts(tagID))
	var res struct {
		Code int            `json:"code"`
		Data model.AvidData `json:"data"`
	}
	if err := d.httpR.Get(c, d.lpRegionURI, "", params, &res); err != nil {
		PromError("大数据分区接口", "d.httpR.Get(%s) error(%+v)", d.lpRegionURI+"?"+params.Encode(), err)
		return nil, 0, err
	}
	if res.Code != ecode.OK.Code() {
		PromError("大数据分区接口", "dynamic url(%s) res code(%d) or res.data(%v)", d.lpRegionURI+"?"+params.Encode(), res.Code, res.Data)
		return nil, 0, errors.Wrap(ecode.Int(res.Code), d.lpRegionURI+"?"+params.Encode())
	}
	if len(res.Data.Avid) == 0 {
		log.Error("dynamic url(%s) res(%v)", d.lpRegionURI+"?"+params.Encode(), res)
	}
	return res.Data.Avid, res.Data.Count, nil
}

// RegionTagArcs get new dynamic by bigData API.
func (d *Dao) RegionTagArcs(c context.Context, rid int32, tagID int64, remoteIP string) (aids []int64, err error) {
	params := url.Values{}
	params.Set("rid", strconv.FormatInt(int64(rid), 10))
	params.Set("tag_id", strconv.FormatInt(tagID, 10))
	var res struct {
		Code int            `json:"code"`
		Data model.AvidData `json:"data"`
	}
	if err = d.httpR.Get(c, d.regionTagURI, remoteIP, params, &res); err != nil {
		PromError("大数据Tag接口", "d.httpR.Get(%s) error(%+v)", d.regionTagURI+"?"+params.Encode(), err)
		return
	}
	if res.Code != ecode.OK.Code() {
		PromError("大数据Tag接口", "dynamic url(%s) res code(%d) or res.data(%v)", d.regionTagURI+"?"+params.Encode(), res.Code, res.Data)
		err = ecode.Int(res.Code)
		return
	}
	if len(res.Data.Avid) == 0 {
		log.Error("dynamic url(%s) res(%v)", d.regionTagURI+"?"+params.Encode(), res)
	}
	aids = res.Data.Avid
	return
}
