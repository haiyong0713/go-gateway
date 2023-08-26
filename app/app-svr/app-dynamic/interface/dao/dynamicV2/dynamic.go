package dynamicV2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-common/library/sync/errgroup.v2"

	xecode "go-gateway/app/app-svr/app-dynamic/ecode"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyndrawrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/draw"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	"github.com/pkg/errors"
)

const (
	_dynUpdOffset      = "/dynamic_svr/v0/dynamic_svr/vd_upd_offset"
	_vdUpList          = "/dynamic_svr/v0/dynamic_svr/vd_uplist"
	_dynAdditionFollow = "/dynamic_mix/v1/dynamic_mix/attach_card_button"
	_topicSquare       = "/topic_svr/v1/topic_svr/hot_entry"
	_common            = "/common_biz/v0/common_biz/fetch_biz"
	_dynAllUpdOffset   = "/dynamic_svr/v0/dynamic_svr/upd_offset"
	_vote              = "/vote_svr/v1/vote_svr/do_vote_v2"
	_voteResult        = "/vote_svr/v1/vote_svr/vote_info"
)

func (d *Dao) DynVideoList(ctx context.Context, uid int64, updateBaseLine, assistBaseLine string, dynType []string, attention *dyncommongrpc.AttentionInfo, build, platform, mobiApp, buvid, devide, ip, from string) (*mdlv2.DynListRes, error) {
	req := &dyngrpc.VideoNewReq{
		Uid:            uid,
		UpdateBaseline: updateBaseLine,
		AssistBaseline: assistBaseLine,
		TypeList:       dynType,
		AttentionInfo:  attention,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:     build,
			Platform:  platform,
			MobiApp:   mobiApp,
			Buvid:     buvid,
			Device:    devide,
			Ip:        ip,
			From:      from,
			FromSpmid: "dt.video-dt.0.0.pv",
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
	}
	data, err := d.dynamicGRPC.VideoNew(ctx, req)
	if err != nil {
		log.Error("动态服务 视频页 DynVideoList error(%v)", err)
		return nil, errors.WithStack(err)
	}
	ret := &mdlv2.DynListRes{}
	ret.FromVideoNew(data, uid)
	return ret, nil
}

func (d *Dao) DynVideoHistory(ctx context.Context, uid int64, offset string, page int64, dynType []string, attention *dyncommongrpc.AttentionInfo, build, platform, mobiApp, buvid, devide, ip, from string) (*mdlv2.DynListRes, error) {
	req := &dyngrpc.VideoHistoryReq{
		Uid:           uid,
		Offset:        offset,
		Page:          page,
		TypeList:      dynType,
		AttentionInfo: attention,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:    build,
			Platform: platform,
			MobiApp:  mobiApp,
			Buvid:    buvid,
			Device:   devide,
			Ip:       ip,
			From:     from,
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
	}
	data, err := d.dynamicGRPC.VideoHistory(ctx, req)
	if err != nil {
		log.Error("动态服务 视频页 DynVideoHistory error(%v)", err)
		return nil, errors.WithStack(err)
	}
	ret := &mdlv2.DynListRes{}
	ret.FromVideoHistory(data, uid)
	return ret, nil
}

func (d *Dao) DynBriefs(ctx context.Context, dynIDs []int64, build, platform, mobiApp, buvid, devide, ip, from, fromSpmid string, needLikeUsers, needBottom bool, uid int64) (*mdlv2.DynListRes, error) {
	req := &dyngrpc.DynBriefsReq{
		Uid:    uid,
		DynIds: dynIDs,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:     build,
			Platform:  platform,
			MobiApp:   mobiApp,
			Buvid:     buvid,
			Device:    devide,
			Ip:        ip,
			From:      from,
			FromSpmid: fromSpmid,
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          needLikeUsers,
			NeedLimitFoldStatement: true,
			NeedBottom:             needBottom,
			NeedTopicInfo:          true,
			NeedLikeIcon:           false,
			NeedRepostNum:          true,
		},
	}
	data, err := d.dynamicGRPC.DynBriefs(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ret := &mdlv2.DynListRes{}
	ret.FromDynBriefs(data, uid)
	return ret, nil
}

func (d *Dao) DynSimpleInfos(ctx context.Context, dynIDs []int64) (map[int64]*dyngrpc.DynSimpleInfo, error) {
	var max50 = 50
	g := errgroup.WithContext(ctx)
	mu := sync.Mutex{}
	res := make(map[int64]*dyngrpc.DynSimpleInfo, len(dynIDs))
	for i := 0; i < len(dynIDs); i += max50 {
		var tmpDynIDs []int64
		if i+max50 > len(dynIDs) {
			tmpDynIDs = dynIDs[i:]
		} else {
			tmpDynIDs = dynIDs[i : i+max50]
		}
		g.Go(func(c context.Context) (err error) {
			req := &dyngrpc.DynSimpleInfosReq{
				DynIds: tmpDynIDs,
			}
			data, err := d.dynamicGRPC.DynSimpleInfos(c, req)
			if err != nil {
				return err
			}
			mu.Lock()
			for _, v := range data.DynSimpleInfos {
				res[v.DynId] = v
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) DynVideoPersonal(ctx context.Context, hostUid, uid int64, IsPreload bool, offset, build, platform, mobiApp, buvid, devide, ip, from, footprint string, attention *dyncommongrpc.AttentionInfo, dynType []string) (*mdlv2.VideoPersonal, error) {
	req := &dyngrpc.VideoPersonalReq{
		HostUid:        hostUid,
		IsPreload:      IsPreload,
		Offset:         offset,
		Uid:            uid,
		AttentionUsers: attention,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:     build,
			Platform:  platform,
			MobiApp:   mobiApp,
			Buvid:     buvid,
			Device:    devide,
			Ip:        ip,
			From:      from,
			FromSpmid: "dt.dt-video-quick-cosume.0.0.pv",
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
		Footprint: footprint,
		TypeList:  dynType,
	}
	data, err := d.dynamicGRPC.VideoPersonal(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ret := &mdlv2.VideoPersonal{}
	ret.FromVideoPersonal(data, uid)
	return ret, nil
}

func (d *Dao) VdUpList(ctx context.Context, teenager int, uid int64, buvid string) (*mdlv2.VdUpListRsp, error) {
	params := url.Values{}
	params.Set("teenagers_mode", strconv.Itoa(teenager))
	params.Set("uid", strconv.FormatInt(uid, 10))
	params.Set("buvid", buvid)
	upList := d.c.Hosts.VcCo + _vdUpList
	var ret struct {
		Code int                `json:"code"`
		Msg  string             `json:"msg"`
		Data *mdlv2.VdUpListRsp `json:"data"`
	}
	if err := d.client.Get(ctx, upList, "", params, &ret); err != nil {
		log.Error("动态服务 视频页 最近访问up主头像列表 vd_uplist error(%v)", err)
		return nil, errors.WithStack(err)
	}
	if ret.Code != 0 {
		log.Errorc(ctx, "VdUpList failed to HTTP GET: %v. params: %v.  code: %v. msg: %v", upList, params.Encode(), ret.Code, ret.Msg)
		return nil, errors.Wrapf(ecode.Int(ret.Code), ret.Msg)
	}
	return ret.Data, nil
}

func (d *Dao) DynVideoUpdateOffset(c context.Context, uid, hostUid int64, offset, footprint string) error {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(uid, 10))
	params.Set("host_uid", strconv.FormatInt(hostUid, 10))
	params.Set("read_offset", offset)
	params.Set("footprint", footprint)
	updOffsetURL := d.c.Hosts.VcCo + _dynUpdOffset
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := d.client.Get(c, updOffsetURL, "", params, &ret); err != nil {
		return errors.WithStack(err)
	}
	if ret.Code != 0 {
		return errors.Wrapf(ecode.Int(ret.Code), "DynUpdOffset url(%v) code(%v) msg(%v)", updOffsetURL, ret.Code, ret.Msg)
	}
	return nil
}

func (d *Dao) AdditionFollow(c context.Context, aType string, dynamicID int64, state string) error {
	params := url.Values{}
	params.Set("attach_card_type", aType)
	params.Set("dynamic_id", strconv.FormatInt(dynamicID, 10))
	params.Set("cur_btn_status", state)
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := d.client.Get(c, d.c.Hosts.VcCo+_dynAdditionFollow, "", params, &ret); err != nil {
		return errors.WithStack(err)
	}
	if ret.Code != 0 {
		return errors.Wrapf(ecode.Int(ret.Code), "AdditionFollow url(%v) code(%v) msg(%v)", d.c.Hosts.VcCo+_dynAdditionFollow, ret.Code, ret.Msg)
	}
	return nil
}

func (d *Dao) DynMixNew(c context.Context, general *mdlv2.GeneralParam, param *api.DynAllReq, dynType []string, attention *dyncommongrpc.AttentionInfo) (*mdlv2.DynListRes, error) {
	var (
		adExtra   string
		dislikeTs int64
	)
	if param.GetAdParam() != nil {
		adExtra = param.GetAdParam().GetAdExtra()
	}
	if param.GetRcmdUpsParam() != nil {
		dislikeTs = param.GetRcmdUpsParam().GetDislikeTs()
	}
	req := &dyngrpc.GeneralNewReq{
		Uid:            general.Mid,
		UpdateBaseline: param.UpdateBaseline,
		TypeList:       dynType,
		AttentionInfo:  attention,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			ColdStart:    param.ColdStart,
			From:         param.From,
			FromSpmid:    "dt.dt.0.0.pv",
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
		AdParam:      &dyngrpc.AdParam{AdExtra: adExtra},
		RcmdUpsParam: &dyngrpc.RcmdUPsParam{DislikeTs: dislikeTs},
		TabRecall: &dyngrpc.TabRecallUp{
			TabRecallUid:  param.TabRecallUid,
			TabRecallType: dyngrpc.StyleType(param.TabRecallType),
		},
	}
	// 6.23.5才发新用户推荐
	if feature.GetBuildLimit(c, d.c.Feature.FeatureBuildLimit.DynUpRcmdOldUi, &feature.OriginResutl{
		BuildLimit: (general.IsIPhonePick() && general.GetBuild() < d.c.BuildLimit.DynUpRcmdOldUiIOS) || (general.IsPad() && general.GetBuild() < d.c.BuildLimit.DynUpRcmdOldUiIOSPad) ||
			(general.IsAndroidPick() && general.GetBuild() < d.c.BuildLimit.DynUpRcmdOldUiAndroid) || (general.IsPadHD() && general.GetBuild() < d.c.BuildLimit.DynUpRcmdOldUiIOSHD) || general.IsAndroidHD()}) {
		req.VersionCtrl.UpRcmdOldUi = true
	}
	data, err := d.dynamicGRPC.GeneralNew(c, req)
	if err != nil {
		log.Error("动态服务 综合页 GeneralNew error(%v)", err)
		return nil, err
	}
	ret := &mdlv2.DynListRes{}
	ret.FromMixNew(data, general.Mid)
	return ret, nil
}

func (d *Dao) DynMixHistory(c context.Context, general *mdlv2.GeneralParam, param *api.DynAllReq, dynType []string, attention *dyncommongrpc.AttentionInfo) (*mdlv2.DynListRes, error) {
	var adExtra string
	if param.GetAdParam() != nil {
		adExtra = param.GetAdParam().GetAdExtra()
	}
	req := &dyngrpc.GeneralHistoryReq{
		Uid:           general.Mid,
		HistoryOffset: param.Offset,
		Page:          int64(param.Page),
		TypeList:      dynType,
		AttentionInfo: attention,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			From:         param.From,
			FromSpmid:    "dt.dt.0.0.pv",
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
		AdParam: &dyngrpc.AdParam{AdExtra: adExtra},
	}
	data, err := d.dynamicGRPC.GeneralHistory(c, req)
	if err != nil {
		log.Error("动态服务 综合页 GeneralHistory error(%v)", err)
		return nil, err
	}
	ret := &mdlv2.DynListRes{}
	ret.FromMixHistory(data, general.Mid)
	return ret, nil
}

func (d *Dao) MixUpList(c context.Context, general *mdlv2.GeneralParam, param *api.DynAllReq) (*dyngrpc.MixUpListRsp, error) {
	res, err := d.dynamicGRPC.MixUpList(c, &dyngrpc.MixUpListReq{
		Uid:              general.Mid,
		LastReqTimestamp: int64(general.LocalTime),
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           metadata.String(c, metadata.RemoteIP),
			From:         param.From,
			TeenagerMode: int32(general.GetTeenagerInt()),
			//ColdStart:    0,
			Version: general.GetVersion(),
			//Network:     general.GetNetWork(),
			//Scene:       0,
			//FromSpmid:   "",
			//UpRcmdOldUi: false,
			CloseRcmd: int32(general.GetDisableRcmdInt()),
		},
		From:          dyngrpc.MixUpListFrom_MIX_UPLIST_FROM_GW,
		TabRecallUid:  param.TabRecallUid,
		TabRecallType: dyngrpc.StyleType(param.TabRecallType),
	})
	if err != nil {
		log.Error("动态服务 综合页 最近访问up主头像列表 MixUpList error(%v)", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) VideoUpList(c context.Context, general *mdlv2.GeneralParam, param *api.DynVideoReq) (*dyngrpc.VideoUpListRsp, error) {
	res, err := d.dynamicGRPC.VideoUpList(c, &dyngrpc.VideoUpListReq{
		Uid: general.Mid,
		Meta: &dyncommongrpc.MetaDataCtrl{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.GetRemoteIP(),
			From:         param.GetFrom(),
			TeenagerMode: int32(general.GetTeenagerInt()),
			//ColdStart:  0,
			Version: general.GetVersion(),
			//Network:    general.GetNetWork(),
			//FromSpmid:  "",
		},
		From: dyngrpc.MixUpListFrom_MIX_UPLIST_FROM_GW,
	})
	if err != nil {
		log.Errorc(c, "动态服务 视频页 最近访问up主头像列表 VideoUpList error(%v)", err)
		return nil, err
	}
	return res, nil
}

// 老话题广场
func (d *Dao) MixTopisSquareOld(c context.Context, general *mdlv2.GeneralParam) (*mdlv2.OldTopicSquareImpl, error) {
	params := url.Values{}
	params.Set("platform", general.GetPlatform())
	params.Set("build", general.GetBuildStr())
	params.Set("mobi_app", general.GetMobiApp())
	params.Set("device", general.GetDevice())
	params.Set("uid", strconv.FormatInt(general.Mid, 10))
	topicSquare := d.c.Hosts.VcCo + _topicSquare
	var ret struct {
		Code int                       `json:"code"`
		Msg  string                    `json:"msg"`
		Data *mdlv2.OldTopicSquareImpl `json:"data"`
	}
	if err := d.client.Get(c, topicSquare, "", params, &ret); err != nil {
		log.Error("动态服务 综合页 话题广场 hot_entry error(%v)", err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Error("MixTopisSquareOld failed to HTTP GET: %v. params: %v.  code: %v. msg: %v", topicSquare, params.Encode(), ret.Code, ret.Msg)
		return nil, errors.Wrapf(ecode.Int(ret.Code), ret.Msg)
	}
	return ret.Data, nil
}

func (d *Dao) DrawDetails(c context.Context, general *mdlv2.GeneralParam, drawIds []int64) (map[int64]*mdlv2.DrawDetailRes, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*mdlv2.DrawDetailRes)
	for i := 0; i < len(drawIds); i += max50 {
		var tmpids []int64
		if i+max50 > len(drawIds) {
			tmpids = drawIds[i:]
		} else {
			tmpids = drawIds[i : i+max50]
		}
		g.Go(func(ctx context.Context) error {
			reply, err := d.drawDetailsGRPC(ctx, general, tmpids)
			if err != nil {
				log.Error("drawDetailsGRPC failed: %+v", err)
				return err
			}
			mu.Lock()
			for k, v := range reply {
				res[k] = v
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) drawDetailsGRPC(c context.Context, general *mdlv2.GeneralParam, drawIds []int64) (map[int64]*mdlv2.DrawDetailRes, error) {
	drawDetailReq := &dyndrawrpc.DrawDetailReq{
		Uid:      general.Mid,
		Ids:      drawIds,
		MetaData: general.ToDynCmnMetaData(),
	}

	resp, err := d.dynDrawGRPC.Detail(c, drawDetailReq)
	if err != nil {
		return nil, err
	}

	ret := make(map[int64]*mdlv2.DrawDetailRes)
	for _, item := range resp.GetDocItems() {
		itemSts := new(mdlv2.DrawDetailRes)
		err = json.Unmarshal([]byte(item.GetItem()), itemSts)
		if err != nil {
			log.Error("drawDetailsGRPC unmarshal resp item error: %+v", err)
			continue
		}
		ret[item.GetDocId()] = itemSts
	}
	return ret, nil
}

func (d *Dao) ListWordText(c context.Context, mid int64, wordIDs []int64) (map[int64]string, error) {
	res, err := d.dynamicGRPC.ListWordText(c, &dyngrpc.WordTextReq{Uid: mid, Rids: wordIDs})
	if err != nil {
		log.Error("ListWordText mid%v dynIDs %v", mid, wordIDs)
		return nil, err
	}
	return res.GetContent(), nil
}

func (d *Dao) CommonInfos(c context.Context, ids []int64) (map[int64]*mdlv2.DynamicCommonCard, error) {
	type params struct {
		RIDs []int64 `json:"rid"`
	}
	p := &params{
		RIDs: ids,
	}
	bs, _ := json.Marshal(p)
	common := d.c.Hosts.VcCo + _common
	req, _ := http.NewRequest("POST", common, strings.NewReader(string(bs)))
	req.Header.Set("Content-Type", "application/json")
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data *struct {
			Entry []*mdlv2.DynamicCommon `json:"entry"`
		} `json:"data"`
	}
	if err := d.client.Do(c, req, &ret); err != nil {
		xmetric.DyanmicItemAPI.Inc(common, "request_error")
		log.Error("CommonInfos %v", err)
		return nil, err
	}
	if ret.Code != 0 {
		xmetric.DyanmicItemAPI.Inc(common, "reply_code_error")
		err := errors.Wrap(ecode.Int(ret.Code), common)
		log.Error("CommonInfos err %v", err)
		return nil, err
	}
	if ret.Data == nil || len(ret.Data.Entry) == 0 {
		xmetric.DyanmicItemAPI.Inc(common, "reply_data_error")
		err := errors.New("CommonInfos get nothing")
		log.Error("CommonInfos err %v", err)
		return nil, err
	}
	var res = make(map[int64]*mdlv2.DynamicCommonCard)
	for _, engry := range ret.Data.Entry {
		if engry == nil || engry.RID == 0 || engry.Card == "" {
			log.Error("CommonInfos entry err %v", engry)
			continue
		}
		card := &mdlv2.DynamicCommonCard{}
		if err := json.Unmarshal([]byte(engry.Card), &card); err != nil {
			log.Error("CommonInfos json unmarshal entry err %v", err)
			continue
		}
		res[engry.RID] = card
	}
	return res, nil
}

func (d *Dao) DyncApplet(c context.Context, mid int64, appletIDs []int64) (map[int64]*dyncommongrpc.ProgramItem, error) {
	resTmp, err := d.dynamicGRPC.ListWidget(c, &dyngrpc.WidgetReq{Uid: mid, Rids: appletIDs})
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return resTmp.GetItems(), nil
}

func (d *Dao) DynRcmdUpExchange(c context.Context, general *mdlv2.GeneralParam, req *api.DynRcmdUpExchangeReq) (*mdlv2.RcmdUPCard, error) {
	resTmp, err := d.dynamicGRPC.ListUpRecommend(c, &dyngrpc.UpRecommendReq{
		Uid:       general.Mid,
		DislikeTs: req.DislikeTs,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetPlatform(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			From:         req.GetFrom(),
			TeenagerMode: int32(general.GetTeenagerInt()),
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
		},
	})
	if err != nil {
		log.Error("DynRcmdUpExchangeReq err %v", err)
		return nil, err
	}
	if resTmp.GetRcmdUps() == nil {
		return nil, errors.New("ListUpRecommend() return nil")
	}
	res := new(mdlv2.RcmdUPCard)
	res.FromRcmdUPCard(resTmp.GetRcmdUps())
	log.Warn("DynRcmdUpExchange mid(%d) list(%v)", general.Mid, res)
	return res, nil
}

func (d *Dao) Votes(c context.Context, mid int64, voteIDs []int64) (map[int64]*dyncommongrpc.VoteInfo, error) {
	resTmp, err := d.dynVoteClient.ListFeedVotes(c, &dynvotegrpc.ListFeedVotesReq{Uid: mid, VoteIds: voteIDs})
	if err != nil {
		log.Error("Votes %v, err %v", voteIDs, err)
		return nil, err
	}
	return resTmp.GetVoteInfos(), nil
}

func (d *Dao) DynAllPersonal(ctx context.Context, hostUid, uid int64, IsPreload bool, offset, build, platform, mobiApp, buvid, devide, ip, from string, attention *dyncommongrpc.AttentionInfo, footprint string, dynType []string) (*mdlv2.AllPersonal, error) {
	req := &dyngrpc.VideoPersonalReq{
		HostUid:        hostUid,
		IsPreload:      IsPreload,
		Offset:         offset,
		Uid:            uid,
		AttentionUsers: attention,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:     build,
			Platform:  platform,
			MobiApp:   mobiApp,
			Buvid:     buvid,
			Device:    devide,
			Ip:        ip,
			From:      from,
			FromSpmid: "dt.dt-video-quick-cosume.0.0.pv",
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
		Footprint: footprint,
		TypeList:  dynType,
	}
	data, err := d.dynamicGRPC.GeneralPersonal(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ret := &mdlv2.AllPersonal{}
	ret.FromAllPersonal(data)
	return ret, nil
}

func (d *Dao) DynAllUpdOffset(c context.Context, gen *mdlv2.GeneralParam, hostUid int64, offset, footprint string) error {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(gen.Mid, 10))
	params.Set("host_uid", strconv.FormatInt(hostUid, 10))
	params.Set("read_offset", offset)
	params.Set("footprint", footprint)
	params.Set("buvid", gen.GetBuvid())
	params.Set("platform", gen.GetPlatform())
	params.Set("user_ip", gen.IP)
	params.Set("version", gen.GetVersion())
	updOffsetURL := d.c.Hosts.VcCo + _dynAllUpdOffset
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := d.client.Get(c, updOffsetURL, "", params, &ret); err != nil {
		return errors.WithStack(err)
	}
	if ret.Code != 0 {
		return errors.Wrapf(ecode.Int(ret.Code), "DynUpdOffset url(%v) code(%v) msg(%v)", updOffsetURL, ret.Code, ret.Msg)
	}
	return nil
}

func (d *Dao) Vote(c context.Context, general *mdlv2.GeneralParam, voteReq *api.DynVoteReq) (error, string) {
	type params struct {
		VoteID    int64   `json:"vote_id"`
		Votes     []int64 `json:"votes"`
		Status    int     `json:"status"`
		UID       int64   `json:"voter_uid"`
		DynamicID int64   `json:"dynamic_id"`
		OpBit     int     `json:"op_bit"`
	}
	dynid, _ := strconv.ParseInt(voteReq.DynamicId, 10, 64)
	p := &params{
		VoteID:    voteReq.VoteId,
		Votes:     voteReq.Votes,
		Status:    int(voteReq.Status),
		UID:       general.Mid,
		DynamicID: dynid,
	}
	if voteReq.Share {
		p.OpBit = 1
	}
	bs, err := json.Marshal(p)
	if err != nil {
		log.Error("Vote err %v", err)
		return nil, d.c.Resource.Text.DynVoteFaild
	}
	vote := d.c.Hosts.VcCo + _vote
	req, err := http.NewRequest("POST", vote, strings.NewReader(string(bs)))
	if err != nil {
		log.Error("Vote err %v", err)
		return nil, d.c.Resource.Text.DynVoteFaild
	}
	req.Header.Set("Content-Type", "application/json")
	var ret struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		Message string `json:"message"`
	}
	if err := d.client.Do(c, req, &ret); err != nil {
		log.Error("Vote %v", err)
		return nil, d.c.Resource.Text.DynVoteFaild
	}
	if ret.Code != 0 {
		log.Error("Vote err %v", errors.Wrap(ecode.Int(ret.Code), vote))
	}
	return nil, ret.Message
}

func (d *Dao) VoteResult(c context.Context, general *mdlv2.GeneralParam, voteID int64) (*dyncommongrpc.VoteInfo, error) {
	params := url.Values{}
	params.Set("vote_id", strconv.FormatInt(voteID, 10))
	params.Set("uid", strconv.FormatInt(general.Mid, 10))
	voteResult := d.c.Hosts.VcCo + _voteResult
	var res struct {
		Code int               `json:"code"`
		Msg  string            `json:"msg"`
		Data *mdlv2.VoteResule `json:"data"`
	}
	if err := d.client.Get(c, voteResult, "", params, &res); err != nil {
		return nil, errors.WithStack(err)
	}
	if res.Code != 0 {
		return nil, errors.Wrapf(ecode.Int(res.Code), "VoteResult url(%v) code(%v) msg(%v)", voteResult, res.Code, res.Msg)
	}
	if res.Data == nil || res.Data.Info == nil {
		err := errors.New("vote result get nil info")
		return nil, err
	}
	vi := &dyncommongrpc.VoteInfo{
		VoteId:     res.Data.Info.VoteID,
		Title:      res.Data.Info.Title,
		Desc:       res.Data.Info.Desc,
		JoinNum:    res.Data.Info.Cnt,
		Type:       res.Data.Info.Type,
		ChoiceCnt:  res.Data.Info.ChoiceCnt,
		EndTime:    res.Data.Info.Endtime,
		Status:     res.Data.Info.Status,
		BizType:    res.Data.Info.BizType,
		ImgUrl:     res.Data.Info.ImgURL,
		MyVotes:    res.Data.MyVotes,
		OptionsCnt: res.Data.Info.OptionsCnt,
	}
	for _, option := range res.Data.Info.Options {
		if option == nil {
			continue
		}
		op := &dyncommongrpc.VoteOptionInfo{
			OptIdx:  option.Idx,
			OptDesc: option.Desc,
			ImgUrl:  option.ImgURL,
			Cnt:     option.Cnt,
			BtnStr:  option.BtnStr,
			Title:   option.Title,
		}
		vi.Options = append(vi.Options, op)
	}
	return vi, nil
}

// UpListViewMore get UpListViewMore.
func (d *Dao) UpListViewMore(c context.Context, general *mdlv2.GeneralParam, sortType int32) (*dyngrpc.UpListViewMoreRsp, error) {
	UpListViewMoreReply, err := d.dynamicGRPC.UpListViewMore(c, &dyngrpc.UpListViewMoreReq{
		Uid:      general.Mid,
		SortType: sortType,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
		},
	})
	if err != nil || UpListViewMoreReply == nil {
		log.Errorc(c, "Dao.UpListViewMore(mid: %+v) failed. error(%+v)", general.Mid, err)
		return nil, err
	}
	return UpListViewMoreReply, nil
}

func (d *Dao) DynDetail(c context.Context, general *mdlv2.GeneralParam, param *api.DynDetailReq) (*mdlv2.DynDetailRes, error) {
	var adExtra string
	if param.GetAdParam() != nil {
		adExtra = param.GetAdParam().GetAdExtra()
	}
	dynID, _ := strconv.ParseInt(param.DynamicId, 10, 64)
	req := &dyngrpc.DynDetailReq{
		Uid:     general.Mid,
		HostUid: param.Uid,
		DynId:   dynID,
		DynRevsId: &dyncommongrpc.DynRevsId{
			DynType: param.DynType,
			Rid:     param.Rid,
		},
		AdParam: &dyngrpc.AdParam{AdExtra: adExtra},
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			From:         param.From,
			FromSpmid:    "dt.dt-detail.0.0.pv",
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
	}
	reply, err := d.dynamicGRPC.DynDetail(c, req)
	if ecode.Cause(err).Code() > 0 {
		// 业务code
		return nil, xecode.DynViewNotFound
	}
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	res := &mdlv2.DynDetailRes{}
	dynTmp := &mdlv2.Dynamic{}
	dynTmp.FromDynamic(reply.Dyn)
	res.Dynamic = dynTmp
	res.Recommend = reply.Recommend
	return res, nil
}

func (d *Dao) DynLight(c context.Context, general *mdlv2.GeneralParam, param *api.DynLightReq, dynType []string, attention *dyncommongrpc.AttentionInfo) (*mdlv2.DynListRes, error) {
	req := &dyngrpc.GeneralHistoryReq{
		Uid:           general.Mid,
		HistoryOffset: param.HistoryOffset,
		Page:          int64(param.Page),
		TypeList:      dynType,
		AttentionInfo: attention,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			From:         param.From,
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          false,
			NeedLimitFoldStatement: false,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
		AdParam: &dyngrpc.AdParam{},
	}
	data, err := d.dynamicGRPC.DynLight(c, req)
	if err != nil {
		return nil, err
	}
	ret := &mdlv2.DynListRes{}
	ret.FromDynLight(data, general.Mid)
	return ret, nil
}

func (d *Dao) DynUnLoginLight(c context.Context, general *mdlv2.GeneralParam, param *api.DynLightReq, dynType []string) (*mdlv2.DynListRes, error) {
	req := &dyngrpc.UnloginLightReq{
		FakeUid:       param.FakeUid,
		HistoryOffset: param.HistoryOffset,
		TypeList:      dynType,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			From:         param.From,
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          false,
			NeedLimitFoldStatement: false,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
	}
	data, err := d.dynamicGRPC.UnloginLight(c, req)
	if err != nil {
		return nil, err
	}
	ret := &mdlv2.DynListRes{}
	ret.FromDynUnLoginLight(data, req.FakeUid)
	return ret, nil
}

func (d *Dao) LikeList(c context.Context, general *mdlv2.GeneralParam, param *api.LikeListReq, attention *dyncommongrpc.AttentionInfo) (*dyngrpc.LikeListRsp, error) {
	var (
		_max = int64(20)
	)
	dynID, _ := strconv.ParseInt(param.DynamicId, 10, 64)
	req := &dyngrpc.LikeListReq{
		Uid:   general.Mid,
		DynId: dynID,
		DynRevsId: &dyncommongrpc.DynRevsId{
			DynType: param.DynType,
			Rid:     param.Rid,
		},
		UidOffset:     param.UidOffset,
		PageNumber:    int64(param.Page),
		PageSize:      _max,
		AttentionInfo: attention,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
		},
	}
	reply, err := d.dynamicGRPC.LikeList(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) RepostList(c context.Context, general *mdlv2.GeneralParam, param *api.RepostListReq, ps int) (*mdlv2.RepostListRes, error) {
	dynID, _ := strconv.ParseInt(param.DynamicId, 10, 64)
	req := &dyngrpc.RepostListReq{
		Uid:   general.Mid,
		DynId: dynID,
		DynRevsId: &dyncommongrpc.DynRevsId{
			DynType: param.DynType,
			Rid:     param.Rid,
		},
		Offset:   param.Offset,
		PageSize: int64(ps),
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			From:         param.From,
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
	}
	reply, err := d.dynamicGRPC.RepostList(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	res := &mdlv2.RepostListRes{}
	res.FromRepostList(reply)
	return res, nil
}

func (d *Dao) HotRepostList(c context.Context, general *mdlv2.GeneralParam, param *api.RepostListReq, ps int) (*mdlv2.RepostListRes, error) {
	dynID, _ := strconv.ParseInt(param.DynamicId, 10, 64)
	req := &dyngrpc.RepostListReq{
		Uid:   general.Mid,
		DynId: dynID,
		DynRevsId: &dyncommongrpc.DynRevsId{
			DynType: param.DynType,
			Rid:     param.Rid,
		},
		Offset:   param.Offset,
		PageSize: int64(ps),
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			From:         param.From,
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
	}
	reply, err := d.dynamicGRPC.HotRepostList(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	res := &mdlv2.RepostListRes{}
	res.FromRepostList(reply)
	return res, nil
}

func (d *Dao) SpaceHistory(c context.Context, general *mdlv2.GeneralParam, param *api.DynSpaceReq, dynType []string, attention *dyncommongrpc.AttentionInfo) (*mdlv2.DynListRes, error) {
	req := &dyngrpc.SpaceHistoryReq{
		Uid:           general.Mid,
		HostUid:       param.HostUid,
		HistoryOffset: param.HistoryOffset,
		AttentionInfo: attention,
		Page:          param.Page,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
			FromSpmid:    "dt.space-dt.0.0.pv",
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
		TypeList: dynType,
	}
	reply, err := d.dynamicGRPC.SpaceHistory(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	res := &mdlv2.DynListRes{}
	res.FromSpaceHistory(reply, general.Mid)
	return res, nil
}

func (d *Dao) UnLogin(c context.Context, general *mdlv2.GeneralParam) (*dyngrpc.UnLoginRsp, error) {
	req := &dyngrpc.UnLoginReq{
		AppMeta: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
		},
	}
	reply, err := d.dynamicGRPC.UnLogin(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) Search(c context.Context, general *mdlv2.GeneralParam, param *api.DynSearchReq, attention *dyncommongrpc.AttentionInfo, ps int) (*dyngrpc.SearchRsp, *mdlv2.DynListRes, error) {
	req := &dyngrpc.SearchReq{
		S:        param.Keyword,
		PageNum:  param.Page,
		PageSize: int32(ps),
		Uid:      general.Mid,
		Meta: &dyncommongrpc.MetaDataCtrl{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			FromSpmid:    "dt.dt-search-result.0.0.pv",
		},
		AttentionInfo: attention,
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
	}
	reply, err := d.dynamicGRPC.Search(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, nil, err
	}
	res := &mdlv2.DynListRes{}
	res.FromSearch(reply, general.Mid)
	return reply, res, nil
}

func (d *Dao) UnLoginFeed(c context.Context, param *api.DynRcmdReq) (*mdlv2.DynListRes, error) {
	req := &dyngrpc.UnLoginFeedReq{
		FakeUid:   param.FakeUid,
		IsRefresh: param.IsRefresh,
	}
	reply, err := d.dynamicGRPC.UnLoginFeed(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	res := &mdlv2.DynListRes{}
	res.FromUnLoginFeed(reply, param.FakeUid)
	return res, nil
}

func (d *Dao) LbsPoiDetail(c context.Context, param *api.LbsPoiReq) (*dyngrpc.LbsPoiDetailRsp, error) {
	req := &dyngrpc.LbsPoiDetailReq{
		Poi:  param.Poi,
		Type: uint64(param.Type),
	}
	reply, err := d.dynamicGRPC.LbsPoiDetail(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) LbsPoiList(c context.Context, general *mdlv2.GeneralParam, param *api.LbsPoiReq) (*mdlv2.DynListRes, error) {
	req := &dyngrpc.LbsPoiListReq{
		Poi:    param.Poi,
		Type:   uint64(param.Type),
		Uid:    uint64(general.Mid),
		Offset: param.Offset,
	}
	reply, err := d.dynamicGRPC.LbsPoiList(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	res := &mdlv2.DynListRes{}
	res.FromLBS(reply)
	return res, nil
}

func (d *Dao) FeedFilter(ctx context.Context, req *dyngrpc.FeedFilterReq) (resp *mdlv2.DynListRes, err error) {
	res, err := d.dynamicGRPC.FeedFilter(ctx, req)
	if err != nil {
		return nil, err
	}
	resp = &mdlv2.DynListRes{
		HasMore: res.HasMore, HistoryOffset: res.Offset,
		Dynamics: make([]*mdlv2.Dynamic, 0, len(res.Dyns)),
	}
	for _, d := range res.Dyns {
		dyn := new(mdlv2.Dynamic)
		dyn.FromDynamic(d)
		resp.Dynamics = append(resp.Dynamics, dyn)
	}
	return
}
