package dao

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	model "go-gateway/app/app-svr/app-dynamic/interface/model/draw"

	"go-common/library/ecode"
	"go-common/library/log"

	"github.com/pkg/errors"
)

const (
	_urlNearbyLocations = "/lbs_svr/v1/lbs_svr/nearby_poi_list"
	_urlSearchLocations = "/lbs_svr/v1/lbs_svr/search_poi_list"
)

func (d *Dao) GetNearbyLocationsTopK(ctx context.Context, k int, lat, lng float64) (locations []*model.LBSSearchItem, err error) {
	params := url.Values{}
	params.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	params.Set("lng", strconv.FormatFloat(lng, 'f', -1, 64))
	params.Set("page", "0")
	params.Set("page_size", strconv.Itoa(k))
	questUrl := d.conf.Hosts.VcCo + _urlNearbyLocations
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Pois []interface{} `json:"pois"`
		} `json:"data"`
	}
	if err = d.clientLongTimeout.Get(ctx, questUrl, "", params, &ret); err != nil {
		log.Error("%s query failed, params:(%s), error(%v)", _urlNearbyLocations, params.Encode(), err)
		return nil, err
	}
	if ret.Code != ecode.OK.Code() {
		log.Error("%s query failed, params:(%s), error code(%d)", _urlNearbyLocations, params.Encode(), ret.Code)
		err = errors.Wrap(ecode.Int(ret.Code), questUrl+"?"+params.Encode())
		return nil, err
	}
	if len(ret.Data.Pois) == 0 {
		log.Error("%s return empty, params:(%s)", _urlNearbyLocations, params.Encode())
		return
	}
	for _, loc := range ret.Data.Pois {
		bytes, err := json.Marshal(loc)
		if err != nil {
			log.Error("%s json decode failed, params:(%s), error(%v)", _urlNearbyLocations, params.Encode(), err)
			return nil, err
		}
		locations = append(locations, &model.LBSSearchItem{Pio: string(bytes)})
	}
	return
}

func (d *Dao) SearchLabs(ctx context.Context, word string, lat, lng float64, page, pageSize int) (locations []*model.LBSSearchItem, hasMore bool, err error) {
	params := url.Values{}
	params.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	params.Set("lng", strconv.FormatFloat(lng, 'f', -1, 64))
	params.Set("keyword", word)
	params.Set("page", strconv.Itoa(page))
	params.Set("page_size", strconv.Itoa(pageSize))
	questUrl := d.conf.Hosts.VcCo + _urlSearchLocations
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			HasMore bool          `json:"has_more"`
			Pois    []interface{} `json:"pois"`
		} `json:"data"`
	}
	if err = d.clientLongTimeout.Get(ctx, questUrl, "", params, &ret); err != nil {
		log.Error("%s query failed, params:(%s), error(%v)", _urlSearchLocations, params.Encode(), err)
		return nil, hasMore, err
	}
	if ret.Code != ecode.OK.Code() {
		log.Error("%s query failed, params:(%s), err code(%d)", _urlSearchLocations, params.Encode(), ret.Code)
		err = errors.Wrap(ecode.Int(ret.Code), questUrl+"?"+params.Encode())
		return nil, hasMore, err
	}
	if len(ret.Data.Pois) == 0 {
		log.Error("%s return empty, params:(%s)", _urlSearchLocations, params.Encode())
		return
	}
	for _, loc := range ret.Data.Pois {
		bytes, err := json.Marshal(loc)
		if err != nil {
			log.Error("%s json decode failed, params:(%s), error(%v)", _urlSearchLocations, params.Encode(), err)
			return nil, hasMore, err
		}
		locations = append(locations, &model.LBSSearchItem{Pio: string(bytes)})
	}
	hasMore = ret.Data.HasMore
	return
}
