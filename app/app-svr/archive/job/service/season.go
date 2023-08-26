package service

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"go-common/library/railgun"
	jobmdl "go-gateway/app/app-svr/archive/job/model/databus"
)

func (s *Service) SeasonNotifyUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	m := &jobmdl.SeasonWithArchive{}
	if err := json.Unmarshal(msg.Payload(), m); err != nil {
		log.Error("json.Unmarshal(%+v) error(%+v)", string(msg.Payload()), err)
		return nil, err
	}
	return &railgun.SingleUnpackMsg{
		Group: m.SeasonID,
		Item:  m,
	}, nil
}

func (s *Service) SeasonNotifyDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	if item == nil {
		return railgun.MsgPolicyIgnore
	}
	m := item.(*jobmdl.SeasonWithArchive)
	if m == nil {
		return railgun.MsgPolicyIgnore
	}
	if m.SeasonID <= 0 || len(m.Aids) == 0 {
		log.Error("seasonNotify wrong messages(%+v)", m)
		return railgun.MsgPolicyIgnore
	}
	var act int
	s.Prom.Incr(m.Route)
	switch m.Route {
	case jobmdl.SeasonRouteForUpdate:
		act = _arcActionUpSid
	case jobmdl.SeasonRouteForRemove:
		act = _arcActionRmSid
	default:
		log.Error("get wrong route(%s) messages(%+v)", m.Route, m)
		return railgun.MsgPolicyIgnore
	}
	for _, aid := range m.Aids {
		s.arcUpdate(aid, act, m.SeasonID)
	}
	return railgun.MsgPolicyNormal
}
