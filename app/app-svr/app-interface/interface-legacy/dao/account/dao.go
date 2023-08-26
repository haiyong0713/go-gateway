package account

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	memberAPI "git.bilibili.co/bapis/bapis-go/account/service/member"
	shuangqing "git.bilibili.co/bapis/bapis-go/datacenter/shuangqing"
	passportuser "git.bilibili.co/bapis/bapis-go/passport/service/user"
	garbgrpc "git.bilibili.co/bapis/bapis-go/vas/garb/live2d/service"

	"github.com/pkg/errors"
)

func staticKVConfig() map[string]*tableConfig {
	return map[string]*tableConfig{
		"1d": {
			Table: "shuangqing_aggr_taishan1",
			Token: "IlGPpoXApqGiW2n5",
			Zone:  "sh004",
		},
		"7d": {
			Table: "shuangqing_aggr_taishan7",
			Token: "IlGPpoXApqGiW2n5",
			Zone:  "sh004",
		},
		"30d": {
			Table: "shuangqing_aggr_taishan30",
			Token: "IlGPpoXApqGiW2n5",
			Zone:  "sh004",
		},
	}
}

// Dao is account dao.
type Dao struct {
	client *bm.Client
	// rpc
	accGRPC             accgrpc.AccountClient
	memberRPC           memberAPI.MemberClient
	grabGRPC            garbgrpc.GarbCharacterClient
	staticKV            *Taishan
	shuangQingStaticKvs map[string]*Taishan
	passportUser        passportuser.PassportUserClient
}

// New account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: bm.NewClient(c.HTTPClient),
	}
	var err error
	if d.memberRPC, err = memberAPI.NewClient(c.MemClient); err != nil {
		panic(err)
	}
	if d.accGRPC, err = accgrpc.NewClient(c.AccountGRPC); err != nil {
		panic(err)
	}
	if d.grabGRPC, err = garbgrpc.NewClient(c.GrabClient); err != nil {
		panic(err)
	}
	if d.shuangQingStaticKvs, _, err = NewKVs(c, staticKVConfig()); err != nil {
		panic(err)
	}
	if d.passportUser, err = passportuser.NewClient(c.PassportUser); err != nil {
		panic(err)
	}
	return
}

// BlockTime get user blocktime
func (d *Dao) BlockTime(c context.Context, mid int64) (blockTime int64, err error) {
	info, err := d.memberRPC.BlockInfo(c, &memberAPI.MemberMidReq{Mid: mid})
	if err != nil {
		err = errors.Wrapf(err, "%v", mid)
		return
	}
	if info.EndTime > 0 {
		blockTime = info.EndTime
	}
	return
}

// Profile3 get profile
func (d *Dao) Profile3(c context.Context, mid int64) (card *accgrpc.ProfileStatReply, err error) {
	arg := &accgrpc.MidReq{Mid: mid}
	if card, err = d.accGRPC.ProfileWithStat3(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
	}
	return
}

func (d *Dao) ProfilesWithoutPrivacy3(c context.Context, mids []int64) (map[int64]*accgrpc.ProfileWithoutPrivacy, error) {
	arg := &accgrpc.MidsReq{Mids: mids}
	card, err := d.accGRPC.ProfilesWithoutPrivacy3(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "%+v", arg)
	}
	return card.ProfilesWithoutPrivacy, nil
}

// Card get card
func (d *Dao) Card(c context.Context, mid int64) (card *accgrpc.Card, err error) {
	var cardTmp *accgrpc.CardReply
	if cardTmp, err = d.accGRPC.Card3(c, &accgrpc.MidReq{Mid: mid}); err != nil {
		err = errors.Wrapf(err, "%v", mid)
		return
	}
	card = cardTmp.Card
	return
}

// ProfileByName3 rpc card get by name
func (d *Dao) ProfileByName3(c context.Context, name string) (card *accgrpc.ProfileStatReply, err error) {
	var infos map[int64]*accgrpc.Info
	infosTmp, err := d.accGRPC.InfosByName3(c, &accgrpc.NamesReq{Names: []string{name}})
	if err != nil {
		err = errors.Wrapf(err, "%v", name)
		return
	}
	infos = infosTmp.Infos
	if len(infos) == 0 {
		err = ecode.NothingFound
		return
	}
	for mid := range infos {
		card, err = d.Profile3(c, mid)
		break
	}
	return
}

// Infos3 rpc info get by mids .
func (d *Dao) Infos3(c context.Context, mids []int64) (res map[int64]*accgrpc.Info, err error) {
	var resTmp *accgrpc.InfosReply
	arg := &accgrpc.MidsReq{Mids: mids}
	if resTmp, err = d.accGRPC.Infos3(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = resTmp.Infos
	return
}

func (d *Dao) CheckRegTime(ctx context.Context, req *accgrpc.CheckRegTimeReq) bool {
	res, err := d.accGRPC.CheckRegTime(ctx, req)
	if err != nil {
		log.Error("d.accGRPC.CheckRegTime req=%+v", req)
		return false
	}
	return res.GetHit()
}

// Relations3 relations.
func (d *Dao) Relations3(c context.Context, owners []int64, mid int64) (follows map[int64]bool) {
	if len(owners) == 0 {
		return nil
	}
	follows = make(map[int64]bool, len(owners))
	var (
		am        *accgrpc.RelationsReply
		err       error
		ip        = metadata.String(c, metadata.RemoteIP)
		ownersMap = make(map[int64]struct{}, len(owners))
		ownerLeft []int64
	)
	for _, owner := range owners {
		if _, ok := ownersMap[owner]; ok {
			continue
		}
		follows[owner] = false
		ownersMap[owner] = struct{}{}
		ownerLeft = append(ownerLeft, owner)
	}
	arg := &accgrpc.RelationsReq{Owners: ownerLeft, Mid: mid, RealIp: ip}
	if am, err = d.accGRPC.Relations3(c, arg); err != nil {
		log.Error("d.accRPC.Relations2(%v) error(%v)", arg, err)
		return
	}
	for i, a := range am.Relations {
		if _, ok := follows[i]; ok {
			follows[i] = a.Following
		}
	}
	return
}

// RichRelations3 rich relations.
func (d *Dao) RichRelations3(c context.Context, owner, mid int64) (rel int, err error) {
	var (
		res *accgrpc.RichRelationsReply
		ip  = metadata.String(c, metadata.RemoteIP)
	)
	arg := &accgrpc.RichRelationReq{Mids: []int64{mid}, Owner: owner, RealIp: ip}
	if res, err = d.accGRPC.RichRelations3(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if res != nil {
		if r, ok := res.RichRelations[mid]; ok {
			rel = int(r)
		}
	}
	return
}

// Cards3 is
func (d *Dao) Cards3(c context.Context, mids []int64) (res map[int64]*accgrpc.Card, err error) {
	var cardTmp *accgrpc.CardsReply
	arg := &accgrpc.MidsReq{Mids: mids}
	if cardTmp, err = d.accGRPC.Cards3(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = cardTmp.Cards
	return
}

// UserCheck 各种入口白名单
// https://www.tapd.cn/20055921/prong/stories/view/1120055921001066980  动态互推TAPD在此！！
func (d *Dao) UserCheck(c context.Context, mid int64, checkURL string) (ok bool, err error) {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Status int `json:"status"`
		} `json:"data"`
	}
	if err = d.client.Get(c, checkURL, "", params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), checkURL+"?"+params.Encode())
		return
	}
	if res.Data.Status == 1 {
		ok = true
	}
	return
}

// RedDot 我的页小红点逻辑
func (d *Dao) RedDot(c context.Context, mid int64, redDotURL string) (ok bool, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			RedDot bool `json:"red_dot"`
		} `json:"data"`
	}
	if err = d.client.Get(c, redDotURL, "", params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), redDotURL+"?"+params.Encode())
		return
	}
	log.Warn("reddot response mid(%d) url(%s) res(%t)", mid, redDotURL+"?"+params.Encode(), res.Data.RedDot)
	ok = res.Data.RedDot
	return
}

// Prompting get up face&name has updated
func (d *Dao) Prompting(c context.Context, mid int64) (prompt *accgrpc.PromptingReply, err error) {
	arg := &accgrpc.MidReq{Mid: mid}
	if prompt, err = d.accGRPC.Prompting(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
	}
	return
}

func (d *Dao) CharacterUsageStatus(c context.Context, mid, build int64, platform, mobiApp, device, buvid string) (*garbgrpc.UsageReply, error) {
	req := &garbgrpc.UsageReq{
		Mid:      mid,
		Buvid:    buvid,
		Build:    build,
		Platform: platform,
		MobiApp:  mobiApp,
		Device:   device,
	}
	reply, err := d.grabGRPC.CharacterUsageStatusV2(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) RealnameTeenAgeCheck(c context.Context, mid int64, ip string) (*memberAPI.RealnameTeenAgeCheckReply, error) {
	req := &memberAPI.MidReq{Mid: mid, RealIP: ip}
	rly, err := d.memberRPC.RealnameTeenAgeCheck(c, req)
	if err != nil {
		log.Error("Fail to request memberAPI.RealnameTeenAgeCheck, req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly, nil
}

func statisticsKey(mid int64, date time.Time) string {
	return fmt.Sprintf("%d-%s", mid, date.Format("20060102"))
}

func (d *Dao) ExportStatistics(ctx context.Context, mid int64, date time.Time, sel string) (*shuangqing.ShuangQing, error) {
	staticKv, ok := d.shuangQingStaticKvs[sel]
	if !ok {
		return nil, errors.Errorf("get d.shuangQingStaticKvs error mid=%d, sel=%s", mid, sel)
	}
	req := staticKv.NewGetReq([]byte(statisticsKey(mid, date)))
	reply, err := staticKv.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	out := &shuangqing.ShuangQing{}
	if err := out.Unmarshal(reply.Columns[0].Value); err != nil {
		return nil, errors.WithStack(err)
	}
	return out, nil
}

func (d *Dao) UserDetail(ctx context.Context, in *passportuser.UserDetailReq) (*passportuser.UserDetailReply, error) {
	reply, err := d.passportUser.UserDetail(ctx, in)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply, nil
}

func (d *Dao) NFTBatchInfo(ctx context.Context, in *memberAPI.NFTBatchInfoReq) (*memberAPI.NFTBatchInfoReply, error) {
	reply, err := d.memberRPC.NFTBatchInfo(ctx, in)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply, nil
}

func (d *Dao) UserActiveLocation(ctx context.Context, in *passportuser.MidReq) (*passportuser.UserActiveLocationReply, error) {
	reply, err := d.passportUser.UserActiveLocation(ctx, in)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply, nil
}

func (d *Dao) GetUserExtraValueSingleKey(ctx context.Context, in *memberAPI.UserExtraValueSingleKeyReq) (*memberAPI.UserExtraValueSingleKeyReply, error) {
	reply, err := d.memberRPC.GetUserExtraValueSingleKey(ctx, in)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply, nil
}

func (d *Dao) GetUserExtraBasedOnKeys(ctx context.Context, in *memberAPI.GetUserExtraBasedOnKeyReq) (*memberAPI.UserExtraValues, error) {
	reply, err := d.memberRPC.GetUserExtraBasedOnKeys(ctx, in)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply, nil
}
