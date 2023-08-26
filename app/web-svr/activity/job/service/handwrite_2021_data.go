package service

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	handwritemdl "go-gateway/app/web-svr/activity/job/model/handwrite"
	sourcemdl "go-gateway/app/web-svr/activity/job/model/source"
	"go-gateway/pkg/idsafe/bvid"
	"strconv"
	"strings"
	"time"
)

const (
	allHandwriteLimit = 1000
	// handWorkMidResultFileName 手书活动-用户任务完成结果
	handWorkMidResultFileName = "手书活动-用户任务完成结果"
	// handWorkArchiveFileName 手书活动-稿件情况
	handWorkArchiveFileName = "手书活动-稿件情况"
)

var handwrite2021DataCtx context.Context

func handwrite2021DatasCtxInit() {
	handwrite2021DataCtx = trace.SimpleServerTrace(context.Background(), "handwrite2021 data")
}

// Handwrite2021Data 手书2021
func (s *Service) Handwrite2021Data() {
	s.handwrite2021DataRunning.Lock()
	defer s.handwrite2021DataRunning.Unlock()

	handwrite2021DatasCtxInit()
	ctx := handwrite2021DataCtx
	err := s.handwriteData(ctx)
	if err != nil {
		// 错误处理
		err = s.sendWechat(ctx, "[手书任务 数据产出]", fmt.Sprintf("%v", err), "zhangtinghua")
		if err != nil {
			log.Errorc(ctx, "handwrite2021 s.sendWechat (%v)", err)
		}
	}
}

func (s *Service) handwriteData(c context.Context) error {
	midInfo, archiveInfo, err := s.doHandwrite2021Data(c)
	if err != nil {
		return err
	}
	err = s.handwriteMemberInfoToCsv(c, midInfo)
	if err != nil {
		return err
	}
	err = s.handwriteArcToCsv(c, archiveInfo)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) getTime(timeInt int64) (dateStr string) {
	if timeInt > 0 {
		//返回time对象
		t := time.Unix(timeInt, 0)
		//返回string
		dateStr = t.Format("2006-01-02 15:04:05")
	}
	return
}

func (s *Service) handwriteMemberInfoToCsv(c context.Context, midInfo map[int64]*handwritemdl.MidTaskAllData) error {
	categoryHeader := []string{"用户ID", "昵称", "粉丝数", "神仙模式（个数）", "神仙模式相关稿件aid", "神仙模式完成时间", "佛系模式", "佛系模式相关稿件aid", "佛系模式完成时间", "爆肝模式（中）", "爆肝模式（中）相关稿件aid", "爆肝模式（中）完成时间", "爆肝模式（高）", "爆肝模式（高）相关稿件aid", "爆肝模式（高）完成时间", "金额（分）"}
	data := [][]string{}
	if midInfo == nil {
		return nil
	}
	for _, v := range midInfo {
		rows := []string{}
		var midStr, godstr, godTime, fans, tired1str, tired1Time, tired2str, tired2Time, tired3str, tired3Time, money string
		midStr = strconv.FormatInt(v.Mid, 10)
		fans = strconv.FormatInt(v.Fans, 10)
		godstr = strconv.Itoa(v.God)

		godTime = s.getTime(v.GodTime)
		tired1Time = s.getTime(v.TiredLevel1Time)
		tired2Time = s.getTime(v.TiredLevel2Time)
		tired3Time = s.getTime(v.TiredLevel3Time)

		tired1str = strconv.Itoa(v.TiredLevel1)
		tired2str = strconv.Itoa(v.TiredLevel2)
		tired3str = strconv.Itoa(v.TiredLevel3)
		money = strconv.FormatInt(v.Money, 10)
		rows = append(rows, midStr, v.NickName, fans, godstr, v.GodDetail, godTime, tired1str, v.TiredLevel1Detail, tired1Time, tired2str, v.TiredLevel2Detail, tired2Time, tired3str, v.TiredLevel3Detail, tired3Time, money)
		data = append(data, rows)
	}
	fileName := fmt.Sprintf("%v_%v.csv", handWorkMidResultFileName, time.Now().Format("200601021504"))
	err := s.createCsvAndSend(c, s.c.Handwrite2021.FilePath, fileName, handWorkMidResultFileName, categoryHeader, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) handwriteArcToCsv(c context.Context, archiveInfo map[int64]*sourcemdl.Archive) (err error) {
	categoryHeader := []string{"稿件AID", "BVID", "稿件标题", "投稿时间", "播放", "点赞", "硬币", "收藏"}
	data := [][]string{}
	if archiveInfo == nil {
		return nil
	}
	for _, v := range archiveInfo {
		rows := []string{}
		var aidStr, bvidstr, pubtime, view, like, coin, fav string
		aidStr = strconv.FormatInt(v.Aid, 10)
		bvidstr, err := bvid.AvToBv(v.Aid)
		if err != nil {
			log.Errorc(c, "Failed to switch bv to av: %s %+v", v.Aid, err)
			continue
		}
		view = strconv.FormatInt(v.View, 10)
		like = strconv.FormatInt(v.Like, 10)
		coin = strconv.FormatInt(v.Coin, 10)
		fav = strconv.FormatInt(v.Fav, 10)
		pubtime = s.getTime(v.PubTime)

		rows = append(rows, aidStr, bvidstr, v.Title, pubtime, view, like, coin, fav)
		data = append(data, rows)
	}
	fileName := fmt.Sprintf("%v_%v.csv", handWorkArchiveFileName, time.Now().Format("200601021504"))
	err = s.createCsvAndSend(c, s.c.Handwrite2021.FilePath, fileName, handWorkArchiveFileName, categoryHeader, data)
	if err != nil {
		return err
	}
	return nil
}

// doHandwrite2021Data 数据产出
func (s *Service) doHandwrite2021Data(ctx context.Context) (map[int64]*handwritemdl.MidTaskAllData, map[int64]*sourcemdl.Archive, error) {
	// task 结果产出
	midAllTask, err := s.getAllMidTaskResult(ctx)
	if err != nil {
		log.Errorc(ctx, "s.getAllMidTaskResult err(%v)", err)
		return nil, nil, err
	}
	midsMap := make(map[int64]struct{})
	mids := make([]int64, 0)
	aids := make([]int64, 0)
	aidsMap := make(map[int64]struct{})
	midTaskDetail := make(map[int64]*handwritemdl.MidTaskAllData)
	var (
		memberFansMap  map[int64]int64
		memberNickanme map[int64]string
		arcInfo        map[int64]*api.Arc
		award          *handwritemdl.AwardCountNew
	)
	for _, v := range midAllTask {
		var taskDetail string
		var taskDetailStruct = make([]string, 0)
		if len(v.TaskDetailStruct) > 0 {
			for _, item := range v.TaskDetailStruct {
				taskDetailStruct = append(taskDetailStruct, fmt.Sprintf("%d", item))
			}
		}
		if len(taskDetailStruct) > 0 {
			taskDetail = strings.Join(taskDetailStruct, ";")

		}

		if _, ok := midTaskDetail[v.Mid]; !ok {
			midTaskDetail[v.Mid] = &handwritemdl.MidTaskAllData{}
			midTaskDetail[v.Mid].Mid = v.Mid
		}
		if v.TaskType == handwritemdl.TaskTypeGod {
			midTaskDetail[v.Mid].God = v.FinishCount
			midTaskDetail[v.Mid].GodTime = v.FinishTime
			midTaskDetail[v.Mid].GodDetail = taskDetail
		}
		if v.TaskType == handwritemdl.TaskTypeTiredLevel1 {
			midTaskDetail[v.Mid].TiredLevel1 = v.FinishCount
			midTaskDetail[v.Mid].TiredLevel1Time = v.FinishTime
			midTaskDetail[v.Mid].TiredLevel1Detail = taskDetail
		}
		if v.TaskType == handwritemdl.TaskTypeTiredLevel2 {
			midTaskDetail[v.Mid].TiredLevel2 = v.FinishCount
			midTaskDetail[v.Mid].TiredLevel2Time = v.FinishTime
			midTaskDetail[v.Mid].TiredLevel2Detail = taskDetail
		}
		if v.TaskType == handwritemdl.TaskTypeTiredLevel3 {
			midTaskDetail[v.Mid].TiredLevel3 = v.FinishCount
			midTaskDetail[v.Mid].TiredLevel3Time = v.FinishTime
			midTaskDetail[v.Mid].TiredLevel3Detail = taskDetail
		}
		if _, ok := midsMap[v.Mid]; !ok {
			midsMap[v.Mid] = struct{}{}
			mids = append(mids, v.Mid)
		}
		if v.TaskDetailStruct != nil && len(v.TaskDetailStruct) > 0 {
			for _, aid := range v.TaskDetailStruct {
				if _, ok := aidsMap[aid]; !ok {
					aidsMap[aid] = struct{}{}
					aids = append(aids, aid)
				}
			}
		}
	}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if memberFansMap, err = s.memberFollowerNum(ctx, mids); err != nil {
			log.Errorc(ctx, "s.memberFollowerNum() error(%v)", err)
			return ecode.ActivityWriteHandFansErr
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if memberNickanme, err = s.memberNickname(ctx, mids); err != nil {
			log.Errorc(ctx, "s.memberNickname() error(%v)", err)
			return ecode.ActivityWriteHandMemberErr
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if arcInfo, err = s.sourceSvr.ArchiveInfo(ctx, aids); err != nil {
			log.Errorc(ctx, "s.sourceSvr.ArchiveInfo(ctx, aids)", err)
			return err
		}
		return nil
	})
	// 总数统计
	eg.Go(func(ctx context.Context) (err error) {

		award, err = s.handWrite.GetTaskCount(ctx)
		if err != nil {
			log.Errorc(ctx, "s.GetTaskCount err(%v)", err)
		}
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, nil, err
	}

	for mid := range midTaskDetail {
		var money int64
		if award != nil {
			if award.God != 0 {
				money += (s.c.Handwrite2021.GodAllMoney / award.God) * int64(midTaskDetail[mid].God)
			}
			if award.TiredLevel1 != 0 {
				money += (s.c.Handwrite2021.Tired1Money / award.TiredLevel1) * int64(midTaskDetail[mid].TiredLevel1)
			}
			if award.TiredLevel2 != 0 {
				money += (s.c.Handwrite2021.Tired2Money / award.TiredLevel2) * int64(midTaskDetail[mid].TiredLevel2)
			}
			if award.TiredLevel3 != 0 {
				money += (s.c.Handwrite2021.Tired3Money / award.TiredLevel3) * int64(midTaskDetail[mid].TiredLevel3)
			}
		}
		if fans, ok := memberFansMap[mid]; ok {
			midTaskDetail[mid].Fans = fans
		}
		if nickName, ok := memberNickanme[mid]; ok {
			midTaskDetail[mid].NickName = nickName
		}
		midTaskDetail[mid].Money = money
	}
	var archiveList = make(map[int64]*sourcemdl.Archive)
	if arcInfo != nil {
		for _, v := range arcInfo {
			archiveList[v.Aid] = &sourcemdl.Archive{
				Aid:     v.Aid,
				Mid:     v.Author.Mid,
				View:    int64(v.Stat.View),
				Danmaku: int64(v.Stat.Danmaku),
				Reply:   int64(v.Stat.Reply),
				Fav:     int64(v.Stat.Fav),
				Coin:    int64(v.Stat.Coin),
				Share:   int64(v.Stat.Share),
				Like:    int64(v.Stat.Like),
				Videos:  int64(v.Videos),
				PubTime: int64(v.PubDate),
				Title:   v.Title,
			}
		}
	}
	return midTaskDetail, archiveList, nil

}

// getAllCollege 获取所有学校
func (s *Service) getAllMidTaskResult(c context.Context) (handwriteList []*handwritemdl.MidTaskDB, err error) {
	var offset int64
	handwriteList = make([]*handwritemdl.MidTaskDB, 0)
	for {
		handwrite, err := s.handWrite.GetAllMidTask(c, offset, allHandwriteLimit)
		if err != nil {
			log.Errorc(c, "s.college.GetAllCollege error(%v)", err)
			return nil, err
		}
		if len(handwrite) > 0 {
			handwriteList = append(handwriteList, handwrite...)
		}
		if len(handwrite) < allHandwriteLimit {
			break
		}
		offset += allHandwriteLimit
	}

	return handwriteList, nil
}
