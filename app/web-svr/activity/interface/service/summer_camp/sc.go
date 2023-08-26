package summer_camp

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	actApi "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/model/summer_camp"
	"go-gateway/app/web-svr/activity/interface/rewards"
	"go-gateway/pkg/idsafe/bvid"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	act_plat "git.bilibili.co/bapis/bapis-go/platform/common/act-plat"
	actPlat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
)

const (
	JOIN      = 1
	QUIT_JOIN = 2

	JOIN_STATUS      = 1
	QUIT_JOIN_STATUS = 0
)

func (s *Service) GetLotteryActivityID() string {
	return s.c.SummerCampConf.ActivityIdRW
}

func (s *Service) GetLotterySid() string {
	return s.c.SummerCampConf.LotteryPoolId
}

func (s *Service) GetLotteryCid() int64 {
	return s.c.SummerCampConf.LotteryCid
}

func (s *Service) GetLotteryActionType() int {
	return int(s.c.SummerCampConf.LotteryActionType)
}

// UseLotteryPoint
func (s *Service) UseLotteryPoint(ctx context.Context, mid int64, orderID, lotteryID, activityID string) (err error) {
	return s.costPointDao.UserCostForLottery(ctx, mid, orderID, lotteryID, activityID)
}

// GetUserTaskInfo获取用户任务相关数据(打卡天数、我的积分、视频任务、分享任务、投稿任务)
func (s *Service) GetUserTaskInfo(c context.Context, mid int64) (*summer_camp.UserInfoRes, error) {
	var res = &summer_camp.UserInfoRes{IsJoin: 1}
	taskInfo := new(summer_camp.TaskInfo)
	activityId := s.c.SummerCampConf.ActivityIdRW
	// 分别获取打卡天数、我的积分、视频任务、分享任务、投稿任务
	eg := errgroup.WithContext(c)
	// 获取用户打卡天数(不连续)
	eg.Go(func(ctx context.Context) (err error) {
		signDays, err := s.TaskTotalProgress(ctx, mid, activityId)
		res.SignDays = signDays
		return
	})
	// 获取用户我的积分
	eg.Go(func(ctx context.Context) (err error) {
		totalLeft, _, _, err := s.costPointDao.GetUserTotalPoint(ctx, mid, activityId)
		if err != nil {
			return
		}
		taskInfo.TotalPoint = totalLeft
		return
	})
	// 获取用户视频任务情况
	eg.Go(func(ctx context.Context) (err error) {
		totalPoints, err := s.TaskCounterRes(ctx, mid, activityId, s.c.SCCounterCourseMap["view"])
		taskInfo.ViewVideosTask = totalPoints
		return
	})
	// 获取用户分享任务情况
	eg.Go(func(ctx context.Context) (err error) {
		points, err := s.TaskCounterRes(ctx, mid, activityId, s.c.SCCounterCourseMap["share"])
		taskInfo.ShareTask = points
		return
	})
	// 获取用户投稿任务情况
	eg.Go(func(ctx context.Context) (err error) {
		points, err := s.TaskCounterRes(ctx, mid, activityId, s.c.SCCounterCourseMap["archive"])
		taskInfo.TougaoTask = points
		return
	})

	if err := eg.Wait(); err != nil {
		log.Errorc(c, "GetUserTaskInfo err,err is (%v).", err)
		return nil, err
	}
	res.TaskInfo = taskInfo
	return res, nil

}

// TaskTotalProgress 用户打卡任务完成历史数据
func (s *Service) TaskTotalProgress(ctx context.Context, mid int64, activityId string) (int64, error) {
	var (
		signDays   int64 = 0
		countReply *actPlat.GetTotalResResp
	)
	countReply, err := client.ActPlatClient.GetTotalRes(ctx, &actPlat.GetTotalResReq{
		Activity: activityId,
		Counter:  "view_day",
		Mid:      mid,
	})
	if err != nil {
		log.Errorc(ctx, "SummerCamp get grpc client.ActPlatClient.GetTotalRes() mid(%d) error(%+v)", mid, err)
		return signDays, err
	}
	if countReply == nil {
		log.Warnc(ctx, "SummerCamp get grpc client.ActPlatClient.GetTotalRes() mid(%d) historyReply is nil", mid)
		return signDays, nil
	}

	signDays = countReply.Total
	return signDays, nil
}

// TaskCounterRes 任务当天完成情况
func (s *Service) TaskCounterRes(ctx context.Context, mid int64, activityId string, counter string) (int64, error) {
	var (
		start []byte
		res   int64 = 0
		now         = time.Now().Unix()
	)
	for {
		var (
			countReply *actPlat.GetCounterResResp
		)
		countReply, err := client.ActPlatClient.GetCounterRes(ctx, &actPlat.GetCounterResReq{
			Activity: activityId,
			Counter:  counter,
			Mid:      mid,
			Time:     now,
			Start:    start,
		})
		if err != nil {
			log.Errorc(ctx, "SummerCamp grpc client.ActPlatClient.GetCounterRes() mid(%d) error(%+v)", mid, err)
			return res, err
		}
		if countReply == nil || countReply.CounterList == nil {
			log.Warnc(ctx, "SummerCamp grpc client.ActPlatClient.GetCounterRes() mid(%d) historyReply is nil", mid)
			return res, nil
		}
		// 计算已经打卡天数
		for _, v := range countReply.CounterList {
			res += v.Val
		}

		start = countReply.Next
		if len(start) == 0 || countReply.Next == nil {
			break
		}
	}
	return res, nil
}

// GetCourseList 获取课程列表
func (s *Service) GetCourseList(ctx context.Context, pn, ps int) (*summer_camp.CourseListRes, error) {
	res := &summer_camp.CourseListRes{}
	// 读缓存
	dbCourseList, err := s.summerCampDao.CacheGetCourseList(ctx)
	if err == nil && dbCourseList != nil {
		res.Page = pn
		res.Size = len(dbCourseList)
		res.Total = len(dbCourseList)
		for _, v := range dbCourseList {
			tmp := &summer_camp.OneCourseRes{
				ID:         v.CourseID,
				CourseName: v.CourseTitle,
				CourseIcon: v.PicCover,
				BodanId:    v.BodanId,
			}
			res.List = append(res.List, tmp)
		}
		return res, nil
	} else {
		log.Errorc(ctx, "SummerCamp service.GetCourseList get CacheGetCourseList err,error is (%v).or res is nil.", err)
	}
	// 回源，去db中取数据
	offset := (pn - 1) * ps
	limit := ps
	dbList, err := s.summerCampDao.GetCourseList(ctx, offset, limit)
	if err != nil {
		log.Errorc(ctx, "SummerCamp service.GetCourseList err,error is (%v).", err)
		return nil, ecode.GetPlanListErr
	}
	if dbList != nil {
		for _, one := range dbList {
			tmp := &summer_camp.OneCourseRes{
				ID:         one.CourseID,
				CourseName: one.CourseTitle,
				CourseIcon: one.PicCover,
				BodanId:    one.BodanId,
			}
			res.List = append(res.List, tmp)
		}
		res.Page = pn
		res.Size = len(dbList)
		res.Total = len(dbList)
	}
	// 写缓存
	err = s.summerCampDao.CacheSetCourseList(ctx, dbList)
	if err != nil {
		log.Errorc(ctx, "SummerCamp service.summerCampDao.CacheSetCourseList err,error is (%v).", err)
	}
	return res, nil
}

// UserCourseInfo 获取用户课程列表情况
func (s *Service) UserCourseInfo(ctx context.Context, mid int64, pn, ps int) (*summer_camp.UserCourseInfoRes, error) {
	res := &summer_camp.UserCourseInfoRes{}
	activityId := s.c.SummerCampConf.ActivityIdRW
	// 用户选课列表
	offset := (pn - 1) * ps
	limit := ps
	userCourse, courseIds, err := s.summerCampDao.GetUserCourse(ctx, mid, offset, limit)
	if err != nil {
		return nil, ecode.SummerCampUserCourseInfoErr
	}
	if userCourse != nil && courseIds != nil {
		var (
			courseUserJoinTimeMap = make(map[int64]int64)
			courseDetailMap       = make(map[int64]*summer_camp.OneCourseRes)
			userCourseSignDaysRes = make(map[int64]int)
			courseTotalDaysMap    = make(map[int64]int)
			l1                    = sync.Mutex{}
		)
		for _, course := range userCourse {
			courseUserJoinTimeMap[course.CourseID] = int64(course.JoinTime)

		}
		// 获取课程总天数map
		courseBodanRes, courseDetailMap, err := s.getCourseBodanMap(ctx)
		if err != nil {
			return nil, ecode.SummerCampUserCourseInfoErr
		}
		if courseBodanRes != nil {
			for courseId, bodans := range courseBodanRes {
				courseTotalDaysMap[courseId] = len(bodans)
			}
		}
		// 获取每门课程的打卡情况map
		eg := errgroup.WithContext(ctx)
		for _, courseId := range courseIds {
			tmpCourseId := courseId
			// 获取用户打卡天数(不连续)
			eg.Go(func(ctx context.Context) (err error) {
				if v, ok := s.c.SCCounterCourseMap[strconv.FormatInt(tmpCourseId, 10)]; ok {
					signDays, _, err := s.TaskHistoryDays(ctx, mid, activityId, v,
						courseUserJoinTimeMap[tmpCourseId], false)
					if err != nil {
						return err
					}
					l1.Lock()
					userCourseSignDaysRes[tmpCourseId] = signDays
					l1.Unlock()
				}
				return
			})
		}
		if err := eg.Wait(); err != nil {
			return nil, err
		}

		// 填充res
		res.Page = pn
		res.Size = len(userCourse)
		res.Total = len(userCourse)
		newList := make([]*summer_camp.UserCourseSignInfo, 0)
		for _, course := range userCourse {
			isFinished := false
			if userCourseSignDaysRes[course.CourseID] >= courseTotalDaysMap[course.CourseID] {
				// 说明用户已结业
				isFinished = true
			}
			tmp := &summer_camp.UserCourseSignInfo{
				CourseId:   course.CourseID,
				CourseName: course.CourseTitle,
				CourseIcon: courseDetailMap[course.CourseID].CourseIcon,
				JoinTime:   course.JoinTime,
				CourseDays: courseTotalDaysMap[course.CourseID],
				SignedDays: userCourseSignDaysRes[course.CourseID],
				IsFinished: isFinished,
			}
			res.List = append(res.List, tmp)
		}
		// 对返回列表进行排序
		newList = sortUserCourseList(res.List)
		res.List = newList

	}
	return res, nil

}

func sortUserCourseList(list []*summer_camp.UserCourseSignInfo) (res []*summer_camp.UserCourseSignInfo) {
	if list == nil {
		return
	}
	res = make([]*summer_camp.UserCourseSignInfo, 0)
	// 排序逻辑：1、用户打卡天数高的排前面，2、最新选择的的排前面
	// 已经结业的单独拿出来
	finishedCourse := make([]*summer_camp.UserCourseSignInfo, 0)
	unFinishedCourse := make([]*summer_camp.UserCourseSignInfo, 0)
	for _, one := range list {
		if one.IsFinished == true {
			finishedCourse = append(finishedCourse, one)
		} else {
			unFinishedCourse = append(unFinishedCourse, one)
		}

	}
	//sortBySignedDays(unFinishedCourse, 0, len(unFinishedCourse)-1)
	//sortBySignedDays(finishedCourse, 0, len(finishedCourse)-1)
	sort.Slice(unFinishedCourse, func(i, j int) bool {
		return unFinishedCourse[i].SignedDays > unFinishedCourse[j].SignedDays && unFinishedCourse[i].JoinTime > unFinishedCourse[j].JoinTime
	})
	sort.Slice(finishedCourse, func(i, j int) bool {
		return unFinishedCourse[i].SignedDays > unFinishedCourse[j].SignedDays && unFinishedCourse[i].JoinTime > unFinishedCourse[j].JoinTime
	})

	res = append(res, unFinishedCourse...)
	res = append(res, finishedCourse...)
	return

}

func sortBySignedDays(arr []*summer_camp.UserCourseSignInfo, first, last int) {
	flag := first
	left := first
	right := last

	if first >= last {
		return
	}
	// 将大于arr[flag]的都放在右边，小于的，都放在左边
	for first < last {
		// 如果flag从左边开始，那么是必须先从有右边开始比较，也就是先在右边找比flag小的
		for first < last {
			if arr[last].SignedDays <= arr[flag].SignedDays && arr[last].JoinTime <= arr[flag].JoinTime {
				last--
				continue
			}
			// 交换数据
			arr[last], arr[flag] = arr[flag], arr[last]
			flag = last
			break
		}
		for first < last {
			if arr[first].SignedDays >= arr[flag].SignedDays && arr[first].JoinTime >= arr[flag].JoinTime {
				first++
				continue
			}
			arr[first], arr[flag] = arr[flag], arr[first]
			flag = first
			break
		}
	}

	sortBySignedDays(arr, left, flag-1)
	sortBySignedDays(arr, flag+1, right)
}

// TaskHistoryDays 任务完成情况
func (s *Service) TaskHistoryDays(ctx context.Context, mid int64, activityId string, counter string, joinTime int64, yesterday bool) (int, []*act_plat.HistoryContent, error) {
	var (
		start []byte
		days  int = 0
		day       = make(map[string]struct{}, 0)
		list      = make([]*act_plat.HistoryContent, 0)
	)
	// 今天凌晨时间节点
	todayZero := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 00, 0, 0, 0, time.Local).Unix()
	for {
		var (
			countReply *actPlat.GetHistoryResp
		)
		countReply, err := client.ActPlatClient.GetHistory(ctx, &actPlat.GetHistoryReq{
			Activity: activityId,
			Counter:  counter,
			Mid:      mid,
			Start:    start,
		})
		if err != nil {
			log.Errorc(ctx, "SummerCamp grpc client.ActPlatClient.GetHistory() mid(%d) error(%+v)", mid, err)
			return days, nil, err
		}
		if countReply == nil || countReply.History == nil {
			log.Warnc(ctx, "SummerCamp grpc client.ActPlatClient.GetHistory() mid(%d) history is nil", mid)
			return days, nil, nil
		}
		// 计算已经打卡天数
		for _, v := range countReply.History {
			if yesterday == true {
				if v.Timestamp > joinTime && v.Timestamp < todayZero {
					day[time.Unix(v.Timestamp, 0).Format("20060102")] = struct{}{}
				}
			} else {
				if v.Timestamp > joinTime {
					day[time.Unix(v.Timestamp, 0).Format("20060102")] = struct{}{}
				}
			}
			list = append(list, v)

		}
		days += len(day)
		start = countReply.Next
		if len(start) == 0 || countReply.Next == nil {
			break
		}
	}
	return days, list, nil
}

// getCourseBodanMap 每个课程对应的全部播单数据
func (s *Service) getCourseBodanMap(ctx context.Context) (courseBodansMap map[int64][]int64, courseDetailMap map[int64]*summer_camp.OneCourseRes, err error) {
	courseList, err := s.GetCourseList(ctx, 1, 50)
	if err != nil {
		log.Errorc(ctx, "SummerCamp service.getCourseMap err,error is (%v).", err)
		return nil, nil, err
	}

	courseBodansMap = make(map[int64][]int64)
	courseDetailMap = make(map[int64]*summer_camp.OneCourseRes)
	if courseList != nil {
		for _, oneCourse := range courseList.List {
			mlids := make([]int64, 0)
			bodanArr := strings.Split(oneCourse.BodanId, "-")
			for _, bodanId := range bodanArr {
				tmp, err := strconv.ParseInt(bodanId, 10, 64)
				if err != nil {
					return nil, nil, ecode.StringToInt64Err
				}
				mlids = append(mlids, tmp)
			}
			courseBodansMap[oneCourse.ID] = mlids
			courseDetailMap[oneCourse.ID] = oneCourse
		}
	}
	return
}

// UserTodayVideosById 获取用户课程今天应该展示的课程课程列表
func (s *Service) UserTodayVideosById(ctx context.Context, mid int64, courseId int64, pn, ps int) (res *summer_camp.CourseBodanList, err error) {
	activityId := s.c.SummerCampConf.ActivityIdRW
	// 判断用户是否报名该课程
	isJoinCourse, err := s.summerCampDao.GetUserCourseById(ctx, mid, courseId)
	if err != nil {
		return nil, err
	}
	if isJoinCourse == nil || isJoinCourse.ID == 0 {
		return nil, ecode.SCUserNotJoinCourse
	}
	// 查询用户课程任务完成天数(排除今天)
	signDays, _, err := s.TaskHistoryDays(ctx, mid, activityId, s.c.SCCounterCourseMap[strconv.FormatInt(courseId, 10)],
		int64(isJoinCourse.JoinTime), true)
	if err != nil {
		return nil, err
	}

	// 根据courseId查询天数+1天课程播单
	courseBodanMap, _, err := s.getCourseBodanMap(ctx)
	if err != nil {
		return nil, ecode.SummerCampUserCourseInfoErr
	}
	var mlid int64
	if v, ok := courseBodanMap[courseId]; ok {
		// 返回打卡天数+1的播单，signDays = 下标
		if signDays <= len(v)-1 {
			mlid = v[signDays]
		} else {
			return nil, ecode.SCcourseNotExit
		}
	} else {
		return nil, ecode.SCcourseNotExit
	}

	// 获取播单视频
	res, err = s.getMlidVideos(ctx, mlid, mid, pn, ps)

	return

}

func (s *Service) getMlidVideos(ctx context.Context, mlid int64, mid int64, pn, ps int) (res *summer_camp.CourseBodanList, err error) {
	res = &summer_camp.CourseBodanList{}
	// grpc获取播单详情 type=2代表视频类收藏夹
	reply, err := s.favDao.Folders(ctx, []int64{mlid}, 2)
	if err != nil {
		log.Errorc(ctx, "SummerCamp service.getMlidVideos grpc s.favDao.Folders err!error is (%v)", err)
		return nil, ecode.SCcourseNotExit
	}
	if reply.Res == nil || len(reply.Res) <= 0 {
		log.Errorc(ctx, "SummerCamp service.getMlidVideos  grpc s.favDao.Folders is nil,mlid is:(%v)", mlid)
		return nil, ecode.SCcourseNotExit
	}
	folder := reply.Res[0]
	var (
		aids []int64
		// 排序用
		aidsSortMap = map[int64]*summer_camp.VideoDetail{}
	)

	res.BodanTitle = folder.Name
	res.BodanDesc = folder.Description

	// grpc获取播单里的视频列表
	fvideos, err := s.favDao.FavoritesAll(ctx, 2, mid, folder.Mid, folder.ID, int32(pn), int32(ps))
	if err != nil {
		log.Errorc(ctx, "service.getMlidVideos & get s.favDao.FavoritesAll err!"+
			"error is (%v),mid is (%v),uid is (%v),folderid is (%v)", err, mid, folder.Mid, folder.ID)
		return nil, err
	}
	if fvideos.Res.List == nil {
		log.Errorc(ctx, "service.getMlidVideos & get fav.FavoritesAll result is nil")
		return nil, nil
	}
	for _, v := range fvideos.Res.List {
		aids = append(aids, v.Oid)
	}
	// 获取视频信息
	var archive map[int64]*api.Arc
	if len(aids) > 0 {
		archive, err = s.archive.AllArchiveInfo(ctx, aids)
		if err != nil {
			log.Errorc(ctx, "service.getMlidVideos.getAllArchiveInfo err(%v)", err)
			return nil, err
		}
	}
	// 填充视频字段
	for aid, v := range archive {
		bvidStr, _ := bvid.AvToBv(v.Aid)
		aidsSortMap[aid] = &summer_camp.VideoDetail{
			Bvid:     bvidStr,
			Title:    v.Title,
			Duration: v.Duration,
			Cover:    v.Pic,
			CntInfo: summer_camp.VideoCnt{
				Collect: v.Stat.Fav,
				Play:    v.Stat.View,
				Danmaku: v.Stat.Danmaku,
			},
			Link: v.ShortLinkV2,
		}
	}
	for _, v := range aids {
		res.List = append(res.List, aidsSortMap[v])
	}
	res.Total = len(res.List)
	res.Page = pn
	res.Size = len(res.List)
	return res, nil
}

// UserOneDayVideosById获取用户某一天的视频列表
func (s *Service) UserOneDayVideosById(ctx context.Context, mid int64, courseId int64, day, pn, ps, showTab int) (res *summer_camp.CourseBodanList, err error) {
	activityId := s.c.SummerCampConf.ActivityIdRW
	// 判断用户是否报名该课程
	isJoinCourse, err := s.summerCampDao.GetUserCourseById(ctx, mid, courseId)
	if err != nil {
		log.Errorc(ctx, "UserOneDayVideosById summerCampDao.GetUserCourseById err:(%v)", err)
		return nil, ecode.SCUserNotJoinCourse
	}
	if isJoinCourse == nil {
		return nil, ecode.SCUserNotJoinCourse
	}
	// 查询用户课程任务完成天数(排除今天)
	signDays, _, err := s.TaskHistoryDays(ctx, mid, activityId, s.c.SCCounterCourseMap[strconv.FormatInt(courseId, 10)],
		int64(isJoinCourse.JoinTime), true)
	if err != nil {
		log.Errorc(ctx, "service.UserOneDayVideosById.TaskHistoryDays err:(%v)", err)
		return nil, ecode.SCTaskErr
	}
	// 如果请求的day>打卡天数=无权限查看
	if day > signDays+1 {
		log.Errorc(ctx, "service.UserOneDayVideosById user has no right,mid is(%v), signDays is(%v)", mid, signDays)
		return nil, ecode.SCUserNoRight
	}

	// 根据courseId查询第day天课程播单
	courseBodanMap, _, err := s.getCourseBodanMap(ctx)
	if err != nil {
		return nil, ecode.SummerCampUserCourseInfoErr
	}
	var mlid int64
	if v, ok := courseBodanMap[courseId]; ok {
		// 返回第day天播单，day-1 = 下标
		if day <= len(v) {
			mlid = v[day-1]
		} else {
			return nil, ecode.SCcourseNotExit
		}
	} else {
		return nil, ecode.SCcourseNotExit
	}

	// 获取播单视频
	res, err = s.getMlidVideos(ctx, mlid, mid, pn, ps)
	// 是否需要展示往期播单title
	if showTab == 1 {

		if day == 1 {
			res.TabList = append(res.TabList, &summer_camp.TabBodan{
				Title: res.BodanTitle,
				Desc:  res.BodanDesc,
			})
		} else {
			mlids := make([]int64, 0)
			for k, v := range courseBodanMap[courseId] {
				if k+1 > day {
					break
				}
				mlids = append(mlids, v)
			}
			tabBodansMap, err := s.getMlidsInfo(ctx, mlids)
			if err != nil {
				return nil, err
			}
			// 填充res
			for _, v := range mlids {
				res.TabList = append(res.TabList, tabBodansMap[v])
			}
		}
	}

	return

}

func (s *Service) getMlidsInfo(ctx context.Context, mlids []int64) (res map[int64]*summer_camp.TabBodan, err error) {
	res = make(map[int64]*summer_camp.TabBodan)
	// grpc获取播单详情 type=2代表视频类收藏夹
	reply, err := s.favDao.Folders(ctx, mlids, 2)
	if err != nil {
		log.Errorc(ctx, "SummerCamp service.getMlidsInfo grpc s.favDao.Folders err!error is (%v)", err)
		return nil, ecode.SCcourseNotExit
	}
	if reply.Res == nil || len(reply.Res) <= 0 {
		log.Errorc(ctx, "SummerCamp service.getMlidsInfo grpc s.favDao.Folders is nil,mlids is:(%v)", mlids)
		return nil, ecode.SCcourseNotExit
	}
	for _, folder := range reply.Res {
		res[folder.Mlid] = &summer_camp.TabBodan{
			Title: folder.Name,
			Desc:  folder.Description,
		}

	}
	return
}

// StartPlan 用户开启计划
func (s *Service) StartPlan(ctx context.Context, mid int64, courseIds []int64, reserveId int64) (res *summer_camp.StartPlanRes, err error) {
	res = &summer_camp.StartPlanRes{RewardPoint: int(s.c.SummerCampConf.StartPlanPoint)}
	if reserveId != s.c.SummerCampConf.ReserveId {
		return res, ecode.SCActivityIdErr
	}
	// 判断redis标记位

	// 判断用户是否预约&保证加入预约
	err = s.likeSrv.AsyncReserve(ctx, reserveId, mid, 1, new(likemdl.ReserveReport))
	if err != nil {
		if err == ecode.ActivityRepeatSubmit { // 已经预约
			err = nil
		} else {
			// 重试
			return res, ecode.SystemNetWorkBuzyErr
		}
	}
	// 并发-观看任务报名，分享、投稿报名
	eg := errgroup.WithContext(ctx)
	// 找到courseId对应的counter
	counters := make([]string, 0)
	for _, courseId := range courseIds {
		if v, ok := s.c.SCCounterCourseMap[strconv.FormatInt(courseId, 10)]; ok {
			counters = append(counters, v)
		}

	}
	for _, counter := range counters {
		counterT := counter
		eg.Go(func(ctx context.Context) (err error) {
			// 报名观看任务 过期半年
			_, err = client.ActPlatClient.AddFilterMemberInt(ctx, &actPlat.SetFilterMemberIntReq{
				Activity: s.c.SummerCampConf.ActivityIdRW,
				Counter:  counterT,
				Filter:   "filter_mid",
				Values:   []*actPlat.FilterMemberInt{{Value: mid, ExpireTime: 15768000}},
			})
			if err != nil {
				log.Errorc(ctx, "SummerCamp StartPlan client.ActPlatClient.AddFilterMemberInt mid(%d) err(%v)", mid, err)
				return err
			}
			return
		})
	}
	// 分享、投稿报名 过期半年
	eg.Go(func(ctx context.Context) (err error) {
		_, err = client.ActPlatClient.AddSetMemberInt(ctx, &actPlat.SetMemberIntReq{
			Activity: s.c.SummerCampConf.ActivityIdRW,
			Name:     "filter_mid",
			Values:   []*actPlat.SetMemberInt{{Value: mid, ExpireTime: 15768000}},
		})
		return
	})

	if err := eg.Wait(); err != nil {
		return res, ecode.SCTaskSetErr
	}
	// 插入用户课程表
	records := make([]*summer_camp.DBUserCourse, 0)
	_, courseInfoMap, err := s.getCourseBodanMap(ctx)
	if err != nil {
		return nil, err
	}
	for _, courseId := range courseIds {
		if v, ok := courseInfoMap[courseId]; ok {
			tmp := &summer_camp.DBUserCourse{
				Mid:         mid,
				CourseID:    courseId,
				CourseTitle: v.CourseName,
				Status:      1,
			}
			records = append(records, tmp)
		}
	}
	rowsNum, errG := s.summerCampDao.MultiInsertUserCourse(ctx, records)
	if errG != nil || rowsNum <= 0 {
		log.Errorc(ctx, "SummerCamp StartPlan summerCampDao.MultiInsertUserCourse mid(%d) err(%v)", mid, errG)
		if errG != nil && strings.Contains(errG.Error(), "Duplicate entry") {
			return
		}
		return nil, ecode.SCStartCourseDBErr

	}
	// 添加redis标记位

	return

}

// JoinCourse
func (s *Service) JoinCourse(ctx context.Context, mid int64, courseIds []int64, typ int) (res *summer_camp.StartPlanRes, err error) {
	// 1.加db
	records := make([]*summer_camp.DBUserCourse, 0)
	_, courseInfoMap, err := s.getCourseBodanMap(ctx)
	if err != nil {
		return nil, err
	}
	var (
		status         = JOIN_STATUS
		rowsNum        = int64(0)
		errG           error
		taskExpireTime int64 = 15768000
	)
	if typ == QUIT_JOIN {
		// db
		status = QUIT_JOIN_STATUS
		// task
		taskExpireTime = 1
	}
	for _, courseId := range courseIds {
		if v, ok := courseInfoMap[courseId]; ok {
			tmp := &summer_camp.DBUserCourse{
				Mid:         mid,
				CourseID:    courseId,
				CourseTitle: v.CourseName,
				Status:      status,
			}
			records = append(records, tmp)
		}
	}
	if typ == JOIN {
		rowsNum, errG = s.summerCampDao.MultiInsertOrUpdateUserCourse(ctx, mid, records)
	} else {
		rowsNum, errG = s.summerCampDao.SingleQuitJoin(ctx, mid, records[0])
	}

	if errG != nil {
		log.Errorc(ctx, "SummerCamp JoinCourse summerCampDao.MultiInsertUserCourse or "+
			"SingleQuitJoin err.mid(%d) err(%v),typ is (%v)", mid, errG, typ)
		if strings.Contains(errG.Error(), "Duplicate entry") {
			return
		}
		return nil, ecode.SCStartCourseDBErr

	}
	if rowsNum <= 0 {
		if typ == JOIN {
			log.Errorc(ctx, "SummerCamp JoinCourse rows lessthan zero,summerCampDao.MultiInsertUserCourse or "+
				"SingleQuitJoin err.mid(%d) err(%v),typ is (%v)", mid, errG, typ)
			return nil, ecode.SCStartCourseDBErr
		} else {
			return
		}
	}
	// 2.修改任务计入/退出
	eg := errgroup.WithContext(ctx)
	// 找到courseId对应的counter
	counters := make([]string, 0)
	for _, courseId := range courseIds {
		if v, ok := s.c.SCCounterCourseMap[strconv.FormatInt(courseId, 10)]; ok {
			counters = append(counters, v)
		}

	}
	for _, counter := range counters {
		counterT := counter
		eg.Go(func(ctx context.Context) (err error) {
			// 报名观看任务 过期半年
			_, err = client.ActPlatClient.AddFilterMemberInt(ctx, &actPlat.SetFilterMemberIntReq{
				Activity: s.c.SummerCampConf.ActivityIdRW,
				Counter:  counterT,
				Filter:   "filter_mid",
				Values:   []*actPlat.FilterMemberInt{{Value: mid, ExpireTime: taskExpireTime}},
			})
			if err != nil {
				log.Errorc(ctx, "SummerCamp JoinCourse client.ActPlatClient.AddFilterMemberInt mid(%d) err(%v)", mid, err)
				return err
			}
			return
		})
	}
	if err := eg.Wait(); err != nil {
		return res, ecode.SCTaskSetErr
	}
	return
}

// ExchangeAward 积分兑换奖品
func (s *Service) ExchangeAward(ctx context.Context, mid int64, activityId int64, awardId string) (err error) {
	// 校验
	if activityId != s.c.SummerCampConf.ReserveId {
		err = ecode.SCActivityIdErr
		return
	}
	// 积分兑换奖品
	activityIdStr := s.c.SummerCampConf.ActivityIdRW
	todayDate := time.Now().Format("20060102")
	// 唯一标识，幂等
	orderId := fmt.Sprintf("%d_%s_%s", mid, todayDate, awardId)
	err = s.costPointDao.UserCostForExchange(ctx, activityIdStr, awardId, mid, orderId)
	if err != nil {
		return
	}
	// 发放奖励 todo 后期拆出异步 binlog+databus
	awardInt, err := strconv.ParseInt(awardId, 10, 64)
	if err != nil {
		log.Errorc(ctx, "ExchangeAward stringtoint64 err,err is :(%v),awardId is :%s", err, awardId)
		err = ecode.ExchangeErr
		return
	}
	rewardOrderID := s.md5(fmt.Sprintf("%s_%s", s.c.SummerCampConf.ReserveId, orderId))
	log.Infoc(ctx, "SummerCamp ExchangeAward send award mid (%d) orderID(%v)", mid, rewardOrderID)
	_, err = rewards.Client.SendAwardByIdAsync(ctx, mid, rewardOrderID, "reward_conf", awardInt, true, true)
	if err != nil {
		log.Errorc(ctx, "ExchangeAward rewards.Client.SendAwardByIdAsync error: mid: %v, uniqueId: %v, err: %v", mid, rewardOrderID, err)
		return
	}
	return

}

const (
	ObtainType        = 0 // 获得积分
	CostTypeLottery   = 1 // 抽奖消耗
	CostTypeExchange  = 2 // 积分兑换消耗
	ObtainViewName    = "完成看视频任务"
	ObtainShareName   = "完成分享成绩任务"
	ObtainArchiveName = "完成投稿任务"
	ObtainNew         = "新人奖励"
	ObtainAward       = "额外奖励"
	CostLotteryName   = "参与抽奖"
)

// UserPointHistory 用户积分明细
func (s *Service) UserPointHistory(ctx context.Context, mid int64, join_time int64, pn, ps int) (res *summer_camp.UserPointHistoryRes, err error) {
	// 获得的积分;来源：1 share*10+archive*30+view_total*20+award*300+150
	res = new(summer_camp.UserPointHistoryRes)
	list := make([]*summer_camp.PointInfo, 0)
	list = append(list, &summer_camp.PointInfo{
		Tim:        join_time,
		Type:       ObtainType,
		RewardName: ObtainNew,
		Point:      s.c.SummerCampConf.StartPlanPoint,
		Left:       0,
	})
	var lock sync.Mutex
	eg := errgroup.WithContext(ctx)
	// 观看获得的积分
	eg.Go(func(ctx context.Context) (err error) {
		_, viewHis, err := s.TaskHistoryDays(ctx, mid, s.c.SummerCampConf.ActivityIdRW,
			s.c.SCCounterCourseMap["view"], join_time, false)
		if err != nil {
			log.Errorc(ctx, "UserPointHistory TaskHistoryDays err,err is (%v),counter is (%s)",
				err, s.c.SummerCampConf.ActivityIdRW, s.c.SCCounterCourseMap["view"])
			return
		}
		lock.Lock()
		for _, v := range viewHis {
			list = append(list, &summer_camp.PointInfo{
				Tim:        v.Timestamp,
				Type:       ObtainType,
				RewardName: ObtainViewName,
				Point:      s.c.SummerCampConf.ViewPoint,
				Left:       0,
			})
		}
		lock.Unlock()

		return
	})
	// 分享获得的积分
	eg.Go(func(ctx context.Context) (err error) {
		_, viewHis, err := s.TaskHistoryDays(ctx, mid, s.c.SummerCampConf.ActivityIdRW,
			s.c.SCCounterCourseMap["share"], join_time, false)
		if err != nil {
			log.Errorc(ctx, "UserPointHistory TaskHistoryDays err,err is (%v),counter is (%s)",
				err, s.c.SummerCampConf.ActivityIdRW, s.c.SCCounterCourseMap["share"])
			return
		}
		lock.Lock()
		for _, v := range viewHis {
			list = append(list, &summer_camp.PointInfo{
				Tim:        v.Timestamp,
				Type:       ObtainType,
				RewardName: ObtainShareName,
				Point:      s.c.SummerCampConf.SharePoint,
				Left:       0,
			})
		}
		lock.Unlock()

		return
	})
	// 投稿获得的积分
	eg.Go(func(ctx context.Context) (err error) {
		_, viewHis, err := s.TaskHistoryDays(ctx, mid, s.c.SummerCampConf.ActivityIdRW,
			s.c.SCCounterCourseMap["archive"], join_time, false)
		if err != nil {
			log.Errorc(ctx, "UserPointHistory TaskHistoryDays err,err is (%v),counter is (%s)",
				err, s.c.SummerCampConf.ActivityIdRW, s.c.SCCounterCourseMap["archive"])
			return
		}
		lock.Lock()
		for _, v := range viewHis {
			list = append(list, &summer_camp.PointInfo{
				Tim:        v.Timestamp,
				Type:       ObtainType,
				RewardName: ObtainArchiveName,
				Point:      s.c.SummerCampConf.ArchivePoint,
				Left:       0,
			})
		}
		lock.Unlock()

		return
	})
	// 额外获得的积分
	eg.Go(func(ctx context.Context) (err error) {
		_, viewHis, err := s.TaskHistoryDays(ctx, mid, s.c.SummerCampConf.ActivityIdRW,
			s.c.SCCounterCourseMap["award"], join_time, false)
		if err != nil {
			log.Errorc(ctx, "UserPointHistory TaskHistoryDays err,err is (%v),counter is (%s)",
				err, s.c.SummerCampConf.ActivityIdRW, s.c.SCCounterCourseMap["award"])
			return
		}
		lock.Lock()
		for _, v := range viewHis {
			list = append(list, &summer_camp.PointInfo{
				Tim:        v.Timestamp,
				Type:       ObtainType,
				RewardName: ObtainAward,
				Point:      s.c.SummerCampConf.AwardPoint,
				Left:       0,
			})
		}
		lock.Unlock()

		return
	})
	// 消耗
	eg.Go(func(ctx context.Context) (err error) {
		// todo 读缓存
		_, costList, err := s.costPointDao.GetUserAllCost(ctx, s.c.SummerCampConf.ActivityIdRW, mid, false)
		if err != nil {
			log.Errorc(ctx, "UserPointHistory s.costPointDao.GetUserAllCost err,err is (%v)", err)
			return
		}
		if costList != nil && len(costList) > 0 {
			lock.Lock()
			for _, v := range costList {
				typ := 0
				rewarName := ""
				if v.CostType == 1 {
					typ = CostTypeLottery
					rewarName = CostLotteryName
				} else {
					typ = CostTypeExchange

					awardIdInt, err := strconv.ParseInt(v.AwardId, 10, 64)
					if err != nil {
						log.Errorc(ctx, "UserPointHistory strconv.ParseInt awardid err,"+
							"err is (%v),awardId is (%v).", err, v.AwardId)
						return err
					}
					info, err := client.ActivityClient.RewardsGetAwardConfigById(ctx, &actApi.RewardsGetAwardConfigByIdReq{
						Id: awardIdInt,
					})
					if err != nil {
						return err
					}
					rewarName = info.Name
				}
				list = append(list, &summer_camp.PointInfo{
					Tim:        int64(v.Ctime),
					Type:       typ,
					RewardName: rewarName,
					Point:      int64(v.CostValue),
					Left:       0,
				})
			}
			lock.Unlock()
		}

		return
	})
	if err = eg.Wait(); err != nil {
		return nil, ecode.PointHisListErr
	}

	// list排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Tim > list[j].Tim
	})
	// 填入剩余积分
	for i := len(list) - 1; i >= 0; i-- {
		if list[i].Type == ObtainType {
			if i == len(list)-1 {
				list[i].Left = list[i].Point
			} else {
				list[i].Left = list[i+1].Left + list[i].Point
			}

		} else {
			if i == len(list)-1 {
				list[i].Left = 0
			}
			list[i].Left = list[i+1].Left - list[i].Point
			if list[i].Left < 0 {
				// 做个非负兜底补救
				list[i].Left = 0
			}

		}

	}

	res.List = list
	res.Total = len(list)
	res.Size = ps // 留口子后续分页
	res.Page = pn
	return
}

// ExchangeAwardList 每日奖品列表
func (s *Service) ExchangeAwardList(ctx context.Context, mid int64, pn, ps int) (res *summer_camp.AwardListRes, err error) {
	res = new(summer_camp.AwardListRes)
	// 查找列表
	awardList, errG := s.rewardConfDao.GetTodayAwardList(ctx, s.c.SummerCampConf.ActivityIdRW, 2)
	if errG != nil {
		log.Errorc(ctx, "ExchangeAwardList err ,err is (%v).", errG)
		return nil, ecode.RewardListErr
	}
	if awardList == nil {
		return
	}

	var (
		awardStockLeft map[int64]int = make(map[int64]int)
		//awardIds       []int64                            = make([]int64, 0)
		stockIds   []int64                            = make([]int64, 0)
		awardsInfo map[int64]*actApi.RewardsAwardInfo = make(map[int64]*actApi.RewardsAwardInfo)
		exMap      map[string]int                     = make(map[string]int)
		l1                                            = sync.Mutex{}
		l2                                            = sync.Mutex{}
	)
	eg := errgroup.WithContext(ctx)
	// 查询奖品库存&奖品详情
	eg.Go(func(ctx context.Context) (err error) {
		for _, award := range awardList {
			awardIdInt, err := strconv.ParseInt(award.AwardId, 10, 64)
			if err != nil {
				log.Errorc(ctx, "ExchangeAwardList strconv.ParseInt awardid err,"+
					"err is (%v),awardId is (%v).", err, award.AwardId)
				return err
			}
			// 查询奖品详情
			info, err := client.ActivityClient.RewardsGetAwardConfigById(ctx, &actApi.RewardsGetAwardConfigByIdReq{
				Id: awardIdInt,
			})
			if err != nil {
				return err
			}
			l1.Lock()
			awardsInfo[awardIdInt] = info
			l1.Unlock()
			if err != nil {
				log.Errorc(ctx, "ExchangeAwardList client.ActivityClient.RewardsGetAwardConfigById err,"+
					"err is (%v),awardId is (%v).", err, award.AwardId)
				return err
			}
			stockIds = append(stockIds, award.StockId)

		}
		//awardStockLeft, err = s.stockDao.GetGiftStocks(ctx, s.c.SummerCampConf.ActivityIdRW, awardIds)
		awardStockLeftRes, err := client.ActivityClient.GetStocksByIds(ctx, &actApi.GetStocksReq{
			StockIds:  stockIds,
			SkipCache: false,
			Mid:       mid,
		})
		if err != nil || awardStockLeftRes == nil || awardStockLeftRes.StockMap == nil {
			log.Errorc(ctx, "ExchangeAwardList client.ActivityClient.GetStocksByIds err,err is (%v).", err)
			err = nil
		}
		for k, v := range awardStockLeftRes.StockMap {
			awardStockLeft[k] = int(v.List[0].StockNum)
		}
		return
	})
	// 查询用户今日兑换
	eg.Go(func(ctx context.Context) (err error) {
		exList, err := s.costPointDao.TodayUserHasExchangedPrizes(ctx, mid, s.c.SummerCampConf.ActivityIdRW)
		if err != nil {
			log.Errorc(ctx, "ExchangeAwardList getcostPointDao.TodayUserHasExchangedPrizes err ,err is (%v).", err)
			return
		}
		// 奖品是否兑换 Map
		if exList != nil {
			for _, one := range exList {
				if one.AwardId != "" {
					l2.Lock()
					exMap[one.AwardId] = 1
					l2.Unlock()
				}
			}
		}
		return
	})

	if err := eg.Wait(); err != nil {
		return nil, ecode.RewardListErr
	}

	// 构造返回
	for _, info := range awardList {
		canExchange := true
		if _, ok := exMap[info.AwardId]; ok {
			canExchange = false
		}
		awardIdInt, _ := strconv.ParseInt(info.AwardId, 10, 64)
		stockLeft := 0
		awardName := "null"
		awardIcon := ""
		if _, ok := awardStockLeft[info.StockId]; ok {
			stockLeft = awardStockLeft[info.StockId]
		}
		// 库存为0不可兑换
		if stockLeft == 0 {
			canExchange = false
		}
		if _, ok := awardsInfo[awardIdInt]; ok {
			awardName = awardsInfo[awardIdInt].Name
			awardIcon = awardsInfo[awardIdInt].IconUrl
		}
		res.List = append(res.List, &summer_camp.AwardInfo{
			AwardId:         info.AwardId,
			AwardName:       awardName,
			AwardCost:       int64(info.CostValue),
			StockLeft:       int64(stockLeft),
			Icon:            awardIcon,
			UserCanExchange: canExchange,
		})
	}
	res.Size = len(awardList)
	res.Page = pn
	res.Total = len(awardList)
	return

}
