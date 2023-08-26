package dao

import (
	"context"
	"database/sql"

	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/vogue"
)

const (
	_goodsListSQL    = "SELECT id,name,picture,type,score,send,stock,want,attr FROM act_vogue_goods WHERE is_delete=0"
	_goodsSQL        = "SELECT id,name,picture,type,score,send,stock,want,attr FROM act_vogue_goods WHERE id=? AND is_delete=0"
	_goodsAddSendSQL = "UPDATE act_vogue_goods SET send=send+1 WHERE id=? AND stock-send>0"
	_goodsAddWantSQL = "UPDATE act_vogue_goods SET want=want+1 WHERE id=?"
)

func (d *Dao) RawGoodsList(c context.Context) (res []*model.Goods, err error) {
	res = make([]*model.Goods, 0, 0)
	rows, err := d.db.Query(c, _goodsListSQL)
	if err != nil {
		log.Error("dmReader.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &model.Goods{}
		if err = rows.Scan(&r.Id, &r.Name, &r.Picture, &r.Type, &r.Score, &r.Send, &r.Stock, &r.Want, &r.Attr); err != nil {
			log.Error("row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) RawGoods(c context.Context, id int64) (res *model.Goods, err error) {
	res = new(model.Goods)
	row := d.db.QueryRow(c, _goodsSQL, id)
	if err = row.Scan(&res.Id, &res.Name, &res.Picture, &res.Type, &res.Score, &res.Send, &res.Stock, &res.Want, &res.Attr); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("RawGoods(%d) error(%v)", id, err)
	}
	return
}

func (d *Dao) GoodsAddSend(c context.Context, id int64) (affect int64, err error) {
	res, err := d.db.Exec(c, _goodsAddSendSQL, id)
	if err != nil {
		log.Error("d.InsertTask(%v) error(%v)", id, err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) GoodsAddWant(c context.Context, id int64) (err error) {
	_, err = d.db.Exec(c, _goodsAddWantSQL, id)
	if err != nil {
		log.Error("d.InsertTask(%v) error(%v)", id, err)
		return
	}
	return
}
