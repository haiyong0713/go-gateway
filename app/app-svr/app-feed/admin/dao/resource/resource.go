package resource

import (
	"context"
	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	"go-common/library/database/elastic"
	"go-common/library/database/orm"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	model "go-gateway/app/app-svr/app-feed/admin/model/resource"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

// Dao is
type Dao struct {
	Conf             *conf.Config
	DB               *gorm.DB
	client           *bm.Client
	es               *elastic.Elastic
	simpleArchiveURL string
	archiveClient    archiveGRPC.ArchiveClient
	//Consumer         *databus.Databus
}

// New is
func New(c *conf.Config) *Dao {
	dao := &Dao{
		Conf:             c,
		DB:               orm.NewMySQL(c.ORMResource),
		client:           bm.NewClient(c.HTTPClient.Read),
		es:               elastic.NewElastic(c.ES),
		simpleArchiveURL: c.Host.Archive + _simpleArchiveURL,
	}
	var err error
	if dao.archiveClient, err = archiveGRPC.NewClient(nil); err != nil {
		panic(err)
	}
	return dao
}

// GetCustomConfig is
func (d *Dao) GetCustomConfig(ctx context.Context, id int64) (*model.CustomConfig, error) {
	out := &model.CustomConfig{}
	if err := d.DB.Table("custom_config").Where("id=?", id).First(out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// GetCustomConfigBy is
func (d *Dao) GetCustomConfigBy(ctx context.Context, tp int64, oid int64) (*model.CustomConfig, error) {
	out := &model.CustomConfig{}
	if err := d.DB.Table("custom_config").Where("tp=?", tp).Where("oid=?", oid).First(out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// CCList is
func (d *Dao) CCList(ctx context.Context, req *model.CCListReq) (*model.CCListReply, error) {
	idSQL := "select id from custom_config where "
	if req.TP > 0 {
		idSQL += "tp=" + strconv.FormatInt(req.TP, 10) + " and "
	}
	if req.OidNum > 0 {
		idSQL += "oid=" + strconv.FormatInt(req.OidNum, 10) + " and "
	}
	idSQL += "origin_type=" + strconv.FormatInt(int64(req.OriginType), 10) + " order by id desc "
	idSQL += "limit " + strconv.FormatInt(req.PS, 10) + " offset " + strconv.FormatInt((req.PN-1)*req.PS, 10)

	ccs := []*model.CustomConfig{}
	if err := d.DB.Table("custom_config").Joins("inner join (" + idSQL + ") as ids on custom_config.id = ids.id").Find(&ccs).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	ccsr := make([]*model.CustomConfigReply, 0, len(ccs))
	for _, cc := range ccs {
		ccsr = append(ccsr, &model.CustomConfigReply{
			CustomConfig: *cc,
			Status:       cc.ResolveStatusAt(now),
		})
	}
	reply := &model.CCListReply{
		Data: ccsr,
		Page: common.Page{
			Num:  int(req.PN),
			Size: int(req.PS),
		},
	}
	return reply, nil
}

// CCAdd is
func (d *Dao) CCAdd(ctx context.Context, req *model.CCAddReq) (rows int64, err error) {
	res := d.DB.Exec(`INSERT IGNORE INTO custom_config(tp,oid,content,url,highlight_content,image,image_big,stime,etime,state,origin_type,audit_code) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		req.TP, req.OidNum, req.Content, req.URL, req.HighlightContent, req.Image, req.ImageBig, req.STime, req.ETime, model.CustomConfigStateEnable, req.OriginType, req.AuditCode)
	if res.Error != nil {
		err = res.Error
	}
	rows = res.RowsAffected
	return
}

// CCUpdate is
func (d *Dao) CCUpdate(ctx context.Context, req *model.CCUpdateReq) error {
	if err := d.DB.Exec(`UPDATE custom_config SET oid=?,content=?,url=?,highlight_content=?,image=?,image_big=?,stime=?,etime=?,origin_type=?,audit_code=? WHERE id=?`,
		req.OidNum, req.Content, req.URL, req.HighlightContent, req.Image, req.ImageBig, req.STime, req.ETime, req.OriginType, req.AuditCode, req.ID).Error; err != nil {
		return err
	}
	return nil
}

// CCUpdateState is
func (d *Dao) CCUpdateState(ctx context.Context, id int64, state int64) error {
	if err := d.DB.Exec(`UPDATE custom_config SET state=? WHERE id=? LIMIT 1`, state, id).Error; err != nil {
		return err
	}
	return nil
}

// CCUpdateAuditCode is
func (d *Dao) CCUpdateAuditCode(_ context.Context, id int64, auditCode int32) error {
	if err := d.DB.Exec(`UPDATE custom_config SET audit_code=? WHERE id=? LIMIT 1`, auditCode, id).Error; err != nil {
		return err
	}
	return nil
}

func (d *Dao) CCListTotal(req *model.CCListReq) (total int, err error) {
	query := d.DB.Table("custom_config")
	if req.TP > 0 {
		query = query.Where("tp=?", req.TP)
	}
	if req.OidNum > 0 {
		query = query.Where("oid=?", req.OidNum)
	}
	query = query.Where("origin_type = ?", req.OriginType)

	err = query.Count(&total).Error
	if err != nil {
		log.Error("CCListTotal error: %s", err.Error())
	}

	return total, nil
}
