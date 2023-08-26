package service

import (
	"context"
	"encoding/json"
	"time"

	xecode "go-gateway/app/app-svr/kvo/ecode"
	v1 "go-gateway/app/app-svr/kvo/interface/api"
	"go-gateway/app/app-svr/kvo/interface/model"
	"go-gateway/app/app-svr/kvo/interface/model/module"

	"go-common/library/cache"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
)

// Document get document
func (s *Service) Document(c context.Context, mid int64, moduleKey string, timestamp int64, checkSum int64) (setting *module.Setting, err error) {
	var (
		uc          *model.UserConf
		rm          json.RawMessage
		moduleKeyID int
	)
	if moduleKeyID = v1.VerifyModuleKey(moduleKey); moduleKeyID == 0 {
		err = ecode.RequestErr
		return
	}
	uc, err = s.userConf(c, mid, moduleKeyID)
	if err != nil {
		return
	}
	if uc.CheckSum == 0 || uc.Timestamp == 0 {
		err = ecode.NotModified
		log.Warn("document empty(mid:%d)", mid)
		return
	}
	// 数据没有变动
	if uc.CheckSum == checkSum && uc.Timestamp == timestamp {
		err = ecode.NotModified
		return
	}
	rm, err = s.document(c, uc.CheckSum)
	if err != nil {
		return
	}
	setting = &module.Setting{
		Timestamp: uc.Timestamp,
		CheckSum:  uc.CheckSum,
		Data:      rm,
	}
	return
}

func (s *Service) DocumentMid(c context.Context, mid int64, moduleKey string, timestamp int64, checkSum int64, platForm string) (setting *module.Setting, err error) {
	var (
		uc          *model.UserConf
		rm          json.RawMessage
		moduleKeyID int
	)
	if moduleKeyID = module.VerifyModuleKey(moduleKey); moduleKeyID == 0 {
		err = ecode.RequestErr
		return
	}
	uc, err = s.userConf(c, mid, moduleKeyID)
	if err != nil {
		return
	}
	if uc.CheckSum == 0 || uc.Timestamp == 0 {
		err = ecode.NothingFound
		rm = nil
	} else {
		rm, err = s.document(c, uc.CheckSum)
	}
	if err != nil && ecode.Cause(err) != ecode.NothingFound {
		return
	}
	rm, err = s.MergeIncrMap(c, mid, "", moduleKeyID, rm)
	if rm == nil {
		err = ecode.NotModified
		return
	}
	setting = &module.Setting{
		Timestamp: uc.Timestamp,
		CheckSum:  uc.CheckSum,
		Data:      rm,
	}
	return
}

func (s *Service) MergeIncrMap(c context.Context, mid int64, buvid string, moduleKeyID int, rm json.RawMessage) (res json.RawMessage, err error) {
	var (
		ucDoc map[string]string
		bm    []byte
	)
	ucDoc, err = s.da.HgetAllUserDoc(c, mid, buvid, moduleKeyID)
	if err != nil {
		log.Error("s.da.HgetAllUserDoc(mid:%d, modulekey:%d) err(%v)", mid, moduleKeyID, err)
		res = rm
		return
	}
	if len(ucDoc) == 0 {
		res = rm
		return
	}
	player := v1.NewConfig(moduleKeyID, nil)
	player.Default()
	if player == nil {
		err = xecode.KvoModuleNotExist
		return
	}
	if rm != nil {
		if err = json.Unmarshal(rm, &player); err != nil {
			res = rm
			return
		}
	}
	player.Merge(ucDoc)
	if bm, err = json.Marshal(player); err != nil {
		res = rm
		return
	}
	res = json.RawMessage(bm)
	return
}

func (s *Service) userConf(c context.Context, mid int64, moduleKeyID int) (uc *model.UserConf, err error) {
	uc, err = s.da.UserConfRds(c, mid, moduleKeyID)
	if err != nil {
		log.Error("service.userConf.UserConfRds(%v,%v) err:%v", mid, moduleKeyID, err)
	}
	log.Warn("userConf cache (mid:%d uc:%+v)", mid, uc)
	if uc != nil {
		return
	}
	uc, err = s.da.UserConf(c, mid, moduleKeyID)
	if err != nil {
		log.Error("service.userConf(%v,%v) err:%v", mid, moduleKeyID, err)
		return
	}
	log.Warn("userConf db (mid:%d uc:%+v)", mid, uc)
	if uc == nil {
		uc = &model.UserConf{
			Mid:       mid,
			ModuleKey: moduleKeyID,
		}
	}
	s.da.SetUserConfRds(c, uc)
	return
}

func (s *Service) document(c context.Context, checkSum int64) (rm json.RawMessage, err error) {
	var (
		doc *model.Document
	)
	if v, ok := s.localCache.Get(checkSum); ok {
		cache.MetricHits.Add(1, "bts:document:local")
		rm = v
		return
	}
	cache.MetricMisses.Add(1, "bts:document:local")
	rm, err = s.da.DocumentRds(c, checkSum)
	if err != nil {
		log.Error("service.document.DocumentRds(%v) err:%v", checkSum, err)
	}

	if rm != nil {
		cache.MetricHits.Add(1, "bts:document:redis")
		s.localCache.Add(checkSum, rm)
		return
	}
	cache.MetricMisses.Add(1, "bts:document:redis")
	doc, err = s.da.Document(c, checkSum)
	if err != nil {
		log.Error("service.document(%v) err:%v", checkSum, err)
		return
	}
	if doc == nil {
		err = ecode.NothingFound
		return
	}
	rm = json.RawMessage(doc.Doc)
	s.da.SetDocumentRds(c, checkSum, rm)
	return
}

// AddDocument add a user document
func (s *Service) AddDocument(c context.Context, mid int64, moduleKey string, data string, timestamp int64, oldSum int64, now time.Time) (resp *model.UserConf, err error) {
	var (
		uc          *model.UserConf
		doc         *model.Document
		rm          json.RawMessage
		checkSum    int64
		tx          *sql.Tx
		moduleKeyID int
	)
	if moduleKeyID = module.VerifyModuleKey(moduleKey); moduleKeyID == 0 {
		return nil, ecode.RequestErr
	}
	if rm, checkSum, err = module.Result(moduleKeyID, data); err != nil {
		log.Error("service.GetModule(%v,%s) err:%v", moduleKey, data, err)
		return nil, ecode.RequestErr
	}
	if len(rm) > s.docLimit {
		err = xecode.KvoDataOverLimit
		return
	}
	if uc, err = s.da.UserConfMaster(c, mid, moduleKeyID); err != nil {
		log.Error("service.AddDocument.UserConf(%v,%v) err:%v", mid, moduleKeyID, err)
		return
	}
	if uc != nil {
		if uc.CheckSum == checkSum {
			log.Warn("s.AddDocument(mid:%d,uc:%+v)", mid, uc)
			s.da.SetUserConfRds(c, uc)
			err = ecode.NotModified
			return
		}
	}
	// trans
	tx, err = s.da.BeginTx(c)
	if err != nil {
		log.Error("s.da.BeginTx err:%v", err)
		return
	}
	if err = s.da.TxUpUserConf(c, tx, mid, moduleKeyID, checkSum, now); err != nil {
		log.Error("s.da.TxUpUserConf(%v,%v,%v) error(%v)", mid, moduleKeyID, checkSum, err)
		tx.Rollback()
		return
	}
	doc, err = s.da.Document(c, checkSum)
	if err != nil {
		tx.Rollback()
		return
	}
	if doc == nil {
		if err = s.da.TxUpDocuement(c, tx, checkSum, string(rm), now); err != nil {
			log.Error("s.da.TxUpDocuement(%v,%v,%v) error(%v)", mid, moduleKeyID, checkSum, err)
			tx.Rollback()
			return
		}
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit(), error(%v)", err)
	}
	resp = &model.UserConf{
		CheckSum:  checkSum,
		Timestamp: now.Unix(),
	}
	log.Warn("updateUcCache (mid:%d,moduleKeyID:%d,checkSum:%d)", mid, moduleKeyID, checkSum)
	newUc := &model.UserConf{
		Mid:       mid,
		Timestamp: now.Unix(),
		CheckSum:  checkSum,
		ModuleKey: moduleKeyID,
	}
	s.da.SetUserConfRds(c, newUc)
	s.da.DelUserDoc(c, mid, "", moduleKeyID)
	return
}

func (s *Service) DocumentBuvid(c context.Context, buvid string, moduleKey string, platForm string) (setting *module.Setting, err error) {
	var (
		rm          json.RawMessage
		moduleKeyID int
	)
	if moduleKeyID = v1.VerifyModuleKey(moduleKey); moduleKeyID == 0 {
		err = ecode.RequestErr
		return
	}
	rm, err = s.userDoc(c, 0, buvid, moduleKeyID)
	if err != nil {
		return
	}
	rm, err = s.MergeIncrMap(c, 0, buvid, moduleKeyID, rm)
	if rm == nil {
		err = ecode.NotModified
		return
	}
	setting = &module.Setting{
		Timestamp: time.Now().Unix(),
		CheckSum:  0,
		Data:      rm,
	}
	return
}
