package service

import (
	"context"
	"encoding/json"
	"time"

	xecode "go-gateway/app/app-svr/kvo/ecode"
	v1 "go-gateway/app/app-svr/kvo/interface/api"
	"go-gateway/app/app-svr/kvo/job/internal/model"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/queue/databus/databusutil"
)

const initSize = 1024

func (s *Service) initBuvid(cfg *databus.Config, cfgutil *databusutil.Config) {
	Buvid := databusutil.NewGroup(cfgutil, databus.New(cfg).Messages())
	Buvid.New = s.newMessageBuvid
	Buvid.Split = s.splitBuvid
	Buvid.Do = s.doBuvid
	Buvid.Start()
	s.buvid = Buvid
}

func (s *Service) newMessageBuvid(msg *databus.Message) (res interface{}, err error) {
	var moduleID int
	action := new(model.Action)
	if err = json.Unmarshal(msg.Value, action); err != nil {
		log.Error("s.newMessagePlayer() json.Unmarshal() databus message(%+v) error(%v)", msg, err)
		return
	}
	if moduleID = v1.VerifyModuleKey(action.Action); moduleID == 0 {
		log.Error("s.VerifyModuleKey (action:%s)", action.Action)
		return
	}
	playerConfig := new(model.CfgMessage)
	playerConfig.Body = v1.NewConfigModify(moduleID)
	if err = json.Unmarshal(action.Data, playerConfig); err != nil {
		log.Error("s.newMessagePlayer() json.Unmarshal() databus action(%s) error(%v)", action.Data, err)
		return
	}
	playerConfig.Action = action.Action
	log.Info("s.newMessagePlayer(%+v)", playerConfig)
	res = playerConfig
	return
}

func (s *Service) splitBuvid(msg *databus.Message, data interface{}) int {
	var ret int
	switch v := data.(type) {
	case *model.CfgMessage:
		return v.Buvid.Crc63()
	default:
		log.Error("s.splitPlayer() get data filed, message(%+v) data(%+v)", msg, data)
	}
	return ret
}

func (s *Service) doBuvid(msgs []interface{}) {
	var (
		ctx       = context.TODO()
		i         int
		MergeMsgs = make(map[string]*model.MergeCfgMessage, initSize)
	)
	for _, msg := range msgs {
		if msg == nil {
			continue
		}
		switch v := msg.(type) {
		case *model.CfgMessage:
			moduleId := v1.VerifyModuleKey(v.Action)
			if playerConfig, ok := MergeMsgs[string(v.Buvid)]; !ok {
				MergeMsgs[string(v.Buvid)] = &model.MergeCfgMessage{
					Mid: v.Mid,
					Bodys: map[int]v1.ConfigModify{
						moduleId: v.Body,
					},
					Platform: v.Platform,
					Buvid:    v.Buvid,
				}
			} else {
				playerConfig.Platform = v.Platform
				playerConfig.Merge(moduleId, v.Body)
			}
		default:
			log.Error("s.doPlayerBuvid() get data filed, message(%+v)", msg)
		}
	}
	for _, msg := range MergeMsgs {
		log.Info("s.doPlayerBuvid start buvid(%v) body(%+v)", msg.Buvid, msg.Bodys)
		if err := s.addConfig(ctx, string(msg.Buvid), msg.Bodys, msg.Platform); err != nil {
			log.Error("s.doPlayerBuvid(buvid:%v,body:%+v) err(%v)", msg.Buvid, msg.Bodys, err)
		}
		i += 1
		if i == s.cfg.DoInterval {
			log.Warn("s.doPlayerBuvid sleep %d", i)
			time.Sleep(time.Second)
			i = 0
		}
	}
	log.Info("doPlayer len(%d) merge(%d)", len(msgs), len(MergeMsgs))
	return
}

func (s *Service) addConfig(ctx context.Context, buvid string, reqs map[int]v1.ConfigModify, platform string) (err error) {
	var (
		doc         json.RawMessage
		playerBytes []byte
	)
	for moduleKeyID, req := range reqs {
		if doc, err = s.userDoc(ctx, 0, buvid, moduleKeyID); err != nil {
			return
		}
		player := v1.NewConfig(moduleKeyID, nil)
		player.Default()
		if doc != nil {
			if err = json.Unmarshal([]byte(doc), &player); err != nil {
				log.Error("s.addDmConfig json.Unmarshal(document:%s) buvid(%s) err(%v)", doc, buvid, err)
				return
			}
		}
		if modify := player.Change(req); !modify {
			err = xecode.KvoNotModified
			log.Warn("s.addDmConfig eq (buvid:%s, key:%d, player:%+v,req:%+v)", buvid, moduleKeyID, player, req)
			return
		}
		if playerBytes, err = json.Marshal(player); err != nil {
			log.Error("s.addDmConfig json.Marshal(player:%+v) buvid(%s) module(%d) err(%v)", player, buvid, moduleKeyID, err)
			return
		}
		log.Warn("updateDmBvCache (buvid:%s, moduleKeyID:%d) player(%s)", buvid, moduleKeyID, string(playerBytes))
		_ = s.dao.SetUserDocRds(ctx, 0, buvid, moduleKeyID, playerBytes)
		if err = s.addUserDocTaiShan(ctx, 0, buvid, moduleKeyID, player); err != nil {
			log.Error("s.addUserDocTaiShan(buvid:%s, key:%d, player:%+v, req:%+v) err(%v)", buvid, moduleKeyID, player, req, err)
			return
		}
	}
	return
}
