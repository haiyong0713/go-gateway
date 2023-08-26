package rank

import (
	"context"
	"go-common/library/database/elastic"
	"go-common/library/database/orm"
	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/admin/conf"
	rankmdl "go-gateway/app/web-svr/activity/admin/model/rank_v3"

	"github.com/jinzhu/gorm"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package rank

// Dao dao interface
type Dao interface {
	Close()
	Create(c context.Context, tx *xsql.Tx, rank *rankmdl.Base, sources []*rankmdl.Source, blackWhite []*rankmdl.BlackWhite) (id int64, err error)
	Update(c context.Context, tx *xsql.Tx, rank *rankmdl.Base, sources []*rankmdl.Source, blackWhite []*rankmdl.BlackWhite) (err error)
	GetRankByID(c context.Context, id int64) (rank *rankmdl.Base, err error)
	AddRankRule(c context.Context, tx *xsql.Tx, rule *rankmdl.Rule, score []*rankmdl.ScoreConfig) (err error)
	UpdateRankRule(c context.Context, tx *xsql.Tx, rule *rankmdl.Rule, score []*rankmdl.ScoreConfig) (err error)
	GetSource(c context.Context, baseID int64, sourceType int) (list []*rankmdl.Source, err error)
	GetRule(c context.Context, baseID int64) (list []*rankmdl.Rule, err error)
	GetRuleByID(c context.Context, id int64) (n *rankmdl.Rule, err error)
	ListTotal(c context.Context, state int, keyword string, rankType int, validateTime int64) (total int, err error)
	GetRankList(c context.Context, pn, ps, state int, keyword string, rankType int, validateTime int64) (list []*rankmdl.Base, err error)
	GetSourceBatch(c context.Context, baseIDs []int64) (list []*rankmdl.Source, err error)
	GetRankByStateAndTime(c context.Context, state int, sqlCondition string) (list []*rankmdl.Rule, err error)
	UpdateRuleState(c context.Context, baseIDs []int64, state int) (err error)
	UpdateRankRuleState(c context.Context, rule *rankmdl.Rule) (err error)
	GetBlackOrWhite(c context.Context, baseID int64, interventionType, objectType int) (list []*rankmdl.BlackWhite, err error)
	AddBlackWhite(c context.Context, id int64, whiteBlack []*rankmdl.BlackWhite) (err error)
	UpdateRankAdjust(c context.Context, baseID int64, adjust *rankmdl.Adjust) (err error)
	RankArchive(c context.Context, baseID, rankID int64, batch int, tagID, mid []int64, offset, limit int) (list []*rankmdl.Result, err error)
	GetAdjust(c context.Context, baseID int64, rankID int64, objectType int) (list []*rankmdl.Adjust, err error)
	RankTag(c context.Context, baseID, rankID int64, batch int, tagID []int64) (list []*rankmdl.Result, err error)
	RankUp(c context.Context, baseID, rankID int64, batch int, mid []int64) (list []*rankmdl.Result, err error)
	UpdateRankArchive(c context.Context, tx *xsql.Tx, result *rankmdl.Result) (err error)
	UpdateRankUp(c context.Context, tx *xsql.Tx, result *rankmdl.Result) (err error)
	UpdateRankTag(c context.Context, tx *xsql.Tx, result *rankmdl.Result) (err error)
	UpdateRuleBatch(c context.Context, tx *xsql.Tx, ruleID int64, showBatch int, showBatchTime int64) (err error)
	BatchOidResult(c context.Context, tx *xsql.Tx, baseID int64, result []*rankmdl.ResultOid) (err error)
	BatchOidArchiveResult(c context.Context, tx *xsql.Tx, baseID int64, result []*rankmdl.ResultOidArchive) (err error)
	GetRankArchive(c context.Context, baseID, ruleID int64, batch int) (list []*rankmdl.ResultOidArchive, err error)
	GetRankOid(c context.Context, baseID, ruleID int64, batch int) (list []*rankmdl.ResultOid, err error)
	AddRankSourceEliminateOld(ctx context.Context, tx *xsql.Tx, baseID int64, sources []*rankmdl.Source) (err error)
	SourceListTotal(c context.Context, baseID int64) (total int, err error)
	GetSourceList(c context.Context, pn, ps int, baseID int64) (list []*rankmdl.Source, err error)
	UpdateRankRuleShow(c context.Context, ruleID int64, unit int, precision int, description string) (err error)

	BeginTran(c context.Context) (tx *xsql.Tx, err error)
}

// BeginTran begin transcation.
func (d *dao) BeginTran(c context.Context) (tx *xsql.Tx, err error) {
	return d.db.Begin(c)
}

// Dao dao.
type dao struct {
	c  *conf.Config
	db *xsql.DB
	DB *gorm.DB
	es *elastic.Elastic
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:  c,
		DB: orm.NewMySQL(c.ORM),
		es: elastic.NewElastic(c.EsClient),
		db: xsql.NewMySQL(c.MySQL.Lottery),
	}
	return
}

func (d *dao) GetDB() *gorm.DB {
	return d.DB
}

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return newDao(c)
}

// Close Dao
func (d *dao) Close() {
}
