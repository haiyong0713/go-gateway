package act

import (
	"context"
	"net/url"
	"strconv"
	"time"

	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	xhttp "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"go-gateway/app/web-svr/native-page/interface/conf"
)

const (
	_updateActSubjectURI = "/x/admin/activity/subject/up"
)

type Dao struct {
	actClient           actGRPC.ActivityClient
	httpClient          *xhttp.Client
	updateActSubjectURL string
}

func New(cfg *conf.Config) *Dao {
	actClient, err := actGRPC.NewClient(cfg.ActClient)
	if err != nil {
		panic(err)
	}
	return &Dao{
		actClient:           actClient,
		httpClient:          xhttp.NewClient(cfg.HTTPActAdmin),
		updateActSubjectURL: cfg.Host.ActAdmin + _updateActSubjectURI,
	}
}

func (d *Dao) ReserveProgress(c context.Context, sid, mid, ruleID, typ, dataType int64, dimension actGRPC.GetReserveProgressDimension) (int64, error) {
	if sid == 0 {
		return 0, nil
	}
	req := &actGRPC.GetReserveProgressReq{
		Sid: sid,
		Mid: mid,
		Rules: []*actGRPC.ReserveProgressRule{
			{Dimension: dimension, RuleId: ruleID, Type: typ, DataType: dataType},
		},
	}
	rly, err := d.actClient.GetReserveProgress(c, req)
	if err != nil {
		log.Error("Fail to get reserveProgress, req=%+v error=%+v", req, err)
		return 0, err
	}
	for _, v := range rly.Data {
		if v == nil || v.Rule == nil {
			continue
		}
		if v.Rule.Dimension == dimension && v.Rule.RuleId == ruleID && v.Rule.Type == typ && v.Rule.DataType == dataType {
			return v.Progress, nil
		}
	}
	return 0, nil
}

// ReserveFollowings .
func (d *Dao) ReserveFollowings(c context.Context, mid int64, sids []int64) (res map[int64]*actGRPC.ReserveFollowingReply, err error) {
	var (
		rly *actGRPC.ReserveFollowingsReply
	)
	if rly, err = d.actClient.ReserveFollowings(c, &actGRPC.ReserveFollowingsReq{Sids: sids, Mid: mid}); err != nil {
		log.Error(" d.actRPC.ReserveFollowings(%v,%d) error(%v)", sids, mid, err)
		return
	}
	if rly != nil {
		res = rly.List
	}
	return
}

func (d *Dao) LotteryUnusedTimes(c context.Context, mid int64, lotteryID string) (*actGRPC.LotteryUnusedTimesReply, error) {
	return d.actClient.LotteryUnusedTimes(c, &actGRPC.LotteryUnusedTimesdReq{Sid: lotteryID, Mid: mid})
}

func (d *Dao) UpList(c context.Context, sid, pn, ps, mid int64, typ string) (*actGRPC.UpListReply, error) {
	req := &actGRPC.UpListReq{Sid: sid, Type: typ, Pn: pn, Ps: ps, Mid: mid}
	rly, err := d.actClient.UpList(c, req)
	if err != nil {
		log.Error("Fail to get upList, req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly, nil
}

func (d *Dao) ActivityProgress(c context.Context, sid, typ, mid int64, gids []int64) (*actGRPC.ActivityProgressReply, error) {
	req := &actGRPC.ActivityProgressReq{Sid: sid, Gids: gids, Type: typ, Mid: mid, Time: time.Now().Unix()}
	rly, err := d.actClient.ActivityProgress(c, req)
	if err != nil {
		log.Error("Fail to request ActivityProgress, req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly, nil
}

// UpActReserveRelationInfo.
func (d *Dao) UpActReserveRelationInfo(c context.Context, mid int64, sids []int64) (map[int64]*actGRPC.UpActReserveRelationInfo, error) {
	req := &actGRPC.UpActReserveRelationInfoReq{Sids: sids, Mid: mid}
	rly, err := d.actClient.UpActReserveRelationInfo(c, req)
	if err != nil {
		log.Error("Fail to request UpActReserveRelationInfo, req=%+v error=%+v", req, err)
		return nil, err
	}
	if rly == nil {
		return make(map[int64]*actGRPC.UpActReserveRelationInfo), nil
	}
	return rly.List, nil
}

func (d *Dao) ActRelationInfo(c context.Context, sid, mid int64) (*actGRPC.ActRelationInfoReply, error) {
	req := &actGRPC.ActRelationInfoReq{Id: sid, Mid: mid, Specific: "reserve"}
	rly, err := d.actClient.ActRelationInfo(c, req)
	if err != nil {
		log.Error("Fail to request actClient.ActRelationInfo(), req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly, nil
}

func (d *Dao) GetVoteActivityRank(c context.Context, actID, groupID, pn, ps, sort, mid int64) (*actGRPC.GetVoteActivityRankResp, error) {
	req := &actGRPC.GetVoteActivityRankReq{ActivityId: actID, SourceGroupId: groupID, Pn: pn, Ps: ps, Sort: sort, Mid: mid}
	rly, err := d.actClient.GetVoteActivityRank(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *Dao) OfflineActSubject(c context.Context, sid int64) error {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("id", strconv.FormatInt(sid, 10))
	params.Set("author", "up-sponsor")
	params.Set("etime", time.Now().Format("2006-01-02 15:04:05"))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var res struct {
		Code int   `json:"code"`
		Data int64 `json:"data"`
	}
	if err := d.httpClient.Post(c, d.updateActSubjectURL, ip, params, &res); err != nil {
		log.Error("Fail to request OfflineActSubject, req=%+v error=%+v", params.Encode(), err)
		return err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.updateActSubjectURL+"?"+params.Encode())
		log.Error("Fail to request OfflineActSubject, req=%+v error=%+v", params.Encode(), err)
		return err
	}
	return nil
}

func (d *Dao) ActSubProtocol(c context.Context, sid int64) (*actGRPC.ActSubProtocolReply, error) {
	rly, err := d.actClient.ActSubProtocol(c, &actGRPC.ActSubProtocolReq{Sid: sid})
	if err != nil {
		log.Error("Fail to request actGRPC.ActSubProtocol, sid=%d error=%+v", sid, err)
		return nil, err
	}
	return rly, nil
}

func (d *Dao) RankResult(c context.Context, id, pn, ps int64) (*actGRPC.RankResultResp, error) {
	return d.actClient.RankResult(c, &actGRPC.RankResultReq{RankID: id, Pn: pn, Ps: ps})
}

func (d *Dao) ActSubject(c context.Context, sid int64) (*actGRPC.ActSubjectReply, error) {
	rly, err := d.actClient.ActSubject(c, &actGRPC.ActSubjectReq{Sid: sid})
	if err != nil {
		log.Error("Fail to reqeust actGRPC.ActSubject, sid=%d error=%+v", sid, err)
		return nil, err
	}
	return rly, nil
}
