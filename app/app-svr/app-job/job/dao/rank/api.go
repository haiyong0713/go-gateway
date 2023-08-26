package rank

import (
	"context"
	"errors"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/model/rank"
)

const (
	_rankBangumiAppURL = "/data/rank/all_region-33-app.json"
	_rankRegionAppURL  = "/data/rank/region-%d-app.json"
	_rankAllAppURL     = "/data/rank/all-app.json"
	_rankOriginAppURL  = "/data/rank/recent_origin-app.json"
)

// RankAppBangumi rank bangumi rid 33
func (d *Dao) RankAppBangumi(c context.Context) (list []*rank.List, err error) {
	var res struct {
		Code int          `json:"code"`
		List []*rank.List `json:"list"`
	}
	if err = d.client.Get(c, d.rankBangumiAppURL, "", nil, &res); err != nil {
		log.Error("recommend bangumi rank hots url(%s) error(%v)", d.rankBangumiAppURL, err)
		return
	}
	if res.Code != 0 {
		log.Error("recommend bangumi rank hots url(%s) error(%v)", d.rankBangumiAppURL, res.Code)
		err = fmt.Errorf("recommend bangumi rank hots api response code(%v)", res)
		return
	}
	if err = rankCheck(res.List); err != nil {
		return
	}
	list = res.List
	return
}

// RankAppRegion app reion rank
func (d *Dao) RankAppRegion(c context.Context, rid int) (list []*rank.List, err error) {
	var res struct {
		Code int          `json:"code"`
		List []*rank.List `json:"list"`
	}
	api := fmt.Sprintf(d.rankRegionAppURL, rid)
	if err = d.client.Get(c, api, "", nil, &res); err != nil {
		log.Error("recommend region rank hots url(%s) error(%v)", api, err)
		return
	}
	if res.Code != 0 {
		log.Error("recommend region rank hots url(%s) error(%v)", api, res.Code)
		err = fmt.Errorf("recommend region rank hots api response code(%v)", res)
		return
	}
	if err = rankCheck(res.List); err != nil {
		return
	}
	list = res.List
	return
}

// RankAppAll all rank
func (d *Dao) RankAppAll(c context.Context) (list []*rank.List, err error) {
	var res struct {
		Code int          `json:"code"`
		List []*rank.List `json:"list"`
	}
	if err = d.client.Get(c, d.rankAllAppURL, "", nil, &res); err != nil {
		log.Error("recommend All rank hots url(%s) error(%v)", d.rankAllAppURL, err)
		return
	}
	if res.Code != 0 {
		log.Error("recommend All rank hots url(%s) error(%v)", d.rankAllAppURL, res.Code)
		err = fmt.Errorf("recommend All rank hots api response code(%v)", res)
		return
	}
	if err = rankCheck(res.List); err != nil {
		return
	}
	list = res.List
	return
}

// RankAppOrigin origin rank
func (d *Dao) RankAppOrigin(c context.Context) (list []*rank.List, err error) {
	var res struct {
		Code int          `json:"code"`
		List []*rank.List `json:"list"`
	}
	if err = d.client.Get(c, d.rankOriginAppURL, "", nil, &res); err != nil {
		log.Error("recommend Origin rank hots url(%s) error(%v)", d.rankOriginAppURL, err)
		return
	}
	if res.Code != 0 {
		log.Error("recommend Origin rank hots url(%s) error(%v)", d.rankOriginAppURL, res.Code)
		err = fmt.Errorf("recommend Origin rank hots api response code(%v)", res)
		return
	}
	if err = rankCheck(res.List); err != nil {
		return
	}
	list = res.List
	return
}

func rankCheck(list []*rank.List) error {
	if len(list) == 0 {
		return errors.New("AllRank list struct is nil")
	}
	for _, item := range list {
		if item == nil {
			return errors.New("AllRank list item struct is nil")
		}
	}
	return nil
}
