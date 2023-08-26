package act

import (
	"context"
	"fmt"
	"strconv"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"

	"go-gateway/app/app-svr/app-view/interface/conf"
)

const (
	_actInfo      = "/matsuri/api/get/videoviewinfo"
	_lotteryTimes = "/matsuri/api/get/act/mylotterytimes"
)

// Dao is elec dao.
type Dao struct {
	client       *httpx.Client
	actInfo      string
	lotteryTimes string
	actGRPC      actgrpc.ActivityClient
}

// New elec dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:       httpx.NewClient(c.HTTPClient),
		actInfo:      c.Host.Activity + _actInfo,
		lotteryTimes: c.Host.Activity + _lotteryTimes,
	}
	var err error
	if d.actGRPC, err = actgrpc.NewClient(c.ActivityClient); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	return
}

// ActProtocol get act subject & protocol
func (d *Dao) ActProtocol(c context.Context, messionID int64) (protocol *actgrpc.ActSubProtocolReply, err error) {
	arg := &actgrpc.ActSubProtocolReq{Sid: messionID, HaveProtocol: 1}
	protocol, err = d.actGRPC.ActSubProtocol(c, arg)
	return
}

func (d *Dao) Reserve(c context.Context, id, mid, oid, action int64, spmid, from, typ string, dev device.Device) error {
	if action == 1 {
		req := &actgrpc.RelationReserveCancelReq{
			Id:       id,
			Mid:      mid,
			From:     from,
			Typ:      typ,
			Oid:      strconv.FormatInt(oid, 10),
			Ip:       metadata.String(c, metadata.RemoteIP),
			Platform: dev.RawPlatform,
			Mobiapp:  dev.RawMobiApp,
			Buvid:    dev.Buvid,
			Spmid:    spmid,
		}
		if _, err := d.actGRPC.RelationReserveCancel(c, req); err != nil {
			log.Error("d.actGRPC.RelationReserveCancel err(%+v) req(%+v)", err, req)
			return err
		}
		return nil
	}
	req := &actgrpc.GRPCDoRelationReq{
		Id:       id,
		Mid:      mid,
		From:     from,
		Typ:      typ,
		Oid:      strconv.FormatInt(oid, 10),
		Ip:       metadata.String(c, metadata.RemoteIP),
		Platform: dev.RawPlatform,
		Mobiapp:  dev.RawMobiApp,
		Buvid:    dev.Buvid,
		Spmid:    spmid,
	}
	if _, err := d.actGRPC.GRPCDoRelation(c, req); err != nil {
		log.Error("d.actGRPC.GRPCDoRelation err(%+v) req(%+v)", err, req)
		return err
	}
	return nil
}

func (d *Dao) IsReserveAct(c context.Context, id, mid int64) bool {
	req := &actgrpc.ActRelationInfoReq{
		Id:       id,
		Mid:      mid,
		Specific: "reserve"}
	res, err := d.actGRPC.ActRelationInfo(c, req)
	if err != nil {
		log.Error("d.actGRPC.GRPCDoRelation err(%+v) req(%+v)", err, req)
		return false
	}
	if res != nil && res.ReserveItems != nil && res.ReserveItems.State == 1 {
		return true
	}
	return false
}

func (d *Dao) LiveBooking(c context.Context, mid, upmid int64) (*actgrpc.UpActReserveRelationInfo, error) {
	res, err := d.actGRPC.UpActReserveRelationInfo4Live(c, &actgrpc.UpActReserveRelationInfo4LiveReq{Mid: mid, Upmid: upmid, From: 5})
	if err != nil {
		log.Error("LiveBooking error(%+v)", err)
		return nil, err
	}
	if res == nil || len(res.List) == 0 {
		return nil, ecode.NothingFound
	}
	return res.List[0], nil
}

// 通过稿件ID查询首映预约ID
func (d *Dao) GetPremiereSid(c context.Context, aid int64) (*actgrpc.GetPremiereSidByAidReply, error) {
	req := &actgrpc.GetPremiereSidByAidReq{
		Aid: aid,
	}
	res, err := d.actGRPC.GetPremiereSidByAid(c, req)
	if err != nil {
		log.Error("GetPremiereSid error(%+v)", err)
		return nil, err
	}
	return res, err
}

// 预约状态
func (d *Dao) ReserveState(c context.Context, sid, mid int64) (*actgrpc.ReserveFollowingReply, error) {
	req := &actgrpc.ReserveFollowingReq{
		Sid: sid,
		Mid: mid,
	}
	res, err := d.actGRPC.ReserveFollowing(c, req)
	if err != nil {
		log.Error("ReserveState error(%+v)", err)
		return nil, err
	}
	return res, err
}

// 取消预约
func (d *Dao) ReserveCancel(c context.Context, sid, mid int64) error {
	req := &actgrpc.DelReserveReq{
		Sid: sid,
		Mid: mid,
	}
	_, err := d.actGRPC.DelReserve(c, req)
	if err != nil {
		log.Error("ReserveCancel error(%+v)", err)
		return err
	}
	return nil
}

// 预约
func (d *Dao) AddReserve(c context.Context, sid, mid, upMid int64) error {
	req := &actgrpc.AddReserveReq{
		Sid:  sid,
		Mid:  mid,
		From: "player",
		Typ:  "release",
		Oid:  strconv.FormatInt(upMid, 10),
	}
	_, err := d.actGRPC.AddReserve(c, req)
	if err != nil {
		return err
	}
	return nil
}
