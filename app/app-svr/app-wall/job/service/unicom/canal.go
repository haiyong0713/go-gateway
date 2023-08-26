package unicom

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-wall/job/model/unicom"
)

const (
	_userBindTable    = "unicom_user_bind"
	_userPacksTable   = "unicom_user_packs"
	_unicomOrderTable = "unicom_order"
	_mobileOrderTable = "mobile_order"
	_usermobInfoTable = "unicom_usermob_info"
)

func (s *Service) initCanalRailGun(cfg *railgun.DatabusV1Config, pcfg *railgun.SingleConfig) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(pcfg, s.canalUnpack, s.canalDo)
	g := railgun.NewRailGun("订阅binlog", nil, inputer, processor)
	s.canalRailGun = g
	g.Start()
}

func (s *Service) canalUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	var v *unicom.CanalMsg
	if err := json.Unmarshal(msg.Payload(), &v); err != nil {
		return nil, err
	}
	if v == nil || v.New == nil {
		return nil, nil
	}
	return &railgun.SingleUnpackMsg{
		Group: time.Now().Unix(),
		Item:  v,
	}, nil
}

func (s *Service) canalDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	res := item.(*unicom.CanalMsg)
	switch res.Table {
	case _userBindTable:
		var data *unicom.UnicomUserBind
		if err := json.Unmarshal(res.New, &data); err != nil {
			log.Error("%+v", err)
			return railgun.MsgPolicyIgnore
		}
		if err := s.updateUserBind(ctx, data.Mid); err != nil {
			log.Error("%+v", err)
			return railgun.MsgPolicyAttempts
		}
		log.Info("canal consumer user bind success,mid:%v", data.Mid)
	case _userPacksTable:
		var data *unicom.UnicomUserPacks
		if err := json.Unmarshal(res.New, &data); err != nil {
			log.Error("%+v", err)
			return railgun.MsgPolicyIgnore
		}
		if err := s.dao.DeleteUserPackCache(ctx, data.ID); err != nil {
			log.Error("%+v", err)
			return railgun.MsgPolicyAttempts
		}
		log.Info("canal consumer user pack success,id:%v", data.ID)
	case _unicomOrderTable:
		var data *unicom.UnicomOrder
		if err := json.Unmarshal(res.New, &data); err != nil {
			log.Error("%+v", err)
			return railgun.MsgPolicyIgnore
		}
		if err := s.dao.DeleteUnicomCache(ctx, data.Usermob); err != nil {
			log.Error("%+v", err)
			return railgun.MsgPolicyAttempts
		}
		log.Info("canal consumer unicom order success,id:%v", data.ID)
	case _mobileOrderTable:
		var data *unicom.MobileOrder
		if err := json.Unmarshal(res.New, &data); err != nil {
			log.Error("%+v", err)
			return railgun.MsgPolicyIgnore
		}
		if err := s.dao.DeleteMobileCache(ctx, data.Userpseudocode); err != nil {
			log.Error("%+v", err)
			return railgun.MsgPolicyAttempts
		}
		log.Info("canal consumer mobile order success,id:%v", data.ID)
	case _usermobInfoTable:
		var data *unicom.UnicomUsermobInfo
		if err := json.Unmarshal(res.New, &data); err != nil {
			log.Error("%+v", err)
			return railgun.MsgPolicyIgnore
		}
		if err := s.dao.DelUsermobInfoCache(ctx, data.FakeID, data.Period); err != nil {
			log.Error("canal consumer delete usermob info cache error:%+v", err)
			return railgun.MsgPolicyAttempts
		}
		log.Info("canal consumer unicom usermob info success,id:%v", data.ID)
	}
	return railgun.MsgPolicyNormal
}
