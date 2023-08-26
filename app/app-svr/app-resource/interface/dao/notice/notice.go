package notice

import (
	"context"
	"fmt"
	"time"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/component"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/notice"

	"github.com/pkg/errors"
)

const (
	_getSQL            = `SELECT id,plat,title,content,build,conditions,area,url,type,ef_time,ex_time FROM notice WHERE state=1 AND ef_time<? AND ex_time>? ORDER BY mtime DESC`
	_getPackagePushMsg = `SELECT id,title,text,package_url,popup_title,app_name,app_current_version,app_update_time,app_developer,permission_purpose,privacy_policy,crowed_name,crowed_business,icon,apk_size,jump_to_app_store from package_push WHERE is_deleted=0 AND stime<? AND etime>?`
)

// Dao is notice dao.
type Dao struct {
	db  *sql.DB
	get *sql.Stmt
}

// New new a notice dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db: component.GlobalDB,
	}
	d.get = d.db.Prepared(_getSQL)
	return
}

// GetAll get all notice data.
func (d *Dao) All(ctx context.Context, now time.Time) (res []*notice.Notice, err error) {
	rows, err := d.get.Query(ctx, now, now)
	if err != nil {
		log.Error("query error (%v)", err)
		return
	}
	defer rows.Close()
	res = []*notice.Notice{}
	for rows.Next() {
		b := &notice.Notice{}
		if err = rows.Scan(&b.ID, &b.Plat, &b.Title, &b.Content, &b.Build, &b.Condition, &b.Area, &b.URI, &b.Type, &b.Start, &b.End); err != nil {
			log.Error("rows.Scan err (%v)", err)
			return nil, err

		}
		res = append(res, b)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

func (d *Dao) PackagePushList(ctx context.Context) (map[string]*notice.PushDetail, error) {
	now := time.Now().Unix()
	rows, err := d.db.Query(ctx, _getPackagePushMsg, now, now)
	if err != nil {
		return nil, errors.Wrapf(err, "PackagePushList query error")
	}
	defer rows.Close()
	pushCrowedMap := make(map[string]*notice.PushDetail)
	for rows.Next() {
		detail := &notice.PushDetail{}
		if err = rows.Scan(&detail.ID, &detail.Title, &detail.Text, &detail.PackageUrl, &detail.PopupTitle, &detail.AppName,
			&detail.AppCurrentVersion, &detail.AppUpdateTime, &detail.AppDeveloper, &detail.PermissionPurpose, &detail.PrivacyPolicy,
			&detail.CrowedName, &detail.CrowedBusiness, &detail.ICON, &detail.ApkSize, &detail.JumpToAppStore); err != nil {
			return nil, errors.Wrapf(err, "PackagePushList rows.Scan error")
		}
		key := fmt.Sprintf("%s:%s", detail.CrowedBusiness, detail.CrowedName)
		if _, ok := pushCrowedMap[key]; ok {
			continue
		}
		pushCrowedMap[key] = detail
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "PackagePushList rows.Err() error")
	}
	return pushCrowedMap, nil
}

// Close close memcache resource.
func (dao *Dao) Close() {
	if dao.db != nil {
		dao.db.Close()
	}
}
