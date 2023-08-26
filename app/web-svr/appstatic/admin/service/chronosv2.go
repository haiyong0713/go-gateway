package service

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/appstatic/admin/model"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	_create = "create"
	_update = "update"
	_rank   = "rank"
	_delete = "delete"
)

func (s *Service) ShowAppInfoList(ctx context.Context) ([]*model.AppInfo, error) {
	infoList, err := s.dao.ShowAppInfoList(ctx)
	if err != nil {
		return nil, err
	}
	return infoList, nil
}

func (s *Service) ShowAppInfoDetail(ctx context.Context, appKey string) (*model.AppInfo, error) {
	info, err := s.dao.ShowAppInfoDetail(ctx, appKey)
	if err != nil {
		return nil, errors.Wrapf(err, "show app detail appkey(%s)", appKey)
	}
	return info, nil
}

func (s *Service) SaveAppInfo(ctx context.Context, info *model.AppInfo) error {
	if info.ID > 0 {
		if err := s.dao.UpdateAppInfo(ctx, info); err != nil {
			return errors.Wrapf(err, "update info(%+v)", info)
		}
		return nil
	}
	if err := s.dao.CreateAppInfo(ctx, info); err != nil {
		return errors.Wrapf(err, "create info(%+v)", info)
	}
	return nil
}

func (s *Service) DeleteAppInfo(ctx context.Context, appKey string) error {
	if err := s.dao.DeleteAppInfo(ctx, appKey); err != nil {
		return errors.Wrapf(err, "update appkey(%s)", appKey)
	}
	return nil
}

func (s *Service) SaveServiceInfo(ctx context.Context, info *model.ServiceInfo) error {
	if info.ID > 0 {
		if err := s.dao.UpdateServiceInfo(ctx, info); err != nil {
			return errors.Wrapf(err, "update info(%+v)", info)
		}
		return nil
	}
	if err := s.dao.CreateServiceInfo(ctx, info); err != nil {
		return errors.Wrapf(err, "create info(%+v)", info)
	}
	return nil
}

func (s *Service) DeleteServiceInfo(ctx context.Context, serviceKey string) error {
	if err := s.dao.DeleteServiceInfo(ctx, serviceKey); err != nil {
		return errors.Wrapf(err, "delete servicekey(%s)", serviceKey)
	}
	return nil
}

func (s *Service) ShowServiceInfoList(ctx context.Context) ([]*model.ServiceInfo, error) {
	infoList, err := s.dao.ShowServiceInfoList(ctx)
	if err != nil {
		return nil, err
	}
	return infoList, nil
}

func (s *Service) ShowServiceInfoDetail(ctx context.Context, serviceKey string) (*model.ServiceInfo, error) {
	info, err := s.dao.ShowServiceInfoDetail(ctx, serviceKey)
	if err != nil {
		return nil, errors.Wrapf(err, "show service detail servicekey(%s)", serviceKey)
	}
	return info, nil
}

func (s *Service) SavePackageToAudit(ctx context.Context, packageInfo *model.PackageInfo, username string) (*model.PackageOpReply, error) {
	preInfo, err := s.getAuditPreInfoByID(ctx, packageInfo.ID)
	if err != nil {
		//只有create操作允许找不到version
		skipNothingFoundError := func() bool {
			if packageInfo.ID != 0 {
				return false
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return false
			}
			return true
		}()
		if !skipNothingFoundError {
			return nil, errors.Wrapf(err, "getAuditPreInfoByID info(%+v)", packageInfo)
		}
	}
	if v, ok := preInfo[packageInfo.ID]; ok && v != nil && v.Version != packageInfo.Version {
		log.Warn("version in the mysql is inconsistent with request")
		packageInfo.Version = v.Version
	}
	action := _update
	if packageInfo.ID == 0 {
		action = _create
	}
	packageAuditInfo := genPackageAudit(packageInfo, username, packageInfo.AppKey, packageInfo.ServiceKey, action)
	if packageAuditInfo == nil {
		log.Error("packageAuditInfo lost")
		return nil, ecode.NothingFound
	}
	auditID, err := s.dao.CreatePackageAudit(ctx, packageAuditInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "CreatePackageAudit info(%+v)", packageAuditInfo)
	}
	return &model.PackageOpReply{AuditID: auditID, Behavior: packageAuditInfo.Behavior}, nil
}

func (s *Service) DeletePackageToAudit(ctx context.Context, id int64, username string) (*model.PackageOpReply, error) {
	preInfo, err := s.getAuditPreInfoByID(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "getAuditPreInfoByID id (%d)", id)
	}
	pi, ok := preInfo[id]
	if !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "failed to get pre info in db(%d)", id)
	}
	deleteInfo := &model.DeleteInfo{
		ID:      pi.ID,
		Version: pi.Version,
	}
	packageAuditInfo := genPackageAudit(deleteInfo, username, pi.AppKey, pi.ServiceKey, _delete)
	if packageAuditInfo == nil {
		log.Error("packageAuditInfo lost")
		return nil, errors.Wrapf(err, "packageAuditInfo lost")
	}
	auditID, err := s.dao.CreatePackageAudit(ctx, packageAuditInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "CreatePackageAudit info(%+v)", packageAuditInfo)
	}
	return &model.PackageOpReply{AuditID: auditID, Behavior: packageAuditInfo.Behavior}, nil
}

func (s *Service) RankPackageToAudit(ctx context.Context, packageIDRank map[int64]int64, username string) (*model.PackageOpReply, error) {
	var (
		ids        []int64
		appKey     string
		serviceKey string
	)
	for id := range packageIDRank {
		ids = append(ids, id)
	}
	preInfo, err := s.getAuditPreInfoByID(ctx, ids...)
	if err != nil {
		return nil, errors.Wrapf(err, "getAuditPreInfoByID ids (%+v)", ids)
	}
	rankInfos := &model.RankInfos{}
	for _, v := range preInfo {
		if v == nil {
			continue
		}
		if appKey == "" && serviceKey == "" { //一次排序操作只属于一个app_key和service_key
			appKey = v.AppKey
			serviceKey = v.ServiceKey
		}
		rankInfos.Infos = append(rankInfos.Infos, &model.RankInfo{
			ID:      v.ID,
			Version: v.Version,
			Rank:    packageIDRank[v.ID],
		})
	}
	packageAuditInfo := genPackageAudit(rankInfos, username, appKey, serviceKey, _rank)
	if packageAuditInfo == nil {
		return nil, errors.Wrapf(err, "packageAuditInfo lost")
	}
	auditID, err := s.dao.CreatePackageAudit(ctx, packageAuditInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "CreatePackageAudit info(%+v)", packageAuditInfo)
	}
	return &model.PackageOpReply{AuditID: auditID, Behavior: packageAuditInfo.Behavior}, nil
}

func (s *Service) getAuditPreInfoByID(ctx context.Context, id ...int64) (map[int64]*model.PrePackageInfo, error) {
	res, err := s.dao.BatchGetAuditPrePackageInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Service) AuditApproved(ctx context.Context, auditID int64) error {
	//先拿到该auditID下的所有信息
	auditInfo, err := s.dao.GetPackageAuditInfo(ctx, auditID)
	if err != nil {
		return errors.Wrapf(err, "GetPackageAuditInfo info(%+v)", auditInfo)
	}
	behaviorList := &model.PackageAuditBehaviorList{}
	if err := json.Unmarshal([]byte(auditInfo.Behavior), behaviorList); err != nil {
		return errors.Wrapf(err, "json.Unmarshal")
	}
	packageInfoInOrder := make(map[int64]*model.PackageInfo)
	for k, v := range behaviorList.PackageBehavior {
		switch v.Action {
		case _create:
			if err := s.createPackage(ctx, v.Update, v.Version, auditID); err != nil {
				return err
			}
		case _update:
			if err := s.updatePackage(ctx, v.Update, v.Version, auditID); err != nil {
				return err
			}
		case _delete:
			if err := s.deletePackage(ctx, k, v.Version, auditID); err != nil {
				return err
			}
		case _rank:
			idInt, err := strconv.ParseInt(k, 10, 64)
			if err != nil {
				return errors.Wrapf(err, "strconv.ParseInt")
			}
			if !s.legalPackageVersion(ctx, idInt, v.Version) {
				return errors.Wrapf(ecode.AccessDenied, "rank version denied")
			}
			packageInfoInOrder[idInt] = &model.PackageInfo{Version: v.Version, Rank: v.RankResult}
		default:
			log.Warn("unknown action")
		}
	}
	if len(packageInfoInOrder) > 0 {
		if err := s.rankPackage(ctx, packageInfoInOrder, auditID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) createPackage(ctx context.Context, info *model.PackageInfo, version, auditID int64) error {
	if info == nil || version != 0 {
		return ecode.Error(ecode.NothingFound, "新建package失败")
	}
	if err := s.dao.CreatePackageAndAftermath(ctx, info, auditID); err != nil {
		return err
	}
	return nil
}

func (s *Service) updatePackage(ctx context.Context, info *model.PackageInfo, version, audiID int64) error {
	if info == nil {
		return ecode.Error(ecode.NothingFound, "更新package失败")
	}
	if !s.legalPackageVersion(ctx, info.ID, version) {
		return ecode.Error(ecode.AccessDenied, "version denied")
	}
	info.VersionAdder(1)
	if err := s.dao.UpdatePackageAndAftermath(ctx, info, audiID); err != nil {
		return err
	}
	return nil
}

func (s *Service) deletePackage(ctx context.Context, id string, version, auditID int64) error {
	idInt, _ := strconv.ParseInt(id, 10, 64)
	if !s.legalPackageVersion(ctx, idInt, version) {
		return ecode.Error(ecode.AccessDenied, "version denied")
	}
	if err := s.dao.DeletePackageAndAftermath(ctx, idInt, auditID, version); err != nil {
		return err
	}
	return nil
}

func (s *Service) rankPackage(ctx context.Context, packageInfoInOrder map[int64]*model.PackageInfo, auditID int64) error {
	if err := s.dao.RankPackageAndAftermath(ctx, packageInfoInOrder, auditID); err != nil {
		return err
	}
	return nil
}

func (s *Service) legalPackageVersion(ctx context.Context, id int64, version int64) bool {
	preInfo, err := s.dao.BatchGetAuditPrePackageInfo(ctx, []int64{id})
	if err != nil {
		log.Error("legalPackageVersion s.dao.BatchGetPackageVersion error(%+v) id(%d)", err, id)
		return false
	}
	pi, ok := preInfo[id]
	if !ok {
		log.Error("Failed to fetch version when update")
		return false
	}
	if pi.Version > version {
		log.Warn("current version is bigger than audit version, update forbidden!")
		return false
	}
	return true
}

func (s *Service) AuditReject(ctx context.Context, auditID int64) error {
	if err := s.dao.AuditReject(ctx, auditID); err != nil {
		return err
	}
	return nil
}

func (s *Service) AuditList(ctx context.Context, appKey, serviceKey string) ([]*model.PackageAudit, error) {
	list, err := s.dao.AuditList(ctx, appKey, serviceKey)
	if err != nil {
		return nil, errors.Wrapf(err, "info(%s, %s)", appKey, serviceKey)
	}
	return list, nil
}

func (s *Service) ShowPackageInfoList(ctx context.Context, appKey, serviceKey string) ([]*model.PackageInfo, error) {
	infos, err := s.dao.ShowPackageInfoList(ctx, appKey, serviceKey)
	if err != nil {
		return nil, errors.Wrapf(err, "info(%s, %s)", appKey, serviceKey)
	}
	return infos, nil
}

func (s *Service) ShowPackageInfoDetail(ctx context.Context, uuid string) (*model.PackageInfo, error) {
	info, err := s.dao.ShowPackageInfoDetail(ctx, uuid)
	if err != nil {
		return nil, errors.Wrapf(err, "uuid(%s)", uuid)
	}
	return info, nil
}

func genPackageAudit(behavior model.BehaviorListHandler, username, appKey, serviceKey, action string) *model.PackageAudit {
	as, err := json.Marshal(behavior.GenBehaviorList(action))
	if err != nil {
		log.Error("genPackageAudit json.Marshal error(%+v)", err)
		return nil
	}
	return &model.PackageAudit{
		Operator:   username,
		AppKey:     appKey,
		ServiceKey: serviceKey,
		Behavior:   string(as),
	}
}

// BatchSavePackage packages为前端全量提交的包
func (s *Service) BatchSavePackage(ctx context.Context, packages []*model.PackageInfo, appKey, serviceKey string) error {
	packagesFromDB, err := s.dao.ShowPackageInfoList(ctx, appKey, serviceKey)
	if err != nil {
		return errors.Wrapf(err, "Failed to fetch packages from db appkey(%s), servicekey(%s)", appKey, serviceKey)
	}
	var packagesFromDBMap = make(map[string]struct{})
	for _, v := range packagesFromDB {
		packagesFromDBMap[v.UUID] = struct{}{}
	}
	var (
		toUpdate []*model.PackageInfo
		toDelete []string
		toCreate []*model.PackageInfo
	)
	for _, v := range packages {
		if _, ok := packagesFromDBMap[v.UUID]; ok {
			//新传入的packages在db中存在，update
			toUpdate = append(toUpdate, v)
			delete(packagesFromDBMap, v.UUID)
			continue
		}
		//新传入的packages在db中不存在，create
		toCreate = append(toCreate, v)
	}
	//packagesFromDBMap残留的是需要删除的
	for uuid := range packagesFromDBMap {
		toDelete = append(toDelete, uuid)
	}
	return s.dao.BatchSavePackage(toUpdate, toCreate, toDelete)
}
