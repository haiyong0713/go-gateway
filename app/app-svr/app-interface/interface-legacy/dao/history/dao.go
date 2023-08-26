package history

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	. "go-gateway/app/app-svr/app-interface/interface-legacy/model/history"
	spm "go-gateway/app/app-svr/app-interface/interface-legacy/model/space"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"

	"github.com/pkg/errors"
)

// Dao is history dao
type Dao struct {
	client    *bm.Client
	rpcClient arcgrpc.ArchiveClient
	hisClient hisApi.HistoryClient
}

// New initial history dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: bm.NewClient(c.HTTPClient),
	}
	var err error
	if d.rpcClient, err = arcgrpc.NewClient(c.ArchiveGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClient error (%+v)", err))
	}
	if d.hisClient, err = hisApi.NewClient(c.HistoryGRPC); err != nil {
		panic(fmt.Sprintf("hisApi.NewClient error (%+v)", err))
	}
	return
}

// History get history
func (d *Dao) History(c context.Context, mid int64, pn, ps int32) (res []*hisApi.ModelResource, err error) {
	var (
		arg = &hisApi.HistoriesReq{Mid: mid, Pn: pn, Ps: ps}
		rep *hisApi.HistoriesReply
	)
	if rep, err = d.hisClient.Histories(c, arg); err != nil {
		err = errors.Wrapf(err, "d.hisClient.Histories(%+v)", arg)
		return
	}
	if rep != nil {
		res = rep.Res
	}
	return
}

// Archive get archive info
func (d *Dao) Archive(c context.Context, aids []int64) (info map[int64]*arcgrpc.ViewReply, err error) {
	var (
		viewReply *arcgrpc.ViewsReply
	)
	arg := &arcgrpc.ViewsRequest{Aids: aids}
	if viewReply, err = d.rpcClient.Views(c, arg); err != nil {
		err = errors.Wrapf(err, "d.rpcClient.Views(%+v)", arg)
		return
	}
	info = viewReply.Views
	return
}

// HistoryByTP histroy by tp
func (d *Dao) HistoryByTP(c context.Context, mid int64, pn, ps int32, business string) (res []*hisApi.ModelResource, err error) {
	var (
		arg = &hisApi.HistoriesReq{Mid: mid, Pn: pn, Ps: ps, Business: business}
		rep *hisApi.HistoriesReply
	)
	if rep, err = d.hisClient.Histories(c, arg); err != nil {
		err = errors.Wrapf(err, "d.hisClient.Histories(%+v)", arg)
	}
	if rep != nil {
		res = rep.Res
	}
	return
}

// Cursor 5.28游标由MaxOid+MaxTP唯一确定 改为 由ViewAt唯一确定（防止客户端改动对客户端仍用max字段）
func (d *Dao) Cursor(c context.Context, mid, max int64, ps int32, business string, businesses []string, buvid string) (res []*hisApi.ModelResource, err error) {
	var (
		arg = &hisApi.HistoryCursorReq{Mid: mid, Max: max, Ps: ps, Business: business, ViewAt: max, Businesses: businesses, Buvid: buvid}
		rep *hisApi.HistoryCursorReply
	)
	if rep, err = d.hisClient.HistoryCursor(c, arg); err != nil {
		err = errors.Wrapf(err, "d.hisClient.HistoryCursor(%+v)", arg)
	}
	if rep != nil {
		res = rep.Res
	}
	return
}

func (d *Dao) NativeHistory(c context.Context, mid int64, businesses []string, buvid string, deviceType int8, max int64, ps int32, business string) ([]*hisApi.ModelResource, error) {
	rep, err := d.hisClient.NativeHistory(c, &hisApi.NativeHistoryReq{Mid: mid, Businesses: businesses, Buvid: buvid, DeviceType: int32(deviceType), Max: max, Ps: ps, ViewAt: max, Business: business})
	if err != nil {
		return nil, err
	}
	if rep == nil {
		return nil, ecode.NothingFound
	}
	return rep.Res, nil
}

// Del for history
func (d *Dao) Del(c context.Context, mid int64, hisRes []*hisApi.ModelHistory, buvid string) (err error) {
	arg := &hisApi.DeleteReq{Mid: mid, His: hisRes, Buvid: buvid}
	if _, err = d.hisClient.Delete(c, arg); err != nil {
		err = errors.Wrapf(err, "d.hisClient.Delete(%+v)", arg)
	}
	return
}

// Clear for history
func (d *Dao) Clear(c context.Context, mid int64, businesses []string, buvid string) (err error) {
	arg := &hisApi.ClearHistoryReq{Mid: mid, Businesses: businesses, Buvid: buvid}
	if _, err = d.hisClient.ClearHistory(c, arg); err != nil {
		err = errors.Wrapf(err, "d.hisClient.ClearHistory(%+v)", arg)
	}
	return
}

func (d *Dao) Position(c context.Context, mid, uid int64, business string) (*spm.HistoryPosition, error) {
	arg := &hisApi.PositionReq{Mid: mid, Aid: uid, Business: business}
	reply, err := d.hisClient.Position(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "d.hisClient.ClearHistory(%+v)", arg)
		return nil, err
	}
	position := &spm.HistoryPosition{}
	if reply != nil && reply.Res != nil {
		position.Offset = int(reply.Res.Sid)
		position.Desc = int(reply.Res.Cid)
		position.Oid = int(reply.Res.Epid)
		return position, nil
	}
	return nil, errors.New("d.hisClient.Position reply is nil")
}

// Search 历史记录支持搜索
func (d *Dao) Search(c context.Context, mid int64, pn, ps int32, keyword string, businesses []string, buvid string) (res []*hisApi.ModelResource, total int32, err error) {
	var (
		arg = &hisApi.SearchHistoryReq{Mid: mid, Key: keyword, Pn: pn, Ps: ps, Businesses: businesses, Buvid: buvid}
		rep *hisApi.SearchHistoryReply
	)
	if rep, err = d.hisClient.SearchHistory(c, arg); err != nil {
		err = errors.Wrapf(err, "d.hisClient.SearchHistory(%+v)", arg)
	}
	if rep != nil {
		res = rep.Res
		total = rep.Total
	}
	return
}

// HasHistory is 历史记录是否存在
func (d *Dao) HasHistory(c context.Context, query *SearchQuery) (map[string]bool, error) {
	req := &hisApi.HasHistoryReq{Mid: query.Mid, Businesses: query.Businesses, Buvid: query.Buvid}
	rep, err := d.hisClient.HasHistory(c, req)
	if err != nil {
		return nil, err
	}
	if rep == nil {
		return nil, errors.New("HasHistory rep is nil")
	}
	return rep.Res, nil
}

// SearchHasHistory is 搜索历史记录是否存在
func (d *Dao) SearchHasHistory(c context.Context, query *SearchQuery) (map[string]bool, error) {
	req := &hisApi.SearchHasHistoryReq{Mid: query.Mid, Businesses: query.Businesses, Buvid: query.Buvid, Key: query.Keyword}
	rep, err := d.hisClient.SearchHasHistory(c, req)
	if err != nil {
		return nil, err
	}
	if rep == nil {
		return nil, errors.New("SearchHasHistory rep is nil")
	}
	return rep.Res, nil
}

func (d *Dao) GetHistoryFrequent(ctx context.Context, req *hisApi.HistoryFrequentReq) (*hisApi.HistoryFrequentReply, error) {
	return d.hisClient.HistoryFrequent(ctx, req)
}
