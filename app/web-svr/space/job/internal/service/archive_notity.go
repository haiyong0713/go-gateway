package service

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	"go-common/library/railgun.v2"
	"go-common/library/railgun.v2/message"
	"go-common/library/railgun.v2/processor/single"
	"go-gateway/app/web-svr/space/job/internal/model"

	archiveapi "git.bilibili.co/bapis/bapis-go/archive/service"
)

const (
	_archive        = "archive"
	_update         = "update"
	_noticeTextType = 1
)

func (s *Service) initArcNotifyRailGun() {
	processor := single.New(s.arcNotifyUnpack, s.arcNotifyDo)
	// 指定railgun平台的处理器ID
	consumer, err := railgun.NewConsumer(s.ac.RailgunV2.ArchiveNotify, processor)
	if err != nil {
		panic(err)
	}
	s.arcConsumer = consumer
}

func (s *Service) arcNotifyUnpack(msg message.Message) (*single.UnpackMessage, error) {
	var v *model.ArcMsg
	if err := json.Unmarshal(msg.Payload(), &v); err != nil {
		return nil, err
	}
	if v == nil || v.Table != _archive || v.Action != _update {
		return nil, nil
	}
	return &single.UnpackMessage{
		Group: time.Now().Unix(),
		Item:  v,
	}, nil
}

func (s *Service) arcNotifyDo(ctx context.Context, item interface{}, extra *single.Extra) message.Policy {
	res := item.(*model.ArcMsg)
	// 下线
	if (res.Old != nil && (res.Old.State >= 0 || res.Old.State == -6)) &&
		(res.New != nil && res.New.State < 0 && res.Old.State != -6) && res.New.Aid > 0 && res.New.Mid > 0 {
		s.cancelTopPhotoArc(ctx, res.New.Mid)
		// 联合投稿查询staff
		if (res.New.Attribute>>archiveapi.AttrBitIsCooperation)&int32(1) == archiveapi.AttrYes {
			arcReply, err := s.archiveGRPC.Arc(ctx, &archiveapi.ArcRequest{Aid: res.New.Aid})
			if err != nil {
				log.Error("arcNotifyproc s.archiveGRPC.Arc aid:%d error:%v", res.New.Aid, err)
				return message.Ignore
			}
			for _, v := range arcReply.GetArc().GetStaffInfo() {
				if v.GetMid() > 0 {
					s.cancelTopPhotoArc(ctx, v.GetMid())
				}
			}
		}
	}
	time.Sleep(10 * time.Millisecond)
	return message.Success
}

func (s *Service) cancelTopPhotoArc(ctx context.Context, mid int64) {
	topArcData, err := s.dao.TopPhotoArc(ctx, mid)
	if err != nil {
		log.Error("cancelTopPhotoArc TopPhotoArc mid:%d error:%v", mid, err)
		return
	}
	if topArcData == nil || topArcData.Aid <= 0 {
		log.Warn("cancelTopPhotoArc TopPhotoArc mid:%d no arc", mid)
		return
	}
	// 下线
	if _, err = s.dao.TopPhotoArcCancel(ctx, mid); err != nil {
		log.Error("cancelTopPhotoArc s.dao.TopPhotoArcCancel mid:%d error:%v", mid, err)
		return
	}
	if err = retry(func() error {
		return s.dao.DelCacheTopPhotoArc(ctx, mid)
	}); err != nil {
		log.Error("cancelTopPhotoArc DelCacheTopPhotoArc mid:%d error:%v", mid, err)
	}
	// 发通知
	if err = retry(func() error {
		return s.dao.SendLetter(ctx, &model.LetterParam{
			RecverIDs: []uint64{uint64(mid)},
			SenderUID: s.ac.Msg.SenderUID,
			MsgType:   _noticeTextType,
			Content:   s.ac.Msg.NotifyMsg,
		})
	}); err != nil {
		log.Error("cancelTopPhotoArc s.dao.SendLetter recverID:%d senderUID:%d error:%v", mid, s.ac.Msg.SenderUID, err)
		return
	}
	log.Warn("cancelTopPhotoArc success mid:%d", mid)
}

func retry(callback func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		if err = callback(); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return err
}
