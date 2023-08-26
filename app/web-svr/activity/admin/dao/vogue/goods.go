package vogue

import (
	"context"
	"fmt"

	"go-common/library/log"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"
)

const (
	_tableGoods    = "act_vogue_goods"
	_delete        = 1
	_setSoldOutSQL = "UPDATE `act_vogue_goods` SET attr = attr | (1 << 1) WHERE id = ?"
	_updateSQL     = "UPDATE `act_vogue_goods` SET `name`=?,`picture`=?,`type`=?,`score`=?,`stock`=?,`attr`= attr & 2 | (? << 0) WHERE id = ?"
	_insertSQL     = "INSERT INTO `act_vogue_goods` (`name`, `picture`, `type`, `score`, `stock`, `attr`) VALUES (?,?,?,?,?,?)"
	_goodsList     = "act_goods_list"
	_goods         = "act_goods_%d"
)

// List get goods list
func (d *Dao) GoodsList(c context.Context) (list []*voguemdl.GoodsData, err error) {
	db := d.DB.Table(_tableGoods)
	if err = db.Where("is_delete = 0").Order("id desc").Find(&list).Error; err != nil {
		log.Error("[VogueGoodsList] d.DB.Find, error(%v)", err)
	}
	return
}

// Delete goods
func (d *Dao) DelGoods(c context.Context, id int) (err error) {
	if err = d.DB.Table(_tableGoods).Where("id=? AND want=0", id).Update(map[string]interface{}{"is_delete": _delete}).Error; err != nil {
		log.Error("[DelVogueGoods] d.DB.Delete(), ID(%d) error(%v)", id, err)
	}
	return
}

// Modify goods
func (d *Dao) ModifyGoods(c context.Context, params *voguemdl.GoodsModifyParam) (err error) {
	if err = d.DB.Table(_tableGoods).Exec(_updateSQL, params.Name, params.Picture, params.Type, params.Score, params.Stock, params.AttrReal, params.ID).Error; err != nil {
		log.Error("[ModifyVogueGoods] d.DB.Modify(%d,%s,%s,%s,%d,%d,%d) error(%v)", params.ID, params.Name, params.Picture, params.Type, params.AttrReal, params.Score, params.Stock, err)
	}
	return
}

// Set goods soldOut
func (d *Dao) SetSoldOutGoods(c context.Context, id int) (err error) {
	if err = d.DB.Table(_tableGoods).Exec(_setSoldOutSQL, id).Error; err != nil {
		log.Error("[SetSoldOutVogueGoods] d.DB.SetSoldOut(), ID(%d) error(%v)", id, err)
	}
	_ = d.DelCacheConfig(c, _goodsList)
	_ = d.DelCacheConfig(c, fmt.Sprintf(_goods, id))
	return
}

// Add goods
func (d *Dao) AddGoods(c context.Context, params *voguemdl.GoodsAddParam) (err error) {
	if err = d.DB.Table(_tableGoods).Exec(_insertSQL, params.Name, params.Picture, params.Type, params.Score, params.Stock, params.AttrReal).Error; err != nil {
		log.Error("[AddVogueGoods] d.DB.Create(%s,%s,%s,%d,%d,%d) error(%v)", params.Name, params.Picture, params.Type, params.AttrReal, params.Score, params.Stock, err)
	}
	return
}
