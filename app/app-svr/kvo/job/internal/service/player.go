package service

import (
	"context"
	"encoding/json"
	"time"

	xecode "go-gateway/app/app-svr/kvo/ecode"
	v1 "go-gateway/app/app-svr/kvo/interface/api"
	"go-gateway/app/app-svr/kvo/job/internal/model"

	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/queue/databus/databusutil"
)

func (s *Service) initPlayer(cfg *databus.Config, cfgutil *databusutil.Config) {
	player := databusutil.NewGroup(cfgutil, databus.New(cfg).Messages())
	player.New = s.newMessagePlayer
	player.Split = s.splitPlayer
	player.Do = s.doPlayer
	player.Start()
	s.player = player
}

func (s *Service) newMessagePlayer(msg *databus.Message) (res interface{}, err error) {
	action := new(model.Action)
	if err = json.Unmarshal(msg.Value, action); err != nil {
		log.Error("s.newMessagePlayer() json.Unmarshal() databus message(%+v) error(%v)", msg, err)
		return
	}
	playerConfig := new(model.PlayerConfig)
	if err = json.Unmarshal(action.Data, playerConfig); err != nil {
		log.Error("s.newMessagePlayer() json.Unmarshal() databus action(%+v) error(%v)", action.Data, err)
		return
	}
	playerConfig.Action = action.Action
	log.Info("s.newMessagePlayer(%+v)", playerConfig)
	res = playerConfig
	return
}

func (s *Service) splitPlayer(msg *databus.Message, data interface{}) int {
	var ret int
	switch v := data.(type) {
	case *model.PlayerConfig:
		return int(v.Mid)
	default:
		log.Error("s.splitPlayer() get data filed, message(%+v) data(%+v)", msg, data)
	}
	return ret
}

func (s *Service) doPlayer(msgs []interface{}) {
	var (
		ok           bool
		ctx          = context.TODO()
		playerConfig *model.PlayerConfig
		i            int
	)
	mergeMsgs := make(map[int64]*model.PlayerConfig, len(msgs))
	for _, msg := range msgs {
		if msg == nil {
			continue
		}
		switch v := msg.(type) {
		case *model.PlayerConfig:
			if playerConfig, ok = mergeMsgs[v.Mid]; !ok {
				mergeMsgs[v.Mid] = v
			} else {
				playerConfig.Platform = v.Platform
				playerConfig.Body.Merge(v.Body)
			}
		default:
			log.Error("s.doPlayer() get data filed, message(%+v)", msg)
		}
	}
	for _, msg := range mergeMsgs {
		switch msg.Action {
		case v1.DmPlayerConfig:
			log.Info("s.addPlayerConfig start mid(%d) body(%+v)", msg.Mid, msg.Body)
			if err := s.addPlayerConfig(ctx, msg.Mid, msg.Body, msg.Platform); err != nil {
				log.Error("s.addPlayerConfig(mid:%d,body:%+v) err(%v)", msg.Mid, msg.Body, err)
			}
			i += 1
			if i == s.cfg.DoInterval {
				log.Warn("s.addPlayerConfig sleep %d", i)
				time.Sleep(time.Second)
				i = 0
			}
		}
	}
	log.Info("doPlayer len(%d) merge(%d)", len(msgs), len(mergeMsgs))
	return
}

func (s *Service) addPlayerConfig(ctx context.Context, mid int64, req *v1.DmPlayerConfigReq, platform string) (err error) {
	var (
		moduleKeyID    = v1.VerifyModuleKey(v1.DmPlayerConfig)
		uc             *model.UserConf
		player         *v1.DanmuPlayerConfig
		modify         bool
		doc, playerArr json.RawMessage
		checkSum       int64
		now            = time.Now()
		tx             *sql.Tx
		addDocCache    bool
	)
	if uc, err = s.dao.UserConf(ctx, mid, moduleKeyID); err != nil {
		return
	}
	player = &v1.DanmuPlayerConfig{}
	player.Default()
	if uc != nil {
		if doc, err = s.dao.Document(ctx, uc.CheckSum); err != nil {
			log.Error("s.dao.Document(mid:%d,uc:%+v) err(%v)", mid, uc, err)
			return
		}
		if err = json.Unmarshal([]byte(doc), &player); err != nil {
			log.Error("addPlayerConfig json.Unmarshal(document:%s) mid(%d) err(%v)", doc, mid, err)
			return
		}
	}
	modify = player.Change(req)
	if !modify {
		err = xecode.KvoNotModified
		log.Warn("s.addPlayerConfig eq (mid:%d, player:%+v,req:%+v)", mid, player, req)
		return
	}
	if playerArr, checkSum, err = model.Result(player); err != nil {
		log.Error("s.addPlayerConfig Result(mid:%d, player:%+v) error(%v)", mid, player, err)
		return
	}
	if uc != nil {
		if uc.CheckSum == checkSum {
			if string(playerArr) != string(doc) {
				log.Error("hash conflict mid(%d) doc(%s) newdoc(%s)", mid, string(doc), string(playerArr))
				err = xecode.KvoHashConflict
				return
			}
			log.Warn("updateUcCache (mid:%d,moduleKeyID:%d,checkSum:%d)", mid, moduleKeyID, checkSum)
			s.dao.SetUserConf(ctx, uc)
			err = ecode.NotModified
			return
		}
	}
	tx, err = s.dao.BeginTx(ctx)
	if err != nil {
		log.Error("s.da.BeginTx mid:%d err:%v", mid, err)
		return
	}
	if err = s.dao.TxUpUserConf(ctx, tx, mid, moduleKeyID, checkSum, now); err != nil {
		log.Error("s.da.TxUpUserConf(%v,%v,%v) error(%v)", mid, moduleKeyID, checkSum, err)
		tx.Rollback()
		return
	}
	if _, err = s.dao.Document(ctx, checkSum); err != nil {
		if ecode.Cause(err) != ecode.NothingFound {
			tx.Rollback()
			return
		}
		if err = s.dao.TxUpDocuement(ctx, tx, checkSum, string(playerArr), now); err != nil {
			log.Error("s.da.TxUpDocuement(%v,%v,%v) error(%v)", mid, moduleKeyID, checkSum, err)
			tx.Rollback()
			return
		}
		addDocCache = true
	}
	if err = tx.Commit(); err != nil {
		log.Error("mid:%d key:%d tx.Commit(), error(%v)", mid, moduleKeyID, err)
		return
	}
	log.Warn("updateUcCache (mid:%d,moduleKeyID:%d,checkSum:%d)", mid, moduleKeyID, checkSum)
	s.dao.SetUserConf(ctx, &model.UserConf{
		ModuleKey: moduleKeyID,
		Mid:       mid,
		CheckSum:  checkSum,
		Timestamp: now.Unix(),
	})
	if addDocCache {
		log.Warn("updateDocCache (mid:%d,moduleKeyID:%d,checkSum:%d,player:%s)", mid, moduleKeyID, checkSum, string(playerArr))
		s.dao.SetDocument(ctx, checkSum, playerArr)
	}
	//_ = s.addUserDocTaiShan(ctx, mid, "", moduleKeyID, player)
	return
}
