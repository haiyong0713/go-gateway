package like

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	relaapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	relmdl "git.bilibili.co/bapis/bapis-go/account/service/relation"
	upapi "git.bilibili.co/bapis/bapis-go/archive/service/up"
	tagapi "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/conf"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
	"go-gateway/app/web-svr/activity/interface/model/like"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/pkg/idsafe/bvid"
	"go-main/app/account/usersuit/service/api"

	"go-common/library/sync/errgroup.v2"
)

const (
	_awardTypeSuit  = "suit"
	_retryTime      = 3
	_imgRankTypeDay = 1
	_imgRankTypeAll = 2
	_imgDftRank     = -1
	_aidPartSize    = 100
)

var _emptyArcInfo = make(map[int][]*lmdl.ArcBvInfo)

// UpSpecial .
func (s *Service) UpSpecial(c context.Context, mid int64) (res *lmdl.UpSpecial, err error) {
	var (
		rely *upapi.UpSpecialReply
	)
	if rely, err = s.upClient.UpSpecial(c, &upapi.UpSpecialReq{Mid: mid}); err != nil {
		log.Error("s.upClient.UpSpecial(%d) error(%v)", mid, err)
		return
	}
	if rely != nil && rely.UpSpecial != nil {
		res = &lmdl.UpSpecial{GroupIDs: rely.UpSpecial.GroupIDs}
	}
	return
}

// ReceiveCoupon .
func (s *Service) ReceiveCoupon(c context.Context, sid, mid int64) (err error) {
	if sid != s.c.Image.TenSid {
		err = xecode.RequestErr
		return
	}
	var hasCheck bool
	if hasCheck, err = s.dao.RsSetNX(c, couponKey(mid, s.c.Rule.TenCoupon), s.c.Rule.TenCouponExpire); err != nil || !hasCheck {
		err = ecode.ActivityHasAward
		return
	}
	err = s.bnjDao.GrantCoupon(c, mid, s.c.Rule.TenCoupon)
	if err != nil {
		log.Error("ReceiveCoupon s.bnjDao.GrantCoupon mid(%d) coupon(%s) error(%v)", mid, s.c.Rule.TenCoupon, err)
	}
	return
}

func couponKey(mid int64, coupon string) string {
	return fmt.Sprintf("ten_coupon_%d_%s", mid, coupon)
}

// SteinList .
func (s *Service) SteinList(c context.Context) (data *lmdl.SteinList) {
	data = new(lmdl.SteinList)
	if list, err := s.dao.SteinList(c); err != nil {
		log.Error("SteinList s.dao.SteinList error(%v)", err)
	} else if list != nil {
		data.AwardOne = list.AwardOne
		data.AwardTwo = list.AwardTwo
	}
	return
}

// UserMatchCheck .
func (s *Service) UserMatchCheck(c context.Context, mid int64) (sid int64) {
	var err error
	if sid, err = s.dao.UserMatchCheck(c, mid, s.c.Rule.MatchSids); err != nil {
		log.Error("UserMatchCheck s.dao.UserMatchCheck(%v)", err)
	}
	return
}

// SingleAward .
func (s *Service) SingleAward(c context.Context, mid, sid int64) (err error) {
	cfg, ok := s.awardConf[sid]
	if !ok {
		err = ecode.ActivityNoAward
		return
	}
	// checkout award time
	nowTs := time.Now().Unix()
	if nowTs < cfg.Stime {
		err = ecode.ActivityNotStart
		return
	}
	if nowTs > cfg.Etime {
		err = ecode.ActivityOverEnd
		return
	}
	if !cfg.NoRule {
		// check submit act archive use es,cache 5 min
		var (
			submitCnt int64
		)
		sids := []int64{sid}
		if cfg.ExtraSid > 0 {
			sids = append(sids, cfg.ExtraSid)
		}
		if submitCnt, err = s.dao.LikeMidTotal(c, mid, sids); err != nil {
			log.Error("SingleAward s.dao.LikeMidTotal(%d,%v) error(%v)", mid, sids, err)
			err = ecode.ActivityNotJoin
			return
		}
		if submitCnt <= 0 {
			err = ecode.ActivityNotJoin
			return
		}
	}
	switch cfg.Type {
	case _awardTypeSuit:
		var check bool
		if check, err = s.dao.RsSetNX(c, awardCheckKey(mid, sid), cfg.LimitExpire); err != nil || !check {
			err = ecode.ActivityHasAward
			return
		}
		if _, err = s.suitClient.GrantByMids(c, &api.GrantByMidsReq{Mids: []int64{mid}, Pid: cfg.AwardID, Expire: cfg.AwardExpire}); err != nil {
			err = ecode.ActivityBnjRewardFail
			// del nx key
			s.cache.Do(c, func(ctx context.Context) {
				if e := s.dao.RsDelNX(ctx, awardCheckKey(mid, sid)); e != nil {
					log.Error("SingleAward s.dao.UserMatchCheck(%v)", e)
				}
			})
		}
	}
	return
}

func (s *Service) SingleAwardState(c context.Context, mid, sid int64) (state int, err error) {
	_, ok := s.awardConf[sid]
	if !ok {
		err = ecode.ActivityNoAward
		return
	}
	var val string
	if val, err = s.dao.RsGet(c, awardCheckKey(mid, sid)); err != nil {
		log.Error("SingleAwardState s.dao.RsGet(%s) error(%v)", awardCheckKey(mid, sid), err)
		err = nil
		return
	}
	if val != "" {
		if state, err = strconv.Atoi(val); err != nil {
			log.Error("SingleAwardState strconv.Atoi(%s) error(%v)", val, err)
			err = nil
		}
	}
	return
}

func awardCheckKey(mid, sid int64) string {
	return fmt.Sprintf("award_c_k_%d_%d", mid, sid)
}

// ArchiveList
func (s *Service) ArchiveList(c context.Context, mid, sid, tid int64) (res []*like.ArcInfo, err error) {
	var (
		data       *lmdl.ArcListData
		aids, mids []int64
		arcs       *arcmdl.ArcsReply
	)
	switch sid {
	//case s.c.Scholarship.ArcVid:
	//	data = s.scholarshipArcData
	//case s.c.Eleven.ArcVid:
	//	data = s.arcListData
	//case s.c.SpringCardAct.ArcVid:
	//	data = s.springCardArcData
	//case s.c.Shad.Vid:
	//	data = s.shaDArcData
	//case s.c.Restart2020.Vid:
	//	data = s.restartArcData
	//case s.c.YellowGrean.Vid:
	//	data = s.yellowGreenArcData
	//case s.c.MobileGame.Vid:
	//	data = s.mobileGameArcData
	//case s.c.Stupid.Vid:
	//	data = s.stupidArcData
	case s.c.GameHoliday.Vid:
		data = s.gameHolidayArcData
	case s.c.S10Contribution.TotalVid:
		data = s.totalRankArcData
	case s.c.Funny.Vid:
		data = s.funnyVideoListArcData
	default:
		return
	}
	if data == nil {
		return
	}
	for _, v := range data.List {
		if v.ID == strconv.FormatInt(tid, 10) {
			for _, val := range strings.Split(v.Data.Aids, ",") {
				if strings.HasPrefix(val, "BV") {
					avid, err := bvid.BvToAv(val)
					if err != nil {
						log.Error("Failed to switch bv to av: %s %+v", val, err)
						continue
					}
					aids = append(aids, avid)
				} else {
					if avid, _ := strconv.ParseInt(val, 10, 64); avid > 0 {
						aids = append(aids, avid)
					}
				}
			}
		}
	}
	if len(aids) == 0 {
		return
	}
	if arcs, err = client.ArchiveClient.Arcs(c, &arcmdl.ArcsRequest{Aids: aids}); err != nil {
		log.Error("s.arcClient.Archives3(%v) error(%v)", aids, err)
		return
	}
	for _, arc := range arcs.Arcs {
		mids = append(mids, arc.Author.Mid)
	}
	relReq := &relmdl.RelationsReq{
		Mid: mid,
		Fid: mids,
	}
	relRsp, e := s.relClient.Relations(c, relReq)
	if e != nil {
		log.Errorc(c, "s.relClient.Relations.mids(%v) error(%v)", mids, e)
	}
	for _, aid := range aids {
		if arc, ok := arcs.Arcs[aid]; ok {
			var isFollow int64
			bvidStr, e := bvid.AvToBv(arc.Aid)
			if e != nil {
				continue
			}
			if relRsp != nil {
				if FollowInfo, ok := relRsp.FollowingMap[arc.Author.Mid]; ok && FollowInfo != nil && FollowInfo.Attribute < 128 {
					isFollow = 1
				}
			}
			HideArcAttribute(arc)
			res = append(res, &like.ArcInfo{
				Arc:      arc,
				Bvid:     bvidStr,
				IsFollow: isFollow,
			})
		}
	}
	return
}

func HideArcAttribute(arc *arcmdl.Arc) {
	arc.AttributeV2 = 0
	arc.Attribute = 0
	arc.Access = 0
}

// ArcLists
func (s *Service) ArcLists(c context.Context, sid, defaultTid, specialTid int64) (res *lmdl.ArcLists, err error) {
	var (
		data                     *lmdl.ArcListData
		defaultAids, specialAids []int64
		aids                     []int64
	)
	res = &lmdl.ArcLists{}
	switch sid {
	case s.c.S10Contribution.DayVid:
		data = s.daySelectArcData
	case s.c.DoubleEleven.VideoListVid:
		data = s.doubl11VidelArcData
	case s.c.Timemachine.Vid:
		data = s.timeMachineArcData
	default:
		return
	}
	if data == nil {
		return
	}
	for _, v := range data.List {
		if v.ID == strconv.FormatInt(defaultTid, 10) {
			for _, val := range strings.Split(v.Data.Aids, ",") {
				if strings.HasPrefix(val, "BV") {
					avid, err := bvid.BvToAv(val)
					if err != nil {
						log.Error("Failed to switch bv to av: %s %+v", val, err)
						continue
					}
					defaultAids = append(defaultAids, avid)
				} else {
					if avid, _ := strconv.ParseInt(val, 10, 64); avid > 0 {
						defaultAids = append(defaultAids, avid)
					}
				}
			}
		} else if v.ID == strconv.FormatInt(specialTid, 10) {
			for _, val := range strings.Split(v.Data.Aids, ",") {
				if strings.HasPrefix(val, "BV") {
					avid, err := bvid.BvToAv(val)
					if err != nil {
						log.Error("Failed to switch bv to av: %s %+v", val, err)
						continue
					}
					specialAids = append(specialAids, avid)
				} else {
					if avid, _ := strconv.ParseInt(val, 10, 64); avid > 0 {
						specialAids = append(specialAids, avid)
					}
				}
			}
		}
	}
	if len(defaultAids) > 0 {
		aids = append(aids, defaultAids...)
	}
	if len(specialAids) > 0 {
		aids = append(aids, specialAids...)
	}
	aidsLen := len(aids)
	if aidsLen == 0 {
		return
	}
	arcs := make(map[int64]*arcmdl.Arc, aidsLen)
	for i := 0; i < aidsLen; i += _aidPartSize {
		var partAids []int64
		if i+_aidPartSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidPartSize]
		}
		partArcs, err := client.ArchiveClient.Arcs(c, &arcmdl.ArcsRequest{Aids: partAids})
		if err != nil {
			log.Error("ArcLists s.arcClient.Arcs partAids(%v) error(%v)", partAids, err)
			continue
		}
		for _, v := range partArcs.GetArcs() {
			if v != nil && v.IsNormal() {
				HideArcAttribute(v)
				arcs[v.Aid] = v
			}
		}
	}
	for _, aid := range defaultAids {
		if arc, ok := arcs[aid]; ok {
			bvidStr, e := bvid.AvToBv(arc.Aid)
			if e != nil {
				continue
			}
			res.Default = append(res.Default, &like.ArcBvInfo{
				Arc:  arc,
				Bvid: bvidStr,
			})
		}
	}
	for _, aid := range specialAids {
		if arc, ok := arcs[aid]; ok {
			bvidStr, e := bvid.AvToBv(arc.Aid)
			if e != nil {
				continue
			}
			res.Special = append(res.Special, &like.ArcBvInfo{
				Arc:  arc,
				Bvid: bvidStr,
			})
		}
	}
	return
}

// ChannelArcs
func (s *Service) ChannelArcs(c context.Context, sid int64, tids []int) (res map[int][]*lmdl.ArcBvInfo, err error) {
	var (
		data       *lmdl.ArcListData
		aids       []int64
		tidMap     map[string]int
		tidMapAids map[int][]int64
	)
	tidCount := len(tids)
	if tidCount == 0 {
		res = _emptyArcInfo
		return
	}
	tidMap = make(map[string]int, tidCount)
	tidMapAids = make(map[int][]int64, tidCount)
	for _, tid := range tids {
		tidMap[strconv.Itoa(tid)] = tid
	}
	switch sid {
	case s.c.DoubleEleven.ChannelListID:
		data = s.doubl11ChannelArcData
	default:
		res = _emptyArcInfo
		return
	}
	if data == nil {
		res = _emptyArcInfo
		return
	}
	for _, v := range data.List {
		if tid, ok := tidMap[v.ID]; ok {
			for _, val := range strings.Split(v.Data.Aids, ",") {
				if strings.HasPrefix(val, "BV") {
					avid, err := bvid.BvToAv(val)
					if err != nil {
						log.Error("Failed to switch bv to av: %s %+v", val, err)
						continue
					}
					aids = append(aids, avid)
					tidMapAids[tid] = append(tidMapAids[tid], avid)
				} else {
					if avid, _ := strconv.ParseInt(val, 10, 64); avid > 0 {
						aids = append(aids, avid)
						tidMapAids[tid] = append(tidMapAids[tid], avid)
					}
				}
			}
		}
	}
	aidsLen := len(aids)
	if aidsLen == 0 {
		res = _emptyArcInfo
		return
	}
	arcs := make(map[int64]*arcmdl.Arc, aidsLen)
	for i := 0; i < aidsLen; i += _aidPartSize {
		var partAids []int64
		if i+_aidPartSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidPartSize]
		}
		partArcs, err := client.ArchiveClient.Arcs(c, &arcmdl.ArcsRequest{Aids: partAids})
		if err != nil {
			log.Error("ArcLists s.arcClient.Arcs partAids(%v) error(%v)", partAids, err)
			continue
		}
		for _, v := range partArcs.GetArcs() {
			if v != nil && v.IsNormal() {
				HideArcAttribute(v)
				arcs[v.Aid] = v
			}
		}
	}
	res = make(map[int][]*lmdl.ArcBvInfo, len(tidMapAids))
	for tid, TidAids := range tidMapAids {
		for _, aid := range TidAids {
			if arc, ok := arcs[aid]; ok {
				bvidStr, e := bvid.AvToBv(arc.Aid)
				if e != nil {
					continue
				}
				HideArcAttribute(arc)
				res[tid] = append(res[tid], &like.ArcBvInfo{
					Arc:  arc,
					Bvid: bvidStr,
				})
			}
		}
	}
	return
}

func (s *Service) SingleWebData(c context.Context, sid, vid int64) (data interface{}, err error) {
	if sid != s.c.Taaf.Vid && sid != s.c.Ent.Vid && sid != s.c.Ent.VidV2 {
		err = xecode.RequestErr
		return
	}
	if sid == s.c.Taaf.Vid {
		tmp, ok := s.taafData[strconv.FormatInt(vid, 10)]
		if !ok {
			err = xecode.NothingFound
			return
		}
		data = tmp
		return
	}
	if sid == s.c.Ent.Vid {
		tmp, ok := s.entData[vid]
		if !ok {
			err = xecode.NothingFound
			return
		}
		entData := &lmdl.EntRes{Lid: tmp.Lid, Mid: tmp.Mid, TagID: tmp.TagID}
		var archives map[int64]*arcmdl.Arc
		group := errgroup.WithContext(c)
		if tmp.Lid > 0 {
			group.Go(func(ctx context.Context) error {
				likeCont, e := s.dao.LikeContent(ctx, []int64{tmp.Lid})
				if e != nil {
					log.Error("s.dao.LikeContent lid(%d) error(%v)", tmp.Lid, e)
					return nil
				}
				if item, ok := likeCont[tmp.Lid]; ok && item != nil {
					entData.Content = item
				}
				return nil
			})
		}
		if aids, _ := xstr.SplitInts(tmp.Aid); len(aids) > 0 {
			group.Go(func(ctx context.Context) error {
				reply, e := client.ArchiveClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: aids})
				if e != nil {
					log.Error("s.arcClient.Arcs aids(%v) error(%v)", aids, e)
					return nil
				}
				archives = reply.Arcs
				for _, aid := range aids {
					if arc, ok := archives[aid]; ok && arc != nil && arc.IsNormal() {
						entData.Arcs = append(entData.Arcs, lmdl.CopyFromArc(arc))
					}
				}
				return nil
			})
		}
		group.Wait()
		data = entData
	}
	if sid == s.c.Ent.VidV2 {
		//tmp, ok := s.entDataV2[vid]
		//if !ok {
		err = xecode.NothingFound
		return
		//}
		//entDataV2 := &lmdl.EntResV2{Lid: tmp.Lid, Mid: tmp.Mid}
		//var archives map[int64]*arcmdl.Arc
		//group := errgroup.WithContext(c)
		//if tmp.Lid > 0 {
		//	group.Go(func(ctx context.Context) error {
		//		likeCont, e := s.dao.LikeContent(ctx, []int64{tmp.Lid})
		//		if e != nil {
		//			log.Error("s.dao.LikeContent lid(%d) error(%v)", tmp.Lid, e)
		//			return nil
		//		}
		//		if item, ok := likeCont[tmp.Lid]; ok && item != nil {
		//			entDataV2.Content = item
		//		}
		//		return nil
		//	})
		//}
		//if aids, _ := xstr.SplitInts(tmp.Aid); len(aids) > 0 {
		//	group.Go(func(ctx context.Context) error {
		//		reply, e := s.arcClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: aids})
		//		if e != nil {
		//			log.Error("s.arcClient.Arcs aids(%v) error(%v)", aids, e)
		//			return nil
		//		}
		//		archives = reply.Arcs
		//		for _, aid := range aids {
		//			if arc, ok := archives[aid]; ok && arc != nil && arc.IsNormal() {
		//				entDataV2.Arcs = append(entDataV2.Arcs, lmdl.CopyFromArc(arc))
		//			}
		//		}
		//		return nil
		//	})
		//}
		//group.Wait()
		//data = entDataV2
	}
	return
}

func (s *Service) Iir(_ context.Context, mobiapp string, build int64) bool {
	auditBuilds, ok := s.resAuditData[mobiapp]
	if !ok {
		return false
	}
	for _, auditBuild := range auditBuilds {
		if build == auditBuild {
			return true
		}
	}
	return false
}

//func (s *Service) loadArchiveList() {
//	res, err := s.dao.SourceItem(context.Background(), s.c.Eleven.ArcVid)
//	if err != nil {
//		log.Error("loadArchiveList s.dao.SourceItem(%d) error(%v)", s.c.Eleven.ArcVid, err)
//		return
//	}
//	tmp := new(like.ArcListData)
//	if err = json.Unmarshal(res, &tmp); err != nil {
//		log.Error("loadArchiveList json.Unmarshal(%s) error(%v)", res, err)
//		return
//	} else {
//		s.arcListData = tmp
//	}
//	log.Info("loadArchiveList success")
//}

//func (s *Service) springCardArcDataproc() {
//	if s.springCardArcData == nil {
//		s.springCardArcData = new(like.ArcListData)
//	}
//	res, err := s.dao.SourceItem(context.Background(), s.c.SpringCardAct.ArcVid)
//	if err != nil {
//		log.Error("springCardArcDataproc s.dao.SourceItem(%d) error(%v)", s.c.SpringCardAct.ArcVid, err)
//		return
//	}
//	tmp := new(like.ArcListData)
//	if err = json.Unmarshal(res, &tmp); err != nil {
//		log.Error("springCardArcDataproc json.Unmarshal(%s) error(%v)", res, err)
//		return
//	}
//	s.springCardArcData = tmp
//	log.Info("springCardArcDataproc success")
//}

//func (s *Service) loadShaDArcData() {
//	if s.shaDArcData == nil {
//		s.shaDArcData = new(like.ArcListData)
//	}
//	res, err := s.dao.SourceItem(context.Background(), s.c.Shad.Vid)
//	if err != nil {
//		log.Error("loadShaDArcData s.dao.SourceItem(%d) error(%v)", s.c.Shad.Vid, err)
//		return
//	}
//	tmp := new(like.ArcListData)
//	if err = json.Unmarshal(res, &tmp); err != nil {
//		log.Error("loadShaDArcData json.Unmarshal(%s) error(%v)", res, err)
//		return
//	}
//	s.shaDArcData = tmp
//	log.Info("loadShaDArcData success")
//}

//func (s *Service) loadRestartArcData() {
//	if s.restartArcData == nil {
//		s.restartArcData = new(like.ArcListData)
//	}
//	res, err := s.dao.SourceItem(context.Background(), s.c.Restart2020.Vid)
//	if err != nil {
//		log.Error("loadRestartArcData s.dao.SourceItem(%d) error(%v)", s.c.Restart2020.Vid, err)
//		return
//	}
//	tmp := new(like.ArcListData)
//	if err = json.Unmarshal(res, &tmp); err != nil {
//		log.Error("loadRestartArcData json.Unmarshal(%s) error(%v)", res, err)
//		return
//	}
//	s.restartArcData = tmp
//	log.Info("loadRestartArcData success")
//}

//func (s *Service) loadScholarshipArcData() {
//	res, err := s.dao.SourceItem(context.Background(), s.c.Scholarship.ArcVid)
//	if err != nil {
//		log.Error("loadScholarshipArcData s.dao.SourceItem(%d) error(%v)", s.c.Scholarship.ArcVid, err)
//		return
//	}
//	tmp := new(like.ArcListData)
//	if err = json.Unmarshal(res, &tmp); err != nil {
//		log.Error("loadScholarshipArcData json.Unmarshal(%s) error(%v)", res, err)
//		return
//	} else {
//		s.scholarshipArcData = tmp
//	}
//	log.Info("loadScholarshipArcData success")
//}

func (s *Service) loadTaafWebData() {
	res, err := s.dao.SourceItem(context.Background(), s.c.Taaf.Vid)
	if err != nil {
		log.Error("loadTaafWebData s.dao.SourceItem(%d) error(%v)", s.c.Taaf.Vid, err)
		return
	}
	tmp := new(like.TaafWebData)
	if err = json.Unmarshal(res, tmp); err != nil {
		log.Error("loadTaafWebData s.dao.SourceItem(%d) error(%v)", s.c.Scholarship.ArcVid, err)
		return
	}
	if len(tmp.List) == 0 {
		log.Error("loadTaafWebData data len 0")
		return
	}
	tmpData := make(map[string]*like.TaafData, len(tmp.List))
	for _, v := range tmp.List {
		if v == nil || v.Data == nil {
			continue
		}
		tmpData[v.Name] = v.Data
		if v.Data.Lidnew != "" {
			tmpData[v.Data.Lidnew] = v.Data
		}
	}
	s.taafData = tmpData
	log.Info("loadTaafWebData() success")
}

func (s *Service) loadEntWebData() {
	res, err := s.dao.SourceItem(context.Background(), s.c.Ent.Vid)
	if err != nil {
		log.Error("loadEntWebData s.dao.SourceItem(%d) error(%v)", s.c.Taaf.Vid, err)
		return
	}
	tmp := new(struct {
		List []*struct {
			ID   string        `json:"id"`
			Name string        `json:"name"`
			Data *lmdl.EntData `json:"data"`
		} `json:"list"`
	})
	if err = json.Unmarshal(res, tmp); err != nil {
		log.Error("loadEntWebData s.dao.SourceItem(%d) error(%v)", s.c.Scholarship.ArcVid, err)
		return
	}
	if len(tmp.List) == 0 {
		log.Error("loadEntWebData data len 0")
		return
	}
	tmpData := make(map[int64]*like.EntData, len(tmp.List))
	for _, v := range tmp.List {
		if v == nil || v.Data == nil {
			continue
		}
		tmpData[v.Data.Lid] = v.Data
	}
	s.entData = tmpData
	log.Info("loadEntWebData() success")
}

//func (s *Service) loadEntV2WebData() {
//	res, err := s.dao.SourceItem(context.Background(), s.c.Ent.VidV2)
//	if err != nil {
//		log.Error("loadEntV2WebData s.dao.SourceItem(%d) error(%v)", s.c.Ent.VidV2, err)
//		return
//	}
//	tmp := new(struct {
//		List []*struct {
//			ID   string          `json:"id"`
//			Name string          `json:"name"`
//			Data *lmdl.EntDataV2 `json:"data"`
//		} `json:"list"`
//	})
//	if err = json.Unmarshal(res, tmp); err != nil {
//		log.Error("loadEntV2WebData s.dao.SourceItem(%d) error(%v)", s.c.Ent.VidV2, err)
//		return
//	}
//	if len(tmp.List) == 0 {
//		log.Error("loadEntV2WebData data len 0")
//		return
//	}
//	tmpData := make(map[int64]*like.EntDataV2, len(tmp.List))
//	for _, v := range tmp.List {
//		if v == nil || v.Data == nil {
//			continue
//		}
//		tmpData[v.Data.Lid] = v.Data
//	}
//	s.entDataV2 = tmpData
//}

func (s *Service) loadTaafLikes() {
	tmp := s.loadAllLikes(s.c.Taaf.Sid)
	if len(tmp.List) == 0 {
		log.Error("loadTaafLikes len == 0")
		return
	}
	s.taafLikes = tmp
	log.Info("loadTaafLikes() success")
}

func (s *Service) loadTmLikes() {
	if s.c.Timemachine.FlagSid == 0 {
		return
	}
	tmp := s.loadAllLikes(s.c.Timemachine.FlagSid)
	if len(tmp.List) == 0 {
		log.Error("loadTaafLikes len == 0")
		return
	}
	s.timemachineLikes = tmp
	log.Info("loadTmLikes() success")
}

func (s *Service) loadAllLikes(sid int64) (res *like.ListInfo) {
	res = new(like.ListInfo)
	ctx := context.Background()
	var lids []int64
	data, err := s.dao.LikesBySid(ctx, 0, sid)
	if err != nil {
		log.Error("loadAllLikes sid(%d) error(%v)", sid, err)
		return
	}
	if len(data) == 0 {
		log.Warn("loadAllLikes load data finish")
		return
	}
	for _, v := range data {
		lids = append(lids, v.ID)
	}
	batchSize := 100
	contents := make(map[int64]*like.LikeContent, len(lids))
	// 分批获取like content 数据
	for len(lids) > 0 {
		if batchSize > len(lids) {
			batchSize = len(lids)
		}
		tmpLids := lids[:batchSize]
		lids = lids[batchSize:]
		conts, err := s.dao.RawLikeContent(ctx, tmpLids)
		if err != nil {
			log.Error("loadAllLikes sid(%d) error(%v)", s.c.Taaf.Sid, err)
			return
		}
		for _, v := range conts {
			contents[v.ID] = v
		}
		time.Sleep(100 * time.Millisecond)
	}
	for _, v := range data {
		if cont, ok := contents[v.ID]; ok && cont != nil {
			res.List = append(res.List, &like.List{
				Item:   v,
				Object: map[string]interface{}{"cont": cont},
			})
		}
	}
	return
}

func (s *Service) loadResAuditData() {
	data, err := s.dao.ResAudit(context.Background())
	if err != nil {
		log.Error("loadResAuditData s.dao.ResAudit error(%v)", err)
		return
	}
	if len(data) == 0 {
		log.Warn("loadResAuditData len(data) == 0")
		return
	}
	s.resAuditData = data
}

func (s Service) SpecialArcList(c context.Context, id, sid int64, pn, ps int) (res *lmdl.SpecialArcListReply, err error) {
	start := int64((pn - 1) * ps)
	end := start + int64(ps)
	grade, ok := s.specialArcData[id]
	if !ok || grade == nil || len(grade.Subject) == 0 {
		log.Warn("SpecialArcList subject nil")
		return
	}
	aidStr := ""
	for _, v := range grade.Subject {
		if v.ID == sid {
			aidStr = v.Aids
			break
		}
	}
	if aidStr == "" {
		log.Warn("SpecialArcList 找不到对应的sid")
		return
	}
	aids := make([]int64, 0)
	for _, val := range strings.Split(aidStr, ",") {
		trim := strings.TrimSpace(val)
		if i, _ := strconv.ParseInt(trim, 10, 64); i > 0 {
			aids = append(aids, i)
		}
	}
	total := int64(len(aids))
	if total < end {
		end = total
	}
	if start >= total {
		return
	}
	arcs := new(arcmdl.ArcsReply)
	if arcs, err = client.ArchiveClient.Arcs(c, &arcmdl.ArcsRequest{Aids: aids[start:end]}); err != nil {
		log.Error("s.arcClient.Archives3(%v) error(%v)", aids, err)
		return
	}
	list := make([]*arcmdl.Arc, 0)
	for _, aid := range aids {
		if arc, ok := arcs.Arcs[aid]; ok {
			HideArcAttribute(arc)
			list = append(list, arc)
		}
	}
	page := &lmdl.Page{Num: pn, Size: ps, Total: total}
	res = &lmdl.SpecialArcListReply{List: list, Page: page}
	return
}

func (s *Service) loadSpecialArcData() {
	urls := s.c.Image.SpecialJsonList
	tmp := make(map[int64]*lmdl.SpecialArcList)
	for k, v := range urls {
		t := time.Now().Unix()
		for i := 0; i < _retryTime; i++ {
			data, err := s.dao.SpecialData(context.Background(), v, t)
			if err == nil {
				id, e := strconv.ParseInt(k, 10, 64)
				if e != nil {
					log.Error("loadSpecialArcData strconv.ParseInt(%s) error(%v)", k, e)
					break
				}
				tmp[id] = data
				break
			}
			log.Error("loadSpecialArcData(%s,%d) error(%+v)", v, i, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	s.specialArcData = tmp
	log.Info("loadSpecialArcData success(%+v)", s.specialArcData)
}

// yellow & green act load arc
//func (s *Service) loadYellowGreenArcData() {
//	if s.yellowGreenArcData == nil {
//		s.yellowGreenArcData = new(like.ArcListData)
//	}
//	res, err := s.dao.SourceItem(context.Background(), s.c.YellowGrean.Vid)
//	if err != nil {
//		log.Error("loadYellowGreenArcData s.dao.SourceItem(%d) error(%v)", s.c.YellowGrean.Vid, err)
//		return
//	}
//	tmp := new(like.ArcListData)
//	if err = json.Unmarshal(res, &tmp); err != nil {
//		log.Error("loadYellowGreenArcData json.Unmarshal(%s) error(%v)", res, err)
//		return
//	}
//	s.yellowGreenArcData = tmp
//	log.Info("loadYellowGreenArcData success")
//}

func (s *Service) ReadDay(c context.Context, mid int64) (*lmdl.ReadDay, error) {
	var totalCount, myCount int64
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		total, err := s.dao.LikeCount(ctx, s.c.ReadDay.Sid, 0)
		if err != nil {
			log.Error("ReadDay LikeCount sid:%d error:%v", s.c.ReadDay.Sid, err)
			return nil
		}
		totalCount = total
		return nil
	})
	if mid > 0 {
		group.Go(func(ctx context.Context) error {
			stat, err := s.dao.MyListTotalStateFromEs(ctx, s.c.ReadDay.Sid, mid, 0)
			if err != nil {
				log.Error("ReadDay MyListTotalStateFromEs sid:%d mid:%d error:%v", s.c.ReadDay.Sid, mid, err)
				return nil
			}
			myCount = stat.Count
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	if s.c.ReadDay.Multi > 1 {
		totalCount = int64(float64(totalCount) * s.c.ReadDay.Multi)
	}
	return &lmdl.ReadDay{
		TotalCount: totalCount,
		Max:        s.c.ReadDay.Max,
		EndTime:    s.c.ReadDay.EndTime.Unix(),
		MyCount:    myCount,
	}, nil
}

func (s *Service) Bml20Follow(c context.Context, sid, mid int64, reSrc uint8, ck string) (err error) {
	cfg := func() *conf.Bml20 {
		if s.c.Bml20 == nil {
			return nil
		}
		for _, v := range s.c.Bml20 {
			if v != nil && v.Sid == sid {
				return v
			}
		}
		return nil
	}()
	if cfg == nil {
		err = xecode.RequestErr
		return
	}
	if _, err = s.Reserve(c, sid, mid, 1); err != nil {
		log.Error("Bml20Follow Reserve sid:%d mid:%d error(%d)", sid, mid, err)
		return
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if _, e := s.relClient.AddFollowing(ctx, &relaapi.FollowingReq{Mid: mid, Fid: cfg.Mid, Source: uint32(reSrc)}); e != nil {
			log.Error("Bml20Follow AddFollowing mid:%d fid:%d error(%v)", mid, cfg.Mid, e)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if _, e := s.relClient.AddFollowing(ctx, &relaapi.FollowingReq{Mid: mid, Fid: cfg.SnsMid, Source: uint32(reSrc)}); e != nil {
			log.Error("Bml20Follow AddFollowing sns mid:%d fid:%d error(%v)", mid, cfg.SnsMid, e)
		}
		return nil
	})
	if cfg.ShopID > 0 {
		group.Go(func(ctx context.Context) error {
			if e := s.dao.TicketAddWish(ctx, cfg.ShopID, ck); e != nil {
				log.Error("Bml20Follow TicketAddWish mid:%d ShopID:%d error(%v)", mid, cfg.ShopID, e)
			}
			return nil
		})
	}
	group.Go(func(ctx context.Context) error {
		if e := s.tagDao.AddSub(ctx, &tagapi.AddSubReq{Mid: mid, Tids: []int64{cfg.DyID}}); e != nil {
			log.Error("Bml20Follow s.tagDao.AddSub mid:%d DyID:%d error(%v)", mid, cfg.DyID, e)
		}
		return nil
	})
	group.Wait()
	return
}

func (s *Service) ImageUserRank(c context.Context, mid int64, typ int) (res *lmdl.ImgUserRank, err error) {
	var (
		rank, otherRank   int64
		score, otherScore float64
		infos             map[int64]*accapi.Info
		stats             map[int64]*relaapi.StatReply
	)
	res = &lmdl.ImgUserRank{Self: &lmdl.ImageSelf{DayRank: _imgDftRank, TotalRank: _imgDftRank}}
	now := time.Now()
	nowDay := now.Format("20060102")
	if now.Hour() < 1 && now.Minute() < 30 {
		nowDay = now.AddDate(0, 0, -1).Format("20060102")
	}
	if now.Unix() > s.c.ImageV2.Etime.Unix() {
		nowDay = s.c.ImageV2.Etime.Format("20060102")
	}
	limit := s.c.ImageV2.DayLimit
	if typ == _imgRankTypeAll {
		limit = s.c.ImageV2.AllLimit
	}
	list, err := s.dao.ImageUserRankList(c, s.c.ImageV2.Sid, nowDay, typ, limit-1)
	if err != nil {
		log.Error("ImageUserRankList s.dao.ImageUserRankList sid:%d day:%s typ:%d error(%v)", s.c.ImageV2.Sid, nowDay, typ, err)
		return
	}
	var mids []int64
	for _, v := range list {
		if v.Mid > 0 {
			mids = append(mids, v.Mid)
		}
	}
	if len(mids) == 0 {
		log.Warn("ImageUserRankList sid:%d day:%s typ:%d len(mids) == 0", s.c.ImageV2.Sid, nowDay, typ)
		return
	}
	if mid > 0 {
		mids = append(mids, mid)
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		var e error
		infos, e = s.accInfos(ctx, mids)
		if e != nil {
			return e
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		var e error
		stats, e = s.accStats(ctx, mids)
		if e != nil {
			log.Error("ImageUserRankList s.accStats mids(%v) error(%v)", mids, e)
		}
		return nil
	})
	if mid > 0 {
		group.Go(func(ctx context.Context) error {
			var e error
			rank, score, e = s.dao.ImageUserRank(ctx, s.c.ImageV2.Sid, mid, nowDay, typ)
			if e != nil {
				log.Error("ImageUserRankList sid:%d mid:%d day:%s typ:%d error(%v)", s.c.ImageV2.Sid, mid, nowDay, typ, e)
				return nil
			}
			if rank != bwsmdl.DefaultRank {
				rank = rank + 1
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			var e error
			otherTyp := 3 - typ
			otherRank, otherScore, e = s.dao.ImageUserRank(ctx, s.c.ImageV2.Sid, mid, nowDay, otherTyp)
			if e != nil {
				log.Error("ImageUserRankList sid:%d mid:%d day:%s typ:%d error(%v)", s.c.ImageV2.Sid, mid, nowDay, otherTyp, e)
				return nil
			}
			if otherRank != bwsmdl.DefaultRank {
				otherRank = otherRank + 1
			}
			return nil
		})
	}
	if err = group.Wait(); err != nil {
		log.Error("ImageUserRankList sid:%d day:%s typ:%d group.Wait error(%v)", s.c.ImageV2.Sid, nowDay, typ, err)
		return
	}
	for i, v := range list {
		imgUser := &lmdl.ImgUser{SimpleUser: &lmdl.SimpleUser{Mid: v.Mid}, ImageRank: int64(i + 1), ImageScore: v.ImageScore}
		if info, ok := infos[v.Mid]; ok && info != nil {
			imgUser.Name = info.Name
			imgUser.Face = info.Face
		}
		if stat, ok := stats[v.Mid]; ok && stat != nil {
			imgUser.Follower = stat.Follower
		}
		res.List = append(res.List, imgUser)
	}
	if mid > 0 {
		res.Self.SimpleUser = &lmdl.SimpleUser{Mid: mid}
		if info, ok := infos[mid]; ok && info != nil {
			res.Self.SimpleUser.Name = info.Name
			res.Self.SimpleUser.Face = info.Face
		}
		if typ == _imgRankTypeDay {
			res.Self.DayRank = rank
			res.Self.DayScore = score
			res.Self.TotalRank = otherRank
			res.Self.TotalScore = otherScore
		} else {
			res.Self.DayRank = otherRank
			res.Self.DayScore = otherScore
			res.Self.TotalRank = rank
			res.Self.TotalScore = score
		}
		if res.Self.DayRank > s.c.ImageV2.DayLimit {
			res.Self.DayRank = int64(bwsmdl.DefaultRank)
		}
		if res.Self.TotalRank > s.c.ImageV2.AllLimit {
			res.Self.TotalRank = int64(bwsmdl.DefaultRank)
		}
	}
	return
}

func (s *Service) accInfos(c context.Context, mids []int64) (infos map[int64]*accapi.Info, err error) {
	mutex := sync.Mutex{}
	midsLen := len(mids)
	group := errgroup.WithContext(c)
	infos = make(map[int64]*accapi.Info, midsLen)
	for i := 0; i < midsLen; i += _aidBulkSize {
		var partMids []int64
		if i+_aidBulkSize > midsLen {
			partMids = mids[i:]
		} else {
			partMids = mids[i : i+_aidBulkSize]
		}
		group.Go(func(ctx context.Context) (err error) {
			infoReqs, e := s.accClient.Infos3(c, &accapi.MidsReq{Mids: partMids})
			if e != nil {
				log.Error("s.accClient.Infos3(%v) error(%v)", partMids, err)
				return e
			}
			mutex.Lock()
			for _, v := range infoReqs.GetInfos() {
				infos[v.Mid] = v
			}
			mutex.Unlock()
			return
		})
	}
	err = group.Wait()
	return
}

func (s *Service) accStats(c context.Context, mids []int64) (stats map[int64]*relaapi.StatReply, err error) {
	mutex := sync.Mutex{}
	midsLen := len(mids)
	group := errgroup.WithContext(c)
	stats = make(map[int64]*relaapi.StatReply, midsLen)
	for i := 0; i < midsLen; i += _aidBulkSize {
		var partMids []int64
		if i+_aidBulkSize > midsLen {
			partMids = mids[i:]
		} else {
			partMids = mids[i : i+_aidBulkSize]
		}
		group.Go(func(ctx context.Context) (err error) {
			infoReqs, e := s.relClient.Stats(c, &relaapi.MidsReq{Mids: partMids})
			if e != nil {
				log.Error("s.relClient.Stats(%v) error(%v)", partMids, err)
				return e
			}
			mutex.Lock()
			for _, v := range infoReqs.GetStatReplyMap() {
				stats[v.Mid] = v
			}
			mutex.Unlock()
			return
		})
	}
	err = group.Wait()
	return
}

// load mobile game act arc data
//func (s *Service) loadMobileGameArcData() {
//	if s.mobileGameArcData == nil {
//		s.mobileGameArcData = new(like.ArcListData)
//	}
//	res, err := s.dao.SourceItem(context.Background(), s.c.MobileGame.Vid)
//	if err != nil {
//		log.Error("failed to load SourceItem(%d,%v)", s.c.MobileGame.Vid, err)
//		return
//	}
//	tmp := new(like.ArcListData)
//	if err = json.Unmarshal(res, &tmp); err != nil {
//		log.Error("loadMobileGameArcData json.Unmarshal(%s) error(%v)", res, err)
//		return
//	}
//	s.mobileGameArcData = tmp
//	log.Info("loadMobileGameArcData success")
//}

func (s *Service) ChildhoodRank(c context.Context) (res []*lmdl.FactionRes, err error) {
	data, err := s.dao.FactionRank(c)
	if err != nil {
		log.Error("ChildhoodRank s.dao.ChildhoodRank error(%v)", err)
		return
	}
	var mids []int64
	for _, v := range data {
		if v == nil {
			continue
		}
		for _, item := range v.List {
			if item != nil && item.Mid > 0 {
				mids = append(mids, item.Mid)
			}
		}
	}
	var infos *accapi.InfosReply
	if len(mids) > 0 {
		if infos, err = s.accClient.Infos3(c, &accapi.MidsReq{Mids: mids}); err != nil {
			log.Error("ChildhoodRank Infos3 mids(%v) error(%v)", mids, err)
			err = nil
		}
	}
	for _, v := range data {
		if v == nil {
			continue
		}
		tmp := &lmdl.FactionRes{
			Sid:   v.Sid,
			Name:  s.c.Faction.Sids[strconv.FormatInt(v.Sid, 10)],
			Score: v.Score,
		}
		for _, item := range v.List {
			if item != nil && item.Mid > 0 {
				if acc, ok := infos.GetInfos()[item.Mid]; ok && acc != nil {
					tmp.AccList = append(tmp.AccList, &lmdl.FactionUser{
						SimpleUser: &lmdl.SimpleUser{
							Mid:  acc.Mid,
							Name: acc.Name,
							Face: acc.Face,
						},
						Score: item.Score,
					})
				}
			}
		}
		res = append(res, tmp)
	}
	return
}

func (s *Service) StupidList(ctx context.Context, sid, mid int64) (*lmdl.StupidListReply, error) {
	total, arcs, err := s.stupidGlobalList(ctx, sid)
	if err != nil {
		log.Error("Failed to fetch stupid global data: %d %+v", sid, err)
		return nil, err
	}
	var (
		target1, target2, target3 int64
		mine1, mine2, mine3       bool
	)
	for _, arc := range arcs {
		if arc.Mid == mid {
			mine1, mine2, mine3 = s.matchTarget(arc.Vv)
		}
		if arc.Vv < s.c.Stupid.Target1 {
			continue
		}
		if arc.Vv < s.c.Stupid.Target2 {
			target1++
			continue
		}
		if arc.Vv < s.c.Stupid.Target3 {
			target1++
			target2++
			continue
		}
		target1++
		target2++
		target3++
	}
	return &lmdl.StupidListReply{
		Global: &lmdl.StupidGolbal{
			Total:   total,
			Target1: target1,
			Target2: target2,
			Target3: target3,
		},
		Individual: &lmdl.StupidIndividual{
			Target1: mine1,
			Target2: mine2,
			Target3: mine3,
		},
	}, nil
}

func (s *Service) stupidGlobalList(ctx context.Context, sid int64) (int64, []*lmdl.StupidVv, error) {
	total, err := s.dao.CacheStupidTotal(ctx, sid)
	if err != nil {
		return 0, nil, err
	}
	arcs, err := s.dao.CacheStupidArcs(ctx, sid)
	if err != nil {
		return 0, nil, err
	}
	return total, arcs, nil
}

func (s *Service) matchTarget(vv int64) (bool, bool, bool) {
	var mine1, mine2, mine3 bool
	if vv >= s.c.Stupid.Target1 {
		mine1 = true
	}
	if vv >= s.c.Stupid.Target2 {
		mine2 = true
	}
	if vv >= s.c.Stupid.Target3 {
		mine3 = true
	}
	return mine1, mine2, mine3
}

// load stupid v2 act arc data
//func (s *Service) loadStupidArcData() {
//	if s.stupidArcData == nil {
//		s.stupidArcData = new(like.ArcListData)
//	}
//	res, err := s.dao.SourceItem(context.Background(), s.c.Stupid.Vid)
//	if err != nil {
//		log.Error("Failed to load SourceItem(%d,%v)", s.c.Stupid.Vid, err)
//		return
//	}
//	tmp := new(like.ArcListData)
//	if err = json.Unmarshal(res, &tmp); err != nil {
//		log.Error("Failed to json unmarshal:%+v", err)
//		return
//	}
//	s.stupidArcData = tmp
//	log.Info("loadStupidArcData success")
//}

// loadGameHolidayArcData stupid v2 act arc data
func (s *Service) loadGameHolidayArcData() {
	if s.gameHolidayArcData == nil {
		s.gameHolidayArcData = new(like.ArcListData)
	}
	res, err := s.dao.SourceItem(context.Background(), s.c.GameHoliday.Vid)
	if err != nil {
		log.Error("Failed to load SourceItem(%d,%v)", s.c.GameHoliday.Vid, err)
		return
	}
	tmp := new(like.ArcListData)
	if err = json.Unmarshal(res, &tmp); err != nil {
		log.Error("Failed to json unmarshal:%+v", err)
		return
	}
	s.gameHolidayArcData = tmp
	log.Info("loadGameHolidayArcData success")
}

// loadTotalRankArcData v2 act arc data
func (s *Service) loadTotalRankArcData() {
	if s.totalRankArcData == nil {
		s.totalRankArcData = new(like.ArcListData)
	}
	res, err := s.dao.SourceItem(context.Background(), s.c.S10Contribution.TotalVid)
	if err != nil {
		log.Error("Failed to load SourceItem(%d,%v)", s.c.S10Contribution.TotalVid, err)
		return
	}
	tmp := new(like.ArcListData)
	if err = json.Unmarshal(res, &tmp); err != nil {
		log.Error("Failed to json unmarshal:%+v", err)
		return
	}
	s.totalRankArcData = tmp
	log.Info("load S10 TotalRank success")
}

// loadDaySelectArcData v2 act arc data
func (s *Service) loadDaySelectArcData() {
	if s.daySelectArcData == nil {
		s.daySelectArcData = new(like.ArcListData)
	}
	res, err := s.dao.SourceItem(context.Background(), s.c.S10Contribution.DayVid)
	if err != nil {
		log.Error("Failed to load SourceItem(%d,%v)", s.c.S10Contribution.DayVid, err)
		return
	}
	tmp := new(like.ArcListData)
	if err = json.Unmarshal(res, &tmp); err != nil {
		log.Error("Failed to json unmarshal:%+v", err)
		return
	}
	s.daySelectArcData = tmp
	log.Info("load S10 DaySelect success")
}

// loadDouble11VideoData
func (s *Service) loadDouble11VideoData() {
	if s.doubl11VidelArcData == nil {
		s.doubl11VidelArcData = new(like.ArcListData)
	}
	res, err := s.dao.SourceItem(context.Background(), s.c.DoubleEleven.VideoListVid)
	if err != nil {
		log.Error("Failed to load SourceItem(%d,%v)", s.c.DoubleEleven.VideoListVid, err)
		return
	}
	tmp := new(like.ArcListData)
	if err = json.Unmarshal(res, &tmp); err != nil {
		log.Error("Failed to json unmarshal:%+v", err)
		return
	}
	s.doubl11VidelArcData = tmp
	log.Info("load double eleven video arcs success")
}

// loadActKnowledgeData
func (s *Service) loadActKnowledgeData() error {
	s.knowledge = make(map[int64][]*lmdl.LIDWithVote, 0)
	StrSIDs := s.c.Knowledge.Sid
	SIDs := strings.Split(StrSIDs, ",")
	tmp := make(map[int64][]*lmdl.LIDWithVote, 0)
	for _, sid := range SIDs {
		v, _ := strconv.ParseInt(sid, 10, 64)
		res, err := s.dao.GetVoteTotalBySid(context.Background(), v)
		if err != nil || res == nil {
			log.Error("Failed to load loadActKnowledgeData sid:%v err:%v", v, err)
			return nil
		}
		log.Info("loadActKnowledgeData into computer cache sid:%v succ:%+v", v, res)
		//s.knowledge[v] = res
		tmp[v] = res
	}
	s.knowledge = tmp
	return nil
}

// loadDouble11ChannelData
func (s *Service) loadDouble11ChannelData() {
	if s.doubl11ChannelArcData == nil {
		s.doubl11ChannelArcData = new(like.ArcListData)
	}
	res, err := s.dao.SourceItem(context.Background(), s.c.DoubleEleven.ChannelListID)
	if err != nil {
		log.Error("Failed to load SourceItem(%d,%v)", s.c.DoubleEleven.ChannelListID, err)
		return
	}
	tmp := new(like.ArcListData)
	if err = json.Unmarshal(res, &tmp); err != nil {
		log.Error("Failed to json unmarshal:%+v", err)
		return
	}
	s.doubl11ChannelArcData = tmp
	log.Info("load double eleven channel arcs success")
}

// loadTimeMachineData
func (s *Service) loadTimeMachineData() {
	if s.timeMachineArcData == nil {
		s.timeMachineArcData = new(like.ArcListData)
	}
	res, err := s.dao.SourceItem(context.Background(), s.c.Timemachine.Vid)
	if err != nil {
		log.Error("Failed to load SourceItem(%d,%v)", s.c.Timemachine.Vid, err)
		return
	}
	tmp := new(like.ArcListData)
	if err = json.Unmarshal(res, &tmp); err != nil {
		log.Error("Failed to json unmarshal:%+v", err)
		return
	}
	s.timeMachineArcData = tmp
	log.Info("load double eleven channel arcs success")
}

//func (s *Service) loadFunnyVideoArcData() {
//	if s.funnyVideoListArcData == nil {
//		s.funnyVideoListArcData = new(like.ArcListData)
//	}
//	res, err := s.dao.SourceItem(context.Background(), s.c.Funny.Vid)
//	if err != nil {
//		log.Error("Failed to load SourceItem(%d,%v)", s.c.Funny.Vid, err)
//		return
//	}
//	tmp := new(like.ArcListData)
//	if err = json.Unmarshal(res, &tmp); err != nil {
//		log.Error("Failed to json unmarshal:%+v", err)
//		return
//	}
//	s.funnyVideoListArcData = tmp
//	log.Info("funnyVideoListArcData success")
//}

func (s *Service) StupidStatus(ctx context.Context, sid string, mid int64) (*lmdl.StupidStatus, error) {
	if sid != s.c.Stupid.LotterySid {
		return nil, xecode.NothingFound
	}
	status := &lmdl.StupidStatus{IsAfrican: false}
	if mid <= 0 {
		return status, nil
	}
	lottery, lotteryTimesConf, err := s.fetchLotteryConfig(ctx, sid)
	if err != nil {
		return nil, err
	}
	status.Lottery = lottery
	status.LotteryTimesConf = lotteryTimesConf
	if res, err := s.dao.RiGet(ctx, fmt.Sprintf(lockKey, lottery.ID, mid)); err != nil || res < 3 {
		return status, nil
	}
	if err := checkTimesConf(lottery, lotteryTimesConf); err != nil {
		return nil, err
	}
	records, err := s.lotteryRecord(ctx, lottery.ID, mid, 0, s.c.Stupid.Num-1)
	if err != nil {
		log.Error("Failed to fetch lottery record: %d %d %+v", lottery.ID, mid, err)
		return nil, err
	}
	i := int64(0)
	for _, record := range records {
		if record.GiftID != 0 {
			return status, nil
		}
		i++
	}
	if i == s.c.Stupid.Num {
		status.IsAfrican = true
	}
	return status, nil
}
