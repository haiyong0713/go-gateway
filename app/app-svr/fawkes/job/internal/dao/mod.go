package dao

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"go-common/library/log"

	"go-gateway/app/app-svr/fawkes/job/internal/model/mod"

	"github.com/pkg/errors"
)

func (d *dao) VersionByID(ctx context.Context, id int64) (*mod.Version, error) {
	const _versionByIDSQL = "SELECT id,module_id,env,version,from_ver_id,state FROM mod_version WHERE id=?"
	v := &mod.Version{}
	row := d.fawkesDB.QueryRow(ctx, _versionByIDSQL, id)
	if err := row.Scan(&v.ID, &v.ModuleID, &v.Env, &v.Version, &v.FromVerID, &v.State); err != nil {
		return nil, err
	}
	return v, nil
}

func (d *dao) VersionByIDs(ctx context.Context, ids []int64) ([]*mod.Version, error) {
	const _versionByIDsSQL = "SELECT id,module_id,env,version,from_ver_id,state FROM mod_version WHERE id IN (%s)"
	if len(ids) == 0 {
		return []*mod.Version{}, nil
	}
	var (
		sqls []string
		args []interface{}
	)
	for _, id := range ids {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	rows, err := d.fawkesDB.Query(ctx, fmt.Sprintf(_versionByIDsSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mod.Version
	for rows.Next() {
		f := &mod.Version{}
		if err = rows.Scan(&f.ID, &f.ModuleID, &f.Env, &f.Version, &f.FromVerID, &f.State); err != nil {
			return nil, err
		}
		res = append(res, f)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *dao) OriginalFile(ctx context.Context, versionID int64) (*mod.File, error) {
	const _originFileSQL = "SELECT id,version_id,name,content_type,size,md5,url,is_patch,from_ver FROM mod_file WHERE version_id=? AND is_patch=0 AND from_ver=0"
	f := &mod.File{}
	row := d.fawkesDB.QueryRow(ctx, _originFileSQL, versionID)
	if err := row.Scan(&f.ID, &f.VersionID, &f.Name, &f.ContentType, &f.Size, &f.Md5, &f.URL, &f.IsPatch, &f.FromVer); err != nil {
		return nil, err
	}
	return f, nil
}

func (d *dao) LastVersionList(ctx context.Context, moduleID, version, limit int64, env mod.Env) ([]*mod.Version, error) {
	const _versionListSQL = "SELECT id,module_id,version,state FROM mod_version WHERE module_id=? AND version<? AND env=? AND state!=? ORDER BY version DESC LIMIT ?"
	rows, err := d.fawkesDB.Query(ctx, _versionListSQL, moduleID, version, env, mod.VersionDisable, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mod.Version
	for rows.Next() {
		v := &mod.Version{}
		if err = rows.Scan(&v.ID, &v.ModuleID, &v.Version, &v.State); err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *dao) VersionList(ctx context.Context, moduleID int64, versions []int64, env mod.Env) (verList []*mod.Version, err error) {
	const _versionListSQL = "SELECT id,module_id,version,state FROM mod_version WHERE module_id=? AND version IN (%s) AND env=? AND state!=? ORDER BY version DESC"
	if len(versions) == 0 {
		return
	}
	var (
		sqls []string
		args []interface{}
	)
	args = append(args, moduleID)
	for _, version := range versions {
		sqls = append(sqls, "?")
		args = append(args, version)
	}
	args = append(args, env, mod.VersionDisable)
	rows, err := d.fawkesDB.Query(ctx, fmt.Sprintf(_versionListSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mod.Version
	for rows.Next() {
		v := &mod.Version{}
		if err = rows.Scan(&v.ID, &v.ModuleID, &v.Version, &v.State); err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *dao) OriginalFileList(ctx context.Context, versionIDs []int64) ([]*mod.File, error) {
	const _originalFileSQL = "SELECT id,version_id,name,url FROM mod_file WHERE version_id IN(%s) AND is_patch=0 AND from_ver=0 ORDER BY id DESC"
	var (
		sqls []string
		args []interface{}
	)
	for _, id := range versionIDs {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	rows, err := d.fawkesDB.Query(ctx, fmt.Sprintf(_originalFileSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*mod.File
	for rows.Next() {
		f := &mod.File{}
		if err = rows.Scan(&f.ID, &f.VersionID, &f.Name, &f.URL); err != nil {
			return nil, err
		}
		res = append(res, f)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *dao) VersionSucceed(ctx context.Context, id int64) error {
	const _versionStateSQL = "UPDATE mod_version SET state=? WHERE id=? AND state=?"
	_, err := d.fawkesDB.Exec(ctx, _versionStateSQL, mod.VersionSucceeded, id, mod.VersionProcessing)
	return err
}

func (d *dao) PatchAdd(ctx context.Context, version *mod.Version, patchFiles []*mod.File) error {
	const (
		_insertFileSQL   = "INSERT INTO mod_file (version_id,name,content_type,size,md5,url,is_patch,from_ver) VALUES %s ON DUPLICATE KEY UPDATE name=VALUES(name),content_type=VALUES(content_type),size=VALUES(size),md5=VALUES(md5),url=VALUES(url)"
		_versionStateSQL = "UPDATE mod_version SET state=? WHERE id=? AND state=?"
	)
	if len(patchFiles) == 0 {
		return nil
	}
	var (
		sqls []string
		args []interface{}
	)
	for _, f := range patchFiles {
		sqls = append(sqls, "(?,?,?,?,?,?,?,?)")
		args = append(args, version.ID, f.Name, f.ContentType, f.Size, f.Md5, f.URL, f.IsPatch, f.FromVer)
	}
	tx, err := d.fawkesDB.Begin(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if r := recover(); r != nil {
			if err1 := tx.Rollback(); err != nil {
				log.Error("tx.Rollback error:%+v", err1)
			}
			log.Error("recover:%+v", r)
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
	if _, err = tx.Exec(fmt.Sprintf(_insertFileSQL, strings.Join(sqls, ",")), args...); err != nil {
		return err
	}
	if _, err = tx.Exec(_versionStateSQL, mod.VersionSucceeded, version.ID, mod.VersionProcessing); err != nil {
		return err
	}
	return nil
}

func (d *dao) DownloadFile(ctx context.Context, url, filePath string) error {
	req, err := d.client.NewRequest(http.MethodGet, url, "", nil)
	if err != nil {
		return errors.Wrap(err, url)
	}
	res, err := d.client.Raw(ctx, req)
	if err != nil {
		return errors.Wrap(err, url)
	}
	f, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, filePath)
	}
	if _, err = io.Copy(f, bytes.NewReader(res)); err != nil {
		return errors.Wrap(err, filePath)
	}
	return nil
}
