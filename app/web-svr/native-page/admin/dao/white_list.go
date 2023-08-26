package dao

import (
	"context"
	"fmt"
	"strings"

	"go-gateway/app/web-svr/native-page/admin/model"

	"go-common/library/log"
)

const (
	tableWhiteList = "white_list"
	StateInvalid   = 0
	StateValid     = 1
	batchAddSQL    = "INSERT INTO `white_list` (`mid`,`creator`,`creator_uid`,`modifier`,`modifier_uid`,`state`,`from_type`) VALUES %s"
)

func (d *Dao) AddWhiteList(c context.Context, req *model.WhiteListRecord) (int, error) {
	if err := d.DB.Table(tableWhiteList).Create(req).Error; err != nil {
		log.Errorc(c, "Fail to create whiteList, req=%+v error=%+v", req, err)
		return 0, err
	}
	return req.ID, nil
}

func (d *Dao) UpdateWhiteList(c context.Context, id int, attrs map[string]interface{}) error {
	if err := d.DB.Table(tableWhiteList).Where("id=?", id).Update(attrs).Error; err != nil {
		log.Errorc(c, "Fail to update whiteList, id=%+v attrs=%+v error=%+v", id, attrs, err)
		return err
	}
	return nil
}

func (d *Dao) WhiteList(c context.Context, mid int64, pn, ps int64) ([]*model.WhiteListRecord, int64, error) {
	var list []*model.WhiteListRecord
	db := d.DB.Table(tableWhiteList).Where("state=?", StateValid)
	if mid != 0 {
		db = db.Where("mid=?", mid)
	}
	var count int64
	if err := db.Count(&count).Error; err != nil {
		log.Errorc(c, "Fail to count whiteList, mid=%+v error=%+v", mid, err)
		return nil, 0, err
	}
	if count == 0 {
		return []*model.WhiteListRecord{}, count, nil
	}
	var offset int64
	if pn >= 1 {
		offset = (pn - 1) * ps
	}
	if err := db.Offset(offset).Limit(ps).Order("id desc").Find(&list).Error; err != nil {
		log.Errorc(c, "Fail to find whiteList, mid=%+v error=%+v", mid, err)
		return nil, 0, err
	}
	return list, count, nil
}

func (d *Dao) WhiteListByMids(c context.Context, mids []int64) (map[int64]*model.WhiteListRecord, error) {
	if len(mids) == 0 {
		return map[int64]*model.WhiteListRecord{}, nil
	}
	var list []*model.WhiteListRecord
	db := d.DB.Table(tableWhiteList).Where("mid in (?)", mids).Where("state=?", StateValid)
	if err := db.Find(&list).Error; err != nil {
		log.Errorc(c, "Fail to batch get whiteList, mids=%+v error=%+v", mids, err)
		return nil, err
	}
	result := make(map[int64]*model.WhiteListRecord, len(list))
	for _, v := range list {
		result[v.Mid] = v
	}
	return result, nil
}

func (d *Dao) BatchAddWhiteList(c context.Context, attrs []*model.WhiteListRecord) error {
	if len(attrs) == 0 {
		return nil
	}
	sqlParts := make([]string, 0, len(attrs))
	sqlArgs := make([]interface{}, 0, len(attrs))
	for _, v := range attrs {
		sqlParts = append(sqlParts, "(?,?,?,?,?,?,?)")
		sqlArgs = append(sqlArgs, v.Mid, v.Creator, v.CreatorUID, v.Modifier, v.ModifierUID, v.State, v.FromType)
	}
	if err := d.DB.Exec(fmt.Sprintf(batchAddSQL, strings.Join(sqlParts, ",")), sqlArgs...).Error; err != nil {
		log.Errorc(c, "Fail to batch add whiteList, attrs=%+v error=%+v", attrs, err)
		return err
	}
	return nil
}
