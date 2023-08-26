package fawkes

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	xsql "go-common/library/database/sql"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/fawkes/service/model/mod"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/pkg/errors"
)

const _limit = 1000

func (d *Dao) ModPoolList(ctx context.Context, appKey string) ([]*mod.Pool, error) {
	const _poolListSQL = "SELECT id,app_key,name,remark,module_count_limit,module_size_limit,module_count,module_size FROM mod_pool WHERE app_key=? ORDER BY id DESC"
	rows, err := d.db.Query(ctx, _poolListSQL, appKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mod.Pool
	for rows.Next() {
		p := &mod.Pool{}
		if err = rows.Scan(&p.ID, &p.AppKey, &p.Name, &p.Remark, &p.ModuleCountLimit, &p.ModuleSizeLimit, &p.ModuleCount, &p.ModuleSize); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModPoolByPoolIDs(ctx context.Context, poolIDs []int64) (map[int64]*mod.Pool, error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, poolID := range poolIDs {
		sqls = append(sqls, "?")
		args = append(args, poolID)
	}
	const _poolListByPoolIDsSQL = "SELECT id,app_key,name,remark,module_count_limit,module_size_limit,module_count,module_size FROM mod_pool WHERE id IN (%s) ORDER BY id DESC"
	rows, err := d.db.Query(ctx, fmt.Sprintf(_poolListByPoolIDsSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[int64]*mod.Pool{}
	for rows.Next() {
		p := &mod.Pool{}
		if err = rows.Scan(&p.ID, &p.AppKey, &p.Name, &p.Remark, &p.ModuleCountLimit, &p.ModuleSizeLimit, &p.ModuleCount, &p.ModuleSize); err != nil {
			return nil, err
		}
		res[p.ID] = p
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModModuleList(ctx context.Context, poolID int64) ([]*mod.Module, error) {
	const _moduleListSQL = "SELECT id,pool_id,name,remark,compress,is_wifi,state,deleted,zip_check FROM mod_module WHERE pool_id=? AND deleted=0 ORDER BY id DESC"
	rows, err := d.db.Query(ctx, _moduleListSQL, poolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mod.Module
	for rows.Next() {
		m := &mod.Module{}
		if err = rows.Scan(&m.ID, &m.PoolID, &m.Name, &m.Remark, &m.Compress, &m.IsWifi, &m.State, &m.Deleted, &m.ZipCheck); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModVersionList(ctx context.Context, moduleID int64, env mod.Env, offset, limit int64) ([]*mod.Version, error) {
	const _versionListSQL = "SELECT id,module_id,env,version,remark,from_ver_id,released,release_time,state FROM mod_version WHERE module_id=? AND env=? ORDER BY version DESC LIMIT ?,?"
	rows, err := d.db.Query(ctx, _versionListSQL, moduleID, env, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mod.Version
	for rows.Next() {
		v := &mod.Version{}
		if err = rows.Scan(&v.ID, &v.ModuleID, &v.Env, &v.Version, &v.Remark, &v.FromVerID, &v.Released, &v.ReleaseTime, &v.State); err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModVersionList2(ctx context.Context, moduleID int64, env mod.Env, versions []int64) ([]*mod.Version, error) {
	const _versionListSQL = "SELECT id,module_id,env,version,remark,from_ver_id,released,release_time,state FROM mod_version WHERE module_id=? AND env=? AND version IN (%s)"
	var res []*mod.Version
	if len(versions) == 0 {
		return res, nil
	}
	var (
		sqls []string
		args []interface{}
	)
	args = append(args, moduleID, env)
	for _, v := range versions {
		sqls = append(sqls, "?")
		args = append(args, v)
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_versionListSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		v := &mod.Version{}
		if err = rows.Scan(&v.ID, &v.ModuleID, &v.Env, &v.Version, &v.Remark, &v.FromVerID, &v.Released, &v.ReleaseTime, &v.State); err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModVersionCount(ctx context.Context, moduleID int64, env mod.Env) (int64, error) {
	const _versionCountSQL = "SELECT COUNT(1) FROM mod_version WHERE module_id=? AND env=?"
	var count int64
	row := d.db.QueryRow(ctx, _versionCountSQL, moduleID, env)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (d *Dao) ModVersionFile(ctx context.Context, versionIDs []int64) (map[int64][]*mod.File, error) {
	const _versionFileSQL = "SELECT version_id,id,name,size,md5,url,is_patch,from_ver,ctime FROM mod_file WHERE version_id IN (%s) ORDER BY id desc"
	var (
		sqls []string
		args []interface{}
	)
	for _, id := range versionIDs {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_versionFileSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[int64][]*mod.File{}
	for rows.Next() {
		var versionID int64
		f := &mod.File{}
		if err = rows.Scan(&versionID, &f.ID, &f.Name, &f.Size, &f.Md5, &f.URL, &f.IsPatch, &f.FromVer, &f.Ctime); err != nil {
			return nil, err
		}
		f.SetURL(d.c.Mod.ModCDN)
		res[versionID] = append(res[versionID], f)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModFile(ctx context.Context, versionID int64) (*mod.File, error) {
	const _fileSQL = "SELECT id,name,size,md5,url,ctime FROM mod_file WHERE version_id=? AND is_patch=0 AND from_ver=0"
	rows := d.db.QueryRow(ctx, _fileSQL, versionID)
	f := &mod.File{}
	if err := rows.Scan(&f.ID, &f.Name, &f.Size, &f.Md5, &f.URL, &f.Ctime); err != nil {
		return nil, err
	}
	f.SetURL(d.c.Mod.ModCDN)
	return f, nil
}

func (d *Dao) ModPatchList(ctx context.Context, versionID int64) ([]*mod.Patch, error) {
	const _patchListSQL = "SELECT id,name,size,md5,url,from_ver,ctime FROM mod_file WHERE version_id=? AND is_patch=1 ORDER BY from_ver desc"
	rows, err := d.db.Query(ctx, _patchListSQL, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mod.Patch
	for rows.Next() {
		p := &mod.Patch{}
		if err = rows.Scan(&p.ID, &p.Name, &p.Size, &p.Md5, &p.URL, &p.FromVer, &p.Ctime); err != nil {
			return nil, err
		}
		p.SetURL(d.c.Mod.ModCDN)
		res = append(res, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModVersionAdd(ctx context.Context, moduleID int64, env mod.Env, remark string, file *mod.File) (*mod.Version, error) {
	const (
		_maxVersionSQL    = "SELECT MAX(version) FROM mod_version WHERE module_id=?"
		_insertVersionSQL = "INSERT INTO mod_version (module_id,env,version,remark,state) VALUES (?,?,?,?,?)"
		_insertFileSQL    = "INSERT INTO mod_file (version_id,name,content_type,size,md5,url,is_patch,from_ver) VALUES (?,?,?,?,?,?,?,?)"
	)
	var maxVersion sql.NullInt64
	row := d.db.QueryRow(ctx, _maxVersionSQL, moduleID)
	if err := row.Scan(&maxVersion); err != nil && err != xsql.ErrNoRows {
		return nil, err
	}
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback error:%+v", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit error:%+v", err)
		}
	}()
	version := maxVersion.Int64 + 1
	res, err := tx.Exec(_insertVersionSQL, moduleID, env, version, remark, mod.VersionProcessing)
	if err != nil {
		return nil, err
	}
	versionID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	if res, err = tx.Exec(_insertFileSQL, versionID, file.Name, file.ContentType, file.Size, file.Md5, file.URL, file.IsPatch, file.FromVer); err != nil {
		return nil, err
	}
	if file.ID, err = res.LastInsertId(); err != nil {
		return nil, err
	}
	file.SetURL(d.c.Mod.ModCDN)
	return &mod.Version{
		ID:      versionID,
		Version: version,
		File:    file,
	}, nil
}

func (d *Dao) ModVersionRelease(ctx context.Context, versionID int64, released bool, releaseTime xtime.Time) error {
	const _moduleReleaseSQL = "UPDATE mod_version SET released=?,release_time=? WHERE id=? AND state=?"
	_, err := d.db.Exec(ctx, _moduleReleaseSQL, released, releaseTime, versionID, mod.VersionSucceeded)
	return err
}

func (d *Dao) ModVersionByID(ctx context.Context, versionID int64) (*mod.Version, error) {
	const _versionByIDSQL = "SELECT id,module_id,env,version,remark,from_ver_id,released,release_time,state FROM mod_version WHERE id=?"
	row := d.db.QueryRow(ctx, _versionByIDSQL, versionID)
	v := &mod.Version{}
	if err := row.Scan(&v.ID, &v.ModuleID, &v.Env, &v.Version, &v.Remark, &v.FromVerID, &v.Released, &v.ReleaseTime, &v.State); err != nil {
		return nil, err
	}
	return v, nil
}

func (d *Dao) ModProdVersionExist(ctx context.Context, moduleID int64, version int64) (int64, error) {
	const _versionExistSQL = "SELECT id FROM mod_version WHERE module_id=? AND env=? AND version=?"
	row := d.db.QueryRow(ctx, _versionExistSQL, moduleID, mod.EnvProd, version)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (d *Dao) ModVersionPush(ctx context.Context, v *mod.Version) (int64, error) {
	const _insertVersionSQL = "INSERT INTO mod_version (module_id,env,version,remark,from_ver_id,state) VALUES (?,?,?,?,?,?)"
	res, err := d.db.Exec(ctx, _insertVersionSQL, v.ModuleID, mod.EnvProd, v.Version, v.Remark, v.ID, v.State)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) ModVersionPushWithConfig(ctx context.Context, v *mod.Version, config *mod.Config, gray *mod.Gray) (versionID int64, err error) {
	const (
		_inSQL        = "INSERT INTO mod_version (module_id,env,version,remark,from_ver_id,state) VALUES (?,?,?,?,?,?)"
		_addConfigSQL = "INSERT INTO mod_version_config (version_id,priority,app_ver,sys_ver,stime,etime) VALUES (?,?,?,?,?,?)"
		_addGraySQL   = "INSERT INTO mod_version_gray (version_id,strategy,salt,bucket_start,bucket_end,whitelist,whitelist_url,manual_download) VALUES (?,?,?,?,?,?,?,?)"
	)
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback error:%+v", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit error:%+v", err)
		}
	}()
	res, err := tx.Exec(_inSQL, v.ModuleID, mod.EnvProd, v.Version, v.Remark, v.ID, v.State)
	if err != nil {
		return 0, err
	}
	if versionID, err = res.LastInsertId(); err != nil {
		return 0, err
	}
	if config != nil {
		if _, err := tx.Exec(_addConfigSQL, versionID, config.Priority, config.AppVer, config.SysVer, config.Stime, config.Etime); err != nil {
			return 0, err
		}
	}
	if gray != nil {
		if _, err := tx.Exec(_addGraySQL, versionID, gray.Strategy, gray.Salt, gray.BucketStart, gray.BucketEnd, gray.Whitelist, gray.WhitelistURL, gray.ManualDownload); err != nil {
			return 0, err
		}
	}
	return versionID, nil
}

func (d *Dao) ModVersionConfig(ctx context.Context, versionID int64) (*mod.Config, error) {
	const _versionConfigSQL = "SELECT id,version_id,priority,app_ver,sys_ver,stime,etime FROM mod_version_config WHERE version_id=?"
	row := d.db.QueryRow(ctx, _versionConfigSQL, versionID)
	c := &mod.Config{}
	if err := row.Scan(&c.ID, &c.VersionID, &c.Priority, &c.AppVer, &c.SysVer, &c.Stime, &c.Etime); err != nil {
		return nil, err
	}
	return c, nil
}

func (d *Dao) ModVersionConfigAdd(ctx context.Context, config *mod.Config) (int64, error) {
	const (
		_addConfigSQL = "INSERT INTO mod_version_config (version_id,priority,app_ver,sys_ver,stime,etime) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE priority=VALUES(priority),app_ver=VALUES(app_ver),sys_ver=VALUES(sys_ver),stime=VALUES(stime),etime=VALUES(etime)"
	)
	res, err := d.db.Exec(ctx, _addConfigSQL, config.VersionID, config.Priority, config.AppVer, config.SysVer, config.Stime, config.Etime)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) ModVersionConfigApplyAdd(ctx context.Context, config *mod.Config) (int64, error) {
	const (
		_addConfigSQL = "INSERT INTO mod_version_config_apply (version_id,priority,app_ver,sys_ver,stime,etime,state) VALUES (?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE priority=VALUES(priority),app_ver=VALUES(app_ver),sys_ver=VALUES(sys_ver),stime=VALUES(stime),etime=VALUES(etime),state=VALUES(state)"
	)
	res, err := d.db.Exec(ctx, _addConfigSQL, config.VersionID, config.Priority, config.AppVer, config.SysVer, config.Stime, config.Etime, mod.ApplyStateChecking)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) ModVersionGray(ctx context.Context, versionID int64) (*mod.Gray, error) {
	const _versionGraySQL = "SELECT id,version_id,strategy,salt,bucket_start,bucket_end,whitelist,whitelist_url,manual_download FROM mod_version_gray WHERE version_id=?"
	row := d.db.QueryRow(ctx, _versionGraySQL, versionID)
	g := &mod.Gray{}
	if err := row.Scan(&g.ID, &g.VersionID, &g.Strategy, &g.Salt, &g.BucketStart, &g.BucketEnd, &g.Whitelist, &g.WhitelistURL, &g.ManualDownload); err != nil {
		return nil, err
	}
	return g, nil
}

func (d *Dao) ModVersionGrayAdd(ctx context.Context, gray *mod.Gray) (int64, error) {
	const (
		_addGraySQL = "INSERT INTO mod_version_gray (version_id,strategy,salt,bucket_start,bucket_end,whitelist,whitelist_url,manual_download) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE strategy=VALUES(strategy),salt=VALUES(salt),bucket_start=VALUES(bucket_start),bucket_end=VALUES(bucket_end),whitelist=VALUES(whitelist),whitelist_url=VALUES(whitelist_url),manual_download=VALUES(manual_download)"
	)
	res, err := d.db.Exec(ctx, _addGraySQL, gray.VersionID, gray.Strategy, gray.Salt, gray.BucketStart, gray.BucketEnd, gray.Whitelist, gray.WhitelistURL, gray.ManualDownload)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) ModVersionGrayApplyAdd(ctx context.Context, gray *mod.Gray) (int64, error) {
	const (
		_addGraySQL = "INSERT INTO mod_version_gray_apply (version_id,strategy,salt,bucket_start,bucket_end,whitelist,whitelist_url,manual_download,state) VALUES (?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE strategy=VALUES(strategy),salt=VALUES(salt),bucket_start=VALUES(bucket_start),bucket_end=VALUES(bucket_end),whitelist=VALUES(whitelist),whitelist_url=VALUES(whitelist_url),manual_download=VALUES(manual_download),state=VALUES(state)"
	)
	res, err := d.db.Exec(ctx, _addGraySQL, gray.VersionID, gray.Strategy, gray.Salt, gray.BucketStart, gray.BucketEnd, gray.Whitelist, gray.WhitelistURL, gray.ManualDownload, mod.ApplyStateChecking)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) ModModuleDelete(ctx context.Context, moduleID int64) error {
	const _moduleDeleteSQL = "UPDATE mod_module SET deleted=1 WHERE id=? AND deleted=0"
	_, err := d.db.Exec(ctx, _moduleDeleteSQL, moduleID)
	return err
}

func (d *Dao) ModModuleState(ctx context.Context, moduleID int64, state mod.ModuleState) error {
	const _moduleStateSQL = "UPDATE mod_module SET state=? WHERE id=? AND deleted=0"
	_, err := d.db.Exec(ctx, _moduleStateSQL, state, moduleID)
	return err
}

func (d *Dao) ModModuleUpdate(ctx context.Context, moduleID int64, remark string, isWIFI, zipCheck bool, compress mod.Compress) error {
	const _moduleUpdateSQL = "UPDATE mod_module SET %s WHERE id=? AND deleted=0"
	var (
		args []interface{}
		sqls []string
	)
	if remark != "" {
		args = append(args, remark)
		sqls = append(sqls, "remark=?")
	}
	args = append(args, isWIFI, zipCheck)
	sqls = append(sqls, "is_wifi=?", "zip_check=?")
	if compress != "" {
		args = append(args, compress)
		sqls = append(sqls, "compress=?")
	}
	args = append(args, moduleID)
	_, err := d.db.Exec(ctx, fmt.Sprintf(_moduleUpdateSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(ctx, "ModuleUpdateSQL slq[%s] args[%#v] error[%#v]", fmt.Sprintf(_moduleUpdateSQL, strings.Join(sqls, ",")), args, err)
	}
	return err
}

func (d *Dao) ModModuleExist(ctx context.Context, poolID int64, name string) (int64, error) {
	const _moduleExistSQL = "SELECT id FROM mod_module WHERE pool_id=? AND name=?"
	row := d.db.QueryRow(ctx, _moduleExistSQL, poolID, name)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (d *Dao) ModModuleAdd(ctx context.Context, poolID int64, name, remark string, isWiFI bool, compress mod.Compress, state mod.ModuleState, zipCheck bool) (int64, error) {
	const _moduleAddSQL = "INSERT INTO mod_module (pool_id,name,remark,compress,state,zip_check,is_wifi) VALUES (?,?,?,?,?,?,?)"
	res, err := d.db.Exec(ctx, _moduleAddSQL, poolID, name, remark, compress, state, zipCheck, isWiFI)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) ModPoolExist(ctx context.Context, appKey, name string) (int64, error) {
	const _poolExistSQL = "SELECT id FROM mod_pool WHERE app_key=? AND name=?"
	row := d.db.QueryRow(ctx, _poolExistSQL, appKey, name)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (d *Dao) ModPoolByName(ctx context.Context, appKey, name string) (*mod.Pool, error) {
	const _selSQL = "SELECT id,app_key,name,remark,module_count_limit,module_size_limit,module_count,module_size FROM mod_pool WHERE app_key=? AND name=?"
	row := d.db.QueryRow(ctx, _selSQL, appKey, name)
	p := &mod.Pool{}
	if err := row.Scan(&p.ID, &p.AppKey, &p.Name, &p.Remark, &p.ModuleCountLimit, &p.ModuleSizeLimit, &p.ModuleCount, &p.ModuleSize); err != nil {
		return nil, err
	}
	return p, nil
}

func (d *Dao) ModPoolAdd(ctx context.Context, appKey, name, remark string, moduleCountLimit, moduleSizeLimit int64) (int64, error) {
	const _moduleAddSQL = "INSERT INTO mod_pool (app_key,name,remark,module_count_limit,module_size_limit) VALUES (?,?,?,?,?)"
	res, err := d.db.Exec(ctx, _moduleAddSQL, appKey, name, remark, moduleCountLimit, moduleSizeLimit)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) ModPoolUpdate(ctx context.Context, poolID int64, moduleCountLimit, moduleSizeLimit int64) error {
	const _moduleUpdateSQL = "UPDATE mod_pool SET module_count_limit=?,module_size_limit=? WHERE id=?"
	_, err := d.db.Exec(ctx, _moduleUpdateSQL, moduleCountLimit, moduleSizeLimit, poolID)
	return err
}

func (d *Dao) ModPermissionList(ctx context.Context, poolID int64) ([]*mod.Permission, error) {
	const _permissionListSQL = "SELECT id,username,pool_id,permission FROM mod_permission WHERE pool_id=? AND deleted=0"
	rows, err := d.db.Query(ctx, _permissionListSQL, poolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mod.Permission
	for rows.Next() {
		p := &mod.Permission{}
		if err = rows.Scan(&p.ID, &p.Username, &p.PoolID, &p.Permission); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModPermissionAdd(ctx context.Context, username string, poolID int64, permission mod.Perm) (int64, error) {
	const (
		_addSQL = "INSERT INTO mod_permission (username,pool_id,permission,deleted) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE permission=VALUES(permission),deleted=VALUES(deleted)"
		_delSQL = "UPDATE mod_permission SET deleted=1 WHERE username=? AND pool_id=?"
	)
	if permission == mod.PermNone {
		_, err := d.db.Exec(ctx, _delSQL, username, poolID)
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	res, err := d.db.Exec(ctx, _addSQL, username, poolID, permission, false)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) ModPermissionDelete(ctx context.Context, permissionID int64) error {
	const _permissionDeleteSQL = "UPDATE mod_permission SET deleted=1 WHERE id=?"
	_, err := d.db.Exec(ctx, _permissionDeleteSQL, permissionID)
	return err
}

func (d *Dao) ModPermissionByUsername(ctx context.Context, username string, poolID int64) (mod.Perm, error) {
	const _permissionByUsernameSQL = "SELECT permission FROM mod_permission WHERE username=? AND pool_id=? AND deleted=0"
	row := d.db.QueryRow(ctx, _permissionByUsernameSQL, username, poolID)
	var permission mod.Perm
	if err := row.Scan(&permission); err != nil {
		return "", err
	}
	return permission, nil
}

func (d *Dao) ModModuleByID(ctx context.Context, moduleID int64) (*mod.Module, error) {
	const _moduleByIDSQL = "SELECT id,pool_id,name,remark,compress,is_wifi,state,zip_check FROM mod_module WHERE id=? AND deleted=0"
	row := d.db.QueryRow(ctx, _moduleByIDSQL, moduleID)
	m := &mod.Module{}
	if err := row.Scan(&m.ID, &m.PoolID, &m.Name, &m.Remark, &m.Compress, &m.IsWifi, &m.State, &m.ZipCheck); err != nil {
		return nil, err
	}
	return m, nil
}

func (d *Dao) ModModuleByName(ctx context.Context, poolID int64, name string) (*mod.Module, error) {
	const _moduleByIDSQL = "SELECT id,pool_id,name,remark,compress,is_wifi,state,zip_check FROM mod_module WHERE pool_id=? AND name=? AND deleted=0"
	row := d.db.QueryRow(ctx, _moduleByIDSQL, poolID, name)
	m := &mod.Module{}
	if err := row.Scan(&m.ID, &m.PoolID, &m.Name, &m.Remark, &m.Compress, &m.IsWifi, &m.State, &m.ZipCheck); err != nil {
		return nil, err
	}
	return m, nil
}

func (d *Dao) ModPoolByID(ctx context.Context, poolID int64) (*mod.Pool, error) {
	const _poolByIDSQL = "SELECT id,app_key,name,remark,module_count_limit,module_size_limit,module_count,module_size FROM mod_pool WHERE id=? AND deleted=0"
	row := d.db.QueryRow(ctx, _poolByIDSQL, poolID)
	p := &mod.Pool{}
	if err := row.Scan(&p.ID, &p.AppKey, &p.Name, &p.Remark, &p.ModuleCountLimit, &p.ModuleSizeLimit, &p.ModuleCount, &p.ModuleSize); err != nil {
		return nil, err
	}
	return p, nil
}

func (d *Dao) ModPermissionByID(ctx context.Context, permissionID int64) (*mod.Permission, error) {
	const _permissionByIDSQL = "SELECT id,username,pool_id,permission FROM mod_permission WHERE id=? AND deleted=0"
	row := d.db.QueryRow(ctx, _permissionByIDSQL, permissionID)
	p := &mod.Permission{}
	if err := row.Scan(&p.ID, &p.Username, &p.PoolID, &p.Permission); err != nil {
		return nil, err
	}
	return p, nil
}

func (d *Dao) ModBusAppKeyList(ctx context.Context) ([]string, error) {
	const _appKeyList = "SELECT DISTINCT app_key FROM mod_pool WHERE deleted=0 ORDER BY id DESC"
	rows, err := d.db.Query(ctx, _appKeyList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []string
	for rows.Next() {
		var appKey string
		if err := rows.Scan(&appKey); err != nil {
			return nil, err
		}
		res = append(res, appKey)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// ModBusPoolList appKey-poolName-mod.BusPool
func (d *Dao) ModBusPoolList(ctx context.Context) (map[string]map[string]*mod.BusPool, error) {
	const _poolList = "SELECT id,app_key,name,remark FROM mod_pool WHERE deleted=0 ORDER BY id DESC"
	rows, err := d.db.Query(ctx, _poolList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[string]map[string]*mod.BusPool{}
	for rows.Next() {
		p := &mod.BusPool{}
		if err := rows.Scan(&p.ID, &p.AppKey, &p.Name, &p.Remark); err != nil {
			return nil, err
		}
		v, ok := res[p.AppKey]
		if !ok {
			v = map[string]*mod.BusPool{}
			res[p.AppKey] = v
		}
		v[p.Name] = p
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// ModBusModuleList return pool-moduleName-mod.BusModule
func (d *Dao) ModBusModuleList(ctx context.Context, poolIDs []int64) (map[int64]map[string]*mod.BusModule, error) {
	const _moduleList = "SELECT id,pool_id,name,remark,compress,is_wifi,zip_check FROM mod_module WHERE pool_id IN (%s) AND state=? AND deleted=0"
	var shard int
	if len(poolIDs) < _limit {
		shard = 1
	} else {
		shard = len(poolIDs) / _limit
		if len(poolIDs)%(shard*_limit) != 0 {
			shard++
		}
	}
	idss := make([][]int64, shard)
	for i, id := range poolIDs {
		idss[i%shard] = append(idss[i%shard], id)
	}
	res := map[int64]map[string]*mod.BusModule{}
	for _, poolIDs := range idss {
		if len(poolIDs) == 0 {
			continue
		}
		var (
			sqls []string
			args []interface{}
		)
		for _, v := range poolIDs {
			sqls = append(sqls, "?")
			args = append(args, v)
		}
		args = append(args, mod.ModuleOnline)
		rows, err := d.db.Query(ctx, fmt.Sprintf(_moduleList, strings.Join(sqls, ",")), args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			m := &mod.BusModule{}
			if err := rows.Scan(&m.ID, &m.PoolID, &m.Name, &m.Remark, &m.Compress, &m.IsWifi, &m.ZipCheck); err != nil {
				return nil, err
			}
			v, ok := res[m.PoolID]
			if !ok {
				v = map[string]*mod.BusModule{}
				res[m.PoolID] = v
			}
			v[m.Name] = m
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (d *Dao) ModBusVersionListByModuleIDs(ctx context.Context, moduleIDs []int64, env mod.Env) (map[int64]map[mod.Env][]*mod.BusVersion, error) {
	const _versionList = "SELECT id,module_id,env,version,remark,from_ver_id,release_time,state,mtime FROM mod_version WHERE module_id IN (%s) AND env=? AND state=? AND released=1 ORDER BY version DESC"
	var shard int
	if len(moduleIDs) < _limit {
		shard = 1
	} else {
		shard = len(moduleIDs) / _limit
		if len(moduleIDs)%(shard*_limit) != 0 {
			shard++
		}
	}
	idss := make([][]int64, shard)
	for i, id := range moduleIDs {
		idss[i%shard] = append(idss[i%shard], id)
	}
	res := map[int64]map[mod.Env][]*mod.BusVersion{}
	for _, moduleIDs := range idss {
		if len(moduleIDs) == 0 {
			continue
		}
		var (
			sqls []string
			args []interface{}
		)
		for _, v := range moduleIDs {
			sqls = append(sqls, "?")
			args = append(args, v)
		}
		args = append(args, env, mod.VersionSucceeded)
		rows, err := d.db.Query(ctx, fmt.Sprintf(_versionList, strings.Join(sqls, ",")), args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			v := &mod.BusVersion{}
			if err := rows.Scan(&v.ID, &v.ModuleID, &v.Env, &v.Version, &v.Remark, &v.FromVerID, &v.ReleaseTime, &v.State, &v.Mtime); err != nil {
				return nil, err
			}
			val, ok := res[v.ModuleID]
			if !ok {
				val = map[mod.Env][]*mod.BusVersion{}
				res[v.ModuleID] = val
			}
			val[v.Env] = append(val[v.Env], v)
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (d *Dao) ModBusVersionListByModuleIDs2(ctx context.Context, moduleIDs []int64, env mod.Env) (map[int64]map[mod.Env][]*mod.BusVersion, error) {
	const _versionList = "SELECT id,module_id,env,version,remark,from_ver_id,release_time,state,mtime FROM mod_version WHERE module_id IN (%s) AND env=? AND state=? ORDER BY version DESC"
	var shard int
	if len(moduleIDs) < _limit {
		shard = 1
	} else {
		shard = len(moduleIDs) / _limit
		if len(moduleIDs)%(shard*_limit) != 0 {
			shard++
		}
	}
	idss := make([][]int64, shard)
	for i, id := range moduleIDs {
		idss[i%shard] = append(idss[i%shard], id)
	}
	res := map[int64]map[mod.Env][]*mod.BusVersion{}
	for _, moduleIDs := range idss {
		if len(moduleIDs) == 0 {
			continue
		}
		var (
			sqls []string
			args []interface{}
		)
		for _, v := range moduleIDs {
			sqls = append(sqls, "?")
			args = append(args, v)
		}
		args = append(args, env, mod.VersionSucceeded)
		rows, err := d.db.Query(ctx, fmt.Sprintf(_versionList, strings.Join(sqls, ",")), args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			v := &mod.BusVersion{}
			if err := rows.Scan(&v.ID, &v.ModuleID, &v.Env, &v.Version, &v.Remark, &v.FromVerID, &v.ReleaseTime, &v.State, &v.Mtime); err != nil {
				return nil, err
			}
			val, ok := res[v.ModuleID]
			if !ok {
				val = map[mod.Env][]*mod.BusVersion{}
				res[v.ModuleID] = val
			}
			val[v.Env] = append(val[v.Env], v)
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (d *Dao) ModBusVersionListByIDs(ctx context.Context, ids []int64) (map[int64]*mod.BusVersion, error) {
	const _versionList = "SELECT id,module_id,env,version,remark,from_ver_id,release_time,state,mtime FROM mod_version WHERE id IN (%s)"
	var shard int
	if len(ids) < _limit {
		shard = 1
	} else {
		shard = len(ids) / _limit
		if len(ids)%(shard*_limit) != 0 {
			shard++
		}
	}
	idss := make([][]int64, shard)
	for i, id := range ids {
		idss[i%shard] = append(idss[i%shard], id)
	}
	res := map[int64]*mod.BusVersion{}
	for _, ids := range idss {
		if len(ids) == 0 {
			continue
		}
		var (
			sqls []string
			args []interface{}
		)
		for _, v := range ids {
			sqls = append(sqls, "?")
			args = append(args, v)
		}
		rows, err := d.db.Query(ctx, fmt.Sprintf(_versionList, strings.Join(sqls, ",")), args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			v := &mod.BusVersion{}
			if err := rows.Scan(&v.ID, &v.ModuleID, &v.Env, &v.Version, &v.Remark, &v.FromVerID, &v.ReleaseTime, &v.State, &v.Mtime); err != nil {
				return nil, err
			}
			res[v.ID] = v
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (d *Dao) ModBusFileList(ctx context.Context, versionIDs []int64) (map[int64][]*mod.BusFile, error) {
	const _fileList = "SELECT id,version_id,name,content_type,size,md5,url,is_patch,from_ver FROM mod_file WHERE version_id IN (%s) ORDER BY from_ver DESC"
	var shard int
	if len(versionIDs) < _limit {
		shard = 1
	} else {
		shard = len(versionIDs) / _limit
		if len(versionIDs)%(shard*_limit) != 0 {
			shard++
		}
	}
	idss := make([][]int64, shard)
	for i, id := range versionIDs {
		idss[i%shard] = append(idss[i%shard], id)
	}
	res := map[int64][]*mod.BusFile{}
	for _, versionIDs := range idss {
		if len(versionIDs) == 0 {
			continue
		}
		var (
			sqls []string
			args []interface{}
		)
		for _, v := range versionIDs {
			sqls = append(sqls, "?")
			args = append(args, v)
		}
		if err := func() error {
			rows, err := d.db.Query(ctx, fmt.Sprintf(_fileList, strings.Join(sqls, ",")), args...)
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				f := &mod.BusFile{}
				if err := rows.Scan(&f.ID, &f.VersionID, &f.Name, &f.ContentType, &f.Size, &f.Md5, &f.URL, &f.IsPatch, &f.FromVer); err != nil {
					return err
				}
				if !f.SetURL(d.c.Mod) {
					log.Error("日志告警 file url 前缀未识别,file:%+v", f)
					continue
				}
				res[f.VersionID] = append(res[f.VersionID], f)
			}
			return rows.Err()
		}(); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (d *Dao) ModBusVersionConfigList(ctx context.Context, versionIDs []int64) (map[int64]*mod.BusVersionConfig, error) {
	const _versionConfigList = "SELECT id,version_id,priority,app_ver,sys_ver,stime,etime,scale,forbiden_device,arch,mtime FROM mod_version_config WHERE version_id IN (%s)"
	var shard int
	if len(versionIDs) < _limit {
		shard = 1
	} else {
		shard = len(versionIDs) / _limit
		if len(versionIDs)%(shard*_limit) != 0 {
			shard++
		}
	}
	idss := make([][]int64, shard)
	for i, id := range versionIDs {
		idss[i%shard] = append(idss[i%shard], id)
	}
	res := map[int64]*mod.BusVersionConfig{}
	for _, versionIDs := range idss {
		if len(versionIDs) == 0 {
			continue
		}
		var (
			sqls []string
			args []interface{}
		)
		for _, v := range versionIDs {
			sqls = append(sqls, "?")
			args = append(args, v)
		}
		rows, err := d.db.Query(ctx, fmt.Sprintf(_versionConfigList, strings.Join(sqls, ",")), args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			c := &mod.BusVersionConfig{}
			if err := rows.Scan(&c.ID, &c.VersionID, &c.Priority, &c.AppVer, &c.SysVer, &c.Stime, &c.Etime, &c.Scale, &c.ForbidenDevice, &c.Arch, &c.Mtime); err != nil {
				return nil, err
			}
			res[c.VersionID] = c
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (d *Dao) ModBusVersionGrayList(ctx context.Context, versionIDs []int64) (map[int64]*mod.BusVersionGray, error) {
	const _versionGrayList = "SELECT id,version_id,strategy,salt,bucket_start,bucket_end,whitelist,whitelist_url,manual_download,mtime FROM mod_version_gray WHERE version_id IN (%s)"
	var shard int
	if len(versionIDs) < _limit {
		shard = 1
	} else {
		shard = len(versionIDs) / _limit
		if len(versionIDs)%(shard*_limit) != 0 {
			shard++
		}
	}
	idss := make([][]int64, shard)
	for i, id := range versionIDs {
		idss[i%shard] = append(idss[i%shard], id)
	}
	res := map[int64]*mod.BusVersionGray{}
	for _, versionIDs := range idss {
		if len(versionIDs) == 0 {
			continue
		}
		var (
			sqls []string
			args []interface{}
		)
		for _, v := range versionIDs {
			sqls = append(sqls, "?")
			args = append(args, v)
		}
		rows, err := d.db.Query(ctx, fmt.Sprintf(_versionGrayList, strings.Join(sqls, ",")), args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			g := &mod.BusVersionGray{}
			if err := rows.Scan(&g.ID, &g.VersionID, &g.Strategy, &g.Salt, &g.BucketStart, &g.BucketEnd, &g.Whitelist, &g.WhitelistURL, &g.ManualDownload, &g.Mtime); err != nil {
				return nil, err
			}
			if !g.SetWhitelistURL(d.c.Mod.ModCDN) {
				log.Error("日志告警 gray whitelist url 前缀未识别,:%+v", g)
				continue
			}
			res[g.VersionID] = g
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (d *Dao) ModRoleApplyList(ctx context.Context, appKey, username string, state mod.ApplyState, offset, limit int64) ([]*mod.RoleApply, error) {
	sql := "SELECT id,app_key,pool_id,username,permission,operator,state,ctime,mtime FROM mod_role_apply WHERE app_key=?"
	var args []interface{}
	args = append(args, appKey)
	if username != "" {
		sql += " AND (username=? OR operator=?)"
		args = append(args, username)
		args = append(args, username)
	}
	if state != "" {
		sql += " AND state=?"
		args = append(args, state)
	}
	sql += " ORDER BY id DESC LIMIT ?,?"
	args = append(args, offset)
	args = append(args, limit)
	rows, err := d.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []*mod.RoleApply{}
	for rows.Next() {
		m := &mod.RoleApply{}
		if err := rows.Scan(&m.ID, &m.AppKey, &m.PoolID, &m.Username, &m.Permission, &m.Operator, &m.State, &m.Ctime, &m.Mtime); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModRoleApplyCount(ctx context.Context, appKey, username string, state mod.ApplyState) (int64, error) {
	sql := "SELECT COUNT(1) FROM mod_role_apply WHERE app_key=?"
	var args []interface{}
	args = append(args, appKey)
	if username != "" {
		sql += " AND (username=? OR operator=?)"
		args = append(args, username)
		args = append(args, username)
	}
	if state != "" {
		sql += " AND state=?"
		args = append(args, state)
	}
	var count int64
	row := d.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (d *Dao) ModRoleApplyAdd(ctx context.Context, appKey, username string, poolID int64, permission mod.Perm, operator string) (int64, error) {
	const _roleApplyAddSQL = "INSERT INTO mod_role_apply (app_key,username,pool_id,permission,operator,state) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE app_key=VALUES(app_key),permission=VALUES(permission),operator=VALUES(operator),state=VALUES(state)"
	res, err := d.db.Exec(ctx, _roleApplyAddSQL, appKey, username, poolID, permission, operator, mod.ApplyStateChecking)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) ModRoleApplyByUsernamePoolID(ctx context.Context, username string, poolID int64) (*mod.RoleApply, error) {
	const _roleApplySQL = "SELECT id,app_key,pool_id,username,permission,operator,state FROM mod_role_apply,ctime,mtime WHERE username=? AND pool_id=?"
	row := d.db.QueryRow(ctx, _roleApplySQL, username, poolID)
	res := &mod.RoleApply{}
	if err := row.Scan(&res.ID, &res.AppKey, &res.PoolID, &res.Username, &res.Permission, &res.Operator, &res.State, &res.Ctime, &res.Mtime); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModRoleOperatorList(ctx context.Context, poolID int64) ([]*mod.Permission, error) {
	const _permissionListSQL = "SELECT id,username,pool_id,permission FROM mod_permission WHERE pool_id=? AND permission=? AND deleted=0"
	rows, err := d.db.Query(ctx, _permissionListSQL, poolID, mod.PermAdmin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mod.Permission
	for rows.Next() {
		p := &mod.Permission{}
		if err = rows.Scan(&p.ID, &p.Username, &p.PoolID, &p.Permission); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModRoleApply(ctx context.Context, applyID int64) (*mod.RoleApply, error) {
	const _roleApplySQL = "SELECT id,app_key,pool_id,username,permission,operator,state,ctime,mtime FROM mod_role_apply WHERE id=?"
	row := d.db.QueryRow(ctx, _roleApplySQL, applyID)
	res := &mod.RoleApply{}
	if err := row.Scan(&res.ID, &res.AppKey, &res.PoolID, &res.Username, &res.Permission, &res.Operator, &res.State, &res.Ctime, &res.Mtime); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModRoleApplyPass(ctx context.Context, apply *mod.RoleApply) (int64, error) {
	const (
		_roleApplyStateSQL = "UPDATE mod_role_apply SET state=? WHERE id=? AND state=?"
		_permissionAddSQL  = "INSERT INTO mod_permission (username,pool_id,permission,deleted) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE permission=VALUES(permission),deleted=VALUES(deleted)"
	)
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback error:%+v", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit error:%+v", err)
		}
	}()
	if _, err = tx.Exec(_roleApplyStateSQL, mod.ApplyStatePassed, apply.ID, mod.ApplyStateChecking); err != nil {
		return 0, err
	}
	res, err := tx.Exec(_permissionAddSQL, apply.Username, apply.PoolID, apply.Permission, false)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) ModRoleApplyRefuse(ctx context.Context, applyID int64) error {
	const _roleApplyStateSQL = "UPDATE mod_role_apply SET state=? WHERE id=? AND state=?"
	_, err := d.db.Exec(ctx, _roleApplyStateSQL, mod.ApplyStateRefused, applyID, mod.ApplyStateChecking)
	return err
}

func (d *Dao) ModVersionApplyAdd(ctx context.Context, appKey, username string, versionID int64, operator, remark string, now time.Time) (int64, error) {
	const _addSQL = "INSERT INTO mod_version_apply (app_key,username,version_id,operator,remark,state,ctime) VALUES (?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE app_key=VALUES(app_key),username=VALUES(username),operator=VALUES(operator),remark=VALUES(remark),state=VALUES(state),ctime=VALUES(ctime)"
	res, err := d.db.Exec(ctx, _addSQL, appKey, username, versionID, operator, remark, mod.ApplyStateChecking, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) ModVersionApplyExist(ctx context.Context, versionID int64) (int64, error) {
	const _selSQL = "SELECT id FROM mod_version_apply WHERE version_id=? AND state=?"
	row := d.db.QueryRow(ctx, _selSQL, versionID, mod.ApplyStateChecking)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (d *Dao) ModVersionApplyByVersionID(ctx context.Context, versionIDs []int64) (map[int64]*mod.VersionApply, error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, id := range versionIDs {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	const _selSQL = "SELECT id,app_key,username,version_id,operator,remark,state,ctime,mtime FROM mod_version_apply WHERE version_id IN (%s)"
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[int64]*mod.VersionApply{}
	for rows.Next() {
		r := &mod.VersionApply{}
		if err = rows.Scan(&r.ID, &r.AppKey, &r.Username, &r.VersionID, &r.Operator, &r.Remark, &r.State, &r.Ctime, &r.Mtime); err != nil {
			return nil, err
		}
		res[r.VersionID] = r
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModVersionConfigApplyByVersionID(ctx context.Context, versionIDs []int64) (map[int64]*mod.ConfigApply, error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, id := range versionIDs {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	const _selSQL = "SELECT id,version_id,priority,app_ver,sys_ver,stime,etime,state FROM mod_version_config_apply WHERE version_id IN (%s)"
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[int64]*mod.ConfigApply{}
	for rows.Next() {
		r := &mod.ConfigApply{}
		if err = rows.Scan(&r.ID, &r.VersionID, &r.Priority, &r.AppVer, &r.SysVer, &r.Stime, &r.Etime, &r.State); err != nil {
			return nil, err
		}
		res[r.VersionID] = r
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModVersionGrayApplyByVersionID(ctx context.Context, versionIDs []int64) (map[int64]*mod.GrayApply, error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, id := range versionIDs {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	const _selSQL = "SELECT id,version_id,strategy,salt,bucket_start,bucket_end,whitelist,whitelist_url,manual_download,state FROM mod_version_gray_apply WHERE version_id IN (%s)"
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[int64]*mod.GrayApply{}
	for rows.Next() {
		r := &mod.GrayApply{}
		if err = rows.Scan(&r.ID, &r.VersionID, &r.Strategy, &r.Salt, &r.BucketStart, &r.BucketEnd, &r.Whitelist, &r.WhitelistURL, &r.ManualDownload, &r.State); err != nil {
			return nil, err
		}
		res[r.VersionID] = r
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModVersionApplyNotify(ctx context.Context, appKey, operator string) (int64, error) {
	const _selSQL = "SELECT COUNT(1) FROM mod_version_apply WHERE app_key=? AND operator=? AND state=?"
	row := d.db.QueryRow(ctx, _selSQL, appKey, operator, mod.ApplyStateChecking)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (d *Dao) ModVersionApplyList(ctx context.Context, appKey, username string) ([]*mod.VersionApply, error) {
	sql := "SELECT id,app_key,username,version_id,operator,remark,state,ctime,mtime FROM mod_version_apply WHERE app_key=? AND state=?"
	var args []interface{}
	args = append(args, appKey)
	args = append(args, mod.ApplyStateChecking)
	if username != "" {
		sql += " AND (username=? OR operator=?)"
		args = append(args, username)
		args = append(args, username)
	}
	rows, err := d.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mod.VersionApply
	for rows.Next() {
		r := &mod.VersionApply{}
		if err = rows.Scan(&r.ID, &r.AppKey, &r.Username, &r.VersionID, &r.Operator, &r.Remark, &r.State, &r.Ctime, &r.Mtime); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModVersionByVersionIDs(ctx context.Context, versionIDs []int64) (map[int64]*mod.Version, error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, id := range versionIDs {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	const _selSQL = "SELECT id,module_id,env,version,remark,from_ver_id,released,release_time,state FROM mod_version WHERE id IN (%s) ORDER BY id DESC"
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[int64]*mod.Version{}
	for rows.Next() {
		v := &mod.Version{}
		if err = rows.Scan(&v.ID, &v.ModuleID, &v.Env, &v.Version, &v.Remark, &v.FromVerID, &v.Released, &v.ReleaseTime, &v.State); err != nil {
			return nil, err
		}
		res[v.ID] = v
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModModuleByModuleIDs(ctx context.Context, moduleIDs []int64) (map[int64]*mod.Module, error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, id := range moduleIDs {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	const _selSQL = "SELECT id,pool_id,name,remark,compress,is_wifi,state,deleted,zip_check FROM mod_module WHERE id IN (%s) ORDER BY id DESC"
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[int64]*mod.Module{}
	for rows.Next() {
		v := &mod.Module{}
		if err = rows.Scan(&v.ID, &v.PoolID, &v.Name, &v.Remark, &v.Compress, &v.IsWifi, &v.State, &v.Deleted, &v.ZipCheck); err != nil {
			return nil, err
		}
		res[v.ID] = v
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModVersionApply(ctx context.Context, applyID int64) (*mod.VersionApply, error) {
	const _selSQL = "SELECT id,app_key,username,version_id,operator,remark,state,ctime,mtime FROM mod_version_apply WHERE id=?"
	row := d.db.QueryRow(ctx, _selSQL, applyID)
	r := &mod.VersionApply{}
	if err := row.Scan(&r.ID, &r.AppKey, &r.Username, &r.VersionID, &r.Operator, &r.Remark, &r.State, &r.Ctime, &r.Mtime); err != nil {
		return nil, err
	}
	return r, nil
}

func (d *Dao) ModVersionConfigApply(ctx context.Context, versionID int64) (*mod.ConfigApply, error) {
	const _versionConfigSQL = "SELECT id,version_id,priority,app_ver,sys_ver,stime,etime,state FROM mod_version_config_apply WHERE version_id=?"
	row := d.db.QueryRow(ctx, _versionConfigSQL, versionID)
	c := &mod.ConfigApply{}
	if err := row.Scan(&c.ID, &c.VersionID, &c.Priority, &c.AppVer, &c.SysVer, &c.Stime, &c.Etime, &c.State); err != nil {
		return nil, err
	}
	return c, nil
}

func (d *Dao) ModVersionConfigApplyExist(ctx context.Context, versionID int64) (int64, error) {
	const _selSQL = "SELECT id FROM mod_version_config_apply WHERE version_id=? AND state=?"
	row := d.db.QueryRow(ctx, _selSQL, versionID, mod.ApplyStateChecking)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (d *Dao) ModVersionGrayApply(ctx context.Context, versionID int64) (*mod.GrayApply, error) {
	const _versionGraySQL = "SELECT id,version_id,strategy,salt,bucket_start,bucket_end,whitelist,whitelist_url,manual_download,state FROM mod_version_gray_apply WHERE version_id=?"
	row := d.db.QueryRow(ctx, _versionGraySQL, versionID)
	g := &mod.GrayApply{}
	if err := row.Scan(&g.ID, &g.VersionID, &g.Strategy, &g.Salt, &g.BucketStart, &g.BucketEnd, &g.Whitelist, &g.WhitelistURL, &g.ManualDownload, &g.State); err != nil {
		return nil, err
	}
	return g, nil
}

func (d *Dao) ModVersionGrayApplyExist(ctx context.Context, versionID int64) (int64, error) {
	const _selSQL = "SELECT id FROM mod_version_gray_apply WHERE version_id=? AND state=?"
	row := d.db.QueryRow(ctx, _selSQL, versionID, mod.ApplyStateChecking)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (d *Dao) ModVersionApplyPass(ctx context.Context, apply *mod.VersionApply, config *mod.ConfigApply, gray *mod.GrayApply, released bool, releaseTime xtime.Time) error {
	const (
		_addConfigSQL      = "INSERT INTO mod_version_config (version_id,priority,app_ver,sys_ver,stime,etime) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE priority=VALUES(priority),app_ver=VALUES(app_ver),sys_ver=VALUES(sys_ver),stime=VALUES(stime),etime=VALUES(etime)"
		_addGraySQL        = "INSERT INTO mod_version_gray (version_id,strategy,salt,bucket_start,bucket_end,whitelist,whitelist_url,manual_download) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE strategy=VALUES(strategy),salt=VALUES(salt),bucket_start=VALUES(bucket_start),bucket_end=VALUES(bucket_end),whitelist=VALUES(whitelist),whitelist_url=VALUES(whitelist_url),manual_download=VALUES(manual_download)"
		_upVersionSQL      = "UPDATE mod_version SET released=?,release_time=? WHERE id=? AND state=?"
		_upConfigApplySQL  = "UPDATE mod_version_config_apply SET state=? WHERE version_id=? AND state=?"
		_upGrayApplySQL    = "UPDATE mod_version_gray_apply SET state=? WHERE version_id=? AND state=?"
		_upVersionApplySQL = "UPDATE mod_version_apply SET state=? WHERE id=? AND state=?"
	)
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback error:%+v", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit error:%+v", err)
		}
	}()
	if config != nil {
		if _, err = tx.Exec(_addConfigSQL, apply.VersionID, config.Priority, config.AppVer, config.SysVer, config.Stime, config.Etime); err != nil {
			return err
		}
		if _, err = tx.Exec(_upConfigApplySQL, mod.ApplyStatePassed, apply.VersionID, mod.ApplyStateChecking); err != nil {
			return err
		}
	}
	if gray != nil {
		if _, err = tx.Exec(_addGraySQL, apply.VersionID, gray.Strategy, gray.Salt, gray.BucketStart, gray.BucketEnd, gray.Whitelist, gray.WhitelistURL, gray.ManualDownload); err != nil {
			return err
		}
		if _, err = tx.Exec(_upGrayApplySQL, mod.ApplyStatePassed, apply.VersionID, mod.ApplyStateChecking); err != nil {
			return err
		}
	}
	if _, err = tx.Exec(_upVersionSQL, released, releaseTime, apply.VersionID, mod.VersionSucceeded); err != nil {
		return err
	}
	if _, err = tx.Exec(_upVersionApplySQL, mod.ApplyStatePassed, apply.ID, mod.ApplyStateChecking); err != nil {
		return err
	}
	return nil
}

func (d *Dao) ModVersionApplyRefuse(ctx context.Context, applyID, versionID int64) (err error) {
	const (
		_upVersionApplySQL = "UPDATE mod_version_apply SET state=? WHERE id=? AND state=?"
		_upConfigApplySQL  = "UPDATE mod_version_config_apply SET state=? WHERE version_id=? AND state=?"
		_upGrayApplySQL    = "UPDATE mod_version_gray_apply SET state=? WHERE version_id=? AND state=?"
	)
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback error:%+v", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit error:%+v", err)
		}
	}()
	if _, err := d.db.Exec(ctx, _upVersionApplySQL, mod.ApplyStateRefused, applyID, mod.ApplyStateChecking); err != nil {
		return err
	}
	if _, err := d.db.Exec(ctx, _upConfigApplySQL, mod.ApplyStateRefused, versionID, mod.ApplyStateChecking); err != nil {
		return err
	}
	if _, err := d.db.Exec(ctx, _upGrayApplySQL, mod.ApplyStateRefused, versionID, mod.ApplyStateChecking); err != nil {
		return err
	}
	return err
}

func (d *Dao) ModModuleListByName(ctx context.Context, name string) ([]*mod.Module, error) {
	const _moduleListSQL = "SELECT id,pool_id,name,remark,compress,is_wifi,state,deleted,zip_check FROM mod_module WHERE name=? AND deleted=0 ORDER BY id DESC"
	rows, err := d.db.Query(ctx, _moduleListSQL, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mod.Module
	for rows.Next() {
		m := &mod.Module{}
		if err = rows.Scan(&m.ID, &m.PoolID, &m.Name, &m.Remark, &m.Compress, &m.IsWifi, &m.State, &m.Deleted, &m.ZipCheck); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModOriginalFile(ctx context.Context, versionID int64) (*mod.File, error) {
	const _fileSQL = "SELECT id,name,size,md5,url,ctime FROM mod_file WHERE version_id=? AND is_patch=0 AND from_ver=0"
	rows := d.db.QueryRow(ctx, _fileSQL, versionID)
	f := &mod.File{}
	if err := rows.Scan(&f.ID, &f.Name, &f.Size, &f.Md5, &f.URL, &f.Ctime); err != nil {
		return nil, err
	}
	return f, nil
}

func (d *Dao) ModSyncAdd(ctx context.Context, toModuleID, toVersionID int64, v *mod.Version, config *mod.Config, gray *mod.Gray) (versionID int64, err error) {
	const (
		_fileSQL          = "SELECT id,name,size,md5,url,ctime FROM mod_file WHERE version_id=? AND is_patch=0 AND from_ver=0"
		_maxVersionSQL    = "SELECT MAX(version) FROM mod_version WHERE module_id=?"
		_insertVersionSQL = "INSERT INTO mod_version (module_id,env,version,remark,state) VALUES (?,?,?,?,?)"
		_insertFileSQL    = "INSERT INTO mod_file (version_id,name,content_type,size,md5,url,is_patch,from_ver) VALUES (?,?,?,?,?,?,?,?)"
		_addConfigSQL     = "INSERT INTO mod_version_config (version_id,priority,app_ver,sys_ver,stime,etime) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE priority=VALUES(priority),app_ver=VALUES(app_ver),sys_ver=VALUES(sys_ver),stime=VALUES(stime),etime=VALUES(etime)"
		_addGraySQL       = "INSERT INTO mod_version_gray (version_id,strategy,salt,bucket_start,bucket_end,whitelist,whitelist_url,manual_download) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE strategy=VALUES(strategy),salt=VALUES(salt),bucket_start=VALUES(bucket_start),bucket_end=VALUES(bucket_end),whitelist=VALUES(whitelist),whitelist_url=VALUES(whitelist_url),manual_download=VALUES(manual_download)"
	)
	rows := d.db.QueryRow(ctx, _fileSQL, v.ID)
	file := &mod.File{}
	if err := rows.Scan(&file.ID, &file.Name, &file.Size, &file.Md5, &file.URL, &file.Ctime); err != nil {
		return 0, err
	}
	var maxVersion sql.NullInt64
	if toVersionID == 0 {
		row := d.db.QueryRow(ctx, _maxVersionSQL, toModuleID)
		if err := row.Scan(&maxVersion); err != nil && err != xsql.ErrNoRows {
			return 0, err
		}
	}
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback error:%+v", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit error:%+v", err)
		}
	}()
	if err = func() error {
		if toVersionID != 0 {
			return nil
		}
		version := maxVersion.Int64 + 1
		res, err := tx.Exec(_insertVersionSQL, toModuleID, v.Env, version, v.Remark, mod.VersionProcessing)
		if err != nil {
			return errors.WithStack(err)
		}
		if toVersionID, err = res.LastInsertId(); err != nil {
			return errors.WithStack(err)
		}
		if _, err = tx.Exec(_insertFileSQL, toVersionID, file.Name, file.ContentType, file.Size, file.Md5, file.URL, file.IsPatch, file.FromVer); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}(); err != nil {
		return 0, err
	}
	if config != nil {
		if _, err = tx.Exec(_addConfigSQL, toVersionID, config.Priority, config.AppVer, config.SysVer, config.Stime, config.Etime); err != nil {
			return 0, errors.WithStack(err)
		}
	}
	if gray != nil {
		if _, err = tx.Exec(_addGraySQL, toVersionID, gray.Strategy, gray.Salt, gray.BucketStart, gray.BucketEnd, gray.Whitelist, gray.WhitelistURL, gray.ManualDownload); err != nil {
			return 0, errors.WithStack(err)
		}
	}
	return toVersionID, nil
}

func (d *Dao) ReleasedVersionExists(ctx context.Context, moduleID int64) (bool, error) {
	const _versionCountSQL = "SELECT COUNT(1) FROM mod_version WHERE module_id=? AND released=1"
	var count int64
	row := d.db.QueryRow(ctx, _versionCountSQL, moduleID)
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *Dao) ModulePushOffline(ctx context.Context, moduleID int64, appKey, poolName, moduleName string) (bool, error) {
	const (
		_maxVersionSQL   = "SELECT env,MAX(version) FROM mod_version WHERE module_id=? GROUP BY env"
		_moduleStateSQL  = "UPDATE mod_module SET state=? WHERE id=? AND state=? AND deleted=0"
		_versionStateSQL = "UPDATE mod_version SET state=?,released=0 WHERE module_id=? AND version<=? AND env=? AND state!=?"
	)
	maxVersiom := map[mod.Env]int64{}
	rows, err := d.db.Query(ctx, _maxVersionSQL, moduleID)
	if err != nil {
		return false, errors.WithStack(err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			env        mod.Env
			maxVersion int64
		)
		if err = rows.Scan(&env, &maxVersion); err != nil {
			return false, errors.WithStack(err)
		}
		maxVersiom[env] = maxVersion
	}
	if err := rows.Err(); err != nil {
		return false, errors.WithStack(err)
	}
	if len(maxVersiom) == 0 {
		return false, nil
	}
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return false, errors.WithStack(err)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			log.Error("recover:%+v", r)
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback error:%+v", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit error:%+v", err)
		}
	}()
	res, err := tx.Exec(_moduleStateSQL, mod.ModuleOffline, moduleID, mod.ModuleOnline)
	if err != nil {
		return false, errors.WithStack(err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return false, errors.WithStack(err)
	}
	if ra == 0 {
		return false, nil
	}
	for env, maxVersion := range maxVersiom {
		if _, err = tx.Exec(_versionStateSQL, mod.VersionDisable, moduleID, maxVersion, env, mod.VersionDisable); err != nil {
			return false, errors.WithStack(err)
		}
		if ra, err = res.RowsAffected(); err != nil {
			return false, errors.WithStack(err)
		}
		if ra == 0 {
			continue
		}
		if err = d.setModuleDisable(ctx, appKey, env, poolName, moduleName, maxVersion); err != nil {
			return false, errors.WithStack(err)
		}
	}
	return true, nil
}

func (d *Dao) setModuleDisable(ctx context.Context, appKey string, env mod.Env, poolName, moduleName string, version int64) error {
	key := fmt.Sprintf("mod_disable_%s_%s", appKey, env)
	_, err := d.redis.Do(ctx, "HSET", key, fmt.Sprintf("%s_%s", poolName, moduleName), version)
	return err
}
