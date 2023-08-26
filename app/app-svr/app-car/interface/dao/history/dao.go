package history

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/queue/databus.v2"
	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/history"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
)

const (
	_hisMax         = 1200
	_sourceCarAudio = "car-audio"
)

// Dao is history dao
type Dao struct {
	hisClient        hisApi.HistoryClient
	fmReportClient   databus.Client
	fmReportProducer databus.Producer
}

// New initial history dao
func New(c *conf.Config) *Dao {
	d := &Dao{}
	var err error
	if d.hisClient, err = hisApi.NewClient(c.HistoryGRPC); err != nil {
		panic(fmt.Sprintf("hisApi.NewClient error (%+v)", err))
	}
	ctx := context.Background()
	if d.fmReportClient, d.fmReportProducer, err = NewProducer(ctx, c.FmReportMq); err != nil {
		panic(fmt.Sprintf("fmReport NewProducer error (%+v)", err))
	}
	return d
}

func (d *Dao) Close() {
	d.fmReportClient.Close()
}

func (d *Dao) HistoryCursor(c context.Context, mid, max, viewAt int64, ps int32, business, buvid string, businesses []string) ([]*hisApi.ModelResource, error) {
	var (
		arg = &hisApi.HistoryCursorReq{Mid: mid, Max: max, Ps: ps, Business: business, ViewAt: viewAt, Businesses: businesses, Ip: metadata.String(c, metadata.RemoteIP), Buvid: buvid}
	)
	reply, err := d.hisClient.HistoryCursor(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.Res, nil
}

func (d *Dao) HistoryCursorV2(c context.Context, mid, max, viewAt int64, ps int, business, buvid string, businesses []string) ([]*hisApi.ModelResource, error) {
	var (
		arg = &hisApi.HistoryCursorReq{
			Mid:        mid,
			Buvid:      buvid,
			Ps:         int32(ps),
			Max:        max,
			ViewAt:     viewAt,
			Business:   business,
			Businesses: businesses,
		}
	)
	reply, err := d.hisClient.HistoryCursor(c, arg)
	if err != nil {
		log.Error("HistoryCursorV2 d.hisClient.HistoryCursor error:%+v, arg:%+v", err, arg)
		return nil, err
	}
	return reply.Res, nil
}

func (d *Dao) HistoryCursorAll(c context.Context, mid, max int64, ps int32, business, buvid string, businesses []string, isAudio bool, build int) ([]*hisApi.ModelResource, error) {
	var (
		// 一次拉取所有的数据，然后额外筛选音频还是视频内容
		arg = &hisApi.HistoryCursorReq{Mid: mid, Max: 0, Ps: _hisMax, Business: "", ViewAt: 0, Businesses: businesses, Ip: metadata.String(c, metadata.RemoteIP), Buvid: buvid}
	)
	reply, err := d.hisClient.HistoryCursor(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	var (
		res []*hisApi.ModelResource
		ok  bool
	)
	if max == 0 {
		ok = true
	}
	for _, v := range reply.Res {
		if v.Business == business && v.Unix == max {
			ok = true
			// 跳过一次去掉相同的一个数据
			continue
		}
		if !ok {
			continue
		}
		// nolint:gomnd
		if build >= 1100000 {
			// 如果是音频模式过滤掉非音频的数据
			if isAudio && v.Source != _sourceCarAudio {
				continue
			}
			// 如果是非音频模式过滤掉音频的数据
			if !isAudio && v.Source == _sourceCarAudio {
				continue
			}
		}
		res = append(res, v)
		if len(res) >= int(ps) {
			break
		}
	}
	return res, nil
}

// Progress is  archive plays progress .
func (d *Dao) Progress(c context.Context, aid, mid int64, buvid string) (*hisApi.ModelHistory, error) {
	arg := &hisApi.ProgressReq{Mid: mid, Aids: []int64{aid}, Buvid: buvid}
	his, err := d.hisClient.Progress(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if his == nil {
		return nil, ecode.NothingFound
	}
	hi, ok := his.Res[aid]
	if !ok {
		return nil, ecode.NothingFound
	}
	return hi, nil

}

// BatchProgress 批量获取ugc与pgc的历史记录.
func (d *Dao) BatchProgress(c context.Context, mid int64, buvid string, aids, sids []int64) (map[int64]*hisApi.ModelHistory, map[int64]*hisApi.ModelHistory, error) {
	req := &hisApi.BatchProgressReq{
		Mid:   mid,
		Buvid: buvid,
		Metas: []*hisApi.BatchProgressReqMeta{
			{
				Business: model.UgcBusinesses,
				Kids:     aids,
			},
			{
				Business: model.PgcBusinesses,
				Kids:     sids,
			},
		},
	}
	reply, err := d.hisClient.BatchProgress(c, req)
	if err != nil {
		log.Error("BatchProgress d.hisClient.BatchProgress err=%+v, aids=%+v, sids=%+v", err, aids, sids)
		return nil, nil, err
	}
	ugcHistory := make(map[int64]*hisApi.ModelHistory)
	pgcHistory := make(map[int64]*hisApi.ModelHistory)
	if reply != nil {
		if reply.Res[model.UgcBusinesses] != nil {
			ugcHistory = reply.Res[model.UgcBusinesses].Records
		}
		if reply.Res[model.PgcBusinesses] != nil {
			pgcHistory = reply.Res[model.PgcBusinesses].Records
		}
	}
	return ugcHistory, pgcHistory, nil
}

func (d *Dao) Report(c context.Context, mid int64, buvid string, tp, dt int, param *history.ReportParam) error {
	arg := &hisApi.ReportReq{
		Mid:    mid,
		Buvid:  buvid,
		Oid:    param.Aid,
		Cid:    param.Cid,
		Sid:    param.SeasonId,
		Epid:   param.EpId,
		Tp:     int32(tp),
		Stp:    int64(param.SeasonType),
		Dt:     int32(dt),
		Pro:    param.Progress,
		ViewAt: param.Timestamp,
		Source: param.Source,
	}
	if _, err := d.hisClient.Report(c, arg); err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}
