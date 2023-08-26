package series

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	seriesgrpc "git.bilibili.co/bapis/bapis-go/platform/interface/series"
)

// Dao is series dao.
type Dao struct {
	c *conf.Config
	//grpc
	seriesClient seriesgrpc.SeriesClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	//grpc
	var err error
	if d.seriesClient, err = seriesgrpc.NewClient(c.SeriesClient); err != nil {
		panic(fmt.Sprintf("seriesgrpc NewClientt error (%+v)", err))
	}
	return
}

// ListSeries list series
func (d *Dao) ListSeries(c context.Context, mid, aid int64) (*seriesgrpc.ListSeriesResp, error) {
	arg := &seriesgrpc.ListSeriesReq{Mid: mid, State: seriesgrpc.SeriesOnline, Aid: aid}
	seriesReply, err := d.seriesClient.ListSeries(c, arg)
	if err != nil {
		return nil, err
	}
	return seriesReply, nil
}

// ListArchivesCursor list archives cursor
func (d *Dao) ListArchivesCursor(c context.Context, mid, seriesId, ps, next int64, sort string) (*seriesgrpc.ListArchivesCursorResp, error) {
	arg := &seriesgrpc.ListArchivesCursorReq{Mid: mid, SeriesId: seriesId, Cursor: &seriesgrpc.CursorReq{Ps: ps, Next: next}, Sort: sort}
	seriesReply, err := d.seriesClient.ListArchivesCursor(c, arg)
	if err != nil {
		return nil, err
	}
	return seriesReply, nil
}

// Series is
func (d *Dao) Series(c context.Context, seriesId int64) (*seriesgrpc.SeriesData, error) {
	arg := &seriesgrpc.SeriesReq{
		SeriesId: seriesId,
	}
	reply, err := d.seriesClient.Series(c, arg)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
