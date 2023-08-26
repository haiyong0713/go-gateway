package service

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/job/model/college"
	"go-gateway/app/web-svr/activity/job/model/rank"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	// oneSecond 1秒=1000ms
	oneSecond = 1000
	// concurrencyGetMidScore 并发获取用户积分100
	concurrencyGetMidScore = 100
	// collegeDbChannel channel长度
	collegeDbChannel = 10
	// maxBatchLimit 批量拉取数量
	maxBatchLimit = 1000
	// personalInfoBatch 批量保存的用户信息数量
	personalInfoBatch = 100
	// concurrencyPersonal 批量保存的用户信息并发数
	concurrencyPersonal = 1
	// rankType
	rankTypeNationwide = 0
	// collegeInfoBatch 批量保存的学校排行数量
	collegeInfoBatch = 300
	// concurrencycollege 批量保存的学校信息并发数
	concurrencycollege = 1
	// collegeDetailBatch 批量保存学校详情数量
	collegeDetailBatch = 500
	// concurrencycollegeDetail 批量保存的学校详情并发数
	concurrencycollegeDetail = 2
	// allCollegeLimit 一次取学校的个数
	allCollegeLimit = 500
)

var collegeCtx context.Context
var collegeVersionCtx context.Context

func collegeCtxInit() {
	collegeCtx = trace.SimpleServerTrace(context.Background(), "collegeRank")
}

func collegeVersionCtxInit() {
	collegeVersionCtx = trace.SimpleServerTrace(context.Background(), "collegeVersionRank")
}

// CollegeRank 学院排行
func (s *Service) CollegeRank() {
	start := time.Now()
	s.collegeRankRunning.Lock()
	if start.Unix() > s.c.College.ActivityEnd.Unix() {
		return
	}
	defer s.collegeRankRunning.Unlock()
	collegeCtxInit()
	log.Infoc(collegeCtx, "college rank start (%d)", start.Unix())
	// 获取本次脚本的版本号
	version, err := s.college.GetCollegeUpdateVersion(collegeCtx)
	if err != nil {
		log.Errorc(collegeCtx, "s.GetCollegeUpdateVersion(c) err(%v)", err)
		return
	}
	collegeList, err := s.getAllCollege(collegeCtx)
	if err != nil {
		log.Errorc(collegeCtx, "s.getAllCollege(c) err(%v)", err)
		return
	}
	err = s.collegeAllArchive(collegeCtx, collegeList)
	if err != nil {
		log.Errorc(collegeCtx, "s.collegeAllArchive(c) err(%v)", err)
		return
	}
	log.Infoc(collegeCtx, "collegeTagTopArchive  begin")
	s.collegeTagTopArchive(collegeCtx, collegeList)

	log.Infoc(collegeCtx, "college Rank begin")
	err = s.collegeRank(collegeCtx, collegeList, version.Version+1)
	if err != nil {
		log.Errorc(collegeCtx, "s.collegeRank(c) err(%v)", err)
		return
	}
	// 计算完成，版本号+1
	version.Version = version.Version + 1
	version.Time = s.updateTime(collegeCtx, start)
	err = s.college.SetCollegeUpdateVersion(collegeCtx, version)
	if err != nil {
		log.Errorc(collegeCtx, "s.college.SetCollegeUpdateVersion(c) err(%v)", err)
		return
	}
	end := time.Now()
	spend := end.Unix() - start.Unix()
	log.Infoc(collegeCtx, "CollegeRank success() spend(%d)", spend)

}

func (s *Service) updateTime(c context.Context, start time.Time) int64 {
	year := start.Year()   //年
	month := start.Month() //月
	day := start.Day()     //日
	hour := start.Hour()   //小时
	var newHour int64
	if hour < 12 {
		newHour = 12
	} else {
		newHour = 18
	}
	var monthStr, dayStr string
	monthStr = fmt.Sprintf("%02d", month)
	dayStr = fmt.Sprintf("%02d", day)

	timeStr := fmt.Sprintf("%d-%s-%s %2d:00:00", year, monthStr, dayStr, newHour)
	loc, _ := time.LoadLocation("Asia/Shanghai")                       //设置时区
	tt, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, loc) //2006-01-02 15:04:05是转换的格式如php的"Y-m-d H:i:s"
	log.Infoc(c, "version update time (%d) (%s)", tt.Unix(), tt)
	return tt.Unix()
}

func (s *Service) collegeTagTopArchive(c context.Context, collegeList []*college.College) {
	// todo 获取各个学校的稿件
	for k, v := range collegeList {
		collegeList[k].Aids = make([]int64, 0)
		if s.collegeArchiveTopList != nil {
			if archiveList, ok := s.collegeArchiveTopList[v.ID]; ok {
				collegeList[k].Aids = archiveList
			}
		}
	}
}

// collegeRank 学校各个积分排行
func (s *Service) collegeRank(c context.Context, collegeList []*college.College, version int) error {

	for k, v := range collegeList {
		if k%500 == 0 {
			log.Infoc(c, "collegeAllMidRankAndScore index(%d)", k)
		}
		// 计算不同学校的用户排行及获取总分
		score, err := s.collegeAllMidRankAndScore(c, v, version)
		if err != nil {
			log.Errorc(c, "s.collegeAllMidRankAndScore(%d) err(%v)", v.ID, err)
			return err
		}
		collegeList[k].Score = score
	}
	provinceCollegeList := s.provinceAllCollege(c, collegeList)

	// 处理全国排行
	log.Infoc(c, "nationwiderank begin")
	nationwideRank, err := s.collegeNationwideRank(c, collegeList, version)
	if err != nil {
		log.Errorc(c, "s.collegeNationwideRank err(%v)", err)
		return err
	}
	// 处理省排行
	log.Infoc(c, "provincerank begin")
	provinceRank, err := s.collegeProvinceRank(c, provinceCollegeList, version)
	if err != nil {
		log.Errorc(c, "s.collegeProvinceRank err(%v)", err)
		return err
	}
	return s.collegeDetail(c, collegeList, provinceRank, nationwideRank, version)
}

// collegeDetail 学校维度
func (s *Service) collegeDetail(c context.Context, collegeList []*college.College, provinceRank map[int64][]*rank.Redis, nationwideRank []*rank.Redis, version int) error {
	collegeDetail := make([]*college.Detail, 0)
	proviceCollegeMap := make(map[int64]*rank.Redis)
	nationwideCollegeMap := make(map[int64]*rank.Redis)
	for _, v := range provinceRank {
		for _, college := range v {
			proviceCollegeMap[college.Mid] = college
		}
	}
	for _, v := range nationwideRank {
		nationwideCollegeMap[v.Mid] = v
	}
	for _, v := range collegeList {
		nation, nationOk := nationwideCollegeMap[v.ID]
		province, provinceOk := proviceCollegeMap[v.ID]
		detail := &college.Detail{
			Score:       v.Score,
			TabList:     v.TabList,
			ID:          v.ID,
			Name:        v.Name,
			RelationMid: v.RelationMid,
			Province:    v.Province,
		}
		if nationOk {
			detail.NationwideRank = nation.Rank
		}
		if provinceOk {
			detail.ProvinceRank = province.Rank
		}
		if s.collegeTabList != nil {
			if tabList, ok := s.collegeTabList[v.ID]; ok {
				detail.TabList = tabList
			}
		}
		collegeDetail = append(collegeDetail, detail)
	}
	return s.collegeDetailSave(c, collegeDetail, version)
}

// collegeDetailSave ..
func (s *Service) collegeDetailSave(c context.Context, college []*college.Detail, version int) error {
	for _, v := range college {
		err := s.college.SetCollegeDetail(c, v, version)
		if err != nil {
			log.Errorc(c, "s.college.SetCollegeDetail collegeID(%d)", v.ID)
		}
	}
	return nil
}

// collegeNationwideRank 全国排行
func (s *Service) collegeNationwideRank(c context.Context, collegeList []*college.College, version int) ([]*rank.Redis, error) {
	collegeRank := s.collegeRankStruct(c, collegeList, rankTypeNationwide, version)
	redisRank, err := s.collegeRankResultSave(c, rankTypeNationwide, collegeRank, s.collegeRankSid(c, rankTypeNationwide), version)
	if err != nil {
		log.Errorc(c, "s.collegeRankResultSave() rankType (%d) error(%v)", rankTypeNationwide, err)
	}
	return redisRank, err
}

// collegeProvinceRank 省份排行
func (s *Service) collegeProvinceRank(c context.Context, proviceCollegeList map[int64][]*college.College, version int) (map[int64][]*rank.Redis, error) {
	provinceRank := make(map[int64][]*rank.Redis)
	for provinceID, v := range proviceCollegeList {
		provinceCollegeRank := s.collegeRankStruct(c, v, provinceID, version)
		redisRank, err := s.collegeRankResultSave(c, provinceID, provinceCollegeRank, s.collegeRankSid(c, provinceID), version)
		provinceRank[provinceID] = redisRank
		if err != nil {
			log.Errorc(c, "s.collegeRankResultSave() rankType (%d) error(%v)", rankTypeNationwide, err)
			return nil, err
		}
	}
	return provinceRank, nil
}

// collegeRankSid 获取活动id
func (s *Service) collegeRankSid(c context.Context, provinceID int64) int64 {
	sidStr := fmt.Sprintf("%d%03d", s.c.College.CollegeSID, provinceID)
	sid, _ := strconv.ParseInt(sidStr, 10, 64)
	return sid
}

func (s *Service) collegeRankStruct(c context.Context, collegeList []*college.College, rankType int64, version int) []*rank.College {
	data := make(map[int64]*rank.College)
	for _, v := range collegeList {
		data[v.ID] = &rank.College{
			ID:    v.ID,
			Score: v.Score,
			Aids:  v.Aids,
		}
	}
	s.collegeSetCollegeHistoryRank(c, rankType, version, data)
	collegeData := make([]*rank.College, 0)
	for _, v := range data {
		collegeData = append(collegeData, v)
	}
	collegeRankData := rank.CollegeScore{}
	collegeRankData.Data = collegeData
	collegeRankData.TopLength = len(collegeList)
	rank.Sort(&collegeRankData)
	return collegeRankData.Data
}

// collegeRankResultSave 学院排名结果保存
func (s *Service) collegeRankResultSave(c context.Context, rankType int64, collegeScore []*rank.College, sid int64, version int) ([]*rank.Redis, error) {
	eg := errgroup.WithContext(c)
	redisRank := make([]*rank.Redis, 0)
	// redis 存储
	eg.Go(func(ctx context.Context) (err error) {
		redisRank, err = s.collegeRedisCollegeRank(c, rankType, version, collegeScore)
		return err
	})
	// mysql 存储
	eg.Go(func(ctx context.Context) error {
		return s.collegeDBCollegeRank(c, collegeScore, sid)
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, err
	}
	return redisRank, nil
}

// getAllCollege 获取所有学校
func (s *Service) getAllCollege(c context.Context) (collegeList []*college.College, err error) {
	var offset int64
	collegeList = make([]*college.College, 0)
	for {
		college, err := s.college.GetAllCollege(c, offset, allCollegeLimit)
		if err != nil {
			log.Errorc(c, "s.college.GetAllCollege error(%v)", err)
			return nil, err
		}
		if len(college) > 0 {
			collegeList = append(collegeList, college...)
		}
		if len(college) < allCollegeLimit {
			break
		}
		offset += allCollegeLimit
	}

	return collegeList, nil
}

// provinceAllCollege 分省处理
func (s *Service) provinceAllCollege(c context.Context, collegeList []*college.College) map[int64][]*college.College {
	var provinceCollege = make(map[int64][]*college.College, 0)
	if collegeList != nil {
		for _, v := range collegeList {
			if _, ok := provinceCollege[v.ProvinceID]; !ok {
				provinceCollege[v.ProvinceID] = make([]*college.College, 0)
			}
			provinceCollege[v.ProvinceID] = append(provinceCollege[v.ProvinceID], v)
		}
	}
	return provinceCollege
}

// collegeRedisCollegeRank redis rank data
func (s *Service) collegeRedisCollegeRank(c context.Context, rankType int64, version int, collegeScore []*rank.College) (rankRedis []*rank.Redis, err error) {
	rankRedis = make([]*rank.Redis, 0)
	for i, v := range collegeScore {
		if v != nil {
			rankRedis = append(rankRedis, &rank.Redis{
				Mid:   v.ID,
				Score: v.Score,
				Rank:  i + 1,
				Aids:  v.Aids,
			})
		}
	}
	if len(rankRedis) > 0 {
		err = s.rank.SetRank(c, s.collegeCollegeRankKey(c, rankType, version), rankRedis)
		if err != nil {
			log.Errorc(c, "s.rank.SetRank(%s) error(%v)", s.collegeCollegeRankKey(c, rankType, version), err)
			err = errors.Wrapf(err, "s.SetRank")
		}
	}
	return rankRedis, err
}

func (s *Service) collegeDbRankInsert(c context.Context, rankDb []*rank.DB) (err error) {
	var times int
	patch := collegeInfoBatch
	concurrency := concurrencycollege
	times = len(rankDb) / patch / concurrency
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			b := batch
			i := index
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(rankDb) {
					return nil
				}
				reqMids := rankDb[start:]
				end := start + patch
				if end < len(rankDb) {
					reqMids = rankDb[start:end]
				}
				if len(reqMids) > 0 {
					err = s.rank.BatchAddRank(c, reqMids)
					if err != nil {
						err = errors.Wrapf(err, "s.college.BatchAddRank")
						return err
					}
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Errorc(c, "eg.Wait error(%v)", err)
			return err
		}
	}
	return nil
}

// collegeDBCollegeRank redis rank data
func (s *Service) collegeDBCollegeRank(c context.Context, collegeScore []*rank.College, sid int64) (err error) {
	rankDb := make([]*rank.DB, 0)
	mids := make([]int64, 0)
	for _, v := range collegeScore {
		mids = append(mids, v.ID)
	}
	hourString := time.Now().Format("2006010215")
	hour, _ := strconv.ParseInt(hourString, 10, 64)
	for i, v := range collegeScore {
		if v != nil {
			rank := &rank.DB{
				Mid:   v.ID,
				Score: v.Score,
				Rank:  i + 1,
				Batch: hour,
				SID:   sid,
			}
			rankDb = append(rankDb, rank)
		}
	}
	if len(rankDb) > 0 {
		err = s.collegeDbRankInsert(c, rankDb)
		if err != nil {
			log.Errorc(c, "s.rank.BatchAddRank(%v) error(%v)", mids, err)
			err = errors.Wrapf(err, "s.rank.BatchAddRank %v", mids)
		}
	}
	return err
}

func (s *Service) collegeCollegeRankKey(c context.Context, rankType int64, version int) string {
	return fmt.Sprintf("college:%d:%d", version, rankType)
}

func (s *Service) collegeSetCollegeHistoryRank(c context.Context, rankType int64, version int, mapCollegeInfo map[int64]*rank.College) error {
	historyRank, err := s.rank.GetRank(c, s.collegeCollegeRankKey(c, rankType, version-1))
	if err != nil {
		err = errors.Wrapf(err, "s.rank.GetRank")
		return err
	}
	if historyRank != nil {
		for _, v := range historyRank {
			if _, ok := (mapCollegeInfo)[v.Mid]; ok {
				(mapCollegeInfo)[v.Mid].History = v.Rank
			}
		}
	}
	return nil
}

// getWaitTime 返回等待时间
func (s *Service) getWaitTime(startTime, endTime int64) int64 {
	diff := endTime - startTime
	if diff >= oneSecond {
		return 0
	}
	return oneSecond - diff
}

// CollegeVersion 学院排行
func (s *Service) CollegeVersion() {

	s.collegeVersionUpdateRunning.Lock()
	defer s.collegeVersionUpdateRunning.Unlock()
	collegeVersionCtxInit()
	// 获取本次脚本的版本号
	version, err := s.college.GetCollegeUpdateVersion(collegeVersionCtx)
	if err != nil {
		log.Errorc(collegeVersionCtx, "s.college.GetCollegeUpdateVersion err (%v) version(%d)", err, version)
		return
	}
	now := time.Now().Unix()
	if s.c.College.VersionTest == 1 || (version.Version > 0 && version.Time < now) {
		err = s.college.SetCollegeVersion(collegeVersionCtx, version)
		if err != nil {
			log.Errorc(collegeVersionCtx, "s.college.SetCollegeVersion err (%v) version(%d)", err, version)
			return
		}
		log.Infoc(collegeVersionCtx, "CollegeUpdateVersion update()")
	}
	log.Infoc(collegeVersionCtx, "CollegeUpdateVersion success()")
}
