package dao

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/model"
)

const (
	_rankURI          = "%s.json"
	_rankAllURI       = "all-%d"
	_rankAllRidURI    = "all_region-%d-%d"
	_rankAllRecURI    = "recent_all-%d"
	_rankAllRecRidURI = "recent_region-%d-%d"
	_rankOriAllURI    = "all_origin-%d"
	_rankOriAllRidURI = "all_region_origin-%d-%d"
	_rankOriRecURI    = "recent_origin-%d"
	_rankOriRecRidURI = "recent_region_origin-%d-%d"
	_rankAllNewURI    = "all_rookie-%d"
	_rankAllNewRidURI = "all_region_rookie-%d-%d"
	_rankRegionURI    = "recent_region%s-%d-%d.json"
	_rankRecURI       = "reco_region-%d.json"
	_rankTagURI       = "/tag/hot/web/%d/%d.json"
	_rankIndexURI     = "reco-%d.json"
	_customURI        = "game_custom_2.json"
	_hotLabelURI      = "/recommand"
	_webTopURI        = "reco-webtop.json"
	_promoteURI       = "old/promote-4.json"
)

// Ranking get ranking data from new api
func (d *Dao) Ranking(c context.Context, rid int16, rankType, day, arcType int) (res *model.RankNew, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	suffix := rankURI(rid, model.RankType[rankType], day, arcType)
	var rs struct {
		Code int                     `json:"code"`
		Note string                  `json:"note"`
		List []*model.RankNewArchive `json:"list"`
	}
	if err = d.httpBigData.RESTfulGet(c, d.rankURL, ip, params, &rs, suffix); err != nil {
		log.Error("d.httpBigData.RESTfulGet(%s) error(%v)", suffix, err)
		return
	}
	if rs.Code != ecode.OK.Code() {
		log.Error("d.httpBigData.RESTfulGet(%s) error code(%d)", suffix, rs.Code)
		err = ecode.Int(rs.Code)
		return
	}
	res = &model.RankNew{Note: rs.Note, List: rs.List}
	return
}

// HotLabel get hot label aids.
func (d *Dao) HotLabel(c context.Context) (aids []int64, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("cmd", "hot")
	params.Set("from", "10")
	timeout := time.Duration(d.c.Custom.RecommendTimeout) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("ignore_custom", "1")
	var res struct {
		Code int `json:"code"`
		Data []struct {
			Id int64 `json:"id"`
		} `json:"data"`
	}
	if err = d.httpBigData.Get(c, d.hotLabelURL, ip, params, &res); err != nil {
		log.Error("d.httpBigData.Get(%s) error(%v)", d.hotLabelURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("d.httpBigData.Get(%s) error(%v)", d.hotLabelURL, err)
		err = ecode.Int(res.Code)
		return
	}
	for _, v := range res.Data {
		if v.Id > 0 {
			aids = append(aids, v.Id)
		}
	}
	return
}

func rankURI(rid int16, rankType string, day, arcType int) string {
	if rankType == model.RankType[1] {
		if arcType == 1 {
			if rid > 0 {
				return fmt.Sprintf(_rankAllRecRidURI, rid, day)
			}
			return fmt.Sprintf(_rankAllRecURI, day)
		}
		if rid > 0 {
			return fmt.Sprintf(_rankAllRidURI, rid, day)
		}
		return fmt.Sprintf(_rankAllURI, day)
	} else if rankType == model.RankType[2] {
		if arcType == 1 {
			if rid > 0 {
				return fmt.Sprintf(_rankOriRecRidURI, rid, day)
			}
			return fmt.Sprintf(_rankOriRecURI, day)
		}
		if rid > 0 {
			return fmt.Sprintf(_rankOriAllRidURI, rid, day)
		}
		return fmt.Sprintf(_rankOriAllURI, day)
	}
	if rid > 0 {
		return fmt.Sprintf(_rankAllNewRidURI, rid, day)
	}
	return fmt.Sprintf(_rankAllNewURI, day)
}
