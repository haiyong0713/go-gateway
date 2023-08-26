package service

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math"
	"math/rand"
	"net/url"
	"sort"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/archive/middleware"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/model"
	"go-gateway/app/app-svr/archive/service/model/archive"
	ugcmdl "go-gateway/app/app-svr/ugc-season/service/api"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	vasApi "git.bilibili.co/bapis/bapis-go/vas/trans/service"
	batch "git.bilibili.co/bapis/bapis-go/video/vod/playurlugcbatch"
	volume "git.bilibili.co/bapis/bapis-go/video/vod/playurlvolume"

	"github.com/thoas/go-funk"
)

const (
	_ispCU          = "联通"
	_ispCT          = "电信"
	_ispCM          = "移动"
	_MaxRedirectCnt = 20
)

// ArcWithStat is
func (s *Service) ArcWithStat(c context.Context, arg *api.ArcRequest) (*api.Arc, error) {
	var (
		arc  *api.Arc
		stat *api.Stat
	)
	g := errgroup.WithContext(c)
	g.Go(func(ctx context.Context) (err error) {
		if arc, err = s.arc.Arc(ctx, arg.Aid); err != nil {
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) (err error) {
		if stat, err = s.arc.Stat3(ctx, arg.Aid); err != nil {
			log.Error("Arc aid(%d) error(%+v)", arg.Aid, err)
			stat = &api.Stat{Aid: arg.Aid}
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		log.Error("Arc aid(%d) error(%+v)", arg.Aid, err)
		return nil, err
	}
	arc.Stat = *stat
	arc.FillStat()
	s.checkArcPremiereInfo(c, arc)
	s.checkArcIpControl(arc)
	arcMap := make(map[int64]*api.Arc)
	arcMap[arc.Aid] = arc
	s.payCheck(c, arg.Mid, arcMap)
	return arc, nil
}

func (s *Service) MultiPlayerBk(c context.Context, arg *api.ArcsWithPlayurlRequest) (map[int64]*api.ArcPlayer, error) {
	aps, err := s.ArcsWithPlayurl(c, arg)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]*api.ArcPlayer)
	for aid, ap := range aps {
		if ap == nil {
			continue
		}
		playerInfo := make(map[int64]*api.PlayerInfo)
		var pgcExtra *api.PGCPlayerExtra
		var progress int64
		if ap.EpisodeId > 0 {
			pgcExtra = &api.PGCPlayerExtra{
				IsPreview:   ap.IsPreview,
				EpisodeId:   ap.EpisodeId,
				SubType:     ap.SubType,
				PgcSeasonId: ap.PgcSeasonId,
			}
		}
		if ap.GetHistory().GetCid() == ap.Arc.FirstCid {
			progress = ap.GetHistory().GetProgress()
		}
		playerInfo[ap.FirstCid] = &api.PlayerInfo{
			Playurl: ap.Playurl,
			PlayerExtra: &api.PlayerExtra{
				Dimension:      &ap.Dimension,
				PgcPlayerExtra: pgcExtra,
				Progress:       progress,
				Cid:            ap.FirstCid,
			},
		}
		res[aid] = &api.ArcPlayer{
			Arc:              ap.Arc,
			PlayerInfo:       playerInfo,
			DefaultPlayerCid: ap.FirstCid,
		}
	}
	return res, nil
}

// nolint:gocognit
func (s *Service) ArcsPlayerSvr(c context.Context, arg *api.ArcsPlayerRequest) (map[int64]*api.ArcPlayer, error) {
	if arg == nil {
		return nil, ecode.RequestErr
	}
	aids, noPlayAids, aToC, aidHighQn, err := arg.Parse()
	if err != nil {
		return nil, err
	}
	//开关灾备
	if s.c.Switch.NoMultiPlayer {
		return s.MultiPlayerBk(c, &api.ArcsWithPlayurlRequest{Aids: aids, BatchPlayArg: arg.BatchPlayArg, AidsWithoutPlayurl: noPlayAids})
	}
	//秒开参数校验
	if arg.BatchPlayArg == nil {
		arg.BatchPlayArg = &api.BatchPlayArg{}
	}
	as, history, isVip, isControl, err := s.baseForBatchAv(c, aids, arg.BatchPlayArg.Mid, arg.BatchPlayArg.MobiApp, arg.BatchPlayArg.Device, arg.BatchPlayArg.Buvid, arg.BatchPlayArg.AutoplayAreaValidate)
	if err != nil {
		return nil, err
	}
	if len(as) == 0 {
		return nil, ecode.NothingFound
	}
	ap := make(map[int64]*api.ArcPlayer, len(aids))
	var paids []int64
	var spAids = make(map[int64]bool)
	var noPlayAidsMap = make(map[int64]struct{})
	for aid, arc := range as {
		ap[aid] = new(api.ArcPlayer)
		ap[aid].Arc = arc
		ap[aid].DefaultPlayerCid = arc.FirstCid
		ap[aid].PlayerInfo = make(map[int64]*api.PlayerInfo)
		ap[aid].PlayerInfo[arc.FirstCid] = &api.PlayerInfo{
			PlayerExtra: &api.PlayerExtra{
				Cid:       arc.FirstCid,
				Dimension: &arc.Dimension,
			},
		}
		//付费稿件，autoplay=0禁止自动播放，且不获取秒开地址
		if arc.Duration > s.c.Custom.DurationLimit || !arc.IsNormal() || funk.ContainsInt64(noPlayAids, aid) ||
			arc.AttrValV2(api.AttrBitV2Pay) == api.AttrYes {
			noPlayAidsMap[aid] = struct{}{}
			delete(aToC, aid)
			continue
		}
		if arg.BatchPlayArg.ShowPgcPlayurl && arc.AttrVal(archive.AttrBitIsPGC) == archive.AttrYes {
			paids = append(paids, aid)
		}
		if arc.Rights.Autoplay != 1 {
			noPlayAidsMap[aid] = struct{}{}
			continue
		}
		if (isVip && !isControl) || arc.Author.Mid == arg.BatchPlayArg.Mid || s.isVipFree(aid) {
			spAids[aid] = true
		}
		hisCid := s.historyDefault(arg.BatchPlayArg.MobiApp, arg.BatchPlayArg.Build, aid, history)
		if arc.Videos > 1 && hisCid > 0 {
			aToC[aid] = append(aToC[aid], hisCid)
		}
		//保证有首p
		if len(aToC[aid]) == 0 {
			aToC[aid] = append(aToC[aid], arc.FirstCid)
		}
		aToC[aid] = funk.UniqInt64(aToC[aid])
	}
	if !s.c.PlayerSwitch || !arg.BatchPlayArg.AllowRequestBvc() {
		return ap, nil
	}
	// 过滤出真实存在的aidToCids的map，获取aid下有效的cids
	vs, err := s.arc.VideosByAidCids(c, aToC)
	if err != nil {
		//返回error，但是不阻塞秒开返回，默认下面逻辑会返回首p秒开数据
		log.Error("s.arc.VideosByAidCids is err aToC(%+v) %+v", aToC, err)
	}
	//过滤稿件下有效的cids
	for aid, pages := range vs {
		if _, ok := aToC[aid]; !ok {
			delete(aToC, aid)
			continue
		}
		var checkedCids []int64
		for _, p := range pages {
			checkedCids = append(checkedCids, p.Cid)
		}
		aToC[aid] = checkedCids
	}
	//获取默认DefaultPlayerCid
	for aid, a := range ap {
		hisCid := s.historyDefault(arg.BatchPlayArg.MobiApp, arg.BatchPlayArg.Build, aid, history)
		if hisCid > 0 && funk.ContainsInt64(aToC[aid], hisCid) {
			ap[aid].DefaultPlayerCid = hisCid
			continue
		}
		if _, ok := noPlayAidsMap[aid]; ok {
			continue
		}
		if !funk.ContainsInt64(aToC[aid], a.Arc.FirstCid) {
			aToC[aid] = append(aToC[aid], a.Arc.FirstCid)
		}
	}
	var cidArr []*batch.RequestVideoItem
	for aid, cids := range aToC {
		for _, cid := range cids {
			cidArr = append(cidArr, &batch.RequestVideoItem{Cid: uint64(cid), IsSp: spAids[aid]})
		}
	}
	if len(cidArr) == 0 && len(paids) == 0 {
		return ap, nil
	}

	return s.ArcsWithUrl(c, arg, cidArr, paids, ap, vs, history, aToC, aidHighQn), nil
}

func (s *Service) joinPCDNToPlayUrlInfo(bpcdnInfo map[string]string, item *batch.ResponseItem) {
	if len(bpcdnInfo) == 0 || item == nil || item.Dash == nil {
		return
	}
	for _, v := range item.Dash.Video {
		bpcdn, ok := bpcdnInfo[fmt.Sprintf("%d_%d", v.Id, v.Codecid)]
		if !ok {
			continue
		}
		newBaseUrl := v.BaseUrl
		if strings.Contains(newBaseUrl, "?") {
			newBaseUrl += "&"
		} else {
			newBaseUrl += "?"
		}
		params := url.Values{}
		params.Add("bpcdn", bpcdn)
		paramStr := params.Encode()
		// 重新encode的时候空格变成了+号问题修复
		if strings.IndexByte(paramStr, '+') > -1 {
			paramStr = strings.Replace(paramStr, "+", "%20", -1)
		}
		v.BaseUrl = fmt.Sprintf("%s%s", newBaseUrl, paramStr)
		s.infoProm.Incr("bpcdn-video-hit")
	}
	for _, a := range item.Dash.Audio {
		bpcdn, ok := bpcdnInfo[fmt.Sprintf("%d_%d", a.Id, a.Codecid)]
		if !ok {
			continue
		}
		newBaseUrl := a.BaseUrl
		if strings.Contains(newBaseUrl, "?") {
			newBaseUrl += "&"
		} else {
			newBaseUrl += "?"
		}
		params := url.Values{}
		params.Add("bpcdn", bpcdn)
		paramStr := params.Encode()
		// 重新encode的时候空格变成了+号问题修复
		if strings.IndexByte(paramStr, '+') > -1 {
			paramStr = strings.Replace(paramStr, "+", "%20", -1)
		}
		a.BaseUrl = fmt.Sprintf("%s%s", newBaseUrl, paramStr)
		s.infoProm.Incr("bpcdn-audio-hit")
	}
}

func (s *Service) historyDefault(mobiApp string, build, aid int64, history map[int64]*hisApi.ModelHistory) int64 {
	// 新版本才要选择返回历史记录，老版本仅返回首p
	if (mobiApp == archive.MobileAppIphone && build >= s.c.Custom.HistoryPlayUrlBuildIphone) || (mobiApp == archive.MobileAppAndroid && build >= s.c.Custom.HistoryPlayUrlBuildAndroid) {
		if his, ok := history[aid]; ok {
			return his.Cid
		}
	}
	return 0
}

func (s *Service) archiveAutoPlayValidate(ctx context.Context, arcs map[int64]*api.Arc, incArcs map[int64]*api.ArcInternal) error {
	aids := []int64{}
	//1、找出需要进行校验auto_play的aid
	for _, arc := range arcs {
		var inArc *api.ArcInternal
		if _, ok := incArcs[arc.Aid]; ok {
			inArc = incArcs[arc.Aid]
		}
		// 先跳过允许 autoplay 的稿件
		if api.CalcAutoplayV2(arc, inArc) == 1 {
			continue
		}
		// 不允许 autoplay，但当跳过 limit area 后，返回的是 1，说明只有地区受限
		if api.CalcAutoplayV2(arc, inArc, api.SkipLimitArea()) == 1 {
			//地区限制需要调用location接口进一步判断auto_play的值
			if arc.AttrVal(api.AttrBitLimitArea) == api.AttrYes {
				aids = append(aids, arc.Aid)
			}
		}
	}
	if len(aids) == 0 {
		return nil
	}
	//2、调用location接口
	res, err := s.locationDao.ArchiveAuthBatch(ctx, aids)
	if err != nil {
		log.Error("s.locationDao.ArchiveAuthBatch is err %+v", err)
		return err
	}
	if len(res) == 0 {
		return nil
	}
	//3、重新赋值auto_play
	for aid, auth := range res {
		if auth.Play == int64(locgrpc.Status_Forbidden) {
			continue
		}
		if _, ok := arcs[aid]; ok {
			arcs[aid].Rights.Autoplay = 1
		}
	}
	return nil
}

// ArcsWithPlayurl grpc with playurl
func (s *Service) ArcsWithPlayurl(c context.Context, arg *api.ArcsWithPlayurlRequest) (ap map[int64]*api.ArcWithPlayurl, err error) {
	if arg == nil || (len(arg.Aids) == 0 && len(arg.AidsWithoutPlayurl) == 0) {
		err = ecode.RequestErr
		return
	}
	aids := append(arg.Aids, arg.AidsWithoutPlayurl...)
	ap = make(map[int64]*api.ArcWithPlayurl, len(aids))
	arg.ResetBatchArg()
	as, history, isVip, isControl, err := s.baseForBatchAv(c, aids, arg.BatchPlayArg.Mid, arg.BatchPlayArg.MobiApp, arg.BatchPlayArg.Device, arg.BatchPlayArg.Buvid, arg.BatchPlayArg.AutoplayAreaValidate)
	if err != nil {
		return nil, err
	}
	if len(as) == 0 {
		return
	}
	for aid, arc := range as {
		ap[aid] = new(api.ArcWithPlayurl)
		ap[aid].Arc = arc
		if h, ok := history[aid]; ok {
			if h == nil {
				continue
			}
			progress := h.Pro
			if h.Pro > 0 { // 历史进度单位统一毫秒
				progressRatio := int64(1000)
				progress = h.Pro * progressRatio
			}
			ap[aid].History = &api.History{
				Cid:      h.Cid,
				Progress: progress,
			}
		}
	}
	if !s.c.PlayerSwitch || !arg.BatchPlayArg.AllowRequestBvc() {
		return
	}
	var (
		cidArr []*batch.RequestVideoItem
		paids  []int64
	)
	//需要playurl的aids继续执行（相关推荐只有前三位需要
	for _, aid := range arg.Aids {
		av, ok := ap[aid]
		if !ok {
			continue
		}
		if av.Duration > s.c.Custom.DurationLimit { //秒开文件长度太长客户端会奔溃
			continue
		}
		if arg.BatchPlayArg.ShowPgcPlayurl && av.AttrVal(archive.AttrBitIsPGC) == archive.AttrYes {
			paids = append(paids, av.Aid)
		}
		if av.Rights.Autoplay != 1 {
			continue
		}
		isSp := false
		// 1.稿件是自己的
		// 2.是vip&&没有被管控
		// 3.vip限免视频
		if (isVip && !isControl) || av.Author.Mid == arg.BatchPlayArg.Mid || s.isVipFree(aid) {
			isSp = true
		}
		cidArr = append(cidArr, &batch.RequestVideoItem{Cid: uint64(av.FirstCid), IsSp: isSp})
	}
	if len(cidArr) == 0 && len(paids) == 0 {
		return
	}
	if s.simplePlayurl(arg.BatchPlayArg.MobiApp, arg.BatchPlayArg.Build, arg.BatchPlayArg.Mid) {
		ap, err = s.ArcsWithSP(c, arg, cidArr, paids, ap)
	} else {
		ap, err = s.ArcsWithAP(c, arg, cidArr, paids, ap)
	}
	return
}

// ArcsWithAP is all playurl with all qn
func (s *Service) ArcsWithAP(c context.Context, arg *api.ArcsWithPlayurlRequest, cidArr []*batch.RequestVideoItem, paids []int64, ap map[int64]*api.ArcWithPlayurl) (res map[int64]*api.ArcWithPlayurl, err error) {
	pm, pgcm, _, _ := s.batchPlayURL(c, cidArr, paids, arg.BatchPlayArg)
	if len(pm) == 0 && len(pgcm) == 0 {
		return ap, nil
	}
	res = make(map[int64]*api.ArcWithPlayurl)
	for k, arc := range ap {
		pi, piok := pm[uint64(arc.FirstCid)]
		if piok {
			var askQn int64
			// story模式下 秒开使用720p
			if arg.BatchPlayArg.From == model.PlayurlFromStory {
				askQn = model.QnFlv720
			}
			arc.Playurl = new(api.BvcVideoItem)
			arc.Playurl.FromBatch(pi, askQn)
		}
		if pgci, ok := pgcm[arc.FirstCid]; ok {
			arc.IsPreview = pgci.IsPreview
			arc.EpisodeId = pgci.EpisodeId
			arc.Playurl = pgci.PlayerInfo
			arc.Rights.Autoplay = 1
			arc.SubType = pgci.SeasonType
			arc.PgcSeasonId = pgci.SeasonID
		}
		res[k] = arc
	}
	return
}

// ArcsWithSP is simple playurl with miaokai qn and user qn
func (s *Service) ArcsWithSP(c context.Context, arg *api.ArcsWithPlayurlRequest, cidArr []*batch.RequestVideoItem, paids []int64, ap map[int64]*api.ArcWithPlayurl) (res map[int64]*api.ArcWithPlayurl, err error) {
	ugcm, pgcm, _, _ := s.batchPlayURL(c, cidArr, paids, arg.BatchPlayArg)
	if len(ugcm) == 0 && len(pgcm) == 0 {
		return ap, nil
	}
	askQn := s.setAskQn(arg.BatchPlayArg)
	var formatsNew bool
	if (arg.BatchPlayArg.MobiApp == "iphone" && arg.BatchPlayArg.Build > s.c.Custom.HdrIOS) || (arg.BatchPlayArg.MobiApp == "android" && arg.BatchPlayArg.Build > s.c.Custom.HdrAnd) {
		formatsNew = true
	}
	res = make(map[int64]*api.ArcWithPlayurl)
	for k, arc := range ap {
		ugci, uok := ugcm[uint64(arc.FirstCid)]
		if uok {
			arc.Playurl = new(api.BvcVideoItem)
			arc.Playurl.FromBatchV2(ugci, arg.BatchPlayArg.Qn, askQn, formatsNew, s.isVipFree(arc.Aid), nil, "", 0, false)
		}
		if pgci, ok := pgcm[arc.FirstCid]; ok {
			arc.IsPreview = pgci.IsPreview
			arc.EpisodeId = pgci.EpisodeId
			arc.Playurl = pgci.PlayerInfo
			arc.Rights.Autoplay = 1
			arc.SubType = pgci.SeasonType
			arc.PgcSeasonId = pgci.SeasonID
		}
		res[k] = arc
	}
	return
}

func (s *Service) PgcParamsMerge(pgcm map[int64]*archive.PGCPlayurl, arcPlayer *api.ArcPlayer, history map[int64]*hisApi.ModelHistory, pages map[int64][]*api.Page) *api.ArcPlayer {
	arcPlayer.PlayerInfo = make(map[int64]*api.PlayerInfo)
	var tmpVideoItem *api.BvcVideoItem
	tmpExtra := &api.PlayerExtra{
		Cid:       arcPlayer.DefaultPlayerCid,
		Progress:  s.playerProgress(history, arcPlayer.Arc.Aid, arcPlayer.DefaultPlayerCid),
		Dimension: s.getDimension(arcPlayer, arcPlayer.DefaultPlayerCid, pages),
	}
	if pgci, ok := pgcm[arcPlayer.Arc.FirstCid]; ok {
		tmpExtra.PgcPlayerExtra = &api.PGCPlayerExtra{
			IsPreview:   pgci.IsPreview,
			EpisodeId:   pgci.EpisodeId,
			SubType:     pgci.SeasonType,
			PgcSeasonId: pgci.SeasonID,
		}
		tmpVideoItem = pgci.PlayerInfo
		arcPlayer.Arc.Rights.Autoplay = 1
	}
	arcPlayer.PlayerInfo[arcPlayer.DefaultPlayerCid] = &api.PlayerInfo{
		Playurl:     tmpVideoItem,
		PlayerExtra: tmpExtra,
	}
	return arcPlayer
}

func (s *Service) UgcParamsMerge(ugc map[uint64]*batch.ResponseItem, arcPlayer *api.ArcPlayer, history map[int64]*hisApi.ModelHistory, pages map[int64][]*api.Page, cids []int64, arg *api.ArcsPlayerRequest, cdnScore map[string]map[string]string, volMap map[uint64]*volume.VolumeItem, isHighQn bool) *api.ArcPlayer {
	arcPlayer.PlayerInfo = make(map[int64]*api.PlayerInfo)
	for _, cid := range cids {
		var tmpVideoItem *api.BvcVideoItem
		tmpExtra := &api.PlayerExtra{
			Cid:       cid,
			Progress:  s.playerProgress(history, arcPlayer.Arc.Aid, cid),
			Dimension: s.getDimension(arcPlayer, cid, pages),
		}
		if ugci, ok := ugc[uint64(cid)]; ok {
			tmpVideoItem = new(api.BvcVideoItem)
			askQn := s.setAskQn(arg.BatchPlayArg)
			//针对登陆用户：qn=32下发480P+720P
			if s.QnChangeGrey(arg.BatchPlayArg) && arg.BatchPlayArg.Mid > 0 && arg.BatchPlayArg.Qn == 32 && arg.BatchPlayArg.From != model.PlayurlFromStory {
				askQn = model.QnFlv720
			}
			if s.simplePlayurl(arg.BatchPlayArg.MobiApp, arg.BatchPlayArg.Build, arg.BatchPlayArg.Mid) {
				var formatsNew bool
				if (arg.BatchPlayArg.MobiApp == "iphone" && arg.BatchPlayArg.Build > s.c.Custom.HdrIOS) || (arg.BatchPlayArg.MobiApp == "android" && arg.BatchPlayArg.Build > s.c.Custom.HdrAnd) {
					formatsNew = true
				}
				tmpVideoItem.FromBatchV2(ugci, arg.BatchPlayArg.Qn, askQn, formatsNew, s.isVipFree(arcPlayer.Arc.Aid), cdnScore, arg.BatchPlayArg.From, arg.BatchPlayArg.NetType, isHighQn)
			} else {
				tmpVideoItem.FromBatch(ugci, askQn)
			}
			//登陆用户 + 用户请求qn=32 + 排除story 返回quality = 32(后期产品排期优化此逻辑)
			if s.QnChangeGrey(arg.BatchPlayArg) && arg.BatchPlayArg.Mid > 0 && arg.BatchPlayArg.Qn == 32 && arg.BatchPlayArg.From != model.PlayurlFromStory {
				tmpVideoItem.Quality = model.Qn480
			}
			//获取音量信息
			tmpVideoItem.Volume = s.getVolume(cid, volMap)
		}
		arcPlayer.PlayerInfo[cid] = &api.PlayerInfo{
			Playurl:     tmpVideoItem,
			PlayerExtra: tmpExtra,
		}
	}
	return arcPlayer
}

func (s *Service) getVolume(cid int64, volMap map[uint64]*volume.VolumeItem) *api.VolumeInfo {
	if vol, ok := volMap[uint64(cid)]; ok {
		return &api.VolumeInfo{
			MeasuredI:         vol.MeasuredI,
			MeasuredLra:       vol.MeasuredLra,
			MeasuredTp:        vol.MeasuredTp,
			MeasuredThreshold: vol.MeasuredThreshold,
			TargetOffset:      vol.TargetOffset,
			TargetI:           vol.TargetI,
			TargetTp:          vol.TargetTp,
		}
	}
	return nil
}

func (s *Service) QnChangeGrey(batchArg *api.BatchPlayArg) bool {
	miaoMod := crc32.ChecksumIEEE([]byte(batchArg.Buvid+"qn_change_grey")) % uint32(100)
	return miaoMod < s.c.Custom.QnChangeGrey
}

func (s *Service) getDimension(arcPlayer *api.ArcPlayer, cid int64, pages map[int64][]*api.Page) *api.Dimension {
	//dimension
	dimension := &api.Dimension{}
	if cid == arcPlayer.Arc.FirstCid { //单p的直接用arc.dimension赋值
		dimension = &arcPlayer.Arc.Dimension
	} else if tmpVs, ok := pages[arcPlayer.Arc.Aid]; ok { //多p的再pages里找
		// pages 改成 map[aid][cid]的形式
		tmp := funk.ToMap(tmpVs, "Cid")
		mapping := tmp.(map[int64]*api.Page)
		if p, ok := mapping[cid]; ok {
			dimension = &p.Dimension
		}
	}
	return dimension
}

func (s *Service) ArcsWithUrl(c context.Context, arg *api.ArcsPlayerRequest, cidArr []*batch.RequestVideoItem, paids []int64, ap map[int64]*api.ArcPlayer, pages map[int64][]*api.Page, history map[int64]*hisApi.ModelHistory, aToC map[int64][]int64, aidHighQn map[int64]bool) map[int64]*api.ArcPlayer {
	ugcm, pgcm, locInfo, volItem := s.batchPlayURL(c, cidArr, paids, arg.BatchPlayArg)
	if len(ugcm) == 0 && len(pgcm) == 0 {
		return ap
	}
	// third cdn choose
	cdnScore := s.calCdnScore(c, ugcm, arg.BatchPlayArg.Buvid, locInfo, arg.BatchPlayArg.Mid)
	for aid, arcPlayer := range ap {
		//pgc
		if funk.Contains(paids, aid) {
			s.PgcParamsMerge(pgcm, arcPlayer, history, pages)
			continue
		}
		//ugc
		cids, ok := aToC[aid]
		if !ok {
			continue
		}
		isHighQn := aidHighQn[aid] //是否需要高清晰度
		s.UgcParamsMerge(ugcm, arcPlayer, history, pages, cids, arg, cdnScore, volItem, isHighQn)
	}
	return ap
}

// nolint:gocognit
// ArchivesWithPlayer with player 老版本使用不迭代
func (s *Service) ArchivesWithPlayer(c context.Context, arg *archive.ArgPlayer, showPGCPlayurl bool) (ap map[int64]*archive.ArchiveWithPlayer, err error) {
	if arg == nil || (len(arg.Aids) == 0 && len(arg.AidsWithoutPlayurl) == 0) {
		err = ecode.RequestErr
		return
	}
	var (
		qn        = arg.Qn
		platform  = arg.Platform
		ip        = arg.RealIP
		fnver     = arg.Fnver
		fnval     = arg.Fnval
		session   = arg.Session
		forceHost = arg.ForceHost
		build     = arg.Build
		mid       = arg.Mid
	)
	aids := append(arg.Aids, arg.AidsWithoutPlayurl...)
	ap = make(map[int64]*archive.ArchiveWithPlayer, len(aids))
	as, _, isVip, isControl, err := s.baseForBatchAv(c, aids, mid, platform, "", arg.Buvid, false)
	if err != nil {
		return nil, err
	}
	if len(as) == 0 {
		return
	}
	for aid, arc := range as {
		ap[aid] = new(archive.ArchiveWithPlayer)
		ap[aid].Arc = arc
	}
	if !s.c.PlayerSwitch || platform == "" || platform == "unknown" || ip == "" || ip == "0.0.0.0" {
		return
	}
	if (platform == "iphone" && build <= model.QnIOSBuild) || (platform == "android" && build < model.QnAndroidBuild) {
		qn = s.c.Custom.PlayerQn
	}
	var (
		cidArr []*batch.RequestVideoItem
		paids  []int64
		eg     = errgroup.WithContext(c)
		pm     map[uint64]*batch.ResponseItem
		pgcm   map[int64]*archive.PlayerInfo
	)
	for _, aid := range arg.Aids {
		av, ok := ap[aid]
		if !ok {
			continue
		}
		if av.Duration > s.c.Custom.DurationLimit { //秒开文件长度太长客户端会奔溃
			continue
		}
		if showPGCPlayurl && av.AttrVal(archive.AttrBitIsPGC) == archive.AttrYes {
			paids = append(paids, av.Aid)
		}
		if av.Rights.Autoplay != 1 {
			continue
		}
		isSp := false
		// 1.稿件是自己的
		// 2.是vip&&没有被管控
		if (isVip && !isControl) || av.Author.Mid == mid {
			isSp = true
		}
		cidArr = append(cidArr, &batch.RequestVideoItem{Cid: uint64(av.FirstCid), IsSp: isSp})
	}
	if len(cidArr) == 0 && len(paids) == 0 {
		return
	}
	if len(cidArr) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			// player
			batchArg := &api.BatchPlayArg{
				Build:     arg.Build,
				Qn:        qn,
				MobiApp:   platform,
				Fnval:     fnval,
				Fnver:     fnver,
				Ip:        ip,
				ForceHost: forceHost,
				Buvid:     arg.Buvid,
				Mid:       mid,
			}
			if pm, err = s.playurldao.PlayurlBatch(ctx, cidArr, batchArg, 0, false); err != nil {
				log.Error("s.playurldao.PlayurlBatch err(%+v)", err)
			}
			return nil
		})
	}
	if len(paids) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			// pgc player
			if pgcm, err = s.arc.PGCPlayerInfos(ctx, paids, platform, ip, session, fnval, fnver); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait err(%+v)", err)
		return nil, err
	}
	for _, arc := range ap {
		pi, piok := pm[uint64(arc.FirstCid)]
		if piok {
			arc.PlayerInfo = new(archive.PlayerInfo)
			arc.PlayerInfo.Cid = int64(pi.Cid)
			arc.PlayerInfo.ExpireTime = int64(pi.ExpireTime)
			arc.PlayerInfo.FileInfo = make(map[int][]*archive.PlayerFileInfo)
			for qn, files := range pi.FileInfo {
				for _, f := range files.Infos {
					arc.PlayerInfo.FileInfo[int(qn)] = append(arc.PlayerInfo.FileInfo[int(qn)], &archive.PlayerFileInfo{
						FileSize:   int64(f.Filesize),
						TimeLength: int64(f.Timelength),
					})
				}
			}
			for _, v := range pi.SupportQuality {
				arc.PlayerInfo.SupportQuality = append(arc.PlayerInfo.SupportQuality, int(v))
			}
			arc.PlayerInfo.SupportFormats = pi.SupportFormats
			arc.PlayerInfo.SupportDescription = pi.SupportDescription
			arc.PlayerInfo.Quality = int(pi.Quality)
			arc.PlayerInfo.URL = pi.Url
			arc.PlayerInfo.VideoCodecid = pi.VideoCodecid
			arc.PlayerInfo.VideoProject = pi.VideoProject
			arc.PlayerInfo.Fnver = int(pi.Fnver)
			arc.PlayerInfo.Fnval = int(pi.Fnval)
			if pi.Dash != nil {
				arc.PlayerInfo.Dash = archive.FromDash(pi.Dash)
			}
			arc.PlayerInfo.NoRexcode = pi.NoRexcode
		}
		if pgci, ok := pgcm[arc.FirstCid]; ok {
			arc.PlayerInfo = pgci
			arc.Rights.Autoplay = 1
		}
	}
	return
}

// nolint:unparam
// mid以后也许要用
func (s *Service) simplePlayurl(mobiApp string, build, mid int64) bool {
	//白名单逻辑先注释，统一走简化后秒开结果
	//for _, locMid := range s.c.Custom.PlayurlMids {
	//	if mid == locMid {
	//		return false
	//	}
	//}
	return (mobiApp == "iphone" && build > s.c.Custom.SimplePlayurlIOS) || (mobiApp == "android" && build > s.c.Custom.SimplePlayurlAnd) || (mobiApp == "ipad" && build > s.c.Custom.SimplePlayurlIpad)
}

func (s *Service) setAskQn(batchArg *api.BatchPlayArg) int64 {
	// story模式下 秒开使用720p
	if batchArg.From == model.PlayurlFromStory {
		return s.setStoryQn(batchArg)
	}
	if batchArg.Mid == 0 {
		// 未登录用户最高可看480p清晰度
		if batchArg.Qn > model.Qn480 {
			batchArg.Qn = model.Qn480
		}
		// 未登录用户清晰度实验：https://www.tapd.bilibili.co/20095661/prong/stories/view/1120095661001941850
		return s.setNologinQn(batchArg)
	}
	// 秒开清晰度实验
	askQn := batchArg.Qn
	if askQn > model.QnFlv720 {
		askQn = model.QnFlv720
	}
	// mid白名单
	for _, mid := range s.c.Custom.UserQnGrayMids {
		if batchArg.Mid == mid {
			return askQn
		}
	}
	for _, mid := range s.c.Custom.WifiUserQnGrayMids {
		if batchArg.Mid == mid && batchArg.NetType == api.NetworkType_WIFI {
			return askQn
		}
	}
	mod := uint32(100)
	miaoMod := crc32.ChecksumIEEE([]byte(batchArg.Buvid+"miaokai_test")) % mod
	if miaoMod < s.c.Custom.UserQnGray {
		// 秒开清晰度下发【用户默认选择清晰度】，但最高为720P
		return askQn
	}
	if miaoMod >= s.c.Custom.UserQnGray && miaoMod < s.c.Custom.WifiUserQnGray {
		// 仅在Wi-Fi下秒开清晰度下发【用户默认选择清晰度】最高为720P；非Wi-Fi下保持逻辑不变，仍为固定480P起播
		if batchArg.NetType == api.NetworkType_WIFI {
			return askQn
		}
	}
	return 0
}

func (s *Service) setStoryQn(batchArg *api.BatchPlayArg) int64 {
	//wifi下1080p，非wifi下720p
	if batchArg.NetType == api.NetworkType_WIFI {
		return model.Qn1080
	}
	return model.QnFlv720
}

func (s *Service) setNologinQn(batchArg *api.BatchPlayArg) int64 {
	// 仅在wifi下做实验
	if batchArg.NetType != api.NetworkType_WIFI {
		return batchArg.Qn
	}
	// buvid白名单
	for _, buvid := range s.c.Custom.NologinQnBuvids {
		if batchArg.Buvid == buvid {
			return model.QnFlv720
		}
	}
	mod := uint32(100)
	nologinMod := crc32.ChecksumIEEE([]byte(batchArg.Buvid+"denglu_test")) % mod
	if nologinMod < s.c.Custom.NologinQnGray {
		return model.QnFlv720
	}
	return batchArg.Qn
}

// Archives3 multi get archives.
func (s *Service) Archives3(c context.Context, aids []int64, mid int64, mobiApp, device string) (map[int64]*api.Arc, error) {
	as, _, err := s.arc.Arcs(c, aids, mid, mobiApp, device)
	if err != nil {
		return nil, err
	}
	for _, a := range as {
		s.checkArcPremiereInfo(c, a)
		s.checkArcIpControl(a)
	}
	s.payCheck(c, mid, as)
	return as, nil
}

// ArchivesAndInc multi get archives.
func (s *Service) ArchivesAndInc(c context.Context, aids []int64, mid int64, mobiApp, device string) (map[int64]*api.Arc, map[int64]*api.ArcInternal, error) {
	as, incArcs, err := s.arc.Arcs(c, aids, mid, mobiApp, device)
	if err != nil {
		return nil, nil, err
	}
	for _, a := range as {
		s.checkArcPremiereInfo(c, a)
		s.checkArcIpControl(a)
	}
	// 对付费稿件进行鉴权
	s.payCheck(c, mid, as)
	return as, incArcs, nil
}

// Creators multi get Creators.
func (s *Service) Creators(c context.Context, aids []int64) (map[int64]*api.Creators, error) {
	res := make(map[int64]*api.Creators)
	as, err := s.Archives3(c, aids, 0, "", "")
	if err != nil {
		return nil, err
	}
	for aid, arc := range as {
		creator := &api.Creators{
			Owner: &api.Owner{Mid: arc.Author.Mid},
			Staff: arc.StaffInfo,
		}
		res[aid] = creator
	}
	return res, nil
}

// SimpleArc is
func (s *Service) SimpleArc(c context.Context, req *api.SimpleArcRequest) (*api.SimpleArc, error) {
	sa, err := s.arc.SimpleArc(c, req.Aid)
	if err != nil {
		return nil, err
	}
	s.checkSimpleArcPremiereInfo(c, sa)
	// 对付费稿件进行鉴权
	saMap := make(map[int64]*api.SimpleArc, 1)
	saMap[sa.Aid] = sa
	s.payCheckForSimpleArc(c, req.Mid, saMap)
	return sa, nil
}

func (s *Service) ArcsInner(c context.Context, aids []int64) (map[int64]*api.ArcInner, error) {
	ac, err := s.arc.ArcsInner(c, aids)
	if err != nil {
		return nil, err
	}
	ry := make(map[int64]*api.ArcInner)
	for k, v := range ac {
		if v == nil {
			continue
		}
		ry[k] = &api.ArcInner{Limit: v.UnfoldLimit()}
	}
	return ry, nil
}

// SimpleArc is
func (s *Service) SimpleArcs(c context.Context, aids []int64, mid int64) (map[int64]*api.SimpleArc, error) {
	sas, err := s.arc.SimpleArcs(c, aids)
	if err != nil {
		return nil, err
	}
	for _, a := range sas {
		s.checkSimpleArcPremiereInfo(c, a)
	}
	s.payCheckForSimpleArc(c, mid, sas)
	return sas, nil
}

// 是否大会员限免视频
func (s *Service) isVipFree(aid int64) bool {
	for _, v := range s.c.Custom.VipFreeAids {
		if aid == v {
			return true
		}
	}
	return false
}

func (s *Service) calCdnScore(c context.Context, pm map[uint64]*batch.ResponseItem, buvid string, locInfo *locgrpc.InfoCompleteReply, mid int64) map[string]map[string]string {
	if hit := s.hitTest(mid, buvid); !hit {
		return nil
	}
	var provinceID int64
	zoneID := locInfo.GetInfo().GetZoneId()
	zoneLen := 3
	if len(zoneID) >= zoneLen {
		provinceID = zoneID[2]
	}
	if len(pm) == 0 || provinceID == 0 {
		return nil
	}
	// 运营商改为完全匹配，忽略三大运营商之外请求
	isp := ""
	locIsp := locInfo.GetInfo().Isp
	switch locIsp {
	case _ispCM:
		isp = "cm"
	case _ispCT:
		isp = "ct"
	case _ispCU:
		isp = "cu"
	default:
		return nil
	}
	// 将需要处理的第三方域名拿出来
	tmpVideoHost := make(map[string]struct{})
	for _, p := range pm {
		// 提取需要进行处理的域名（音频不处理）
		if p.Dash == nil {
			continue
		}
		for _, v := range p.Dash.Video {
			for _, bkp := range v.BackupUrl {
				if bdm := api.GetThirdDomain(bkp); bdm != "" {
					tmpVideoHost[bdm] = struct{}{}
				}
			}
			if dm := api.GetThirdDomain(v.BaseUrl); dm != "" {
				tmpVideoHost[dm] = struct{}{}
			}
		}
		for _, v := range p.Dash.Audio {
			for _, bkp := range v.BackupUrl {
				if bdm := api.GetThirdDomain(bkp); bdm != "" {
					tmpVideoHost[bdm] = struct{}{}
				}
			}
			if dm := api.GetThirdDomain(v.BaseUrl); dm != "" {
				tmpVideoHost[dm] = struct{}{}
			}
		}
	}
	// 如果没有第三方域名
	if len(tmpVideoHost) == 0 {
		return nil
	}
	// 获取第三方域名选中的ip
	cdnMap := make(map[string]map[string]string)
	for vh := range tmpVideoHost {
		chooseIp := s.chooseCdn(c, middleware.CdnZoneKey(vh, isp, provinceID))
		if len(chooseIp) == 0 {
			continue
		}
		cdnMap[vh] = chooseIp
	}
	return cdnMap
}

func (s *Service) hitTest(mid int64, buvid string) bool {
	for _, cmid := range s.c.Custom.CdnMids {
		if mid == cmid {
			return true
		}
	}
	mod := uint32(1000)
	return crc32.ChecksumIEEE([]byte(buvid))%mod < s.c.Custom.CdnScoreGray
}

func (s *Service) chooseCdn(c context.Context, cdnZoneKey string) map[string]string {
	s.cdnScoresMu.RLock()
	score, ok := s.cdnScores[cdnZoneKey]
	s.cdnScoresMu.RUnlock()
	if !ok || time.Now().Unix()-score.LastUpTime > s.c.Custom.ScoreInternal { //暂定60秒更新一次可配置
		s.cache.Do(c, func(c context.Context) {
			s.setCdnCache(cdnZoneKey)
		})
	}

	res := make(map[string]string, 2)
	rand.Seed(time.Now().UnixNano())
	if len(score.WifiScoreIps) != 0 {
		ix := rand.Intn(len(score.WifiScoreIps))
		res["wifi"] = score.WifiScoreIps[ix]
	}
	if len(score.WwanScoreIps) != 0 {
		ix := rand.Intn(len(score.WwanScoreIps))
		res["wwan"] = score.WwanScoreIps[ix]
	}
	return res
}

func (s *Service) setCdnCache(cdnZoneKey string) {
	//从redis获取host对应ip及评分
	conn := s.cdnRedis.Get(context.Background())
	defer conn.Close()
	bs, err := redis.Bytes(conn.Do("GET", cdnZoneKey))
	if err != nil {
		if err == redis.ErrNil {
			s.writeCdnScore(cdnZoneKey, CdnScore{LastUpTime: time.Now().Unix()})
			return
		}
		log.Error("calCdnScore redis.ByteSlices err(%+v) cdnZoneKey(%s)", err, cdnZoneKey)
		return
	}
	var cdnScore map[string]map[string]float64
	if err := json.Unmarshal(bs, &cdnScore); err != nil {
		log.Error("calCdnScore json.Unmarshal err(%+v) cdnZoneKey(%s)", err, cdnZoneKey)
		return
	}
	s.writeCdnScore(cdnZoneKey, CdnScore{WwanScoreIps: s.getBestIpList(cdnScore["wwan"]),
		WifiScoreIps: s.getBestIpList(cdnScore["wifi"]), LastUpTime: time.Now().Unix()})
}

func (s *Service) getBestIpList(ipScores map[string]float64) []string {
	if len(ipScores) == 0 {
		return nil
	}
	var list []*api.IpScore
	for ip, score := range ipScores {
		list = append(list, &api.IpScore{
			Ip:    ip,
			Score: score,
		})
	}
	// 分数越小越靠前
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Score < list[j].Score
	})
	ix := int64(math.Floor(float64(len(list)) * s.c.Custom.ScoreRank))
	if ix < 1 {
		ix = 1
	}
	list = list[:ix]
	var bestIpList []string
	for _, v := range list {
		bestIpList = append(bestIpList, v.Ip)
	}
	return bestIpList
}

func (s *Service) writeCdnScore(cdnZoneKey string, cdnScore CdnScore) {
	log.Warn("writeCdnScore cdnZoneKey(%s) cdnScore(%+v)", cdnZoneKey, cdnScore)
	s.cdnScoresMu.Lock()
	s.cdnScores[cdnZoneKey] = cdnScore
	s.cdnScoresMu.Unlock()
}

func (s *Service) ArcsRedirectPolicy(c context.Context, aids []int64) (map[int64]*api.RedirectPolicy, error) {
	//参数校验
	aids, err := s.arcsRedirectParamsValidate(aids)
	if err != nil {
		return nil, err
	}
	redirects, err := s.arc.GetArcsRedirect(c, aids)
	if err != nil {
		return nil, err
	}
	return redirects, nil
}

func (s *Service) arcsRedirectParamsValidate(aids []int64) ([]int64, error) {
	if len(aids) == 0 || len(aids) > _MaxRedirectCnt {
		return nil, ecode.RequestErr
	}
	res := []int64{}
	for _, aid := range aids {
		if aid <= 0 {
			continue
		}
		res = append(res, aid)
	}
	//去重
	res = funk.UniqInt64(res)
	return res, nil
}

func (s *Service) ArcRedirectPolicyAddSrv(c context.Context, req *api.ArcRedirectPolicyAddRequest) error {
	//写入
	if err := s.arc.AddRedirect(c, req); err != nil {
		return err
	}
	return nil
}

func (s *Service) UpPremiereArcsSrv(c context.Context, req *api.UpPremiereArcsRequest) (map[int64]*api.UpArcs, error) {
	res := make(map[int64]*api.UpArcs, len(req.Mids))
	if len(req.Mids) == 0 {
		return res, nil
	}

	if len(s.recentPremiereArc) == 0 {
		return res, nil
	}

	//获取mid对应的最先首映中的稿件id
	aids := make([]int64, 0, len(req.Mids))
	for _, mid := range req.Mids {
		aids = append(aids, s.recentPremiereArc[mid]...)
	}

	if len(aids) == 0 {
		return res, nil
	}

	//获取稿件信息
	arcs, err := s.arc.SimpleArcs(c, aids)
	if err != nil {
		log.Error("GetUpPremiereArcs d.RawArcs(%v) error(%v)", req.Mids, err)
		return nil, err
	}
	for _, aid := range aids {
		if arcs[aid] == nil || arcs[aid].Premiere == nil {
			log.Error("GetUpPremiereArcs mids(%+v) arc or premiere miss aid(%d)", req.Mids, aid)
			continue
		}

		if res[arcs[aid].Mid] != nil {
			continue
		}

		if arcs[aid].IsNormalPremiere() && calPremiereState(time.Unix(arcs[aid].Premiere.StartTime, 0), arcs[aid].Duration) == api.PremiereState_premiere_in {
			upArcList := make([]*api.UpArc, 0, 1)
			upArc := &api.UpArc{
				Aid: aid,
			}
			upArcList = append(upArcList, upArc)
			res[arcs[aid].Mid] = &api.UpArcs{
				UpArc: upArcList,
			}
		}
	}
	return res, nil
}

func (s *Service) checkSimpleArcPremiereInfo(c context.Context, a *api.SimpleArc) {
	if !a.IsNormalPremiere() {
		return
	}
	//实时计算首映状态
	if a.Premiere != nil {
		a.Premiere.State = calPremiereState(time.Unix(a.Premiere.StartTime, 0), a.Duration)
		return
	}
	log.Error("首映稿件缺少信息 缓存未写入 aid(%d)", a.Aid)
	p, err := s.getArcPremiereInfo(c, a.Aid, a.Duration)
	if err != nil {
		log.Error("首映稿件缺少信息 获取信息失败 aid(%d)", a.Aid)
		return
	}
	a.Premiere = p
	log.Error("首映稿件缺少信息 填充信息 aid(%d) premiere(%+v)", a.Aid, a.Premiere)
}

func (s *Service) checkArcPremiereInfo(c context.Context, a *api.Arc) {
	if !a.IsNormalPremiere() {
		return
	}
	//实时计算首映状态
	if a.Premiere != nil {
		a.Premiere.State = calPremiereState(time.Unix(a.Premiere.StartTime, 0), a.Duration)
		return
	}
	log.Error("首映稿件缺少信息 缓存未写入 aid(%d)", a.Aid)
	p, err := s.getArcPremiereInfo(c, a.Aid, a.Duration)
	if err != nil {
		log.Error("首映稿件缺少信息 获取信息失败 aid(%d)", a.Aid)
		return
	}
	a.Premiere = p
	log.Error("首映稿件缺少信息 填充信息 aid(%d) premiere(%+v)", a.Aid, a.Premiere)
}

func (s *Service) getArcPremiereInfo(c context.Context, aid int64, duration int64) (*api.Premiere, error) {
	expand, err := s.arc.RawArchiveExpand(c, []int64{aid})
	if err != nil {
		return nil, err
	}
	if expand == nil || expand[aid] == nil {
		return nil, err
	}
	return &api.Premiere{
		State:     calPremiereState(expand[aid].PremiereTime.Time(), duration),
		StartTime: expand[aid].PremiereTime.Time().Unix(),
		RoomId:    expand[aid].RoomId,
	}, nil
}

func calPremiereState(premiereTime time.Time, duration int64) (state api.PremiereState) {
	cur := time.Now()
	end := premiereTime.Unix() + duration
	if cur.Before(premiereTime) {
		state = api.PremiereState_premiere_before
	} else if cur.After(time.Unix(end, 0)) {
		state = api.PremiereState_premiere_after
	} else {
		state = api.PremiereState_premiere_in
	}
	return state
}

// payCheck 检查稿件是否已付费, 绑定商品信息
func (s *Service) payCheck(c context.Context, mid int64, as map[int64]*api.Arc) {
	var (
		ugcPayAids []int64
		seasonIds  []int64
	)

	for aid, arc := range as {
		if arc.AttrValV2(api.AttrBitV2Pay) == api.AttrNo {
			continue
		}
		if arc.Pay == nil {
			log.Error("日志告警 付费稿件 缺少付费类型属性 Aid(%d)", aid)
			missAddits, err := s.arc.RawAddits(c, []int64{arc.Aid})
			if err != nil {
				log.Error("日志告警 付费稿件 获取付费类型失败 aid(%d) d.RawAddits error(%+v)", arc.Aid, err)
				continue
			}
			if missAddits[arc.Aid] != nil {
				arc.Pay = &api.PayInfo{
					PayAttr: missAddits[arc.Aid].Subtype,
				}
			}
			if arc.Pay.AttrVal(api.PaySubTypeAttrBitSeason) == api.AttrYes {
				episode, err := s.arc.RawSeasonEpisode(c, arc.SeasonID, arc.Aid)
				if err != nil {
					log.Error("日志告警 付费稿件 获取免费试看失败 aid(%d) d.RawAddits error(%+v)", arc.Aid, err)
					continue
				}
				if episode != nil {
					arc.Rights.ArcPayFreeWatch = episode.AttrVal(ugcmdl.EpisodeAttrSnFreeWatch)
				}
			}
			log.Error("日志告警 付费稿件 成功设置付费类型属性 Aid(%d)", aid)
		}
		ugcPayAids = append(ugcPayAids, aid)
		//如果是合集付费类型
		if arc.Pay.AttrVal(api.PaySubTypeAttrBitSeason) == api.AttrYes && arc.SeasonID > 0 {
			seasonIds = append(seasonIds, arc.SeasonID)
		}
	}

	if len(ugcPayAids) == 0 {
		return
	}

	sgm := s.getPaidSeasonGoodsInfo(c, mid, seasonIds)
	for _, aid := range ugcPayAids {
		goodsInfo := make([]*api.GoodsInfo, 0, 1)
		//添加付费合集的商品信息
		if as[aid].SeasonID != 0 {
			sg := &api.GoodsInfo{
				PayState:  api.PayState_PayStateUnknown,
				Category:  api.Category_CategorySeason,
				FreeWatch: as[aid].Rights.ArcPayFreeWatch == 1,
			}
			if sgm[as[aid].SeasonID] != nil {
				sg.PayState = sgm[as[aid].SeasonID].PayState
				sg.GoodsId = sgm[as[aid].SeasonID].GoodsId
				sg.GoodsName = sgm[as[aid].SeasonID].GoodsName
				sg.GoodsPrice = sgm[as[aid].SeasonID].GoodsPrice
				sg.GoodsPriceFmt = sgm[as[aid].SeasonID].GoodsPriceFmt
			}
			goodsInfo = append(goodsInfo, sg)
		}
		as[aid].Pay.GoodsInfo = goodsInfo
	}
}

// payCheckForSimpleArc 检查稿件是否已付费, 0未付费, 1已付费
func (s *Service) payCheckForSimpleArc(c context.Context, mid int64, as map[int64]*api.SimpleArc) {
	//获取付费稿件
	var (
		ugcPayAids []int64
		seasonIds  []int64
	)

	for aid, arc := range as {
		if arc.AttrValV2(api.AttrBitV2Pay) == api.AttrNo {
			continue
		}
		if arc.Pay == nil {
			log.Error("日志告警 付费稿件 缺少付费类型属性 simpleArc Aid(%d)", aid)
			continue
		}
		ugcPayAids = append(ugcPayAids, aid)
		//如果是合集付费类型
		if arc.Pay.AttrVal(api.PaySubTypeAttrBitSeason) == api.AttrYes && arc.SeasonId > 0 {
			seasonIds = append(seasonIds, arc.SeasonId)
		}
	}

	if len(ugcPayAids) == 0 {
		return
	}

	sgm := s.getPaidSeasonGoodsInfo(c, mid, seasonIds)
	for _, aid := range ugcPayAids {
		goodsInfo := make([]*api.GoodsInfo, 0, 1)
		//添加付费合集的商品信息
		if as[aid].SeasonId != 0 {
			sg := &api.GoodsInfo{
				PayState:  api.PayState_PayStateUnknown,
				Category:  api.Category_CategorySeason,
				FreeWatch: as[aid].Rights.ArcPayFreeWatch == 1,
			}
			if sgm[as[aid].SeasonId] != nil {
				sg.PayState = sgm[as[aid].SeasonId].PayState
				sg.GoodsId = sgm[as[aid].SeasonId].GoodsId
				sg.GoodsName = sgm[as[aid].SeasonId].GoodsName
				sg.GoodsPrice = sgm[as[aid].SeasonId].GoodsPrice
				sg.GoodsPriceFmt = sgm[as[aid].SeasonId].GoodsPriceFmt
			}
			goodsInfo = append(goodsInfo, sg)
		}
		as[aid].Pay.GoodsInfo = goodsInfo
	}
}

// getPaidSeasonGoodsInfo 获取付费合集的商品信息以及付费状态
func (s *Service) getPaidSeasonGoodsInfo(c context.Context, mid int64, seasonIds []int64) map[int64]*api.GoodsInfo {
	goodsInfoMap := make(map[int64]*api.GoodsInfo, len(seasonIds))

	if len(seasonIds) == 0 {
		return goodsInfoMap
	}

	//获取用户付费状态
	res, err := s.vasDao.SeasonUserVoucherBatch(c, mid, seasonIds)
	if err != nil || res == nil {
		log.Error("日志告警 付费稿件 s.vasDao.SeasonUserVoucherBatch mid(%d) seasonIds(%+v) res(%+v) error(%+v)",
			mid, seasonIds, res, err)
		return goodsInfoMap
	}

	for _, sid := range seasonIds {
		paid := res.Result[sid] == vasApi.VoucherState_VoucherStateActive

		item := res.Items[sid]
		if item == nil {
			log.Error("日志告警 付费稿件 缺少商品信息 mid(%d) sid(%+v) res(%+v) error(%+v)",
				mid, sid, res, err)
			continue
		}

		//绑定合集付费商品
		payState := api.PayState_PayStateUnknown
		//降级开启时，默认为未付费
		if !s.c.Switch.DegreePayCheck && mid != 0 && paid {
			payState = api.PayState_PayStateActive
		}
		goodsInfo := &api.GoodsInfo{
			GoodsId:       item.ProductId,
			GoodsPrice:    item.Price,
			GoodsPriceFmt: item.PriceFmt,
			GoodsName:     item.Name,
			PayState:      payState,
		}
		goodsInfoMap[sid] = goodsInfo
	}
	return goodsInfoMap
}

// checkArcIpControl 检查IP地址管控
func (s *Service) checkArcIpControl(a *api.Arc) {
	if fixedIp, ok := s.buvidFixedIp[a.Aid]; ok {
		a.PubLocation = fixedIp
		return
	}
	if fixedIp, ok := s.upMidFixedIp[a.Author.Mid]; ok {
		a.PubLocation = fixedIp
		return
	}
}
