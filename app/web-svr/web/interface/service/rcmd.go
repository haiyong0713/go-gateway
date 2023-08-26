package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	"go-gateway/pkg/idsafe/bvid"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/stat/prom"
	"go-common/library/sync/errgroup.v2"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	xecode "go-gateway/app/web-svr/web/ecode"
	"go-gateway/app/web-svr/web/interface/model"
	"go-gateway/app/web-svr/web/interface/model/rcmd"

	api "git.bilibili.co/bapis/bapis-go/account/service"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	"github.com/pkg/errors"
)

const (
	_Group0 uint64 = iota
	_Group1
	_Group2
	_Group3
	_Group4
	_Group5
	_Group6
	_Group7
	_Group8
	_Group9
	_GroupHit
)

func (s *Service) WebTopRcmd(ctx context.Context, buvid string, mid int64, freshType, version, count int, ip string, freshIdx, freshIdx1h int64) (*rcmd.TopRcmd, error) {
	var (
		item []*rcmd.Item
		err  error
	)
	group := s.abTest(mid, buvid, version)
	switch group {
	case rcmd.GroupA: // 对照组，未登录用户，热门数据，存放在内存中
		// 为了不记录index，所以一次性返回较多的数据，前端分页
		ps := 30
		if version == 1 {
			ps = 24
		}
		item, err = s.topHot(ctx, ps, false)
	case rcmd.GroupB: // 实验组，登录用户｜新首页未登录用户分组推荐，实时调用AI接口
		// ps不能过大，资源池有限
		ps := 10
		if version == 1 {
			ps = 8 // 新版返回8个
			if count > 0 {
				ps = count
			}
		}
		api := "/x/web-interface/index/top/rcmd"
		ignoreFeedVersion := "V0"
		item, err = s.topRcmd(ctx, buvid, mid, freshType, ps, ip, api, 0, freshIdx, freshIdx1h, ignoreFeedVersion, 0)
	}
	if err != nil {
		return nil, err
	}
	tmp := item
	for _, v := range tmp {
		v.AvFeature = nil
	}
	res := &rcmd.TopRcmd{
		Item: tmp,
		Abtest: &rcmd.Abtest{
			Group: group,
		},
		UserFeature: nil,
	}
	return res, nil
}

func (s *Service) abTest(mid int64, buvid string, version int) rcmd.Group {
	// a:对照组，b：实验组
	if mid < 1 {
		if version == 1 && s.groupTest(buvid) { // 对新首页的未登录用户增加分组推荐实验
			return rcmd.GroupB
		}
		return rcmd.GroupA
	}
	for _, val := range s.c.Rcmd.Whitelist {
		if mid == val {
			return rcmd.GroupB
		}
	}
	h := md5.New()
	if _, err := h.Write([]byte(strconv.FormatInt(mid, 10))); err != nil {
		log.Error("%+v", err)
	}
	b, err := strconv.ParseUint(hex.EncodeToString(h.Sum(nil))[18:], 16, 64)
	if err != nil {
		log.Error("日志告警 分组错误,error:%+v", err)
		return rcmd.GroupA
	}
	if b%100 < s.c.Rcmd.Bucket {
		return rcmd.GroupB
	}
	return rcmd.GroupA
}

// nolint: gocognit,gomnd
func (s *Service) topRcmd(ctx context.Context, buvid string, mid int64, freshType, ps int, ip, api string, isFeed, freshIdx, freshIdx1h int64, feedVersion string, yNum int64) (item []*rcmd.Item, err error) {
	const (
		_av = "av"
	)
	var (
		data        []*rcmd.AITopRcmd
		code        int
		isRec       int
		trackid     string
		showlist    *rcmd.Showlist
		userFeature json.RawMessage
	)
	defer func() {
		bs, _ := json.Marshal(showlist)
		s.InfocRcmd(&model.RcmdInfoc{
			API:         api,
			IP:          ip,
			Mid:         mid,
			Buvid:       buvid,
			Ptype:       1,
			Time:        time.Now().Unix(),
			FreshType:   freshType,
			IsRec:       isRec,
			Trackid:     trackid,
			ReturnCode:  code,
			UserFeature: string(userFeature),
			Showlist:    string(bs),
			IsFeed:      isFeed,
			FreshIdx:    freshIdx,
			FreshIdx1h:  freshIdx1h,
			FeedVersion: feedVersion,
			YNum:        yNum,
		})
	}()
	if item, err = func() ([]*rcmd.Item, error) {
		prom.BusinessInfoCount.Incr("首页天马AI接口调用")
		data, userFeature, code, err = s.rcmdDao.TopRcmd(ctx, mid, freshType, ps, ip, buvid, isFeed, freshIdx, freshIdx1h, yNum, feedVersion)
		if err != nil {
			prom.BusinessErrCount.Incr(fmt.Sprintf("首页天马AI接口调用失败,%v", code))
			return nil, err
		}
		var aids []int64
		for _, v := range data {
			switch v.Goto {
			case _av:
				aids = append(aids, v.ID)
			default:
				log.Error("日志告警 首页天马实验组无效的goto:%v", v.Goto)
			}
		}
		if len(aids) == 0 {
			return nil, errors.New("top rcmd aids len is 0")
		}
		arcs, err1 := s.batchArchives(ctx, aids)
		if err1 != nil {
			return nil, err1
		}
		var (
			rcmdItem    []*rcmd.Item
			sectionItem []*rcmd.SectionItem
		)
		pos := 1
		for _, v := range data {
			switch v.Goto {
			case _av:
				if arc, ok := arcs[v.ID]; ok && arc != nil && arc.IsNormal() {
					i := &rcmd.Item{}
					i.FromArc(arc, v)
					rcmdItem = append(rcmdItem, i)
					sectionItem = append(sectionItem, &rcmd.SectionItem{
						AvFeature: v.AvFeature,
						Goto:      v.Goto,
						ID:        v.ID,
						Pos:       pos,
						Source:    v.Source,
					})
					pos++
				}
			}
		}
		if len(rcmdItem) < ps {
			return nil, errors.New(fmt.Sprintf("top rcmd card count less than %v", ps))
		}
		isRec = 1
		trackid = data[0].Trackid
		showlist = &rcmd.Showlist{
			Section: &rcmd.Section{
				Items: sectionItem,
			},
		}
		return rcmdItem, nil
	}(); err != nil {
		// 1、换一换场景
		//    a、ai侧出现异常，返回码为500，走网关灾备
		//    b、如果网关侧过滤掉结果不足10个，走网关灾备
		//    c、如果ai侧返回码为 -3 时，前端提示 “没有新内容”
		//    d、其他返回码时，走网关灾备
		// 2、初始化首页，如下情况走网关灾备
		//    a、ai侧出现异常，返回码为500
		//    b、用户刷新，ai侧return_code = -3
		//    c、用户刷新，ai侧返回网关10个结果，网关过滤后，不足10个
		// 3、网关灾备样式
		//    网关灾备时，前端会继续出新样式，“换一换”的样式
		// 4、网关灾备
		//    从获取的结果30个结果中， 随机出10个
		log.Error("首页天马AI接口错误,error:%+v", err)
		if ecode.EqualError(ecode.Int(-3), err) {
			if freshType == 3 {
				return nil, xecode.CardNothingFound
			}
		}
		if item, err = s.topHot(ctx, ps, false); err != nil {
			return nil, err
		}
		index := func() []int {
			if len(item) < ps {
				return nil
			}
			var index []int
			exists := map[int]struct{}{}
			rand.Seed(time.Now().UnixNano())
			for {
				val := rand.Intn(len(item))
				if _, ok := exists[val]; !ok {
					exists[val] = struct{}{}
					index = append(index, val)
					if len(index) == ps {
						return index
					}
					continue
				}
			}
		}()
		var (
			randItem    []*rcmd.Item
			sectionItem []*rcmd.SectionItem
		)
		pos := 1
		for _, val := range index {
			v := item[val]
			randItem = append(randItem, v)
			sectionItem = append(sectionItem, &rcmd.SectionItem{
				AvFeature: nil,
				Goto:      v.Goto,
				ID:        v.ID,
				Pos:       pos,
				Source:    "",
			})
			pos++
		}
		if len(randItem) < ps {
			return nil, errors.New(fmt.Sprintf("top rcmd card count less than %v", ps))
		}
		item = randItem
		showlist = &rcmd.Showlist{
			Section: &rcmd.Section{
				Items: sectionItem,
			},
		}
	}
	return item, nil
}

func randomSelect(arr []*rcmd.Item, ps int) []*rcmd.Item {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(arr))

	res := []*rcmd.Item{}

	for i := 0; i < ps; i++ {
		if index >= len(arr) {
			index = index % len(arr)
		}
		res = append(res, arr[index])
		index++
	}

	return res

}

func (s *Service) topHot(ctx context.Context, ps int, needRandom bool) ([]*rcmd.Item, error) {
	arcs, err := func() ([]*arcmdl.Arc, error) {
		aids := s.webTopData
		if len(aids) == 0 {
			return nil, errors.New("top hot aids len is 0")
		}
		arcs, err := s.batchArchives(ctx, aids)
		if err != nil {
			return nil, err
		}
		var res []*arcmdl.Arc
		for _, aid := range aids {
			if arc, ok := arcs[aid]; ok && arc != nil && arc.IsNormal() {
				res = append(res, arc)
			}
		}
		if len(res) < ps {
			return nil, errors.New(fmt.Sprintf("top hot card ps less than %v", ps))
		}
		return res, nil
	}()
	if err != nil {
		log.Error("日志告警 首页天马对照组获取热门数据错误:%+v", err)
		if arcs, err = s.dao.WebTopHotBakCache(ctx); err != nil {
			log.Error("日志告警 首页天马对照组获取兜底数据错误:%+v", err)
			return nil, xecode.CardNothingFound
		}
	}
	var item []*rcmd.Item
	for _, arc := range arcs {
		i := &rcmd.Item{}
		i.FromArc(arc, nil)
		item = append(item, i)
	}
	if len(item) > ps {

		if needRandom {

			item = randomSelect(item, ps)
		} else {
			item = item[:ps]
		}

	}

	return item, nil
}

// 对buvid计算分组,检查对应分组开关是否开启,返回实验开启状态
func (s *Service) groupTest(buvid string) bool {
	if buvid == "" {
		return s.c.Rcmd.GroupTest.EmptyBuvid
	}
	ctx := md5.New()
	ctx.Write([]byte(buvid + "CF246FE41AD8452C"))
	f, err := strconv.ParseUint(hex.EncodeToString(ctx.Sum(nil))[18:], 16, 64)
	if err != nil {
		log.Error("分组错误,error:%+v", err)
		return false
	}
	g := f % _GroupHit
	switch g {
	case _Group0:
		return s.c.Rcmd.GroupTest.Group0
	case _Group1:
		return s.c.Rcmd.GroupTest.Group1
	case _Group2:
		return s.c.Rcmd.GroupTest.Group2
	case _Group3:
		return s.c.Rcmd.GroupTest.Group3
	case _Group4:
		return s.c.Rcmd.GroupTest.Group4
	case _Group5:
		return s.c.Rcmd.GroupTest.Group5
	case _Group6:
		return s.c.Rcmd.GroupTest.Group6
	case _Group7:
		return s.c.Rcmd.GroupTest.Group7
	case _Group8:
		return s.c.Rcmd.GroupTest.Group8
	case _Group9:
		return s.c.Rcmd.GroupTest.Group9
	}
	return false
}

func (s *Service) WebTopFeedRcmd(ctx context.Context, mid, freshType, ps, freshIdx, freshIdx1h, yNum int64, feedVersion string, ip, buvid string) (*rcmd.TopRcmd, error) {
	api := "/x/web-interface/index/top/feed/rcmd"
	item, err := s.topRcmd(ctx, buvid, mid, int(freshType), int(ps), ip, api, 1, freshIdx, freshIdx1h, feedVersion, yNum)
	if err != nil {
		return nil, err
	}
	res := &rcmd.TopRcmd{
		Item:        item,
		UserFeature: nil,
	}
	return res, nil
}

// 获取位置信息
func (s *Service) LocalInfo(c context.Context, ip string) (res *locgrpc.InfoComplete, err error) {
	var (
		args = &locgrpc.InfoCompleteReq{Addr: ip}
		info *locgrpc.InfoCompleteReply
	)
	if info, err = s.locGRPC.InfoComplete(c, args); err != nil {
		log.Error("%+v", err)
		return
	}
	res = info.GetInfo()
	return
}

func (s *Service) WebTopFeedRcmdV2(ctx context.Context, req *rcmd.TopRcmdReq) (*rcmd.TopRcmd, error) {
	var (
		info     *locgrpc.InfoComplete
		localErr error
	)
	if info, localErr = s.LocalInfo(ctx, req.Ip); localErr != nil {
		log.Error("%+v", localErr)
	}
	if info != nil {
		req.Country = info.Country
		req.Province = info.Province
		req.City = info.City
	}
	req.Api = "/x/web-interface/index/top/feed/rcmd"
	if req.Buvid == "" && req.Mid == 0 {
		item, err := s.topHot(ctx, req.Ps, true)

		if err != nil {
			return nil, err
		}

		return &rcmd.TopRcmd{
			Item: item,
		}, nil

	} else {
		ret, err := s.topFeedRcmd(ctx, req)
		if err != nil {
			return nil, err
		}
		tmp := ret.DataItem
		for _, v := range tmp {
			v.AvFeature = nil
		}
		res := &rcmd.TopRcmd{
			Item:                  tmp,
			BusinessCard:          ret.BusinessItem,
			FloorInfo:             ret.FloorInfo,
			UserFeature:           nil,
			PreloadExposePct:      ret.PreloadExposePct,
			PreloadFloorExposePct: ret.PreloadFloorExposePct,
		}
		return res, nil
	}

}

// feeds 补量计划
func (s *Service) topFeedRcmd(ctx context.Context, req *rcmd.TopRcmdReq) (*rcmd.TopFeedRcmdReply, error) {
	// 变量定义
	var (
		code                    int
		isRec                   int
		trackid                 string
		showlist                *rcmd.Showlist
		userFeature             json.RawMessage
		dataItems, businessItem []*rcmd.Item
		floorInfos              []*rcmd.AIRcmdFloorInfo
		err                     error
		preloadExposePct        float32
		preloadFloorExposePct   float32
	)
	// 关闭时上报
	defer func() { s.reportRcmd(req, isRec, code, trackid, string(userFeature), showlist) }()
	// 开始获取数据
	if dataItems, businessItem, err = func() ([]*rcmd.Item, []*rcmd.Item, error) {
		prom.BusinessInfoCount.Incr("首页Feed天马AI接口调用")
		res, code, err := s.rcmdDao.TopFeedRcmd(ctx, req)
		if err != nil || code != 0 {
			prom.BusinessErrCount.Incr(fmt.Sprintf("首页Feed天马AI接口调用失败,%v", code))
			return nil, nil, err
		}
		// 设置楼层信息
		floorInfos = res.FloorInfos

		preloadExposePct = res.PreloadExposePct
		preloadFloorExposePct = res.PreloadFloorExposePct

		dataIds, businessIds := s.getResIds(res)
		// business 数据区 先处理保证容灾情况时，business有数据
		var businessCards []*rcmd.Item
		if businessIds != nil && businessIds.HasData {
			businessCards, _, _ = s.getItemsByIds(ctx, businessIds, res.BusinessCards, true)
		}

		aidsLen := len(dataIds.Aids)
		// 判断视频稿件是否数量满足
		if aidsLen == 0 {
			return nil, businessCards, errors.New("top feed rcmd aids len is 0")
		}
		// data 数据区
		rcmdItem, sectionItem, err := s.getItemsByIds(ctx, dataIds, res.Data, false)
		if err != nil {
			return nil, businessCards, err
		}
		if len(rcmdItem) < req.Ps {
			return nil, businessCards, errors.New(fmt.Sprintf("top rcmd card count less than %v", req.Ps))
		}
		// 填充上报数据
		isRec = 1
		trackid = rcmdItem[0].TrackId
		showlist = &rcmd.Showlist{
			Section: &rcmd.Section{
				Items: sectionItem,
			},
		}
		userFeature = res.UserFeature
		return rcmdItem, businessCards, nil
	}(); err != nil {
		// 容灾处理
		rcmdItem, sectionItem, err := s.disasterRec(ctx, code, err, req)
		if err != nil {
			return nil, err
		}
		showlist = &rcmd.Showlist{
			Section: &rcmd.Section{
				Items: sectionItem,
			},
		}
		dataItems = rcmdItem
	}
	return &rcmd.TopFeedRcmdReply{
		DataItem:              dataItems,
		BusinessItem:          businessItem,
		FloorInfo:             floorInfos,
		UserFeature:           userFeature,
		PreloadExposePct:      preloadExposePct,
		PreloadFloorExposePct: preloadFloorExposePct,
	}, nil
}

// nolint:gocognit
func (s *Service) getItemsByIds(ctx context.Context, in *rcmd.TopRcmdIds, origin []*rcmd.AITopRcmd, isBusiness bool) ([]*rcmd.Item, []*rcmd.SectionItem, error) {
	if len(origin) <= 0 {
		return nil, nil, nil
	}
	var (
		rcmdItem                   []*rcmd.Item
		sectionItem                []*rcmd.SectionItem
		arcs                       map[int64]*arcmdl.Arc
		rooms                      map[int64]*model.LiveRoomInfo
		seasons                    map[int32]*seasongrpc.CardInfoProto
		liveUpMids                 []int64
		arcErr, liveErr, seasonErr error
		accInfos                   map[int64]*api.Info
		aids                       []int64
		totalAids                  []int64
	)
	const CreativeVideo = int8(1)
	totalAids = append(totalAids, in.Aids...)
	for _, v := range origin {
		if v.BusinessInfo == nil || v.BusinessInfo.AdInfo == nil || v.BusinessInfo.AdInfo.Info == nil {
			continue
		}
		info := v.BusinessInfo.AdInfo.Info
		if info.CreativeType == CreativeVideo {
			aids = append(aids, info.CreativeContent.VideoID)
			totalAids = append(totalAids, aids...)
		}
	}
	group := errgroup.WithContext(ctx)
	if len(totalAids) > 0 {
		// 视频稿件
		group.Go(func(ctx context.Context) error {
			if arcs, arcErr = s.batchArchives(ctx, totalAids); arcErr != nil {
				log.Error("【@getItemsByIds】首页Feed天马获取直播数据失败 Aids(%v), error(%v)", totalAids, arcErr)
				return arcErr
			}
			return nil
		})
	}
	if len(in.LiveIds) > 0 {
		// 直播
		group.Go(func(ctx context.Context) error {
			if rooms, liveErr = s.liveDao.RoomByIds(ctx, in.LiveIds); liveErr != nil {
				log.Error("【@getItemsByIds】首页Feed天马获取直播数据失败 roomids(%v), error(%v)", in.LiveIds, liveErr)
				return liveErr
			}
			// 加入mid list 用于查询
			for _, v := range rooms {
				if v == nil {
					continue
				}
				liveUpMids = append(liveUpMids, v.Uid)
			}
			return nil
		})
	}
	if len(in.SeasonIds) > 0 {
		// ogv
		group.Go(func(ctx context.Context) error {
			if seasons, seasonErr = s.ogvSeasonCards(ctx, in.SeasonIds); seasonErr != nil {
				log.Error("【@getItemsByIds】首页Feed天马获取Pgc数据失败 seasonIds(%v), error(%v)", in.SeasonIds, seasonErr)
				return seasonErr
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, nil, err
	}
	if len(liveUpMids) > 0 {
		// 查询直播的up主
		accInfos = func() map[int64]*api.Info {
			accReply, err := s.accGRPC.Infos3(ctx, &api.MidsReq{Mids: liveUpMids})
			if err != nil {
				log.Error("【@getItemsByIds】s.accGRPC.Infos3 mids(%v) error(%v)", liveUpMids, err)
				return nil
			}
			return accReply.GetInfos()
		}()
	}
	// 填充数据
	pos := 1
	for _, v := range origin {
		var i *rcmd.Item
		switch v.Goto {
		case rcmd.AV:
			if arc, ok := arcs[v.ID]; ok && arc != nil && arc.IsNormal() {
				i = &rcmd.Item{}
				i.FromArc(arc, v)
			}
		case rcmd.Live:
			if info, ok := rooms[v.ID]; ok && info != nil {
				i = &rcmd.Item{}
				i.FromLive(info, v)
				liveUser, ok := accInfos[info.Uid]
				if !ok || liveUser == nil {
					continue
				}
				// 填充名字头像
				i.Owner.Name = liveUser.Name
				i.Owner.Face = liveUser.Face
			}
		case rcmd.Ogv:
			if info, ok := seasons[int32(v.ID)]; ok && info != nil {
				i = &rcmd.Item{}
				i.FromOgv(info, v)
			}
		case rcmd.Ad:
			i = &rcmd.Item{}
			i.Goto = "ad"
			i.FromAd(v)
			if i.BusinessInfo != nil {
				if arcInfo, ok := arcs[i.BusinessInfo.Aid]; ok {
					bvID, _ := bvid.AvToBv(i.BusinessInfo.Aid)
					i.BusinessInfo.Archive = &rcmd.ArchiveBV{
						Arc:  arcInfo,
						BVID: bvID,
					}
				}
			}
		}
		if i != nil {
			i.IsStock = v.IsStock
			const _isStock = 1
			if v.IsStock == _isStock {
				i.FromAd(v)
			}
			rcmdItem = append(rcmdItem, i)
			if !isBusiness {
				sectionItem = append(sectionItem, &rcmd.SectionItem{
					AvFeature: v.AvFeature,
					Goto:      v.Goto,
					ID:        v.ID,
					Pos:       pos,
					Source:    v.Source,
				})
				pos++
			}
		}
	}
	if isBusiness && len(rcmdItem) > 0 {
		// 如果是business card，应前端需要，填充空的占位的av卡
		var businessCard []*rcmd.Item
		sort.SliceStable(rcmdItem, func(i, j int) bool {
			return rcmdItem[i].Pos < rcmdItem[j].Pos
		})
		index := 0
		max := rcmdItem[len(rcmdItem)-1].Pos
		for i := 1; i <= max; i++ {
			if i < rcmdItem[index].Pos {
				businessCard = append(businessCard, &rcmd.Item{
					Goto: rcmd.AV,
					Pos:  i,
				})
				continue
			}
			businessCard = append(businessCard, rcmdItem[index])
			index++
		}
		return businessCard, sectionItem, nil
	}
	return rcmdItem, sectionItem, nil
}

// 通过天马返回结果获取ids
func (s *Service) getResIds(in *rcmd.TopFeedRcmdRep) (dataIds *rcmd.TopRcmdIds, businessIds *rcmd.TopRcmdIds) {
	if in == nil && in.Code != 0 {
		return
	}
	if in.Data != nil && len(in.Data) > 0 {
		dataIds = &rcmd.TopRcmdIds{
			HasData: true,
		}
		for _, v := range in.Data {
			switch v.Goto {
			case rcmd.AV:
				dataIds.Aids = append(dataIds.Aids, v.ID)
			case rcmd.Live:
				dataIds.LiveIds = append(dataIds.LiveIds, v.ID)
			case rcmd.Ad:
				dataIds.AdIds = append(dataIds.AdIds, v.ID)
			case rcmd.Ogv:
				dataIds.SeasonIds = append(dataIds.SeasonIds, int32(v.ID))
			default:
				log.Error("日志告警 首页Feed天马无效的Data goto:%v", v.Goto)
			}
		}
	}
	if in.BusinessCards != nil && len(in.BusinessCards) > 0 {
		businessIds = &rcmd.TopRcmdIds{
			HasData: true,
		}
		for _, v := range in.BusinessCards {
			switch v.Goto {
			case rcmd.Live:
				businessIds.LiveIds = append(businessIds.LiveIds, v.ID)
			case rcmd.Ogv:
				businessIds.SeasonIds = append(businessIds.SeasonIds, int32(v.ID))
			case rcmd.Ad:
				businessIds.AdIds = append(businessIds.AdIds, v.ID) // 仅占位
			default:
				log.Error("日志告警 首页Feed天马无效的Business goto:%v", v.Goto)
			}
		}
	}
	return
}

/*
***
容灾策略
1、换一换场景

	a、ai侧出现异常，返回码为500，走网关灾备
	b、如果网关侧过滤掉结果不足10个，走网关灾备
	c、如果ai侧返回码为 -3 时，前端提示 “没有新内容”
	d、其他返回码时，走网关灾备

2、初始化首页，如下情况走网关灾备

	a、ai侧出现异常，返回码为500
	b、用户刷新，ai侧return_code = -3
	c、用户刷新，ai侧返回网关10个结果，网关过滤后，不足10个

3、网关灾备样式

	网关灾备时，前端会继续出新样式，“换一换”的样式

4、网关灾备
*/
func (s *Service) disasterRec(ctx context.Context, code int, inerr error, req *rcmd.TopRcmdReq) (item []*rcmd.Item, secItem []*rcmd.SectionItem, err error) {
	// 常量
	const (
		_lackNum = -3
		_change  = 3 // 换一换
	)
	ps := req.Ps
	log.Error("首页Feed天马AI接口错误,error:%+v", err)
	if code == _lackNum || ecode.EqualError(ecode.Int(-3), inerr) {
		if req.FreshType == _change {
			return nil, nil, xecode.CardNothingFound
		}
	}
	if item, err = s.topHot(ctx, req.Ps, false); err != nil {
		return nil, nil, err
	}
	index := func() []int {
		if len(item) < ps {
			return nil
		}
		var index []int
		exists := map[int]struct{}{}
		rand.Seed(time.Now().UnixNano())
		for {
			val := rand.Intn(len(item))
			if _, ok := exists[val]; !ok {
				exists[val] = struct{}{}
				index = append(index, val)
				if len(index) == ps {
					return index
				}
				continue
			}
		}
	}()
	var (
		randItem    []*rcmd.Item
		sectionItem []*rcmd.SectionItem
	)
	pos := 1
	for _, val := range index {
		v := item[val]
		randItem = append(randItem, v)
		sectionItem = append(sectionItem, &rcmd.SectionItem{
			AvFeature: nil,
			Goto:      v.Goto,
			ID:        v.ID,
			Pos:       pos,
			Source:    "",
		})
		pos++
	}
	if len(randItem) < ps {
		return nil, nil, errors.New(fmt.Sprintf("【@disasterRec】top rcmd card count less than %v", ps))
	}
	item = randItem
	return randItem, sectionItem, nil
}

// 写入hive表
func (s *Service) reportRcmd(req *rcmd.TopRcmdReq, isRec, code int, trackid, userFeature string, showlist *rcmd.Showlist) {
	bs, _ := json.Marshal(showlist)
	s.InfocRcmd(&model.RcmdInfoc{
		API:         req.Api,
		IP:          req.Ip,
		Mid:         req.Mid,
		Buvid:       req.Buvid,
		Ptype:       1,
		Time:        time.Now().Unix(),
		FreshType:   req.FreshType,
		IsRec:       isRec,
		Trackid:     trackid,
		ReturnCode:  code,
		UserFeature: userFeature,
		Showlist:    string(bs),
		IsFeed:      int64(req.IsFeed),
		FreshIdx:    int64(req.FreshIdx),
		FreshIdx1h:  int64(req.FreshIdx1h),
		FeedVersion: req.FeedVersion,
		YNum:        int64(req.YNum),
	})
}

// 获取ogv card（service中已经有了pgc season grpc，遂不再使用单独的dao）
func (s *Service) ogvSeasonCards(ctx context.Context, seasonIds []int32) (map[int32]*seasongrpc.CardInfoProto, error) {
	reply, err := s.seasonGRPC.Cards(ctx, &seasongrpc.SeasonInfoReq{SeasonIds: seasonIds})
	if err != nil {
		err = errors.Wrapf(err, "%v", seasonIds)
		log.Error("【OgvSeasonCards】get cards error: (%v)", err)
		return nil, err
	}
	return reply.GetCards(), nil
}
