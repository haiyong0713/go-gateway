package show

import (
	"context"

	"go-common/library/log"

	"github.com/pkg/errors"

	"go-gateway/app/app-svr/app-feed/admin/model/icon"
)

// IconSave .
func (d *Dao) IconSave(c context.Context, ic *icon.Icon) (int64, error) {
	if ic == nil {
		return 0, nil
	}
	if ic.ID > 0 {
		upParam := map[string]interface{}{
			"module":         ic.Module,
			"icon":           ic.Icon,
			"global_red_dot": ic.GlobalRed,
			"effect_group":   ic.EffectGroup,
			"effect_url":     ic.EffectURL,
			"stime":          ic.Stime,
			"etime":          ic.Etime,
			"operator":       ic.Operator,
		}
		if err := d.DB.Model(&icon.Icon{}).Where("id=?", ic.ID).Update(upParam).Error; err != nil {
			err = errors.Wrapf(err, "db update err param(%+v)", upParam)
			return 0, err
		}
	} else {
		if err := d.DB.Model(&icon.Icon{}).Create(ic).Error; err != nil {
			err = errors.Wrapf(err, "db Create err param(%+v)", ic)
			return 0, err
		}
	}
	return ic.ID, nil
}

// UpdateIconState .
func (d *Dao) UpdateIconState(c context.Context, id int64, state int) (rows int64, err error) {
	query := d.DB.Model(&icon.Icon{}).Where("id=?", id).Update(map[string]int{"state": state})
	rows = query.RowsAffected
	err = query.Error
	return
}

// Icon .
func (d *Dao) Icon(c context.Context, id int64) (res *icon.Icon, err error) {
	res = new(icon.Icon)
	if err = d.DB.Model(&icon.Icon{}).Where("id=?", id).First(&res).Error; err != nil {
		log.Error("Icon First id(%d) error(%v)", id, err)
	}
	return
}

// Icons .
func (d *Dao) Icons(c context.Context, pn, ps int) (res []*icon.Icon, total int64, err error) {
	query := d.DB.Model(&icon.Icon{}).Where("state>?", icon.StateDel)
	if err = query.Count(&total).Error; err != nil {
		log.Error("Icons count error(%v)", err)
		return
	}
	if total == 0 {
		return
	}
	res = make([]*icon.Icon, 0)
	if err = query.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&res).Error; err != nil {
		log.Error("Icons Find error(%v)", err)
	}
	return
}
