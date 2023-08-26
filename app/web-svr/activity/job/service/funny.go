package service

import (
	"context"
	"encoding/json"
	"fmt"
	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"go-common/library/log"
	"go-common/library/net/trace"
	go_common_library_time "go-common/library/time"
	grpcArc "go-gateway/app/app-svr/archive/service/api"
	likemdl "go-gateway/app/web-svr/activity/job/model/like"
	mdlmail "go-gateway/app/web-svr/activity/job/model/mail"
	"go-gateway/pkg/idsafe/bvid"
	"strconv"
	"strings"
	"sync"
	"time"
)

type part1MailFileData struct {
	mid      int64
	uname    string
	bvID     string
	title    string
	view     int32
	isNew    string
	pubTime  go_common_library_time.Time
	like     int32
	duration int64
	state    int32
	typeID   int32
}

type part2MailFileData struct {
	mid        int64
	production int32
	view       int32
	duration   int64
}

type relation struct {
	num      int32
	view     int32
	duration int64
	like     int32
	pubTime  go_common_library_time.Time
	uname    string
	bvID     string
	title    string
}

var (
	funnyPartOne context.Context
	funnyPartTwo context.Context
)

// 同步运营同学视频数据到数据中心
//func (s *Service) FunnySyncVideoData() {
//	c := context.Background()
//	log.Infoc(c, "Start FunnySyncVideoData Script : time %s", time.Now().Format("2006-01-02 15:04"))
//	s.funnySyncRunning.Lock()
//	defer s.funnySyncRunning.Unlock()
//	aids, err := s.funnyGetAids(c)
//	if err != nil {
//		log.Errorc(c, "FunnySyncVideoData funnyGetAids Err(%v)", err)
//		return
//	}
//	err = s.syncFunnyAidsToActPlat(c, aids)
//	if err != nil {
//		log.Errorc(c, "FunnySyncVideoData syncFunnyAidsToActPlat Err aids:%v err:%v", aids, err)
//		return
//	}
//	log.Infoc(c, "FunnySyncVideoData Sync Success aids: %v", aids)
//	return
//}

// 计算搞笑新人福利人数
func (s *Service) CaculatePartOne() {
	var (
		limit = int64(50)
		page  = int64(0)
		cData sync.Map
		//num             int
		mailFileDataSet []part1MailFileData
	)
	type arcInfo struct {
		state    int32
		like     int32
		pubTime  go_common_library_time.Time
		duration int64
		uname    string
		bvID     string
		title    string
		view     int32
		typeID   int32
	}

	if time.Now().Unix() > 1605180540 {
		return
	}

	partOneCtx := trace.SimpleServerTrace(context.Background(), "funnyPartOne")
	if err := s.SendTextMail(partOneCtx, s.buildReceivers(), "搞笑迎新大会", "搞笑迎新大会第一部分人数计算脚本开始执行"); err != nil {
		log.Warnc(partOneCtx, "Start CaculatePartOne Script Email Err: time:%s err:%v", time.Now().Format("2006-01-02 15:04"), err)
	}
	log.Infoc(partOneCtx, "Start CaculatePartOne Script : time %s", time.Now().Format("2006-01-02 15:04"))
	for {
		// 批量获取稿件ID
		aids, err := s.funny.GetUserBatchData(partOneCtx, s.c.Funny.ActSid, limit, page)
		if err != nil {
			// retry
			aids, err = s.funny.GetUserBatchData(partOneCtx, s.c.Funny.ActSid, limit, page)
			if err != nil {
				log.Errorc(partOneCtx, "%v", err)
				return
			}
		}
		page++
		// GRPC请求
		arcsResply, err := s.arcClient.Arcs(partOneCtx, &grpcArc.ArcsRequest{Aids: aids})
		if err != nil {
			// retry
			arcsResply, err = s.arcClient.Arcs(partOneCtx, &grpcArc.ArcsRequest{Aids: aids})
			if err != nil {
				log.Errorc(partOneCtx, "CaculatePartOne Get Arcs By GRPC Err aids:%v", aids)
				return
			}
		}

		for _, aid := range aids {
			arc, ok := arcsResply.Arcs[aid]

			if !ok {
				log.Errorc(partOneCtx, "CaculatePartOne Get Arcs By GRPC Range Data Err no aid:%v", aid)
				return
			}

			if arc.TypeID != 138 {
				continue
			}

			mid := arc.Author.Mid // 用户id
			isNew, err := s.funny.IsNewUser(partOneCtx, mid)
			if err != nil {
				log.Errorc(partOneCtx, "%v", err)
				return
			}

			if isNew == false {
				log.Infoc(partOneCtx, "mid:%v is not activity new user", mid)
				continue
			}

			bvID, err := bvid.AvToBv(arc.Aid)
			if err != nil {
				continue
			}

			// 内存获取数据
			arcData, ok := cData.Load(mid)
			if !ok { // 数据不存在
				cData.Store(mid, arcInfo{
					state:    arc.State,
					like:     arc.Stat.Like,
					pubTime:  arc.PubDate,
					duration: arc.Duration,
					uname:    arc.Author.Name,
					bvID:     bvID,
					title:    arc.Title,
					view:     arc.Stat.View,
					typeID:   arc.TypeID,
				})
				continue
			}
			// 存在的话 拿出来比较pubTime map中数据日期靠后 则被替换 否则不动
			if arcData.(arcInfo).pubTime > arc.PubDate {
				// 用新的替换旧的
				cData.Store(mid, arcInfo{
					state:    arc.State,
					like:     arc.Stat.Like,
					pubTime:  arc.PubDate,
					duration: arc.Duration,
					uname:    arc.Author.Name,
					bvID:     bvID,
					title:    arc.Title,
					view:     arc.Stat.View,
					typeID:   arc.TypeID,
				})
			}
		}

		// 数据计算完毕
		if len(aids) < 50 {
			break
		}
	}

	// 批量剔除不符合规则的数据
	cData.Range(func(key, value interface{}) bool {
		item := value.(arcInfo)
		if item.typeID != 138 || item.like < s.c.Funny.PartOneLikesLimit || item.duration <= s.c.Funny.FunnyVideoDuration {
			log.Infoc(partOneCtx, "delete item state:%v like:%v mid:%v duration：%v", item.state, item.like, key.(int64), s.c.Funny.FunnyVideoDuration)
			cData.Delete(key)
		} else {
			mailFileDataSet = append(mailFileDataSet, part1MailFileData{
				mid:     key.(int64),
				isNew:   "新用户",
				pubTime: item.pubTime,
				like:    item.like,
				uname:   item.uname,
				bvID:    item.bvID,
				title:   item.title,
				view:    item.view,
				state:   item.state,
				typeID:  item.typeID,
			})
		}
		return true
	})

	// 统计总数
	//num = len(mailFileDataSet)
	//
	//// 统计数字 写入缓存
	//err := s.funny.SetTask1Data(partOneCtx, num)
	//if err != nil {
	//	log.Errorc(partOneCtx, "%v", err)
	//	return
	//}

	if err := s.sendEmailWithCSVFilePart1(partOneCtx, mailFileDataSet); err != nil {
		log.Warnc(partOneCtx, "Start CaculatePartOne Script Email Err: time:%s err:%v", time.Now().Format("2006-01-02 15:04"), err)
	}

}

// 计算第二部分的数据
func (s *Service) CaculatePartTwo() {
	var (
		limit = int64(50)
		page  = int64(0)
		cData sync.Map
		//num             int
		mailFileDataSet []part2MailFileData
		relationData    sync.Map
	)
	type arcInfo struct {
		num      int32
		view     int32
		duration int64
	}

	if time.Now().Unix() > 1605022526 {
		return
	}

	partTwoCtx := trace.SimpleServerTrace(context.Background(), "funnyPartTwo")
	if err := s.SendTextMail(partTwoCtx, s.buildReceivers(), "搞笑迎新大会", "搞笑迎新大会第二部分计算脚本开始执行"); err != nil {
		log.Warnc(partTwoCtx, "Start CaculatePartTwo Script Email Err: time:%s err:%v", time.Now().Format("2006-01-02 15:04"), err)
	}
	log.Infoc(partTwoCtx, "Start CaculatePartTwo Script : time %s", time.Now().Format("2006-01-02 15:04"))
	for {
		// 批量获取稿件ID
		aids, err := s.funny.GetUserBatchData(partTwoCtx, s.c.Funny.ActSid, limit, page)
		if err != nil {
			// retry
			aids, err = s.funny.GetUserBatchData(partTwoCtx, s.c.Funny.ActSid, limit, page)
			if err != nil {
				log.Errorc(partTwoCtx, "%v", err)
				return
			}
		}
		page++
		// GRPC请求
		arcsResply, err := s.arcClient.Arcs(partTwoCtx, &grpcArc.ArcsRequest{Aids: aids})
		if err != nil {
			// retry
			arcsResply, err = s.arcClient.Arcs(partTwoCtx, &grpcArc.ArcsRequest{Aids: aids})
			if err != nil {
				log.Errorc(partTwoCtx, "CaculatePartTwo Get Arcs By GRPC Err aids:%v", aids)
				return
			}
		}

		for _, aid := range aids {
			arc, ok := arcsResply.Arcs[aid]
			if !ok {
				log.Errorc(partTwoCtx, "CaculatePartTwo Get Arcs By GRPC Range Data Err no aid:%v", aid)
				return
			}

			mid := arc.Author.Mid // 用户id
			// 查看这条数据是否在线
			if arc.IsNormal() == false {
				log.Infoc(partTwoCtx, "mid:%v aid:%v is not normal", mid, aid)
				continue
			}

			// 时长 <= 30s 剔除
			if arc.Duration <= s.c.Funny.FunnyVideoDuration {
				log.Infoc(partTwoCtx, "mid:%v aid:%v duration less 30s", mid, aid)
				continue
			}

			// 内存获取数据
			arcOldData, ok := cData.Load(mid)
			if !ok { // 数据不存在
				cData.Store(mid, arcInfo{
					num:  1,
					view: arc.Stat.View,
				})
			} else {
				// 存在的话 做累计
				oldData := arcOldData.(arcInfo)
				cData.Store(mid, arcInfo{
					num:  oldData.num + 1,
					view: oldData.view + arc.Stat.View,
				})
			}

			// 往关联数据里面append数据
			bvID, err := bvid.AvToBv(arc.Aid)
			if err != nil {
				log.Errorc(partTwoCtx, "bvid.AvToBv failed err:%v aid:%v", err, arc.Aid)
				continue
			}

			r, ok := relationData.Load(mid)
			if !ok { // 未找到 直接append数据
				var rd []relation
				rd = append(rd, relation{
					like:     arc.Stat.Like,
					pubTime:  arc.PubDate,
					duration: arc.Duration,
					uname:    arc.Author.Name,
					bvID:     bvID,
					title:    arc.Title,
					view:     arc.Stat.View,
				})
				relationData.Store(mid, rd)
			} else {
				// 找到断言 append数据
				rd, ok := r.([]relation)
				if !ok {
					log.Errorc(partTwoCtx, "assert error info:%v", r)
					continue
				}
				rd = append(rd, relation{
					like:     arc.Stat.Like,
					pubTime:  arc.PubDate,
					duration: arc.Duration,
					uname:    arc.Author.Name,
					bvID:     bvID,
					title:    arc.Title,
					view:     arc.Stat.View,
				})
				relationData.Store(mid, rd)
			}
		}

		// 数据计算完毕
		if len(aids) < 50 {
			break
		}
	}

	// 批量剔除不符合规则的数据
	cData.Range(func(key, value interface{}) bool {
		item := value.(arcInfo)
		if item.num < s.c.Funny.PartTwoVideoNumLimit || item.view < s.c.Funny.PartTwoVideoViewLimit {
			log.Infoc(partTwoCtx, "delete item num:%v view:%v mid:%v", item.num, item.view, key.(int64))
			cData.Delete(key)
			relationData.Delete(key)
		} else {
			mailFileDataSet = append(mailFileDataSet, part2MailFileData{
				mid:        key.(int64),
				production: item.num,
				view:       item.view,
			})
		}
		return true
	})

	// 统计总数
	//num = len(mailFileDataSet)
	//
	//// 统计数字 写入缓存
	//err := s.funny.SetTask2Data(partTwoCtx, num)
	//if err != nil {
	//	log.Errorc(partTwoCtx, "%v", err)
	//	return
	//}

	if err := s.sendEmailWithCSVFilePart2(partTwoCtx, mailFileDataSet, relationData); err != nil {
		log.Warnc(partTwoCtx, "Start CaculatePartTwo Script Email Err: time:%s err:%v", time.Now().Format("2006-01-02 15:04"), err)
	}
}

func (s *Service) funnyGetAids(c context.Context) ([]int64, error) {
	res, err := s.dao.SourceItem(context.Background(), s.c.Funny.Vid)
	if err != nil {
		log.Errorc(c, "FunnySyncVideoData funnyGetAids SourceItem Err vid:%v res:%v err:%v", s.c.Funny.Vid, res, err)
		return nil, err
	}
	tmp := new(likemdl.ArcListData)
	if err = json.Unmarshal(res, &tmp); err != nil {
		log.Errorc(c, "FunnySyncVideoData funnyGetAids JSON Unmatshal Err res:%v err:%v", res, err)
		return nil, err
	}
	log.Infoc(c, "FunnySyncVideoData funnyGetAids SourceItem Data : %v", tmp)
	aids := []int64{}
	if tmp != nil && tmp.List != nil {
		for _, v := range tmp.List {
			for _, val := range strings.Split(v.Data.Aids, ",") {
				if strings.HasPrefix(val, "BV") {
					avid, err := bvid.BvToAv(val)
					if err != nil {
						log.Errorc(c, "FunnySyncVideoData funnyGetAids switch bv to av: %s %+v", val, err)
						continue
					}
					aids = append(aids, avid)
					continue
				}
				if avid, _ := strconv.ParseInt(val, 10, 64); avid > 0 {
					aids = append(aids, avid)
				}
			}
		}
	}
	log.Infoc(c, "FunnySyncVideoData funnyGetAids aids : %v", aids)
	return aids, nil
}
func (s *Service) syncFunnyAidsToActPlat(c context.Context, aids []int64) error {
	values := []*actplatapi.FilterMemberInt{}
	expireTime := int64(600)
	for _, i := range aids {
		values = append(values, &actplatapi.FilterMemberInt{Value: i, ExpireTime: expireTime})
	}
	_, err := s.actplatClient.AddFilterMemberInt(c, &actplatapi.SetFilterMemberIntReq{
		Activity: s.c.Funny.ActPlatActivity,
		Counter:  s.c.Funny.ActPlatCounter,
		Filter:   "filter_aid_sources",
		Values:   values,
	})
	return err
}

func (s *Service) sendEmailWithCSVFilePart1(c context.Context, dataSet []part1MailFileData) error {
	categoryHeader := []string{"uid", "是否是新用户", "稿件BV号", "稿件标题", "播放量", "点赞数", "首投作品过审时间", "当前时间作品状态", "分区ID"}
	data := [][]string{}
	for _, v := range dataSet {
		rows := []string{}
		rows = append(rows, strconv.FormatInt(v.mid, 10), "是", v.bvID, v.title, fmt.Sprint(v.view), fmt.Sprint(v.like), v.pubTime.Time().String(), fmt.Sprint(v.state), fmt.Sprint(v.typeID))
		data = append(data, rows)
	}
	fileName := fmt.Sprintf("%v_%v.csv", "搞笑视频大会任务一符合条件用户明细", time.Now().Format("20060102"))
	err := s.funnyCreateCsvAndSend(c, s.c.Dubbing.FilePath, fileName, fileName, categoryHeader, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) sendEmailWithCSVFilePart2(c context.Context, dataSet []part2MailFileData, relationData sync.Map) error {
	categoryHeader := []string{"uid", "是否是新用户", "稿件BV号", "稿件标题", "播放量", "点赞数", "作品总数", "首投作品过审时间"}
	data := [][]string{}
	for _, v := range dataSet {
		// 拿到用户mid
		mID := v.mid
		// 关联数据中取出来
		syncMapRelationData, _ := relationData.Load(mID)
		relationData := syncMapRelationData.([]relation)
		for _, item := range relationData {
			rows := []string{}
			rows = append(rows, strconv.FormatInt(mID, 10), "新用户", item.bvID, item.title, fmt.Sprint(item.view), fmt.Sprint(item.like), fmt.Sprint(v.production), item.pubTime.Time().String())
			data = append(data, rows)
		}
	}
	fileName := fmt.Sprintf("%v_%v.csv", "搞笑视频大会任务二符合条件用户明细", time.Now().Format("20060102"))
	err := s.funnyCreateCsvAndSend(c, s.c.Dubbing.FilePath, fileName, fileName, categoryHeader, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) funnyCreateCsvAndSend(c context.Context, filePath, fileName string, subject string, categoryHeader []string, data [][]string) error {
	base := &mdlmail.Base{
		Host:    s.c.Mail.Host,
		Port:    s.c.Mail.Port,
		Address: s.c.Mail.Address,
		Pwd:     s.c.Mail.Pwd,
		Name:    s.c.Mail.Name,
	}
	return s.activityCreateCsvAndSend(c, "./data/", fileName, subject, base, s.buildReceivers(), []*mdlmail.Address{}, []*mdlmail.Address{}, categoryHeader, data)
}

func (s *Service) buildReceivers() []*mdlmail.Address {
	var mailReceivers []*mdlmail.Address

	receivers := strings.Split(s.c.Funny.EmailReceivers, ",")
	for _, v := range receivers {
		user := &mdlmail.Address{
			Address: v,
			Name:    "",
		}
		mailReceivers = append(mailReceivers, user)
	}
	return mailReceivers
}
