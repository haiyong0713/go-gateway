package dao

import (
	"go-gateway/app/web-svr/appstatic/admin/model"

	"go-common/library/ecode"
)

func (d *Dao) FetchDolbyWhiteList() ([]*model.DolbyWhiteList, error) {
	reply := make([]*model.DolbyWhiteList, 0)
	if err := d.GWDB.Model(&model.DolbyWhiteList{}).Where("is_deleted=?", 0).Find(&reply).Error; err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) AddDolbyWhiteList(in *model.DolbyWhiteList) error {
	return d.GWDB.Model(&model.DolbyWhiteList{}).Create(in).Error
}

func (d *Dao) SaveDolbyWhiteList(in *model.DolbyWhiteList) error {
	return d.GWDB.Model(&model.DolbyWhiteList{}).Where("id=?", in.ID).Update(in).Error
}

func (d *Dao) DeleteDolbyWhiteList(id int64) error {
	return d.GWDB.Table("dolby_vision_whitelist").Where("id=?", id).Update("is_deleted", 1).Error
}

func (d *Dao) FetchQnBlackList() ([]*model.QnBlackList, error) {
	reply := make([]*model.QnBlackList, 0)
	if err := d.GWDB.Model(&model.QnBlackList{}).Where("is_deleted=?", 0).Find(&reply).Error; err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) AddQnBlackList(in *model.QnBlackList) error {
	return d.GWDB.Model(&model.QnBlackList{}).Create(in).Error
}

func (d *Dao) SaveQnBlackList(in *model.QnBlackList) error {
	return d.GWDB.Model(&model.QnBlackList{}).Where("id=?", in.ID).Update(in).Error
}

func (d *Dao) DeleteQnBlackList(id int64) error {
	return d.GWDB.Table("qn_blacklist").Where("id=?", id).Update("is_deleted", 1).Error
}

func (d *Dao) FetchLimitFreeList() ([]*model.LimitFreeInfo, error) {
	reply := make([]*model.LimitFreeInfo, 0)
	if err := d.GWDB.Model(&model.LimitFreeInfo{}).Where("is_deleted=?", 0).Find(&reply).Error; err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) FetchLimitFree(id int64) (*model.LimitFreeInfo, error) {
	reply := &model.LimitFreeInfo{}
	if err := d.GWDB.Model(&model.LimitFreeInfo{}).Where("id=?", id).Find(&reply).Error; err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) FetchLimitFreeByAid(aid int64) (*model.LimitFreeInfo, error) {
	reply := &model.LimitFreeInfo{}
	if err := d.GWDB.Model(&model.LimitFreeInfo{}).Where("aid=? and is_deleted=?", aid, 0).Find(&reply).Error; err != nil {
		if ecode.Cause(err) == ecode.NothingFound {
			return reply, nil
		}
		return nil, err
	}
	return reply, nil
}

func (d *Dao) AddLimitFreeInfo(in *model.LimitFreeInfo) error {
	return d.GWDB.Model(&model.LimitFreeInfo{}).Create(in).Error
}

func (d *Dao) EditLimitFreeInfo(in *model.LimitFreeInfo) error {
	return d.GWDB.Model(&model.LimitFreeInfo{}).Where("id=?", in.ID).Save(in).Error
}

func (d *Dao) DeleteLimitFreeInfo(id int64) error {
	return d.GWDB.Table("resolution_limit_free").Where("id=?", id).Update("is_deleted", 1).Error
}
