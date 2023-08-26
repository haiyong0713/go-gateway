package service

import (
	"fmt"

	"go-common/library/log"
	"go-gateway/app/web-svr/space/admin/model"

	"github.com/jinzhu/gorm"
)

const (
	_upStatSQL = `INSERT INTO space_official (uid,name,icon,scheme,rcmd,ios_url,android_url,button,deleted) VALUES (?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE uid=?,name=?,icon=?,scheme=?,rcmd=?,ios_url=?,android_url=?,button=?,deleted=?`
)

// AddOfficial .
func (s *Service) AddOfficial(value *model.SpaceOfficial) (err error) {
	if err = s.dao.DB.Exec(_upStatSQL, value.Uid, value.Name, value.Icon, value.Scheme, value.Rcmd, value.IosUrl, value.AndroidUrl, value.Button, value.Deleted, value.Uid, value.Name, value.Icon, value.Scheme, value.Rcmd, value.IosUrl, value.AndroidUrl, value.Button, value.Deleted).Error; err != nil {
		log.Error("InsertUpdateOfficial value(%v) error(%v)", value, err)
		return
	}
	return
}

// UpdateOfficial .
func (s *Service) UpdateOfficial(req *model.SpaceOfficial) (err error) {
	if req.ID <= 0 {
		return fmt.Errorf("ID参数错误")
	}
	tmp := &model.SpaceOfficial{}
	if err = s.dao.DB.Where("uid = ?", req.Uid).First(tmp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("找不到uid为(%d)的数据", req.Uid)
		}
		log.Error("UpdateOfficial First req(%v) error(%v)", req, err)
		return
	}
	if tmp.ID != req.ID && tmp.Deleted == model.NotDelete {
		return fmt.Errorf("已存在相同UID(%d)数据", req.Uid)
	}
	//存在已删除数据 需要先删掉本条数据 然后更新已删除 相同uid的值 为有效
	tx := s.dao.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("UpdateOfficial recover req(%+v) error(%v)", req, r)
		}
		if err != nil {
			if err1 := tx.Rollback().Error; err1 != nil {
				log.Error("UpdateOfficial Rollback req(%+v) error(%v),error1(%v)", req, err, err1)
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("UpdateOfficial Commit req(%+v) error(%v)", req, err)
			return
		}
	}()
	if err = tx.Model(&model.SpaceOfficial{}).Where("id = ?", req.ID).UpdateColumn("deleted", model.Deleted).Error; err != nil {
		return
	}
	req.ID = tmp.ID
	if err = tx.Model(&model.SpaceOfficial{}).Save(req).Error; err != nil {
		return
	}
	return
}

// DeleteOfficial .
func (s *Service) DeleteOfficial(id int64) (err error) {
	if err = s.dao.DB.Model(&model.SpaceOfficial{}).Where("id = ?", id).UpdateColumn("deleted", model.Deleted).Error; err != nil {
		log.Error("DeleteOfficial value(%v) error(%v)", id, err)
		return
	}
	return
}

// Official .
func (s *Service) Official(req *model.SpaceOfficialParam) (pager *model.SpaceOfficialPager, err error) {
	pager = &model.SpaceOfficialPager{
		Page: model.Page{
			Num:  req.Pn,
			Size: req.Ps,
		},
	}
	value := make([]*model.SpaceOfficial, 0)
	w := map[string]interface{}{
		"deleted": model.NotDelete,
	}
	query := s.dao.DB.Model(&model.SpaceOfficial{})
	if err = query.Where(w).Count(&pager.Page.Total).Error; err != nil {
		log.Error("Official count error(%v)", err)
		return
	}
	if pager.Page.Total == 0 {
		return
	}
	if err = query.Offset((req.Pn - 1) * req.Ps).Where(w).Order("mtime DESC").Limit(req.Ps).Find(&value).Error; err != nil {
		log.Error("Official req(%v) error(%v)", req, err)
		return
	}
	pager.Item = value
	return
}
