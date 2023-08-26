package service

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/web-svr/web-goblin/admin/internal/model"

	"go-common/library/sync/errgroup.v2"
)

const (
	_notDeleted = 0
	_deleted    = 1
	_hint       = 1
	_self       = 2
	_contact    = 3
	_guess      = 4
	_strategy   = 5
)

// AddBusiness .
func (s *Service) AddBusiness(c context.Context, param *model.GbCustomerBusiness) (err error) {
	preData := new(model.GbCustomerBusiness)
	s.dao.DB().Where("business=? AND customer_type = ? AND is_deleted = 0", param.Business, param.CustomerType).First(&preData)
	if preData.ID > 0 {
		return fmt.Errorf("业务名重复")
	}
	if err = s.dao.DB().Model(&model.GbCustomerBusiness{}).Create(param).Error; err != nil {
		log.Error("AddBusiness s.dao.DB().Model Create(%+v) error(%v)", param, err)
	}
	return
}

// EditBusiness .
func (s *Service) EditBusiness(c context.Context, param *model.GbCustomerBusiness) (err error) {
	if param.ID <= 0 {
		return fmt.Errorf("id不存在")
	}
	preData := new(model.GbCustomerBusiness)
	s.dao.DB().Where("id != ?", param.ID).Where("business = ? AND customer_type = ? AND is_deleted = 0", param.Business, param.CustomerType).First(&preData)
	if preData.ID > 0 {
		log.Error("EditBusiness s.dao.DB().Where id(%d) business(%s) error(%v)", param.ID, param.Business, err)
		return fmt.Errorf("业务名重复")
	}
	if err = s.dao.DB().Model(&model.GbCustomerBusiness{}).Update(param).Error; err != nil {
		log.Error("EditBusiness s.dao.DB().Model Update(%+v) error(%v)", param, err)
	}
	return
}

// DelBusiness .
func (s *Service) DelBusiness(c context.Context, id int64) (err error) {
	preData := new(model.GbCustomerBusiness)
	if err = s.dao.DB().Where("id=?", id).First(&preData).Error; err != nil {
		log.Error("DelBusiness s.dao.DB().Where id(%d) error(%v)", id, err)
		return
	}
	tx := s.dao.DB().Begin()
	if err = tx.Model(&model.GbCustomerBusiness{}).Where("id = ?", id).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
		log.Error("DelBusinesss dao.DB().Model DELETE id(%d) error(%v)", id, err)
		err = tx.Rollback().Error
		return
	}
	if err = tx.Model(&model.GbCustomerCenters{}).Where("business_type = ?", id).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
		log.Error("DelBusinesss GbCustomerCenters  dao.DB().Model DELETE business_type(%d) error(%v)", id, err)
		err = tx.Rollback().Error
		return
	}
	err = tx.Commit().Error
	return
}

// ListBusiness .
func (s *Service) ListBusiness(_ context.Context, customerType int64) (list []*model.GbCustomerBusiness, err error) {
	query := s.dao.DB().Model(&model.GbCustomerBusiness{}).Where("is_deleted = ?", _notDeleted)
	if customerType != 0 {
		query = query.Where("customer_type = ?", customerType)
	}
	if err = query.Order("rank DESC,id ASC").Find(&list).Error; err != nil {
		log.Error("ListBusiness s.dao.DB().Model Find Error (%v)", err)
	}
	return
}

// InfoBusiness .
func (s *Service) InfoBusiness(c context.Context, id int64) (data *model.GbCustomerBusiness, err error) {
	data = new(model.GbCustomerBusiness)
	if err = s.dao.DB().Model(&model.GbCustomerBusiness{}).Where("id=?", id).First(&data).Error; err != nil {
		log.Error("InfoBusiness Error (%v)", err)
	}
	return
}

// AddCustomer .
func (s *Service) AddCustomer(c context.Context, param *model.GbCustomerCenters) (err error) {
	if err = s.checkP(param); err != nil {
		return
	}
	if err = s.dao.DB().Model(&model.GbCustomerCenters{}).Create(param).Error; err != nil {
		log.Error("AddCustomer s.dao.DB().Model Create(%+v) error(%v)", param, err)
	}
	return
}

// EditCustomer .
func (s *Service) EditCustomer(c context.Context, param *model.GbCustomerCenters) (err error) {
	if err = s.checkP(param); err != nil {
		return
	}
	preData := new(model.GbCustomerCenters)
	if err = s.dao.DB().Where("id=?", param.ID).First(&preData).Error; err != nil {
		log.Error("EditCustomer s.dao.DB().Where id(%d) error(%d)", param.ID, err)
		return
	}
	if err = s.dao.DB().Model(&model.GbCustomerCenters{}).Update(param).Error; err != nil {
		log.Error("EditCustomer s.dao.DB().Model Update(%+v) error(%v)", param, err)
	}
	return
}

// DelCustomer .
func (s *Service) DelCustomer(c context.Context, id int64) (err error) {
	preData := new(model.GbCustomerCenters)
	if err = s.dao.DB().Where("id=?", id).First(&preData).Error; err != nil {
		log.Error("DelCustomer s.dao.DB().Where id(%d) error(%d)", id, err)
		return
	}
	if err = s.dao.DB().Model(&model.GbCustomerCenters{}).Where("id = ?", id).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
		log.Error("DelCustomer s.dao.DB().Model  DELETE id(%d) error(%v)", id, err)
	}
	return
}

// ListCustomer .
func (s *Service) ListCustomer(c context.Context, pn, ps int64) (list []*model.GbCustomerCenters, count int64, err error) {
	var (
		businesses []*model.GbCustomerBusiness
		bMap       map[int64]string
	)
	wg := errgroup.WithContext(c)
	wg.Go(func(ctx context.Context) (err error) {
		s.dao.DB().Model(&model.GbCustomerCenters{}).Where("is_deleted = ?", _notDeleted).Count(&count)
		return
	})
	wg.Go(func(ctx context.Context) (err error) {
		if err = s.dao.DB().Model(&model.GbCustomerCenters{}).Where("is_deleted = ?", _notDeleted).Offset((pn - 1) * ps).Limit(ps).Order("customer_type ASC,rank DESC,id ASC").Find(&list).Error; err != nil {
			log.Error("ListCustomer s.dao.DB().Model Find Error (%v)", err)
		}
		return
	})
	wg.Go(func(ctx context.Context) (err error) {
		if err = s.dao.DB().Model(&model.GbCustomerBusiness{}).Where("is_deleted = ?", _notDeleted).Find(&businesses).Error; err != nil {
			log.Error("ListBusiness s.dao.DB().Model Find Error (%v)", err)
		}
		return
	})
	if err = wg.Wait(); err != nil {
		return
	}
	bMap = make(map[int64]string, len(businesses))
	for _, b := range businesses {
		bMap[b.ID] = b.Business
	}
	for _, customer := range list {
		if customer.BusinessType > 0 {
			if name, ok := bMap[customer.BusinessType]; ok {
				customer.BusinessName = name
			}
		}
	}
	return
}

// InfoCustomer .
func (s *Service) InfoCustomer(c context.Context, id int64) (data *model.GbCustomerCenters, err error) {
	data = new(model.GbCustomerCenters)
	if err = s.dao.DB().Model(&model.GbCustomerCenters{}).Where("id=?", id).First(&data).Error; err != nil {
		log.Error("InfoCustomer Error (%v)", err)
	}
	return
}

func (s *Service) hintCheck(param *model.GbCustomerCenters) (err error) {
	var hintList []*model.GbCustomerCenters
	if param.CustomerType == _hint {
		if err = s.dao.DB().Model(&model.GbCustomerCenters{}).Where("is_deleted = ?", _notDeleted).Where("customer_type = ?", _hint).Find(&hintList).Error; err != nil {
			log.Error("ListCustomer s.dao.DB().Model Find Error (%v)", err)
			return
		}
		if len(hintList) > 0 {
			for _, cus := range hintList {
				if param.ID == cus.ID {
					continue
				}
				if param.Stime >= cus.Stime && param.Stime <= cus.Etime {
					log.Error("Customer hint stime param(%+v)  error(%v)", param, err)
					return fmt.Errorf("开始时间重复")
				} else if param.Etime >= cus.Stime && param.Etime <= cus.Etime {
					log.Error("Customer hint etime  param(%+v)  error(%v)", param, err)
					return fmt.Errorf("结束时间重复")
				}
			}
		}
	}
	return
}

// nolint:gocognit
func (s *Service) checkP(param *model.GbCustomerCenters) (err error) {
	if param.BusinessType <= 0 && (param.CustomerType == _contact) {
		return fmt.Errorf("业务ID错误")
	}
	if param.Title == "" && (param.CustomerType == _contact || param.CustomerType == _self || param.CustomerType == _strategy || param.CustomerType == _guess) {
		return fmt.Errorf("标题不能为空")
	}
	if param.Copywriting == "" && (param.CustomerType == _hint || param.CustomerType == _guess) {
		return fmt.Errorf("文案不能为空")
	}
	if param.Image == "" && (param.CustomerType == _self || param.CustomerType == _strategy) {
		return fmt.Errorf("图片不能为空")
	}
	if param.WebUrl == "" && (param.CustomerType == _self || param.CustomerType == _contact || param.CustomerType == _strategy) {
		return fmt.Errorf("web连接不能为空")
	}
	if param.H5Url == "" && (param.CustomerType == _self || param.CustomerType == _contact || param.CustomerType == _strategy) {
		return fmt.Errorf("h5连接不能为空")
	}
	if param.Stime <= 0 && (param.CustomerType == _hint || param.CustomerType == _strategy) {
		return fmt.Errorf("上线时间不能为空")
	}
	if param.Etime <= 0 && (param.CustomerType == _hint || param.CustomerType == _strategy) {
		return fmt.Errorf("下线时间不能为空")
	}
	if param.CustomerType == _hint {
		if err = s.hintCheck(param); err != nil {
			return
		}
	}
	return
}
