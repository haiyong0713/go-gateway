package like

import (
	"context"
	"encoding/json"
	"fmt"
	api "git.bilibili.co/bapis/bapis-go/crm/service/profile-manager"
	upratingGRPC "git.bilibili.co/bapis/bapis-go/crm/service/uprating"
	dynapi "git.bilibili.co/bapis/bapis-go/dynamic/service/publish"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/tools/lib/function"
	"net/http"
	"net/url"
	"strconv"
)

const _dynamicCreateURI = "/dynamic_svr/v0/dynamic_svr/icreate"

func (d *Dao) meetEMFThreshold(ctx context.Context, mid int64) (res bool, err error) {
	var ratingReply *upratingGRPC.RatingReply
	for i := 0; i < 3; i++ {
		req := &upratingGRPC.MidReq{Mid: mid}
		ratingReply, err = client.RatingClient.Rating(ctx, req)
		log.Infoc(ctx, "client.RatingClient.Rating:req:(%+v), reply:(%+v)", req, ratingReply)
		if err == nil {
			break
		}
	}
	if err != nil {
		err = errors.Wrap(err, "ratingReply err")
		return
	}
	if ratingReply == nil {
		err = fmt.Errorf("ratingReply is nil")
		return
	}

	// 电磁力等级>=3时有稿件预约权限
	res = ratingReply.Rating != nil && ratingReply.Rating.Level >= d.c.UpActReserveAuthConf.EMFLevelThreshold
	return
}

func (d *Dao) inUpGroup(ctx context.Context, mid int64, upGroupIDs []int64) (res bool, err error) {
	var upgroupReply *api.UpGroupMemberExistReply
	for i := 0; i < 3; i++ {
		req := &api.UpGroupMemberExistReq{Groups: upGroupIDs, Mid: mid}
		upgroupReply, err = client.UpGroupClient.UpGroupMemberExist(ctx, req)
		log.Infoc(ctx, "client.UpGroupClient.UpGroupMemberExist:req:(%+v), reply:(%+v)", req, upgroupReply)
		if err == nil {
			break
		}
	}
	if err != nil {
		return
	}
	if upgroupReply == nil {
		err = errors.Wrap(err, "upgroupReply is nil")
		return
	}
	res = len(upgroupReply.InGroups) != 0
	return
}

func (d *Dao) CanUpActReserveSpecific(ctx context.Context, mid int64, reserveType pb.UpActReserveRelationType) (res map[int64]pb.UpCreateActReserveQualification, err error) {
	res = make(map[int64]pb.UpCreateActReserveQualification)
	// 检测用户是否有发起权限
	if reserveType == pb.UpActReserveRelationType_Archive {
		// 稿件预约使用电磁力+人群包条件校验
		// 优先检测电磁力是否达标
		var emfLevelAuth bool
		emfLevelAuth, err = d.meetEMFThreshold(ctx, mid)
		if err != nil {
			err = errors.Wrapf(err, "d.meetEMFThreshold err, mid: %v", mid)
			return
		}
		//若电磁力不达标，则检查是否在人群包中
		if !emfLevelAuth {
			var upGroupAuth bool
			upGroupAuth, err = d.inUpGroup(ctx, mid, []int64{d.c.UpActReserveAuthConf.ArcUpGroup})
			if err != nil {
				err = errors.Wrapf(err, "d.inUpGroup err, mid: %v", mid)
				return
			}
			if !upGroupAuth {
				return
			}
		}
	} else if reserveType == pb.UpActReserveRelationType_Live {
		var liveAuth bool
		liveAuth, err = d.inUpGroup(ctx, mid, []int64{d.c.UpActReserveAuthConf.LiveUpGroup})
		if err != nil {
			err = errors.Wrapf(err, "d.inUpGroup err, mid: %v", mid)
			return
		}
		if !liveAuth {
			return
		}
	}

	// 查询满足限制策略流转状态
	pbState := d.CanCreateUpActReserveRelationState()
	state := make([]int64, 0)
	for _, v := range pbState {
		state = append(state, int64(v))
	}

	switch reserveType {
	case pb.UpActReserveRelationType_Archive:
		maxNumLimit := d.c.UpActReserveCreateConfig.ArcMaxNumLimit
		var data map[int64]*like.UpActReserveRelationInfo
		data, err = d.RawGetUpActRelationReserveListWithLimit(ctx, mid, []int64{int64(pb.UpActReserveRelationType_Archive)}, state, maxNumLimit)
		if err != nil {
			return
		}
		// 默认不允许
		res[int64(pb.UpActReserveRelationType_Archive)] = pb.UpCreateActReserveQualification_QualificationStrategy
		if int64(len(data)) < maxNumLimit {
			res[int64(pb.UpActReserveRelationType_Archive)] = pb.UpCreateActReserveQualification_QualificationAllow
		}
	case pb.UpActReserveRelationType_Live:
		maxNumLimit := d.c.UpActReserveCreateConfig.LiveMaxNumLimit
		// 官号增加发起限制
		if limit, ok := d.c.UpActReserveCreateMoreLiveReserveMIDs[strconv.FormatInt(mid, 10)]; ok {
			maxNumLimit = limit
		}
		var data map[int64]*like.UpActReserveRelationInfo
		data, err = d.RawGetUpActRelationReserveListWithLimit(ctx, mid, []int64{int64(pb.UpActReserveRelationType_Live)}, state, maxNumLimit)
		if err != nil {
			return
		}
		// 默认不允许
		res[int64(pb.UpActReserveRelationType_Live)] = pb.UpCreateActReserveQualification_QualificationStrategy
		if int64(len(data)) < maxNumLimit {
			res[int64(pb.UpActReserveRelationType_Live)] = pb.UpCreateActReserveQualification_QualificationAllow
		}
	}
	return
}

func (d *Dao) CanCreateUpActReserve(ctx context.Context, mid int64, from pb.UpCreateActReserveFrom) (res map[int64]pb.UpCreateActReserveQualification, err error) {
	res = make(map[int64]pb.UpCreateActReserveQualification)

	arcRes, err := d.CanUpActReserveSpecific(ctx, mid, pb.UpActReserveRelationType_Archive)
	if err != nil {
		err = errors.Wrapf(err, "d.CanUpActReserveSpecific err, mid:%v, type: %v", mid, pb.UpActReserveRelationType_Archive)
	}
	for k, v := range arcRes {
		res[k] = v
	}
	liveRes, err := d.CanUpActReserveSpecific(ctx, mid, pb.UpActReserveRelationType_Live)
	if err != nil {
		err = errors.Wrapf(err, "d.CanUpActReserveSpecific err, mid:%v, type: %v", mid, pb.UpActReserveRelationType_Live)
	}
	for k, v := range liveRes {
		res[k] = v
	}
	return
}

func (d *Dao) UpActReserveRelationInfoRelationState() []pb.UpActReserveRelationState {
	return []pb.UpActReserveRelationState{
		pb.UpActReserveRelationState_UpReserveReject,
		pb.UpActReserveRelationState_UpReserveCancelExpired,
		pb.UpActReserveRelationState_UpReserveCancel,
		pb.UpActReserveRelationState_UpReserveRelated,
		pb.UpActReserveRelationState_UpReserveRelatedOnline,
		pb.UpActReserveRelationState_UpReserveRelatedWaitCallBack,
		pb.UpActReserveRelationState_UpReserveRelatedCallBackCancel,
		pb.UpActReserveRelationState_UpReserveRelatedCallBackDone,
	}
}

func (d *Dao) CanCreateUpActReserveRelationState() []pb.UpActReserveRelationState {
	return []pb.UpActReserveRelationState{
		pb.UpActReserveRelationState_UpReserveRelated,
		pb.UpActReserveRelationState_UpReserveRelatedOnline,
	}
}

func (d *Dao) CancelUpActReserveRelationState() []pb.UpActReserveRelationState {
	return []pb.UpActReserveRelationState{
		pb.UpActReserveRelationState_UpReserveCancel,
		pb.UpActReserveRelationState_UpReserveRelated,
		pb.UpActReserveRelationState_UpReserveRelatedOnline,
	}
}

// 允许up主关闭预约流转状态集
func (d *Dao) CancelUpActReserveSubjectState() []int64 {
	return []int64{
		like.ActSubjectStateCancel,
		like.ActSubjectStateAudit,
		like.ActSubjectStateNormal,
	}
}

// ActSubjectType map UpActReserveRelationType
func (d *Dao) GetActSubjectTypeByUpActReserveRelationType(ctx context.Context, input int64) (output pb.UpActReserveRelationType, err error) {
	switch input {
	case like.UPRESERVATIONARC:
		output = pb.UpActReserveRelationType_Archive
	case like.UPRESERVATIONLIVE:
		output = pb.UpActReserveRelationType_Live
	default:
		err = fmt.Errorf("GetActSubjectTypeByUpActReserveRelationType case type err input(%+v)", input)
		return
	}
	return
}

// UpActReserveRelationType map ActSubjectType
func (d *Dao) GetUpActReserveRelationTypeByActSubjectType(ctx context.Context, input int64) (output int64, err error) {
	switch input {
	case int64(pb.UpActReserveRelationType_Archive):
		output = like.UPRESERVATIONARC
	case int64(pb.UpActReserveRelationType_Live):
		output = like.UPRESERVATIONLIVE
	default:
		err = fmt.Errorf("GetUpActReserveRelationTypeByActSubjectType case type err intput(%+v)", input)
		return
	}
	return
}

// 请求来源参数from 做校验过滤 + 转换
func (d *Dao) GetSourceFromType(ctx context.Context, input int64) (output pb.UpCreateActReserveFrom, err error) {
	switch input {
	case int64(pb.UpCreateActReserveFrom_FromDynamic):
		output = pb.UpCreateActReserveFrom_FromDynamic
	case int64(pb.UpCreateActReserveFrom_FromDanmaku):
		output = pb.UpCreateActReserveFrom_FromDanmaku
	case int64(pb.UpCreateActReserveFrom_FromArchiveCron):
		output = pb.UpCreateActReserveFrom_FromArchiveCron
	case int64(pb.UpCreateActReserveFrom_FromBiliApp):
		output = pb.UpCreateActReserveFrom_FromBiliApp
	case int64(pb.UpCreateActReserveFrom_FromBiliLive):
		output = pb.UpCreateActReserveFrom_FromBiliLive
	case int64(pb.UpCreateActReserveFrom_FROMPCBILILIVE):
		output = pb.UpCreateActReserveFrom_FROMPCBILILIVE
	case int64(pb.UpCreateActReserveFrom_FROMBILIWEB):
		output = pb.UpCreateActReserveFrom_FROMBILIWEB
	default:
		err = fmt.Errorf("GetSourceFromType case type err intput(%+v)", input)
		return
	}
	return
}

// 类型转换 + 参数过滤
func (d *Dao) GetUpActReserveRelationTypeByH5Type(ctx context.Context, input int64) (output pb.UpActReserveRelationType, err error) {
	switch input {
	case int64(pb.UpActReserveRelationType_Archive):
		output = pb.UpActReserveRelationType_Archive
	case int64(pb.UpActReserveRelationType_Live):
		output = pb.UpActReserveRelationType_Live
	default:
		err = fmt.Errorf("GetUpActReserveRelationTypeByH5Type case type err intput(%+v)", input)
		return
	}
	return
}

func (d *Dao) UpActReserveCanBindListState() []pb.UpActReserveRelationState {
	return []pb.UpActReserveRelationState{
		pb.UpActReserveRelationState_UpReserveRelated,
		pb.UpActReserveRelationState_UpReserveRelatedOnline,
	}
}

func (d *Dao) UpActReserveCanBindListArcState() []pb.UpActReserveRelationState {
	return []pb.UpActReserveRelationState{
		pb.UpActReserveRelationState_UpReserveRelated,
		pb.UpActReserveRelationState_UpReserveRelatedOnline,
	}
}

func (d *Dao) UpActReserveBindListState() []pb.UpActReserveRelationState {
	return []pb.UpActReserveRelationState{
		pb.UpActReserveRelationState_UpReserveReject,
		pb.UpActReserveRelationState_UpReserveCancel,
		pb.UpActReserveRelationState_UpReserveRelated,
		pb.UpActReserveRelationState_UpReserveRelatedOnline,
		pb.UpActReserveRelationState_UpReserveRelatedWaitCallBack,
		pb.UpActReserveRelationState_UpReserveRelatedCallBackCancel,
		pb.UpActReserveRelationState_UpReserveRelatedCallBackDone,
	}
}

// 稿件审核中
func (d *Dao) ArcsForbidWaitState() []int64 {
	return []int64{
		like.StateForbidWait,
		like.StateForbidLater,
		like.StateForbidPatched,
		like.StateForbidWaitXcode,
		like.StateForbidAdminDelay,
		like.StateForbidOnlyComment,
		like.StateForbidDispatch,
		like.StateForbidSubmit,
	}
}

// 稿件审核通过 定时发布中
func (d *Dao) ArcsDelayState() []int64 {
	return []int64{
		like.StateForbidUserDelay,
	}
}

// 审核中以及定时发布 不展示数据
func (d *Dao) ArcsUnExpectState() []int64 {
	return append(d.ArcsForbidWaitState(), d.ArcsDelayState()...)
}

// 允许绑定状态
func (d *Dao) BindActReserveState() []pb.UpActReserveRelationState {
	return []pb.UpActReserveRelationState{
		pb.UpActReserveRelationState_UpReserveRelated,
		pb.UpActReserveRelationState_UpReserveRelatedOnline,
	}
}

// 空间触达关联信息状态
func (d *Dao) UpActUserSpaceCardState() []pb.UpActReserveRelationState {
	return []pb.UpActReserveRelationState{
		pb.UpActReserveRelationState_UpReserveRelated,
		pb.UpActReserveRelationState_UpReserveRelatedOnline,
	}
}

func (d *Dao) UpActReserveType() []pb.UpActReserveRelationType {
	return []pb.UpActReserveRelationType{
		pb.UpActReserveRelationType_Archive,
		pb.UpActReserveRelationType_Live,
	}
}

// 核销态的值
func (d *Dao) UpActReserveVerification4CancelState() []pb.UpActReserveRelationState {
	return []pb.UpActReserveRelationState{
		pb.UpActReserveRelationState_UpReserveRelatedWaitCallBack,
		pb.UpActReserveRelationState_UpReserveRelatedCallBackCancel,
		pb.UpActReserveRelationState_UpReserveRelatedCallBackDone,
	}
}

// 直播核销
func (d *Dao) LiveUpActReserveVerification4CancelState() []pb.UpActReserveRelationState {
	return []pb.UpActReserveRelationState{
		pb.UpActReserveRelationState_UpReserveRelated,
	}
}

// 审核过期状态
func (d *Dao) LiveUpActReserveStateExpire() []pb.UpActReserveRelationState {
	return []pb.UpActReserveRelationState{
		pb.UpActReserveRelationState_UpReserveRelated,
	}
}

// 空间触达关联信息状态
func (d *Dao) UpActUserUGCAndStoryState() []pb.UpActReserveRelationState {
	return []pb.UpActReserveRelationState{
		pb.UpActReserveRelationState_UpReserveRelated,
	}
}

func (d *Dao) UpActReserveIsAudit(relation *like.UpActReserveRelationInfo) pb.UpActVisible {
	if relation.Type == int64(pb.UpActReserveRelationType_Archive) {
		if relation.Audit == like.UpActReserveAudit {
			return pb.UpActVisible_OnlyUpVisible
		}
	}
	if relation.Type == int64(pb.UpActReserveRelationType_Live) {
		if relation.LotteryType == 0 { // 非抽奖直播规则同稿件一样
			if relation.Audit == like.UpActReserveAudit {
				return pb.UpActVisible_OnlyUpVisible
			}
		} else { // 抽奖直播 需要动态+抽奖+预约 均审核通过才可见
			if !function.InInt64Slice(relation.Audit, []int64{like.UpActReservePass, like.UpActReservePassDelayAudit}) ||
				relation.LotteryAudit != int64(pb.UpActReserveRelationDependAuditResult_UpActReserveRelationDependAuditResultPass) ||
				relation.DynamicAudit != int64(pb.UpActReserveRelationDependAuditResult_UpActReserveRelationDependAuditResultPass) {
				return pb.UpActVisible_OnlyUpVisible
			}
		}
	}
	return pb.UpActVisible_DefaultVisible
}

func (d *Dao) CanRelateUpActReserve(ctx context.Context, mid int64) (res map[int64]pb.UpCreateActReserveQualification, err error) {
	res = make(map[int64]pb.UpCreateActReserveQualification)
	var activeReserves map[int64]*like.UpActReserveRelationInfo

	for _, t := range []pb.UpActReserveRelationType{pb.UpActReserveRelationType_Archive, pb.UpActReserveRelationType_Live} {
		state := make([]int64, 0)
		for _, v := range d.BindActReserveState() {
			state = append(state, int64(v))
		}
		maxNumLimit := int64(1)
		activeReserves, err = d.RawGetUpActRelationReserveListWithLimit(ctx, mid, []int64{int64(t)}, state, maxNumLimit)
		if err != nil {
			err = errors.Wrap(err, "s.RawGetUpActRelationReserveListWithLimit error")
			return
		}
		if len(activeReserves) > 0 {
			res[int64(t)] = pb.UpCreateActReserveQualification_QualificationAllow
		}
	}

	return
}

// todo:原有http接口不再使用，改为grpc接口
func (d *Dao) CreateDynamic(c context.Context, mid, sid int64) (dynamicID int64, err error) {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(mid, 10))
	// 组装动态数据
	createDynamicCardExtension := &like.CreateDynamicExtension{}
	createDynamicCardExtension.FlagCfg.Reserve.ReserveID = sid
	createDynamicCardExtension.FlagCfg.Reserve.ReserveSource = 0
	extension, _ := json.Marshal(createDynamicCardExtension)
	params.Set("extension", string(extension))
	params.Set("type", strconv.FormatInt(4, 10))
	params.Set("content", "点击预约按钮，不错过直播")
	params.Set("from", like.CreateDynamicFrom)

	var req *http.Request
	if req, err = d.client.NewRequest(http.MethodPost, d.dynamicCreateURL, metadata.String(c, metadata.RemoteIP), params); err != nil {
		return
	}

	var res struct {
		Errno int    `json:"errno"`
		Msg   string `json:"msg"`
		Data  struct {
			DynamicID    int64  `json:"dynamic_id"`
			ErrMsg       string `json:"errmsg"`
			DynamicIDStr string `json:"dynamic_id_str"`
		} `json:"data"`
	}
	err = d.client.Do(c, req, &res)
	log.Infoc(c, "CreateDynamic d.client.Do req:%+v, reply:%+v", req, res)
	if err != nil {
		err = errors.Wrapf(err, "CreateDynamic d.client.Do req:%+v, reply:%+v", req, res)
		return
	}

	if res.Errno != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Errno), "CreateDynamic mid:%d sid:%d", mid, sid)
	}
	dynamicID = res.Data.DynamicID
	return
}

func (d *Dao) CreateShadowDynamic(c context.Context, mid, sid int64) (dynamicID int64, err error) {
	res1 := &dynapi.GetDynamicIdRsp{}
	err = retry.WithAttempts(c, "GetDynamicId", 3, netutil.BackoffConfig{}, func(c context.Context) (err error) {
		res1, err = client.PublishClient.GetDynamicId(c, &dynapi.GetDynamicIdReq{}) // 仅生成动态id，不绑定用户
		log.Infoc(c, "client.PublishClient.GetDynamicId req: &dynapi.GetDynamicIdReq{}, res:%+v", res1)
		return
	})
	if err != nil {
		err = errors.Wrapf(err, "client.PublishClient.GetDynamicId err")
		return
	}
	if res1 == nil || res1.DynId == 0 {
		err = errors.Wrapf(err, "client.PublishClient.GetDynamicId get res err, res:%+v", res1)
		return
	}

	res2 := &dynapi.ICreateReserveDynRsp{}
	req := &dynapi.ICreateReserveDynReq{
		Uid:       mid,
		DynId:     res1.DynId,
		DynType:   dynapi.ReserveCreateDynType_RESERVE_CREATE_DYN_TYPE_WORD,
		ReserveId: sid,
		Content:   "视频更新预告",
		ShowDyn:   true,
	}
	err = retry.WithAttempts(c, "ICreateReserveDyn", 3, netutil.BackoffConfig{}, func(c context.Context) (err error) {
		res2, err = client.PublishClient.ICreateReserveDyn(c, req) // 创建预约动态，支持幂等
		log.Infoc(c, "client.PublishClient.ICreateReserveDyn req:%+v, res:%+v", req, res2)
		if err != nil {
			err = errors.Wrapf(err, "client.PublishClient.ICreateReserveDyn err")
		}
		return
	})

	if err != nil {
		err = errors.Wrapf(err, "retry.WithAttempts client.PublishClient.ICreateReserveDyn err")
	}
	dynamicID = res1.DynId
	return
}
