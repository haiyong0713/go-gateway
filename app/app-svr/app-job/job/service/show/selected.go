package show

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-common/library/sync/errgroup.v2"

	xecode "go-gateway/app/app-svr/app-card/ecode"
	"go-gateway/app/app-svr/app-job/job/model"
	"go-gateway/app/app-svr/app-job/job/model/show"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

const (
	_sTypeWeeklySelected      = "weekly_selected"
	_aggregation              = "hotword"
	_archiveHonorUpdate       = "update"
	_archiveHonorDelete       = "delete"
	_archiveHonorWeeklyLink   = "?num=%d&navhide=1"
	_archiveHonorWeeklyLinkV2 = "?current_tab=week-%d"
)

// try 5 times to send ArchiveHonorSub
// nolint:bilirailguncheck
func (s *Service) sendArcHonor(ctx context.Context, msg *show.HonorMsg) (err error) {
	if !model.EnvRun() { // 非云立方不发送
		return
	}
	for i := 0; i < 5; i++ {
		if err = s.archiveHonorPub.Send(ctx, strconv.FormatInt(msg.Aid, 10), msg); err == nil {
			log.Info("sendArcHonor msg %+v succ", msg)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Error("sendArcHonor msg %+v error %v", msg, err)
	return
}

func (s *Service) alertAI(sType string) (err error) {
	var (
		serie *show.Serie
		ctx   = context.Background()
		aiCnt int
	)
	log.Info("AlertAI Begin")
	// pick the serie according to the insertion time
	if serie, err = s.dao.PickSerie(ctx, sType); err != nil {
		log.Error("AlertAI sType %s, Err %v", sType, err)
		return
	}
	// if there are already some AI data, we refuse new data
	if aiCnt, err = s.dao.AICount(ctx, serie.ID); err != nil {
		log.Error("AlertAI serie.ID %d, Err %v", serie.ID, err)
		return
	}
	if aiCnt == 0 {
		if err = s.dao.MerakNotify(ctx, s.c.WechatAlert.AI); err != nil {
			log.Error("AlertAI serie.ID %d, MerakNotify Err %v", serie.ID, err)
		}
		return
	}
	log.Info("AlertAI Succ, SID %d, AI Data Count %d", serie.ID, aiCnt)
	return
}

func (s *Service) alertAuditor(sType string) (err error) {
	var (
		serie *show.Serie
		ctx   = context.Background()
	)
	log.Info("alertAuditor Begin")
	// pick the serie according to the insertion time
	if serie, err = s.dao.PickSerie(ctx, sType); err != nil {
		log.Error("AlertAI sType %s, Err %v", sType, err)
		return
	}
	if !serie.Passed() {
		if err = s.dao.MerakNotify(ctx, s.c.WechatAlert.Audit); err != nil {
			log.Error("AlertAuditor serie.ID %d, MerakNotify Err %v", serie.ID, err)
		}
		return
	}
	log.Info("AlertAuditor Succ, SID %d, Auditing Status %d", serie.ID, serie.Status)
	return
}

func (s *Service) weeklyInsertion() {
	defer s.waiter.Done()
	var (
		ctx = context.Background()
		err error
	)
	for {
		msg, ok := <-s.weeklySelSub.Messages()
		if !ok {
			log.Error("[databus: app-job weeklyInsertion] consumer exit!")
			return
		}
		_ = msg.Commit()
		// analyse the AI data and insert them into DB
		log.Info("[databus: app-job weeklyInsertion] New Message: %s", msg.Value)
		if err = s.treatAIData(ctx, msg.Value); err != nil {
			continue
		}
	}
}

func (s *Service) allowAI(ctx context.Context, sType string) (allow bool, sid int64, err error) {
	var (
		serie *show.Serie
		aiCnt int
	)
	// pick the serie according to the insertion time
	if serie, err = s.dao.PickSerie(ctx, sType); err != nil {
		return
	}
	// if there are already some AI data, we refuse new data
	if aiCnt, err = s.dao.AICount(ctx, serie.ID); err != nil {
		return
	}
	if aiCnt > 0 {
		err = xecode.AIDataExist
		return
	}
	allow = true
	sid = serie.ID
	return
}

// insertRes inserts the resources data into the given serie
func (s *Service) insertRes(ctx context.Context, sid int64, selRes []*show.SerieRes) (err error) {
	var maxPos int64
	if maxPos, err = s.dao.MaxPosition(ctx, sid); err != nil {
		return
	}
	return s.dao.AIInsertion(ctx, sid, maxPos, selRes)
}

func (s *Service) treatAIData(ctx context.Context, value json.RawMessage) (err error) {
	var (
		sid     int64
		aiAllow bool
		selRes  []*show.SerieRes
		m       = &show.AIPopular{}
	)
	if err = json.Unmarshal(value, m); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", value, err)
		return
	}
	switch m.Route {
	case _sTypeWeeklySelected: // 每周必看
		if aiAllow, sid, err = s.allowAI(ctx, _sTypeWeeklySelected); err != nil || !aiAllow {
			log.Error("[databus: app-job weeklyInsertion] AI Data Refuse, Possible Err %v", err)
			return
		}
		weeklySel := &show.AIWeeklyItems{}
		if err = json.Unmarshal(value, weeklySel); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", value, err)
			return
		}
		if len(weeklySel.List) == 0 {
			log.Error("[databus: app-job weeklyInsertion] AI Data Empty")
			return
		}
		for _, item := range weeklySel.List { // travel aids from AI
			var resource = &show.SerieRes{}
			resource.FromAv(item.AID)
			selRes = append(selRes, resource)
		}
		if err = s.insertRes(ctx, sid, selRes); err != nil {
			log.Error("[databus: app-job weeklyInsertion] Sid %d,  DBInsertion Err %v!", sid, err)
		}
	case _aggregation: // 热门热点
		aggregation := &show.AIAggregationItems{}
		if err = json.Unmarshal(value, aggregation); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", value, err)
			return
		}
		if len(aggregation.List) == 0 {
			log.Error("[databus: app-job aggregation] AI Data Empty")
			return
		}
		for _, item := range aggregation.List {
			s.dealHotWord(ctx, item)
		}
	default:
		log.Warn("[databus: app-job] Route is empty !")
	}
	return
}

// consumePopular consumes binlog message
func (s *Service) consumePopular() {
	defer s.waiter.Done()
	for {
		msg, ok := <-s.selResBinlog.Messages()
		if !ok {
			log.Info("databus: consumePopular consumer exit!")
			return
		}
		_ = msg.Commit()
		log.Info("[consumePopular] New Message: %s", msg.Value)
		var (
			m   = &show.DatabusRes{}
			err error
			ctx = context.Background()
		)
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", msg, err)
			continue
		}
		switch m.Table {
		case "hotword_aggregation": // 热门热点
			agg := &show.HotWordDatabus{}
			if err = json.Unmarshal(msg.Value, agg); err != nil {
				log.Error("json.Unmarshal(%s) error(%v)", msg.Value, err)
				continue
			}
			if agg.New != nil { // mc 存储最新的
				s.addHotMc(ctx, agg.New)
			}
		case "good_history": // 入站必刷
			if !model.EnvRun() { // 仅云立方更新播单+成就系统即可
				continue
			}
			his := new(show.GoodHisDatabus)
			if err = json.Unmarshal(msg.Value, his); err != nil {
				log.Error("json.Unmarshal(%s) error(%v)", msg.Value, err)
				continue
			}
			if his == nil || his.New == nil {
				continue
			}
			s.dealGoodHis(ctx, his)
		}
	}
}

func (s *Service) dealHotWord(ctx context.Context, item *show.AIAggregation) {
	var (
		err      error
		tagIDcnt int
		tagRes   *taggrpc.TagReply
		hotRes   []*show.Aggregation
	)
	gp := errgroup.Group{}
	gp.Go(func(ctx context.Context) error {
		if tagIDcnt, err = s.dao.TagIDCount(ctx, item.TagID); err != nil {
			log.Error("[treatAIData] s.dao.TagIDCount() error(%v)", err)
			return err
		}
		if tagIDcnt != 0 {
			log.Warn("[treatAIData] tagID(%d) already exist !", item.TagID)
			err = fmt.Errorf("treatAIData tagID(%d) already exist", item.TagID)
			return err
		}
		return nil
	})
	gp.Go(func(ctx context.Context) error {
		if tagRes, err = s.dao.Tag(ctx, item.TagID); err != nil {
			log.Error("[treatAIData] s.dao.Tag(%d) error(%v)", item.TagID, err)
		}
		return err
	})
	if err = gp.Wait(); err != nil {
		log.Error("gp.Wait() error(%v)", err)
		return
	}
	if tagRes == nil || tagRes.Tag == nil {
		log.Warn("[treatAIData] tag is not exist ! TagID(%d)", item.TagID)
		return
	}
	if hotRes, err = s.dao.HotWordName(ctx, tagRes.Tag.Name); err != nil {
		log.Error("[treatAIData] s.dao.HotWordName(%s) error(%v)", tagRes.Tag.Name, err)
		return
	}
	if len(hotRes) == 0 { // tagid不存在,直接插入
		if err = s.dao.AddAggregation(ctx, tagRes.Tag.Name, item.TagID); err != nil {
			log.Error("[treatAIData] s.dao.AddAggregation() TagID(%d) error(%v)", item.TagID, err)
			return
		}
	} else if len(hotRes) == 1 && hotRes[0].State == show.NoAuditing { // tagid存在且未审核，插入
		if err = s.dao.AddTagID(ctx, hotRes[0].ID, item.TagID); err != nil {
			log.Error("[treatAIData] s.dao.AddTagID() TagID(%d) error(%v)", item.TagID, err)
		}
	} else { // 企业微信警告
		wechatMsg := *s.c.WechatAlert.Aggregation
		wechatMsg.Template = fmt.Sprintf(wechatMsg.Template, item.TagID, tagRes.Tag.Name)
		if err = s.dao.MerakNotify(ctx, &wechatMsg); err != nil {
			log.Error("AlertAI MerakNotify Err %v", err)
		}
	}
}

func (s *Service) selectedHonorURL(number int64) string {
	return s.c.WeeklySel.HonorLink + fmt.Sprintf(_archiveHonorWeeklyLink, number)
}

func (s *Service) selectedNaHonorURL(number int64) string {
	return s.c.WeeklySel.HonorLinkV2 + fmt.Sprintf(_archiveHonorWeeklyLinkV2, number)
}

func (s *Service) addHotMc(ctx context.Context, agg *show.Aggregation) {
	var err error
	for i := 0; i < 3; i++ {
		if err = s.dao.AddHotMC(ctx, agg); err == nil {
			return
		}
	}
	log.Error("[addHotMc] addHotMC error(%v)", err)
}

func (s *Service) refreshHonorLink(number, maxNumber int64) {
	s.waiter.Add(1)
	defer s.waiter.Done()
	for {
		if number > maxNumber {
			break
		}
		var res []*show.SerieRes
		retryErr := retry.WithAttempts(context.Background(), "refreshHonorLink", 3, netutil.DefaultBackoffConfig, func(context.Context) error {
			serieId, err := s.dao.SerieID(context.Background(), number)
			if err != nil {
				if err == sql.ErrNoRows {
					return nil
				}
				return err
			}
			res, err = s.dao.SerieRes(context.Background(), serieId)
			if err != nil {
				return err
			}
			for _, v := range res {
				if !v.IsArc() {
					continue
				}
				msg := new(show.HonorMsg)
				msg.FromWeeklySelected(v.RID, number, s.selectedHonorURL(number), _archiveHonorUpdate, s.selectedNaHonorURL(number))
				err = s.sendArcHonor(context.Background(), msg)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if retryErr != nil {
			log.Error("日志告警 retry three times error(%+v)", retryErr)
		}
		number++
	}
}
