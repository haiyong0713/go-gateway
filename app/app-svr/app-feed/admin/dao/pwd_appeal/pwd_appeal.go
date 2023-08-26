package pwd_appeal

import (
	"github.com/jinzhu/gorm"
	"go-common/library/log"

	model "go-gateway/app/app-svr/app-feed/admin/model/pwd_appeal"
)

func (d *Dao) SearchPwdAppeal(req *model.ListReq, needTotal bool) ([]*model.PwdAppeal, int64, error) {
	rly := make([]*model.PwdAppeal, 0, req.Ps)
	db := d.db.Model(&model.PwdAppeal{})
	if req.Mid > 0 {
		db = db.Where("mid = ?", req.Mid)
	}
	if req.DeviceToken != "" {
		db = db.Where("device_token = ?", req.DeviceToken)
	}
	if req.State > 0 {
		db = db.Where("state = ?", req.State)
	}
	if req.Mode > 0 {
		db = db.Where("mode = ?", req.Mode)
	}
	if req.BeginTime != "" {
		db = db.Where("ctime >= ?", req.BeginTime)
	}
	if req.EndTime != "" {
		db = db.Where("ctime <= ?", req.EndTime)
	}
	total := func() int64 {
		if !needTotal {
			return 0
		}
		var total int64
		if err := db.Count(&total).Error; err != nil {
			log.Error("Fail to query pwd_appeal total, req=%+v error=%+v", req, err)
			return 0
		}
		return total
	}()
	if req.Pn > 0 && req.Ps > 0 {
		offset := (req.Pn - 1) * req.Ps
		db = db.Offset(offset).Limit(req.Ps)
	}
	if err := db.Order("id DESC").Find(&rly).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, nil
		}
		log.Error("Fail to search pwd_appeal, req=%+v error=%+v", req, err)
		return nil, 0, err
	}
	return rly, total, nil
}

func (d *Dao) PwdAppeal(id int64) (*model.PwdAppeal, error) {
	appeal := &model.PwdAppeal{}
	if err := d.db.Model(&model.PwdAppeal{}).Where("id = ?", id).First(&appeal).Error; err != nil {
		log.Error("Fail to query pwd_appeal, id=%d error=%+v", id, err)
		return nil, err
	}
	return appeal, nil
}

func (d *Dao) PassAppeal(id int64, pwd, operator string) error {
	fields := map[string]interface{}{"pwd": pwd, "state": model.StatePass, "operator": operator}
	if err := d.db.Model(&model.PwdAppeal{}).Where("id=?", id).Where("state=?", model.StatePending).Update(fields).Error; err != nil {
		log.Error("Fail to update pwd_appeal of pass, id=%d fields=%+v error=%+v", id, fields, err)
		return err
	}
	return nil
}

func (d *Dao) RejectAppeal(id int64, reason, operator string) error {
	fields := map[string]interface{}{"reject_reason": reason, "state": model.StateReject, "operator": operator}
	if err := d.db.Model(&model.PwdAppeal{}).Where("id=?", id).Where("state=?", model.StatePending).Update(fields).Error; err != nil {
		log.Error("Fail to update pwd_appeal of reject, id=%d fields=%+v error=%+v", id, fields, err)
		return err
	}
	return nil
}

func (d *Dao) CreatePwdAppeal(pwdAppeal *model.PwdAppeal) (int64, error) {
	if err := d.db.Model(&model.PwdAppeal{}).Create(pwdAppeal).Error; err != nil {
		log.Error("Fail to create pwd_appeal, data=%+v error=%+v", pwdAppeal, err)
		return 0, err
	}
	return pwdAppeal.ID, nil
}

func (d *Dao) PendingPwdAppeal(mobile int64) (int64, error) {
	appeal := &model.PwdAppeal{}
	db := d.db.Model(&model.PwdAppeal{}).Select("id").Where("mobile=?", mobile).Where("state=?", model.StatePending)
	if err := db.First(&appeal).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		log.Error("Fail to query pending pwd_appeal, mobile=%d error=%+v", mobile, err)
		return 0, err
	}
	return appeal.ID, nil
}

func (d *Dao) DelUploadKey(id int64) error {
	fields := map[string]interface{}{"upload_key": ""}
	if err := d.db.Model(&model.PwdAppeal{}).Where("id=?", id).Update(fields).Error; err != nil {
		log.Error("Fail to del upload_key of pwd_appeal, id=%+v error=%+v", id, err)
		return err
	}
	return nil
}
