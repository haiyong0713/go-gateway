package dao

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"go-common/library/ecode"
	"net/url"
	"strconv"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/space/admin/model"
)

const (
	_topPhotoUrl     = "/x/internal/space/clear/topphoto/arc"
	_tablewhitelist  = "whitelist"
	_whitelistAddSQL = "INSERT INTO whitelist (mid,mid_name,state,stime,etime,username) VALUES (?,?,?,?,?,?)"
)

// WhitelistAdd add whitelist
func (d *Dao) WhitelistAdd(args []*model.WhitelistAdd) (err error) {
	tx := d.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("dao.WhitelistAdd.Begin error(%v)", err)
		return
	}
	if len(args) > 0 {
		for _, white := range args {
			if err = tx.Model(&model.Whitelist{}).Exec(_whitelistAddSQL, white.Mid, white.MidName, white.State, white.Stime, white.Etime, white.Username).Error; err != nil {
				log.Error("dao.WhitelistAdd(%+v) error(%v)", white, err)
				if errRollback := tx.Rollback().Error; errRollback != nil {
					log.Error("dao.WhitelistAdd rollback error(%v)", errRollback)
				}
				return
			}
		}

	}
	err = tx.Commit().Error
	return
}

func (d *Dao) WhitelistFindById(id int64) (ret *model.WhitelistAdd, err error) {
	ret = &model.WhitelistAdd{}
	if err = d.DB.Table(_tablewhitelist).Where("id = ?", id).First(ret).Error; err != nil {
		log.Error("dao.WhitelistFindByMid find(%v) error(%v)", ret, err)
		return
	}
	return
}

func (d *Dao) ValidWhitelistMid(mid int64) (ok bool, err error) {
	tmp := &model.WhitelistAdd{}
	if err = d.DB.Table(_tablewhitelist).Where("mid=? and state != 3 and deleted=0", mid).First(tmp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return true, nil
		}
		log.Error("dao.ValidWhitelistMid find(%v) error(%v)", tmp, err)
		return false, err
	}
	return false, nil
}

// WhitelistIn query whiteist count
func (d *Dao) WhitelistIn(mids []int64) (whites map[int64]*model.Whitelist, err error) {
	var (
		whitelist []*model.Whitelist
	)
	whites = make(map[int64]*model.Whitelist, len(mids))
	if len(mids) == 0 {
		return nil, fmt.Errorf("mid不能为空")
	}
	if err = d.DB.Model(&model.Whitelist{}).Where("mid in (?)", mids).Find(&whitelist).
		Error; err != nil {
		log.Error("dao.WhitelistIn.Count(%+v) error(%v)", mids, err)
		return
	}
	for _, v := range whitelist {
		whites[v.Mid] = v
	}
	return
}

// WhitelistUp whiteist update
func (d *Dao) WhitelistUp(arg *model.WhitelistAdd) (err error) {
	w := map[string]interface{}{
		"id":      arg.ID,
		"deleted": 0,
	}
	if err = d.DB.Model(&model.Whitelist{}).Where(w).Update(arg).Error; err != nil {
		log.Error("dao.WhitelistUp.Update error(%v)", err)
		return
	}
	return
}

func (d *Dao) WhitelistDelete(id int64, t int) (err error) {
	up := map[string]interface{}{
		"deleted": model.Deleted,
	}
	if t == 1 {
		if err = d.DB.Table(_tablewhitelist).Where("deleted=0").Update(up).Error; err != nil {
			log.Error("dao.WhitelistDelte id(%d) error(%v)", id, err)
			return
		}
		return
	}
	if err = d.DB.Table(_tablewhitelist).Where("id = ?", id).Update(up).Error; err != nil {
		log.Error("dao.WhitelistDelte id(%d) error(%v)", id, err)
		return
	}
	return
}

func (d *Dao) ClearWhitelistInfo(c context.Context, mid int64) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))

	if err = d.http.Post(c, d.clearTopPhotoURL, "", params, &res); err != nil {
		log.Error("dao.ClearWhitelistInfo Get(%v) error(%v)", mid, err)
		return
	}
	if res.Code != 0 {
		// err = errors.Wrap(err, fmt.Sprintf("code(%d)", res.Code))
		log.Error("dao.ClearWhitelistInfo d.client.Get(%s) error(%v)", _topPhotoUrl+"?"+params.Encode(), err)
		err = ecode.Int(res.Code)
	}
	return
}

// WhitelistIndex whiteist
func (d *Dao) WhitelistIndex(mid int64, pn, ps, status int) (pager *model.WhitelistPager, err error) {
	var (
		whitelist []*model.Whitelist
	)
	w := map[string]interface{}{
		"deleted": 0,
	}
	pager = &model.WhitelistPager{
		Page: model.Page{
			Num:  pn,
			Size: ps,
		},
	}
	query := d.DB.Table(_tablewhitelist).Where(w)
	if mid != 0 {
		query = query.Where("mid = ?", mid)
	}
	if status != 0 {
		query = query.Where("state = ?", status)
	}
	if err = query.Count(&pager.Page.Total).Error; err != nil {
		log.Error("dao.WhitelistIndex.Count error(%v)", err)
		return
	}
	if err = query.Order("`mtime` DESC").Offset((pn - 1) * ps).Limit(ps).Find(&whitelist).Error; err != nil {
		log.Error("dao.WhitelistIndex.Find error(%v)", err)
		return
	}
	pager.Item = whitelist
	return
}

func (d *Dao) ChangeStatus(id int64, status int) (err error) {
	up := map[string]interface{}{
		"state": status,
	}
	if err = d.DB.Table(_tablewhitelist).Where("id = ?", id).Update(up).Error; err != nil {
		log.Error("dao.ChangeStatus id(%d) error(%v)", id, err)
		return
	}
	return
}

func (d *Dao) FindWhiteByStatus(status int) (ret []*model.WhitelistAdd, err error) {
	now := time.Now().Format(dateFormat)
	query := d.DB.Table(_tablewhitelist)
	ret = make([]*model.WhitelistAdd, 0)
	if status == model.StatusValid {
		if err = query.Where("etime < ? and state=1", now).Find(&ret).Error; err != nil {
			log.Error("dao.FindWhitelistByStatus.Find error(%v)", err)
			return
		}
	}
	if status == model.StatusReady {
		if err = query.Where("stime < ? and state=1", now).Find(&ret).Error; err != nil {
			log.Error("dao.FindWhitelistByStatus.Find error(%v)", err)
			return
		}
	}
	return
}
