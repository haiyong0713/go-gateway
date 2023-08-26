package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"

	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"

	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	garb "git.bilibili.co/bapis/bapis-go/garb/service"
)

const (
	_archiveURL  = "https://www.bilibili.com/video/av"
	_archiveURL2 = "http://www.bilibili.com/video/av"

	_maxRetryTime = 5
)

var _emptyDySearch = make([]*model.DySeach, 0)

// DySearch load Res info to cache
func (s *Service) DySearch() []*model.DySeach {
	if len(s.dySearch) == 0 {
		s.dySearch = _emptyDySearch
	}
	return s.dySearch
}

func (s *Service) loadResWithRetry() (err error) {
	for i := 0; i < _maxRetryTime; i++ {
		err = s.loadRes()
		if err != nil {
			log.Error("loadResWithRetry s.loadRes() error(%+v)", err)
			continue
		}
		return
	}
	log.Error("loadResWithRetry exceed max try time s.loadRes() error(%+v)", err)
	return
}

// LoadRes load Res info to cache
func (s *Service) loadRes() (err error) {
	// load default banner.
	var posTmp map[int][]int
	resTmp, err := s.res.Resources(context.TODO())
	if err != nil {
		log.Error("s.res.Resources error(%v)", err)
		return
	}
	s.resCache = resTmp
	resCacheMap := make(map[int]*model.Resource)
	posTmp = make(map[int][]int)
	for _, res := range resTmp {
		resCacheMap[res.ID] = res
		if res.Counter == 0 && res.Parent != 0 {
			posTmp[res.Parent] = append(posTmp[res.Parent], res.ID)
		}
	}
	s.posCache = posTmp
	s.resCacheMap = resCacheMap
	// load default banner.
	asgTmp, err := s.res.Assignment(context.TODO())
	if err != nil {
		log.Error("s.res.Assignment error(%v)", err)
		return
	}
	var asgNewTmp []*model.Assignment
	asgNewm, err := s.res.AssignmentNew(context.TODO())
	if err != nil {
		log.Error("s.res.AssignmentNew error(%v)", err)
		return
	}
	for _, asgNews := range asgNewm {
		var (
			weightBanners = make(map[int][]*model.Assignment)
			weightm       = make(map[int]struct{})
		)
		for _, asgNew := range asgNews {
			weightm[asgNew.PositionWeight] = struct{}{}
			weightBanners[asgNew.PositionWeight] = append(weightBanners[asgNew.PositionWeight], asgNew)
		}
		var weights []int
		for weight := range weightm {
			weights = append(weights, weight)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(weights)))
		for _, weight := range weights {
			rand.Seed(time.Now().Unix())
			count := len(weightBanners[weight])
			if count == 0 {
				continue
			}
			asg := weightBanners[weight][rand.Intn(count)]
			if asg != nil {
				asgNewTmp = append(asgNewTmp, asg)
				break
			}
		}
	}
	asgNewTmp = append(asgNewTmp, asgTmp...)
	categoryTmp, err := s.res.CategoryAssignment(context.TODO())
	if err != nil {
		log.Error("s.res.CategoryAssignment error(%v)", err)
		return
	}
	asgNewTmp = append(asgNewTmp, categoryTmp...)
	bossTmp, err := s.res.BossAssignment(context.TODO())
	if err != nil {
		log.Error("s.res.BossAssignment error(%v)", err)
		return
	}
	asgNewTmp = append(asgNewTmp, bossTmp...)
	s.asgCache = asgNewTmp
	resArchiveWarn, resURLWarn := s.formWarnInfo(asgNewTmp)
	s.resArchiveWarnCache = resArchiveWarn
	s.resURLWarnCache = resURLWarn
	asgCacheMap := make(map[int][]*model.Assignment)
	for _, asg := range asgNewTmp {
		asgCacheMap[asg.ResID] = append(asgCacheMap[asg.ResID], asg)
		// get activity ids
		var data struct {
			MissionID int64 `json:"mission_id"`
		}
		if asg.Rule != "" {
			e := json.Unmarshal([]byte(asg.Rule), &data)
			if e != nil {
				log.Error("json.Unmarshal (%s) error(%v)", asg.Rule, e)
			} else {
				if data.MissionID > 0 {
					act, err := s.actgrpc.ActSubject(context.Background(), &actgrpc.ActSubjectReq{Sid: data.MissionID})
					if err != nil {
						log.Error("%v", err)
						continue
					}
					if act != nil && act.Subject != nil {
						asg.ActivityID = data.MissionID
						asg.ActivitySTime = act.Subject.Stime
						asg.ActivityETime = act.Subject.Etime
					}
				}
			}
		}
	}
	s.asgCacheMap = asgCacheMap
	// load default banner.
	bannerTmp, err := s.res.DefaultBanner(context.TODO())
	if err != nil {
		log.Error("s.res.DefaultBanner error(%v)", err)
		return
	}
	s.defBannerCache = bannerTmp
	// index icon
	tmpIndexIcon, err := s.res.IndexIcon(context.TODO())
	if err != nil {
		log.Error("s.res.IndexIcon() error(%v)", err)
		return
	}
	s.indexIcon = tmpIndexIcon
	return
}

func (s *Service) formWarnInfo(asgNewTmp []*model.Assignment) (resArchive map[int64][]*model.ResWarnInfo, resURL map[string][]*model.ResWarnInfo) {
	resArchive = make(map[int64][]*model.ResWarnInfo)
	resURL = make(map[string][]*model.ResWarnInfo)
	for _, asg := range asgNewTmp {
		var (
			aid int64
			url string
			err error
			rw  *model.ResWarnInfo
		)
		if (asg.Atype == model.AsgTypeVideo) || (asg.Atype == model.AsgTypeAv) {
			if aid, err = strconv.ParseInt(asg.URL, 10, 64); err != nil {
				log.Error("formWarnInfo url(%v) error(%v)", asg.URL, err)
				err = nil
				continue
			}
		} else if (asg.Atype == model.AsgTypePic) || (asg.Atype == model.AsgTypeURL) {
			if strings.HasPrefix(asg.URL, _archiveURL) {
				urls := strings.Split(asg.URL, "?")
				aidURL := strings.TrimPrefix(urls[0], _archiveURL)
				aidURL = strings.TrimSuffix(aidURL, "/")
				if aid, err = strconv.ParseInt(aidURL, 10, 64); err != nil {
					log.Error("formWarnInfo url(%v),aidURL(%v) error(%v)", asg.URL, aidURL, err)
					err = nil
					continue
				}
			} else if strings.HasPrefix(asg.URL, _archiveURL2) {
				urls := strings.Split(asg.URL, "?")
				aidURL := strings.TrimPrefix(urls[0], _archiveURL2)
				aidURL = strings.TrimSuffix(aidURL, "/")
				if aid, err = strconv.ParseInt(aidURL, 10, 64); err != nil {
					log.Error("formWarnInfo url(%v) error(%v)", asg.URL, err)
					err = nil
					continue
				}
			} else {
				url = asg.URL
			}
		}
		if aid == 0 && url == "" {
			continue
		}
		rw = &model.ResWarnInfo{
			AssignmentID:   asg.AsgID,
			AssignmentName: asg.Name,
			STime:          asg.STime,
			ETime:          asg.ETime,
			UserName:       asg.Username,
			ApplyGroupID:   asg.ApplyGroupID,
			MaterialID:     asg.ID,
		}
		if re, ok := s.resCacheMap[asg.ResID]; ok {
			if re.Counter > 0 {
				rw.ResourceID = re.ID
				if rep, ok := s.resCacheMap[re.Parent]; ok {
					rw.ResourceName = fmt.Sprintf("%v_%v", rep.Name, re.Name)
					continue
				}
				rw.ResourceName = re.Name
			} else {
				rw.ResourceID = re.Parent
				rw.ResourceName = re.Name
			}
		}
		if aid != 0 {
			rw.AID = aid
			resArchive[aid] = append(resArchive[aid], rw)
		} else {
			rw.URL = url
			resURL[url] = append(resURL[url], rw)
		}
	}
	return
}

// ResourceAll get all resource
func (s *Service) ResourceAll(c context.Context) (res []*model.Resource) {
	res = s.resCache
	return
}

// AssignmentAll get all assignment
func (s *Service) AssignmentAll(c context.Context) (ass []*model.Assignment) {
	// TODO delete
	for _, asc := range s.asgCache {
		as := &model.Assignment{}
		*as = *asc
		as.Weight = 0
		as.Operater = model.Operater
		ass = append(ass, as)
	}
	return
}

// Resource get resource by resource_id or positon_id
func (s *Service) Resource(c context.Context, resID int) (res *model.Resource) {
	var (
		ok  bool
		pos []int
	)
	if res, ok = s.resCacheMap[resID]; !ok {
		return
	}
	// Safe first!! Prevent res nil panic.
	if res.Counter == 0 {
		if len(s.asgCacheMap[resID]) > 0 {
			res.Assignments = s.asgCacheMap[resID]
			return
		}
		res.Assignments = s.asgCacheMap[res.Parent]
	} else {
		if pos, ok = s.posCache[resID]; !ok {
			return
		}
		var (
			tmpNormalRes   []*model.Assignment
			tmpCategoryRes = s.asgCacheMap[resID]
		)
		for _, pid := range pos {
			tmpNormalRes = append(tmpNormalRes, s.asgCacheMap[pid]...)
		}
		for _, nr := range tmpNormalRes {
			if nr.Weight > len(tmpCategoryRes) {
				tmpCategoryRes = append(tmpCategoryRes, nr)
			} else {
				tmpCategoryRes = append(tmpCategoryRes[:nr.Weight-1], append([]*model.Assignment{nr}, tmpCategoryRes[nr.Weight-1:]...)...)
			}
		}
		if len(tmpCategoryRes) > res.Counter {
			res.Assignments = tmpCategoryRes[:res.Counter]
		} else {
			res.Assignments = tmpCategoryRes
		}
	}
	return
}

// Resources get resources by resource_ids or position_ids
func (s *Service) Resources(c context.Context, resIDs []int) (res map[int]*model.Resource) {
	if len(resIDs) == 0 {
		res = _emptyResources
		return
	}
	res = make(map[int]*model.Resource)
	for _, rid := range resIDs {
		if resTmp := s.Resource(c, rid); resTmp != nil {
			res[rid] = resTmp
		}
	}
	return
}

// DefBanner get defbanner config
func (s *Service) DefBanner(c context.Context) (defbanner *model.Assignment) {
	defbanner = s.defBannerCache
	return
}

// IndexIcon get index icon
func (s *Service) IndexIcon(c context.Context) (icons map[string][]*model.IndexIcon) {
	icons = map[string][]*model.IndexIcon{
		model.IconTypes[model.IconTypeFix]:    s.indexIcon[model.IconTypeFix],
		model.IconTypes[model.IconTypeRandom]: s.indexIcon[model.IconTypeRandom],
	}
	return
}

// get icon from garb
func (s *Service) getUserPurchasedIcon(c context.Context, mid int64) (re *model.PlayerIcon, err error) {
	in := &garb.PlayIconUserEquipReq{
		Mid: mid,
	}
	var out *garb.PlayIconUserEquipReply
	if out, err = s.garbGRPC.PlayIconUserEquip(c, in); err != nil || out == nil || out.PlayIcon == nil {
		return nil, err
	}
	s.infoProm.Incr("PlayIconUserEquip-HasValue")
	re = &model.PlayerIcon{
		URL1:         out.PlayIcon.DragIcon,
		URL2:         out.PlayIcon.Icon,
		Hash1:        out.PlayIcon.DragIconHash,
		Hash2:        out.PlayIcon.IconHash,
		CTime:        xtime.Time(out.PlayIcon.Ver),
		DragLeftPng:  out.PlayIcon.DragLeftPng,
		MiddlePng:    out.PlayIcon.MiddlePng,
		DragRightPng: out.PlayIcon.DragRightPng,
	}
	if out.PlayIcon.DragData != nil {
		re.DragData = &pb.IconData{
			MetaJson:  out.PlayIcon.DragData.MetaJson,
			SpritsImg: out.PlayIcon.DragData.SpritsImg,
		}
	}
	if out.PlayIcon.NodragData != nil {
		re.NoDragData = &pb.IconData{
			MetaJson:  out.PlayIcon.NodragData.MetaJson,
			SpritsImg: out.PlayIcon.NodragData.SpritsImg,
		}
	}
	return re, nil

}

// PlayerIcon get player icon
func (s *Service) PlayerIcon(c context.Context, aid int64, tagIds []int64, typeId int32, mid int64, showPlayicon, isUnder604 bool) (re *model.PlayerIcon, err error) {
	// archive > purchased > tag > type > overall
	// isUnder604:true 不返回运营icon
	var ok bool
	if aid != 0 && !isUnder604 {
		if re, ok = s.playIconArchive[aid]; ok && re != nil {
			return
		}
	}
	//获取用户购买的icon
	if mid > 0 && showPlayicon {
		if re, err = s.getUserPurchasedIcon(c, mid); re != nil && err == nil {
			return re, nil
		}
	}
	//安卓<=6040000 不返回运营icon
	if isUnder604 {
		s.infoProm.Incr("play_icon_empty")
		err = ecode.NothingFound
		return
	}
	var reTmp *model.PlayerIcon
	for _, tagId := range tagIds {
		if reTmp, ok = s.playIconTag[tagId]; ok {
			if re == nil || re.MTime < reTmp.MTime {
				re = reTmp
			}
		}
	}
	if re != nil {
		return
	}
	if typeId != 0 {
		if re, ok = s.playIconType[typeId]; ok && re != nil {
			return
		}
		tid := strconv.Itoa(int(typeId))
		//nolint:gosec
		ptid, _ := strconv.Atoi(s.typeList[tid])
		if re, ok = s.playIconType[int32(ptid)]; ok && re != nil {
			return
		}
	}
	if re = s.playIcon; re == nil {
		err = ecode.NothingFound
	}
	return
}

// PlayerPgcIcon get pgc player icon
func (s *Service) PlayerPgcIcon(c context.Context, sid, mid int64, showPlayicon bool) (res *model.PlayerIcon) {
	var ok bool
	if res, ok = s.playIconPgc[sid]; ok && res != nil {
		return
	}
	if mid > 0 && showPlayicon {
		if res, err := s.getUserPurchasedIcon(c, mid); res != nil && err == nil {
			return res
		}
	}
	//全局
	if s.playIcon != nil {
		res = s.playIcon
	}
	return
}

//nolint:gocognit
func (s *Service) WebPlayerIcon(c context.Context, req *pb.WebPlayerIconRequest) (*pb.WebPlayerIconReply, error) {
	// avid > purchased > tag > typeID > all > seasonID
	data := func() *model.PlayerIcon {
		var (
			ok  bool
			res *model.PlayerIcon
		)
		if req.GetAid() > 0 {
			if res, ok = s.playIconArchive[req.GetAid()]; ok && res != nil {
				return res
			}
		}
		if req.GetMid() > 0 {
			if res, err := s.getUserPurchasedIcon(c, req.GetMid()); res != nil && err == nil {
				return res
			}
		}
		var reTmp *model.PlayerIcon
		for _, tagId := range req.GetTagIDs() {
			if reTmp, ok = s.playIconTag[tagId]; ok && reTmp != nil {
				if res == nil || res.MTime < reTmp.MTime {
					res = reTmp
				}
			}
		}
		if res != nil {
			return res
		}
		if req.GetTypeID() > 0 {
			if res, ok = s.playIconType[req.GetTypeID()]; ok && res != nil {
				return res
			}
			tid := strconv.Itoa(int(req.GetTypeID()))
			//nolint:gosec
			ptid, _ := strconv.Atoi(s.typeList[tid])
			if res, ok = s.playIconType[int32(ptid)]; ok && res != nil {
				return res
			}
		}
		if res = s.playIcon; res != nil {
			return res
		}
		if req.GetSeasonID() > 0 {
			if res, ok = s.playIconPgc[req.GetSeasonID()]; ok && res != nil {
				return res
			}
		}
		return nil
	}()
	if data == nil {
		return nil, ecode.NothingFound
	}
	return &pb.WebPlayerIconReply{
		Icon: &pb.PlayerIcon{
			URL1:  data.URL1,
			Hash1: data.Hash1,
			URL2:  data.URL2,
			Hash2: data.Hash2,
			Ctime: data.CTime,
		},
	}, nil
}

// Cmtbox get live danmaku box
func (s *Service) Cmtbox(c context.Context, id int64) (re *model.Cmtbox, err error) {
	var ok bool
	if re, ok = s.cmtbox[id]; !ok {
		err = ecode.NothingFound
	}
	return
}

//func (s *Service) loadCustomConfig() error {
//	ccs, err := s.res.CustomConfigs(context.Background())
//	if err != nil {
//		log.Error("Failed to get custom configs: %+v", err)
//		return err
//	}
//	s.customConfigStore = ccs
//	return nil
//}

// loadSearchOgvConfig .
func (s *Service) loadSearchOgvConfig() {
	var (
		err error
	)
	tmp, err := s.show.SearchOgv(context.Background())
	if err != nil {
		log.Error("loadSearchOgvConfig error(%+v)", err)
		return
	}
	if tmp != nil {
		s.searchOgvCache = tmp
	}
	log.Info("loadSearchOgvConfig Success ID(%v)", s.searchOgvCache)
}

// CustomConfig is
func (s *Service) CustomConfig(ctx context.Context, req *pb.CustomConfigRequest) (rep *pb.CustomConfigReply, err error) {
	var cc *model.CustomConfig
	if cc, err = s.res.GetCustomConfigBySF(ctx, req.TP, req.Oid); err != nil {
		return nil, err
	} else if cc == nil {
		return nil, ecode.NothingFound
	}
	reply := &pb.CustomConfigReply{
		TP:               cc.TP,
		Oid:              cc.Oid,
		Content:          cc.Content,
		URL:              cc.URL,
		HighlightContent: cc.HighlightContent,
		Image:            cc.Image,
		ImageBig:         cc.ImageBig,
		STime:            cc.STime.Unix(),
		ETime:            cc.ETime.Unix(),
		State:            1,
	}
	return reply, nil
}
