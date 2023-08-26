package fawkes

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	dsql "database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/database/bfs"
	"go-common/library/database/orm"
	"go-common/library/database/sql"
	xhttp "go-common/library/net/http/blademaster"

	fissiGrpc "git.bilibili.co/bapis/bapis-go/account/service/fission"
	databusV2 "go-common/library/queue/databus.v2"

	"go-gateway/app/app-svr/fawkes/service/conf"
	clickhouseGenerator "go-gateway/app/app-svr/fawkes/service/dao/database"
	"go-gateway/app/app-svr/fawkes/service/dao/oss"
	"go-gateway/app/app-svr/fawkes/service/model"
	"go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/jinzhu/gorm"
)

// Main Dao
type Dao struct {
	c *conf.Config
	// Databases
	db          *sql.DB // fawkes-admin db
	ORMDB       *gorm.DB
	mdb         *sql.DB                 // macross db (will be deprecated)
	bsdb        *sql.DB                 // bili_show db (will be deprecated)
	veda        *sql.DB                 // veda_crash_db
	redis       *redis.Redis            // fawkes-admin redis
	clickhouse  *clickhouseGenerator.DB // mobile-ep 新集群
	clickhouse2 *clickhouseGenerator.DB // 运维旧集群
	// Business
	httpClient      *xhttp.Client
	bfsCli          *bfs.BFS
	broadcastClient *Broadcast
	databusClient   databusV2.Client
	// URLs
	treeToken string
	treeRole  string
	treeAuth  string
	treeApp   string
	bfsCache  string

	topResource string
	// ossWapper
	ossWrapper *oss.Dao
	fission    fissiGrpc.FissionClient
}

// New dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// Databases
		db:          sql.NewMySQL(c.MySQL.Fawkes),
		mdb:         sql.NewMySQL(c.MySQL.Macross),
		ORMDB:       orm.NewMySQL(c.ORM),
		bsdb:        sql.NewMySQL(c.MySQL.Show),
		veda:        sql.NewMySQL(c.MySQL.Veda),
		redis:       redis.NewRedis(c.Redis.Fawkes),
		clickhouse:  clickhouseGenerator.NewClickhouse(c.ClickHouse.Monitor),
		clickhouse2: clickhouseGenerator.NewClickhouse(c.ClickHouse.Monitor2),
		// Business
		bfsCli:          bfs.New(c.BFS),
		broadcastClient: NewBroadcast(),
		httpClient:      xhttp.NewClient(c.HTTPClient),
		// DataBus
		databusClient: NewDatabus(c.Databus),
		// URLs
		treeToken:   c.Host.Easyst + _token,
		treeRole:    c.Host.Easyst + _role,
		treeAuth:    c.Host.Easyst + _auth,
		treeApp:     c.Host.Easyst + _treeApp,
		bfsCache:    c.Host.Sven + _bfsCache,
		topResource: c.Host.Bender + _topResource,
		// OSS Wrapper
		ossWrapper: oss.New(c),
		fission:    NewFisson(conf.Conf),
	}
	d.initORM()
	return
}

// UpBFS upload to bfs.
func (d *Dao) UpBFS(dir, filePath string) (cdnURL, md5Str string, err error) {
	var f *os.File
	if f, err = os.Open(filePath); err != nil {
		log.Error("%v", err)
		return
	}
	filename := path.Base(filePath)
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, f); err != nil {
		log.Error("%v", err)
		return
	}
	md5Bs := md5.Sum(buf.Bytes())
	md5Str = hex.EncodeToString(md5Bs[:])
	if cdnURL, err = d.Upload(context.Background(), model.BFSBucket, dir, filename, "", buf.Bytes()); err != nil {
		log.Error("%v", err)
	}
	return
}

// UpBFSV2 兼容海外版.
// 国内版本 -> 发布至"国内BFS"
// 海外版本 -> 发布至"海外OSS"
func (d *Dao) UpBFSV2(dir, filePath, appKey string) (cdnURL, md5Str string, err error) {
	var (
		f       *os.File
		appInfo *app.APP
	)
	if f, err = os.Open(filePath); err != nil {
		log.Error("%v", err)
		return
	}
	filename := path.Base(filePath)
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, f); err != nil {
		log.Error("%v", err)
		return
	}
	md5Bs := md5.Sum(buf.Bytes())
	md5Str = hex.EncodeToString(md5Bs[:])
	if appInfo, err = d.AppPass(context.Background(), appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if cdnURL, err = d.Upload(context.Background(), model.BFSBucket, dir, filename, "", buf.Bytes()); err != nil {
		log.Error("%v", err)
		return
	}
	if appInfo.ServerZone == app.AppServerZone_Abroad {
		if cdnURL, _, _, err = d.ossWrapper.FileUploadOss(context.Background(), filePath, path.Join(dir, filename), appInfo.ServerZone); err != nil {
			log.Error("%v", err)
			return
		}
	}
	return
}

// Upload to bfs.
func (d *Dao) Upload(c context.Context, bucket, dir, fileName, contentType string, file []byte) (url string, err error) {
	if url, err = d.bfsCli.Upload(c, &bfs.Request{
		Bucket:      bucket,
		ContentType: contentType,
		Filename:    fileName,
		File:        file,
		Dir:         dir,
	}); err != nil {
		log.Error("Upload(err:%v)", err)
	}
	return
}

// DiffCmd execute the bsdiff command to have the result
func (d *Dao) DiffCmd(folder, filename string, newPath string, oldPath string) (patchPath string, err error) {
	patchPath = path.Join(folder, filename)
	cmd := exec.Command("bsdiff", oldPath, newPath, patchPath)
	// exec Command
	if err = cmd.Run(); err != nil {
		log.Error("%v", err)
	}
	return
}

// BeginTran begin transcation.
func (d *Dao) BeginTran(c context.Context) (tx *sql.Tx, err error) {
	return d.db.Begin(c)
}

// VedaBeginTran begin transcation.
func (d *Dao) VedaBeginTran(c context.Context) (tx *sql.Tx, err error) {
	return d.veda.Begin(c)
}

// GetRowsValue get rows value
func (d *Dao) GetRowsValue(rows *dsql.Rows, structPtr interface{}) (err error) {
	v := reflect.ValueOf(structPtr)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return
	}
	e := v.Elem()
	t := e.Type()
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	var dest_scans = make([]interface{}, len(cols))
	for i, c := range cols {
		for j := 0; j < t.NumField(); j++ {
			if t.Field(j).Tag.Get("json") == c || t.Field(j).Tag.Get("json") == c+",omitempty" {
				dest_scans[i] = e.Field(j).Addr().Interface()
			}
		}
	}
	return rows.Scan(dest_scans...)
}

// Transact is
func (d *Dao) Transact(ctx context.Context, txFunc func(*sql.Tx) error) (err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			//nolint:errcheck
			tx.Rollback()
			// panic(p) // re-throw panic after Rollback
			log.Error("Panic in Transact: %+v", p)
			return
		}
		if err != nil {
			//nolint:errcheck
			tx.Rollback() // err is non-nil; don't change it
			return
		}
		err = tx.Commit() // err is nil; if Commit returns error update err
	}()
	err = txFunc(tx)
	return err
}

// Veda Transact
func (d *Dao) VedaTransact(ctx context.Context, txFunc func(*sql.Tx) error) (err error) {
	tx, err := d.veda.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			//nolint:errcheck
			tx.Rollback()
			// panic(p) // re-throw panic after Rollback
			log.Error("Panic in Transact: %+v", p)
			return
		}
		if err != nil {
			//nolint:errcheck
			tx.Rollback() // err is non-nil; don't change it
			return
		}
		err = tx.Commit() // err is nil; if Commit returns error update err
	}()
	err = txFunc(tx)
	return err
}

// BeginMacrossTran begin macross transcation.
func (d *Dao) BeginMacrossTran(c context.Context) (tx *sql.Tx, err error) {
	return d.mdb.Begin(c)
}

// BeginShowTran begin show transcation.
func (d *Dao) BeginShowTran(c context.Context) (tx *sql.Tx, err error) {
	return d.bsdb.Begin(c)
}

// BeginVedaTran begin show transcation.
func (d *Dao) BeginVedaTran(c context.Context) (tx *sql.Tx, err error) {
	return d.veda.Begin(c)
}

// WriteConfigFile write config file.
func (d *Dao) WriteConfigFile(folder, filename string, contents []byte) (zipFilePath string, err error) {
	// cvs := strconv.FormatInt(cv, 10)
	fileInfo, err := os.Stat(folder)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(folder, 0755); err != nil {
			log.Error("os.MkDirAll(%s) error(%v)", folder, err)
			return
		}
		err = nil // NOTE: folder ok~
	} else if !fileInfo.IsDir() {
		err = fmt.Errorf("%s is not folder", folder)
		log.Error("%v", err)
		return
	}
	var file *os.File
	originFile := path.Join(folder, filename)
	if file, err = os.OpenFile(originFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644); err != nil {
		log.Error("os.OpenFile(%s) error(%v)", originFile, err)
		return
	}
	defer file.Close()
	if _, err = file.Write(contents); err != nil {
		log.Error("%v", err)
		return
	}
	if zipFilePath, err = d.ZIP(folder, filename, originFile); err != nil {
		log.Error("%v", err)
	}
	return
}

// ZIP zip config.
func (d *Dao) ZIP(folder, filename string, originFilePath string) (zipFilePath string, err error) {
	var file *os.File
	if file, err = os.Open(originFilePath); err != nil {
		log.Error("os.OpenFile(%s) error(%v)", originFilePath, err)
		return
	}
	defer file.Close()
	zipFilePath = path.Join(folder, fmt.Sprintf("%v.zip", filename))
	var fileInfo os.FileInfo
	if fileInfo, err = file.Stat(); err != nil {
		log.Error("%v", err)
		return
	}
	var zipfile *os.File
	if zipfile, err = os.Create(zipFilePath); err != nil {
		log.Error("%v", err)
		return
	}
	defer zipfile.Close()
	archive := zip.NewWriter(zipfile)
	defer archive.Close()
	var header *zip.FileHeader
	if header, err = zip.FileInfoHeader(fileInfo); err != nil {
		log.Error("%v", err)
		return
	}
	header.Method = zip.Deflate
	var writer io.Writer
	if writer, err = archive.CreateHeader(header); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = io.Copy(writer, file); err != nil {
		log.Error("%v", err)
	}
	return
}

// Ping dao.
func (d *Dao) Ping(c context.Context) (err error) {
	if err = d.db.Ping(c); err != nil {
		log.Error("d.db error(%v)", err)
	}
	if d.ORMDB != nil {
		err = d.ORMDB.DB().PingContext(c)
		log.Error("d.ormdb error(%v)", err)
	}
	return

}

// Close close kafka connection.
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
	if d.ORMDB != nil {
		d.ORMDB.Close()
	}
}

// FormLike fmt sql like.
func (d *Dao) FormLike(kv string, params []string, ltype string) (sqls string, args []interface{}) {
	var likes []string
	for _, param := range params {
		args = append(args, kv+"%")
		likes = append(likes, param+" LIKE ?")
	}
	sqls = fmt.Sprintf("(%s)", strings.Join(likes, " "+ltype+" "))
	return
}

func (d *Dao) initORM() {
	d.ORMDB.LogMode(true)
}

// // FormOrder fmt sql order.
// func (d *Dao) FormOrder(order, sort string) string {
// 	return fmt.Sprintf(" ORDER BY %s %s", order, sort)
// }

// // FormLimit fmt sql limit.
// func (d *Dao) FormLimit(pn, ps int) string {
// 	return fmt.Sprintf(" LIMIT %d,%d ", (pn-1)*ps, ps)
// }

/*********************************调用接口 统一**********************************/

// ExternalErrorc 外部调用接口
func (d *Dao) ExternalErrorc(ctx context.Context, okStatus, curStatus int64, msg string, args ...interface{}) error {
	if curStatus != okStatus {
		log.Errorc(ctx, fmt.Sprintf("external interface error: status code: %v, msg: %v", curStatus, msg), args...)
		return fmt.Errorf("external interface error %v", msg)
	}
	log.Warnc(ctx, fmt.Sprintf("external interface calling: %v", msg))
	return nil
}
