package service

import (
	"encoding/json"

	"go-common/library/log"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

const (
	_nodeinfoEvent = "player.ugc-video-detail.worldline.0.options.player"
)

func (s *Service) infoc(i interface{}) {
	switch v := i.(type) {
	case *model.InfocNode:
		msg, err := json.Marshal(v.NodeExtended)
		if err != nil {
			log.Error("InfocNode JsonMarshal Aid %d Err %v", v.AID, err)
			return
		}
		log.Info("infoc data %s, %d, %d, %d, %d, %s, %d, %s, %s, %s", _nodeinfoEvent,
			v.FromNID, v.ToNID, v.AID, v.MID, string(msg), v.Build, v.Channel, v.MobiApp, v.Platform)
		if err = s.infocNode.Info(_nodeinfoEvent, v.FromNID, v.ToNID,
			v.AID, v.MID, string(msg), v.Build, v.Channel, v.MobiApp, v.Platform); err != nil {
			log.Error("InfocNode Aid %d Err %v", v.AID, err)
			return
		}
	case *model.InfocMark:
		log.Info("infoc data %d, %d, %d, %d, %d",
			v.AID, v.MID, v.GraphVersion, v.Mark, v.LogTime)
		if err := s.infocMark.Info(v.AID, v.MID, v.GraphVersion, v.Mark, v.LogTime); err != nil {
			log.Error("InfocMark Aid %d Err %v", v.AID, err)
			return
		}
	default:
		log.Warn("infocproc can't process the type")
	}
}

func (s *Service) infocAction(params *model.NodeInfoParam, mid, resquestId, fromID int64, otype int32) {
	if params.Portal == 0 { // 非回溯操作需要上报
		infocMsg := &model.InfocNode{ // infoc 消息
			ToNID: resquestId,
			MID:   mid,
			AID:   params.AID,
			NodeExtended: model.NodeExtended{
				Type:   otype,
				Delay:  params.Delay,
				Screen: params.Screen,
			},
			FromNID:  fromID,
			Build:    params.Build,
			Channel:  params.Channel,
			MobiApp:  params.MobiApp,
			Platform: params.Platform,
		}
		s.infoc(infocMsg)
	}

}
