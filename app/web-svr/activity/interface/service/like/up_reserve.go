package like

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go-common/library/net/netutil"
	"strconv"
	"strings"
	"time"

	audit "git.bilibili.co/bapis/bapis-go/aegis/strategy/service"
	fliapi "git.bilibili.co/bapis/bapis-go/filter/service"
	"git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"github.com/bluele/gcache"
	"go-common/library/log"
	utilRetry "go-common/library/retry"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/tools/lib/function"
)

// 动态附加卡信息
func (s *Service) UpActReserveRelationInfo(ctx context.Context, req *pb.UpActReserveRelationInfoReq) (res *pb.UpActReserveRelationInfoReply, err error) {
	// 返回
	res = &pb.UpActReserveRelationInfoReply{
		List: make(map[int64]*pb.UpActReserveRelationInfo),
	}

	// 获取数据
	relationInfos, err := s.GetUpActReserveRelationReachInfo(ctx, req.Sids, req.Mid)
	if err != nil {
		err = errors.Wrap(err, "s.GetUpActReserveRelationReachInfo err")
		log.Errorc(ctx, err.Error())
		return
	}

	// 获取允许下发附加卡流转状态
	state := s.ConvertPBState2Int64(s.dao.UpActReserveRelationInfoRelationState())
	for _, v := range relationInfos {
		if function.InInt64Slice(int64(v.State), state) {
			res.List[v.Sid] = v
		}
	}

	return
}

// up创建预约活动
func (s *Service) CreateUpActReserve(ctx context.Context, mid int64, req *like.CreateUpActReserveArgs) (lastID int64, err error) {
	var (
		params      = &like.CreateUpActReserveItem{}  // 预约基础数据
		extraParams = &like.CreateUpActReserveExtra{} // relation基础数据
		reply       = &fliapi.FilterV5Reply{}         // 敏感词过滤
		level       int64                             // 敏感词返回级别
	)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	// relation稿件or直播 map act_subject表中的24 or 25 类型
	if params.Type, err = s.dao.GetUpActReserveRelationTypeByActSubjectType(ctx, req.Type); err != nil {
		err = errors.Wrap(err, "s.dao.GetUpActReserveRelationTypeByActSubjectType err")
		return
	}

	// relationType过滤
	relationType, err := s.dao.GetUpActReserveRelationTypeByH5Type(ctx, req.Type)
	if err != nil {
		err = errors.Wrap(err, "s.dao.GetUpActReserveRelationTypeByH5Type err")
		return
	}

	// 区分来源
	relationFrom, err := s.dao.GetSourceFromType(ctx, req.From)
	if err != nil {
		err = errors.Wrap(err, "s.dao.GetSourceFromType err")
		return
	}

	// 权限校验
	data, err := s.dao.CanCreateUpActReserve(ctx, mid, relationFrom)
	if err != nil {
		err = errors.Wrap(err, "s.dao.CanCreateUpActReserve err")
		return
	}
	if _, ok := data[req.Type]; !ok {
		err = ecode.CreateUpActReserveNotInWhiteListErr
		return
	}
	if data[req.Type] != pb.UpCreateActReserveQualification_QualificationAllow {
		err = ecode.CreateUpActReserveExistErr
		return
	}

	// 标题
	var title string
	if title, err = TrimTitle(ctx, req.Title); err != nil {
		err = errors.Wrapf(err, "TrimTitle err err(%+v)", err)
		return
	}
	// 直播添加预约标题：
	if relationType == pb.UpActReserveRelationType_Live {
		title = like.UpActReserveLivePrefix + title
	}

	if err = s.ReserveTimeCheck(req, relationType, extraParams); err != nil {
		err = errors.Wrap(err, "s.ReserveTimeCheck err")
		return
	}

	// 定时发布渠道需要增加
	if relationFrom == pb.UpCreateActReserveFrom_FromArchiveCron {
		if err = s.CheckCronPubReserveParams(ctx, req); err != nil {
			err = errors.Wrap(err, "s.CheckCronPubReserveParams err")
			return
		}
		extraParams.LivePlanStartTime = xtime.Time(req.LivePlanStartTime)
	}

	// 准备数据
	params.Name = title
	params.Stime = xtime.Time(req.Stime)
	params.Etime = xtime.Time(req.Etime)
	extraParams.From = int64(relationFrom)
	extraParams.Audit = like.UpActReservePass
	extraParams.Oid = req.Oid
	extraParams.AuditChannel = like.UpActReserveAuditChannelDefault

	if req.LotteryID != "" && req.LotteryType > 0 {
		extraParams.LotteryID = req.LotteryID
		extraParams.LotteryType = req.LotteryType
		// 只有动态允许发起直播抽奖
		if req.From != int64(pb.UpCreateActReserveFrom_FromDynamic) || relationType != pb.UpActReserveRelationType_Live {
			err = fmt.Errorf("arc no auth req(%+v)", req)
			return
		}
		// 定时开奖暂时没有审核 完全是审核通过
		if extraParams.LotteryType == int64(pb.UpActReserveRelationLotteryType_UpActReserveRelationLotteryTypeCron) {
			extraParams.LotteryAudit = like.UpActReserveDependAudit
		} else {
			err = fmt.Errorf("illegal lottery type req(%+v)", req)
			return
		}
	}

	level, reply, err = s.GetAuditLevelAndReply(ctx, mid, req.From, title, relationFrom)
	if err != nil {
		err = errors.Wrap(err, "s.GetAuditLevelAndReply err")
		return
	}

	var relationState pb.UpActReserveRelationState
	// 根据来源来生成参数
	if req.From == int64(pb.UpCreateActReserveFrom_FromDynamic) {
		params.State = like.ActSubjectStateEdit
		relationState = pb.UpActReserveRelationState_UpReserveEdit
	} else if req.From == int64(pb.UpCreateActReserveFrom_FromDanmaku) || req.From == int64(pb.UpCreateActReserveFrom_FromBiliApp) ||
		req.From == int64(pb.UpCreateActReserveFrom_FromBiliLive) || req.From == int64(pb.UpCreateActReserveFrom_FROMPCBILILIVE) || req.From == int64(pb.UpCreateActReserveFrom_FROMBILIWEB) {
		params.State = like.ActSubjectStateNormal
		relationState = pb.UpActReserveRelationState_UpReserveRelated
		if level == like.SensitiveLevelAudit { // 先审后发
			params.State = like.ActSubjectStateAudit
			extraParams.Audit = like.UpActReserveAudit
			extraParams.AuditChannel = like.UpActReserveAuditChannelPlatform
		}
		if level == like.SensitiveLevelPass { // 先发后审
			extraParams.Audit = like.UpActReservePassDelayAudit
			extraParams.AuditChannel = like.UpActReserveAuditChannelPlatform
		}
	} else if req.From == int64(pb.UpCreateActReserveFrom_FromArchiveCron) {
		params.State = like.ActSubjectStateAudit
		extraParams.Audit = like.UpActReserveAudit
		extraParams.AuditChannel = like.UpActReserveAuditChannelArchive
		relationState = pb.UpActReserveRelationState_UpReserveRelatedOnline
	}

	lastID, err = s.dao.CreateUpActReserveItem(ctx, params, mid, relationType, relationState, extraParams, req.CreateDynamic)
	if err != nil {
		err = errors.Wrap(err, "s.dao.CreateUpActReserveItem err")
		return
	}

	if function.InInt64Slice(level, []int64{like.SensitiveLevelPass, like.SensitiveLevelAudit}) {
		if err = s.Go2Audit(ctx, lastID, mid, title, reply); err != nil {
			err = errors.Wrap(err, "s.Go2Audit err")
			return level, err
		}
	}
	return
}

func (s *Service) ReserveTimeCheck(req *like.CreateUpActReserveArgs, relationType pb.UpActReserveRelationType, extraParams *like.CreateUpActReserveExtra) (err error) {
	ts := function.Now()
	// 开始时间
	req.Stime = ts
	if relationType == pb.UpActReserveRelationType_Archive {
		req.Etime = 2147454847 // 稿件没有结束时间
	}
	if relationType == pb.UpActReserveRelationType_Live {
		if req.LivePlanStartTime == 0 {
			err = ecode.CreateUpActReserveLiveTimeIllegalErr3
			return
		}
		// 不可早于当前时间后5分钟
		if req.LivePlanStartTime < ts+5*60 {
			err = ecode.CreateUpActReserveLiveTimeIllegalErr1
			return
		}
		// 最长不能超过6个月
		if req.LivePlanStartTime > ts+60*60*24*180 {
			err = ecode.CreateUpActReserveLiveTimeIllegalErr2
			return
		}
		// 直播结束时间 预计开播时间 + 半小时 + 稳定5分钟
		req.Etime = req.LivePlanStartTime + s.c.UpActReserveCreateConfig.PlayRange + s.c.UpActReserveCreateConfig.PLayContinue
		extraParams.LivePlanStartTime = xtime.Time(req.LivePlanStartTime)
	}

	if req.Etime <= req.Stime {
		err = ecode.CreateUpActReserveTimeIllegalErr
		return
	}
	return
}

func (s *Service) GetAuditLevelAndReply(ctx context.Context, mid, from int64, title string, relationFrom pb.UpCreateActReserveFrom) (int64, *fliapi.FilterV5Reply, error) {
	if relationFrom == pb.UpCreateActReserveFrom_FromArchiveCron {
		return like.SensitiveLevelNormal, nil, nil
	}

	level, err, reply := FilterTitle(ctx, mid, title)
	if err != nil {
		return level, reply, err
	}

	if function.InInt64Slice(level, []int64{like.SensitiveLevelIntercept20, like.SensitiveLevelIntercept30, like.SensitiveLevelIntercept40}) {
		err = ecode.CreateUpActReserveTitleIllegalErr
		return level, reply, err
	}

	if s.IsInAuditSpecialPeriod(ctx, from, function.Now()) {
		level = like.SensitiveLevelAudit
	}

	return level, reply, nil
}

// up是否可以发起哪些类型的预约
func (s *Service) CanUpCreateActReserve(ctx context.Context, req *pb.CanUpCreateActReserveReq) (res *pb.CanUpCreateActReserveReply, err error) {
	res = new(pb.CanUpCreateActReserveReply)
	data, err := s.dao.CanCreateUpActReserve(ctx, req.Mid, req.From)
	if err != nil {
		err = errors.Wrap(err, "s.dao.CanCreateUpActReserve err")
		return
	}
	res.List = data
	return
}

// 查询用户有哪些状态处于100和120的预约活动
func (s *Service) UpActReserveRelationContinuing(ctx context.Context, mid int64, arg *like.UpActReserveRelationContinuingArg) (list []*pb.UpActReserveRelationInfo, err error) {
	list = make([]*pb.UpActReserveRelationInfo, 0)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	state := make([]int64, 0)
	for _, v := range s.dao.BindActReserveState() {
		state = append(state, int64(v))
	}

	data, err := s.dao.RawGetUpActReserveRelation(ctx, mid, []int64{arg.Type}, state)
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawGetUpActRelationReserve err")
		return
	}

	collectionSids := make([]int64, 0)
	for _, v := range data {
		collectionSids = append(collectionSids, v.Sid)
	}

	list, err = s.GetUpActReserveRelationReachInfo(ctx, collectionSids, mid)

	if arg.InstantID != "" {
		instantIDs := strings.Split(arg.InstantID, ",")
		storeInstantIDs := make([]int64, 0)
		for _, id := range instantIDs {
			var tmp int64
			tmp, err = strconv.ParseInt(id, 10, 64)
			if err != nil {
				log.Errorc(ctx, "s.UpActReserveRelationContinuing format err, mid:%v, instantIDs:%v", mid, instantIDs)
				return
			}
			storeInstantIDs = append(storeInstantIDs, tmp)
		}
		for _, instantID := range storeInstantIDs {
			if !function.InInt64Slice(instantID, collectionSids) {
				var relation *pb.UpActReserveInfoReply
				relation, err = s.UpActReserveInfo(ctx, &pb.UpActReserveInfoReq{Mid: mid, Sids: storeInstantIDs})
				if err != nil {
					log.Errorc(ctx, "s.UpActReserveInfo Err err(%+v)", err)
					return
				}
				if relation == nil {
					log.Errorc(ctx, "s.UpActReserveInfo Err, did not get reserve info, mid:%v, sid:%v", mid, arg.InstantID)
					return
				}
				for _, v := range relation.List {
					list = append(list, &pb.UpActReserveRelationInfo{
						Sid:               v.Sid,
						Title:             v.Title,
						Total:             v.Total,
						Stime:             v.Stime,
						Etime:             v.Etime,
						IsFollow:          v.IsFollow,
						Type:              v.Type,
						LivePlanStartTime: v.LivePlanStartTime,
					})
				}
			}
		}
	}

	if err != nil {
		err = errors.Wrap(err, "s.GetUpActReserveRelationReachInfo err")
		return
	}

	return
}

// 查询用户有哪些可关联的他人的预约
func (s *Service) UpActReserveRelationOthers(ctx context.Context, mid int64, arg *like.UpActReserveRelationOthersArg) (list []*pb.UpActReserveRelationInfo, err error) {
	list = make([]*pb.UpActReserveRelationInfo, 0)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	state := make([]int64, 0)
	for _, v := range s.dao.BindActReserveState() {
		state = append(state, int64(v))
	}

	data, err := s.dao.RawGetUpActReserveRelationOthers(ctx, mid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawGetUpActRelationReserve err")
		return
	}

	collectionSids := make([]int64, 0)
	for _, v := range data {
		collectionSids = append(collectionSids, v.Sid)
	}

	tmpList, err := s.GetUpActReserveRelationReachInfo(ctx, collectionSids, mid)
	if err != nil {
		err = errors.Wrap(err, "s.GetUpActReserveRelationReachInfo err")
		return
	}
	for i := range tmpList {
		if tmpList[i] == nil {
			continue
		}
		if tmpList[i].UpActVisible != pb.UpActVisible_DefaultVisible || int64(tmpList[i].Type) != arg.Type || !function.InInt64Slice(int64(tmpList[i].State), state) {
			continue
		}
		list = append(list, tmpList[i])
	}
	return
}

// 动态发布回调
func (s *Service) CreateUpActReserveRelation(ctx context.Context, req *pb.CreateUpActReserveRelationReq) (res *pb.CreateUpActReserveRelationReply, err error) {
	res = new(pb.CreateUpActReserveRelationReply)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	if req.From != pb.UpCreateActReserveFrom_FromDynamic {
		err = fmt.Errorf("illegal from req(%+v)", req)
		return
	}

	// 查询活动id
	subject, err := s.dao.ActSubjectWithState(ctx, req.Sid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.ActSubjectWithState err")
		return
	}
	if subject == nil || subject.ID <= 0 {
		err = fmt.Errorf("s.dao.ActSubjectWithState result none req(%+v)", req)
		return
	}
	// 状态非编辑态 报错
	if subject.State != like.ActSubjectStateEdit {
		err = fmt.Errorf("subject.State != like.ActSubjectStateEdit subject(%+v)", subject)
		return
	}

	relationType, err := s.dao.GetActSubjectTypeByUpActReserveRelationType(ctx, subject.Type)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.GetActSubjectTypeByUpActReserveRelationType err subject(%+v)", subject)
		return
	}

	// 获取关联关系
	relations, err := s.dao.RawGetUpActReserveRelationInfo(ctx, []int64{req.Sid}, req.Mid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawGetUpActReserveRelationInfo err")
		return
	}
	if _, ok := relations[subject.ID]; !ok {
		err = fmt.Errorf("relation[subject.ID] none relation(%+v) sid(%v)", relations, subject.ID)
		return
	}
	relation := relations[subject.ID]

	// mid 校验
	if relation.Mid != req.Mid {
		err = fmt.Errorf("relation mid != req.Mid relation(%+v) req(%+v)", relation, req)
		return
	}

	if relation.State != int64(pb.UpActReserveRelationState_UpReserveEdit) {
		err = fmt.Errorf("CreateUpActReserveRelation relation.State != pb.UpActReserveRelationState_UpReserveEdit relation(%+v)", relation)
		return
	}

	// 查询是否已经绑定过动态id
	if relation.DynamicID != "" {
		err = ecode.CreateUpActReserveExistDynamicID
		return
	}

	// 发起权限校验
	data, err := s.dao.CanCreateUpActReserve(ctx, req.Mid, req.From)
	if err != nil {
		err = errors.Wrap(err, "s.dao.CanCreateUpActReserve err")
		return
	}
	if _, ok := data[int64(relationType)]; !ok {
		err = ecode.CreateUpActReserveNotInWhiteListErr
		return
	}
	if data[int64(relationType)] != pb.UpCreateActReserveQualification_QualificationAllow {
		err = ecode.CreateUpActReserveExistErr
		return
	}

	if relation.LotteryType > 0 {
		// 抽奖权限校验
		var auth bool
		auth, err = s.dao.GetDynamicLotteryAuth(ctx, req.Mid, relation.Type)
		if err != nil {
			err = errors.Wrap(err, "s.dao.GetDynamicLotteryAuth err")
			return
		}
		if !auth {
			err = fmt.Errorf("s.dao.GetDynamicLotteryAuth no auth")
			return
		}
	}

	// 敏感词检测
	level, err, reply := FilterTitle(ctx, relation.Mid, subject.Name)
	if err != nil {
		err = errors.Wrap(err, "FilterTitle err")
		return
	}

	// 是否需要送审
	var needAudit bool
	// 审核不通过驳回
	if function.InInt64Slice(level, []int64{like.SensitiveLevelIntercept20, like.SensitiveLevelIntercept30, like.SensitiveLevelIntercept40}) {
		err = ecode.CreateUpActReserveTitleIllegalErr
		return
	}

	// 默认insert or update db state
	relationState := pb.UpActReserveRelationState_UpReserveRelated
	subjectState := like.ActSubjectStateNormal
	auditState := like.UpActReservePass
	auditChannelState := like.UpActReserveAuditChannelDefault

	// 审核级别对状态影响
	if level == like.SensitiveLevelAudit { // 先审后发
		needAudit = true
		subjectState = like.ActSubjectStateAudit
		auditState = like.UpActReserveAudit
		auditChannelState = like.UpActReserveAuditChannelPlatform
	} else if level == like.SensitiveLevelPass { // 先发后审
		needAudit = true
		auditState = like.UpActReservePassDelayAudit
		auditChannelState = like.UpActReserveAuditChannelPlatform
	}

	// 特殊时期特殊渠道强制审核时间控制
	if s.IsInAuditSpecialPeriod(ctx, like.SpecialPeriodMustAuditFrom, function.Now()) {
		needAudit = true
		subjectState = like.ActSubjectStateAudit
		auditState = like.UpActReserveAudit
		auditChannelState = like.UpActReserveAuditChannelPlatform
	}

	// 需要送审
	if needAudit {
		err = s.Go2Audit(ctx, relation.Sid, relation.Mid, subject.Name, reply)
		if err != nil {
			err = errors.Wrapf(err, "s.Go2Audit err")
			return
		}
	}

	update := &like.UpActReserveRelationUpdateFields{
		Sid:               req.Sid,
		Mid:               req.Mid,
		SubjectState:      subjectState,
		RelationState:     int64(relationState),
		AuditState:        auditState,
		AuditChannelState: auditChannelState,
		DynamicID:         req.DynamicID,
	}

	// 数据更新
	if err = s.dao.TXUpdateSubjectAndRelationData(ctx, update); err != nil {
		err = errors.Wrapf(err, "s.dao.TXUpdateSubjectAndRelationData err update(%+v)", update)
		return
	}

	// 跟抽奖做绑定
	if relation.LotteryType > 0 {
		if err = utilRetry.WithAttempts(ctx, "BindReserveAndDynamicLottery", 3, netutil.DefaultBackoffConfig, func(tx context.Context) error {
			return s.dao.BindReserveAndDynamicLottery(ctx, req.Mid, relation)
		}); err != nil {
			err = errors.Wrap(err, "retry.WithAttempts s.dao.BindReserveAndDynamicLottery err")
			return
		}
	}

	return
}

// up撤销预约
func (s *Service) CancelUpActReserve(ctx context.Context, req *pb.CancelUpActReserveReq) (res *pb.CancelUpActReserveReply, err error) {
	res = new(pb.CancelUpActReserveReply)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	// 获取关联关系
	relations, err := s.dao.RawGetUpActReserveRelationInfo(ctx, []int64{req.Sid}, req.Mid)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.RawGetUpActReserveRelationInfo err")
		return
	}
	if _, ok := relations[req.Sid]; !ok {
		err = fmt.Errorf("relation[req.Sid] none relations(%+v) req(%+v)", relations, req)
		return
	}
	relation := relations[req.Sid]

	// mid 校验
	if relation.Mid != req.Mid {
		err = fmt.Errorf("relation mid != req.Mid relation(%+v) req(%+v)", relation, req)
		return
	}

	flag1 := false
	allowRelationStateCond := s.ConvertPBState2Int64(s.dao.CancelUpActReserveRelationState())

	for _, state := range allowRelationStateCond {
		if relation.State == state {
			flag1 = true
			break
		}
	}
	if !flag1 {
		err = fmt.Errorf("up cancel reserve state err allowRelationStateCond(%+v) relation(%+v)", allowRelationStateCond, relation)
		return
	}

	// 获取活动基本信息判断状态
	subject, err := s.dao.ActSubjectWithState(ctx, req.Sid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.ActSubjectWithState err")
		return
	}
	if subject == nil || subject.ID <= 0 {
		err = fmt.Errorf("s.dao.ActSubjectWithState nil subject(%+v)", subject)
		return
	}

	flag2 := false
	// 获取允许关闭状态的集合
	allowSubjectState := s.dao.CancelUpActReserveSubjectState()
	for _, state := range allowSubjectState {
		if state == subject.State {
			flag2 = true
			break
		}
	}
	if !flag2 {
		err = fmt.Errorf("up cancel reserve state err allowSubjectState(%+v) subject(%+v)", allowSubjectState, subject)
		return
	}

	subjectState := like.ActSubjectStateCancel
	actRelationState := pb.UpActReserveRelationState_UpReserveCancel
	if err = s.dao.UpActCancelReserve(ctx, req.Sid, req.Mid, subjectState, actRelationState); err != nil {
		err = errors.Wrap(err, "s.dao.UpActCancelReserve err")
		return
	}

	return
}

// 预约活动基本信息 假卡
func (s *Service) UpActReserveInfo(ctx context.Context, req *pb.UpActReserveInfoReq) (res *pb.UpActReserveInfoReply, err error) {
	res = new(pb.UpActReserveInfoReply)
	res.List = make(map[int64]*pb.UpActReserveInfo, 0)

	var (
		subjects     map[int64]*like.SubjectItem
		relations    map[int64]*like.UpActReserveRelationInfo
		reserveInfos map[int64]*like.ActFollowingReply
	)

	// 并发获取数据
	eg := errgroup.WithContext(ctx)
	// 活动基本信息
	eg.Go(func(ctx context.Context) (err error) {
		subjects, err = s.dao.RawActSubjectsWithStateFromMaster(ctx, req.Sids)
		if err != nil {
			log.Errorc(ctx, "s.dao.ActSubjectsWithState Err err(%+v)", err)
			return
		}
		return
	})
	// 获取预约状态和总数
	eg.Go(func(ctx context.Context) (err error) {
		reserveInfos, err = s.ReserveFollowings(ctx, req.Sids, req.Mid)
		if err != nil {
			log.Errorc(ctx, "s.ReserveFollowings Err err(%+v)", err)
			return
		}
		return
	})
	// 请求关联数据
	eg.Go(func(ctx context.Context) (err error) {
		relations, err = s.dao.RawGetUpActReserveRelationInfoFromMaster(ctx, req.Sids)
		if err != nil {
			log.Errorc(ctx, "s.dao.RawGetUpActReserveRelationInfo Err err(%+v)", err)
			return
		}
		return
	})

	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return
	}

	// 整理数据
	for sid, reserveFollow := range reserveInfos {
		if subject, ok1 := subjects[sid]; ok1 {
			if relation, ok2 := relations[sid]; ok2 {
				isFollow := int64(0)
				if reserveFollow.IsFollowing {
					isFollow = 1
				}
				relationType := pb.UpActReserveRelationType(relation.Type)
				tmp := &pb.UpActReserveInfo{
					Sid:               sid,
					Title:             s.HandleTitle(ctx, subject.Name, relationType),
					Total:             reserveFollow.Total,
					Stime:             subject.Stime,
					Etime:             subject.Etime,
					IsFollow:          isFollow,
					Type:              relationType,
					LivePlanStartTime: relation.LivePlanStartTime,
					LotteryType:       pb.UpActReserveRelationLotteryType(relation.LotteryType),
					Upmid:             relation.Mid,
				}
				if relation.LotteryID != "" {
					// 获取奖品信息
					var prizeInfo *pb.UpActReserveRelationPrizeInfo
					prizeInfo, err = s.dao.GetDynamicLotteryPrizeInfo(ctx, relation)
					if err != nil {
						log.Errorc(ctx, "s.dao.GetDynamicLotteryPrizeInfo err prizeInfo(%+v) err(%+v)", prizeInfo, err)
						err = nil
					}
					tmp.PrizeInfo = prizeInfo
				}
				res.List[sid] = tmp
			}
		}
	}

	return
}

// h5查询活动基本信息
func (s *Service) UpActReserveInfoH5(ctx context.Context, sid int64, mid int64) (res *like.UpActReserveInfo, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	// 查询是否存在关联关系数据
	relations, err := s.dao.RawGetUpActReserveRelationInfo(ctx, []int64{sid}, mid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawGetUpActReserveRelationInfo")
		return
	}
	if _, ok := relations[sid]; !ok {
		err = fmt.Errorf("relations[sid] empty relations(%+v) sid(%+v)", relations, sid)
		return
	}
	relation := relations[sid]

	// mid 校验
	if relation.Mid != mid {
		err = fmt.Errorf("relation Mid != req.Mid relation(%+v) mid(%+v)", relation, mid)
		return
	}

	// sid 校验
	if relation.Sid != sid {
		err = fmt.Errorf("relation Sid != sid relation(%+v) sid(%+v)", relation, sid)
		return
	}

	// 活动基本信息
	subjects, err := s.dao.ActSubjectsWithState(ctx, []int64{sid})
	if err != nil {
		err = errors.Wrap(err, "s.dao.ActSubjectsWithState err")
		return
	}
	if _, ok := subjects[sid]; !ok {
		err = fmt.Errorf("subjects[sid] empty subjects(%+v) sid(%+v)", subjects, sid)
		return
	}

	subject := subjects[sid]
	res = &like.UpActReserveInfo{
		ID:                relation.Sid,
		Title:             subject.Name,
		Stime:             subject.Stime,
		Etime:             subject.Etime,
		LivePlanStartTime: relation.LivePlanStartTime,
		LotteryType:       relation.LotteryType,
		LotteryID:         relation.LotteryID,
	}

	// 直播标题前缀处理
	if relation.Type == int64(pb.UpActReserveRelationType_Live) {
		res.Title = strings.Replace(subject.Name, like.UpActReserveLivePrefix, "", 1)
	}

	// actSubject type 映射 upActReserveRelation state
	typ, err := s.dao.GetActSubjectTypeByUpActReserveRelationType(ctx, subject.Type)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.GetActSubjectTypeByUpActReserveRelationType err subject(%+v)", subject)
		return
	}
	res.Type = int64(typ)

	return
}

// 更新up主预约表单
func (s *Service) UpdateUpActReserve(ctx context.Context, mid int64, req *like.UpdateUpActReserveArgs) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	// 查询是否存在关联关系数据
	relations, err := s.dao.RawGetUpActReserveRelationInfo(ctx, []int64{req.ID}, mid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawGetUpActReserveRelationInfo err")
		return
	}
	if _, ok := relations[req.ID]; !ok {
		err = fmt.Errorf("relations[req.ID] empty relations(%+v) req(%+v)", relations, req)
		return
	}
	relation := relations[req.ID]

	// 状态校验
	if relation.State != int64(pb.UpActReserveRelationState_UpReserveEdit) {
		err = fmt.Errorf("relation.State != pb.UpActReserveRelationState_UpReserveEdit relation(%+v)", relations)
		return
	}

	// mid 校验
	if relation.Mid != mid {
		err = fmt.Errorf("relation Mid != mid relation(%+v) mid(%+v)", relation, mid)
		return
	}

	// sid 校验
	if relation.Sid != req.ID {
		err = fmt.Errorf("relation Sid != req.ID relation(%+v) req(%+v)", relation, req)
		return
	}

	params := &like.CreateUpActReserveItem{}
	extraParams := &like.CreateUpActReserveExtra{}

	if params.Type, err = s.dao.GetUpActReserveRelationTypeByActSubjectType(ctx, req.Type); err != nil {
		err = errors.Wrap(err, "s.dao.GetUpActReserveRelationTypeByActSubjectType err")
		return
	}

	relationFrom, err := s.dao.GetSourceFromType(ctx, req.From)
	if relationFrom != pb.UpCreateActReserveFrom_FromDynamic {
		err = fmt.Errorf("relation from illegal")
		return
	}
	if err != nil {
		err = errors.Wrap(err, "s.dao.GetSourceFromType err")
		return
	}

	// 标题
	var title string
	if title, err = TrimTitle(ctx, req.Title); err != nil {
		err = errors.Wrap(err, "TrimTitle err")
		return
	}

	// 直播添加预约标题：
	if relation.Type == int64(pb.UpActReserveRelationType_Live) {
		title = like.UpActReserveLivePrefix + title
	}

	// 敏感词
	level, err, _ := FilterTitle(ctx, mid, title)
	if err != nil {
		err = errors.Wrap(err, "FilterTitle err")
		return
	}

	if function.InInt64Slice(level, []int64{like.SensitiveLevelIntercept20, like.SensitiveLevelIntercept30, like.SensitiveLevelIntercept40}) {
		err = ecode.CreateUpActReserveTitleIllegalErr
		return
	}

	now := function.Now()

	// 开始时间
	req.Stime = now

	// 稿件没有结束时间
	if relation.Type == int64(pb.UpActReserveRelationType_Archive) {
		req.Etime = 2147454847
	} else if relation.Type == int64(pb.UpActReserveRelationType_Live) {
		// 开始直播时间参数校验
		if req.LivePlanStartTime == 0 {
			err = ecode.CreateUpActReserveLiveTimeIllegalErr3
			return
		}
		// 不可早于当前时间后5分钟。
		if req.LivePlanStartTime < now+5*60 {
			err = ecode.CreateUpActReserveLiveTimeIllegalErr1
			return
		}
		// 最长不能超过6个月
		if req.LivePlanStartTime > now+60*60*24*180 {
			err = ecode.CreateUpActReserveLiveTimeIllegalErr2
			return
		}

		// 直播有结束时间 预计开播时间 + 半小时 + 稳定5分钟
		req.Etime = req.LivePlanStartTime + s.c.UpActReserveCreateConfig.PlayRange + s.c.UpActReserveCreateConfig.PLayContinue
		extraParams.LivePlanStartTime = xtime.Time(req.LivePlanStartTime)
	}

	if req.Etime <= req.Stime {
		err = ecode.CreateUpActReserveTimeIllegalErr
		return
	}

	params.Name = title
	params.Stime = xtime.Time(req.Stime)
	params.Etime = xtime.Time(req.Etime)
	extraParams.LotteryType = req.LotteryType
	extraParams.LotteryID = req.LotteryID

	if extraParams.LotteryType > 0 && extraParams.LotteryID != "" {
		extraParams.LotteryAudit = 1
	}

	if err = s.dao.UpdateUpActReserveItem(ctx, relation.Sid, params, extraParams); err != nil {
		err = errors.Wrap(err, "s.dao.UpdateUpActReserveItem err(%v)")
		return
	}

	return
}

func TrimTitle(ctx context.Context, input string) (output string, err error) {
	// 标题
	input = strings.TrimSpace(input)
	if input == "" {
		err = ecode.CreateUpActReserveTitleEmptyErr
		return
	}
	output = input
	return
}

func FilterTitle(ctx context.Context, mid int64, content string) (level int64, err error, reply *fliapi.FilterV5Reply) {
	reply = &fliapi.FilterV5Reply{}
	req := &fliapi.FilterReq{
		Area:    "bullet_up",
		Mid:     mid,
		Message: content,
	}
	if reply, err = client.FilterClient.FilterV5(ctx, req); err != nil {
		err = errors.Wrapf(err, "FilterTitle err(%+v) req(%+v) reply(%+v)", err, req, reply)
		return
	}
	log.Infoc(ctx, "FilterTitle content(%s) mid(%d) reply(%+v)", content, mid, reply)

	// 区分社区 不送审 4 和 5 Source: 4, Desc: "运营规避"  Source: 5, Desc: "OGV运营"
	sourceLevels := []int64{0}
	if len(reply.Rules) > 0 {
		for _, v := range reply.Rules {
			if v.Source == 4 || v.Source == 5 {
				continue
			}
			if v.Level >= 0 {
				sourceLevels = append(sourceLevels, int64(v.Level))
			}
		}
	}

	// 不通source优先级整合
	if function.InInt64Slice(int64(like.SensitiveLevelIntercept20), sourceLevels) { // 标记打回
		level = like.SensitiveLevelIntercept20
	} else if function.InInt64Slice(int64(like.SensitiveLevelIntercept30), sourceLevels) { // 标记打回
		level = like.SensitiveLevelIntercept30
	} else if function.InInt64Slice(int64(like.SensitiveLevelIntercept40), sourceLevels) { // 标记打回
		level = like.SensitiveLevelIntercept40
	} else if function.InInt64Slice(int64(like.SensitiveLevelAudit), sourceLevels) { // 先审后发
		level = like.SensitiveLevelAudit
	} else if function.InInt64Slice(int64(like.SensitiveLevelPass), sourceLevels) { // 先发后审
		level = like.SensitiveLevelPass
	} else if function.InInt64Slice(int64(like.SensitiveLevelTest), sourceLevels) { // 测试标记
		level = like.SensitiveLevelTest
	} else if function.InInt64Slice(int64(like.SensitiveLevelNormal), sourceLevels) { // 通过
		level = like.SensitiveLevelNormal
	} else {
		err = fmt.Errorf("unknow sourceLevels (%+v)", sourceLevels)
	}

	return
}

// 允许关联预约活动列表
func (s *Service) UpActReserveCanBindList(ctx context.Context, req *pb.UpActReserveCanBindListReq) (reply *pb.UpActReserveCanBindListReply, err error) {
	reply = new(pb.UpActReserveCanBindListReply)
	reply.List = make([]*pb.UpActReserveCanBindInfo, 0)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	//state := make([]int64, 0)
	//if req.From == pb.UpCreateActReserveFrom_FromArchive {
	//	state = s.ConvertPBState2Int64(s.dao.UpActReserveCanBindListArcState())
	//} else if req.From == pb.UpCreateActReserveFrom_FromDanmaku {
	//	state = s.ConvertPBState2Int64(s.dao.UpActReserveCanBindListState())
	//}
	state := s.ConvertPBState2Int64(s.dao.UpActReserveCanBindListState())

	// 获取允许绑定的预约id
	candidate, err := s.dao.RawGetUpActReserveRelation(ctx, req.Mid, []int64{int64(req.Type)}, state)
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawGetUpActReserveRelation err")
		return
	}

	relations := make(map[int64]*like.UpActReserveRelationInfo)
	// 转换map用于后续查找
	for _, v := range candidate {
		relations[v.Sid] = v
	}

	// 允许展现的sid
	sids := make([]int64, 0)

	// 稿件类型
	if req.Type == pb.UpActReserveRelationType_Archive {
		// 已绑资源的oid
		checkOids := make([]int64, 0)
		// oid映射捆绑的sid
		oid2Sid := make(map[int64]int64)

		// 提取资源id
		for _, relation := range candidate {
			convertedStrInt64, _ := strconv.ParseInt(relation.Oid, 10, 64)
			if convertedStrInt64 == 0 {
				sids = append(sids, relation.Sid)
			} else {
				// 需要校验oid
				checkOids = append(checkOids, convertedStrInt64)
				oid2Sid[convertedStrInt64] = relation.Sid
			}
		}

		if len(checkOids) > 0 {
			videoReq := &service.ArchiveSimpleBatchReq{Aids: checkOids}
			videoReply, err := client.VideoClient.ArchiveSimpleBatch(ctx, videoReq)
			if err != nil {
				err = errors.Wrapf(err, "client.VideoClient.ArchiveSimpleBatch err(%+v) req(%+v) reply(%+v)", err, videoReq, videoReply)
				return reply, err
			}

			// 去除不合法的稿件状态
			showOids := make([]int64, 0)
			state = s.dao.ArcsUnExpectState()
			for _, arc := range videoReply.Arcs {
				// (稿件开放浏览 && 稿件审核中 && 稿件过审处于定时发布中) 均不可以换绑
				if arc.State < 0 && !function.InInt64Slice(int64(arc.State), state) {
					showOids = append(showOids, arc.Aid)
				}
			}
			// 提取sid
			for _, v := range showOids {
				if _, ok := oid2Sid[v]; ok {
					sids = append(sids, oid2Sid[v])
				}
			}
		}
	} else if req.Type == pb.UpActReserveRelationType_Live {
		// 提取资源id
		for _, relation := range candidate {
			sids = append(sids, relation.Sid)
		}
	}

	if len(sids) == 0 {
		return
	}

	subjects, err := s.dao.RawActSubjectsWithState(ctx, sids)
	if err != nil {
		err = errors.Wrap(err, "s.dao.ActSubjectsWithState err")
		return reply, err
	}
	totals, err := s.GetActSubjectsReserveIDsFollowTotalByOptimization(ctx, sids)
	if err != nil {
		err = errors.Wrap(err, "s.GetActSubjectsReserveIDsFollowTotalByOptimization err")
		return reply, err
	}
	for _, sid := range sids {
		if subject, ok := subjects[sid]; ok {
			if !function.InInt64Slice(subject.State, []int64{like.ActSubjectStateNormal, like.ActSubjectStateAudit}) {
				continue
			}
			if _, ok1 := totals[sid]; ok1 {
				if relation, ok2 := relations[sid]; ok2 {
					reply.List = append(reply.List, &pb.UpActReserveCanBindInfo{Title: subject.Name, Sid: subject.ID, Total: totals[sid], LivePlanStartTime: relation.LivePlanStartTime})
				}
			}
		}
	}

	return
}

// 查询绑定关系
func (s *Service) UpActReserveBindList(ctx context.Context, req *pb.UpActReserveBindListReq) (reply *pb.UpActReserveBindListReply, err error) {
	reply = new(pb.UpActReserveBindListReply)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	state := s.ConvertPBState2Int64(s.dao.UpActReserveBindListState())
	// 获取资源关联数据
	relation, err := s.dao.RawGetUpActRelationReserveByOIDLatestData(ctx, req.Oid, []int64{int64(req.Type)}, state)
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawGetUpActRelationReserveByOIDLatestData err")
		return
	}
	if relation == nil || relation.Sid == 0 {
		return
	}
	subject, err := s.dao.ActSubjectWithState(ctx, relation.Sid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.ActSubjectWithState err")
		return
	}
	reply.Sid = relation.Sid
	reply.Title = subject.Name
	reply.State = relation.State

	return
}

// 绑定关系
func (s *Service) BindActReserve(ctx context.Context, req *pb.BindActReserveReq) (reply *pb.BindActReserveReply, err error) {
	reply = new(pb.BindActReserveReply)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	// 只有100和120状态允许换绑
	state := s.ConvertPBState2Int64(s.dao.BindActReserveState())
	// 获取本条记录 没有指定type 拿出来需要做校验
	relations, err := s.dao.RawUpActReserveRelationInfoWithState(ctx, []int64{req.Sid}, state)
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawUpActReserveRelationInfoWithState err")
		return
	}

	if _, ok := relations[req.Sid]; !ok {
		err = fmt.Errorf("relations[req.Sid] empty relations(%+v) req(%+v)", relations, req)
		return
	}

	relation := relations[req.Sid]
	if relation.Mid != req.Mid {
		err = fmt.Errorf("user mid err relation(%+v) req(%+v)", relation, req)
		return
	}

	if relation.Type != int64(req.Type) {
		err = fmt.Errorf("type err relation(%+v) req(%+v)", relation, req)
		return
	}

	// 查询oid关联老的state为120的数据
	oldRelation, err := s.dao.RawGetUpActRelationReserveByOID(ctx, req.Oid, int64(req.Type), int64(pb.UpActReserveRelationState_UpReserveRelatedOnline))
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawGetUpActRelationReserveByOID err")
		return
	}

	// 是否存在老数据
	// 如果存在 进行事务处理 允许重复绑定 不存在则直接绑定
	if relation.Sid > 0 {
		// 找到老数据这一行
		if err = s.dao.UpdateUpActReserveBindUnion(ctx, req.Oid, oldRelation.Sid, relation.Sid); err != nil {
			err = errors.Wrap(err, "s.dao.UpdateUpActReserveBindUnion err")
			return
		}
	} else {
		if err = s.dao.UpdateUpActReserveBind(ctx, req.Oid, int64(pb.UpActReserveRelationState_UpReserveRelatedOnline), req.Sid); err != nil {
			err = errors.Wrap(err, "s.dao.UpdateUpActReserveBind err")
			return
		}
	}

	return
}

// 获取空间卡
func (s *Service) GetUpActUserSpaceCard(ctx context.Context, req *pb.UpActUserSpaceCardReq) (reply *pb.UpActUserSpaceCardReply, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	if req.Upmid == 0 {
		err = fmt.Errorf("upmid err req(%+v)", req)
		return
	}

	reply = new(pb.UpActUserSpaceCardReply)
	reply.List = make([]*pb.UpActReserveRelationInfo, 0)

	// 查询符合条件的sid
	sids, err := s.dao.GetUpActReserveRelationInfo4SpaceCardIDs(ctx, req.Upmid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.GetUpActReserveRelationInfo4SpaceCardIDs")
		return
	}

	// 获取数据
	relationInfos, err := s.GetUpActReserveRelationReachInfo(ctx, sids, req.Mid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.GetUpActReserveRelationReachInfo")
		return
	}

	for _, relationInfo := range relationInfos {
		reply.List = append(reply.List, relationInfo)
	}

	return
}

// 构造触达卡片信息
func (s *Service) GetUpActReserveRelationReachInfo(ctx context.Context, sids []int64, mid int64) (reply []*pb.UpActReserveRelationInfo, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	// 获取关联基本数据
	relations, err := s.dao.GetUpActReserveRelationInfoBySid(ctx, sids)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.GetUpActReserveRelationInfoBySid err sids(%+v)", sids)
		return
	}

	var (
		actSubjects  = make(map[int64]*like.SubjectItem)
		reservesInfo = make(map[int64]*like.ActFollowingReply)
	)

	reply = make([]*pb.UpActReserveRelationInfo, 0)

	// 并发获取活动信息
	g := errgroup.WithContext(ctx)
	// 活动基本信息
	g.Go(func(ctx context.Context) (err error) {
		actSubjects, err = s.dao.ActSubjectsWithState(ctx, sids)
		if err != nil {
			err = errors.Wrap(err, "s.dao.ActSubjectsWithState err")
			return
		}
		return
	})
	// 获取预约人数和预约状态
	g.Go(func(ctx context.Context) (err error) {
		reservesInfo, err = s.ReserveFollowings(ctx, sids, mid)
		if err != nil {
			err = errors.Wrap(err, "s.ReserveFollowings err")
			return
		}
		return
	})
	if err = g.Wait(); err != nil {
		err = errors.Wrap(err, "g.Wait() err")
		return
	}
	// 组装数据
	for sid, relation := range relations {
		if subject, ok1 := actSubjects[sid]; ok1 {
			if reserve, ok2 := reservesInfo[sid]; ok2 {
				isFollow := int64(0)
				if mid > 0 && reserve.IsFollowing {
					isFollow = 1
				}
				relationType := pb.UpActReserveRelationType(relation.Type)
				tmp := &pb.UpActReserveRelationInfo{
					Sid:                   relation.Sid,
					Title:                 s.HandleTitle(ctx, subject.Name, relationType),
					Total:                 reserve.Total,
					Stime:                 subject.Stime,
					Etime:                 subject.Etime,
					IsFollow:              isFollow,
					State:                 pb.UpActReserveRelationState(relation.State),
					Oid:                   relation.Oid,
					Type:                  relationType,
					Upmid:                 relation.Mid,
					ReserveRecordCtime:    reserve.Ctime,
					LivePlanStartTime:     relation.LivePlanStartTime,
					UpActVisible:          s.dao.UpActReserveIsAudit(relation),
					LotteryType:           pb.UpActReserveRelationLotteryType(relation.LotteryType),
					DynamicId:             relation.DynamicID,
					ReserveTotalShowLimit: s.c.ReserveTotalShowLimitsMap[strconv.FormatInt(relation.Type, 10)],
				}
				if relation.LotteryID != "" {
					// 获取奖品信息
					var prizeInfo *pb.UpActReserveRelationPrizeInfo
					prizeInfo, err = s.dao.GetDynamicLotteryPrizeInfo(ctx, relation)
					if err != nil {
						log.Errorc(ctx, "s.dao.GetDynamicLotteryPrizeInfo err prizeInfo(%+v) err(%+v)", prizeInfo, err)
						err = nil
					}
					tmp.PrizeInfo = prizeInfo
				}
				reply = append(reply, tmp)
			}
		}
	}

	return
}

func (s *Service) Go2Audit(ctx context.Context, sid int64, mid int64, title string, filter *fliapi.FilterV5Reply) (err error) {
	// 审核服务历史遗留问题 需要转换rpc结构体
	positions := make([]*audit.PosProto, 0)
	rules := make([]*audit.RuleProto, 0)
	if len(filter.Positions) > 0 {
		for _, v := range filter.Positions {
			positions = append(positions, &audit.PosProto{From: v.From, To: v.To})
		}
	}
	if len(filter.Rules) > 0 {
		for _, v := range filter.Rules {
			rules = append(rules, &audit.RuleProto{
				Id:       v.ID,
				Mode:     v.Mode,
				Rule:     v.Rule,
				Area:     v.Area,
				Key:      v.Key,
				Level:    v.Level,
				Stime:    v.STime,
				Etime:    v.ETime,
				Comment:  v.Comment,
				Cid:      v.CID,
				Type:     v.Type,
				Source:   v.Source,
				RuleType: v.RuleType,
			})
		}
	}

	req := &audit.AegisProcessReq{
		BusinessId: s.c.UpActReserveAudit.BizID1,
		AddInfo: &audit.AegisAddInfo{
			BusinessId: s.c.UpActReserveAudit.BizID2,
			NetId:      s.c.UpActReserveAudit.NetID,
			Oid:        strconv.FormatInt(sid, 10),
			Mid:        mid,
			Content:    title,
			Filter: &audit.FilterReply{
				Result:    filter.Result,
				Level:     filter.Level,
				Positions: positions,
				Rules:     rules,
				Reason:    filter.Reason,
			},
		},
	}

	for i := 0; i < 3; i++ {
		_, err = client.AuditClient.AegisProcess(ctx, req)
		log.Infoc(ctx, "client.AuditClient.AegisProcess req(%+v)", req)
		if err == nil {
			break
		}
	}

	if err != nil {
		err = errors.Wrapf(err, "client.AuditClient.AegisProcess err req(%+v)", req)
		return
	}

	return
}

// 审核回调信息
func (s *Service) UpActReserveAuditCallBack(ctx context.Context, sid int64, state int64) (res int, err error) {
	log.Infoc(ctx, "UpActReserveAuditCallBack sid(%+v) state(%+v)", sid, state)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	if !function.InInt64Slice(state, []int64{like.UpActReserveAuditPass, like.UpActReserveAuditReject}) {
		err = fmt.Errorf("wrong audit type state(%+v)", state)
		return
	}

	relations, err := s.dao.RawGetUpActReserveRelationInfoBySid(ctx, []int64{sid})
	if err != nil {
		err = errors.Wrapf(err, "s.dao.RawGetUpActReserveRelationInfoBySid sid(%+v)", sid)
		return
	}
	if _, ok := relations[sid]; !ok {
		err = fmt.Errorf("relations[sid] nil relations(%+v) sid(%+v)", relations, sid)
		return
	}

	relation := relations[sid]

	subject, err := s.dao.RawActSubjectWithState(ctx, sid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawActSubjectWithState err")
		return
	}
	if subject == nil || subject.ID == 0 {
		err = fmt.Errorf("subject nil subject(%+v)", subject)
		return
	}

	// 如果审核通过的话
	// 如果是稿件或直播
	if relation.Type == int64(pb.UpActReserveRelationType_Archive) || relation.Type == int64(pb.UpActReserveRelationType_Live) {
		if state == like.UpActReserveAuditPass {
			// 幂等
			if relation.Audit == like.UpActReservePass {
				return
			}
			// 如果非审核平台渠道 报错
			if relation.AuditChannel != like.UpActReserveAuditChannelPlatform {
				err = fmt.Errorf("audit_channel err relation(%+v)", relation)
				return
			}
			if function.InInt64Slice(relation.Audit, []int64{like.UpActReservePassDelayAudit, like.UpActReserveAudit}) {
				update := &like.UpActReserveRelationUpdateFields{
					Sid:               sid,
					Mid:               relation.Mid,
					SubjectState:      like.ActSubjectStateNormal,
					RelationState:     relation.State,
					AuditState:        like.UpActReservePass,
					AuditChannelState: like.UpActReserveAuditChannelDefault,
					DynamicID:         relation.DynamicID,
				}

				err = s.dao.TXUpdateSubjectAndRelationData(ctx, update)
				if err != nil {
					err = errors.Wrapf(err, "s.dao.TXUpdateSubjectAndRelationData err update(%+v)", update)
					return
				}
			} else {
				err = fmt.Errorf("illegal type relation(%+v) subject(%+v) sid(%+v) state(%+v)", relation, subject, sid, state)
				return
			}
		}

		// 审核拒绝 任何情况 都直接拒绝掉
		if state == like.UpActReserveAuditReject {
			// 幂等
			if relation.Audit == like.UpActReserveReject {
				return
			}
			// 如果非审核平台渠道 报错
			if relation.AuditChannel != like.UpActReserveAuditChannelPlatform {
				err = fmt.Errorf("audit_channel err relation(%+v)", relation)
				return
			}
			update := &like.UpActReserveRelationUpdateFields{
				Sid:               sid,
				Mid:               relation.Mid,
				SubjectState:      like.ActSubjectStateReject,
				RelationState:     int64(pb.UpActReserveRelationState_UpReserveReject),
				AuditState:        like.UpActReserveReject,
				AuditChannelState: like.UpActReserveAuditChannelDefault,
				DynamicID:         relation.DynamicID,
			}
			err = s.dao.TXUpdateSubjectAndRelationData(ctx, update)
			if err != nil {
				err = errors.Wrapf(err, "s.dao.TXUpdateSubjectAndRelationData err update(%+v)", update)
				return
			}
		}
	} else {
		err = fmt.Errorf("illegal relation type relations(%+v)", relation)
		return
	}

	return
}

// 核销
func (s *Service) UpActReserveVerification4Cancel(ctx context.Context, req *pb.UpActReserveVerification4CancelReq) (reply *pb.UpActReserveVerification4CancelReply, err error) {
	reply = new(pb.UpActReserveVerification4CancelReply)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	if req.Sid <= 0 {
		err = fmt.Errorf("req.Sid illegal req(%+v)", req)
		return
	}

	// 获取关联信息
	relationInfos, err := s.dao.RawGetUpActReserveRelationInfoBySid(ctx, []int64{req.Sid})
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawGetUpActReserveRelationInfoBySid err")
		return
	}

	// 获取数据
	if _, ok := relationInfos[req.Sid]; !ok {
		err = fmt.Errorf("relationInfos[req.Sid] empty relationInfos(%+v) req(%+v)", relationInfos, req)
		return
	}

	relationInfo := relationInfos[req.Sid]

	// 参数校验
	if req.Mid > 0 {
		// up主校验
		if req.Mid != relationInfo.Mid {
			err = fmt.Errorf("mid error relationInfo(%+v) req(%+v) ", relationInfo, req)
			return
		}
	}

	// 类型校验
	if req.Type != pb.UpActReserveRelationType(relationInfo.Type) {
		err = fmt.Errorf("type error relationInfo(%+v) req(%+v) ", relationInfo, req)
		return
	}

	// 直播的话 oid不能为空
	if req.Type == pb.UpActReserveRelationType_Live {
		if req.Oid == "" {
			err = fmt.Errorf("oid empty req(%+v)", req)
			return
		}
	}

	// 状态核验
	state := s.ConvertPBState2Int64(s.dao.UpActReserveVerification4CancelState())
	if !function.InInt64Slice(int64(req.State), state) {
		err = fmt.Errorf("req.State not in state req(%+v) state(%+v)", req, state)
		return
	}

	relationInfoState := pb.UpActReserveRelationState(relationInfo.State)

	// 状态处理 DB现存状态 == req请求改变状态 证明是幂等操作忽略
	if req.State == relationInfoState {
		return
	}

	// 首先稿件
	if req.Type == pb.UpActReserveRelationType_Archive {
		// 开始核销130 前置条件必须为120 允许请求重试的话 包含状态130
		if req.State == pb.UpActReserveRelationState_UpReserveRelatedWaitCallBack {
			// 已经150 还来调用130 返回特定错误码 给私信需求用
			if relationInfoState == pb.UpActReserveRelationState_UpReserveRelatedCallBackDone {
				err = ecode.CreateUpActReserveVerification4CancelStateErr
				return
			}

			if relationInfoState != pb.UpActReserveRelationState_UpReserveRelatedOnline &&
				relationInfoState != pb.UpActReserveRelationState_UpReserveRelatedWaitCallBack {
				err = fmt.Errorf("before state condition error req(%+v) relationInfo(%+v)", req, relationInfo)
				return
			}
		}
		// 如果状态是140的话 前置条件必须是130或140
		if req.State == pb.UpActReserveRelationState_UpReserveRelatedCallBackCancel {
			if relationInfoState != pb.UpActReserveRelationState_UpReserveRelatedWaitCallBack &&
				relationInfoState != pb.UpActReserveRelationState_UpReserveRelatedCallBackCancel {
				err = fmt.Errorf("before state condition error req(%+v) relationInfo(%+v)", req, relationInfo)
				return
			}
		}
		// 状态150的话 前置条件必须是130 140 150
		if req.State == pb.UpActReserveRelationState_UpReserveRelatedCallBackDone {
			if relationInfoState != pb.UpActReserveRelationState_UpReserveRelatedWaitCallBack &&
				relationInfoState != pb.UpActReserveRelationState_UpReserveRelatedCallBackCancel &&
				relationInfoState != pb.UpActReserveRelationState_UpReserveRelatedCallBackDone {
				err = fmt.Errorf("before state condition error req(%+v) relationInfo(%+v)", req, relationInfo)
				return
			}
		}

		// 变更DB状态
		if err = s.dao.UpActReserveRelationCancel4Arc(ctx, relationInfo.Mid, req.Sid, function.Now(), req.State); err != nil {
			err = errors.Wrap(err, "s.dao.UpActReserveRelationCancel4Arc err")
			return
		}

	} else if req.Type == pb.UpActReserveRelationType_Live {
		// 直播核销130 前置条件100
		if req.State == pb.UpActReserveRelationState_UpReserveRelatedWaitCallBack {
			if relationInfo.State != int64(pb.UpActReserveRelationState_UpReserveRelated) {
				err = fmt.Errorf("before state condition error req(%+v) relationInfo(%+v)", req, relationInfo)
				return
			}
			// 变更DB状态 以及插入场次id
			if err = s.dao.UpActReserveRelationCancel4Live(ctx, relationInfo.Mid, req.Sid, function.Now(), req.State, req.Oid); err != nil {
				err = errors.Wrap(err, "s.dao.UpActReserveRelationCancel4Live err")
				return
			}
		} else if req.State == pb.UpActReserveRelationState_UpReserveRelatedCallBackDone {
			if relationInfo.State != int64(pb.UpActReserveRelationState_UpReserveRelatedWaitCallBack) {
				err = fmt.Errorf("before state condition error req(%+v) relationInfo(%+v)", req, relationInfo)
				return
			}
			if err = s.dao.UpdateUpActReserveState(ctx, int64(req.State), req.Sid); err != nil {
				err = errors.Wrap(err, "s.dao.UpdateUpActReserveState")
				return
			}
		}
	}

	return
}

// 直播获取数据
func (s *Service) UpActReserveRelationInfoByTime(ctx context.Context, req *pb.UpActReserveRelationInfoByTimeReq) (reply *pb.UpActReserveRelationInfoByTimeReply, err error) {
	reply = &pb.UpActReserveRelationInfoByTimeReply{
		List: make(map[int64]*pb.UpActReserveRelationInfo),
	}

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	state := s.ConvertPBState2Int64(s.dao.LiveUpActReserveVerification4CancelState())
	data, err := s.dao.RawGetUpActReserveRelation(ctx, req.Mid, []int64{int64(req.Type)}, state)
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawGetUpActReserveRelation")
		return
	}
	if len(data) == 0 {
		return
	}

	for _, v := range data {
		sTime := v.LivePlanStartTime.Time().Unix() - s.c.UpActReserveCreateConfig.PlayRange
		eTime := v.LivePlanStartTime.Time().Unix() + s.c.UpActReserveCreateConfig.PlayRange
		if req.Time >= sTime && req.Time <= eTime {
			var list []*pb.UpActReserveRelationInfo
			list, err = s.GetUpActReserveRelationReachInfo(ctx, []int64{v.Sid}, req.Mid)
			if err != nil {
				err = errors.Wrap(err, "s.GetUpActReserveRelationReachInfo err")
				return
			}
			if len(list) > 0 {
				reply.List[v.Sid] = list[0]
			}
		}
	}

	return
}

// 条件查询预约数据
func (s *Service) UpActReserveRelationDBInfoByCondition(ctx context.Context, req *pb.UpActReserveRelationDBInfoByConditionReq) (reply *pb.UpActReserveRelationDBInfoByConditionReply, err error) {
	var (
		states       []int64
		reserveTypes []int64
	)

	reply = &pb.UpActReserveRelationDBInfoByConditionReply{
		List: make(map[int64]*pb.UpActReserveRelationInfo),
	}
	if req.From == pb.UpVerifyReserveFrom_FromLiveVerify {
		states = append(states, int64(pb.UpActReserveRelationState_UpReserveRelated))
		reserveTypes = append(reserveTypes, int64(pb.UpActReserveRelationType_Live))
	}
	if len(states) == 0 || len(reserveTypes) == 0 {
		log.Infoc(ctx, "unsupported from type:%v", req.From)
		return
	}
	data, err := s.dao.RawGetUpActReserveRelationOfAllMid(ctx, reserveTypes, states)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.RawGetUpActReserveRelation error(%+v) req(%+v)", err, req)
		log.Errorc(ctx, err.Error())
		return
	}

	for _, v := range data {
		reply.List[v.Sid] = &pb.UpActReserveRelationInfo{
			Sid:               v.Sid,
			Upmid:             v.Mid,
			LivePlanStartTime: v.LivePlanStartTime,
		}
	}

	return
}

// 直播数据过期
func (s *Service) UpActReserveLiveStateExpire(ctx context.Context, req *pb.UpActReserveLiveStateExpireReq) (reply *pb.UpActReserveLiveStateExpireReply, err error) {
	reply = new(pb.UpActReserveLiveStateExpireReply)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	state := s.ConvertPBState2Int64(s.dao.LiveUpActReserveStateExpire())
	relations, err := s.dao.RawGetUpActReserveLiveExpireData(ctx, int64(pb.UpActReserveRelationType_Live), state)
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawGetUpActReserveLiveExpireData err")
		return
	}
	if len(relations) == 0 {
		return
	}

	for _, relation := range relations {
		if function.Now() > relation.LivePlanStartTime.Time().Unix()+s.c.UpActReserveCreateConfig.PlayRange+s.c.UpActReserveCreateConfig.PLayContinue {
			update := &like.UpActReserveRelationUpdateFields{
				Sid:               relation.Sid,
				Mid:               relation.Mid,
				SubjectState:      like.ActSubjectStateCancel,
				RelationState:     int64(pb.UpActReserveRelationState_UpReserveCancelExpired),
				AuditState:        relation.Audit,
				AuditChannelState: relation.AuditChannel,
				DynamicID:         relation.DynamicID,
			}
			if err = s.dao.TXUpdateSubjectAndRelationData(ctx, update); err != nil {
				err = errors.Wrapf(err, "s.dao.TXUpdateSubjectAndRelationData err update(%+v)", update)
				log.Errorc(ctx, err.Error())
			}
		}
	}

	return
}

// 只获取直播类型数据
func (s *Service) UpActReserveRelationInfo4Live(ctx context.Context, req *pb.UpActReserveRelationInfo4LiveReq) (reply *pb.UpActReserveRelationInfo4LiveReply, err error) {
	reply = new(pb.UpActReserveRelationInfo4LiveReply)
	reply.List = make([]*pb.UpActReserveRelationInfo, 0)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	var sid int64
	if req.Upmid == 0 {
		err = fmt.Errorf("upmid empty req(%+v)", req)
		return
	}

	// 内存获取数据
	cacheSid, err := s.dao.UpActReserveRelationInfo4LiveGCache.Get(req.Upmid)
	// 有错误
	if err != nil && err != gcache.KeyNotFoundError {
		err = errors.Wrap(err, "s.dao.UpActReserveRelationInfo4LiveGCache err")
		return
	}
	// 值不存在
	if err == gcache.KeyNotFoundError {
		// 回源数据
		sid, err = s.GetAndSetCacheKey(ctx, req.Upmid)
		if err != nil {
			err = errors.Wrap(err, "s.GetOrSetCacheKey err")
			return
		}
	} else {
		// 值存在 赋值
		sid = cacheSid.(int64)
	}

	if sid <= 0 {
		return
	}

	// 通过方法来获取sid数据
	relations, err := s.GetUpActReserveRelationReachInfo(ctx, []int64{sid}, req.Mid)
	if err != nil {
		err = errors.Wrap(err, "s.GetUpActReserveRelationReachInfo err")
		return
	}

	for _, relation := range relations {
		if relation.Sid == sid {
			reply.List = append(reply.List, relation)
		}
	}

	return
}

func (s *Service) ConvertPBState2Int64(pbState []pb.UpActReserveRelationState) []int64 {
	state := make([]int64, 0)
	for _, v := range pbState {
		state = append(state, int64(v))
	}
	return state
}

func (s *Service) GetAndSetCacheKey(ctx context.Context, upMid int64) (sid int64, err error) {
	// 回源数据
	sid, err = s.dao.GetUpActReserveRelationInfo4Live(ctx, upMid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.GetUpActReserveRelationInfo4Live")
		return
	}
	// 如果数据不存在
	if sid <= 0 {
		sid = 0
	}
	// 缓存到内存
	if err = s.dao.UpActReserveRelationInfo4LiveGCache.SetWithExpire(upMid, sid, time.Second*30); err != nil {
		err = errors.Wrapf(err, "s.dao.UpActReserveRelationInfo4LiveGCache.SetWithExpire err upmid(%+v) sid(%+v)", upMid, sid)
		return
	}
	return
}

func (s *Service) GetSidAndDynamicIDByOid(ctx context.Context, req *pb.GetSidAndDynamicIDByOidReq) (reply *pb.GetSidAndDynamicIDByOidReply, err error) {
	reply = new(pb.GetSidAndDynamicIDByOidReply)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	if req.Oid == "" {
		err = fmt.Errorf("req.Oid empty req(%+v)", req)
		return
	}

	// 动态来获取数据
	if req.From == pb.UpCreateActReserveFrom_FromDynamic {
		res := new(like.UpActReserveRelationBind)
		if res, err = s.dao.GetUpActReserveRelationBindInfo(ctx, req.Oid, int64(req.Type), int64(req.From)); err != nil {
			err = errors.Wrap(err, "s.dao.GetUpActReserveRelationBindInfo err")
			return
		}
		if res != nil && res.ID == 0 {
			return
		}

		reply.Sid = res.Sid
		reply.Oid = res.Oid
		reply.Rid = res.Rid

		return
	}

	return
}

// 特殊时间点开启强制审核 驳回不送审 其他均为先审后发 渠道为0的话 必审
func (s *Service) IsInAuditSpecialPeriod(ctx context.Context, from int64, ts int64) bool {
	// 默认全渠道审
	if from == like.SpecialPeriodMustAuditFrom {
		if ts >= s.c.UpActReserveCreateConfig.ForceAuditStime && ts <= s.c.UpActReserveCreateConfig.ForceAuditEtime {
			return true
		}
	} else {
		if _, ok := s.c.UpActReserveCreateConfig.ForceAuditFrom[strconv.FormatInt(from, 10)]; ok {
			if ts >= s.c.UpActReserveCreateConfig.ForceAuditStime && ts <= s.c.UpActReserveCreateConfig.ForceAuditEtime {
				return true
			}
		}
	}
	return false
}

func (s *Service) CanUpActReserve4Dynamic(ctx context.Context, req *pb.CanUpActReserve4DynamicReq) (res *pb.CanUpActReserve4DynamicReply, err error) {
	res = &pb.CanUpActReserve4DynamicReply{
		List: make(map[int64]*pb.PrivilegeMap),
	}

	// 发起权限
	canCreateReserveAuth, err := s.dao.CanCreateUpActReserve(ctx, req.Mid, req.From)
	if err != nil {
		err = errors.Wrap(err, "s.dao.CanCreateUpActReserve error")
		log.Errorc(ctx, err.Error())
		return
	}

	// 关联权限
	canRelateReserveAuth, err := s.dao.CanRelateUpActReserve(ctx, req.Mid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.CanRelateUpActReserve error")
		log.Errorc(ctx, err.Error())
		return
	}

	// 赋值
	for _, relationType := range []pb.UpActReserveRelationType{pb.UpActReserveRelationType_Archive, pb.UpActReserveRelationType_Live} {
		// 稿件or直播 的[发起]或[关联]权限
		privilege := &pb.PrivilegeMap{
			List: make(map[int64]pb.UpCreateActReserveQualification),
		}
		// 有发起权限 将发起结果的枚举填充到数据
		if v, ok := canCreateReserveAuth[int64(relationType)]; ok {
			privilege.List[int64(pb.PrivilegeType_CreateReserve)] = v
		}
		// 有关联权限 一定是allow枚举 将结果填充到数据
		if v, ok := canRelateReserveAuth[int64(relationType)]; ok {
			privilege.List[int64(pb.PrivilegeType_RelateReserve)] = v
		}
		// 有发起或者关联 即为有这种类型的权限 否则不包含权限
		if len(privilege.List) > 0 {
			res.List[int64(relationType)] = privilege
		}
	}

	return
}

func (s *Service) UpActReserveRecord(ctx context.Context, req *pb.UpActReserveRecordReq) (reply *pb.UpActReserveRecordReply, err error) {
	maxNumLimit := int64(1)
	reply = new(pb.UpActReserveRecordReply)

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	record, err := s.dao.IsUpActRelationReservePublished(ctx, req.Mid, []int64{int64(req.Type)}, maxNumLimit)
	if err != nil {
		err = errors.Wrap(err, "s.dao.IsUpActRelationReservePublished err")
		return
	}
	if record != nil && record.ID == 0 {
		reply.Res = false
		return
	}
	reply.Res = true
	return
}

func (s *Service) UpActReserveRelationDependAudit(ctx context.Context, req *pb.UpActReserveRelationDependAuditReq) (reply *pb.UpActReserveRelationDependAuditReply, err error) {
	reply = new(pb.UpActReserveRelationDependAuditReply)

	relations, err := s.dao.RawGetUpActReserveRelationInfoBySid(ctx, []int64{req.Sid})
	if err != nil {
		err = errors.Wrap(err, "s.dao.RawGetUpActReserveRelationInfoBySid err")
		return
	}

	if _, ok := relations[req.Sid]; !ok {
		err = fmt.Errorf("relations[req.Sid] empty relations(%+v) sid(%+v)", relations, req.Sid)
		return
	}

	relation := relations[req.Sid]
	switch req.Channel {
	case pb.UpActReserveRelationDependAuditChannel_DependAuditChannelDynamic: // 动态
		// 幂等
		if relation.DynamicAudit == int64(req.Audit) {
			return
		}
		// 更新审核状态
		if err = s.dao.UpdateUpActReserveRelationDependAuditState(ctx, relation.Sid, int64(req.Audit), relation.LotteryAudit); err != nil {
			err = errors.Wrap(err, "s.dao.UpdateUpActReserveRelationDependAuditState err")
			return
		}
	default:
		err = fmt.Errorf("req type illegal req(%+v)", req)
	}

	return
}

func (s *Service) CanUpActReserveByType(ctx context.Context, req *pb.CanUpActReserveByTypeReq) (reply *pb.CanUpActReserveByTypeReply, err error) {
	reply = new(pb.CanUpActReserveByTypeReply)
	// 能否创建预约
	data, err := s.dao.CanUpActReserveSpecific(ctx, req.Mid, req.Type)
	if err != nil {
		err = errors.Wrapf(err, "s.CanUpActReserveSpecific error, req:%+v", req)
		log.Errorc(ctx, err.Error())
		return
	}
	var crtExist bool
	if allow, ok := data[int64(req.Type)]; ok {
		crtExist = allow == pb.UpCreateActReserveQualification_QualificationAllow
	}
	arg := &like.UpActReserveRelationContinuingArg{
		Type: int64(req.Type),
	}
	// 是否有进行中的预约
	continuingList, err := s.UpActReserveRelationContinuing(ctx, req.Mid, arg)
	if err != nil {
		err = errors.Wrapf(err, "s.UpActReserveRelationContinuing error, req:%+v", req)
		log.Errorc(ctx, err.Error())
		return
	}
	continuingExist := len(continuingList) > 0

	reply = &pb.CanUpActReserveByTypeReply{
		Exist: crtExist || continuingExist,
		UpActReserveCreateInfo: &pb.UpActReserveCreateInfo{
			Exist: crtExist,
		},
		UpActReserveContinuingInfo: &pb.UpActReserveContinuingInfo{
			Exist: continuingExist,
			//ContinuingList: continuingList,
		},
	}
	return
}

func (s *Service) CanUpActReserveFull(ctx context.Context, req *pb.CanUpActReserveFullReq) (reply *pb.CanUpActReserveFullReply, err error) {
	reply = new(pb.CanUpActReserveFullReply)
	reply.Res = make(map[int64]*pb.CanUpActReserveFullInfo)
	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()
	for _, reserveType := range []pb.UpActReserveRelationType{pb.UpActReserveRelationType_Archive, pb.UpActReserveRelationType_Live} {
		// 能否创建预约
		var data map[int64]pb.UpCreateActReserveQualification
		data, err = s.dao.CanUpActReserveSpecific(ctx, req.Mid, reserveType)
		if err != nil {
			err = errors.Wrapf(err, "s.CanUpActReserveSpecific error, req:%+v", req)
			return
		}
		var crtExist bool
		if allow, ok := data[int64(reserveType)]; ok {
			crtExist = allow == pb.UpCreateActReserveQualification_QualificationAllow
		}

		// 是否有进行中的预约
		arg := &like.UpActReserveRelationContinuingArg{
			Type: int64(reserveType),
		}
		var continuingList []*pb.UpActReserveRelationInfo
		continuingList, err = s.UpActReserveRelationContinuing(ctx, req.Mid, arg)
		if err != nil {
			err = errors.Wrapf(err, "s.UpActReserveRelationContinuing error, req:%+v", req)
			return
		}
		continuingExist := len(continuingList) > 0

		// 是否有可关联的他人的预约
		arg2 := &like.UpActReserveRelationOthersArg{
			Type: int64(reserveType),
		}
		var othersList []*pb.UpActReserveRelationInfo
		othersList, err = s.UpActReserveRelationOthers(ctx, req.Mid, arg2)
		if err != nil {
			err = errors.Wrapf(err, "s.UpActReserveRelationOthers error, req:%+v", req)
			return
		}
		othersExist := len(othersList) > 0

		tmp := &pb.CanUpActReserveFullInfo{
			Exist: crtExist || continuingExist || othersExist,
			UpActReserveCreateInfo: &pb.UpActReserveCreateInfo{
				Exist: crtExist,
			},
			UpActReserveContinuingInfo: &pb.UpActReserveContinuingInfo{
				Exist: continuingExist,
				//ContinuingList: continuingList,
			},
			UpActReserveRelateOthersInfo: &pb.UpActReserveRelateOthersInfo{
				Exist: othersExist,
				//OthersReserveList: othersList,
			},
		}
		reply.Res[int64(reserveType)] = tmp
	}
	return
}

func (s *Service) CanUpRelateOthersActReserve(ctx context.Context, req *pb.CanUpRelateOthersActReserveReq) (reply *pb.CanUpRelateOthersActReserveReply, err error) {
	reply = new(pb.CanUpRelateOthersActReserveReply)
	res, err := s.dao.CanUpActOthersReserve(ctx, req.Mid, req.Sid)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.CanUpActOthersReserve err")
		log.Errorc(ctx, "s.dao.CanUpActOthersReserve err:(%v)", err)
		return
	}
	reply.Auth = res > 0
	return
}

func (s *Service) CanUpRelateReserveAuth(ctx context.Context, req *pb.CanUpRelateReserveAuthReq) (reply *pb.CanUpRelateReserveAuthReply, err error) {
	reply = new(pb.CanUpRelateReserveAuthReply)
	state := s.ConvertPBState2Int64(s.dao.CanCreateUpActReserveRelationState())

	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	res, err := s.dao.GetUpActReserveRelationInfoBySid(ctx, []int64{req.Sid})
	if err != nil {
		err = errors.Wrapf(err, "s.dao.CanUpRelateReserveAuth err")
		return
	}
	if res == nil {
		err = errors.Wrapf(err, "s.dao.GetUpActReserveRelationInfoBySid got nil res")
		return
	}
	relation, ok := res[req.Sid]
	if !ok {
		err = errors.Wrapf(err, "s.dao.GetUpActReserveRelationInfoBySid no relation info")
		return
	}

	if relation.Mid == req.Mid {
		reply.Role = pb.ReserveRelationRole_OwnReserve
		reply.Auth = function.InInt64Slice(relation.State, state)
		return
	}

	res2, err := s.dao.CanUpActOthersReserve(ctx, req.Mid, req.Sid)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.CanUpRelateReserveAuth err")
		return
	}
	reply.Role = pb.ReserveRelationRole_OthersReserve
	reply.Auth = res2 > 0 && s.dao.UpActReserveIsAudit(relation) == pb.UpActVisible_DefaultVisible && function.InInt64Slice(relation.State, state)
	return
}

func (s *Service) CheckCronPubReserveParams(ctx context.Context, req *like.CreateUpActReserveArgs) (err error) {
	if req.Oid == "" {
		err = fmt.Errorf("req.Oid empty req(%+v)", req)
		return
	}
	if req.LivePlanStartTime == 0 {
		err = fmt.Errorf("req.LivePlanStartTime == 0 req(%+v)", req)
		return
	}
	return
}

func (s *Service) HandleTitle(ctx context.Context, title string, typ pb.UpActReserveRelationType) (convTitle string) {
	convTitle = title
	// 稿件类型标题加前缀
	if typ == pb.UpActReserveRelationType_Archive {
		convTitle = "预告：" + convTitle
	}
	return
}
