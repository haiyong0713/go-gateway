package service

import (
	"fmt"
	"regexp"

	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/appstatic/admin/model"

	"github.com/jinzhu/gorm"
)

// Gray .
func (s *Service) Gray(id int64) (res *model.ResourceGray, err error) {
	where := map[string]interface{}{
		"resource_id": id,
	}
	res = &model.ResourceGray{}
	if err = s.DB.Where(where).First(res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			res = nil
			err = nil
		} else {
			log.Error("Gray first ID(%d) error(%v)", id, err)
		}
		return
	}
	return
}

// AddGray .
func (s *Service) AddGray(value *model.ResourceGray) (err error) {
	var (
		ok bool
	)
	if value.ID != 0 {
		return fmt.Errorf("id 必须为空")
	}
	if ok, err = isValidateSalt(value.Salt); err != nil {
		return fmt.Errorf("盐值参数错误%s", err.Error())
	}
	if !ok {
		return fmt.Errorf("盐值必须英文+数字")
	}
	if _, err = xstr.SplitInts(value.WhitelistInput); err != nil {
		return fmt.Errorf("mid文件不合法 %s", err.Error())
	}
	if err = s.DB.Create(value).Error; err != nil {
		return
	}
	return
}

func (s *Service) validateGrayParam(value *model.ResourceGray) (err error) {
	if value.BucketStart < -1 || value.BucketEnd > 999 {
		return fmt.Errorf("命中桶参数需要 -1到999")
	}
	return nil
}

// isValidateSalt .
func isValidateSalt(s string) (bool, error) {
	return regexp.MatchString(`^[a-zA-Z0-9]+$`, s)
}

// SaveGray .
func (s *Service) SaveGray(value *model.ResourceGray) (err error) {
	var (
		ok bool
	)
	if ok, err = isValidateSalt(value.Salt); err != nil {
		return fmt.Errorf("盐值参数错误%s", err.Error())
	}
	if !ok {
		return fmt.Errorf("盐值必须英文+数字")
	}
	if err = s.validateGrayParam(value); err != nil {
		return
	}
	if value.ID == 0 {
		return fmt.Errorf("id 不能为空")
	}
	if _, err = xstr.SplitInts(value.WhitelistInput); err != nil {
		return fmt.Errorf("mid文件不合法 %s", err.Error())
	}
	if err = s.DB.Save(value).Error; err != nil {
		return
	}
	return
}
