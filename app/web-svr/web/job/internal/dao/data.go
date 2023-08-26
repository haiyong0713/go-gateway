package dao

import (
	"context"
	"fmt"
	"net/url"

	"go-common/library/ecode"
	"go-gateway/app/web-svr/web/job/internal/model"

	"github.com/pkg/errors"
)

const _webTopURI = "/data/rank/reco-webtop.json"

func (d *dao) WebTop(ctx context.Context) ([]int64, error) {
	var res struct {
		Code int `json:"code"`
		List []*struct {
			Aid int64 `json:"aid"`
		} `json:"list"`
	}
	if err := d.httpR.Get(ctx, d.webTopURL, "", url.Values{}, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, ecode.Int(res.Code)
	}
	var aids []int64
	for _, v := range res.List {
		if v != nil && v.Aid > 0 {
			aids = append(aids, v.Aid)
		}
	}
	if len(aids) == 0 {
		return nil, errors.New("web top aids is nil")
	}
	return aids, nil
}

const _rankIndexURI = "/data/rank/reco-%d.json"

func (d *dao) RankIndex(ctx context.Context, day int64) ([]int64, error) {
	var res struct {
		Code int `json:"code"`
		Num  int `json:"num"`
		List []*struct {
			Aid int64 `json:"aid"`
		} `json:"list"`
	}
	if err := d.httpR.RESTfulGet(ctx, d.rankIndexURL, "", url.Values{}, &res, day); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, ecode.Int(res.Code)
	}
	var aids []int64
	for _, v := range res.List {
		if v != nil && v.Aid > 0 {
			aids = append(aids, v.Aid)
		}
	}
	if len(aids) == 0 {
		return nil, errors.New("rank index aids is nil")
	}
	return aids, nil
}

const _rankRcmdURI = "/data/rank/reco_region-%d.json"

func (d *dao) RankRecommend(ctx context.Context, rid int64) ([]int64, error) {
	var res struct {
		Code int `json:"code"`
		Num  int `json:"num"`
		List []*struct {
			Aid int64 `json:"aid"`
		} `json:"list"`
	}
	if err := d.httpR.RESTfulGet(ctx, d.rankRcmdURL, "", url.Values{}, &res, rid); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, ecode.Int(res.Code)
	}
	var aids []int64
	for _, v := range res.List {
		if v != nil && v.Aid > 0 {
			aids = append(aids, v.Aid)
		}
	}
	if len(aids) == 0 {
		return nil, errors.New("rank recommend aids is nil")
	}
	return aids, nil
}

const _lpRankRcmdURI = "/data/rank/reco_region-%s.json"

func (d *dao) LpRankRecommend(ctx context.Context, business string) ([]int64, error) {
	var res struct {
		Code int `json:"code"`
		Num  int `json:"num"`
		List []*struct {
			Aid int64 `json:"aid"`
		} `json:"list"`
	}
	if err := d.httpR.RESTfulGet(ctx, d.lpRankRcmdURL, "", url.Values{}, &res, business); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, ecode.Int(res.Code)
	}
	var aids []int64
	for _, v := range res.List {
		if v != nil && v.Aid > 0 {
			aids = append(aids, v.Aid)
		}
	}
	if len(aids) == 0 {
		return nil, errors.New("rank recommend aids is nil")
	}
	return aids, nil
}

const _rankRegionURI = "/data/rank/recent_region%s-%d-%d.json"

func (d *dao) RankRegion(ctx context.Context, rid, day, original int64) ([]*model.RankAid, error) {
	var data struct {
		Code int              `json:"code"`
		List []*model.RankAid `json:"list"`
	}
	if err := d.httpR.RESTfulGet(ctx, d.rankRegionURL, "", url.Values{}, &data, model.OriType[original], rid, day); err != nil {
		return nil, err
	}
	if data.Code != ecode.OK.Code() {
		return nil, ecode.Int(data.Code)
	}
	var res []*model.RankAid
	for _, v := range data.List {
		if v != nil && v.Aid > 0 {
			res = append(res, v)
		}
	}
	if len(res) == 0 {
		return nil, errors.New("rank region list is nil")
	}
	return res, nil
}

const _rankTagURI = "/tag/hot/web/%d/%d.json"

func (d *dao) RankTag(ctx context.Context, rid, tagID int64) ([]*model.RankAid, error) {
	var data struct {
		Code int              `json:"code"`
		List []*model.RankAid `json:"list"`
	}
	if err := d.httpR.RESTfulGet(ctx, d.rankTagURL, "", url.Values{}, &data, rid, tagID); err != nil {
		return nil, err
	}
	if data.Code != ecode.OK.Code() {
		return nil, ecode.Int(data.Code)
	}
	var res []*model.RankAid
	for _, v := range data.List {
		if v != nil && v.Aid > 0 {
			res = append(res, v)
		}
	}
	if len(res) == 0 {
		return nil, errors.New("rank tag list is nil")
	}
	return res, nil
}

const _rankListURI = "/data/rank/%s-web.json"

func (d *dao) RankList(ctx context.Context, typ model.RankListType, rid int64) (*model.RankList, error) {
	var res struct {
		Code int `json:"code"`
		*model.RankList
	}
	if err := d.httpR.RESTfulGet(ctx, d.rankListURL, "", url.Values{}, &res, d.rankListType(typ, rid)); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, ecode.Int(res.Code)
	}
	if res.RankList == nil || len(res.List) == 0 {
		return nil, errors.New("rank list aids is nil")
	}
	return res.RankList, nil
}

const _rankListOldURI = "/data/rank/all_region-%d-3.json"

func (d *dao) RankListOld(ctx context.Context, rid int64) (*model.RankList, error) {
	var res struct {
		Code int `json:"code"`
		*model.RankList
	}
	if err := d.httpR.RESTfulGet(ctx, d.rankListOldURL, "", url.Values{}, &res, rid); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, ecode.Int(res.Code)
	}
	if res.RankList == nil || len(res.List) == 0 {
		return nil, errors.New("rank list aids is nil")
	}
	return res.RankList, nil
}

func (d *dao) rankListType(typ model.RankListType, rid int64) string {
	if rid > 0 {
		return fmt.Sprintf("region-%d", rid)
	}
	switch typ {
	case model.RankListTypeAll:
		return "all"
	case model.RankListTypeOrigin:
		return "origin"
	case model.RankListTypeRookie:
		return "rookie"
	default:
		return ""
	}
}
