package resource

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/queue/databus/databusutil"
	"go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	model "go-gateway/app/app-svr/app-feed/admin/model/resource"
	xtime "time"
)

type archiveEvent struct {
	Event     string `json:"event"`
	AdminID   int64  `json:"admin_id"`
	Aid       int64  `json:"aid"`
	AuditCode int32  `json:"reason_id"`
	Timestamp int64  `json:"timestamp"`
}

var (
	codePriorityMap map[int32]int32
	codeReasonMap   map[int32]string
)

// 404错误页中由于稿件审核状态引起的结果同步
func (s *Service) ConsumeArchiveAuditResult() {
	initBaseMap(s.dao.Conf.Error404Conf)

	dsSub := databus.New(s.dao.Conf.Error404Conf.Databus)
	//defer dsSub.Close()

	g := databusutil.NewGroup(nil, dsSub.Messages())

	g.New = parseMsg
	g.Split = shardingMsg
	g.Do = func(msg []interface{}) {
		s.processMsg(msg)
	}
	g.Start()
	//defer g.Close()
}

func initBaseMap(conf *conf.Error404Config) {
	codePriorityMap = make(map[int32]int32)
	codeReasonMap = make(map[int32]string)
	for _, c := range conf.AuditMap {
		for _, code := range c.Codes {
			codePriorityMap[code] = c.Priority
			codeReasonMap[code] = c.Reason
		}
	}
}

// 解析消息
func parseMsg(msg *databus.Message) (interface{}, error) {
	m := &archiveEvent{}
	if err := json.Unmarshal(msg.Value, &m); err != nil {
		log.Error("404 Consumer: json.Unmarshal(%s) error(%v)", msg.Value, err)
	}
	return m, nil
}

// 拆分消息
func shardingMsg(_ *databus.Message, data interface{}) int {
	msg, ok := data.(*archiveEvent)
	if !ok {
		return 0
	}
	return int(msg.Aid)
}

// 处理消息
func (s *Service) processMsg(msg []interface{}) {
	for _, m := range msg {
		data, ok := m.(*archiveEvent)
		//log.Info("404 Cousume data: %v", data)
		if !ok {
			continue
		}

		// 事件判定
		if data.Event == "up_delete_archive" {
			s.handleArchiveDeletedResult(data)
		} else if data.Event == "audit_video" || data.Event == "audit_archive" {
			s.handleAuditResult(data)
		}
	}
}

//nolint:gocognit
func (s *Service) handleAuditResult(m *archiveEvent) {
	// reason id判定
	var (
		content string
		hasCode bool
	)
	if content, hasCode = codeReasonMap[m.AuditCode]; !hasCode {
		//log.Warn("404 Consumer: read codeReasonMap error - code -%s", m.AuditCode)
		return
	}
	ctx := context.Background()
	operator := s.dao.Conf.Error404Conf.BaseConf.Operator
	if cc, err := s.dao.GetCustomConfigBy(ctx, 1, m.Aid); err != nil && err.Error() != "-404" {
		log.Error("404 Consumer: GetCustomConfigBy error - %s", err.Error())
		return
	} else {
		if cc == nil {
			// 对应aid没有干预数据，在新增记录后直接下一条
			newCC := &model.CCAddReq{
				TP:         1,
				OidNum:     m.Aid,
				Content:    content,
				STime:      time.Time(m.Timestamp),
				ETime:      time.Time(m.Timestamp + s.dao.Conf.Error404Conf.BaseConf.ETimeOffset),
				Operator:   operator,
				OperatorID: m.AdminID,
				AuditCode:  m.AuditCode,
				OriginType: 1,
			}
			if rows, err := s.CCAdd(ctx, newCC); err != nil {
				log.Error("404 Consumer: CCAdd error - %s", err.Error())
			} else if rows == 0 {
				//nolint:govet
				log.Error("404 Consumer: CCAdd fail - %s", newCC)
			}
			return
		} else {
			// step 1：判定是否有人工运营并且生效中
			if cc.OriginType == 0 && cc.State == 1 && cc.ETime.Time().Unix() > xtime.Now().Unix() {
				if err := s.ccAuditCodeUpdate(ctx, m.Aid, operator, m.AdminID, m.AuditCode); err != nil {
					log.Error("404 Consumer: CCUpdate error - %s", err.Error())
				}
				return
			}
			// step 2: 如果没有被操作过, 读取新旧优先级
			var (
				newPriority            int32
				oldPriority            int32
				readNewPrioritySuccess bool
				readOldPrioritySuccess bool
			)
			if newPriority, readNewPrioritySuccess = codePriorityMap[m.AuditCode]; !readNewPrioritySuccess {
				//nolint:govet
				log.Error("404 Consumer: read codePriorityMap error - code -%s", m.AuditCode)
				return
			}
			if oldPriority, readOldPrioritySuccess = codePriorityMap[cc.AuditCode]; !readOldPrioritySuccess {
				//nolint:govet
				log.Error("404 Consumer: read codePriorityMap error - code -%s", cc.AuditCode)
				oldPriority = 4
			}

			// step 3: 如果优先级没有原来的高并且在线的直接跳过, 否则修改现有数据
			if newPriority < oldPriority || cc.State == 0 || cc.ETime.Time().Unix() <= xtime.Now().Unix() {
				// step 2: 如果优先级比原来的高
				modifiedCC := &model.CCUpdateReq{
					TP:               1,
					STime:            time.Time(m.Timestamp),
					ETime:            time.Time(m.Timestamp + s.dao.Conf.Error404Conf.BaseConf.ETimeOffset),
					OidNum:           m.Aid,
					ID:               cc.ID,
					Content:          content,
					URL:              "",
					HighlightContent: "",
					Image:            "",
					ImageBig:         "",
					Operator:         operator,
					OperatorID:       m.AdminID,
					AuditCode:        m.AuditCode,
					OriginType:       1,
				}

				if err := s.CCUpdate(ctx, modifiedCC); err != nil {
					log.Error("404 Consumer: CCUpdate error - %s", err.Error())
					return
				}
				if cc.State == 0 {
					if err := s.ccEnable(ctx, cc.ID, operator, m.AdminID); err != nil {
						log.Error("404 Consumer: ccEnable error - %s", err.Error())
						return
					}
				}
			}
		}
	}
}

func (s *Service) handleArchiveDeletedResult(m *archiveEvent) {
	// reason id判定
	var (
		content string
		hasCode bool
	)
	if content, hasCode = codeReasonMap[-100]; !hasCode {
		//log.Warn("404 Consumer: read codeReasonMap error - code -%s", m.AuditCode)
		return
	}
	ctx := context.Background()
	operator := s.dao.Conf.Error404Conf.BaseConf.Operator
	if cc, err := s.dao.GetCustomConfigBy(ctx, 1, m.Aid); err != nil && err.Error() != "-404" {
		log.Error("404 Consumer: GetCustomConfigBy error - %s", err.Error())
		return
	} else {
		if cc == nil {
			// 对应aid没有干预数据，在新增记录后直接下一条
			newCC := &model.CCAddReq{
				TP:         1,
				OidNum:     m.Aid,
				Content:    content,
				STime:      time.Time(m.Timestamp),
				ETime:      time.Time(m.Timestamp + s.dao.Conf.Error404Conf.BaseConf.ETimeOffset),
				Operator:   operator,
				OperatorID: m.AdminID,
				AuditCode:  -100,
				OriginType: 1,
			}
			if rows, err := s.CCAdd(ctx, newCC); err != nil {
				log.Error("404 Consumer: CCAdd error - %s", err.Error())
			} else if rows == 0 {
				//nolint:govet
				log.Error("404 Consumer: CCAdd fail - %s", newCC)
			}
			return
		} else {
			// step 1：判定是否有人工运营且在生效中
			if cc.OriginType == 0 && cc.State == 1 && cc.ETime.Time().Unix() > xtime.Now().Unix() {
				if err := s.ccAuditCodeUpdate(ctx, m.Aid, operator, m.AdminID, -100); err != nil {
					log.Error("404 Consumer: CCUpdate error - %s", err.Error())
				}
				return
			}

			// step 2: 修改为up主自删除
			modifiedCC := &model.CCUpdateReq{
				TP:               1,
				STime:            time.Time(m.Timestamp),
				ETime:            time.Time(m.Timestamp + s.dao.Conf.Error404Conf.BaseConf.ETimeOffset),
				OidNum:           m.Aid,
				ID:               cc.ID,
				Content:          content,
				URL:              "",
				HighlightContent: "",
				Image:            "",
				ImageBig:         "",
				Operator:         operator,
				OperatorID:       m.AdminID,
				AuditCode:        -100,
				OriginType:       1,
			}

			if err := s.CCUpdate(ctx, modifiedCC); err != nil {
				log.Error("404 Consumer: CCUpdate error - %s", err.Error())
				return
			}
			if cc.State == 0 {
				if err := s.ccEnable(ctx, cc.ID, operator, m.AdminID); err != nil {
					log.Error("404 Consumer: ccEnable error - %s", err.Error())
					return
				}
			}
		}
	}
}
