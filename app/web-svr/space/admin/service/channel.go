package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/space/admin/model"
)

const (
	_channelName    = "name"
	_channelIntro   = "intro"
	_channelNameDft = "未命名频道"
)

func (s *Service) Channel(_ context.Context, mid int64) (list []*model.Channel, err error) {
	data := &model.Channel{Mid: mid}
	if err = s.dao.DB.Table(data.TableName()).Where("mid=?", mid).Find(&list).Error; err != nil {
		log.Error("Masterpiece mid:%d error (%v)", mid, err)
		return
	}
	if len(list) == 0 {
		err = ecode.NothingFound
	}
	return
}

func (s *Service) ChannelClear(_ context.Context, mid, channelID int64, field string) (oldData *model.Channel, err error) {
	oldData = &model.Channel{Mid: mid}
	if err = s.dao.DB.Table(oldData.TableName()).Where("mid=?", mid).Where("id=?", channelID).First(&oldData).Error; err != nil {
		log.Error("ChannelClear mid:%d channelID:%d error (%v)", mid, channelID, err)
		return
	}
	afValue := ""
	if field == _channelName {
		afValue = _channelNameDft
	}
	if err = s.dao.DB.Table(oldData.TableName()).Where("mid=?", mid).Where("id=?", channelID).Update(field, afValue).Error; err != nil {
		log.Error("ChannelClear mid:%d channelID:%d field:%s error(%v)", mid, channelID, field, err)
	}
	return
}
