package share

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/view"

	shareApi "git.bilibili.co/bapis/bapis-go/community/interface/share"

	"github.com/pkg/errors"
)

// Dao is share dao
type Dao struct {
	// grpc
	shareGRPC shareApi.ShareClient
}

// New share dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.shareGRPC, err = shareApi.NewClient(c.ShareClient)
	if err != nil {
		panic(fmt.Sprintf("share NewClient error(%v)", err))
	}
	return
}

func (d *Dao) AddShareClick(c context.Context, params *view.ShareParam, mid, upID int64, buvid, apiType string, actData *shareApi.Metadata) (*shareApi.ServiceClickReply, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	var arg = &shareApi.ServiceClickReq{
		Oid:          params.OID,
		Mid:          mid,
		Type:         model.ShareTypeMap[params.Type],
		Ip:           ip,
		Channel:      params.ShareChannel,
		TraceId:      params.ShareTraceID,
		Client:       params.MobiApp,
		Buvid:        buvid,
		From:         params.From,
		Ssid:         params.SeasonID,
		Epid:         params.EpID,
		UpId:         upID,
		ParentAreaId: params.ParentAreaID,
		AreaId:       params.AreaID,
		Build:        params.Build,
		ApiType:      apiType,
		Spmid:        params.Spmid,
		FromSpmid:    params.FromSpmid,
		Metadata:     actData,
		NeedReport:   true,
		MobiApp:      params.MobiApp,
		Platform:     params.Platform,
		Device:       params.Device,
	}
	res, err := d.shareGRPC.ServiceClick(c, arg)
	if err != nil {
		log.Error("d.shareGRPC.AddShareClickCount err(%+v) arg(%+v)", err, arg)
		return nil, err
	}
	return res, nil
}

func (d *Dao) AddShareComplete(c context.Context, params *view.ShareParam, mid, upID int64, buvid string) (*shareApi.ServiceFinishReply, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	var arg = &shareApi.ServiceFinishReq{
		Oid:          params.OID,
		Mid:          mid,
		Type:         model.ShareTypeMap[params.Type],
		Ip:           ip,
		Channel:      params.ShareChannel,
		TraceId:      params.ShareTraceID,
		Client:       params.MobiApp,
		Buvid:        buvid,
		From:         params.From,
		Ssid:         params.SeasonID,
		Epid:         params.EpID,
		UpId:         upID,
		ParentAreaId: params.ParentAreaID,
		AreaId:       params.AreaID,
		Build:        params.Build,
	}
	res, err := d.shareGRPC.ServiceFinish(c, arg)
	if err != nil {
		log.Error("d.shareGRPC.AddShareCompletedCount err(%+v) arg(%+v)", err, arg)
		return nil, err
	}
	return res, nil
}

func (d *Dao) LastChannel(c context.Context, param *shareApi.LastChannelReq) (*shareApi.LastChannelReply, error) {
	reply, err := d.shareGRPC.LastChannel(c, param)
	if err != nil {
		return nil, errors.Wrapf(err, "d.shareGRPC.LastChannel param(%v)", param)
	}
	if reply == nil {
		return nil, ecode.NothingFound
	}
	return reply, nil
}
