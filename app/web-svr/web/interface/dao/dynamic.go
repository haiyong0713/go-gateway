package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	"go-gateway/app/web-svr/web/interface/model"

	cheesedyngrpc "git.bilibili.co/bapis/bapis-go/cheese/service/dynamic"
	cheeseseasongrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
	dynfeedgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"

	"github.com/pkg/errors"
)

const (
	_drawDetailsV2 = "/dynamic_svr/v0/dynamic_svr/get_dynamic_info"
	_dynCommonBiz  = "/common_biz/v0/common_biz/fetch_biz"
)

func (d *Dao) DrawInfos(c context.Context, mid int64, drawIds []int64) (map[int64]*model.DynamicCard, error) {
	params := url.Values{}
	params.Add("uid[]", strconv.FormatInt(mid, 10))
	for _, id := range drawIds {
		params.Add("dynamic_ids[]", strconv.FormatInt(id, 10))
	}
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		// Data []*dymdl.DrawDetailRes `json:"data"`
		Data struct {
			Cards []*model.DynamicCard `json:"cards"`
		} `json:"data"`
	}
	if err := d.httpR.Get(c, d.drawDetails, "", params, &ret); err != nil {
		log.Errorc(c, "DrawDetails http GET(%s) failed, params:(%s), error(%+v)", d.drawDetails, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "DrawDetails http GET(%s) failed, params:(%s), code: %v, msg: %v", d.drawDetails, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "DrawDetails url(%v) code(%v) msg(%v)", d.drawDetails, ret.Code, ret.Msg)
		return nil, err
	}
	res := make(map[int64]*model.DynamicCard)
	for _, card := range ret.Data.Cards {
		if card == nil {
			continue
		}
		if card.Desc.DynamicID == 0 {
			continue
		}
		res[card.Desc.DynamicID] = card
	}
	return res, nil
}

func (d *Dao) DynamicEntrance(c context.Context, mid, video, article, all int64) (*dynfeedgrpc.WebEntranceInfoRsp, error) {
	res, err := d.dynamicFeedGRPC.WebEntranceInfo(c, &dynfeedgrpc.WebEntranceInfoReq{Uid: mid, VideoOffset: video, ArticleOffset: article, AlltypeOffset: all, NeedHeadIcon: true})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) DynamicAttachPermissionCheck(c context.Context, mid int64) (*cheesedyngrpc.DynamicAttachPermissionCheckReply, error) {
	res, err := d.cheeseDynamicGRPC.DynamicAttachPermissionCheck(c, &cheesedyngrpc.DynamicAttachPermissionCheckReq{Mid: mid})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) DynamicAttachAdd(c context.Context, mid int64, url string) (*cheesedyngrpc.DynamicAttachAddReply, error) {
	res, err := d.cheeseDynamicGRPC.DynamicAttachAdd(c, &cheesedyngrpc.DynamicAttachAddReq{Mid: mid, Url: url})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) DynamicUserSeason(c context.Context, mid int64, pn, ps int32) (*cheeseseasongrpc.UserSeasonReply, error) {
	res, err := d.cheeseSeasonGRPC.UserSeason(c, &cheeseseasongrpc.UserSeasonReq{Mid: mid, NeedAll: 2, FinishSate: 2, Pn: pn, Ps: ps})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) DynamicSeasonProfile(c context.Context, ids []int32) (*cheeseseasongrpc.SeasonProfileReply, error) {
	res, err := d.cheeseSeasonGRPC.Profile(c, &cheeseseasongrpc.SeasonProfileReq{Ids: ids, Require: cheeseseasongrpc.SeasonRequire_REQUIRE_ALL})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) DynSimpleInfo(ctx context.Context, req *dynfeedgrpc.DynSimpleInfosReq) (*dynfeedgrpc.DynSimpleInfosRsp, error) {
	return d.dynamicFeedGRPC.DynSimpleInfos(ctx, req)
}

func (d *Dao) DynamicCommonInfos(ctx context.Context, ids []int64) (map[int64]*dynmdlV2.DynamicCommonCard, error) {
	params := new(struct {
		RIDs []int64 `json:"rid"`
	})
	params.RIDs = ids
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return nil, errors.Wrapf(err, "DynamicCommonInfos json.NewEncoder() params(%+v)", params)
	}
	req, err := http.NewRequest(http.MethodPost, d.dynCommonBiz, body)
	if err != nil {
		return nil, errors.Wrapf(err, "DynamicCommonInfos http.NewRequest() body(%s)", body)
	}
	req.Header.Set("Content-Type", "application/json")
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data *struct {
			Entry []*dynmdlV2.DynamicCommon `json:"entry"`
		} `json:"data"`
	}
	if err = d.httpR.Do(ctx, req, &ret); err != nil {
		return nil, errors.Wrapf(err, "DynamicCommonInfos http Post(%s) failed, req:(%+v)", d.dynCommonBiz, req)
	}
	if ret.Code != 0 {
		return nil, errors.Wrapf(ecode.Int(ret.Code), "DynamicCommonInfos url=%+v code=%+v msg=%+v", d.dynCommonBiz, ret.Code, ret.Msg)
	}
	if ret.Data == nil || len(ret.Data.Entry) == 0 {
		return nil, errors.New("DynamicCommonInfos get nothing")
	}
	var res = make(map[int64]*dynmdlV2.DynamicCommonCard)
	for _, entry := range ret.Data.Entry {
		if entry.RID == 0 || entry.Card == "" {
			log.Error("DynamicCommonInfos entry err=%+v", entry)
			continue
		}
		card := &dynmdlV2.DynamicCommonCard{}
		if err = json.Unmarshal([]byte(entry.Card), &card); err != nil {
			log.Error("DynamicCommonInfos json unmarshal entry err=%+v", err)
			continue
		}
		res[entry.RID] = card
	}
	return res, nil
}
