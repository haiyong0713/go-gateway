package dao

import (
	"context"

	"go-common/library/database/sql"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewDB)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	InsertConfig(ctx context.Context, gs *model.GatewaySchedule) (err error)
	SelectConfigs(ctx context.Context) ([]*model.GatewaySchedule, error)
	SelectValueByKey(ctx context.Context, key string) (string, error)
	UpdateValueByKey(ctx context.Context, key string, value string) error
	SelectRuleId(ctx context.Context, role *model.CodeRule) (int64, error)
	InsertCodeRule(ctx context.Context, role *model.CodeRule) error
	DeleteCodeRule(ctx context.Context, ruleId int64) error
	GetUserService(ctx context.Context, username string) ([]string, error)
	GetPrimaryService(ctx context.Context, username string) ([]string, error)
	GetSecondaryService(ctx context.Context, username string) ([]string, error)
	InsertScript(ctx context.Context, script *model.Script) error
	GetUserScript(ctx context.Context, userid string) ([]*model.Script, error)
	GetScript(ctx context.Context, id string) (*model.Script, error)
	DeleteScript(ctx context.Context, id string) error
}

// dao dao.
type dao struct {
	db *sql.DB
}

// New new a dao and return.
func New(db *sql.DB) (d Dao, cf func(), err error) {
	return newDao(db)
}

//nolint:unparam
func newDao(db *sql.DB) (d *dao, cf func(), err error) {
	d = &dao{
		db: db,
	}
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {
	// TBD
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
