package account

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	mdlaccount "go-gateway/app/app-svr/app-dynamic/interface/model/account"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"

	api "git.bilibili.co/bapis/bapis-go/account/service"
	membergrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	relagrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	passportgrpc "git.bilibili.co/bapis/bapis-go/passport/service/user"
	"github.com/pkg/errors"
)

const (
	_decoCardsURL = "/x/internal/garb/user/card/multi"
)

// Dao is account dao.
type Dao struct {
	c       *conf.Config
	httpCli *bm.Client

	accApi       api.AccountClient
	relaGRPC     relagrpc.RelationClient
	memberGRPC   membergrpc.MemberClient
	passportGRPC passportgrpc.PassportUserClient
}

// New account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:       c,
		httpCli: bm.NewClient(c.HTTPClient),
	}
	var err error
	if d.accApi, err = api.NewClient(c.AccountGRPC); err != nil {
		panic(fmt.Sprintf("accountGRPC error(%v)", err))
	}
	if d.relaGRPC, err = relagrpc.NewClient(c.RelaGRPC); err != nil {
		panic(err)
	}
	if d.memberGRPC, err = membergrpc.NewClient(c.MemberGRPC); err != nil {
		panic(err)
	}
	if d.passportGRPC, err = passportgrpc.NewClient(c.PassportGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) IsAttention(c context.Context, owners []int64, mid int64) (isAtten map[int64]int32) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]int32)
	for i := 0; i < len(owners); i += max50 {
		var partUids []int64
		if i+max50 > len(owners) {
			partUids = owners[i:]
		} else {
			partUids = owners[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			as := d.IsAttentionSlice(ctx, partUids, mid)
			mu.Lock()
			for uid, a := range as {
				res[uid] = a
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("IsAttention owners(%+v) eg.wait(%+v)", owners, err)
		return nil
	}
	return res
}

// IsAttention is attention
func (d *Dao) IsAttentionSlice(c context.Context, owners []int64, mid int64) (isAtten map[int64]int32) {
	if len(owners) == 0 || mid == 0 {
		return
	}
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &api.RelationsReq{Owners: owners, Mid: mid, RealIp: ip}
	res, err := d.accApi.Relations3(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	isAtten = make(map[int64]int32, len(res.Relations))
	for mid, rel := range res.Relations {
		if rel.Following {
			isAtten[mid] = 1
		}
	}
	return
}

func (d *Dao) Cards3(c context.Context, uids []int64) (*api.CardsReply, error) {
	cardReply, err := d.accApi.Cards3(c, &api.MidsReq{Mids: uids})
	if err != nil || cardReply == nil {
		log.Error("Failed to call Cards3(). uids: %+v. error: %+v", uids, errors.WithStack(err))
		return nil, err
	}
	return cardReply, nil
}

func (d *Dao) Cards3New(c context.Context, uids []int64) (map[int64]*api.Card, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*api.Card)
	for i := 0; i < len(uids); i += max50 {
		var partUids []int64
		if i+max50 > len(uids) {
			partUids = uids[i:]
		} else {
			partUids = uids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			cs, err := d.Cards3Slice(ctx, partUids)
			if err != nil {
				return err
			}
			mu.Lock()
			for uid, c := range cs {
				res[uid] = c
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("Cards3 uids(%+v) eg.wait(%+v)", uids, err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) Profile3(c context.Context, uid int64) (*api.Profile, error) {
	reply, err := d.accApi.Profile3(c, &api.MidReq{Mid: uid})
	if err != nil {
		log.Error("Failed to call Profile3(). uids: %+v. error: %+v", uid, errors.WithStack(err))
		return nil, err
	}
	return reply.GetProfile(), nil
}

func (d *Dao) ProfileWithStat3(c context.Context, mid int64) (*api.ProfileStatReply, error) {
	res, err := d.accApi.ProfileWithStat3(c, &api.MidReq{Mid: mid})
	if err != nil {
		log.Errorc(c, "Failed to call ProfileWithStat3(). uids: %v. error: %v", mid, errors.WithStack(err))
		return nil, err
	}
	return res, nil
}

func (d *Dao) Cards3Slice(c context.Context, uids []int64) (map[int64]*api.Card, error) {
	cardReply, err := d.accApi.Cards3(c, &api.MidsReq{Mids: uids})
	if err != nil || cardReply == nil {
		log.Error("Failed to call Cards3(). uids: %+v. error: %+v", uids, errors.WithStack(err))
		return nil, err
	}
	return cardReply.GetCards(), nil
}

func (d *Dao) DecorateCards(c context.Context, uids []int64) (map[int64]*mdlaccount.DecoCards, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*mdlaccount.DecoCards)
	for i := 0; i < len(uids); i += max50 {
		var partUids []int64
		if i+max50 > len(uids) {
			partUids = uids[i:]
		} else {
			partUids = uids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			dcs, err := d.DecorateCardsSlice(ctx, partUids)
			if err != nil {
				return err
			}
			mu.Lock()
			for uid, dc := range dcs {
				res[uid] = dc
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("DecorateCards uids(%+v) eg.wait(%+v)", uids, err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) DecorateCardsSlice(c context.Context, uids []int64) (map[int64]*mdlaccount.DecoCards, error) {
	params := url.Values{}
	params.Set("mids", xstr.JoinInts(uids))
	decoCard := d.c.Hosts.ApiCo + _decoCardsURL
	var ret struct {
		Code int                             `json:"code"`
		Msg  string                          `json:"message"`
		Data map[int64]*mdlaccount.DecoCards `json:"data"`
	}
	if err := d.httpCli.Get(c, decoCard, "", params, &ret); err != nil {
		xmetric.DyanmicItemAPI.Inc(decoCard, "request_error")
		log.Errorc(c, "PGCBatch http GET(%s) failed, params:(%s), error(%+v)", decoCard, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		xmetric.DyanmicItemAPI.Inc(decoCard, "reply_code_error")
		log.Errorc(c, "PGCBatch http GET(%s) failed, params:(%s), code: %v, msg: %v", decoCard, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "PGCBatch url(%v) code(%v) msg(%v)", decoCard, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) Followings(c context.Context, uid int64) (*relagrpc.FollowingsReply, error) {
	following, err := d.relaGRPC.Attentions(c, &relagrpc.MidReq{Mid: uid})
	if err != nil {
		err = errors.WithStack(err)
		log.Errorc(c, "Failed to call Attentions(). uid: %v. error: %+v", uid, err)
		return nil, err
	}
	return following, nil
}

// Interrelations
func (d *Dao) Interrelations(ctx context.Context, mid int64, owners []int64) (res map[int64]*relagrpc.InterrelationReply, err error) {
	fidsMap := make(map[int64]int64)
	fids := []int64{}
	for _, fid := range owners {
		if _, ok := fidsMap[fid]; ok {
			continue
		}
		fidsMap[fid] = fid
		fids = append(fids, fid)
	}
	const _max = 20
	g := errgroup.WithContext(ctx)
	mu := sync.Mutex{}
	res = make(map[int64]*relagrpc.InterrelationReply)
	for i := 0; i < len(fids); i += _max {
		var partFids []int64
		if i+_max > len(fids) {
			partFids = fids[i:]
		} else {
			partFids = fids[i : i+_max]
		}
		g.Go(func(ctx context.Context) (err error) {
			var (
				reply *relagrpc.InterrelationMapReply
				arg   = &relagrpc.RelationsReq{
					Mid: mid,
					Fid: partFids,
				}
			)
			if reply, err = d.relaGRPC.Interrelations(ctx, arg); err != nil {
				log.Error("d.relGRPC.Interrelations(%v) error(%v)", arg, err)
				return nil
			}
			if reply == nil {
				return nil
			}
			mu.Lock()
			for k, v := range reply.InterrelationMap {
				res[k] = v
			}
			mu.Unlock()
			return nil
		})
	}
	err = g.Wait()
	return
}

func (d *Dao) School(ctx context.Context, general *mdlv2.GeneralParam) (*membergrpc.SchoolReply, error) {
	arg := &membergrpc.MidReq{
		Mid:    general.Mid,
		RealIP: general.IP,
	}
	reply, err := d.memberGRPC.School(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) NFTBatchInfo(ctx context.Context, mids []int64) (map[int64]*membergrpc.NFTBatchInfoData, error) {
	const _max = 20
	res := map[int64]*membergrpc.NFTBatchInfoData{}
	g := errgroup.WithContext(ctx)
	mu := sync.Mutex{}
	for i := 0; i < len(mids); i += _max {
		var partUids []int64
		if i+_max > len(mids) {
			partUids = mids[i:]
		} else {
			partUids = mids[i : i+_max]
		}
		g.Go(func(c context.Context) error {
			arg := &membergrpc.NFTBatchInfoReq{
				Mids:   partUids,
				Status: "inUsing",
				Source: "face",
			}
			reply, err := d.memberGRPC.NFTBatchInfo(c, arg)
			if err != nil {
				return err
			}
			mu.Lock()
			for k, v := range reply.GetNftInfos() {
				tmid, err := strconv.ParseInt(k, 10, 64)
				if err != nil {
					continue
				}
				res[tmid] = v
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}

var (
	fixedUserLocCache = make(map[int64]string)
	lastUpdateTime    time.Time
	fixedUserLocRw    sync.RWMutex
)

const (
	_fixedUserLocationUpdateInterval = 5 * time.Minute
)

func (d *Dao) FixedUserLocation(ctx context.Context) map[int64]string {
	fixedUserLocRw.RLock()
	defer fixedUserLocRw.RUnlock()
	if time.Since(lastUpdateTime) > _fixedUserLocationUpdateInterval {
		d.updateFixedUserLocation(ctx)
	}
	return fixedUserLocCache
}

func (d *Dao) updateFixedUserLocation(ctx context.Context) {
	fixedUserLocRw.RUnlock()
	defer fixedUserLocRw.RLock()
	fixedUserLocRw.Lock()
	defer fixedUserLocRw.Unlock()
	// 再次检查时间 避免重复请求
	if time.Since(lastUpdateTime) <= _fixedUserLocationUpdateInterval {
		return
	}
	resp, err := d.passportGRPC.UserFixedLocations(ctx, &passportgrpc.UserFixedLocationsReq{})
	if err != nil {
		log.Errorc(ctx, "error update passportGRPC.UserFixedLocations: %v", err)
		return
	}
	lastUpdateTime = time.Now()
	if resp.GetFixedLocations() != nil {
		fixedUserLocCache = resp.GetFixedLocations()
	} else {
		fixedUserLocCache = make(map[int64]string) // 清空数据
	}
}

func (d *Dao) UserFrequentLoc(ctx context.Context, uids map[int64]struct{}) (map[int64]*passportgrpc.UserActiveLocationReply, error) {
	uidslc := make([]int64, 0, len(uids))
	for uid := range uids {
		uidslc = append(uidslc, uid)
	}
	const (
		_maxIds = 50
	)
	eg := errgroup.WithCancel(ctx)
	mu := sync.Mutex{}
	ret := make(map[int64]*passportgrpc.UserActiveLocationReply)
	for i := 0; i < len(uidslc); i += _maxIds {
		var partUids []int64
		if i+_maxIds > len(uidslc) {
			partUids = uidslc[i:]
		} else {
			partUids = uidslc[i : i+_maxIds]
		}
		eg.Go(func(ctx context.Context) error {
			tmpRes, err := d.passportGRPC.UserActiveLocations(ctx, &passportgrpc.MidsReq{Mids: partUids})
			if err != nil {
				return err
			}
			if len(tmpRes.GetLocations()) > 0 {
				mu.Lock()
				for k, v := range tmpRes.GetLocations() {
					ret[k] = v
				}
				mu.Unlock()
			}
			return nil
		})
	}
	return ret, eg.Wait()
}
