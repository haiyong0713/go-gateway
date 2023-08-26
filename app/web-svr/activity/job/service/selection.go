package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	tagnewapi "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/job/model/like"
	mdlmail "go-gateway/app/web-svr/activity/job/model/mail"

	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

const (
	_selectionScene   = "cartoon_vote"
	_selectionOffline = "cartoon_vote_hit"
)

func (s Service) SelAssistance() {
	var (
		ctx = context.Background()
	)
	log.Infoc(ctx, "SelAssistance s.dao.SelProductRoles start")
	prList, err := s.dao.SelProductRoles(ctx)
	if err != nil {
		log.Errorc(ctx, "SelAssistance s.dao.SelProductRoles error(%+v)", err)
		return
	}
	likeArcs, arcTags, err := s.actArchivesBySid(ctx, s.c.Selection.Sid)
	if err != nil {
		log.Errorc(ctx, "SelAssistance s.loadLikeList sid(%d) error(%v)", s.c.Selection.Sid, err)
		return
	}
	for _, productRole := range prList {
		tmpPr := productRole
		var (
			prTagsMap map[string]string
			product   string
		)
		if tmpPr.Tags == "" {
			continue
		}
		prTagsList := strings.Split(tmpPr.Tags, s.c.Selection.TagSplit)
		prTagsMap = make(map[string]string, len(prTagsList))
		if productRole.TagsType == 1 {
			product = productRole.Product
		}
		for _, prTag := range prTagsList {
			prTagsMap[strings.ToLower(prTag)] = product
		}
		s.prArcAssistance(ctx, tmpPr.ID, arcTags, likeArcs, prTagsMap)
	}
	log.Infoc(ctx, "SelAssistance s.dao.SelProductRoles finish success")
}

func (s *Service) prArcAssistance(ctx context.Context, productroleID int64, arcTags map[int64][]*tagnewapi.Tag, likeArcs map[int64]*arcmdl.Arc, prTagsMap map[string]string) (err error) {
	var prHots []*like.ProductRoleHot
	for _, arc := range likeArcs {
		if arc == nil {
			continue
		}
		if !arc.IsNormal() {
			log.Infoc(ctx, "SelAssistance prArcAssistance arc not is normal aid(%d)", arc.Aid)
			continue
		}
		currentArcTags, ok := arcTags[arc.Aid]
		if !ok {
			log.Infoc(ctx, "SelAssistance prArcAssistance currentArcTags aid(%d) not ok", arc.Aid)
			continue
		}
		for _, tagName := range currentArcTags {
			if product, ok := prTagsMap[strings.ToLower(tagName.Name)]; ok {
				if product != "" { // 还需要命中作品
					var isHitProduct bool
					for _, tagName := range currentArcTags {
						if strings.ToLower(tagName.Name) == strings.ToLower(product) {
							isHitProduct = true
							break
						}
					}
					if !isHitProduct {
						continue
					}
				}
				hotNum := arc.Stat.Like + arc.Stat.Coin*2
				prHots = append(prHots, &like.ProductRoleHot{
					Aid:     arc.Aid,
					PubDate: arc.PubDate.Time().Unix(),
					HotNum:  int64(hotNum),
				})
				break
			}
		}
	}
	if err = s.addAssistance(ctx, productroleID, prHots); err != nil {
		log.Errorc(ctx, "SelAssistance prArcAssistance s.loadLikeList sid(%d) productroleID(%d) error(%v)", s.c.Selection.Sid, productroleID, err)
	}
	return
}

func (s *Service) addAssistance(ctx context.Context, productroleID int64, prHots []*like.ProductRoleHot) (err error) {
	for i := 0; i < _retry; i++ {
		if err = s.dao.AddCacheAssistance(ctx, productroleID, prHots); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) actArchivesBySid(ctx context.Context, sid int64) (archives map[int64]*arcmdl.Arc, arcTags map[int64][]*tagnewapi.Tag, err error) {
	likeArcs, err := s.loadLikeList(ctx, sid, _retryTimes)
	if err != nil {
		log.Error("SelAssistance actArchivesBySid s.loadLikeList sid(%d) error(%v)", sid, err)
		return
	}
	var aids []int64
	for _, v := range likeArcs {
		if v != nil && v.Wid > 0 {
			aids = append(aids, v.Wid)
		}
	}
	aidsLen := len(aids)
	if aidsLen == 0 {
		log.Warn("SelAssistance actArchivesBySid len(aids) == 0")
		return
	}
	if archives, arcTags, err = s.archiveTagName(ctx, aids); err != nil {
		log.Errorc(ctx, "SelAssistance actArchivesBySid s.archiveTagName aidsLen(%d) error(%+v)", aidsLen, err)
	}
	return
}

// archiveTagName get archives and tag name.
func (s *Service) archiveTagName(c context.Context, aids []int64) (archives map[int64]*arcmdl.Arc, arcTags map[int64][]*tagnewapi.Tag, err error) {
	var (
		mutex   = sync.Mutex{}
		aidsLen = len(aids)
	)
	group := errgroup.WithContext(c)
	archives = make(map[int64]*arcmdl.Arc, aidsLen)
	arcTags = make(map[int64][]*tagnewapi.Tag, aidsLen)
	for i := 0; i < aidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		group.Go(func(ctx context.Context) (err error) {
			var arcs *arcmdl.ArcsReply
			arg := &arcmdl.ArcsRequest{Aids: partAids}
			if arcs, err = s.arcClient.Arcs(ctx, arg); err != nil {
				log.Error("SelAssistance s.arcClient.Arcs(%v) error(%v)", partAids, err)
				return
			}
			mutex.Lock()
			for _, v := range arcs.Arcs {
				archives[v.Aid] = v
			}
			mutex.Unlock()
			return
		})
		group.Go(func(ctx context.Context) error {
			reply, tagErr := s.tagNewClient.ResTags(ctx, &tagnewapi.ResTagsReq{Oids: partAids, Type: 3})
			if tagErr != nil {
				log.Error("SelAssistance s.tagClient.ResTags aids(%v) error(%v)", partAids, tagErr)
				return nil
			}
			if reply != nil {
				mutex.Lock()
				for oid, v := range reply.ResTags {
					arcTags[oid] = v.Tags
				}
				mutex.Unlock()
			}
			return nil
		})
	}
	err = group.Wait()
	return
}

// DayVoteReport
func (s *Service) DayVoteReport() {
	ctx := context.Background()
	if len(s.c.Selection.VoteStage) == 0 {
		log.Errorc(ctx, "DayVoteTable s.c.Selection.VoteStage empty")
		return
	}
	nowTime := time.Now()
	if nowTime.Unix() < s.c.Selection.VoteReportBegin {
		log.Warn("DayVoteTable activity not start")
		return
	}
	if nowTime.Unix() > s.c.Selection.VoteReportEnd {
		log.Warn("DayVoteTable activity end")
		return
	}
	// 获取前一日投票人数
	oldTime := nowTime.AddDate(0, 0, -1)
	oldDate := oldTime.Format("2006-01-02")
	midCount, e := s.dao.PrDayMidCnt(ctx, oldDate)
	if e != nil {
		log.Errorc(ctx, "DayVoteTable s.dao.PrDayMidCnt(%s) error(%+v)", oldDate, e)
		return
	}
	subject := oldDate + "参与投票的人数:" + strconv.Itoa(midCount)
	reportDay := nowTime.Day()
	selectionHeader := []string{"奖项类别", "角色名", "作品名"}
	month := nowTime.Format("01月")
	for i := 1; i < reportDay; i++ { //循环日期
		selectionHeader = append(selectionHeader, fmt.Sprintf(month+"%v日", i))
	}
	data := [][]string{}
	if s.c.Selection.ReportProduct == 1 || nowTime.Year() == 2021 {
		for _, category := range s.c.Selection.VoteStage {
			productRoleList, err := s.dao.SelProductRoleByCategory(ctx, category.CategoryID)
			if err != nil {
				log.Errorc(ctx, "DayVoteTable categoryID(%d) error(%+v)", category.CategoryID, err)
				return
			}
			for _, pr := range productRoleList {
				rows := []string{category.CategoryName, pr.Role, pr.Product}
				for i := 1; i < reportDay; i++ { //循环日期
					var dayCount int
					voteDate := fmt.Sprintf(nowTime.Format("2006-01")+"-%v", i)
					if dayCount, err = s.dao.PrDayCnt(ctx, pr.ID, voteDate); err != nil {
						log.Errorc(ctx, "DayVoteTable s.dao.PrDayCnt(%d) date(%s) error(%+v)", pr.ID, voteDate, err)
						return
					}
					rows = append(rows, strconv.Itoa(dayCount))
				}
				data = append(data, rows)
			}
		}
	}
	if err := s.selectionCreateCsvAndSendMail(ctx, subject, selectionHeader, data); err != nil {
		log.Errorc(ctx, "DayVoteTable selectionCreateCsvAndSendMail error(%+v)", err)
	}
}

// VoteReport
func (s *Service) VoteReport() {
	ctx := context.Background()
	if len(s.c.Selection.VoteStage) == 0 {
		log.Errorc(ctx, "VoteReport s.c.Selection.VoteStage empty")
		return
	}
	nowTime := time.Date(2020, 12, 31, 0, 0, 0, 0, time.Local)
	midCount, e := s.dao.PrDayMidCnt(ctx, "2020-12-31")
	if e != nil {
		log.Errorc(ctx, "VoteReport s.dao.PrDayMidCnt(%s) error(%+v)", "2020-12-31", e)
		return
	}
	subject := "2020-12-31参与投票的人数:" + strconv.Itoa(midCount)
	reportDay := nowTime.Day()
	selectionHeader := []string{"奖项类别", "角色名", "作品名"}
	month := "12月"
	for i := 1; i <= reportDay; i++ { //循环日期
		selectionHeader = append(selectionHeader, fmt.Sprintf(month+"%v日", i))
	}
	data := [][]string{}
	for _, category := range s.c.Selection.VoteStage {
		productRoleList, err := s.dao.SelProductRoleByCategory(ctx, category.CategoryID)
		if err != nil {
			log.Errorc(ctx, "VoteReport categoryID(%d) error(%+v)", category.CategoryID, err)
			return
		}
		for _, pr := range productRoleList {
			rows := []string{category.CategoryName, pr.Role, pr.Product}
			for i := 1; i <= reportDay; i++ { //循环日期
				var dayCount int
				voteDate := fmt.Sprintf(nowTime.Format("2006-01")+"-%v", i)
				if dayCount, err = s.dao.PrDayCnt(ctx, pr.ID, voteDate); err != nil {
					log.Errorc(ctx, "VoteReport s.dao.PrDayCnt(%d) date(%s) error(%+v)", pr.ID, voteDate, err)
					return
				}
				rows = append(rows, strconv.Itoa(dayCount))
			}
			log.Infoc(ctx, "VoteReport producdtID(%d)", pr.ID)
			data = append(data, rows)
		}
	}
	if err := s.selectionCreateCsvAndSendMail(ctx, subject, selectionHeader, data); err != nil {
		log.Errorc(ctx, "VoteReport selectionCreateCsvAndSendMail error(%+v)", err)
	}
}

func (s *Service) MidReport() {
	ctx := context.Background()
	voteMids, err := s.dao.VoteMids(ctx)
	if err != nil {
		log.Errorc(ctx, "MidReport s.dao.VoteMids error(%+v)", err)
		return
	}
	subject := "整个活动期间参与总人数:" + strconv.Itoa(len(voteMids))
	selectionHeader := []string{"mid", "参与总天数"}
	data := [][]string{}
	for _, mid := range voteMids {
		rows := []string{strconv.FormatInt(mid, 10)}
		midCount, e := s.dao.MidCountDays(ctx, mid)
		if e != nil {
			log.Errorc(ctx, "MidReport s.dao.MidCountDays mid(%d) error(%+v)", mid, e)
			return
		}
		rows = append(rows, strconv.FormatInt(midCount, 10))
		data = append(data, rows)
	}
	if err := s.selectionCreateCsvAndSendMail(ctx, subject, selectionHeader, data); err != nil {
		log.Errorc(ctx, "MidReport selectionCreateCsvAndSendMail error(%+v)", err)
	}
}

func (s *Service) selectionCreateCsvAndSendMail(c context.Context, subject string, selectionHeader []string, data [][]string) error {
	base := &mdlmail.Base{
		Host:    s.c.Mail.Host,
		Port:    s.c.Mail.Port,
		Address: s.c.Mail.Address,
		Pwd:     s.c.Mail.Pwd,
		Name:    s.c.Mail.Name,
	}
	filePath := "./data/"
	fileName := fmt.Sprintf("%v_%v.csv", "年度动画评选页面B", time.Now().Format("20060102"))
	return s.activityCreateCsvAndSend(c, filePath, fileName, subject, base, s.selectionReceivers(), []*mdlmail.Address{}, []*mdlmail.Address{}, selectionHeader, data)
}

func (s *Service) selectionReceivers() []*mdlmail.Address {
	var mailReceivers []*mdlmail.Address
	receivers := strings.Split(s.c.Selection.EmailReceivers, ",")
	for _, v := range receivers {
		user := &mdlmail.Address{
			Address: v,
			Name:    "",
		}
		mailReceivers = append(mailReceivers, user)
	}
	return mailReceivers
}

func (s *Service) gaiaRiskProc() {
	if s.gaiaRiskSub == nil {
		return
	}
	for {
		msg, ok := <-s.gaiaRiskSub.Messages()
		if !ok {
			break
		}
		if msg != nil && msg.Value != nil {
			log.Info("gaiaRiskProc receive msg: %s", string(msg.Value))
		}
		msg.Commit()
		m := new(like.GaiaRisk)
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			log.Error("gaiaRiskProc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		// 只处理离线数据
		if m.SceneName == _riskArticleDayAction && m.Decision == _riskArticleHitOffline {
			s.gaiaArticleRisk(m.DecisionCtx)
		}
	}
}

func (s *Service) gaiaRisk(m *like.RiskDecisionCtx) {
	var (
		ctx                   = context.Background()
		err                   error
		mid, categoryID, prID int64
	)
	if m == nil {
		log.Errorc(ctx, "gaiaRiskProc gaiaRisk m is nil")
		return
	}
	if mid, err = InterfaceToInt64(m.Mid); err != nil {
		log.Errorc(ctx, "gaiaRiskProc gaiaRisk Mid InterfaceToInt64 m(%+v)", m)
		return
	}
	if categoryID, err = InterfaceToInt64(m.CategoryID); err != nil {
		log.Errorc(ctx, "gaiaRiskProc gaiaRisk categoryID InterfaceToInt64 m(%+v)", m)
		return
	}
	if prID, err = InterfaceToInt64(m.ID); err != nil {
		log.Errorc(ctx, "gaiaRiskProc gaiaRisk ID InterfaceToInt64 m(%+v)", m)
		return
	}
	log.Info("gaiaRiskProc gaiaRisk  mid(%d) category_id(%d) productrole_id(%d) voteDate(%s)", mid, categoryID, prID, m.VoteDate)
	if mid == 0 || categoryID == 0 || prID == 0 || m.VoteDate == "" {
		log.Errorc(ctx, "gaiaRiskProc gaiaRisk s.dao.UpSelectionVoteRisk empty mid(%d) category_id(%d) productrole_id(%d) voteDate(%s)", mid, categoryID, prID, m.VoteDate)
		return
	}
	for i := 0; i < _retry; i++ {
		if err = s.dao.UpSelectionVoteRisk(ctx, mid, categoryID, prID, m.VoteDate); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		log.Errorc(ctx, "gaiaRiskProc gaiaRisk s.dao.UpSelectionVoteRisk mid(%d) category_id(%d) productrole_id(%d) voteDate(%s) error(%+v)", mid, categoryID, prID, m.VoteDate, err)
	}
}

func (s *Service) SetSelectionVoteCache() {
	if s.c.Selection.StopRiskVote == 1 { // 是否停止处理风控减票
		log.Warn("SetSelectionVoteCache is stop")
		return
	}
	ctx := context.Background()
	for categoryID := 1; categoryID <= 6; categoryID++ {
		if err := s.dao.ResetProductRoleVoteNum(ctx, categoryID); err != nil {
			log.Errorc(ctx, "gaiaRiskProc SetSelectionVoteCache s.dao.ResetProductRoleVoteNum categoryID(%d) error(%+v)", categoryID, err)
		}
		time.Sleep(time.Millisecond * 100)
	}
	log.Infoc(ctx, "SetSelectionVoteCache s.dao.ResetProductRoleVoteNum success")
}

func InterfaceToInt64(data interface{}) (value int64, err error) {
	switch data.(type) {
	case nil:
		value = int64(0)
	case int64:
		value = data.(int64)
	case int32:
		value = int64(data.(int32))
	case string:
		value, err = strconv.ParseInt(data.(string), 10, 64)
	case float64:
		value = int64(data.(float64))
	case float32:
		value = int64(data.(float32))
	case json.Number:
		r, _ := data.(json.Number)
		value, err = strconv.ParseInt(r.String(), 10, 64)
	default:
		err = errors.Errorf("error %+v to int64", data)
	}
	return
}
