package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/stat/prom"
	"go-gateway/app/web-svr/activity/ecode"
	bnjmdl "go-gateway/app/web-svr/activity/job/model/bnj"
	"go-gateway/app/web-svr/activity/job/model/like"
	suitapi "go-main/app/account/usersuit/service/api"

	cheeseapi "git.bilibili.co/bapis/bapis-go/cheese/service/coupon"

	"github.com/pkg/errors"
)

const (
	_awardTypePugvCoupon  = 3
	_awardTypeComicCoupon = 4
	_awardTypeMallCoupon  = 5
	_awardTypePendant     = 6
	_awardFinalType       = 7
	_mallCouponSourceID   = 4
	_mallCouponActivityId = "20Dispersing"
	_comicNum             = 10
	_pushTypeHotpot       = 1
	_pushTypeReserve      = 2
)

var awardTypes = []int{_awardTypePugvCoupon, _awardTypeComicCoupon, _awardTypeMallCoupon, _awardTypePendant, _awardFinalType}

func (s *Service) initBnj() {
	value, err := func() (amount int64, err error) {
		for i := 0; i < _retryTimes; i++ {
			if amount, err = s.bnj.RawCurrencyAmount(context.Background(), s.c.Bnj2020.Sid); err == nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		return
	}()
	if err != nil {
		panic(err)
	}
	atomic.StoreInt64(&s.bnjValue, value)
	if value >= s.c.Bnj2020.MaxValue {
		log.Warn("bnjproc bnj20Finish value(%d)", value)
	}
}

func (s *Service) bnjproc() {
	defer s.waiter.Done()
	if s.bnjSub == nil {
		return
	}
	var (
		lastAction = new(bnjmdl.Action)
		changeNum  int64
		pushAction *bnjmdl.Action
	)
	for {
		if s.c.Bnj2020.BlockGame == 1 {
			return
		}
		if s.bnjValue >= s.c.Bnj2020.MaxValue {
			log.Warn("bnjproc bnj20Finish")
			return
		}
		msg, ok := <-s.bnjSub.Messages()
		if !ok {
			log.Info("bnjproc databus exit!")
			return
		}
		msg.Commit()
		m := new(bnjmdl.Action)
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			log.Error("bnjproc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		switch m.Type {
		case bnjmdl.ActionTypeIncr:
			changeNum = changeNum + m.Num
		case bnjmdl.ActionTypeDecr:
			changeNum = changeNum - m.Num
		default:
			log.Error("bnjproc m.Type miss action(%+v)", m)
			continue
		}
		// 选取特殊事件
		if pushAction == nil && m.Message != "" {
			pushAction = m
		}
		if m.Ts > lastAction.Ts {
			s.dealBnjMsg(changeNum, pushAction)
			changeNum = 0
			pushAction = nil
			lastAction = m
		}
	}
}

func (s *Service) bnjAwardproc() {
	defer s.waiter.Done()
	if s.bnjAwardSub == nil {
		return
	}
	for {
		if s.closed {
			log.Info("bnjAwardproc closed!")
			return
		}
		msg, ok := <-s.bnjAwardSub.Messages()
		if !ok {
			log.Info("bnjAwardproc databus exit!")
			return
		}
		msg.Commit()
		award := new(bnjmdl.AwardAction)
		if err := json.Unmarshal(msg.Value, &award); err != nil || award == nil || award.Mid == 0 {
			log.Error("bnjAwardproc json.Unmarshal(%s) error(%+v) award(%+v)", msg.Value, err, award)
			continue
		}
		switch award.Type {
		case _awardTypePugvCoupon, _awardTypeComicCoupon, _awardTypeMallCoupon, _awardTypePendant, _awardFinalType:
			select {
			case s.bnjAwardch[award.Type] <- award:
				// 上报长度
				prom.BusinessInfoCount.State(fmt.Sprintf("bnj-award-chan-%d", award.Type), int64(len(s.bnjAwardch[award.Type])))
			default:
				log.Warn("bnjAwardproc chan(%d) full", award.Type)
			}
		default:
			log.Warn("bnjAwardproc award(%+v) conf error", award)
			continue
		}
		log.Warn("bnjAwardproc award(%+v) success", award)
	}
}

func (s *Service) bnjPugvAwardproc() {
	defer s.waiter.Done()
	ticker := time.NewTicker(time.Millisecond * 5)
	c := context.Background()
	for {
		select {
		case m, ok := <-s.bnjAwardch[_awardTypePugvCoupon]:
			prom.BusinessInfoCount.State(fmt.Sprintf("bnj-award-chan-%d", _awardTypePugvCoupon), int64(len(s.bnjAwardch[_awardTypePugvCoupon])))
			if !ok || s.closed {
				log.Error("bnjPugvAwardproc s.bnjMallAwardproc closed")
				return
			}
			if m == nil {
				continue
			}
			if m.Mirror != "" {
				metadata.NewContext(c, map[string]interface{}{metadata.Mirror: m.Mirror})
			}
			_, err := s.cheeseClient.AsynReceiveCoupon(c, &cheeseapi.AsynReceiveCouponReq{Mid: m.Mid, BatchToken: m.SourceID, SendVc: 1})
			if err != nil {
				// 用户已领取，忽略错误，处理成用户已领取
				if xecode.EqualError(xecode.Int(6009018), err) {
					log.Warn("bnjPugvAwardproc s.cheeseClient.AsynReceiveCoupon mid(%d) has reward", m.Mid)
				} else {
					log.Error("bnjPugvAwardproc s.cheeseClient.AsynReceiveCoupon award(%+v) error(%v)", m, err)
					continue
				}
			}
			err = s.doTask(context.Background(), m.TaskID, m.Mid)
			if err != nil {
				log.Error("bnjPugvAwardproc s.dao.DoTask taskID(%d) mid(%d) error(%v)", m.TaskID, m.Mid, err)
				continue
			}
		case <-ticker.C:
		}
	}
}

func (s *Service) bnjComicAwardproc() {
	defer s.waiter.Done()
	ticker := time.NewTicker(time.Millisecond * 5)
	c := context.Background()
	for {
		select {
		case m, ok := <-s.bnjAwardch[_awardTypeComicCoupon]:
			prom.BusinessInfoCount.State(fmt.Sprintf("bnj-award-chan-%d", _awardTypeComicCoupon), int64(len(s.bnjAwardch[_awardTypeComicCoupon])))
			if !ok || s.closed {
				log.Error("bnjComicAwardproc bnjAwardch s.bnjAwardch closed")
				return
			}
			if m == nil {
				continue
			}
			err := s.bnj.ComicCoupon(c, m.Mid, _comicNum)
			if err != nil {
				log.Error("bnjComicAwardproc s.bnj.ComicCoupon award(%+v) error(%v)", m, err)
				continue
			}
			err = s.doTask(context.Background(), m.TaskID, m.Mid)
			if err != nil {
				log.Error("bnjComicAwardproc s.dao.DoTask taskID(%d) mid(%d) error(%v)", m.TaskID, m.Mid, err)
				continue
			}
		case <-ticker.C:
		}
	}
}

func (s *Service) bnjMallAwardproc() {
	defer s.waiter.Done()
	// qps 200
	ticker := time.NewTicker(time.Millisecond * 5)
	for {
		select {
		case m, ok := <-s.bnjAwardch[_awardTypeMallCoupon]:
			prom.BusinessInfoCount.State(fmt.Sprintf("bnj-award-chan-%d", _awardTypeMallCoupon), int64(len(s.bnjAwardch[_awardTypeMallCoupon])))
			if !ok || s.closed {
				log.Error("bnjMallAwardproc bnjAwardch closed")
				return
			}
			if m == nil {
				continue
			}
			err := s.bnj.MallCoupon(context.Background(), m.Mid, _mallCouponSourceID, m.SourceID, _mallCouponActivityId)
			if err != nil {
				log.Error("bnjMallAwardproc s.bnj.MallCoupon award(%+v) error(%v)", m, err)
				continue
			}
			err = s.doTask(context.Background(), m.TaskID, m.Mid)
			if err != nil {
				log.Error("bnjMallAwardproc s.dao.DoTask taskID(%d) mid(%d) error(%v)", m.TaskID, m.Mid, err)
				continue
			}
		case <-ticker.C:
		}
	}
}

func (s *Service) bnjPendantAwardproc() {
	defer s.waiter.Done()
	// qps 200
	ticker := time.NewTicker(time.Millisecond * 5)
	c := context.Background()
	for {
		select {
		case m, ok := <-s.bnjAwardch[_awardTypePendant]:
			prom.BusinessInfoCount.State(fmt.Sprintf("bnj-award-chan-%d", _awardTypePendant), int64(len(s.bnjAwardch[_awardTypePendant])))
			if !ok || s.closed {
				log.Error("bnjPendantAwardproc bnjAwardch closed")
				return
			}
			if m == nil {
				continue
			}
			pendantID, err := strconv.ParseInt(m.SourceID, 10, 64)
			if err != nil || m.SourceExpire == 0 {
				log.Warn("bnjPendantAwardproc award(%+v) conf error", m)
				continue
			}
			_, err = s.suitClient.GrantByMids(c, &suitapi.GrantByMidsReq{Mids: []int64{m.Mid}, Pid: pendantID, Expire: m.SourceExpire})
			if err != nil {
				log.Error("bnjPendantAwardproc s.suitClient.GrantByMids award(%+v) error(%v)", m, err)
				continue
			}
			err = s.doTask(context.Background(), m.TaskID, m.Mid)
			if err != nil {
				log.Error("bnjPendantAwardproc s.dao.DoTask taskID(%d) mid(%d) error(%v)", m.TaskID, m.Mid, err)
				continue
			}
		case <-ticker.C:
		}
	}
}

func (s *Service) bnjLiveAwardproc() {
	defer s.waiter.Done()
	// qps 200
	ticker := time.NewTicker(time.Millisecond * 5)
	c := context.Background()
	for {
		select {
		case m, ok := <-s.bnjAwardch[_awardFinalType]:
			prom.BusinessInfoCount.State(fmt.Sprintf("bnj-award-chan-%d", _awardFinalType), int64(len(s.bnjAwardch[_awardFinalType])))
			if !ok || s.closed {
				log.Error("bnjLiveAwardproc s.bnjAwardch closed")
				return
			}
			if m == nil {
				continue
			}
			err := s.bnj.SendLiveItem(c, m.Mid)
			if err != nil {
				log.Error("bnjLiveAwardproc s.bnj.SendLiveItem award(%+v) error(%v)", m, err)
				continue
			}
			err = s.doTask(context.Background(), m.TaskID, m.Mid)
			if err != nil {
				log.Error("bnjLiveAwardproc s.dao.DoTask taskID(%d) mid(%d) error(%v)", m.TaskID, m.Mid, err)
				continue
			}
		case <-ticker.C:
		}
	}
}

func (s *Service) bnjReserveproc() {
	total, err := s.dao.RawSubjectStat(context.Background(), s.c.Bnj2020.Sid)
	if err != nil {
		log.Error("bnjReserveproc s.bnj.RawSubjectStat total(%d) sid(%d) error(%v)", total, s.c.Bnj2020.Sid, err)
		return
	}
	if total <= s.bnjReservedCount {
		return
	}
	atomic.StoreInt64(&s.bnjReservedCount, total)
	pushMsg := &bnjmdl.PushReserveBnj20{Type: _pushTypeReserve, ReservedCount: total}
	pushStr, err := json.Marshal(pushMsg)
	if err != nil {
		log.Error("bnjReserveproc json.Marshal(%+v) error(%v)", pushMsg, err)
		return
	}
	if err := s.bnj.PushAll(context.Background(), string(pushStr)); err != nil {
		log.Error("bnjReserveproc s.bnj.PushAll error(%v)", err)
		return
	}
	log.Info("bnjReserveproc pushMsg(%+v)", pushMsg)
}

func (s *Service) dealBnjMsg(changeNum int64, pushAction *bnjmdl.Action) {
	preValue := atomic.LoadInt64(&s.bnjValue)
	if preValue >= s.c.Bnj2020.MaxValue {
		log.Warn("dealBnjMsg preValue(%d) max", preValue)
		return
	}
	afValue := preValue + changeNum
	// 美味值不能比0低
	if afValue < 0 {
		afValue = 0
	}
	atomic.StoreInt64(&s.bnjValue, afValue)
	pushMsg := &bnjmdl.PushHotpotBnj20{Type: _pushTypeHotpot, Value: afValue}
	if pushAction != nil {
		pushMsg.Msg = pushAction.Message
		name := func(c context.Context, mid int64) string {
			info, err := s.accClient.Info3(c, &accapi.MidReq{Mid: mid})
			if err != nil || info == nil {
				log.Error("dealBnjMsg s.accClient.Info3(%d) error(%v)", mid, err)
				return ""
			}
			var name []rune
			runes := []rune(info.Info.Name)
			nameLen := len(runes)
			if nameLen == 2 {
				name = append(runes[0:1], []rune("*")...)
				return string(name)
			}
			if nameLen > 2 {
				var markNum int
				for i, v := range runes {
					if i == 0 {
						name = append(name, v)
					} else if i == nameLen-1 {
						name = append(name, runes[nameLen-1:]...)
					} else {
						if markNum >= 1 {
							continue
						}
						name = append(name, []rune("*")...)
						markNum++
					}
				}
				return string(name)
			}
			name = runes
			return string(name)
		}(context.Background(), pushAction.Mid)
		pushMsg.Msg = name + " " + pushMsg.Msg
	}
	if afValue >= s.c.Bnj2020.MaxValue {
		afValue = s.c.Bnj2020.MaxValue
		pushMsg = &bnjmdl.PushHotpotBnj20{Type: _pushTypeHotpot, Value: afValue, TimelinePic: s.c.Bnj2020.TimelinePic, H5TimelinePic: s.c.Bnj2020.H5TimelinePic}
	}
	atomic.StoreInt64(&s.bnjValue, afValue)
	pushStr, err := json.Marshal(pushMsg)
	if err != nil {
		log.Error("dealBnjMsg json.Marshal(%+v) error(%v)", pushMsg, err)
		return
	}
	if err := s.bnj.PushAll(context.Background(), string(pushStr)); err != nil {
		log.Error("dealBnjMsg s.bnj.PushAll error(%v)", err)
		return
	}
	log.Info("dealBnjMsg action(%+v) msg(%s)", pushAction, string(pushStr))
}

func (s *Service) cronInformationMessage() {
	if s.c.Bnj2020.Sid == 0 || s.c.Bnj2020.MidLimit == 0 {
		log.Error("cronInformationMessage conf error")
		return
	}
	var (
		minID   int64
		content string
		c       = context.Background()
	)
	for _, v := range s.c.Bnj2020.Message {
		if v.Start.Unix() >= time.Now().Unix() {
			content = v.Content
			break
		}
	}
	if content == "" {
		log.Error("cronInformationMessage message conf error")
		return
	}
	cont := &struct {
		Content string `json:"content"`
	}{Content: content}
	sendContent, err := json.Marshal(cont)
	if err != nil {
		log.Error("cronInformationMessage json.Marshal cont(%+v) error(%v)", cont, err)
		return
	}
	// 获取key
	msgKey, err := func(c context.Context, sender int64, content string) (msgKey int64, err error) {
		for i := 0; i < _retryTimes; i++ {
			if msgKey, err = s.bnj.MessageKey(c, sender, content); err == nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		return
	}(c, s.c.Bnj2020.Sender, string(sendContent))
	if err != nil {
		log.Error("cronInformationMessage get message key error(%v)", err)
		return
	}
	for {
		time.Sleep(100 * time.Millisecond)
		list, err := func(c context.Context, sid, minID, limit int64) (list []*like.Reserve, err error) {
			for i := 0; i < _retryTimes; i++ {
				if list, err = s.dao.RawReserveList(c, sid, minID, limit); err == nil {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			return
		}(c, s.c.Bnj2020.Sid, minID, s.c.Bnj2020.MidLimit)
		if err != nil {
			log.Error("cronInformationMessage s.dao.retryReserveList(%d,%d,%d) error(%v)", s.c.Bnj2020.Sid, minID, s.c.Bnj2020.MidLimit, err)
			continue
		}
		if len(list) == 0 {
			log.Info("cronInformationMessage finish")
			break
		}
		var mids []int64
		for i, v := range list {
			if v.Mid > 0 {
				mids = append(mids, v.Mid)
			}
			if i == len(list)-1 {
				minID = v.ID
			}
		}
		if len(mids) == 0 {
			continue
		}
		midList := splitInt64(mids, 50)
		for _, v := range midList {
			if err = s.bnj.AsyncSendNormalMessage(c, msgKey, s.c.Bnj2020.Sender, v); err != nil {
				time.Sleep(100 * time.Millisecond)
				log.Error("Failed to call AsyncSendNormalMessage: msg_key: %d, to: %+v: %+v", msgKey, v, err)
				continue
			}
			log.Info("Succeed to send bnj message to: %+v", v)
		}
	}
}

func splitInt64(buf []int64, limit int) [][]int64 {
	var chunk []int64
	chunks := make([][]int64, 0, len(buf)/limit+1)
	for len(buf) >= limit {
		chunk, buf = buf[:limit], buf[limit:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf)
	}
	return chunks
}

func (s *Service) doTask(c context.Context, taskID, mid int64) (err error) {
	for i := 0; i < _retryTimes; i++ {
		if err = s.dao.DoTask(c, taskID, mid); err == nil {
			break
		}
		if xecode.EqualError(ecode.ActivityTaskHasFinish, err) {
			err = nil
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return
}

// Bnj2020MessageSend is
func (s *Service) Bnj2020MessageSend(ctx context.Context, mids []int64, send int64) (map[string]interface{}, error) {
	if s.c.Bnj2020.Sid == 0 || s.c.Bnj2020.MidLimit == 0 {
		return nil, errors.New("cronInformationMessage conf error")
	}
	content := ""
	for _, v := range s.c.Bnj2020.Message {
		if v.Start.Unix() >= time.Now().Unix() {
			content = v.Content
			break
		}
	}
	if content == "" {
		return nil, errors.New("cronInformationMessage message conf error")
	}
	cont := &struct {
		Content string `json:"content"`
	}{Content: content}
	sendContent, err := json.Marshal(cont)
	if err != nil {
		return nil, errors.Errorf("cronInformationMessage json.Marshal cont(%+v) error(%+v)", cont, err)
	}
	msgKey, err := s.bnj.MessageKey(ctx, s.c.Bnj2020.Sender, string(sendContent))
	if err != nil {
		return nil, err
	}
	if send <= 0 {
		return map[string]interface{}{
			"msg_key":      msgKey,
			"content":      content,
			"send_content": string(sendContent),
			"mids":         mids,
		}, nil
	}
	return nil, s.bnj.AsyncSendNormalMessage(ctx, msgKey, s.c.Bnj2020.Sender, mids)
}
