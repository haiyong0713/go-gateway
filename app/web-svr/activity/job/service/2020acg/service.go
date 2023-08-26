package acg

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/http/blademaster"
	"go-common/library/sync/errgroup.v2"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/dao/like"
	l "go-gateway/app/web-svr/activity/job/model/like"
	xmail "go-gateway/app/web-svr/activity/job/model/mail"
	"go-gateway/app/web-svr/activity/job/service/mail"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	batchDBQuery  = 1000
	batchArcQuery = 100
)

// Service service
type Service struct {
	c         *conf.Config
	dao       *like.Dao
	arcClient arcapi.ArchiveClient
	redis     *redis.Pool
	mail      *mail.Service
	lock      *sync.Mutex
	client    *blademaster.Client
}

type UserTaskState struct {
	Mid        int64
	Task       []*TaskState
	FinishTask int
	Money      int
}

type TaskState struct {
	Finish bool
	Count  int
	Score  int64
}

type ArcDetail struct {
	*l.Item
	*arcapi.Arc
}

func New(c *conf.Config, dao *like.Dao, arcClient arcapi.ArchiveClient) (s *Service) {
	s = &Service{
		c:         c,
		dao:       dao,
		arcClient: arcClient,
		redis:     redis.NewPool(c.Redis.Config),
		mail:      mail.New(c),
		lock:      &sync.Mutex{},
		client:    blademaster.NewClient(c.HTTPClient),
	}
	go s.UpdateTaskStateProc()
	return s
}

func (s *Service) doActionWithAlert(ctx context.Context, action func() error) {
	err := action()
	if err != nil {
		s.SendWeChat(ctx, "UpdateTaskStateProc 执行异常", fmt.Sprintf("执行时间:%s\n异常信息：%v", time.Now(), err), "ouyangkeshou")
	} else {
		//s.SendWeChat(ctx, "UpdateTaskStateProc 执行完成通知", fmt.Sprintf("执行时间:%s", time.Now()), "ouyangkeshou")
	}
}

func (s *Service) UpdateTaskStateProc() {
	ctx := context.Background()
	s.doActionWithAlert(ctx, func() error {
		return s.UpdateTaskState(ctx, time.Now())
	})
	for range time.Tick(time.Duration(s.c.Acg2020.UpdateDuration)) {
		s.doActionWithAlert(ctx, func() error {
			return s.UpdateTaskState(ctx, time.Now())
		})
	}
}

func (s *Service) SendWeChat(c context.Context, title, msg, user string) (err error) {
	var msgBytes []byte
	params := map[string]interface{}{
		"Action":    "NotifyCreate",
		"SendType":  "wechat_message",
		"PublicKey": s.c.Acg2020.WxKey,
		"UserName":  user,
		"Content": map[string]string{
			"subject": title,
			"body":    title + "\n" + msg,
		},
		"TreeId":    "",
		"Signature": "1",
		"Severity":  "P5",
	}
	if msgBytes, err = json.Marshal(params); err != nil {
		return
	}
	var req *http.Request
	if req, err = http.NewRequest(http.MethodPost, s.c.Host.MerakCo, strings.NewReader(string(msgBytes))); err != nil {
		return
	}
	req.Header.Add("content-type", "application/json; charset=UTF-8")
	res := &struct {
		RetCode int `json:"RetCode"`
	}{}
	if err = s.client.Do(c, req, &res); err != nil {
		log.Error("SendWechat d.client.Do(title:%s,msg:%s,user:%s) error(%v)", title, msg, user, err)
		return
	}
	if res.RetCode != 0 {
		err = ecode.Int(res.RetCode)
		log.Error("SendWechat d.client.Do(title:%s,msg:%s,user:%s) error(%v)", title, msg, user, err)
		return
	}
	return
}

func (s *Service) redisKeyUserTaskState(mid int64) string {
	return fmt.Sprintf("acg2020_mid_%d", mid)
}

func (s *Service) UpdateTaskState(ctx context.Context, now time.Time) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	taskNum := len(s.c.Acg2020.Task)
	// 计算当前支持的task
	runningTask := make([]*conf.Acg2020Task, 0, taskNum)
	FinishRoundID := make([]int, 0, taskNum)
	roundAward := make(map[int]int)
	for _, t := range s.c.Acg2020.Task {
		// 结束时间偏差一天，开始时间偏差1秒
		t.StatTime = t.StatTime.Add(time.Second)
		t.EndTime = t.EndTime.Add(time.Second)
		t.StartTime = t.StartTime.Add(-time.Second)
		statEndTime := t.StatTime.Add(time.Duration(s.c.Acg2020.UpdateDuration))
		if t.StartTime.Before(now) {
			// 阶段任务已经开始
			if statEndTime.After(now) {
				// 往后偏差一次执行周期，看看奖励统计截止时间是否到了
				runningTask = append(runningTask, t)
			} else {
				FinishRoundID = append(FinishRoundID, t.Round)
			}
		}
		roundAward[t.Round] = t.Award
	}

	if len(runningTask) == 0 {
		return nil
	}

	// 根据母活动拉取数据源基础信息
	subject, err := s.dao.ActSubject(ctx, s.c.Acg2020.SID)
	if err != nil {
		log.Errorc(ctx, "UpdateTaskState s.dao.ActSubject(ctx, %d) err[%v]", s.c.Acg2020.SID, err)
		return err
	}
	if subject.ChildSids == "" {
		log.Errorc(ctx, "UpdateTaskState subject.ChildSids empty")
		return err
	}
	sids := make([]int64, 0, strings.Count(subject.ChildSids, ",")+1)
	for _, str := range strings.Split(subject.ChildSids, ",") {
		sid, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			log.Errorc(ctx, "UpdateTaskState strconv.ParseInt(str, 10, 64) err[%v]", str, err)
			return err
		}
		sids = append(sids, sid)
	}

	eg := errgroup.WithContext(ctx)
	chanArcCalc := make(chan *ArcDetail)
	chanArcGrpc := make(chan []*l.Item)

	midStat := make(map[int64]*UserTaskState)
	// 稿件维度计算数据,根据mid聚合用户维度信息
	eg.Go(func(ctx context.Context) error {
		for one := range chanArcCalc {
			p, ok := midStat[one.Mid]
			if !ok {
				p = &UserTaskState{
					Mid:  one.Mid,
					Task: make([]*TaskState, taskNum, taskNum),
				}
				for i := 0; i < taskNum; i++ {
					p.Task[i] = &TaskState{}
				}
			}
			if one.Arc != nil {
				for _, t := range runningTask {
					if t.StartTime.Before(one.Arc.Ctime.Time()) && t.EndTime.After(one.Arc.Ctime.Time()) {
						score := Score(one.Arc)
						q := p.Task[t.Round]
						q.Count++
						q.Score += score
						if q.Count >= t.FinishCount && q.Score >= t.FinishScore {
							q.Finish = true
						}
						break
					}
				}
			}
			midStat[one.Mid] = p
		}
		return nil
	})

	// 根据稿件拉取稿件详细数据
	eg.Go(func(ctx context.Context) error {
		defer func() {
			close(chanArcCalc)
			for range chanArcGrpc {
			}
		}()

		for items := range chanArcGrpc {
			req := &arcapi.ArcsRequest{
				Aids: make([]int64, 0, len(items)),
			}
			for _, item := range items {
				req.Aids = append(req.Aids, item.Wid)
			}
			arcsReply, err := s.arcClient.Arcs(ctx, req)
			if err != nil {
				log.Errorc(ctx, "UpdateTaskState s.arcClient.Arcs(ctx, %v) error(%v)", req.Aids, err)
				return err
			}
			for _, item := range items {
				if arc, ok := arcsReply.Arcs[item.Wid]; ok {
					// 稿件状态过滤
					if arc.IsNormal() {
						chanArcCalc <- &ArcDetail{
							Item: item,
							Arc:  arc,
						}
						continue
					}
				} else {
					log.Warnc(ctx, "UpdateTaskState aid(%d) not found", item.Wid)
				}
				chanArcCalc <- &ArcDetail{
					Item: item,
				}
			}
		}
		return nil
	})

	// 拉取多个数据源稿件列表
	eg.Go(func(ctx context.Context) error {
		defer close(chanArcGrpc)
		tmp := make([]*l.Item, 0, batchArcQuery)
		s.dao.ScanLikesBySID(ctx, sids, batchDBQuery, func(ctx context.Context, items []*l.Item) {
			for _, item := range items {
				// 稿件状态过滤
				if item.State != 1 {
					chanArcCalc <- &ArcDetail{
						Item: item,
					}
					continue
				}
				tmp = append(tmp, item)
				if len(tmp) >= batchArcQuery {
					chanArcGrpc <- tmp
					tmp = make([]*l.Item, 0, batchArcQuery)
				}
			}
		})
		if len(tmp) > 0 {
			chanArcGrpc <- tmp
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Errorc(ctx, "UpdateTaskState eg.Wait error(%v)", err)
		return err
	}

	// 加载用户已经完成的任务数据，支持后面重算获奖金额
	if len(FinishRoundID) > 0 {
		f, err := os.Open(s.c.Acg2020.DataFile)
		if err != nil {
			log.Warnc(ctx, "UpdateTaskState os.Open(%s) error(%v)", s.c.Acg2020.DataFile, err)
			// 改从redis加载
			conn := s.redis.Get(ctx)
			defer conn.Close()
			for _, p := range midStat {
				reply, err := redis.Bytes(conn.Do("GET", s.redisKeyUserTaskState(p.Mid)))
				if err != nil {
					if err == redis.ErrNil {
						continue
					}
					log.Errorc(ctx, "UpdateTaskState conn.Do(GET, %s error(%v)", s.redisKeyUserTaskState(p.Mid), err)
					return err
				}
				one := &UserTaskState{}
				if err := json.Unmarshal(reply, &one); err != nil {
					log.Errorc(ctx, "UpdateTaskState  json.Unmarshal(%s) error(%v)", reply, err)
					return err
				}
				for _, round := range FinishRoundID {
					p.Task[round] = one.Task[round]
				}
			}
		} else {
			// 从文件快速加载上次结果
			defer f.Close()
			buf := bufio.NewReader(f)
			for {
				b, _, err := buf.ReadLine()
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Errorc(ctx, "UpdateTaskState buf.ReadLine() error(%v)", err)
					return err
				}
				one := &UserTaskState{}
				if err := json.Unmarshal(b, &one); err != nil {
					log.Errorc(ctx, "UpdateTaskState  json.Unmarshal(%s) error(%v)", b, err)
					return err
				}
				p, ok := midStat[one.Mid]
				if !ok {
					log.Errorc(ctx, "oh no, why, what happened")
					return errors.New("oh no, why, what happened")
				}
				for _, round := range FinishRoundID {
					p.Task[round] = one.Task[round]
				}
			}
		}
	}

	// 聚合活动维度完成总数信息
	roundCount := make([]int, taskNum, taskNum)
	extraCount := make([]int, taskNum+1, taskNum+1)
	for _, one := range midStat {
		for round, t := range one.Task {
			if t.Finish {
				one.FinishTask++
				roundCount[round]++
			}
		}
		extraCount[one.FinishTask]++
	}
	for i := taskNum; i > 0; i-- {
		extraCount[i-1] += extraCount[i]
	}

	// 计算用户基础获奖金额
	for _, one := range midStat {
		for round, t := range one.Task {
			if t.Finish {
				one.Money += roundAward[round] / roundCount[round]
			}
		}
	}

	// 计算用户附加获奖金额
	isLastRound := len(runningTask) == 1 && runningTask[0].Round == taskNum-1
	if isLastRound {
		// 只在最后一轮更新附加奖励
		for _, one := range midStat {
			for _, extra := range s.c.Acg2020.ExtraAward {
				if one.FinishTask >= extra.FinishTask {
					one.Money += extra.Award / extraCount[extra.FinishTask]
				}
			}
		}
	}

	// 明细数据落本地文件
	f, err := os.Create(s.c.Acg2020.DataFile)
	if err != nil {
		log.Errorc(ctx, "UpdateTaskState os.Open(%s) error(%v)", s.c.Acg2020.DataFile, err)
		return err
	}
	defer f.Close()
	for _, one := range midStat {
		b, err := json.Marshal(one)
		if err != nil {
			log.Errorc(ctx, "UpdateTaskState json.Marshal(%v) error(%v)", one, err)
			return err
		}
		if _, err := f.Write(b); err != nil {
			log.Errorc(ctx, "UpdateTaskState f.Write(%s) error(%v)", b, err)
			return err
		}
		if _, err := f.WriteString("\n"); err != nil {
			log.Errorc(ctx, "UpdateTaskState f.WriteString(\\n) error(%v)", err)
			return err
		}
	}

	// 更新redis
	conn := s.redis.Get(ctx)
	defer conn.Close()
	for _, one := range midStat {
		key := s.redisKeyUserTaskState(one.Mid)
		b, err := json.Marshal(one)
		if err != nil {
			log.Errorc(ctx, "UpdateTaskState json.Marshal(%v) error(%v)", one, err)
			return err
		}
		if err := conn.Send("SET", key, b); err != nil {
			log.Errorc(ctx, "UpdateTaskState conn.Send(SET, %s, %s) error(%v)", key, b, err)
			return err
		}
	}
	conn.Flush()
	for range midStat {
		if _, err := conn.Receive(); err != nil {
			if err != nil {
				log.Errorc(ctx, "UpdateTaskState conn.Receive() error(%v)", err)
				return err
			}
		}
	}

	// 判断是否需要邮件发送最终结果
	exportRound := -1
	for _, t := range runningTask {
		if now.After(t.StatTime) {
			exportRound = t.Round
			break
		}
	}
	if exportRound >= 0 {
		return s.sendMail(ctx, midStat, roundAward, roundCount, extraCount, isLastRound, exportRound)
	}

	return nil
}

func (s *Service) sendMail(ctx context.Context, midStat map[int64]*UserTaskState, roundAward map[int]int, roundCount, extraCount []int, isLastRound bool, exportRound int) error {
	filename := s.c.Acg2020.CsvFile
	f, err := os.Create(filename)
	if err != nil {
		log.Errorc(ctx, "UpdateTaskState os.Open(%s) error(%v)", filename, err)
		return err
	}
	defer f.Close()
	f.Write([]byte("\xEF\xBB\xBF"))
	writer := csv.NewWriter(f)
	writer.Write([]string{
		"用户mid",
		"任务1是否完成",
		"任务1稿件数",
		"任务1得分数",
		"任务1瓜分金额",
		"任务2是否完成",
		"任务2稿件数",
		"任务2得分数",
		"任务2瓜分金额",
		"任务3是否完成",
		"任务3稿件数",
		"任务3得分数",
		"任务3瓜分金额",
		"两次任务是否完成",
		"完成两次任务瓜分金额",
		"三次任务是否完成",
		"完成三次任务瓜分金额",
		"获得的总钱数",
	})
	for _, one := range midStat {
		record := make([]string, 0, 11)
		record = append(record, fmt.Sprint(one.Mid))
		for i := 0; i < len(s.c.Acg2020.Task); i++ {
			if one.Task[i].Finish {
				record = append(record, "1")
			} else {
				record = append(record, "0")
			}
			record = append(record, fmt.Sprint(one.Task[i].Count), fmt.Sprint(one.Task[i].Score))
			if one.Task[i].Finish {
				record = append(record, fmt.Sprintf("%.2f", float64(roundAward[i]/roundCount[i])/100))
			} else {
				record = append(record, "0")
			}
		}
		if isLastRound {
			for _, extra := range s.c.Acg2020.ExtraAward {
				if one.FinishTask >= extra.FinishTask {
					record = append(record, "1", fmt.Sprintf("%.2f", float64(extra.Award/extraCount[extra.FinishTask])/100))
				} else {
					record = append(record, "0", "0")
				}
			}
			record = append(record, fmt.Sprintf("%.2f", float64(one.Money)/100))
		} else {
			for range s.c.Acg2020.ExtraAward {
				record = append(record, "", "")
			}
			record = append(record, "")
		}
		writer.Write(record)
	}
	writer.Flush()
	if err := s.mail.SendAttachMail(ctx, s.c.Acg2020.Mail.Receiver,
		fmt.Sprintf("%s[第%d轮]", s.c.Acg2020.Mail.Subject, exportRound+1),
		fmt.Sprintf("%s[第%d轮]", s.c.Acg2020.Mail.Content, exportRound+1),
		&xmail.Attach{
			Name: strings.ReplaceAll(filepath.Base(filename), ".csv", fmt.Sprintf(".%d.csv", exportRound+1)),
			File: filename,
		}); err != nil {
		log.Errorc(ctx, "UpdateTaskState s.mail.SendAttachMail error(%v)", err)
		return err
	}
	return nil
}

func Score(arc *arcapi.Arc) int64 {
	if arc.Stat.View == 0 {
		return 0
	}
	return PlayScore(arc) + OtherScore(arc)
}

func PlayScore(arc *arcapi.Arc) int64 {
	videos := float64(arc.Videos)
	views := float64(arc.Stat.View)
	pRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", 4/(videos+3)), 64)
	aRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", (300000+views)/(2*views)), 64)
	if aRevise > 1 {
		aRevise = 1
	}
	return int64(math.Floor(views*pRevise*aRevise + 0.5))
}

func OtherScore(arc *arcapi.Arc) int64 {
	fav := int64(arc.Stat.Fav)
	coin := int64(arc.Stat.Coin)
	views := int64(arc.Stat.View)
	like := int64(arc.Stat.Like)
	share := int64(arc.Stat.Share)
	return (like*5 + coin*10 + fav*20 + share*50) * (like*5 + coin*10 + fav*20 + share*50) / (views + like*5 + coin*10 + fav*20 + share*50)
}
