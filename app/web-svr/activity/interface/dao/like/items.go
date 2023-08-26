package like

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

// ItemAndContent .
func (d *Dao) ItemAndContent(c context.Context, item *like.Item, cont *like.LikeContent) (res int64, err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = d.db.Begin(c); err != nil {
		log.Error("d.db.Begin() error(%v)", err)
		return
	}
	defer func() {
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if res, err = d.TxAddLike(c, tx, item); err != nil {
		return
	}
	cont.ID = res
	if err = d.TxAddContent(c, tx, cont); err != nil {
		return
	}
	return
}

// ItemAndContentNew .
func (d *Dao) ItemAndContentNew(c context.Context, item *like.Item, cont *like.LikeContent) (res int64, err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = d.db.Begin(c); err != nil {
		log.Errorc(c, "ItemAndContentNew d.db.Begin() error(%v)", err)
		return
	}
	defer func() {
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "ItemAndContentNew tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "ItemAndContentNew tx.Commit() error(%v)", err)
		}
	}()
	if res, err = d.TxAddLike(c, tx, item); err != nil {
		return
	}
	cont.ID = res
	if err = d.TxAddContentNew(c, tx, cont); err != nil {
		return
	}
	return
}
