package like

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) ClockInTag(c context.Context, mid int64) ([]*api.ClockInTag, error) {
	sids := s.clockInSubIDs
	if len(sids) == 0 {
		return nil, xecode.NothingFound
	}
	mu := sync.Mutex{}
	newReply := make(map[int64]*likemdl.HasReserve)
	eg := errgroup.WithContext(c)
	for _, vSid := range sids {
		tmpID := vSid
		eg.Go(func(ctx context.Context) error {
			newR, e := s.dao.ReserveOnly(ctx, tmpID, mid)
			if e != nil {
				log.Error("ClockInTag s.dao.ReserveOnly(%v,%d) error(%v)", tmpID, mid, e)
				return nil
			}
			mu.Lock()
			newReply[tmpID] = newR
			mu.Unlock()
			return nil
		})
	}
	eg.Wait()
	var reserveSids []int64
	for _, sid := range sids {
		if repet, ok := newReply[sid]; ok && repet != nil && repet.ID > 0 && repet.State == 1 {
			reserveSids = append(reserveSids, sid)
		}
	}
	if len(reserveSids) == 0 {
		return nil, nil
	}
	rules, err := s.dao.RawSubjectRulesBySids(c, sids)
	if err != nil {
		log.Error("ClockInTag s.dao.RawSubjectRulesBySids sids(%v) error(%v)", sids, err)
		return nil, err
	}
	// 兼容老版本 如果ctime和etime不等于0000-00-00 00:00:00的话 就是新数据 新数据需要过滤
	var clockInTags []*api.ClockInTag
	now := time.Now().Unix()
	for _, v := range reserveSids {
		if oneRules, ok := rules[v]; ok {
			for _, rule := range oneRules {
				if rule == nil || rule.Tags == "" {
					continue
				}
				// 冻结的tag在这里面不返回 1 正常 2 冻结
				if rule.State == 2 {
					continue
				}
				// 新版返回tag数据需要判断规则时间段
				if rule.Stime > 0 || rule.Etime > 0 {
					if now > rule.Etime.Time().Unix() || now < rule.Stime.Time().Unix() {
						continue
					}
				}
				tagArr := strings.Split(rule.Tags, ",")
				clockInTags = append(clockInTags, &api.ClockInTag{
					TypeIDs: rule.TypeIds,
					Tags:    tagArr[0],
				})
			}
		}
	}
	return clockInTags, nil
}

func (s *Service) loadClockInSubIDs() {
	now := time.Now()
	ids, err := s.dao.RawClockInSubIDs(context.Background(), now)
	if err != nil {
		log.Error("loadClockInSubIDs s.dao.RawClockInSubIDs error(%v)", err)
		return
	}
	if len(ids) == 0 {
		log.Warn("loadClockInSubIDs len(ids) == 0")
	}
	s.clockInSubIDs = ids
}

func (s *Service) checkRuleCounter(counter []*api.SubjectRuleCounter) error {
	if len(counter) == 0 {
		return xecode.Error(xecode.RequestErr, "counter 不能为空")
	}
	for _, c := range counter {
		if c.Name == "" {
			return xecode.Error(xecode.RequestErr, "counter.Name 不能为空")
		}
		if c.Category < 1 {
			return xecode.Error(xecode.RequestErr, "counter.Category 取值错误")
		}
		if c.Sids != "" {
			for _, s := range strings.Split(c.Sids, ",") {
				if p, err := strconv.Atoi(s); err != nil || p <= 0 {
					return xecode.Error(xecode.RequestErr, "counter.Sids 必须为正整数，多个使用逗号分隔")
				}
			}
		}
		if c.Coefficient == "" {
			return xecode.Error(xecode.RequestErr, "counter.Coefficient 不能为空")
		}
		for _, s := range strings.Split(c.Coefficient, ",") {
			if p, err := strconv.ParseFloat(s, 64); err != nil || p < 0 {
				return xecode.Error(xecode.RequestErr, "counter.Coefficient 必须为数字，多个使用逗号分隔")
			}
		}
	}
	return nil
}

// SyncSubjectRules ...
func (s *Service) SyncSubjectRules(c context.Context, sid int64, counter []*api.SubjectRuleCounter) (err error) {
	if err = s.checkRuleCounter(counter); err != nil {
		return
	}
	var subject *likemdl.SubjectItem
	if subject, err = s.dao.ActSubject(c, sid); err != nil {
		log.Error("SyncSubjectRules:s.dao.ActSubject(%d) error(%+v)", sid, err)
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityNotExist
		return
	}
	if subject.Type != likemdl.USERACTIONSTAT {
		err = ecode.ActivityNotExist
		return
	}
	var rules []*likemdl.SubjectRule
	rules, err = s.dao.RawSubjectRulesBySid(c, sid)
	if err != nil {
		log.Error("SyncSubjectRules s.dao.RawSubjectRulesBySids sids(%v) error(%v)", sid, err)
		return
	}
	ins := make([]*likemdl.SubjectRule, 0, len(counter))
	for _, c := range counter {
		update := false
		for _, r := range rules {
			if c.Name == r.RuleName {
				update = true
				r.Attribute = c.Attribute
				r.Coefficient = c.Coefficient
				r.Sids = c.Sids
				r.State = c.State
				r.Tags = c.Tags
				r.TypeIds = c.TypeIDs
				if c.Fav != nil {
					r.AidSourceType = likemdl.RuleAidSourceTypeFav
					favByte, err := json.Marshal(c.Fav)
					if err == nil {
						r.AidSource = string(favByte)
					}
				}
			}
		}
		if !update {
			var aidSourceType int64
			var aidSource string
			if c.Fav != nil {
				aidSourceType = likemdl.RuleAidSourceTypeFav
				favByte, err := json.Marshal(c.Fav)
				if err == nil {
					aidSource = string(favByte)
				}
			}
			ins = append(ins, &likemdl.SubjectRule{
				Sid:           sid,
				Type:          c.Category,
				TypeIds:       c.TypeIDs,
				Tags:          c.Tags,
				State:         c.State,
				Attribute:     c.Attribute,
				RuleName:      c.Name,
				Sids:          c.Sids,
				Coefficient:   c.Coefficient,
				AidSourceType: aidSourceType,
				AidSource:     aidSource,
			})
		}
	}
	if len(rules) > 0 {
		// 触发更新
		for _, r := range rules {
			if err := s.dao.UpdateSubjectRule(c, r); err != nil {
				log.Error("SyncSubjectRules s.dao.UpdateSubjectRule error(%v)", err)
				return err
			}
		}
	}
	if len(ins) > 0 {
		// 写入task，关联rule
		for _, r := range ins {
			t, err := s.taskDao.AddTask(c, "user_action_stat_sub_rule", _taskBusinessID, r.Sid, 0)
			if err != nil {
				log.Error("SyncSubjectRules s.taskDao.AddTask error(%v)", err)
				return err
			}
			r.TaskID = t.ID
		}
		// 触发插入
		return s.dao.InsertSubjectRules(c, ins)
	}
	return nil
}
