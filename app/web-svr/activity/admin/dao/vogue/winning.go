package vogue

import (
	"context"
	"fmt"

	xsql "go-common/library/database/sql"
	"go-common/library/log"

	lotmdl "go-gateway/app/web-svr/activity/admin/model/lottery"
)

const (
	_tableAddress       = "act_lottery_gift_address_%d"
	_tableWin           = "act_lottery_win_%d"
	_giftWinList        = "SELECT a.id,a.mid,a.gift_id,a.cdkey,a.ctime,a.mtime,IFNULL(b.address_id,0) FROM %s AS a LEFT JOIN %s AS b ON a.mid=b.mid WHERE a.mid!=0 LIMIT ? OFFSET ?"
	_giftWinListWithUid = "SELECT a.id,a.mid,a.gift_id,a.cdkey,a.ctime,a.mtime,IFNULL(b.address_id,0) FROM %s AS a LEFT JOIN %s AS b ON a.mid=b.mid WHERE a.mid!=0 AND a.mid = ? LIMIT ? OFFSET ?"
)

// GiftWinList get gift win list
func (d *Dao) GiftWinList(c context.Context, id int64, mid, pn, ps int64) (result []*lotmdl.GiftWinInfo, count int64, err error) {
	var (
		rows      *xsql.Rows
		tableWin  = fmt.Sprintf(_tableWin, id)
		tableAddr = fmt.Sprintf(_tableAddress, id)
	)
	if pn == 0 || ps == 0 {
		pn = 1
		ps = 50000
	}
	if mid > 0 {
		err = d.DB.Table(tableWin).Select("id").Where("mid=?", mid).Count(&count).Error
		if err != nil {
			log.Errorc(c, "lottery@GiftWinList d.db.Query() failed. error(%v)", err)
		}
		if rows, err = d.lotDB.Query(c, fmt.Sprintf(_giftWinListWithUid, tableWin, tableAddr), mid, ps, (pn-1)*ps); err != nil {
			log.Errorc(c, "lottery@GiftWinList d.db.Query() failed. error(%v)", err)
			return
		}
	} else {
		err = d.DB.Table(tableWin).Select("id").Count(&count).Error
		if err != nil {
			log.Errorc(c, "lottery@GiftWinList d.db.Query() failed. error(%v)", err)
		}
		if rows, err = d.lotDB.Query(c, fmt.Sprintf(_giftWinList, tableWin, tableAddr), ps, (pn-1)*ps); err != nil {
			log.Errorc(c, "lottery@GiftWinList d.db.Query() failed. error(%v)", err)
			return
		}
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &lotmdl.GiftWinInfo{}
		if err = rows.Scan(&tmp.ID, &tmp.Mid, &tmp.GiftId, &tmp.CDKey, &tmp.CTime, &tmp.MTime, &tmp.GiftAddrID); err != nil {
			if err == xsql.ErrNoRows {
				tmp = nil
				err = nil
				return
			}
			log.Errorc(c, "lottery@GiftWinList rows.Scan() failed. error(%v)", err)
			return
		}
		result = append(result, tmp)
	}
	err = rows.Err()
	return
}
