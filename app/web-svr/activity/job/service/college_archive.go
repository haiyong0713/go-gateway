package service

import (
	"context"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/job/model/college"
	"go-gateway/app/web-svr/activity/job/model/rank"
	"time"

	tagnewapi "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

const (
	archiveType          = 3
	archivePs            = 50
	aidChannelLength     = 5
	archiveChannelLength = 5
	oneSecondTimes       = 20
	// sourceByTime tag根据时间顺序获取稿件
	sourceByTime = 0
	tagMaxCount  = 1000
)

// collegeAllArchive 分学校获取稿件
func (s *Service) collegeAllArchive(c context.Context, collegeList []*college.College) error {
	s.collegeTabList = make(map[int64][]int64)
	s.collegeArchiveTopList = make(map[int64][]int64)
	adjustArchiveMap, err := s.getCollegeAdjustArchive(c)
	if err != nil {
		log.Errorc(c, "s.getCollegeAdjustArchive err(%v)", err)
		return err
	}
	for _, v := range collegeList {
		whiteList := make(map[int64]struct{})
		if v.White != nil {
			for _, v := range v.White {
				whiteList[v] = struct{}{}
			}
		}
		if v.TagID > 0 {
			err = s.collegeTagArchive(c, v, adjustArchiveMap, whiteList)
			if err != nil {
				log.Errorc(c, "s.collegeTagArchive collegeID(%d) err(%v)", v.ID, err)
				return err
			}
		}

	}
	return nil
}

// getCollegeAdjustArchive 获取调整稿件积分
func (s *Service) getCollegeAdjustArchive(c context.Context) (map[int64]*college.Archive, error) {
	var (
		err   error
		batch int
	)
	archiveList := make([]*college.Archive, 0)
	archiveMap := make(map[int64]*college.Archive)
	for {
		archiveBatchList, err := s.college.GetCollegeAdjustArchive(c, s.collegeMysqlOffset(batch), maxBatchLimit)
		if err != nil {
			log.Errorc(c, "s.college.GetCollegeAdjustArchive: error(%v)", err)
			break
		}
		archiveList = append(archiveList, archiveBatchList...)
		if len(archiveBatchList) < maxBatchLimit {
			break
		}
		time.Sleep(100 * time.Microsecond)
		batch++
	}
	for _, v := range archiveList {
		archiveMap[v.AID] = v
	}
	return archiveMap, err
}

// collegeTagArchive tag下的稿件
func (s *Service) collegeTagArchive(c context.Context, collegeInfo *college.College, archiveAdjuest map[int64]*college.Archive, whiteList map[int64]struct{}) error {
	aidCh := make(chan []int64, aidChannelLength)
	arcCh := make(chan *api.ArcsReply, archiveChannelLength)
	eg := errgroup.WithContext(c)
	var (
		archiveStateMap map[int64]rank.ArchiveBatch
	)
	eg.Go(func(ctx context.Context) (err error) {
		err = s.collegeArchiveIntoChannel(c, collegeInfo.TagID, aidCh)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		archiveStateMap, err = s.collegeArchiveDetail(c, collegeInfo, archiveAdjuest, whiteList, aidCh, arcCh)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return err
	}
	tabArchive := s.collegeAllTabRank(c, collegeInfo, archiveStateMap)
	for tagID, v := range tabArchive {
		err := s.collegeTabArchiveSave(c, collegeInfo, tagID, v)
		if err != nil {
			log.Errorc(c, "s.collegeTabArchiveSave collegeId (%d) tagID(%d), err(%v)", collegeInfo.ID, tagID, err)
			return err
		}
	}
	return nil
}

func (s *Service) collegeTabArchiveSave(c context.Context, collegeInfo *college.College, tabID int64, archiveBatch []*rank.ArchiveScore) error {
	aids := make([]int64, 0)
	for _, v := range archiveBatch {
		aids = append(aids, v.Aid)
	}
	return s.college.SetArchiveTabArchive(c, collegeInfo.ID, tabID, aids)
}

// collegeTabRank 稿件维度排行计算
func (s *Service) collegeAllTabRank(c context.Context, collegeInfo *college.College, mapTabArchive map[int64]rank.ArchiveBatch) map[int64][]*rank.ArchiveScore {
	mapTabArchiveRank := make(map[int64][]*rank.ArchiveScore)
	if mapTabArchive != nil {
		for k, v := range mapTabArchive {
			mapTabArchiveRank[k] = s.collegeArchiveScoreRank(c, collegeInfo, v)
		}
	}
	// 获取稿件top2
	s.collegeGetArchiveTopList(c, collegeInfo, mapTabArchiveRank)

	return mapTabArchiveRank
}

// collegeGetArchiveTopList 获取top2
func (s *Service) collegeGetArchiveTopList(c context.Context, collegeInfo *college.College, mapTabArchiveRank map[int64][]*rank.ArchiveScore) {
	archiveList := make([]int64, 0)
	if tabNormalList, ok := mapTabArchiveRank[college.TabNormal]; ok {
		if len(tabNormalList) >= 1 {
			archiveList = append(archiveList, tabNormalList[0].Aid)
		}
	}
	if tabWhiteList, ok := mapTabArchiveRank[college.TabWhite]; ok {
		if len(tabWhiteList) >= 1 {
			archiveList = append(archiveList, tabWhiteList[0].Aid)
		}
	}
	if len(archiveList) < 2 {
		if tabNormalList, ok := mapTabArchiveRank[college.TabNormal]; ok {
			if len(tabNormalList) >= 2 {
				archiveList = append(archiveList, tabNormalList[1].Aid)
			}
		}
	}
	if len(archiveList) < 2 {
		if tabWhiteList, ok := mapTabArchiveRank[college.TabWhite]; ok {
			if len(tabWhiteList) >= 2 {
				archiveList = append(archiveList, tabWhiteList[1].Aid)
			}
		}
	}
	s.collegeArchiveTopList[collegeInfo.ID] = archiveList
}

// collegeArchiveRank 稿件维度排行计算
func (s *Service) collegeArchiveScoreRank(c context.Context, collegeInfo *college.College, mapArchive rank.ArchiveBatch) []*rank.ArchiveScore {
	archiveScoreMap := mapArchive.Score(collegeArchiveScore)
	return s.collegeArchiveRank(c, *archiveScoreMap, s.c.College.ArchiveTopLength)
}

func (s *Service) collegeArchiveRank(c context.Context, archiveScoreMap rank.ArchiveScoreMap, topLength int) []*rank.ArchiveScore {
	var archiveScoreBatch = rank.ArchiveScoreBatch{}
	for _, v := range archiveScoreMap {
		archiveScoreBatch.Data = append(archiveScoreBatch.Data, v)
	}
	archiveScoreBatch.TopLength = topLength
	rank.Sort(&archiveScoreBatch)
	return archiveScoreBatch.Data
}

func collegeArchiveScore(arc *rank.ArchiveStat) int64 {
	if arc.View == 0 {
		return arc.Adjust
	}
	return getPlayScore(arc) + getQualityScore(arc) + getTopicScore(arc) + arc.Adjust
}

// collegeArchiveIntoChannel 根据名字获取稿件信息
func (s *Service) collegeArchiveIntoChannel(c context.Context, tagID int64, aidCh chan []int64) error {
	var offset string
	defer close(aidCh)
	var count int
	for {
		count++
		startTime := time.Now().UnixNano() / 1e6
		for i := 0; i < oneSecondTimes; i++ {
			rids, err := s.tagNewClient.RidsByTag(c, &tagnewapi.RidsByTagReq{
				Tid:    tagID,
				Offset: offset,
				Ps:     archivePs,
				Typ:    archiveType,
				Source: sourceByTime,
			})
			if err != nil && rids == nil {
				log.Errorc(c, "college tags error (%v)", err)
				if xecode.EqualError(xecode.Deadline, err) {
					time.Sleep(time.Second)
				}
				continue
			}
			offset = rids.Offset
			if len(rids.Rids) > 0 {
				aidCh <- rids.Rids
			}
			if !rids.Hasmore || len(rids.Rids) == 0 {
				return nil
			}
		}
		endTime := time.Now().UnixNano() / 1e6
		waitTime := s.getWaitTime(startTime, endTime)
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Millisecond)
		}
		if count == tagMaxCount {
			log.Errorc(c, "tag get aid list count equal max")
			return nil
		}
	}
}

func (s *Service) collegeArchiveDetail(c context.Context, collegeInfo *college.College, archiveScore map[int64]*college.Archive, whiteList map[int64]struct{}, aidCh chan []int64, arcCh chan *api.ArcsReply) (map[int64]rank.ArchiveBatch, error) {
	var tagArchiveMap = make(map[int64]rank.ArchiveBatch)
	tagArchiveBatchmake := make(map[int64][]*rank.ArchiveStat)
	tagArchiveBatchmake[college.TabWhite] = make([]*rank.ArchiveStat, 0)
	tagArchiveBatchmake[college.TabNormal] = make([]*rank.ArchiveStat, 0)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		defer close(arcCh)
		for v := range aidCh {
			aids := v
			if len(aids) > 0 {
				err = s.archiveInfo(c, aids, arcCh)
			}
		}
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		aidsMap := make(map[int64]bool)
		for v := range arcCh {
			if v == nil || v.Arcs == nil {
				err = ecode.ActivityWriteHandArchiveErr
			}
			for aid, arc := range v.Arcs {
				if arc == nil {
					err = ecode.ActivityWriteHandArchiveErr
				}
				// 防止aid重复
				if _, ok := aidsMap[aid]; !ok && arc.IsNormal() {
					ctime := arc.Ctime.Time().Unix()
					if ctime < s.c.College.ArchiveCtime {
						continue
					}
					var archiveScoreInt int64
					if archive, ok := archiveScore[aid]; ok {
						archiveScoreInt = archive.Score
					}
					archive := &rank.ArchiveStat{
						Mid:     arc.Author.Mid,
						Aid:     aid,
						View:    arc.Stat.View,
						Danmaku: arc.Stat.Danmaku,
						Reply:   arc.Stat.Reply,
						Fav:     arc.Stat.Fav,
						Coin:    arc.Stat.Coin,
						Share:   arc.Stat.Share,
						NowRank: arc.Stat.NowRank,
						Like:    arc.Stat.Like,
						Videos:  arc.Videos,
						Adjust:  archiveScoreInt,
					}
					aidsMap[aid] = true

					if whiteList != nil {
						if _, ok := whiteList[arc.Author.Mid]; ok {
							tagArchiveBatchmake[college.TabWhite] = append(tagArchiveBatchmake[college.TabWhite], archive)
							continue
						}
					}
					tagArchiveBatchmake[college.TabNormal] = append(tagArchiveBatchmake[college.TabNormal], archive)
				}
			}
		}
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, err
	}
	if len(tagArchiveBatchmake[college.TabWhite]) > 0 {
		s.collegeTabList[collegeInfo.ID] = append(s.collegeTabList[collegeInfo.ID], college.TabWhite)
	}
	if len(tagArchiveBatchmake[college.TabNormal]) > 0 {
		s.collegeTabList[collegeInfo.ID] = append(s.collegeTabList[collegeInfo.ID], college.TabNormal)
	}
	tagArchiveMap[college.TabWhite] = make(rank.ArchiveBatch)
	tagArchiveMap[college.TabNormal] = make(rank.ArchiveBatch)
	for _, v := range tagArchiveBatchmake[college.TabWhite] {
		tagArchiveMap[college.TabWhite][v.Aid] = v
	}
	for _, v := range tagArchiveBatchmake[college.TabNormal] {
		tagArchiveMap[college.TabNormal][v.Aid] = v
	}

	return tagArchiveMap, nil
}
