package dynamicV2

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/log"
	"go-common/library/net/metadata"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyncampusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
	"github.com/pkg/errors"
)

func (d *Dao) CampusEntryTab(c context.Context, mid, campusId int64) (*dyncampusgrpc.TabEntryReply, error) {
	return d.homePageSvrClient.TabEntry(c, &dyncampusgrpc.TabEntryReq{
		Uid: mid, CampusId: campusId,
	})
}

func (d *Dao) AlumniDynamics(c context.Context, general *mdlv2.GeneralParam, param *api.CampusRcmdFeedReq, attention *dyncommongrpc.AttentionInfo, zoneID int64) (*mdlv2.DynListRes, error) {
	req := &dyncampusgrpc.AlumniDynamicsReq{
		FromType:  mdlv2.ToCampusFromType(param.GetFromType()),
		Uid:       uint64(general.Mid),
		CampusId:  uint64(param.CampusId),
		FirstTime: uint32(param.FirstTime),
		Page:      uint32(param.Page),
		Scroll:    uint32(param.Scroll),
		ViewedId:  param.ViewDynId,
		VersionCtrl: &dyncommongrpc.VersionCtrlMeta{
			Build:        general.GetBuildStr(),
			Platform:     general.GetPlatform(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			CloseRcmd:    int32(general.GetDisableRcmdInt()),
			FromSpmid:    "dt.campus-moment.0.0.pv",
		},
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers:          true,
			NeedLimitFoldStatement: true,
			NeedBottom:             true,
			NeedTopicInfo:          true,
			NeedLikeIcon:           true,
			NeedRepostNum:          true,
		},
		AttentionInfo: attention,
		Buvid:         general.GetBuvid(),
		Build:         general.GetBuildStr(),
		Plat:          strconv.Itoa(int(model.Plat(general.GetMobiApp(), general.GetDevice()))),
		ZoneId:        strconv.FormatInt(zoneID, 10),
		Network:       general.GetNetWork(),
		Ip:            general.IP,
		MobiApp:       general.GetMobiApp(),
	}
	reply, err := d.dyncampusClient.AlumniDynamics(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	res := &mdlv2.DynListRes{}
	res.FromAlumniDynamics(reply, general.Mid)
	return res, nil
}

func (d *Dao) SchoolSearch(c context.Context, param *api.SchoolSearchReq) ([]*dyncampusgrpc.CampusInfo, error) {
	req := &dyncampusgrpc.SchoolSearchReq{
		FromType: mdlv2.ToCampusFromType(param.GetFromType()),
		Keywords: param.Keyword,
	}
	reply, err := d.dyncampusClient.SchoolSearch(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.Results, nil
}

func (d *Dao) SchoolRecommend(c context.Context, general *mdlv2.GeneralParam, param *api.SchoolRecommendReq) ([]*dyncampusgrpc.CampusInfo, error) {
	req := &dyncampusgrpc.SchoolRecommendReq{
		FromType:  mdlv2.ToCampusFromType(param.GetFromType()),
		Mid:       uint64(general.Mid),
		Latitude:  param.Lat,
		Longitude: param.Lng,
		Ip:        general.IP,
	}
	reply, err := d.dyncampusClient.SchoolRecommend(c, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.Results, nil
}

func (d *Dao) NearbyRcmd(c context.Context, req *api.CampusRcmdReq, general *mdlv2.GeneralParam) (*dyncampusgrpc.PagesReply, error) {
	res, err := d.dyncampusClient.Pages(c, &dyncampusgrpc.PagesReq{
		FromType:   mdlv2.ToCampusFromType(req.GetFromType()),
		Uid:        uint64(general.Mid),
		CampusId:   uint64(req.CampusId),
		CampusName: req.CampusName,
		IpAddr:     metadata.String(c, metadata.RemoteIP),
		Lat:        req.Lat,
		Lng:        req.Lng,
		MetaData: &dyncommongrpc.CmnMetaData{
			Build:    general.GetBuildStr(),
			Platform: general.GetPlatform(),
			MobiApp:  general.GetMobiApp(),
			Buvid:    general.GetBuvid(),
			Device:   general.GetDevice(),
		},
	})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) DynTabShow(c context.Context, req *dyncampusgrpc.TabShowReq) (*dyncampusgrpc.TabShowReply, error) {
	return d.dyncampusClient.TabShow(c, req)
}

func (d *Dao) SetDecision(c context.Context, req *dyncampusgrpc.SetDecisionReq) error {
	_, err := d.dyncampusClient.SetDecision(c, req)
	if err != nil {
		return errors.Wrapf(err, "d.dyncampusClient.SetDecision err(%+v)", req)
	}
	return nil
}

func (d *Dao) SubscribeCampus(c context.Context, req *dyncampusgrpc.SubscribeReq) error {
	_, err := d.dyncampusClient.Subscribe(c, req)
	if err != nil {
		return errors.Wrapf(err, "d.dyncampusClient.Subscribe err(%+v)", req)
	}
	return nil
}

func (d *Dao) SetRcntCampus(c context.Context, req *dyncampusgrpc.SetRecentReq) error {
	_, err := d.dyncampusClient.SetRecent(c, req)
	if err != nil {
		return errors.Wrapf(err, "d.dyncampusClient.SetRecent err(%+v)", req)
	}
	return nil
}

func (d *Dao) OfficialAccounts(c context.Context, param *api.OfficialAccountsReq, general *mdlv2.GeneralParam) (*dyncampusgrpc.OfficialAccountsReply, error) {
	req := &dyncampusgrpc.OfficialAccountsReq{
		FromType:   mdlv2.ToCampusFromType(param.GetFromType()),
		CampusId:   uint64(param.CampusId),
		CampusName: param.CampusName,
		Uid:        uint64(general.Mid),
		Offset:     uint64(param.Offset),
	}
	reply, err := d.dyncampusClient.OfficialAccounts(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) OfficialDynamics(c context.Context, param *api.OfficialDynamicsReq, general *mdlv2.GeneralParam) (*mdlv2.DynListRes, error) {
	req := &dyncampusgrpc.OfficialDynamicsReq{
		FromType:   mdlv2.ToCampusFromType(param.GetFromType()),
		Uid:        uint64(general.Mid),
		CampusId:   uint64(param.CampusId),
		CampusName: param.CampusName,
		Offset:     uint64(param.Offset),
	}
	reply, err := d.dyncampusClient.OfficialDynamics(c, req)
	if err != nil {
		return nil, err
	}
	res := &mdlv2.DynListRes{}
	res.FromOfficialDynamics(reply)
	return res, nil
}

func (d *Dao) CampusRedDot(c context.Context, param *api.CampusRedDotReq, general *mdlv2.GeneralParam) (*dyncampusgrpc.RedDotReply, error) {
	req := &dyncampusgrpc.RedDotReq{
		FromType: mdlv2.ToCampusFromType(param.GetFromType()),
		Uid:      uint64(general.Mid),
		CampusId: uint64(param.CampusId),
	}
	reply, err := d.dyncampusClient.RedDot(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) TopicSquare(c context.Context, param *api.TopicSquareReq, general *mdlv2.GeneralParam) (*dyncampusgrpc.TopicSquareReply, error) {
	req := &dyncampusgrpc.TopicSquareReq{
		FromType: mdlv2.ToCampusFromType(param.GetFromType()),
		Uid:      uint64(general.Mid),
		CampusId: uint64(param.CampusId),
		MetaData: &dyncommongrpc.CmnMetaData{
			Build:    general.GetBuildStr(),
			Platform: general.GetPlatform(),
			MobiApp:  general.GetMobiApp(),
			Buvid:    general.GetBuvid(),
			Device:   general.GetDevice(),
		},
	}
	reply, err := d.dyncampusClient.TopicSquare(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) TopicList(c context.Context, param *api.TopicListReq, general *mdlv2.GeneralParam) (*dyncampusgrpc.TopicListReply, error) {
	req := &dyncampusgrpc.TopicListReq{
		FromType: mdlv2.ToCampusFromType(param.GetFromType()),
		Uid:      uint64(general.Mid),
		CampusId: uint64(param.CampusId),
		Offset:   param.Offset,
		MetaData: &dyncommongrpc.CmnMetaData{
			Build:    general.GetBuildStr(),
			Platform: general.GetPlatform(),
			MobiApp:  general.GetMobiApp(),
			Buvid:    general.GetBuvid(),
			Device:   general.GetDevice(),
		},
	}
	reply, err := d.dyncampusClient.TopicList(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) CampusLikeList(c context.Context, param *api.CampusMateLikeListReq, general *mdlv2.GeneralParam) (*dyncampusgrpc.CampusLikeListRsp, error) {
	req := &dyncampusgrpc.CampusLikeListReq{
		FromType: mdlv2.ToCampusFromType(param.GetFromType()),
		DynId:    param.GetDynamicId(),
		Mid:      general.Mid,
	}
	return d.dyncampusClient.CampusLikeList(c, req)
}

func (d *Dao) CampusFeedback(c context.Context, mid int64, param *api.CampusFeedbackReq) error {
	req := &dyncampusgrpc.FeedbackReq{
		ReqFromType: mdlv2.ToCampusFromType(param.GetFromType()),
		Mid:         mid, FromType: param.From,
	}
	for _, info := range param.GetInfos() {
		req.List = append(req.List, &dyncampusgrpc.FeedbackInfo{
			BizType: int64(info.BizType), BizId: info.BizId,
			CampusId: info.CampusId, Reason: info.Reason,
		})
	}
	_, err := d.dyncampusClient.Feedback(c, req)
	return err
}

func (d *Dao) CampusBillboardMeta(c context.Context, mid, campusID int64, versionCode string, from api.CampusReqFromType) (*mdlv2.CampusBillboardInfo, error) {
	req := &dyncampusgrpc.BillboardReq{
		FromType: mdlv2.ToCampusFromType(from),
		Mid:      mid, CampusId: campusID, VersionCode: versionCode,
	}
	resp, err := d.dyncampusClient.Billboard(c, req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("unexpected nil campus Billboard info")
	}
	ret := &mdlv2.CampusBillboardInfo{}
	ret.FromBillboardReply(resp)
	return ret, nil
}

func (d *Dao) CampusForumSquare(c context.Context, from api.CampusReqFromType, mid, campusID int64, general *mdlv2.GeneralParam) (*mdlv2.CampusForumSquareInfo, error) {
	req := &dyncampusgrpc.ForumSquareReq{
		FromType: mdlv2.ToCampusFromType(from),
		Uid:      mid, CampusId: campusID, MetaData: general.ToDynCmnMetaData(),
	}
	resp, err := d.dyncampusClient.ForumSquare(c, req)
	if err != nil {
		return nil, err
	}
	ret := new(mdlv2.CampusForumSquareInfo)
	ret.FromForumSquareReply(campusID, resp)
	return ret, nil
}

func (d *Dao) CampusForumDynamics(c context.Context, from api.CampusReqFromType, mid, campusID int64, offset string,
	attenions *dyncommongrpc.AttentionInfo, general *mdlv2.GeneralParam) (*mdlv2.CampusForumDynamicsInfo, error) {
	req := &dyncampusgrpc.ForumDynamicsReq{
		FromType: mdlv2.ToCampusFromType(from),
		Uid:      mid, CampusId: campusID, Offset: offset,
		InfoCtrl: &dyncommongrpc.FeedInfoCtrl{
			NeedLikeUsers: true, NeedLimitFoldStatement: true,
			NeedBottom: true, NeedTopicInfo: true,
			NeedLikeIcon: true, NeedRepostNum: true,
		},
		AttentionInfo: attenions, VersionCtrl: general.ToDynVersionCtrlMeta(func(m *dyncommongrpc.VersionCtrlMeta) {
			m.FromSpmid = "dt.campus-moment.0.0.pv"
		}),
	}
	resp, err := d.dyncampusClient.ForumDynamics(c, req)
	if err != nil {
		return nil, err
	}
	ret := new(mdlv2.CampusForumDynamicsInfo)
	ret.FromForumDynamicsReply(resp)
	return ret, nil
}

func (d *Dao) FetchTabSetting(c context.Context, general *mdlv2.GeneralParam) (dyncommongrpc.HomePageTabSttingStatus, error) {
	req := &dyncampusgrpc.FetchTabSettingReq{
		Uid:    general.Mid,
		IpAddr: general.IP,
	}
	reply, err := d.homePageSvrClient.FetchTabSetting(c, req)
	if err != nil {
		return dyncommongrpc.HomePageTabSttingStatus_SETTING_INVALID, err
	}
	return reply.Status, nil
}

func (d *Dao) UpdateTabSetting(c context.Context, general *mdlv2.GeneralParam, param *api.UpdateTabSettingReq) error {
	req := &dyncampusgrpc.UpdateTabSettingReq{
		Uid:    general.Mid,
		Status: dyncommongrpc.HomePageTabSttingStatus(param.Status),
	}
	_, err := d.homePageSvrClient.UpdateTabSetting(c, req)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) CampusSquare(c context.Context, general *mdlv2.GeneralParam, param *api.CampusSquareReq) (*dyncampusgrpc.CampusSquareReply, error) {
	req := &dyncampusgrpc.CampusSquareReq{
		Uid:       general.Mid,
		CampusId:  param.CampusId,
		Latitude:  float32(param.Lat),
		Longitude: float32(param.Lng),
		IpAddr:    general.IP,
	}
	reply, err := d.homePageSvrClient.CampusSquare(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) CampusRecommend(c context.Context, general *mdlv2.GeneralParam, param *api.CampusRecommendReq) (*dyncampusgrpc.CampusRecommendReply, error) {
	req := &dyncampusgrpc.CampusRecommendReq{
		Uid:      general.Mid,
		IpAddr:   general.IP,
		CampusId: param.CampusId,
	}
	reply, err := d.homePageSvrClient.CampusRecommend(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) HomePages(c context.Context, general *mdlv2.GeneralParam, param *api.CampusHomePagesReq) (*dyncampusgrpc.HomePagesReply, error) {
	req := &dyncampusgrpc.HomePagesReq{
		Uid:        general.Mid,
		CampusId:   param.CampusId,
		CampusName: param.CampusName,
		IpAddr:     general.IP,
		Lat:        param.Lat,
		Lng:        param.Lng,
		PageType:   dyncommongrpc.HomePageType(param.PageType),
		MetaData: &dyncommongrpc.CmnMetaData{
			Build:    general.GetBuildStr(),
			Platform: general.GetPlatform(),
			MobiApp:  general.GetMobiApp(),
			Buvid:    general.GetBuvid(),
			Device:   general.GetDevice(),
		},
	}
	reply, err := d.homePageSvrClient.HomePages(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) HomeSubscribe(c context.Context, general *mdlv2.GeneralParam, param *api.HomeSubscribeReq) (*dyncampusgrpc.HomeSubscribeReply, error) {
	req := &dyncampusgrpc.HomeSubscribeReq{
		Uid:        general.Mid,
		CampusId:   param.CampusId,
		CampusName: param.CampusName,
	}
	reply, err := d.homePageSvrClient.HomeSubscribe(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) Identity(c context.Context, general *mdlv2.GeneralParam) (*dyncampusgrpc.CampusIdentityReply, error) {
	req := &dyncampusgrpc.CampusIdentityReq{
		Uid:      general.Mid,
		IpAddr:   general.IP,
		Buvid:    general.GetBuvid(),
		MobiApp:  general.GetMobiApp(),
		Device:   general.GetDevice(),
		Platform: general.GetPlatform(),
		Build:    general.GetBuild(),
		Ua:       general.Device.UserAgent,
	}
	reply, err := d.homePageSvrClient.Identity(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) CampusMngDetail(ctx context.Context, req *dyncampusgrpc.FetchCampusMngDetailReq) (resp *mdlv2.CampusMngDetailRes, err error) {
	reply, err := d.homePageSvrClient.FetchCampusMngDetail(ctx, req)
	if err != nil {
		return nil, err
	}

	ret := &mdlv2.CampusMngDetailRes{
		CampusName: reply.CampusName,
		CampusID:   req.CampusId,
		Items:      reply.Items,
	}
	return ret, nil
}

func (d *Dao) CampusMngSubmit(ctx context.Context, req *dyncampusgrpc.UpdateCampusMngDetailReq) (*mdlv2.CampusMngSubmitRes, error) {
	reply, err := d.homePageSvrClient.UpdateCampusMngDetail(ctx, req)
	if err != nil {
		return nil, err
	}

	ret := &mdlv2.CampusMngSubmitRes{Toast: reply.GetToast()}
	return ret, nil
}

func (d *Dao) CampusQuizList(ctx context.Context, req *dyncampusgrpc.FetchQuestionListReq) (*mdlv2.CampusQuizOperateRes, error) {
	reply, err := d.homePageSvrClient.FetchQuestionList(ctx, req)
	if err != nil {
		return nil, err
	}
	return &mdlv2.CampusQuizOperateRes{List: reply.GetList(), Total: reply.GetTotal()}, nil
}

func (d *Dao) CampusQuizOperate(ctx context.Context, req *dyncampusgrpc.OperateQuestionReq) (*mdlv2.CampusQuizOperateRes, error) {
	_, err := d.homePageSvrClient.OperateQuestion(ctx, req)
	if err != nil {
		return nil, err
	}
	return &mdlv2.CampusQuizOperateRes{}, nil
}
