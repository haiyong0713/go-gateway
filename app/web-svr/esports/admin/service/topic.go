package service

import (
	"context"
	"strconv"
	"strings"
	"sync"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/esports/admin/model"
	"go-gateway/app/web-svr/esports/common/helper"
	pb "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/pkg/idsafe/bvid"

	errGroup "go-common/library/sync/errgroup.v2"
)

const (
	_maxCheckAidCount = 50
	_maxAllAidCount   = 100
)

func checkParamError(param *model.VideoList) error {
	if param.UgcAids == "" && param.GameID == 0 && param.MatchID == 0 && param.YearID == 0 {
		return xecode.Errorf(xecode.RequestErr, "ugc视频和视频库视频至少配置一个才可保存")
	}
	if checkFilterEmpty(param) {
		return xecode.Errorf(xecode.RequestErr, "游戏，赛事，年份必须全部填写")
	}
	return nil
}

func checkFilterEmpty(param *model.VideoList) bool {
	if param.GameID != 0 && (param.MatchID == 0 || param.YearID == 0) {
		return true
	} else if param.MatchID != 0 && (param.GameID == 0 || param.YearID == 0) {
		return true
	} else if param.YearID != 0 && (param.GameID == 0 || param.MatchID == 0) {
		return true
	}
	return false
}

// AddTopicVideoList .
func (s *Service) AddTopicVideoList(ctx context.Context, param *model.VideoList) (res *model.CheckArchive, err error) {
	if err = checkParamError(param); err != nil {
		log.Errorc(ctx, "AddTopicVideoList checkParamError(%+v) error(%v)", param, err)
		return
	}
	if res, err = s.checkSaveVideoList(ctx, param); err != nil {
		return
	}
	if err = s.dao.DB.Model(&model.VideoList{}).Create(param).Error; err != nil {
		log.Errorc(ctx, "AddTopicVideoList s.dao.DB.Model Create(%+v) error(%v)", param, err)
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		if e := s.ClearVideoListCacheByGRPC(param.ID); e != nil {
			log.Errorc(ctx, "AddTopicVideoList s.ClearVideoListCacheByGRPC() id(%d) error(%+v)", param.ID, err)
		}
	})
	return
}

// EditTopicVideoList .
func (s *Service) EditTopicVideoList(ctx context.Context, param *model.VideoList) (res *model.CheckArchive, err error) {
	if err = checkParamError(param); err != nil {
		log.Errorc(ctx, "EditTopicVideoList checkParamError(%+v) error(%v)", param, err)
		return
	}
	preData := new(model.VideoList)
	if err = s.dao.DB.Where("id=?", param.ID).First(&preData).Error; err != nil {
		log.Errorc(ctx, "EditTopicVideoList s.dao.DB.Where id(%d) error(%d)", param.ID, err)
		return
	}
	if res, err = s.checkSaveVideoList(ctx, param); err != nil {
		return
	}
	if err = s.dao.DB.Model(&model.VideoList{}).Update(param).Error; err != nil {
		log.Errorc(ctx, "EditTopicVideoList s.dao.DB.Model Update(%+v) error(%v)", param, err)
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		if e := s.ClearVideoListCacheByGRPC(param.ID); e != nil {
			log.Errorc(ctx, "EditTopicVideoList s.ClearVideoListCacheByGRPC() id(%d) error(%+v)", param.ID, err)
		}
	})
	return
}

func (s *Service) checkSaveVideoList(ctx context.Context, param *model.VideoList) (res *model.CheckArchive, err error) {
	res = &model.CheckArchive{WrongList: make([]string, 0)}
	if param.UgcAids != "" {
		var wrongs, rights []string
		archiveIDs := strings.Split(param.UgcAids, ",")
		if len(archiveIDs) > _maxAllAidCount {
			err = xecode.Errorf(xecode.RequestErr, "最多100个bvid/avid")
			return
		}
		if wrongs, rights, err = s.GetWrongsOrRights(ctx, archiveIDs); err != nil {
			res.WrongList = wrongs
			return
		}
		param.UgcAids = strings.Join(rights, ",")
	}
	return
}

// ForbidTopicVideoList .
func (s *Service) ForbidTopicVideoList(ctx context.Context, id int64, state int) (err error) {
	preVideoList := new(model.VideoList)
	if err = s.dao.DB.Where("id=?", id).First(&preVideoList).Error; err != nil {
		log.Errorc(ctx, "ForbidTopicVideoList s.dao.DB.Where id(%d) error(%d)", id, err)
		return
	}
	if err = s.dao.DB.Model(&model.VideoList{}).Where("id=?", id).Update(map[string]int{"is_deleted": state}).Error; err != nil {
		log.Errorc(ctx, "ForbidTopicVideoList s.dao.DB.Model error(%v)", err)
		return
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		if e := s.ClearVideoListCacheByGRPC(id); e != nil {
			log.Errorc(ctx, "ForbidTopicVideoList s.ClearVideoListCacheByGRPC() id(%d) error(%+v)", id, err)
		}
	})
	return
}

// VideoListInfo .
func (s *Service) VideoListInfo(ctx context.Context, id int64) (data *model.VideoList, err error) {
	data = new(model.VideoList)
	if err = s.dao.DB.Model(&model.VideoList{}).Where("id=?", id).First(&data).Error; err != nil {
		log.Errorc(ctx, "VideoListInfo Error (%v)", err)
	}
	return
}

// TopicVideoLists .
func (s *Service) TopicVideoLists(ctx context.Context, pn, ps int64, title string) (list []*model.VideoList, count int64, err error) {
	source := s.dao.DB.Model(&model.VideoList{}).Where("is_deleted=?", _notDeleted)
	if title != "" {
		source = source.Where("list_name like ?", "%"+title+"%")
	}
	source.Count(&count)
	if err = source.Offset((pn - 1) * ps).Limit(ps).Find(&list).Error; err != nil {
		log.Errorc(ctx, "TopicVideoLists Error (%v)", err)
		return
	}
	return
}

func (s *Service) TopicCheckArchives(ctx context.Context, archiveIDs []string) (res *model.CheckArchive, err error) {
	res = &model.CheckArchive{WrongList: make([]string, 0)}
	if len(archiveIDs) > _maxCheckAidCount {
		err = xecode.Errorf(xecode.RequestErr, "请输入1~50个bvid/avid")
		return
	}
	res.WrongList, _, err = s.GetWrongsOrRights(ctx, archiveIDs)
	return
}

func (s *Service) GetWrongsOrRights(ctx context.Context, archiveIDs []string) (wrong, right []string, err error) {
	wrong = make([]string, 0)
	right = make([]string, 0)
	count := len(archiveIDs)
	if count == 0 {
		return
	}
	var (
		bvidMap   map[string]int64
		paramAids []int64
		rightMap  map[int64]struct{}
	)
	if bvidMap, err = helper.BvidsToAid(ctx, archiveIDs); err != nil {
		err = xecode.Errorf(xecode.RequestErr, "输入bvid/avid不正确(%+v)", err)
		return
	}
	aidMap := make(map[int64]struct{}, count)
	for _, aid := range bvidMap {
		if _, ok := aidMap[aid]; !ok {
			paramAids = append(paramAids, aid)
			aidMap[aid] = struct{}{}
		}
	}
	if rightMap, err = s.rightAllArchive(ctx, paramAids); err != nil {
		err = xecode.Errorf(xecode.RequestErr, "调用稿件服务出错(%+v)", err)
		return
	}
	wrongAids := make(map[int64]struct{}, count)
	for _, aid := range paramAids {
		if _, ok := rightMap[aid]; !ok {
			wrongAids[aid] = struct{}{}
		}
	}
	for strID, aid := range bvidMap {
		if _, ok := wrongAids[aid]; ok {
			wrong = append(wrong, strID)
		}
	}
	if len(wrong) > 0 {
		err = xecode.Errorf(xecode.RequestErr, "不正确的bvid/avid")
		return
	}
	rsMap := make(map[int64]struct{}, 0)
	for _, strID := range archiveIDs {
		var intAid int64
		if strings.HasPrefix(strID, "BV1") {
			if intAid, err = bvid.BvToAv(strID); err != nil {
				err = xecode.Errorf(xecode.RequestErr, "bvid.BvToAv()出错")
				return
			}
		} else {
			if intAid, err = strconv.ParseInt(strID, 10, 64); err != nil {
				err = xecode.Errorf(xecode.RequestErr, "strconv.ParseInt()出错")
				return
			}
		}
		if _, isHave := rsMap[intAid]; !isHave { // 自动去重，取第一个
			right = append(right, strID)
			rsMap[intAid] = struct{}{}
		}
	}
	return
}

func (s *Service) rightAllArchive(ctx context.Context, aids []int64) (arcNormal map[int64]struct{}, err error) {
	var (
		arcErr error
		mutex  = sync.Mutex{}
	)
	group := errGroup.WithContext(ctx)
	aidsLen := len(aids)
	arcNormal = make(map[int64]struct{}, aidsLen)
	for i := 0; i < aidsLen; i += _arcsSize {
		var partAids []int64
		if i+_arcsSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_arcsSize]
		}
		group.Go(func(ctx context.Context) error {
			var tmpRes *arcmdl.ArcsReply
			if tmpRes, arcErr = s.arcClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: partAids}); arcErr != nil {
				log.Errorc(ctx, "wrongAllArchive s.arcClient.Arcs(%v) error %v", partAids, err)
				return arcErr
			}
			if tmpRes != nil {
				for _, arc := range tmpRes.Arcs {
					if arc != nil && arc.IsNormal() {
						mutex.Lock()
						arcNormal[arc.Aid] = struct{}{}
						mutex.Unlock()
					}
				}
			}
			return nil
		})
	}
	if err = group.Wait(); err != nil {
		log.Errorc(ctx, "wrongAllArchive group.Wait() aids(%v) error(%v)", aids, err)
	}
	return
}

func (s *Service) TopicVideoFilter(ctx context.Context, param *model.ParamVideoFilter) (res *pb.VideoListFilterReply, err error) {
	arg := &pb.VideoListFilterRequest{
		GameId:  param.GameID,
		MatchId: param.MatchID,
		YearId:  param.YearID,
	}
	res, err = s.espClient.VideoListFilter(ctx, arg)
	return
}
