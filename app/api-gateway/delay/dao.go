package delay

import (
	"context"
	"errors"

	"go-common/library/database/boss"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	MergeStep(ctx context.Context, apiName, loadPath string) (err error)
	Upload(ctx context.Context, prefix, filename string, fileData []byte) (res string, err error)
	ReadTarPackage(destPath string) (res []byte, err error)
	Compress(filePath string, destPath string) (err error)

	AddRowDB(ctx context.Context, apiName, version string) (id int64, err error)
	GetLatestWF(ctx context.Context, apiName string) (res *WFDetail, err error)
	UpdateLog(ctx context.Context, id int64, dName string, dState int8, logs string) (err error)
	UpdateBoss(ctx context.Context, id int64, boss string, dName string, dState int8) (err error)
	UpdateWFName(ctx context.Context, id int64, wfName string) (err error)
	UpdateWFDis(ctx context.Context, id int64, discoveryID string) (err error)
	UpdateWFImage(ctx context.Context, id int64, wfName string) (err error)
	UpdateWFState(ctx context.Context, id int64, state int) (err error)
	UpdateWFDisplay(ctx context.Context, id int64, dName string, dState int8) (err error)
	UpdateFailedWF(ctx context.Context, id int64, dName string, dState, state int8, logs string) (err error)
	GetWFByApi(ctx context.Context, apiName string) (res []*WFDetail, err error)
	GetAllWF(ctx context.Context) (res []*WFDetail, err error)

	CreateWorkflow(c context.Context, apiName, codeAddress, codeVersion, imageAddr string, dt DeployType) (name, url string, err error)
	GetWorkflowStatus(c context.Context, name string) (phase string, displayName string, err error)
	OutputWorkflow(c context.Context, name, displayName string) (text OutputsData, err error)
	ResumeWorkflow(c context.Context, name, displayName string) (err error)
	GetLogWorkflow(c context.Context, name, displayName string) (text string, err error)
	StopWorkflow(c context.Context, name string) (err error)
}

// dao dao.
type dao struct {
	httpCli *bm.Client
	host    *Host
	boss    *boss.Boss
	db      *sql.DB
}

type Host struct {
	Workflow string
	Boss     string
}

type Cfg struct {
	Client *bm.ClientConfig
	Host   *Host
	Boss   *boss.Config
	DB     *sql.Config
}

func NewDao(cfg Cfg) (d *dao, err error) {
	if cfg.DB == nil || cfg.Client == nil || cfg.Boss == nil || cfg.Host == nil {
		return nil, errors.New("wrong cfg")
	}
	d = &dao{
		httpCli: bm.NewClient(cfg.Client),
		host:    cfg.Host,
		boss:    boss.New(cfg.Boss),
		db:      sql.NewMySQL(cfg.DB),
	}
	return
}

// Close close the resource.
func (d *dao) Close() {
	_ = d.db.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
