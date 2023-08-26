package lottery

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"go-common/library/log"

	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
	"go-gateway/app/web-svr/activity/interface/model/pay"
	"go-gateway/app/web-svr/activity/interface/rewards"
	suitmdl "go-main/app/account/usersuit/service/api"

	couponapi "git.bilibili.co/bapis/bapis-go/account/service/coupon"
	ogvapi "git.bilibili.co/bapis/bapis-go/cheese/service/coupon/v2"
	vipresourceapi "git.bilibili.co/bapis/bapis-go/vip/resource/service"
	"github.com/pkg/errors"
)

const (
	_ = iota
	// 实物奖品
	giftEntity
	// 大会员奖品
	giftVip
	// 头像挂件
	giftGrant
	// 优惠券
	giftCoupon
	// 硬币
	giftCoin
	// 大会员体验券
	giftVipCoupon
	// 其他类型
	giftOther
	// ogv券
	giftOGV
	// 会员购
	giftVipBuy
	// 现金券
	giftMoney
	// 奖励平台
	giftAward
	remarkCoin = "抽奖获得硬币"
	// OgvBizType ogv类型
	OgvBizType = 1
)

const (
	// lotteryActivity ...
	lotteryActivity = "lottery_activity"
	retry           = 3
	pageLimit       = 50
)

// GiftInterface interface
type GiftInterface interface {
	// Check 发放校验
	Check(memberInfo *l.MemberInfo) error
	Init() error
	ResGift() *l.Gift
	Send(c context.Context, s *Service, sid, senderID int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo) (err error)
}

// getGiftByType
func getGiftByType(c context.Context, gift *l.Gift) (GiftInterface, error) {
	switch gift.Type {
	case giftEntity:
		return &GiftMaterialObject{Gift: gift}, nil
	case giftVip:
		return &GiftVip{Gift: gift}, nil
	case giftGrant:
		return &GiftGrant{Gift: gift}, nil
	case giftCoupon:
		return &GiftCoupon{Gift: gift}, nil
	case giftCoin:
		return &GiftCoin{Gift: gift}, nil
	case giftVipCoupon:
		return &GiftVipCoupon{Gift: gift}, nil
	case giftOther:
		return &GiftOther{Gift: gift}, nil
	case giftOGV:
		return &GiftOGV{Gift: gift}, nil
	case giftVipBuy:
		return &GiftVipBuy{Gift: gift}, nil
	case giftMoney:
		return &GiftMoney{Gift: gift}, nil
	case giftAward:
		return &GiftAward{Gift: gift}, nil
	}
	return nil, ecode.ActivityLotteryGiftFoundNoType
}

// GiftMaterialObject 实物奖品
type GiftMaterialObject struct {
	*l.Gift
}

// Init ...
func (g *GiftMaterialObject) Init() error {
	return nil
}

// ResGift gift
func (g *GiftMaterialObject) ResGift() *l.Gift {
	return g.Gift
}

// Send ...
func (g *GiftMaterialObject) Send(c context.Context, s *Service, sid, senderID int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo) (err error) {
	if _, err = s.lottery.InsertLotteryWin(c, sid, g.Gift.ID, member.Mid, member.IP); err != nil {
		log.Errorc(c, "DoLottery s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", sid, g.Gift.ID, member.Mid, err)
		return
	}
	s.sendSysMsg(c, member.Mid, lottery, info, g.ResGift(), senderID, "")
	return nil
}

// Check gift
func (g *GiftMaterialObject) Check(memberInfo *l.MemberInfo) error {
	return memberInfo.IsValidIP()
}

// GiftVip 大会员
type GiftVip struct {
	*l.Gift
	Params *VipParams
}

// VipParams ...
type VipParams struct {
	Token  string `json:"token"`
	AppKey string `json:"app_key"`
}

// Init ...
func (g *GiftVip) Init() (err error) {
	if g.Gift.Source == "" && g.Gift.Params == "" {
		return ecode.ActivityLotteryGiftParamsError
	}
	params := &VipParams{}
	if g.Gift.Source != "" {
		params.AppKey = ""
		params.Token = g.Gift.Source
	}
	if g.Gift.Params != "" {
		err = json.Unmarshal([]byte(g.Gift.Params), &params)
		if err != nil {
			err = errors.Wrapf(err, "GiftGrant json error %s", g.Gift.Params)
			return err
		}
	}

	if params.Token == "" {
		return ecode.ActivityLotteryGiftParamsError
	}
	g.Params = params
	return nil
}

// Send ...
func (g *GiftVip) Send(c context.Context, s *Service, sid, senderID int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo) (err error) {
	if _, err = s.lottery.InsertLotteryWin(c, sid, g.Gift.ID, member.Mid, member.IP); err != nil {
		log.Errorc(c, "DoLottery s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", sid, g.Gift.ID, member.Mid, err)
		return
	}
	if err = s.experienceVip(c, member.Mid, g.Params.Token, l.Remark, g.Params.AppKey); err != nil {
		log.Errorc(c, "sendLotteryAward s.dao.MemberVip mid:%d token(%s) appKey(%s) error(%v)", member.Mid, g.Params.Token, g.Params.AppKey, err)
		return
	}
	if err = s.sendSysMsg(c, member.Mid, lottery, info, g.ResGift(), senderID, ""); err != nil {
		log.Errorc(c, "s.sendSysMsg mid(%d)", member.Mid)
	}
	return nil
}

// experienceVip 体验券
func (s *Service) experienceVip(c context.Context, mid int64, token string, remark string, appKey string) error {
	timestamp := time.Now().Unix()
	midStr := strconv.FormatInt(mid, 10)
	_, err := s.vipResourceClient.ResourceUse(c, &vipresourceapi.ResourceUseReq{
		Mid:        mid,
		BatchToken: token,
		OrderNo:    strconv.FormatInt(time.Now().UnixNano(), 10) + midStr + strconv.FormatInt(rand.Int63n(1000), 10),
		Remark:     remark,
		Appkey:     appKey,
		Ts:         timestamp,
	})
	if err != nil {
		log.Errorc(c, "s.vipResourceClient.ResourceUse(%d) error(%v)", mid, err)
		return err
	}
	return nil
}

// Check gift
func (g *GiftVip) Check(memberInfo *l.MemberInfo) error {
	return nil
}

// ResGift gift
func (g *GiftVip) ResGift() *l.Gift {
	return g.Gift
}

// GiftGrant 头像挂件
type GiftGrant struct {
	*l.Gift
	Params *GrantParams
}

// Send ...
func (g *GiftGrant) Send(c context.Context, s *Service, sid, senderID int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo) (err error) {
	if _, err = s.lottery.InsertLotteryWin(c, sid, g.Gift.ID, member.Mid, member.IP); err != nil {
		log.Errorc(c, "DoLottery s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", sid, g.Gift.ID, member.Mid, err)
		return
	}
	uids := []int64{member.Mid}
	if _, err = s.suitClient.GrantByMids(c, &suitmdl.GrantByMidsReq{Mids: uids, Pid: g.Params.Pid, Expire: g.Params.Expire}); err != nil {
		log.Errorc(c, "s.suitClient.GrantByMids(%d,%d,%d) error(%v)", member.Mid, g.Params.Pid, g.Params.Expire, err)
		return err
	}
	if err = s.sendSysMsg(c, member.Mid, lottery, info, g.ResGift(), senderID, ""); err != nil {
		log.Errorc(c, "s.sendSysMsg mid(%d)", member.Mid)
	}
	return nil
}

// GrantParams 头像资源
type GrantParams struct {
	Pid    int64 `json:"pid"`
	Expire int64 `json:"expire"`
}

// Init ...
func (g *GiftGrant) Init() error {
	params := &GrantParams{}
	if g.Gift.Source == "" && g.Gift.Params == "" {
		return ecode.ActivityLotteryGiftParamsError
	}
	if g.Gift.Source != "" {
		err := json.Unmarshal([]byte(g.Gift.Source), &params)
		if err != nil {
			err = errors.Wrapf(err, "GiftGrant json error %s", g.Gift.Source)
		}
		g.Params = params
	}
	if g.Gift.Params != "" {
		err := json.Unmarshal([]byte(g.Gift.Params), &params)
		if err != nil {
			err = errors.Wrapf(err, "GiftGrant json error %s", g.Gift.Params)
			return err
		}
		g.Params = params
	}
	if params.Pid == 0 || params.Expire == 0 {
		return ecode.ActivityLotteryGiftParamsError
	}

	return nil
}

// ResGift gift
func (g *GiftGrant) ResGift() *l.Gift {
	return g.Gift
}

// Check gift
func (g *GiftGrant) Check(memberInfo *l.MemberInfo) error {
	return nil
}

// GiftCoupon 优惠券
type GiftCoupon struct {
	*l.Gift
}

// Init gift
func (g *GiftCoupon) Init() error {
	return nil
}

// Check gift
func (g *GiftCoupon) Check(memberInfo *l.MemberInfo) error {
	return nil
}

// ResGift gift
func (g *GiftCoupon) ResGift() *l.Gift {
	return g.Gift
}

// Send ...
func (g *GiftCoupon) Send(c context.Context, s *Service, sid, senderID int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo) (err error) {
	var cdKey string
	if _, err = s.lottery.UpdateLotteryWin(c, sid, member.Mid, g.Gift.ID, member.IP); err != nil {
		log.Errorc(c, "sendLotteryAward s.dao.UpdateLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", sid, g.Gift.ID, member.Mid, err)
		return
	}
	if cdKey, err = s.lottery.RawLotteryWinOne(c, sid, member.Mid, g.Gift.ID); err != nil {
		log.Errorc(c, "sendLotteryAward s.dao.RawLotteryWinOne id(%d) mid(%d) gift_id(%d) error(%v)", sid, member.Mid, g.Gift.ID, err)
		return
	}
	s.sendSysMsg(c, member.Mid, lottery, info, g.ResGift(), senderID, "兑换码为："+cdKey)
	return
}

// GiftCoin 硬币
type GiftCoin struct {
	*l.Gift
	Params *CoinParams
}

// CoinParams ...
type CoinParams struct {
	Coin float64
}

// Init ...
func (g *GiftCoin) Init() error {
	params := &CoinParams{}
	if g.Gift.Source == "" && g.Gift.Params == "" {
		return ecode.ActivityLotteryGiftParamsError
	}
	if g.Gift.Source != "" {
		params.Coin, _ = strconv.ParseFloat(g.Gift.Source, 64)
	}
	if g.Gift.Params != "" {
		err := json.Unmarshal([]byte(g.Gift.Params), &params)
		if err != nil {
			err = errors.Wrapf(err, "GiftGrant json error %s", g.Gift.Params)
			return err
		}
		g.Params = params
	}
	if params.Coin == 0 {
		return ecode.ActivityLotteryGiftParamsError
	}

	return nil
}

// Send ...
func (g *GiftCoin) Send(c context.Context, s *Service, sid, senderID int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo) (err error) {
	var giftID int64
	if giftID, err = s.lottery.InsertLotteryWin(c, sid, g.Gift.ID, member.Mid, member.IP); err != nil {
		log.Errorc(c, "sendLotteryAward s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", sid, g.Gift.ID, member.Mid, err)
		return
	}
	orderID := s.md5(fmt.Sprintf("%d_%d_%d_%d", member.Mid, sid, g.Gift.ID, giftID))
	if _, err = s.coinClient.ModifyCoins(c, &coinmdl.ModifyCoinsReq{Mid: member.Mid, Count: g.Params.Coin, Reason: remarkCoin, IP: member.IP, UniqueID: orderID, Caller: lotteryActivity}); err != nil {
		log.Errorc(c, "sendLotteryAward need check coin.ModifyCoin mid:%d count:%v error(%v)", member.Mid, g.Params.Coin, err)
		return
	}
	if err = s.sendSysMsg(c, member.Mid, lottery, info, g.ResGift(), senderID, ""); err != nil {
		log.Errorc(c, "s.sendSysMsg mid(%d)", member.Mid)
	}
	return nil
}

// ResGift gift
func (g *GiftCoin) ResGift() *l.Gift {
	return g.Gift
}

// Check gift
func (g *GiftCoin) Check(memberInfo *l.MemberInfo) error {
	return nil
}

// GiftVipCoupon 大会员抵用券
type GiftVipCoupon struct {
	*l.Gift
	Params *VipCouponParams
}

// ResGift gift
func (g *GiftVipCoupon) ResGift() *l.Gift {
	return g.Gift
}

// Send ...
func (g *GiftVipCoupon) Send(c context.Context, s *Service, sid, senderID int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo) (err error) {
	if _, err = s.lottery.InsertLotteryWin(c, sid, g.Gift.ID, member.Mid, member.IP); err != nil {
		log.Errorc(c, "sendLotteryAward s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", sid, g.Gift.ID, member.Mid, err)
		return
	}
	midStr := strconv.FormatInt(member.Mid, 10)
	orderNo := strconv.FormatInt(time.Now().UnixNano(), 10) + midStr + strconv.FormatInt(rand.Int63n(1000), 10)
	_, err = s.couponClient.AllowanceReceive(c, &couponapi.AllowanceReceiveReq{
		Mid:        member.Mid,
		BatchToken: g.Params.Token,
		OrderNo:    orderNo,
		Appkey:     g.Params.AppKey,
	})
	if err != nil {
		log.Errorc(c, "s.couponClient.AllowanceReceive(%d)  error(%v)", member.Mid, err)
		return
	}
	if err = s.sendSysMsg(c, member.Mid, lottery, info, g.ResGift(), senderID, ""); err != nil {
		log.Errorc(c, "s.sendSysMsg mid(%d)", member.Mid)
	}
	return nil
}

// VipCouponParams ...
type VipCouponParams struct {
	AppKey string `json:"app_key"`
	Token  string `json:"token"`
}

// Init ...
func (g *GiftVipCoupon) Init() error {
	var err error
	if g.Gift.Source == "" && g.Gift.Params == "" {
		return ecode.ActivityLotteryGiftParamsError
	}
	params := &VipCouponParams{}
	if g.Gift.Source != "" {
		params.AppKey = ""
		params.Token = g.Gift.Source
	}
	err = json.Unmarshal([]byte(g.Gift.Params), &params)
	if err != nil {
		err = errors.Wrapf(err, "GiftGrant json error %s", g.Gift.Params)
		return err
	}
	g.Params = params
	if params.Token == "" {
		return ecode.ActivityLotteryGiftParamsError
	}
	return nil
}

// Check gift
func (g *GiftVipCoupon) Check(memberInfo *l.MemberInfo) error {
	return nil
}

// GiftOther 其他奖品
type GiftOther struct {
	*l.Gift
}

// Init gift
func (g *GiftOther) Init() error {
	return nil
}

// ResGift gift
func (g *GiftOther) ResGift() *l.Gift {
	return g.Gift
}

// Check gift
func (g *GiftOther) Check(memberInfo *l.MemberInfo) error {
	return nil
}

// Send ...
func (g *GiftOther) Send(c context.Context, s *Service, sid, senderID int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo) (err error) {
	if _, err = s.lottery.InsertLotteryWin(c, sid, g.Gift.ID, member.Mid, member.IP); err != nil {
		log.Errorc(c, "DoLottery s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", sid, g.Gift.ID, member.Mid, err)
		return
	}
	if err = s.sendSysMsg(c, member.Mid, lottery, info, g.ResGift(), senderID, ""); err != nil {
		log.Errorc(c, "s.sendSysMsg mid(%d)", member.Mid)
	}
	return nil
}

// GiftOGV OGV奖品
type GiftOGV struct {
	*l.Gift
	Params *OGVParams
}

// OGVParams ...
type OGVParams struct {
	Token string `json:"token"`
}

// ResGift gift
func (g *GiftOGV) ResGift() *l.Gift {
	return g.Gift
}

// Init gift
func (g *GiftOGV) Init() error {
	var err error
	if g.Gift.Params == "" {
		return ecode.ActivityLotteryGiftParamsError
	}
	params := &OGVParams{}
	err = json.Unmarshal([]byte(g.Gift.Params), &params)
	if err != nil {
		err = errors.Wrapf(err, "GiftGrant json error %s", g.Gift.Params)
		return err
	}
	if params.Token == "" {
		return ecode.ActivityLotteryGiftParamsError
	}
	g.Params = params
	return nil
}

// Check gift
func (g *GiftOGV) Check(memberInfo *l.MemberInfo) error {
	return nil
}

// Send ...
func (g *GiftOGV) Send(c context.Context, s *Service, sid, senderID int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo) (err error) {
	if _, err = s.lottery.InsertLotteryWin(c, sid, g.Gift.ID, member.Mid, member.IP); err != nil {
		log.Errorc(c, "DoLottery s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", sid, g.Gift.ID, member.Mid, err)
		return
	}
	var crowType ogvapi.CrowType
	if member.IsAnnualVip() == nil {
		crowType = ogvapi.CrowType_YEAR_VIP
	} else if member.IsMonthVip() == nil {
		crowType = ogvapi.CrowType_MONTH_VIP
	} else {
		crowType = ogvapi.CrowType_NORMAL
	}
	_, err = s.ogvClient.ReceiveCoupon(c, &ogvapi.ReceiveCouponReq{
		Mid:        member.Mid,
		BatchToken: g.Params.Token,
		BizType:    OgvBizType,
		CrowType:   crowType,
	})
	if err != nil {
		log.Errorc(c, "s.ogvClient.ReceiveCoupon(%d)  error(%v)", member.Mid, err)
		return
	}
	if err = s.sendSysMsg(c, member.Mid, lottery, info, g.ResGift(), senderID, ""); err != nil {
		log.Errorc(c, "s.sendSysMsg mid(%d)", member.Mid)
	}
	return nil
}

// GiftVipBuy 会员购奖品
type GiftVipBuy struct {
	*l.Gift
	Params *VipBuyParams
}

// VipBuyParams ...
type VipBuyParams struct {
	Token string `json:"token"`
}

// ResGift gift
func (g *GiftVipBuy) ResGift() *l.Gift {
	return g.Gift
}

// Init gift
func (g *GiftVipBuy) Init() error {
	var err error
	if g.Gift.Params == "" {
		return ecode.ActivityLotteryGiftParamsError
	}
	params := &VipBuyParams{}
	err = json.Unmarshal([]byte(g.Gift.Params), &params)
	if err != nil {
		err = errors.Wrapf(err, "GiftVipBuy json error %s", g.Gift.Params)
		return err
	}
	if params.Token == "" {
		return ecode.ActivityLotteryGiftParamsError
	}
	g.Params = params
	return nil
}

// Check gift
func (g *GiftVipBuy) Check(memberInfo *l.MemberInfo) error {
	return nil
}

// Send ...
func (g *GiftVipBuy) Send(c context.Context, s *Service, sid, senderID int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo) (err error) {
	var giftID int64
	if giftID, err = s.lottery.InsertLotteryWin(c, sid, g.Gift.ID, member.Mid, member.IP); err != nil {
		log.Errorc(c, "DoLottery s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", sid, g.Gift.ID, member.Mid, err)
		return
	}
	orderID := fmt.Sprintf("%d_%d_%d", member.Mid, sid, giftID)
	s.retry(c, func() error {
		err = s.lottery.SendVipBuyCoupon(c, member.IP, g.Params.Token, s.c.Lottery.VipBuy.SourceActivityID, orderID, member.Name, s.c.Lottery.VipBuy.SourceID, member.Mid)
		if err != nil {
			log.Errorc(c, "s.SendVipBuyCoupon(%d,%s,%s,%s,%s) error(%v)", member.Mid, member.IP, g.Params.Token, orderID, member.Name, err)
			err = fmt.Errorf("s.SendVipBuyCoupon(%d,%s,%s,%s,%s) error(%v)", member.Mid, member.IP, g.Params.Token, orderID, member.Name, err)
			return err
		}
		return err
	})
	if err = s.sendSysMsg(c, member.Mid, lottery, info, g.ResGift(), senderID, ""); err != nil {
		log.Errorc(c, "s.sendSysMsg mid(%d)", member.Mid)
	}
	return nil
}

// GiftMoney 现金奖品
type GiftMoney struct {
	*l.Gift
	Params *MoneyParams
}

// MoneyParams ...
type MoneyParams struct {
	ActivityID   string `json:"activity_id"`
	TransBalance int64  `json:"money"`
	StartTme     int64  `json:"start_time"`
	TransDesc    string `json:"trans_desc"`
}

// Init gift
func (g *GiftMoney) Init() error {
	var err error
	if g.Gift.Params == "" {
		return ecode.ActivityLotteryGiftParamsError
	}
	params := &MoneyParams{}
	err = json.Unmarshal([]byte(g.Gift.Params), &params)
	if err != nil {
		err = errors.Wrapf(err, "GiftGrant json error %s", g.Gift.Params)
		return err
	}
	if params.ActivityID == "" || params.TransBalance == 0 || params.StartTme == 0 {
		return ecode.ActivityLotteryGiftParamsError
	}
	g.Params = params
	return nil
}

// ResGift gift
func (g *GiftMoney) ResGift() *l.Gift {
	return g.Gift
}

// GiftAward 奖励平台
type GiftAward struct {
	*l.Gift
	Params *AwardParams
}

// AwardParams ...
type AwardParams struct {
	AwardID int64 `json:"award_id"`
}

// Check gift
func (g *GiftAward) Check(memberInfo *l.MemberInfo) error {
	return nil
}

// ResGift ...
func (g *GiftAward) ResGift() *l.Gift {
	return g.Gift
}

// Send ...
func (g *GiftAward) Send(c context.Context, s *Service, sid, senderID int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo) (err error) {
	var giftID int64
	if giftID, err = s.lottery.InsertLotteryWin(c, sid, g.Gift.ID, member.Mid, member.IP); err != nil {
		log.Errorc(c, "DoLottery s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", sid, g.Gift.ID, member.Mid, err)
		return
	}
	orderID := s.md5(fmt.Sprintf("%d_%s_%d_%d_%d", member.Mid, lotteryActivity, sid, g.Gift.ID, giftID))
	log.Infoc(c, "gift send award gift mid (%d) orderID(%v)", member.Mid, orderID)
	_, err = rewards.Client.SendAwardByIdAsync(c, member.Mid, orderID, "lottery", g.Params.AwardID, true, true)
	if err != nil {
		log.Errorc(c, "rewards.Client.SendAwardByIdAsync error: mid: %v, uniqueId: %v, err: %v", member.Mid, orderID, err)
		return err
	}
	return nil
}

// Init ...
func (g *GiftAward) Init() error {
	var err error
	if g.Gift.Params == "" {
		return ecode.ActivityLotteryGiftParamsError
	}
	params := &AwardParams{}
	err = json.Unmarshal([]byte(g.Gift.Params), &params)
	if err != nil {
		err = errors.Wrapf(err, "GiftGrant json error %s", g.Gift.Params)
		return err
	}
	if params.AwardID == 0 {
		return ecode.ActivityLotteryGiftParamsError
	}
	g.Params = params

	return nil
}

// Send ...
func (g *GiftMoney) Send(c context.Context, s *Service, sid, senderID int64, lottery *l.Lottery, info *l.Info, member *l.MemberInfo) (err error) {
	var giftID int64
	if giftID, err = s.lottery.InsertLotteryWin(c, sid, g.Gift.ID, member.Mid, member.IP); err != nil {
		log.Errorc(c, "DoLottery s.dao.InsertLotteryWin id(%d) giftID(%d) mid(%d) error(%v)", sid, g.Gift.ID, member.Mid, err)
		return
	}
	orderID := s.md5(fmt.Sprintf("%d_%s_%d_%d_%d", member.Mid, lotteryActivity, sid, g.Gift.ID, giftID))
	log.Infoc(c, "gift send money mid (%d) orderID(%v)", member.Mid, orderID)
	now := time.Now()
	s.retry(c, func() error {
		reply, err := s.payTransferInner(c, member.Mid, g.Params.TransBalance, orderID, g.Params.TransDesc, now, g.Params.StartTme, g.Params.ActivityID)
		if pay.OrderStatusFail == reply.OrderStatus() {
			log.Errorc(c, "s.payTransferInner(%d,%d,%s) error(%v)", member.Mid, g.Params.TransBalance, orderID, err)
			err = fmt.Errorf("s.payTransferInner(%d,%d,%s) error(%v)", member.Mid, g.Params.TransBalance, orderID, err)
		}
		return err
	})

	if err = s.sendSysMsg(c, member.Mid, lottery, info, g.ResGift(), senderID, ""); err != nil {
		log.Errorc(c, "s.sendSysMsg mid(%d)", member.Mid)
	}
	return nil
}

func (s *Service) retry(c context.Context, f func() error) {
	for i := 0; i < retry; i++ {
		err := f()
		if err == nil {
			return
		}
		log.Errorc(c, "retry info:%v", err)
		time.Sleep(timeSleep)
	}
}

// Check gift
func (g *GiftMoney) Check(memberInfo *l.MemberInfo) error {
	return nil
}

func (s *Service) md5(source string) string {
	md5Str := md5.New()
	md5Str.Write([]byte(source))
	return hex.EncodeToString(md5Str.Sum(nil))
}

// getLatestGiftStore 获得最新库存信息
func (s *Service) getLatestGiftStore(c context.Context, sid, mid int64, giftList []*l.Gift, day string) (newGiftList []*l.Gift, err error) {
	giftIds := make([]int64, 0)
	newGiftList = make([]*l.Gift, 0)
	for _, v := range giftList {
		giftIds = append(giftIds, v.ID)
	}
	var (
		giftNum    map[int64]int64
		giftDayNum map[int64]l.DayNum
		giftMidNum map[int64]l.DayNum
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if giftNum, err = s.getStore(c, sid, giftIds); err != nil {
			err = errors.Wrapf(err, "s.getStore(%d,%v)", sid, giftIds)
			return
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if giftDayNum, giftMidNum, err = s.getDayStore(c, sid, mid, day, giftList); err != nil {
			err = errors.Wrapf(err, "s.getDayStore(%d,%s,%v)", sid, day, giftList)
			return
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return
	}
	for _, v := range giftList {
		gift := v
		if giftNum != nil {
			if num, ok := giftNum[v.ID]; ok {
				gift.SendNum = num
			}
		}
		if giftDayNum != nil {
			if dayNum, ok := giftDayNum[v.ID]; ok {
				gift.DaySendNum = dayNum
			}
		}
		if giftMidNum != nil {
			if dayNum, ok := giftMidNum[v.ID]; ok {
				gift.OtherSendNum = dayNum
			}
		}
		newGiftList = append(newGiftList, gift)
	}
	return
}

// getStore 获得总库存
func (s *Service) getStore(c context.Context, sid int64, giftIds []int64) (resGiftNum map[int64]int64, err error) {
	var (
		giftNum map[int64]int64
	)
	resGiftNum = make(map[int64]int64)
	giftNum, err = s.lottery.CacheSendGiftNum(c, sid, giftIds)
	if err != nil {
		log.Errorc(c, "s.lottery.CacheSendGiftNum(%d, %v) error(%v)", sid, giftIds, err)
		return
	}
	for _, v := range giftIds {
		if giftNum != nil {
			if num, ok := giftNum[v]; ok {
				resGiftNum[v] = num
				continue
			}
		}
		resGiftNum[v] = 0
	}
	return
}

// getGiftDayKey 获取gift 每日上限的key
func (s *Service) getGiftDayKey(giftID int64, giftDayNumKey string) string {
	return fmt.Sprintf("%d_%s", giftID, giftDayNumKey)
}

// getDayStore 获得日库存
func (s *Service) getDayStore(c context.Context, sid, mid int64, day string, giftList []*l.Gift) (giftDayNum map[int64]l.DayNum, giftMidNum map[int64]l.DayNum, err error) {
	giftKey := make([]string, 0)
	giftMidKey := make([]string, 0)
	giftDayNum = make(map[int64]l.DayNum)
	giftMidNum = make(map[int64]l.DayNum)
	for _, v := range giftList {
		if v.DayNum != nil {
			for key, num := range v.DayNum {
				if num > 0 {
					if key == l.GiftMidNumKey {
						giftMidKey = append(giftMidKey, s.getGiftDayKey(v.ID, key))
						continue
					}
					giftKey = append(giftKey, s.getGiftDayKey(v.ID, key))
				}

			}
		}
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if giftDayNum, err = s.getRedisDayOtherNum(c, sid, day, giftKey, giftList); err != nil {
			err = errors.Wrapf(err, "s.getRedisDayOtherNum(%d,%v)", sid, day)
			return
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		midKey := fmt.Sprintf("%d", mid)
		if giftMidNum, err = s.getRedisDayOtherNum(c, sid, midKey, giftMidKey, giftList); err != nil {
			err = errors.Wrapf(err, "s.getRedisDayOtherNum(%d,%v)", sid, mid)
			return
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return
	}

	return giftDayNum, giftMidNum, nil
}

func (s *Service) getRedisDayOtherNum(c context.Context, sid int64, redisKey string, giftKey []string, giftList []*l.Gift) (giftNum map[int64]l.DayNum, err error) {
	giftNum = make(map[int64]l.DayNum)
	giftKeyNum := make(map[string]int64)
	if len(giftKey) > 0 {
		giftKeyNum, err = s.lottery.CacheSendDayGiftNum(c, sid, redisKey, giftKey)
		if err != nil {
			log.Errorc(c, "s.lottery.CacheSendDayGiftNum(%d,%s, %v) error(%v)", sid, redisKey, giftKey, err)
			return
		}
		for _, v := range giftList {
			if v.DayNum != nil {
				var dayNum = make(l.DayNum)
				for key := range v.DayNum {
					if num, ok := giftKeyNum[s.getGiftDayKey(v.ID, key)]; ok {
						dayNum[key] = num
					}
				}
				giftNum[v.ID] = dayNum
			}
		}
	}
	return giftNum, nil

}

// sendStore
func (s *Service) sendStore(c context.Context, sid int64, giftNum map[int64]int) (resStore map[int64]int64, err error) {
	resStore, err = s.lottery.IncrGiftSendNum(c, sid, giftNum)
	if err != nil {
		log.Errorc(c, "s.lottery.IncrGiftSendNum(%d,%v) error(%v)", sid, giftNum, err)
		return
	}
	return
}

// sendDayStore 增加日库存
func (s *Service) sendDayStore(c context.Context, lottery *l.Lottery, mid int64, day string, giftList []*l.Gift, num int) (giftDayNum map[int64]l.DayNum, giftMidNum map[int64]l.DayNum, err error) {
	giftKeyNum := make(map[string]int)
	giftMidKey := make(map[string]int)
	giftDayNum = make(map[int64]l.DayNum)
	giftMidNum = make(map[int64]l.DayNum)
	for _, v := range giftList {
		if v.DayNum != nil {
			for key, n := range v.DayNum {
				if n > 0 {
					if key == l.GiftMidNumKey {
						_, ok := giftMidKey[s.getGiftDayKey(v.ID, key)]
						if ok {
							giftMidKey[s.getGiftDayKey(v.ID, key)] += num
						} else {
							giftMidKey[s.getGiftDayKey(v.ID, key)] = num
						}
						continue
					}
					_, ok := giftKeyNum[s.getGiftDayKey(v.ID, key)]
					if ok {
						giftKeyNum[s.getGiftDayKey(v.ID, key)] += num
					} else {
						giftKeyNum[s.getGiftDayKey(v.ID, key)] = num
					}
				}
			}

		}
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		var expireTime int64
		expireTime = 24 * 30 * 3600
		if giftDayNum, err = s.incrGiftSendDayOrMidNum(c, lottery.ID, day, expireTime, giftList, giftKeyNum); err != nil {
			err = errors.Wrapf(err, "s.getRedisDayOtherNum(%d,%v)", lottery.ID, day)
			return
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		var expireTime int64
		expireTime = s.getLotteryExpireTime(lottery)
		midKey := fmt.Sprintf("%d", mid)
		if giftMidNum, err = s.incrGiftSendDayOrMidNum(c, lottery.ID, midKey, expireTime, giftList, giftMidKey); err != nil {
			err = errors.Wrapf(err, "s.getRedisDayOtherNum(%d,%v)", lottery.ID, mid)
			return
		}
		return
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return
	}
	return giftDayNum, giftMidNum, nil
}

func (s *Service) getLotteryExpireTime(lottery *l.Lottery) int64 {
	return (int64(lottery.Etime) - int64(lottery.Stime)) * 2
}

func (s *Service) incrGiftSendDayOrMidNum(c context.Context, sid int64, redisKey string, expireTime int64, giftList []*l.Gift, giftKeyNum map[string]int) (giftDayNum map[int64]l.DayNum, err error) {
	giftDayNum = make(map[int64]l.DayNum)
	if len(giftKeyNum) > 0 {
		resDayStore, err := s.lottery.IncrGiftSendDayNum(c, sid, redisKey, giftKeyNum, expireTime)
		if err != nil {
			log.Errorc(c, "s.lottery.IncrGiftSendDayNum(%d,%s) error(%v)", sid, redisKey, err)
			return nil, err
		}
		for _, v := range giftList {
			if v.DayNum != nil {
				var dayNum = make(l.DayNum)
				for key := range v.DayNum {
					if num, ok := resDayStore[s.getGiftDayKey(v.ID, key)]; ok {
						dayNum[key] = num
					}
				}
				giftDayNum[v.ID] = dayNum
			}
		}
		return giftDayNum, nil
	}
	return nil, err
}

// checkSendStore 验证本次商品是否可发放
func (s *Service) checkSendStore(c context.Context, lottery *l.Lottery, mid int64, giftList []*l.Gift, day string, num int) (newGiftList []*l.Gift, err error) {
	newGiftList = make([]*l.Gift, 0)
	var (
		giftNum    map[int64]int64
		giftDayNum map[int64]l.DayNum
		giftMidNum map[int64]l.DayNum
	)
	giftMap := make(map[int64]*l.Gift)

	// 首先判断日库存
	giftDayNum, giftMidNum, err = s.sendDayStore(c, lottery, mid, day, giftList, num)
	if err != nil {
		err = errors.Wrapf(err, "s.sendDayStore(%d,%s,%v,%d)", lottery.ID, mid, day, giftList, num)
		return
	}
	minusGiftSendDayNum := make([]*l.Gift, 0)
	giftSendNum := make(map[int64]int)
	defer func() {
		if len(minusGiftSendDayNum) > 0 {
			s.cache.SyncDo(c, func(c context.Context) {
				giftDayNum, giftMidNum, err = s.sendDayStore(c, lottery, mid, day, minusGiftSendDayNum, -num)
				if err != nil {
					log.Errorc(c, "s.sendDayStore lottery(%v)mid(%d) day(%s) minusGiftSendDayNum(%v) -num(%d)", lottery, mid, day, minusGiftSendDayNum, -num)
				}

			})

		}
	}()
	for _, v := range giftList {
		gift := v
		giftMap[v.ID] = v
		if giftDayNum != nil {
			if dayNum, ok := giftDayNum[v.ID]; ok {
				gift.DaySendNum = dayNum
			}
		}
		if giftMidNum != nil {
			if dayNum, ok := giftMidNum[v.ID]; ok {
				gift.OtherSendNum = dayNum
			}
		}
		if gift.CheckSendStore(c) == nil {
			_, ok := giftSendNum[v.ID]
			if ok {
				giftSendNum[v.ID] += num
			} else {
				giftSendNum[v.ID] = num
			}
		} else {
			minusGiftSendDayNum = append(minusGiftSendDayNum, v)
		}

	}
	if len(giftSendNum) > 0 {
		if giftNum, err = s.sendStore(c, lottery.ID, giftSendNum); err != nil {
			err = errors.Wrapf(err, "s.sendStore(%d,%v)", lottery.ID, giftSendNum)
			return
		}
		minusGiftSendNum := make(map[int64]int)
		defer func() {
			if len(minusGiftSendNum) > 0 {
				s.cache.SyncDo(c, func(c context.Context) {
					_, err = s.sendStore(c, lottery.ID, minusGiftSendNum)
					if err != nil {
						log.Errorc(c, "s.sendStore(%d,%v)", lottery.ID, minusGiftSendNum)
					}

				})
			}
		}()
		resultSend := make(map[int64]int)
		for _, v := range giftList {
			gift := v
			if giftNum != nil {
				if num, ok := giftNum[v.ID]; ok {
					gift.SendNum = num
				}
			}
			if gift.CheckSendStore(c) == nil {
				_, ok := resultSend[gift.ID]
				if ok {
					resultSend[gift.ID] += num
				} else {
					resultSend[gift.ID] = num
				}
			} else {
				_, ok := minusGiftSendNum[v.ID]

				if ok {
					minusGiftSendNum[v.ID] -= num
				} else {
					minusGiftSendNum[v.ID] = -num
				}
			}
		}
		for id, num := range resultSend {
			ef, err := s.lottery.UpdatelotteryGiftNumSQL(c, id, num)
			if err != nil {
				log.Errorc(c, "addTimesAndGiftNum s.dao.UpdatelotteryGiftNumSQL id(%d) num(%d) error(%v)", id, num, err)
				continue
			}
			if ef != 0 {
				for i := 0; i < num; i++ {
					newGiftList = append(newGiftList, giftMap[id])
				}
			}
		}

	}

	return
}
