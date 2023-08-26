package dao

import (
	"context"
	"encoding/json"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/appstatic/admin/model"

	"github.com/pkg/errors"
)

func (d *Dao) CreateAppInfo(_ context.Context, info *model.AppInfo) error {
	return d.DB.Model(&model.AppInfo{}).Create(info).Error
}

func (d *Dao) ShowAppInfoList(_ context.Context) ([]*model.AppInfo, error) {
	infoList := make([]*model.AppInfo, 0)
	if err := d.DB.Model(&model.AppInfo{}).Where("is_deleted=?", 0).Find(&infoList).Error; err != nil {
		return nil, err
	}
	return infoList, nil
}

func (d *Dao) ShowAppInfoDetail(_ context.Context, appKey string) (*model.AppInfo, error) {
	detail := new(model.AppInfo)
	if err := d.DB.Model(&model.AppInfo{}).Where("app_key=? AND is_deleted=?", appKey, 0).Find(&detail).Error; err != nil {
		return nil, err
	}
	return detail, nil
}

func (d *Dao) UpdateAppInfo(_ context.Context, info *model.AppInfo) error {
	return d.DB.Model(&model.AppInfo{}).Where("app_key=?", info.AppKey).Update(info).Error
}

func (d *Dao) DeleteAppInfo(_ context.Context, appKey string) error {
	return d.DB.Table("chronos_app").Where("app_key=?", appKey).Update("is_deleted", 1).Error
}

func (d *Dao) ShowServiceInfoList(_ context.Context) ([]*model.ServiceInfo, error) {
	infoList := make([]*model.ServiceInfo, 0)
	if err := d.DB.Model(&model.ServiceInfo{}).Where("is_deleted=?", 0).Find(&infoList).Error; err != nil {
		return nil, err
	}
	return infoList, nil
}

func (d *Dao) ShowServiceInfoDetail(_ context.Context, serviceKey string) (*model.ServiceInfo, error) {
	detail := new(model.ServiceInfo)
	if err := d.DB.Model(&model.ServiceInfo{}).Where("service_key=? AND is_deleted=?", serviceKey, 0).Find(&detail).Error; err != nil {
		return nil, err
	}
	return detail, nil
}

func (d *Dao) CreateServiceInfo(_ context.Context, info *model.ServiceInfo) error {
	return d.DB.Model(&model.ServiceInfo{}).Create(info).Error
}

func (d *Dao) UpdateServiceInfo(_ context.Context, info *model.ServiceInfo) error {
	return d.DB.Model(&model.ServiceInfo{}).Where("service_key=?", info.ServiceKey).Update(info).Error
}

func (d *Dao) DeleteServiceInfo(_ context.Context, serviceKey string) error {
	return d.DB.Table("chronos_service").Where("service_key=?", serviceKey).Update("is_deleted", 1).Error
}

func (d *Dao) CreatePackageAudit(_ context.Context, audit *model.PackageAudit) (int64, error) {
	if err := d.DB.Model(&model.PackageAudit{}).Create(audit).Error; err != nil {
		return 0, err
	}
	return int64(audit.ID), nil
}

func (d *Dao) BatchGetAuditPrePackageInfo(_ context.Context, ids []int64) (map[int64]*model.PrePackageInfo, error) {
	rows, err := d.DB.Table("chronos_package_show").Select("id,uuid,version,app_key,service_key").Where("id IN (?)", ids).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[int64]*model.PrePackageInfo)
	for rows.Next() {
		versionInfo := &model.PrePackageInfo{}
		if err := rows.Scan(&versionInfo.ID, &versionInfo.UUID, &versionInfo.Version, &versionInfo.AppKey, &versionInfo.ServiceKey); err != nil {
			return nil, errors.Wrapf(err, "BatchGetPackageVersion rows Scan error")
		}
		res[versionInfo.ID] = versionInfo
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) GetPackageAuditInfo(_ context.Context, auditID int64) (*model.PackageAudit, error) {
	packageAudit := new(model.PackageAudit)
	if err := d.DB.Model(&model.PackageAudit{}).Where("id=? AND audit_status=?", auditID, 0).Find(&packageAudit).Error; err != nil {
		return nil, errors.Wrapf(err, "GetPackageAuditInfo error auditID(%d)", auditID)
	}
	return packageAudit, nil
}

func (d *Dao) CreatePackageAndAftermath(_ context.Context, info *model.PackageInfo, auditID int64) error {
	var err error
	tx := d.DB.Begin()
	defer func() {
		if err != nil {
			if errRollback := tx.Rollback().Error; errRollback != nil {
				log.Error("tx.Rollback() error(%+v)", err)
				return
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("tx.Commit() error(%v)", err)
			return
		}
	}()
	if err = tx.Model(&model.PackageInfo{}).Create(info).Error; err != nil {
		return err
	}
	if err = tx.Model(&model.PackageAudit{}).Where("id=?", auditID).Update("audit_status", 1).Error; err != nil {
		return err
	}
	return nil
}

func (d *Dao) UpdatePackageAndAftermath(_ context.Context, info *model.PackageInfo, auditID int64) error {
	var err error
	tx := d.DB.Begin()
	defer func() {
		if err != nil {
			if errRollback := tx.Rollback().Error; errRollback != nil {
				log.Error("tx.Rollback() error(%+v)", err)
				return
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("tx.Commit() error(%v)", err)
			return
		}
	}()
	if err = tx.Model(&model.PackageInfo{}).Where("id=?", info.ID).Save(info).Error; err != nil {
		return err
	}
	if err = tx.Model(&model.PackageAudit{}).Where("id=?", auditID).Update("audit_status", 1).Error; err != nil {
		return err
	}
	return nil
}

func (d *Dao) DeletePackageAndAftermath(_ context.Context, id int64, auditID, version int64) error {
	var err error
	tx := d.DB.Begin()
	defer func() {
		if err != nil {
			if errRollback := tx.Rollback().Error; errRollback != nil {
				log.Error("tx.Rollback() error(%+v)", err)
				return
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("tx.Commit() error(%v)", err)
			return
		}
	}()
	if err = tx.Model(&model.PackageInfo{}).Where("id=?", id).Update(map[string]interface{}{"is_deleted": 1, "version": version + 1}).Error; err != nil {
		return err
	}
	if err = tx.Model(&model.PackageAudit{}).Where("id=?", auditID).Update("audit_status", 1).Error; err != nil {
		return err
	}
	return nil
}

func (d *Dao) RankPackageAndAftermath(_ context.Context, packageInfoInOrder map[int64]*model.PackageInfo, auditID int64) error {
	var err error
	tx := d.DB.Begin()
	defer func() {
		if err != nil {
			if errRollback := tx.Rollback().Error; errRollback != nil {
				log.Error("tx.Rollback() error(%+v)", err)
				return
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("tx.Commit() error(%v)", err)
			return
		}
	}()
	for id, v := range packageInfoInOrder {
		if err = tx.Model(&model.PackageInfo{}).Where("id=?", id).Update(map[string]interface{}{"rank": v.Rank, "version": v.Version + 1}).Error; err != nil {
			return err
		}
	}
	if err = tx.Model(&model.PackageAudit{}).Where("id=?", auditID).Update("audit_status", 1).Error; err != nil {
		return err
	}
	return nil
}

func (d *Dao) AuditReject(_ context.Context, auditID int64) error {
	return d.DB.Model(&model.PackageAudit{}).Where("id=? AND audit_status=?", auditID, 0).Error
}

func (d *Dao) AuditList(_ context.Context, appKey, serviceKey string) ([]*model.PackageAudit, error) {
	auditList := make([]*model.PackageAudit, 0)
	err := d.DB.Model(&model.PackageAudit{}).Where("audit_status=? AND app_key=? AND service_key=?", 0, appKey, serviceKey).Find(&auditList).Error
	if err != nil {
		return nil, err
	}
	return auditList, nil
}

func (d *Dao) ShowPackageInfoList(_ context.Context, appKey, serviceKey string) ([]*model.PackageInfo, error) {
	infoList := make([]*model.PackageInfo, 0)
	if err := d.DB.Model(&model.PackageInfo{}).Where("is_deleted=? AND app_key=? AND service_key=?", 0, appKey, serviceKey).Order("rank desc").Find(&infoList).Error; err != nil {
		return nil, err
	}
	return infoList, nil
}

func (d *Dao) ShowPackageInfoDetail(_ context.Context, uuid string) (*model.PackageInfo, error) {
	detail := new(model.PackageInfo)
	if err := d.DB.Model(&model.PackageInfo{}).Where("uuid=? AND is_deleted=?", uuid, 0).Find(&detail).Error; err != nil {
		return nil, err
	}
	return detail, nil
}

func (d *Dao) FetchAllPackageByAppAndService() (map[string][]*model.PackageInfo, error) {
	rows, err := d.DB.Model(&model.PackageInfo{}).Select("rank,app_key,service_key,resource_url,gray,black_list,white_list,video_list,rom_version,net_type,device_type,engine_version,buildlimit_exp,sign,md5").Where("is_deleted=?", 0).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	infoByAppAndServiceKey := make(map[string][]*model.PackageInfo)
	for rows.Next() {
		info := &model.PackageInfo{}
		if err = rows.Scan(&info.Rank, &info.AppKey, &info.ServiceKey, &info.ResourceUrl, &info.Gray, &info.BlackList, &info.WhiteList,
			&info.VideoList, &info.RomVersion, &info.NetType, &info.DeviceType, &info.EngineVersion, &info.BuildLimitExp, &info.Sign, &info.MD5); err != nil {
			return nil, errors.Wrapf(err, "FetchAllPackageInfo rows.Scan error")
		}
		infoByAppAndServiceKey[info.PackageInfoMatchKey()] = append(infoByAppAndServiceKey[info.PackageInfoMatchKey()], info)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return infoByAppAndServiceKey, nil
}

func (d *Dao) SavePackageInfoRulesToAppView(c context.Context, rules map[string][]*model.PackageInfo) error {
	str, err := json.Marshal(rules)
	if err != nil {
		return err
	}
	eg := errgroup.WithContext(c)
	for _, v := range d.playerRedis { // 并发保存
		red := v
		eg.Go(func(ctx context.Context) error {
			conn := red.Get(ctx)
			defer conn.Close()
			if _, err = conn.Do("SET", "chronosV2", str); err != nil {
				return err
			}
			return nil
		})
	}
	return eg.Wait()
}

func (d *Dao) BatchSavePackage(toUpdate, toCreate []*model.PackageInfo, toDelete []string) error {
	var err error
	tx := d.DB.Begin()
	defer func() {
		if err != nil {
			if errRollback := tx.Rollback().Error; errRollback != nil {
				log.Error("tx.Rollback() error(%+v)", err)
				return
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("tx.Commit() error(%v)", err)
			return
		}
	}()
	//delete
	for _, v := range toDelete {
		if err = tx.Model(&model.PackageInfo{}).Where("uuid=?", v).Updates(map[string]interface{}{"is_deleted": 1}).Error; err != nil {
			return err
		}
	}
	//update
	for _, v := range toUpdate {
		if err = tx.Model(&model.PackageInfo{}).Where("uuid=?", v.UUID).Save(v).Error; err != nil {
			return err
		}
	}
	//create
	for _, v := range toCreate {
		if err = tx.Model(&model.PackageInfo{}).Create(v).Error; err != nil {
			return err
		}
	}
	return nil
}
