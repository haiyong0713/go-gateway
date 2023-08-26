package dynamicV2

import (
	"context"
	"math"
	"math/rand"
	"time"

	"go-common/library/conf/env"
	"go-common/library/exp/ab"
	"go-common/library/log"
	v2 "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	"go-common/library/log/infoc.v2"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

func (s *Service) unloginAbtest(c context.Context, general *mdlv2.GeneralParam, uid int64, abtestName, abValue string, flag *ab.StringFlag) bool {
	var (
		groupID int64
	)
	t, ok := ab.FromContext(c)
	if !ok {
		return false
	}
	t.Add(ab.KVString("buvid", general.GetBuvid()))
	exp := flag.Value(t)
	for _, state := range t.Snapshot() {
		if state.Type == ab.ExpHit {
			groupID = state.Value
			break
		}
	}
	event := infoc.NewLogStreamV(s.c.Infoc.DynAbLogID,
		log.Int64(uid),            // 用户uid
		log.String(abtestName),    // 实验变量名
		log.String(exp),           // 实验值
		log.Int64(groupID),        // 实验分组; 未命中是0
		log.String(env.DeployEnv), // 环境
		log.String(time.Now().Format("2006-01-02 15:04:05"))) // 时间
	log.String(general.GetBuvid()) // Buvid
	if err := s.infocV2.Info(c, event); err != nil {
		log.Error("[Lancer] unloginAbtest report failed for log_id:%s, err:%+v", s.c.Infoc.DynAbLogID, err)
	}
	if exp != abValue {
		return false
	}
	return true
}

//nolint:unused
func (s *Service) tabAbtest(c context.Context, acc *accountgrpc.Profile, abtestName, abValue string, flag *ab.StringFlag) bool {
	if acc == nil {
		return false
	}
	// 当前用户的注册时间小于实验时间则跳出实验
	if int64(acc.JoinTime) < s.c.Resource.TabAbtestUserTime {
		return false
	}
	var (
		groupID int64
	)
	t, ok := ab.FromContext(c)
	if !ok {
		return false
	}
	t.Add(ab.KVInt("mid", acc.Mid))
	exp := flag.Value(t)
	if exp == _abMiss {
		return false
	}
	for _, state := range t.Snapshot() {
		if state.Type == ab.ExpHit {
			groupID = state.Value
			break
		}
	}
	event := infoc.NewLogStreamV(s.c.Infoc.TabAbLogID,
		log.Int64(acc.Mid),        // 用户uid
		log.String(abtestName),    // 实验变量名
		log.String(exp),           // 实验值
		log.Int64(groupID),        // 实验分组; 未命中是0
		log.String(env.DeployEnv), // 环境
		log.String(time.Now().Format("2006-01-02 15:04:05"))) // 时间
	if err := s.infocV2.Info(c, event); err != nil {
		log.Error("[Lancer] HotTest report failed for log_id:%s, err:%+v", s.c.Infoc.TabAbLogID, err)
	}
	if exp != abValue {
		return false
	}
	return true
}

// 动态综合页的筛选器AB实验
func (s *Service) dynAllFilterAbtest(ctx context.Context, general *mdlv2.GeneralParam, req *v2.DynTabReq, resp *v2.DynTabReply) {
	// 测试环境没有数据，直接全量
	if env.DeployEnv == env.DeployEnvUat {
		resp.ScreenTab = s.tabFilters(ctx, general, req, true)
		return
	}

	const (
		_dynFilterIOS     = 67700000
		_dynFilterAndroid = 6770000
	)
	// 版本硬限 支持的版本才开实验
	if !general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, _dynFilterAndroid, _dynFilterIOS) {
		return
	}

	const (
		_abTestFilter   = "vdtab"    // 实验开关，必须存在的情况下才开启实验
		_abVdTabOldUser = "vdtabold" // 老用户人群包
		_abVdTabOld2New = "vdtabo2n" // 老转新人群包
	)
	// 实验没开 不做修改
	if !s.dynDao.IsAbTestSwitchOn(ctx, _abTestFilter) {
		return
	}
	var (
		isVdTabOld bool
	)
	// 先看老转新，然后看老用户，两个都不命中那就是新用户
	isVdTabOld2New, err := s.dynDao.IsAbTestMidHit(ctx, _abVdTabOld2New, general.Mid)
	if err != nil {
		log.Errorc(ctx, "skip abTest(%s) due to cache failure: %v", _abTestFilter, err)
		return
	}
	// 如果不是老转新 看一下是不是老用户
	if !isVdTabOld2New {
		isVdTabOld, err = s.dynDao.IsAbTestMidHit(ctx, _abVdTabOldUser, general.Mid)
		if err != nil {
			log.Errorc(ctx, "skip abTest(%s) due to cache failure: %v", _abTestFilter, err)
			return
		}
	} else {
		// 是老转新的情况下 确保不是老用户
		isVdTabOld = false
	}

	t, ok := ab.FromContext(ctx)
	if !ok {
		return
	}
	t.Add(ab.KVInt("mid", general.Mid), ab.KVString("buvid", general.GetBuvid()))

	var (
		abVal string
	)
	if isVdTabOld2New {
		abVal = model.DynVideoTabOld2New.Value(t)
	} else if isVdTabOld {
		abVal = model.DynVideoTabOld.Value(t)
	} else {
		abVal = model.DynVideoTabPureNew.Value(t)
	}

	switch abVal {
	// 1 是对照组 不改效果
	case "1":
		return
	case "2", "3":
		// no op
	default:
		// 未知实验结果 直接跳出
		return
	}
	// 实验组操作
	abTestAction := map[string]func(resp *v2.DynTabReply){
		"1": func(resp *v2.DynTabReply) {
			// 对照组 noop
		},
		"2": func(resp *v2.DynTabReply) {
			// 实验组
			// 展示筛选器，筛选器含视频按钮，顶部隐藏视频tab
			resp.ScreenTab = s.tabFilters(ctx, general, req, true)
			removeVideoTab(resp)
		},
		"3": func(resp *v2.DynTabReply) {
			// 实验组
			// 展示筛选器，筛选器无视频按钮，顶部不隐藏视频tab
			resp.ScreenTab = s.tabFilters(ctx, general, req, false)
		},
	}

	// apply实验效果
	abTestAction[abVal](resp)
}

func removeVideoTab(resp *v2.DynTabReply) {
	if resp == nil {
		return
	}
	res := make([]*v2.DynTab, 0, len(resp.DynTab))
	for i := range resp.DynTab {
		if resp.DynTab[i] != tabVideo {
			res = append(res, resp.DynTab[i])
		}
	}
	resp.DynTab = res
}

func (s *Service) dynDetailBottomBar(ctx context.Context, general *mdlv2.GeneralParam) bool {
	if env.DeployEnv == env.DeployEnvUat {
		return true
	}
	const (
		_dynDetailBarIOS     = 68200000
		_dynDetailBarAndroid = 6820000
	)
	// 版本硬限 支持的版本才开实验
	if !general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, _dynDetailBarAndroid, _dynDetailBarIOS) {
		return false
	}

	t, ok := ab.FromContext(ctx)
	if !ok {
		return false
	}
	t.Add(ab.KVString("buvid", general.GetBuvid()), ab.KVInt("mid", general.Mid))

	val := model.DynDetailBar.Value(t)

	return val == "1"
}

const (
	_expTopicCtrl = iota // 对照组
	_expTopicA           // 话题广场样式A（翻页点点在左上角）
	_expTopicB           // 话题广场样式B（翻页点点在最下面正中间）
)

var dynAllTopicSquareStyleMap = map[string]int32{
	"1": _expTopicCtrl,
	"2": _expTopicA,
	"3": _expTopicB,
}

func (s *Service) dynAllTopicSquareStyle(ctx context.Context, general *mdlv2.GeneralParam) int32 {
	if env.DeployEnv == env.DeployEnvUat {
		if rand.Int() <= math.MaxInt32 {
			return _expTopicA
		}
		return _expTopicB
	}
	const (
		_dynAllTopicIOS     = 68300000
		_dynAllTopicIOSHD   = 34700000
		_dynAllTopicAndroid = 6830000
	)
	// 版本硬限 支持的版本才开实验
	if !general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, _dynAllTopicAndroid, _dynAllTopicIOS) &&
		!(general.IsPad() && general.GetBuild() >= _dynAllTopicIOS) &&
		!(general.IsPadHD() && general.GetBuild() >= _dynAllTopicIOSHD) {
		return _expTopicCtrl
	}

	t, ok := ab.FromContext(ctx)
	if !ok {
		return _expTopicCtrl
	}
	t.Add(ab.KVInt("mid", general.Mid), ab.KVString("buvid", general.GetBuvid()))

	val := model.DynAllTopicSquare.Value(t)

	return dynAllTopicSquareStyleMap[val]
}
