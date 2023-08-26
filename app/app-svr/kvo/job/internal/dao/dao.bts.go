package dao

import (
	"context"
	"encoding/json"

	"go-gateway/app/app-svr/kvo/job/internal/model"

	"go-common/library/ecode"
	"go-common/library/log"
)

func (d *dao) Document(ctx context.Context, checkSum int64) (rm json.RawMessage, err error) {
	var doc *model.Document
	rm, err = d.DocumentRds(ctx, checkSum)
	if err != nil {
		log.Error("dao.DocumentRds(%v) err:%v", checkSum, err)
	}
	if rm != nil {
		return
	}
	doc, err = d.documentDB(ctx, checkSum)
	if err != nil {
		log.Error("dao.document(%v) err:%v", checkSum, err)
		return
	}
	if doc == nil {
		err = ecode.NothingFound
		return
	}
	rm = json.RawMessage(doc.Doc)
	d.SetDocumentRds(ctx, checkSum, rm)
	return
}

func (d *dao) AsyncSetUserConf(ctx context.Context, uc *model.UserConf) {
	_ = d.cache.Do(ctx, func(c context.Context) {
		if err := d.SetUserConfRds(c, uc); err != nil {
			log.Error("d.setUserConfCache(uc:%+v) err(%v)", uc, err)
		}
		return
	})
	return
}

func (d *dao) AsyncSetDocument(ctx context.Context, checkSum int64, bm json.RawMessage) {
	_ = d.cache.Do(ctx, func(c context.Context) {
		if err := d.SetDocumentRds(c, checkSum, bm); err != nil {
			log.Error("d.setDocumentCache(checksum:%d,bm:%s) err(%v)", checkSum, string(bm), err)
		}
		return
	})
	return
}

func (d *dao) SetUserConf(ctx context.Context, uc *model.UserConf) {
	if err := d.SetUserConfRds(ctx, uc); err != nil {
		log.Error("d.setUserConfRds(uc:%+v) err(%v)", uc, err)
	}
	return
}

func (d *dao) SetDocument(ctx context.Context, checkSum int64, bm json.RawMessage) {
	if err := d.SetDocumentRds(ctx, checkSum, bm); err != nil {
		log.Error("d.setDocumentRds(checksum:%d,bm:%s) err(%v)", checkSum, string(bm), err)
	}
	return
}

func (d *dao) UserConf(ctx context.Context, mid int64, moduleKey int) (userConf *model.UserConf, err error) {
	if userConf, err = d.UserConfRds(ctx, mid, moduleKey); err != nil {
		err = nil
	}
	if userConf != nil {
		if userConf.CheckSum == 0 || userConf.Timestamp == 0 {
			userConf = nil
		}
		return
	}
	if userConf, err = d.userConfDB(ctx, mid, moduleKey); err != nil {
		log.Error("d.userConfDB(mid:%d, modulekey:%d) err(%v)", mid, moduleKey, err)
	}
	return
}
