package fawkes

import (
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/ecode"

	"github.com/pkg/errors"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_txAddApk = `INSERT INTO apks (version_code,version_id,version_name,cdn_addr,inet_addr,mapping_addr,local_path,
m_d5,size,is_gray,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`
	_txAddDiffPatch = `INSERT INTO diffs (target_version,target_version_code,target_version_id,source_version,
source_version_code,source_version_id,cdn_addr,inet_addr,local_path,m_d5,size,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`
)

// TxAddApk insert apk.
func (d *Dao) TxAddApk(tx *xsql.Tx, versionCode, buildID int64, version, cdnURL, packURL, mappingURL, packPath, md5 string,
	size int64, isGray int) (r int64, err error) {
	ctime := time.Now().Unix()
	res, err := tx.Exec(_txAddApk, versionCode, strconv.FormatInt(buildID, 10), version, cdnURL, packURL, mappingURL,
		packPath, md5, size, isGray, ctime, ctime)
	if err != nil {
		log.Error("TxAddApk %v", err)
		return
	}
	return res.RowsAffected()
}

// TxAddDiffPatch insert diff patch.
func (d *Dao) TxAddDiffPatch(tx *xsql.Tx, targetVersion, originVersion string, targetBuildID, originBuildID, targetVersionCode,
	originVersionCode int64, cdnURL, patchURL, patchPath, md5 string, size int64) (r int64, err error) {
	ctime := time.Now().Unix()
	res, err := tx.Exec(_txAddDiffPatch, targetVersion, targetVersionCode, strconv.FormatInt(targetBuildID, 10),
		originVersion, originVersionCode, strconv.FormatInt(originBuildID, 10), cdnURL, patchURL, patchPath, md5, size, ctime, ctime)
	if err != nil {
		log.Error("TxAddDiffPatch %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) MacrossFileInfo(c context.Context, url string) (content string, err error) {
	// nolint:gosec
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.Wrap(ecode.Int(resp.StatusCode), url)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	content = string(body)
	return
}
