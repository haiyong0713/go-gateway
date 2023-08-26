package dao

import (
	"context"
	"strconv"

	"go-common/library/log"

	"go-gateway/app/app-svr/archive-extra/service/api"

	"github.com/pkg/errors"
)

const (
	_insertAct = iota + 1
	_updateAct
	_deleteAct
)

// ExtraByAid 根据Aid查找
func (d *Dao) ExtraByAid(c context.Context, aid int64) (res map[string]string, err error) {
	if res, err = d.ArchiveExtraCacheByAid(c, aid); err != nil {
		log.Error("d.ArchiveExtraCacheByAid aid(%d) err(%+v)", aid, err)
	}
	return d.filterExtra(aid, res), err
}

// ExtraByAids 根据Aids查找
func (d *Dao) ExtraByAids(c context.Context, aids []int64) (res map[int64]*api.ArchiveExtraValueReply, err error) {
	valueMap := make(map[int64]map[string]string)
	res = make(map[int64]*api.ArchiveExtraValueReply)
	if valueMap, err = d.ArchiveExtraCacheByAids(c, aids); err != nil {
		log.Error("d.ArchiveExtraCacheByAids aids(%d) err(%+v)", aids, err)
	}
	for aid, value := range valueMap {
		res[aid] = &api.ArchiveExtraValueReply{
			ExtraInfo: d.filterExtra(aid, value),
		}
	}
	return res, err
}

// ExtraByKeys 根据Keys查找
func (d *Dao) ExtraByKeys(c context.Context, aid int64, keys []string) (res map[string]string, err error) {
	var value map[string]string
	res = make(map[string]string)
	if value, err = d.ArchiveExtraCacheByAid(c, aid); err != nil {
		log.Error("d.ArchiveExtraCacheByAid aid(%d) err(%+v)", aid, err)
	}
	for _, key := range keys {
		if _, ok := value[key]; ok {
			res[key] = value[key]
		}
	}
	return d.filterExtra(aid, res), err
}

// ExtraUpdate 更新/新增
func (d *Dao) ExtraUpdate(c context.Context, aid int64, bizType, bizValue string) (rows int64, err error) {
	id, err := d.QueryExtraId(c, aid, bizType)
	if err != nil {
		return
	}
	var act int
	// 如果存在则更新，不存在则新增
	if id > 0 {
		act = _updateAct
	} else {
		act = _insertAct
	}

	if err = d.TXExtraAndLog(c, id, aid, bizType, bizValue, act); err != nil {
		err = errors.Wrapf(err, "d.UpExtra err aid(%d) bizType(%s) bizValue(%s)", aid, bizType, bizValue)
		return
	}

	return
}

// ExtraDel 删除
func (d *Dao) ExtraDel(c context.Context, aid int64, bizType string) (err error) {
	id, err := d.QueryExtraId(c, aid, bizType)
	if err != nil {
		return
	}
	// 删除db
	if id > 0 {
		if err = d.TXExtraAndLog(c, id, aid, bizType, "", _deleteAct); err != nil {
			err = errors.Wrapf(err, "d.DelExtra err aid(%d) type(%s)", aid, bizType)
			return
		}
	} else {
		err = errors.Wrapf(err, "not exist err aid(%d) bizType(%s)", aid, bizType)
		return
	}
	return
}

// filterExtra 服务降级
func (d *Dao) filterExtra(aid int64, m map[string]string) map[string]string {
	filterKeys := make(map[string]struct{})
	values, ok := d.c.DemotionExtra.AidKeys[strconv.FormatInt(aid, 10)]
	if ok { // 降级过滤aid命中
		// 标记降级的bizKey
		for _, bizKey := range values {
			filterKeys[bizKey] = struct{}{}
		}
		res := make(map[string]string)
		for k, v := range m {
			// 命中降级，跳过
			if _, ok := filterKeys[k]; ok {
				continue
			}
			res[k] = v
		}
		return res
	}
	return m
}
