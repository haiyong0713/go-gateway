package timemachine

import (
	"context"
	"fmt"
	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	artapimodel "git.bilibili.co/bapis/bapis-go/article/model"
	artapi "git.bilibili.co/bapis/bapis-go/article/service"
	dataapi "git.bilibili.co/bapis/bapis-go/crm/service/datamart"
	api "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	"git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"go-common/library/cache/memcache"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/timemachine"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTypeID = 999
	filterMcKey   = "u_y_r_f_d_2020"
)

var (
	tagInfo          = make(map[string]*timemachine.UserReport2020TagInfo)
	typeInfo         = make(map[int64]*timemachine.UserReport2020TypeInfo)
	pTypeInfo        = make(map[int64]*timemachine.UserReport2020TypeInfo)
	filterVideoMap   = make(map[int64]struct{})
	filterCardMap    = make(map[int32]struct{})
	filterArticleMap = make(map[int64]struct{})
)

func (s *Service) LoadUserReportBaseData() {
	// 加载tag数据
	if res, err := s.dao.RawUserReport2020TagInfo(context.Background()); err != nil {
		log.Error("s.dao.RawUserReport2020TagInfo err[%v]", err)
	} else {
		tmp := make(map[string]*timemachine.UserReport2020TagInfo)
		for _, one := range res {
			if one.Display == "" {
				one.Display = one.TagName
			}
			tmp[one.TagName] = one
		}
		tagInfo = tmp
	}
	// 加载分区数据
	if res, err := s.dao.RawUserReport2020TypeInfo(context.Background()); err != nil {
		log.Error("s.dao.RawUserReport2020TypeInfo err[%v]", err)
	} else {
		tmp1 := make(map[int64]*timemachine.UserReport2020TypeInfo)
		tmp2 := make(map[int64]*timemachine.UserReport2020TypeInfo)
		for _, one := range res {
			if one.Display == "" {
				one.Display = one.SubTidName
			}
			tmp1[one.Tid] = one
			tmp2[one.Pid] = one
		}
		typeInfo = tmp1
		pTypeInfo = tmp2
	}
	// 加载过滤数据
	s.loadUserReportFilterData()
}

func (s *Service) loadUserReportFilterData() {
	var f struct {
		Aids, Arts []int64
		Cards      []int32
	}
	var ctx = context.Background()
	if err := component.GlobalMC.Get(ctx, filterMcKey).Scan(&f); err != nil {
		return
	}
	tmpM := make(map[int64]struct{})
	for _, id := range f.Aids {
		tmpM[id] = struct{}{}
	}
	filterVideoMap = tmpM
	tmpM = make(map[int64]struct{})
	for _, id := range f.Arts {
		tmpM[id] = struct{}{}
	}
	filterArticleMap = tmpM
	tmpM1 := make(map[int32]struct{})
	for _, id := range f.Cards {
		tmpM1[id] = struct{}{}
	}
	filterCardMap = tmpM1
}

func (s *Service) Publish2020(c context.Context, loginMid, mid int64, aid int64) (*timemachine.UserInfo, error) {
	if mid > 0 && mid != loginMid {
		if _, ok := s.tmMidMap[loginMid]; !ok {
			return nil, ecode.AccessDenied
		}
	}
	if mid == 0 {
		mid = loginMid
	}
	userinfo, err := s.dao.UserInfoByMid(c, mid)
	if err != nil {
		log.Errorc(c, "Publish2020 s.dao.UserInfoByMid(c, %d) err[%v]", mid, err)
		return nil, err
	}
	//if userinfo != nil && userinfo.Aid > 0 {
	//	已经发布过
	//return userinfo, nil
	//}
	if userinfo == nil {
		userinfo = &timemachine.UserInfo{
			Mid: mid,
			Aid: aid,
		}
		if time.Now().Before(s.c.Timemachine.EndLottery) {
			userinfo.LotteryID = s.c.Timemachine.LotteryOld
		}
		err = s.dao.InsertUserInfo(c, userinfo)
		if err != nil {
			log.Errorc(c, "Publish2020 s.dao.InsertUserInfo(c, %d, %d) err[%v]", mid, aid, err)
			return nil, err
		}
	} else {
		userinfo.Aid = aid
		if userinfo.LotteryID == "" && time.Now().Before(s.c.Timemachine.EndLottery) {
			userinfo.LotteryID = s.c.Timemachine.LotteryNew
			if userinfo.IsNew != 1 || time.Now().Sub(userinfo.Mtime.Time()).Minutes() > 10 {
				// 如首次投稿提交失败，超过10分钟还未提交成功，按照非首次用户判断
				userinfo.LotteryID = s.c.Timemachine.LotteryOld
			}
		}
		err = s.dao.UpdateUserInfo(c, userinfo)
		if err != nil {
			log.Errorc(c, "Publish2020 s.dao.UpdateUserInfo(c, %v) err[%v]", *userinfo, err)
			return nil, err
		}
	}
	err = s.dao.DelCacheUserInfoByMid(c, mid)
	if err != nil {
		err = s.dao.DelCacheUserInfoByMid(c, mid)
		log.Errorc(c, "Publish2020 s.dao.DelCacheUserInfoByMid(c, %d) err[%v]", mid, err)
		return nil, err
	}
	return userinfo, nil
}

func (s *Service) BeforePublish2020(c context.Context, loginMid, mid int64) error {
	if mid > 0 && mid != loginMid {
		if _, ok := s.tmMidMap[loginMid]; !ok {
			return ecode.AccessDenied
		}
	}
	if mid == 0 {
		mid = loginMid
	}
	userinfo, err := s.dao.UserInfoByMid(c, mid)
	if err != nil {
		log.Errorc(c, "BeforePublish2020 s.dao.UserInfoByMid(c, %d) err[%v]", mid, err)
		return err
	}
	// 重复检查
	if userinfo != nil && (userinfo.Aid > 0 || userinfo.IsNew > 0) {
		return nil
	}
	// 调用投稿接口
	reply, err := client.DataMartClient.UpArchiveCount(c, &dataapi.UpArchiveCountReq{
		Mid: mid,
	})
	if err != nil {
		log.Errorc(c, "BeforePublish2020 client.DataMartClient.UpArchiveCount(c, %d) err[%v]", mid, err)
		return err
	}
	var isNew int8
	if reply.Count > 0 {
		isNew = 2
	} else {
		isNew = 1
	}
	// 初始化记录
	if userinfo == nil || userinfo.Mid == 0 {
		userinfo = &timemachine.UserInfo{
			Mid:   mid,
			IsNew: isNew,
		}
		err = s.dao.InsertUserInfo(c, userinfo)
		if err != nil {
			log.Errorc(c, "BeforePublish2020 s.dao.InsertUserInfo(c, %v, %v) err[%v]", mid, isNew, err)
			return err
		}
	} else {
		userinfo.IsNew = isNew
		err = s.dao.UpdateUserInfo(c, userinfo)
		if err != nil {
			log.Errorc(c, "BeforePublish2020 s.dao.UpdateUserInfo(c, %v) err[%v]", *userinfo, err)
			return err
		}
	}
	err = s.dao.DelCacheUserInfoByMid(c, mid)
	if err != nil {
		err = s.dao.DelCacheUserInfoByMid(c, mid)
		log.Errorc(c, "BeforePublish2020 s.dao.DelCacheUserInfoByMid(c, %d) err[%v]", mid, err)
	}
	return nil
}

func (s *Service) UserReport2020Filter(c context.Context, aids, arts []int64, cards []int32, cover bool) (interface{}, error) {
	var f struct {
		Aids, Arts []int64
		Cards      []int32
	}
	var ctx = context.Background()
	if err := component.GlobalMC.Get(ctx, filterMcKey).Scan(&f); err != nil && err != memcache.ErrNotFound {
		return nil, err
	}
	if cover {
		if len(aids) > 0 {
			f.Aids = aids
		}
		if len(arts) > 0 {
			f.Arts = arts
		}
		if len(cards) > 0 {
			f.Cards = cards
		}
	} else {
		if len(aids) > 0 {
			m := make(map[int64]struct{})
			for _, id := range f.Aids {
				m[id] = struct{}{}
			}
			for _, id := range aids {
				if _, ok := m[id]; !ok {
					m[id] = struct{}{}
					f.Aids = append(f.Aids, id)
				}
			}
		}
		if len(arts) > 0 {
			m := make(map[int64]struct{})
			for _, id := range f.Arts {
				m[id] = struct{}{}
			}
			for _, id := range arts {
				if _, ok := m[id]; !ok {
					m[id] = struct{}{}
					f.Arts = append(f.Arts, id)
				}
			}
		}
		if len(cards) > 0 {
			m := make(map[int32]struct{})
			for _, id := range f.Cards {
				m[id] = struct{}{}
			}
			for _, id := range cards {
				if _, ok := m[id]; !ok {
					m[id] = struct{}{}
					f.Cards = append(f.Cards, id)
				}
			}
		}
	}
	return f, component.GlobalMC.Set(ctx, &memcache.Item{
		Key:        filterMcKey,
		Object:     &f,
		Expiration: 2592000,
		Flags:      memcache.FlagJSON,
	})
}

func (s *Service) UserReport2020Cache(c context.Context, mid int64) (*timemachine.ResUserReport2020, error) {
	res, err := s.dao.RawUserYearReport2020(c, mid)
	if err != nil && res != nil && res.Mid > 0 {
		return nil, err
	}
	return nil, s.dao.AddCacheUserYearReport2020(c, mid, res)
}

func (s *Service) UserReport2020(c context.Context, loginMid, mid int64) (*timemachine.ResUserReport2020, error) {
	if mid > 0 && mid != loginMid {
		if _, ok := s.tmMidMap[loginMid]; !ok {
			return nil, ecode.AccessDenied
		}
	}
	if mid == 0 {
		mid = loginMid
	}
	// 加载离线计算数据
	report, err := s.dao.UserYearReport2020(c, mid)
	if err != nil {
		log.Errorc(c, "UserReport2020 s.dao.UserYearReport2020(c, %d) err[%v]", mid, err)
		return nil, err
	}
	var userProfile *accapi.ProfileReply
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) (err error) {
		userProfile, err = client.AccountClient.Profile3(c, &accapi.MidReq{Mid: mid})
		if err != nil {
			log.Errorc(c, "UserReport2020 client.AccountClient.Profile3(%v) error(%v)", mid, err)
		}
		return
	})
	if report == nil || report.PlayMinutes == 0 {
		if err := group.Wait(); err != nil {
			return nil, err
		}
		return &timemachine.ResUserReport2020{
			User: &timemachine.ResUserReport2020VideoInfo{
				Mid:      mid,
				Nickname: userProfile.Profile.Name,
				Face:     userProfile.Profile.Face,
			},
			Identification: userProfile.Profile.Identification == 1,
		}, ecode.NothingFound
	}
	userinfo, err := s.dao.UserInfoByMid(c, mid)
	if err != nil {
		log.Errorc(c, "UserReport2020 s.dao.UserInfoByMid(c, %d) err[%v]", mid, err)
		return nil, err
	}
	if userinfo == nil {
		userinfo = &timemachine.UserInfo{}
	}
	res := &timemachine.ResUserReport2020{
		VisitDays:   report.VisitDays,
		PlayVideos:  report.PlayVideos,
		PlayMinutes: report.PlayMinutes,
		HourVisitDays: []*timemachine.ResUserReport2020HourVisitDays{
			{
				Name: "深夜",
				Desc: "今年的熬夜冠军就是你",
			},
			{
				Name: "清晨",
				Desc: "每天起床第一件事\n先给喜欢的UP主打个气",
			},
			{
				Name: "上午",
				Desc: "打开小电视就能开启元气满满的一天",
			},
			{
				Name: "中午",
				Desc: "香喷喷的饭菜\n就该配上香喷喷的下饭视频",
			},
			{
				Name: "午后",
				Desc: "和小电视一起喝下午茶\n发现闪闪发光的宝藏视频吧",
			},
			{
				Name: "晚上",
				Desc: "卸下一天的疲惫\n宝藏视频陪你度过睡前时光",
			},
		},
		FavType:           report.FavType,
		FavTag:            report.FavTag,
		Top6TidScore:      make([]*timemachine.ResUserReport2020Top6TidScore, 0, 6),
		IsShowP4:          report.IsShowP4,
		LatestPlayTime:    report.LatestPlayTime,
		IsShowP5:          report.IsShowP5,
		LongestPlayDay:    report.LongestPlayDay,
		LongestPlayHours:  report.LongestPlayHours,
		LongestPlayTag:    make([]string, 0, 3),
		IsShowP6:          report.IsShowP6,
		MaxVv:             report.MaxVv,
		IsShowP7:          report.IsShowP7,
		SumLike:           report.SumLike,
		SumCoin:           report.SumCoin,
		SumFav:            report.SumFav,
		CoinTime:          report.CoinTime,
		CoinUsers:         report.CoinUsers,
		IsShowP8:          report.IsShowP8,
		RecommendVideo:    make([]*timemachine.ResUserReport2020VideoInfo, 0, 3),
		IsShowP9:          report.IsShowP9,
		FavUpType:         report.FavUpType,
		FavUpVv:           report.FavUpVv,
		IsShowP10:         report.IsShowP10,
		CreateAvs:         report.CreateAvs,
		CreateReads:       report.CreateReads,
		AvVv:              report.AvVv,
		ReadVv:            report.ReadVv,
		BestCreateType:    report.BestCreateType,
		IsShowP11:         report.IsShowP11,
		PlayComic:         report.PlayComic,
		PlayMovie:         report.PlayMovie,
		PlayDrama:         report.PlayDrama,
		PlayDocumentary:   report.PlayDocumentary,
		PlayVariety:       report.PlayVariety,
		FavSeasonID:       report.FavSeasonID,
		FavSeasonType:     report.FavSeasonType,
		IsShowP12:         report.IsShowP12,
		VipDays:           report.VipDays,
		VipAvCount:        report.VipAvCount,
		VipAvPlay:         report.VipAvPlay,
		IsShowP13:         report.IsShowP13,
		LiveHours:         report.LiveHours,
		LiveBeyondPercent: report.LiveBeyondPercent,
		FavLiveUpPlay:     report.FavLiveUpPlay,
		Ctime:             report.Ctime,
		Mtime:             report.Mtime,
		LotteryEnd:        time.Now().After(s.c.Timemachine.EndLottery),
		AID:               userinfo.Aid,
		LotteryID:         userinfo.LotteryID,
	}

	// 加载视频信息，加载用户信息
	aids := make([]int64, 0, 10)
	arts := make([]int64, 0, 10)
	recommendAids := make([]int64, 0, 6)

	// 处理p1用户访问时长描述文案
	switch {
	case res.PlayMinutes <= 300:
		res.PlayDesc = "2021年，B站欢迎你常来呀！"
	case 300 < res.PlayMinutes && res.PlayMinutes <= 900:
		res.PlayDesc = "这里还有很多宝藏视频等你发现！"
	case 900 < res.PlayMinutes && res.PlayMinutes <= 6000:
		res.PlayDesc = "刷B站爽，\n一直刷B站一直爽"
	case 6000 < res.PlayMinutes && res.PlayMinutes <= 30000:
		res.PlayDesc = "这么多的宝藏视频，\n根本停不下来啊！"
	default:
		res.PlayDesc = "原来你就是那个宇宙最强B站星人！"
	}
	// 处理p2高频消费时段数据
	for _, oneVisitDay := range strings.Split(report.HourVisitDays, ",") {
		var index, days int64
		_, err := fmt.Sscanf(oneVisitDay, "%d:%d", &index, &days)
		if err != nil {
			log.Errorc(c, "UserReport2020 fmt.Sscanf(%s) err[%v]", oneVisitDay, err)
			continue
		}
		if index > 0 {
			res.HourVisitDays[index-1].VisitDays = days
		}
	}
	var maxVisitDay int64 = -1
	for i := 0; i < 6; i++ {
		if maxVisitDay < res.HourVisitDays[i].VisitDays {
			maxVisitDay = res.HourVisitDays[i].VisitDays
			res.VisitDesc = res.HourVisitDays[i].Desc
			res.FrequentlyTime = res.HourVisitDays[i].Name
		}
	}
	// P3 分区/tag内容消费倾向 tag/分区描述&链接处理
	switch report.FavType {
	case 1:
		if tag, ok := tagInfo[res.FavTag]; ok {
			res.FavTag = tag.Display
			res.FavTagDesc = tag.Description
			res.FavTagPic = tag.Img
		}
	case 2:
		tid, _ := strconv.ParseInt(res.FavTag, 10, 64)
		if tInfo, ok := typeInfo[tid]; ok {
			res.FavTag = tInfo.Display
			res.FavTagDesc = tInfo.Description
			res.FavTagPic = tInfo.Img
		}
	}

	// P3 分区/tag内容消费倾向 分区战斗力雷达图
	for _, one := range strings.Split(report.Top6TidScore, ",") {
		var tid, score int64
		_, err := fmt.Sscanf(one, "%d:%d", &tid, &score)
		if err != nil {
			log.Errorc(c, "UserReport2020 fmt.Sscanf(%s) err[%v]", one, err)
			continue
		}
		if tid == 999 {
			// 按照产品诉求，999需要特殊case，直接过滤
			continue
		}
		tInfo := &timemachine.ResUserReport2020Top6TidScore{
			Tid:   tid,
			Score: score,
		}
		if t, ok := pTypeInfo[tid]; ok {
			tInfo.TName = t.TidName
			res.Top6TidScore = append(res.Top6TidScore, tInfo)
		}
	}
	appendTypes := []struct {
		Tid   int64
		TName string
	}{
		{
			TName: "生活",
		},
		{
			TName: "游戏",
		},
		{
			TName: "娱乐",
		},
		{
			TName: "影视",
		},
		{
			TName: "动画",
		},
		{
			TName: "知识",
		},
	}
	var j int
	for i := len(res.Top6TidScore); i < 6; i++ {
		res.Top6TidScore = append(res.Top6TidScore, &timemachine.ResUserReport2020Top6TidScore{
			Tid:   appendTypes[j].Tid,
			TName: appendTypes[j].TName,
			Score: 1,
		})
		j++
	}

	// 爆肝看视频
	if report.IsShowP4 == 1 {
		if !s.preCheckVideoID(report.LatestPlayAvid) {
			res.IsShowP4 = 0
		} else {
			aids = append(aids, report.LatestPlayAvid)
			res.LatestPlayVideo = &timemachine.ResUserReport2020VideoInfo{
				Oid: report.LatestPlayAvid,
			}
			s.highlightDecode(c, report.LatestPlayHighlight, res.LatestPlayVideo)
		}
	}
	// 单日最长播放时长
	if report.IsShowP5 == 1 {
		for _, one := range strings.Split(report.LongestPlayTag, ",") {
			var tag string
			var rank int64
			_, err := fmt.Sscanf(one, "%d:%s", &rank, &tag)
			if err != nil {
				log.Errorc(c, "UserReport2020 fmt.Sscanf(%s) err[%v]", one, err)
				continue
			}
			if t, ok := tagInfo[tag]; ok {
				find := false
				for _, tmp := range res.LongestPlayTag {
					if tmp == t.Display {
						find = true
					}
				}
				if find {
					continue
				}
				res.LongestPlayTag = append(res.LongestPlayTag, t.Display)
				if res.LongestPlayTagDesc == "" {
					res.LongestPlayTagDesc = t.Description
					res.LongestPlayTagImg = t.Img
				}
			}
		}
		if report.LongestPlaySubtid != -1 {
			if t, ok := typeInfo[report.LongestPlaySubtid]; ok {
				find := false
				for _, tmp := range res.LongestPlayTag {
					if tmp == t.Display {
						find = true
					}
				}
				if !find {
					res.LongestPlayTag = append(res.LongestPlayTag, t.Display)
					if res.LongestPlayTagDesc == "" {
						res.LongestPlayTagDesc = t.Description
						res.LongestPlayTagImg = t.Img
					}
				}
			}
		}
		if len(res.LongestPlayTag) == 0 {
			res.LongestPlayTag = append(res.LongestPlayTag, typeInfo[defaultTypeID].Display)
			res.LongestPlayTagDesc = typeInfo[defaultTypeID].Description
			res.LongestPlayTagImg = typeInfo[defaultTypeID].Img
		}
		switch {
		case report.LongestPlayHours <= 5:
			res.LongestPlayDesc = "每天逛B站多一点\n每天发现新的宝藏视频"
		case 5 < report.LongestPlayHours && report.LongestPlayHours <= 10:
			res.LongestPlayDesc = "这一天对你或许很特别呢"
		case 10 < report.LongestPlayHours && report.LongestPlayHours <= 14:
			res.LongestPlayDesc = "住在B站了吗？"
		default:
			res.LongestPlayDesc = "眼睛一闭一睁\n一天就在B站里过了"
		}
	}

	// 最常循环视频
	if report.IsShowP6 == 1 {
		if !s.preCheckVideoID(report.MaxVvAvid) {
			res.IsShowP6 = 0
		} else {
			aids = append(aids, report.MaxVvAvid)
			res.MaxVvVideo = &timemachine.ResUserReport2020VideoInfo{
				Oid: report.MaxVvAvid,
			}
			s.highlightDecode(c, report.MaxVvHighlight, res.MaxVvVideo)
			switch {
			case report.MaxVv < 50:
				res.MaxVvDesc = "还要再看一遍吗？"
			case 50 <= report.MaxVv && report.MaxVv < 365:
				res.MaxVvDesc = "开头见"
			case 365 <= report.MaxVv && report.MaxVv < 999:
				res.MaxVvDesc = "每天至少一遍！"
			default:
				res.MaxVvDesc = "“再来亿次”不是说说而已"
			}
		}
	}
	// 三连行为
	if report.IsShowP7 == 1 {
		if !s.preCheckVideoID(report.CoinAvid) {
			res.IsShowP7 = 0
		} else {
			aids = append(aids, report.CoinAvid)
			res.CoinVideo = &timemachine.ResUserReport2020VideoInfo{
				Oid: report.CoinAvid,
			}
			s.highlightDecode(c, report.CoinHighlight, res.CoinVideo)
		}
	}
	// 你可能还想推荐
	if report.IsShowP8 == 1 {
		for _, one := range strings.Split(report.RecommandAvid, ",") {
			var index, aid int64
			_, err := fmt.Sscanf(one, "%d:%d", &index, &aid)
			if err != nil {
				log.Errorc(c, "UserReport2020 fmt.Sscanf(%s) err[%v]", one, err)
				continue
			}
			if !s.preCheckVideoID(aid) {
				continue
			}
			find := false
			// 去重处理
			for _, id := range aids {
				if id == aid {
					find = true
					break
				}
			}
			if find {
				continue
			}
			aids = append(aids, aid)
			recommendAids = append(recommendAids, aid)
		}
		if len(recommendAids) == 0 {
			res.IsShowP8 = 0
		}
	}

	// 年度最喜欢的UP主
	if report.IsShowP9 == 1 {
		if (report.FavUpType == 0 && !s.preCheckVideoID(report.FavUpOid)) ||
			(report.FavUpType == 1 && !s.preCheckArticleID(report.FavUpOid)) {
			res.IsShowP9 = 0
		} else {
			res.FavUpInfo = &timemachine.ResUserReport2020VideoInfo{
				Oid: report.FavUpOid,
			}
			if report.FavUpType == 0 {
				aids = append(aids, report.FavUpOid)
				s.highlightDecode(c, report.FavUpHighlight, res.FavUpInfo)
			} else {
				arts = append(arts, report.FavUpOid)
			}
		}
	}

	// 创作侧
	if report.IsShowP10 == 1 {
		if (report.BestCreateType == 0 && !s.preCheckVideoID(report.BestCreate)) ||
			(report.BestCreateType == 1 && !s.preCheckArticleID(report.BestCreate)) {
			res.IsShowP10 = 0
		} else {
			if report.BestCreateType == 0 {
				aids = append(aids, report.BestCreate)
			} else {
				arts = append(arts, report.BestCreate)
			}
			res.BestCreateInfo = &timemachine.ResUserReport2020VideoInfo{
				Oid: report.BestCreate,
			}
		}
	}

	if report.IsShowP11 == 1 && !s.preCheckCardID(report.FavSeasonID) {
		res.IsShowP11 = 0
	}

	// 大会员
	if report.IsShowP12 == 1 {
		switch {
		case report.VipAvPlay > 120:
			res.VipDesc = "恭喜哔哩哔哩无限矿业公司\n新股东诞生！"
		case 1 <= report.VipAvPlay && report.VipAvPlay <= 120:
			res.VipDesc = "未来的时间\n让大会员接着陪伴你吧"
		default:
			res.VipDesc = "好多宝藏还等着你去发掘哦"
		}
	}

	if report.IsShowP13 == 1 && !s.preCheckVideoID(report.FavLiveUp) {
		res.IsShowP13 = 0
	}

	// 视频 & 用户数据 & 番剧信息拉取
	var arcs *arcapi.ArcsReply
	var aMetas *artapi.ArticleMetasReply
	var cards *api.CardsInfoReply
	var liveProfile *accapi.ProfileReply
	var favProfile *accapi.ProfileReply
	var videoInfo *service.ArchiveStateReply
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			arcs, err = client.ArchiveClient.Arcs(ctx, &arcapi.ArcsRequest{Aids: aids})
			if err != nil {
				log.Errorc(c, "UserReport2020 client.ArchiveClient.Arcs(%v) error(%v)", aids, err)
			}
			return
		})
	}
	// 直播消费
	if res.IsShowP13 == 1 {
		group.Go(func(ctx context.Context) (err error) {
			liveProfile, err = client.AccountClient.Profile3(c, &accapi.MidReq{Mid: report.FavLiveUp})
			if err != nil {
				log.Errorc(c, "UserReport2020 client.AccountClient.Profile3(%v) error(%v)", report.FavLiveUp, err)
			}
			return nil
		})
	}
	// 年度最喜欢的UP主 封禁数据获取
	if res.IsShowP9 == 1 {
		group.Go(func(ctx context.Context) (err error) {
			favProfile, err = client.AccountClient.Profile3(c, &accapi.MidReq{Mid: report.FavUp})
			if err != nil {
				log.Errorc(c, "UserReport2020 client.AccountClient.Profile3(%v) error(%v)", report.FavUp, err)
			}
			return nil
		})
	}
	if len(arts) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			aMetas, err = client.ArticleClient.ArticleMetas(ctx, &artapi.ArticleMetasReq{Ids: arts})
			if err != nil {
				log.Errorc(c, "UserReport2020 s.artClient.ArticleMetas(%v) error(%v)", arts, err)
			}
			return nil
		})
	}
	if res.IsShowP11 == 1 {
		group.Go(func(ctx context.Context) (err error) {
			cards, err = client.SeasonClient.Cards(ctx, &api.SeasonInfoReq{SeasonIds: []int32{report.FavSeasonID}})
			if err != nil {
				log.Errorc(c, "UserReport2020 s.seasonClient.Cards(%v) error(%v)", report.FavSeasonID, err)
			}
			return nil
		})
	}

	if userinfo.Aid > 0 {
		// 请求稿件侧获取稿件状态和最后时间
		group.Go(func(ctx context.Context) (err error) {
			videoInfo, err = client.VideoClient.ArchiveState(ctx, &service.ArchiveStateReq{
				Aid: userinfo.Aid,
			})
			if err != nil {
				log.Errorc(c, "UserReport2020 client.VideoClient.ArchiveState(%v) error(%v)", userinfo.Aid, err)
			}
			return nil
		})
	}

	if err = group.Wait(); err != nil {
		log.Errorc(c, "UserReport2020 mid(%d) group.Wait error(%v)", mid, err)
		return nil, err
	}

	// 处理投稿数据
	if userinfo.Aid > 0 {
		if videoInfo != nil {
			if videoInfo.State != -16 && videoInfo.State != -100 {
				res.PublishStatus = 2
			} else if time.Now().Unix()-videoInfo.Ctime <= s.c.Timemachine.PublishTimeout {
				res.PublishStatus = 1
			}
		} else {
			res.PublishStatus = 2
		}
		if res.PublishStatus == 0 || res.PublishStatus == 1 {
			res.AID = 0
			res.LotteryID = ""
		}
	}

	// 处理p1用户昵称
	res.User = &timemachine.ResUserReport2020VideoInfo{
		Mid:      mid,
		Nickname: userProfile.Profile.Name,
		Face:     userProfile.Profile.Face,
	}
	res.Identification = userProfile.Profile.Identification == 1
	res.Silence = userProfile.Profile.Silence == 1
	res.TelStatus = userProfile.Profile.TelStatus == 1

	// 视频 & 用户数据 & 番剧信息合并
	if res.IsShowP13 == 1 {
		if liveProfile != nil && liveProfile.Profile != nil {
			if liveProfile.Profile.Silence == 1 {
				res.IsShowP13 = 0
			} else {
				res.FavLiveUp = &timemachine.ResUserReport2020VideoInfo{
					Mid:      report.FavLiveUp,
					Nickname: liveProfile.Profile.Name,
					Face:     liveProfile.Profile.Face,
				}
			}
		} else {
			res.IsShowP13 = 0
		}
	}

	// 最喜欢的up主封禁处理
	if res.IsShowP9 == 1 {
		if favProfile == nil || favProfile.Profile == nil || favProfile.Profile.Silence == 1 {
			res.IsShowP9 = 0
		}
	}

	if arcs != nil && len(arcs.Arcs) > 0 {
		// 处理p4爆肝看视频信息
		res.IsShowP4, res.LatestPlayVideo = s.checkAndUpdateVideoInfo(res.IsShowP4, res.LatestPlayVideo, arcs.Arcs)

		// 处理p6最常循环视频信息
		res.IsShowP6, res.MaxVvVideo = s.checkAndUpdateVideoInfo(res.IsShowP6, res.MaxVvVideo, arcs.Arcs)

		// 处理p7三连行为视频信息
		res.IsShowP7, res.CoinVideo = s.checkAndUpdateVideoInfo(res.IsShowP7, res.CoinVideo, arcs.Arcs)

		// 处理p8你可能还想推荐视频信息
		if res.IsShowP8 == 1 {
			for _, aid := range recommendAids {
				tmp, obj := s.checkAndUpdateVideoInfo(1, &timemachine.ResUserReport2020VideoInfo{
					Oid: aid,
				}, arcs.Arcs)
				if tmp == 1 {
					res.RecommendVideo = append(res.RecommendVideo, obj)
					if len(res.RecommendVideo) >= 3 {
						break
					}
				}
			}
		}

		// 年度最喜欢的UP主
		if report.FavUpType == 0 {
			res.IsShowP9, res.FavUpInfo = s.checkAndUpdateVideoInfo(res.IsShowP9, res.FavUpInfo, arcs.Arcs)
		}
		// 创作侧
		if report.BestCreateType == 0 {
			res.IsShowP10, res.BestCreateInfo = s.checkAndUpdateVideoInfo(res.IsShowP10, res.BestCreateInfo, arcs.Arcs)
		}
	}

	if aMetas != nil && len(aMetas.Res) > 0 {
		// 年度最喜欢的UP主
		if report.FavUpType == 1 {
			res.IsShowP9, res.FavUpInfo = s.checkAndUpdateArticleInfo(res.IsShowP9, res.FavUpInfo, aMetas.Res)
		}
		// 创作侧
		if report.BestCreateType == 1 {
			res.IsShowP10, res.BestCreateInfo = s.checkAndUpdateArticleInfo(res.IsShowP10, res.BestCreateInfo, aMetas.Res)
		}
	} else {
		if report.FavUpType == 1 && res.IsShowP9 == 1 {
			res.IsShowP9 = 0
		}
		if report.BestCreateType == 1 && res.IsShowP10 == 1 {
			res.IsShowP10 = 0
		}
	}

	if cards != nil && len(cards.Cards) > 0 {
		// OGV消费
		if res.IsShowP11 == 1 {
			if card, ok := cards.Cards[report.FavSeasonID]; ok {
				res.FavSeasonInfo = &timemachine.ResUserReport2020VideoInfo{
					Oid:   report.BestCreate,
					Title: card.Title,
				}
				if card.NewEp != nil && card.NewEp.Cover != "" {
					res.FavSeasonInfo.Pic = card.NewEp.Cover
				} else {
					res.FavSeasonInfo.Pic = card.Cover
				}
			} else {
				res.IsShowP11 = 0
			}
		}
	} else {
		res.IsShowP11 = 0
	}

	return res, nil
}

func (s *Service) preCheckVideoID(aid int64) bool {
	if aid <= 0 {
		return false
	}
	if _, ok := filterVideoMap[aid]; ok {
		return false
	}
	return true
}

func (s *Service) preCheckArticleID(id int64) bool {
	if id <= 0 {
		return false
	}
	if _, ok := filterArticleMap[id]; ok {
		return false
	}
	return true
}

func (s *Service) preCheckCardID(id int32) bool {
	if id <= 0 {
		return false
	}
	if _, ok := filterCardMap[id]; ok {
		return false
	}
	return true
}

func (s *Service) checkAndUpdateVideoInfo(pageSwitch int64, obj *timemachine.ResUserReport2020VideoInfo, arcs map[int64]*arcapi.Arc) (int64, *timemachine.ResUserReport2020VideoInfo) {
	if pageSwitch == 1 {
		if arc, ok := arcs[obj.Oid]; ok {
			if arc.IsNormal() {
				obj.Mid = arc.Author.Mid
				obj.Nickname = arc.Author.Name
				obj.Face = arc.Author.Face
				obj.Title = arc.Title
				obj.Pic = arc.Pic
			} else {
				pageSwitch = 0
				obj.Oid = 0
				obj.PopularEnd = 0
				obj.PopularStart = 0
				obj.Cid = 0
			}
		} else {
			pageSwitch = 0
			obj.Oid = 0
			obj.PopularEnd = 0
			obj.PopularStart = 0
			obj.Cid = 0
		}
	}
	return pageSwitch, obj
}

func (s *Service) checkAndUpdateArticleInfo(pageSwitch int64, obj *timemachine.ResUserReport2020VideoInfo, arts map[int64]*artapimodel.Meta) (int64, *timemachine.ResUserReport2020VideoInfo) {
	if pageSwitch == 1 {
		if art, ok := arts[obj.Oid]; ok {
			if art.IsNormal() {
				obj.Mid = art.Author.Mid
				obj.Nickname = art.Author.Name
				obj.Face = art.Author.Face
				obj.Title = art.Title
				if len(art.ImageURLs) > 0 {
					obj.Pic = art.ImageURLs[0]
				}
			} else {
				pageSwitch = 0
				obj.Oid = 0
			}
		} else {
			pageSwitch = 0
			obj.Oid = 0
		}
	}
	return pageSwitch, obj
}

func (s *Service) highlightDecode(c context.Context, highlightStr string, obj *timemachine.ResUserReport2020VideoInfo) error {
	var aid, cid, begin, end int64
	_, err := fmt.Sscanf(highlightStr, "%d,%d,%d,%d", &aid, &cid, &begin, &end)
	if err != nil {
		log.Errorc(c, "UserReport2020 fmt.Sscanf(%s) err[%v]", highlightStr, err)
		return err
	}
	obj.Cid = cid
	obj.PopularStart = begin
	obj.PopularEnd = end
	return nil
}
