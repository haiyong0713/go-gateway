package fm

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/database/taishan"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"

	fmRec "git.bilibili.co/bapis/bapis-go/ott-recommend/automotive-channel"
	"github.com/pkg/errors"
)

const (
	_defaultHomePs  = 50
	_defaultPinPs   = 10
	_homePsWithPage = 20
	_notExist       = 0
	_oneHourSec     = 3600

	_build22 = 2_020_000

	_channelAIPrefix = "{channel_heat_score}:"
	_taishanBatchMax = 1000
)

type Dao struct {
	channelGrpc fmRec.AutomotiveChannelRecommenderClient
	db          *xsql.DB
	redisCli    *redis.Redis
	chanTsCli   taishan.TaishanProxyClient
	chanTsTable conf.TaishanTable
}

func New(c *conf.Config) *Dao {
	chanClient, err := fmRec.NewClient(c.FmRecommendGRPC)
	if err != nil {
		panic(fmt.Sprintf("ottRecommend automotiveChannel NewClient error (%+v)", err))
	}
	taishanClient, err := taishan.NewClient(c.Taishan.ChannelClient)
	if err != nil {
		panic(fmt.Sprintf("chanTaishan NewClient error (%+v)", err))
	}
	return &Dao{
		channelGrpc: chanClient,
		db:          xsql.NewMySQL(c.MySQL.Car),
		redisCli:    redis.NewRedis(c.Redis.Entrance),
		chanTsCli:   taishanClient,
		chanTsTable: *c.Taishan.ChannelTable,
	}
}

// AIRegionList ai智能视频tab顺序.
func (d *Dao) AIRegionList(ctx context.Context, mid int64, buvid string, dev model.DeviceInfo) ([]*fmRec.Region, error) {
	res, err := d.channelGrpc.GetRegionList(ctx, &fmRec.ReqRegionList{
		DeviceInfo: &fmRec.DeviceInfo{
			Buvid:     buvid,
			Build:     int32(dev.Build),
			ChannelId: dev.Channel,
			Model:     dev.Model,
		},
		Mid:   mid,
		TabId: fmRec.TabID_TAB_ID_VIDEO,
	})
	if err != nil {
		return nil, err
	}
	if res == nil || len(res.Regions) == 0 {
		return nil, ecode.NothingFound
	}
	return res.Regions, nil
}

// ChannelFeed 垂类稿件列表
func (d *Dao) ChannelFeed(ctx context.Context, mid, channelId, ps, pn int64, buvid string, dev model.DeviceInfo) ([]int64, error) {
	req := &fmRec.ReqChannelFeedRecommend{
		DeviceInfo: &fmRec.DeviceInfo{
			Buvid: buvid,
			Build: int32(dev.Build),
			Model: dev.Model,
		},
		Mid:       mid,
		ChannelId: channelId,
		PageInfo: &fmRec.PageInfo{
			PageSize: ps,
			PageNum:  pn,
		},
	}
	feed, err := d.channelGrpc.ChannelFeed(ctx, req)
	if err != nil {
		return nil, err
	}
	aids := make([]int64, 0)
	for _, v := range feed.Items {
		aids = append(aids, v.Id)
	}
	return aids, nil
}

// FmHome FM首页（v2.3以下使用）
func (d *Dao) FmHome(ctx context.Context, mid int64, buvid string, dev model.DeviceInfo, page *fm_v2.PageReq) (*fm_v2.RecResp, error) {
	if page == nil {
		page = &fm_v2.PageReq{PageSize: GetHomePs(dev), PageNext: &fm_v2.PageInfo{Pn: 0}}
	}
	if page.PageSize == 0 {
		page.PageSize = GetHomePs(dev)
	}
	if page.PageNext == nil {
		page.PageNext = &fm_v2.PageInfo{Pn: 0}
	}
	req := &fmRec.ReqFMHomeRecommend{
		DeviceInfo: &fmRec.DeviceInfo{
			Buvid: buvid,
			Build: int32(dev.Build),
		},
		Mid:      mid,
		PageSize: int64(page.PageSize),
		PageNum:  page.PageNext.Pn,
		Options:  []fmRec.FMHomeOption{fmRec.FMHomeOption(page.ManualRefresh)},
	}
	fmHome, err := d.channelGrpc.FMHome(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(fmHome.Items) == 0 {
		return nil, ecode.NothingFound
	}
	return &fm_v2.RecResp{
		Items: fmHome.Items,
		PageResp: fm_v2.PageResp{
			PageNext: &fm_v2.PageInfo{
				Pn: page.PageNext.Pn + 1,
				Ps: page.PageSize,
			},
			HasNext: fmHome.HasMore,
		},
	}, nil
}

// FmHomeV2 拉取FM首页的推荐（金刚位+精选推荐）（v2.3及以上使用）
func (d *Dao) FmHomeV2(ctx context.Context, param *fm_v2.ShowV2Param, pinPs int, recPage *fm_v2.PageReq) (pin *fmRec.Module, rec *fmRec.Module, err error) {
	if pinPs <= 0 {
		pinPs = _defaultPinPs
	}
	if recPage == nil {
		recPage = &fm_v2.PageReq{PageSize: GetHomePs(param.DeviceInfo), PageNext: &fm_v2.PageInfo{Pn: 0}}
	}
	if recPage.PageSize == 0 {
		recPage.PageSize = GetHomePs(param.DeviceInfo)
	}
	if recPage.PageNext == nil {
		recPage.PageNext = &fm_v2.PageInfo{Pn: 0}
	}

	options := make([]fmRec.FMHomeOption, 0)
	options = append(options, fmRec.FMHomeOption(param.ManualRefresh))
	if param.Mode == model.ClosePersonalAi {
		options = append(options, fmRec.FMHomeOption_FM_HOME_OPTION_PROHIBIT_PERSONALIZATION)
	}
	req := &fmRec.ReqFMHomeRecommendV2{
		DeviceInfo: &fmRec.DeviceInfo{
			Buvid:     param.Buvid,
			Build:     int32(param.Build),
			GuestId:   param.GuestId,
			ChannelId: param.Channel,
			Model:     param.Model,
		},
		Mid:     param.Mid,
		Options: options,
		Modules: []*fmRec.ModuleRequestInfo{
			{
				Type:     fmRec.ModuleType_MODULE_TYPE_BANNER_SMALL_CARDS,
				PageSize: int64(pinPs),
				PageNum:  0,
			},
			{
				Type:     fmRec.ModuleType_MODULE_TYPE_SELECTION,
				PageSize: int64(recPage.PageSize),
				PageNum:  recPage.PageNext.Pn,
			},
		},
	}
	fmHome, err := d.channelGrpc.FMHomeV2(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	for _, v := range fmHome.Modules {
		if v.Type == fmRec.ModuleType_MODULE_TYPE_BANNER_SMALL_CARDS {
			pin = v
		}
		if v.Type == fmRec.ModuleType_MODULE_TYPE_SELECTION {
			rec = v
		}
	}
	if rec == nil { // 金刚位可能不下发
		return nil, nil, errors.Wrap(ecode.NothingFound, "fmHome lack rec")
	}
	return pin, rec, nil
}

// FmPinPage 拉取金刚位"更多"页的推荐（v2.3及以上使用）
func (d *Dao) FmPinPage(ctx context.Context, param *fm_v2.PinPageParam, page *fm_v2.PageReq) (*fmRec.Module, error) {
	if page == nil {
		page = &fm_v2.PageReq{PageSize: GetHomePs(param.DeviceInfo), PageNext: &fm_v2.PageInfo{Pn: 0}}
	}
	if page.PageSize == 0 {
		page.PageSize = GetHomePs(param.DeviceInfo)
	}
	if page.PageNext == nil {
		page.PageNext = &fm_v2.PageInfo{Pn: 0}
	}
	req := &fmRec.ReqFMHomeRecommendV2{
		DeviceInfo: &fmRec.DeviceInfo{
			Buvid:     param.Buvid,
			Build:     int32(param.Build),
			GuestId:   param.GuestId,
			ChannelId: param.Channel,
			Model:     param.Model,
		},
		Mid: param.Mid,
		Modules: []*fmRec.ModuleRequestInfo{{
			Type:     fmRec.ModuleType_MODULE_TYPE_BANNER_SMALL_CARDS_VIEW_MORE,
			PageSize: int64(page.PageSize),
			PageNum:  page.PageNext.Pn,
		}},
	}
	resp, err := d.channelGrpc.FMHomeV2(ctx, req)
	//reqBytes, _ := json.Marshal(req)
	//respBytes, _ := json.Marshal(resp)
	//log.Warnc(ctx, "FmPinPage debug d.channelGrpc.FMHomeV2 req:%s, resp:%s, err:%+v", string(reqBytes), string(respBytes), err)
	if err != nil {
		return nil, err
	}
	if len(resp.Modules) == 0 || resp.Modules[0].Type != fmRec.ModuleType_MODULE_TYPE_BANNER_SMALL_CARDS_VIEW_MORE {
		return nil, ecode.NothingFound
	}
	return resp.Modules[0], nil
}

// FmChannelInfoAI 算法侧提供的频道信息
func (d *Dao) FmChannelInfoAI(ctx context.Context, chanIds []int64) (map[int64]*fm_v2.ChannelInfoAI, error) {
	var (
		records = make([]*taishan.Record, 0)
		res     = make(map[int64]*fm_v2.ChannelInfoAI)
	)
	if len(chanIds) == 0 {
		log.Warnc(ctx, "FmChannelInfoAI empty chanIds")
		return make(map[int64]*fm_v2.ChannelInfoAI), nil
	}
	if len(chanIds) > _taishanBatchMax {
		return nil, errors.Wrap(ecode.RequestErr, "taishan单次查询数量>1000")
	}
	for _, v := range chanIds {
		records = append(records, &taishan.Record{
			Key: []byte(_channelAIPrefix + strconv.FormatInt(v, 10)),
		})
	}
	req := &taishan.BatchGetReq{
		Table: d.chanTsTable.Table,
		Auth: &taishan.Auth{
			Token: d.chanTsTable.Token,
		},
		Records: records,
	}
	resp, err := d.chanTsCli.BatchGet(ctx, req)
	if err != nil {
		return nil, err
	}
	for _, r := range resp.Records {
		var (
			errTmp error
			chanId int64
			info   *fm_v2.ChannelInfoAI
		)
		if r.Status.ErrNo > 0 {
			errTmp = errors.Wrap(ecode.NothingFound, r.Status.Msg)
		} else {
			info = new(fm_v2.ChannelInfoAI)
			errTmp = json.Unmarshal(r.Columns[0].Value, &info)
		}
		if errTmp != nil {
			log.Errorc(ctx, "FmChannelInfoAI error, key is: %s, err is: %s", r.Key, errTmp)
			continue
		}
		chanId, errTmp = strconv.ParseInt(string(r.Key)[len(_channelAIPrefix):], 10, 64)
		if errTmp != nil {
			log.Errorc(ctx, "FmChannelInfoAI key, strconv.ParseInt err, key is: %s, err is: %s", r.Key, errTmp)
			continue
		}
		res[chanId] = info
	}
	return res, nil
}

func GetHomePs(dev model.DeviceInfo) int {
	if dev.Build >= _build22 {
		return _homePsWithPage
	}
	return _defaultHomePs
}
