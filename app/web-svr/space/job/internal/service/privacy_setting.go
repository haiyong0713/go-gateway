package service

import (
	"context"
	"encoding/json"
	"strings"

	// "go-common/library/railgun"
	"go-gateway/app/web-svr/space/job/internal/model"

	"go-common/library/railgun.v2"
	"go-common/library/railgun.v2/message"
	"go-common/library/railgun.v2/processor/single"
)

const (
	_memberPrivacy = "member_privacy"
)

func (s *Service) initSpaceRailGun() {
	processor := single.New(s.spaceUnpack, s.spaceDo)
	// 指定railgun平台的处理器ID
	consumer, err := railgun.NewConsumer(s.ac.RailgunV2.SpaceBinlog, processor)
	if err != nil {
		panic(err)
	}
	s.spaceConsumer = consumer
}

func (s *Service) spaceUnpack(msg message.Message) (m *single.UnpackMessage, err error) {
	var v *model.MemberPrivacyMsg
	if err := json.Unmarshal(msg.Payload(), &v); err != nil {
		return nil, err
	}
	if v == nil || !strings.HasPrefix(v.Table, _memberPrivacy) {
		return nil, nil
	}
	item := v.New
	if item == nil {
		item = v.Old
	}
	if item == nil {
		return nil, nil
	}
	return &single.UnpackMessage{
		// 传入用于计算分组的id 保证同一id排队有序处理
		Group: item.Mid,
		// 传递给do方法的对象 任意类型
		Item: item,
	}, nil
}

func (s *Service) spaceDo(ctx context.Context, item interface{}, extra *single.Extra) message.Policy {
	data := item.(*model.MemberPrivacy)
	mid := data.Mid
	if err := s.dao.DelCachePrivacySetting(ctx, mid); err != nil {
		return message.Retry
	}
	return message.Success
}
