package archive

import (
	"context"
	"fmt"
	"go-common/library/stat/prom"

	"github.com/golang/protobuf/ptypes/empty"

	"go-common/library/database/tidb"
	"go-gateway/app/app-svr/archive/job/conf"

	"github.com/pkg/errors"

	vuapi "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

// Dao is redis dao.
type Dao struct {
	c           *conf.Config
	tidb        *tidb.DB
	videoupGRPC vuapi.VideoUpOpenClient
	errProm     *prom.Prom
	infoProm    *prom.Prom
}

// New is new redis dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:    c,
		tidb: tidb.NewTiDB(c.DB.ArchiveTiDB),
		//迁移逻辑
		errProm:  prom.BusinessErrCount,
		infoProm: prom.BusinessInfoCount,
	}
	var err error
	d.videoupGRPC, err = vuapi.NewClient(c.VideoupClient)
	if err != nil {
		panic(fmt.Sprintf("videoup NewClient error(%v)", err))
	}

	return d
}

func (d *Dao) ArchivesDelay(c context.Context, aid int64) (res *vuapi.ArchivesDelayReply, err error) {
	if res, err = d.videoupGRPC.ArchivesDelay(c, &vuapi.ArchivesDelayReq{Aids: []int64{aid}}); err != nil {
		err = errors.Wrapf(err, "d.videoupGRPC.ArcViewAddit err aid(%d)", aid)
		return
	}
	return
}

// archive_addit表
func (d *Dao) GetArchiveAddit(ctx context.Context, aid int64) (*vuapi.GetArchiveAdditReply, error) {
	req := &vuapi.GetArchiveAdditReq{
		Aid: aid,
	}
	res, err := d.videoupGRPC.GetArchiveAddit(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "d.videoupGRPC.GetArchiveAddit is error %+v", req)
	}
	return res, err
}

// archive表
func (d *Dao) GetArchive(ctx context.Context, aid int64) (*vuapi.GetArchiveReply, error) {
	req := &vuapi.GetArchiveReq{
		Aid: aid,
	}
	res, err := d.videoupGRPC.GetArchive(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// archive_biz表
func (d *Dao) GetArchiveBiz(ctx context.Context, aid int64, state int, bizType int) (*vuapi.GetArchiveBizReply, error) {
	req := &vuapi.GetArchiveBizReq{
		Aid:   aid,
		State: int64(state),
		Tp:    int64(bizType),
	}
	res, err := d.videoupGRPC.GetArchiveBiz(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) GetArchiveFirstPass(ctx context.Context, aid int64) (*vuapi.GetArchiveFirstPassReply, error) {
	req := &vuapi.GetArchiveFirstPassReq{
		Aid: aid,
	}
	res, err := d.videoupGRPC.GetArchiveFirstPass(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// archive表
func (d *Dao) GetArchiveStaff(ctx context.Context, aid int64) (*vuapi.GetArchiveStaffReply, error) {
	req := &vuapi.GetArchiveStaffReq{
		Aid: aid,
	}
	res, err := d.videoupGRPC.GetArchiveStaff(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "d.videoupGRPC.GetArchiveStaff is error %+v", req)
	}
	return res, nil
}

// archive_type表
func (d *Dao) GetArchiveType(ctx context.Context) (map[int64]*vuapi.ArchiveType, error) {
	res, err := d.videoupGRPC.GetArchiveType(ctx, &empty.Empty{})
	if err != nil {
		return nil, errors.Wrapf(err, "d.videoupGRPC.GetArchiveType is error %+v", err)
	}
	return res.ArchiveType, nil
}

func (d *Dao) GetArchiveVideoRelation(ctx context.Context, aid int64) ([]*vuapi.ArchiveVideoRelation, error) {
	req := &vuapi.GetArchiveVideoRelationReq{
		Aid: aid,
	}
	res, err := d.videoupGRPC.GetArchiveVideoRelation(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "d.videoupGRPC.GetArchiveVideoRelation is error %+v", err)
	}
	return res.Relation, nil
}

func (d *Dao) GetArchiveVideoShot(ctx context.Context, cids []int64) (map[int64]*vuapi.ArchiveVideoShot, error) {
	req := &vuapi.GetArchiveVideoShotReq{
		Ids: cids,
	}
	res, err := d.videoupGRPC.GetArchiveVideoShot(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "d.videoupGRPC.GetArchiveVideoShot is error %+v", err)
	}
	return res.VideoShot, nil
}

// Close close connection of db , mc.
func (d *Dao) Close() {
	if d.tidb != nil {
		d.tidb.Close()
	}
}
