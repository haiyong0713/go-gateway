package dao

import (
	"context"
	"time"

	"go-gateway/app/app-svr/kvo/job/internal/model"

	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	"go-common/library/log"
)

const (
	_getDocument = "SELECT check_sum,doc FROM document WHERE check_sum=?"
	_getUserConf = "SELECT mid,module_key,check_sum,timestamp FROM user_conf WHERE mid=? AND module_key=?"
	_upDocument  = "INSERT INTO document(check_sum,doc,ctime,mtime) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE doc=?"
	_upUserConf  = "INSERT INTO user_conf(mid,module_key,check_sum,timestamp,ctime,mtime) VALUES(?,?,?,?,?,?) ON DUPLICATE KEY UPDATE check_sum=?, timestamp=?"
)

var getUserConf *sql.Stmt

func NewDB() (db *sql.DB, err error) {
	var cfg struct {
		Mysql *sql.Config
	}
	if err = paladin.Get("db.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	db = sql.NewMySQL(cfg.Mysql)
	getUserConf = db.Prepared(_getUserConf)
	return
}

// Document get
func (d *dao) documentDB(ctx context.Context, checkSum int64) (doc *model.Document, err error) {
	row := d.db.QueryRow(ctx, _getDocument, checkSum)
	doc = &model.Document{}
	err = row.Scan(&doc.CheckSum, &doc.Doc)
	if err != nil {
		if err == sql.ErrNoRows {
			doc = nil
			err = nil
			return
		}
		log.Error("d.Document row.scan err:%v", err)
	}
	return
}

// UserConf get userconf
func (d *dao) userConfDB(ctx context.Context, mid int64, moduleKey int) (userConf *model.UserConf, err error) {
	row := getUserConf.QueryRow(ctx, mid, moduleKey)
	userConf = &model.UserConf{}
	err = row.Scan(&userConf.Mid, &userConf.ModuleKey, &userConf.CheckSum, &userConf.Timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			userConf = nil
			err = nil
			return
		}
		log.Error("d.UserConf row.Scan err:%v", err)
	}
	return
}

// TxUpDocuement add a document
func (d *dao) TxUpDocuement(ctx context.Context, tx *sql.Tx, checkSum int64, data string, now time.Time) (err error) {
	_, err = tx.Exec(_upDocument, checkSum, data, now, now, data)
	if err != nil {
		log.Error("d.UpDocuement(checksum:%d) db.exec err:%v", checkSum, err)
	}
	return
}

// TxUpUserConf add or update user conf
func (d *dao) TxUpUserConf(ctx context.Context, tx *sql.Tx, mid int64, moduleKey int, checkSum int64, now time.Time) (err error) {
	_, err = tx.Exec(_upUserConf, mid, moduleKey, checkSum, now.Unix(), now, now, checkSum, now.Unix())
	if err != nil {
		log.Error("d.TxUpUserConf (mid:%d,key:%d,checksum:%d) db.exec err:%v", mid, moduleKey, checkSum, err)
	}
	return
}
