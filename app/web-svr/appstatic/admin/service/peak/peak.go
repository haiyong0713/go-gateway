package peak

import (
	"bytes"
	"context"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	libTime "go-common/library/time"
	"go-gateway/app/web-svr/appstatic/admin/conf"
	peakDao "go-gateway/app/web-svr/appstatic/admin/dao/peak"
	peakModel "go-gateway/app/web-svr/appstatic/admin/model/peak"
)

type Service struct {
	dao *peakDao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao: peakDao.New(c),
	}
	return
}

func (s *Service) AddPeak(peak *peakModel.Peak, appVer []peakModel.AppVer) (err error) {
	tx := s.dao.DB.Begin()
	db := tx.Model(&peakModel.Peak{}).Create(peak)
	if err = db.Error; err != nil {
		log.Error("peakSvr.AddPeak create peak error(%v)", err)
		return
	}
	for _, v := range appVer {
		v.PeakID = peak.ID
		v.Deleted = peakModel.NotDeleted
		db := tx.Model(&peakModel.AppVer{}).Create(v)
		if err = db.Error; err != nil {
			log.Error("peakSvr.AddPeak create peak app_ver error(%v)", err)
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		return
	}
	//彩蛋那里增加了一个日志，这里还未添加
	return
}

func (s *Service) IndexPeak(param *peakModel.IndexParam) (values *peakModel.IndexPager, err error) {
	const (
		_transferTimeToFront libTime.Time = 1000
	)
	values = &peakModel.IndexPager{
		Page: peakModel.Page{
			Num:  param.PageNumber,
			Size: param.PageSize,
		},
	}
	w := map[string]interface{}{
		"is_deleted": peakModel.NotDeleted,
	}
	if param.ID != "" {
		w["id"] = param.ID
	}
	query := s.dao.DB.Model(&peakModel.Index{}).Where(w)
	//资源类型
	if param.Type != 0 {
		query = query.Where("type = ?", param.Type)
	}
	// 模糊搜索url
	if param.Url != "" {
		query = query.Where("url like ?", "%"+param.Url+"%")
	}
	// 模糊搜索 文件名
	if param.FileName != "" {
		query = query.Where("file_name like ?", "%"+param.FileName+"%")
	}
	//生效时间
	if param.EffectTime != 0 {
		query = query.Where("effect_time >= ?", param.EffectTime)
	}
	//过期时间
	if param.ExpireTime != 0 {
		query = query.Where("expire_time <= ?", param.ExpireTime)
	}

	if err = query.Order("`id` DESC").Offset((param.PageNumber - 1) * param.PageSize).Limit(param.PageSize).Find(&values.Item).Error; err != nil {
		log.Error("peakSrc.IndexEgg Index list error(%v)", err)
		return
	}
	if err = query.Count(&values.Page.Total).Error; err != nil {
		log.Error("peakSrc.IndexEgg Index count error(%v)", err)
		return
	}
	for k, v := range values.Item {
		if v.ExpireTime < libTime.Time(time.Now().Unix()) {
			v.OnlineStatus = peakModel.NotOnline
		}
		v.EffectTime = v.EffectTime * _transferTimeToFront
		v.ExpireTime = v.ExpireTime * _transferTimeToFront
		appVer := []peakModel.AppVer{}
		w := map[string]interface{}{
			"peak_id":    v.ID,
			"is_deleted": peakModel.NotDeleted,
		}
		if err = s.dao.DB.Model(&peakModel.AppVer{}).Where(w).Find(&appVer).Error; err != nil {
			log.Error("peakSrv.IndexPeak AppVer error(%v)", err)
			return
		}
		values.Item[k].AppVer = appVer
	}
	return
}

func (s *Service) UpdatePeak(peak *peakModel.Peak, appVer []peakModel.AppVer) (err error) {
	tx := s.dao.DB.Begin()
	db := tx.Model(&peakModel.Peak{}).Where("id = ?", peak.ID).Save(peak)
	if err = db.Error; err != nil {
		log.Error("peakSrv.UpdatePeak UpdatePeak peak error(%v)", err)
		return
	}
	db = tx.Model(&peakModel.AppVer{}).Where("peak_id = ?", peak.ID).Update("is_deleted", peakModel.Deleted)
	if err = db.Error; err != nil {
		log.Error("peakSrv.UpdatePeak UpdatePeak peak error(%v)", err)
		return
	}
	for _, v := range appVer {
		v.PeakID = peak.ID
		v.Deleted = peakModel.NotDeleted
		db := tx.Model(&peakModel.AppVer{}).Create(v)
		if err = db.Error; err != nil {
			log.Error("peakSvr.UpdatePeak create peak app_ver error(%v)", err)
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		return
	}
	return
}

func (s *Service) PublishPeak(id uint, onlineStatus uint8) (err error) {
	if err = s.dao.DB.Model(&peakModel.Peak{}).Where("id = ?", id).Update("online_status", onlineStatus).Error; err != nil {
		log.Error("peakSrv.PublishPeak update error(%v)", err)
	}
	return
}

func (s *Service) DeletePeak(id uint) (err error) {
	tx := s.dao.DB.Begin()
	if err = tx.Model(&peakModel.Peak{}).Where("id = ?", id).Update("is_deleted", peakModel.Deleted).Error; err != nil {
		log.Error("peakSrv.DeletePeak Update error(%v)", err)
		return
	}
	if err = tx.Model(&peakModel.AppVer{}).Where("id = ?", id).Update("is_deleted", peakModel.Deleted).Error; err != nil {
		log.Error("peakSrv.DeletePeak Update error(%v)", err)
		return
	}
	if err = tx.Commit().Error; err != nil {
		return
	}
	return
}

func (s *Service) FileMd5(content []byte) (md5Str string, err error) {
	return s.dao.FileMd5(content)
}

func (s *Service) UploadPeak(c context.Context, fileType string, body []byte) (url string, err error) {
	if len(body) == 0 {
		err = ecode.FileNotExists
		return
	}
	url, err = s.dao.Upload(c, fileType, bytes.NewReader(body))
	if err != nil {
		log.Error("s.bfs.Upload error(%v)", err)
	}
	return
}
