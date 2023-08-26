package rewards

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-common/library/naming/discovery"
	"go-common/library/net/http/blademaster/resolver"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/tool"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/go-playground/validator.v9"

	http "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-gateway/app/web-svr/activity/interface/conf"
	dao "go-gateway/app/web-svr/activity/interface/dao/rewards"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
	suit "go-main/app/account/usersuit/service/api"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	blackList "git.bilibili.co/bapis/bapis-go/account/service/account_control_plane"
	accountCoupon "git.bilibili.co/bapis/bapis-go/account/service/coupon"
	class "git.bilibili.co/bapis/bapis-go/cheese/service/coupon"
	garbDiy "git.bilibili.co/bapis/bapis-go/garb/diy/service"
	garb "git.bilibili.co/bapis/bapis-go/garb/service"
	user "git.bilibili.co/bapis/bapis-go/passport/service/user"
	garbCoupon "git.bilibili.co/bapis/bapis-go/vas/coupon/service"
	vip "git.bilibili.co/bapis/bapis-go/vip/resource/service"

	bindDao "go-gateway/app/web-svr/activity/interface/dao/bind"
	bind "go-gateway/app/web-svr/activity/interface/service/bind"
)

var Client = &service{}

// 奖励类型和发奖函数的映射关系
var awardsSendFuncMap = make(map[string] /*awardType*/ func(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (map[string]string, error))

// 奖励类型和奖励配置的映射关系
var awardsConfigMap = make(map[string] /*awardType*/ interface{})

// 奖励类型和发奖前的检查函数的映射关系
var awardsCheckFuncMap = make(map[string] /*awardType*/ func(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) error)

// service ...
type service struct {
	c *conf.Config

	dao *dao.Dao
	//awardsConfigs: 奖励具体配置, 包含奖励类型, 展示名字, 调用发奖方所需的参数.
	//调用不同发奖方所需的参数不同, 泛型实现过于繁琐, 所以使用JsonStr, 推迟到具体发奖函数执行时进行json反序列化.
	awardsConfigs atomic.Value

	//对外链接
	httpClient            *http.Client
	discoveryHttpClient   *http.Client
	comicClient           *http.Client
	suitClient            suit.UsersuitClient
	garbClient            garb.GarbClient
	garbDiyClient         garbDiy.GarbDiyClient
	garbCouponClient      garbCoupon.PlatformCouponClient
	accountClient         account.AccountClient
	accountCouponClient   accountCoupon.CouponClient
	backListClient        blackList.AccountControlPlaneClient
	userAccClient         user.PassportUserClient
	classClient           class.CouponClient
	vipClient             vip.ResourceClient
	broadcastURL          string
	messageURL            string
	msgKeyURL             string
	normalMsgURL          string
	comicBnj2021CouponURL string
	mallCouponURL         string
	mallPrizeURL          string
	liveGoldURL           string
	liveDataBusPub        *databus.Databus
	actPlatPub            *databus.Databus
	bindSvr               *bind.Service
	bindDao               *bindDao.Dao

	//校验工具
	v          *validator.Validate
	validateMu sync.Mutex
}

// New ...
func Init(c *conf.Config) {
	Client.c = c
	Client.dao = dao.New(c)
	Client.awardsConfigs = atomic.Value{}
	Client.httpClient = http.NewClient(c.HTTPClientRewards)
	Client.discoveryHttpClient = http.NewClient(c.HTTPClientRewards, http.SetResolver(resolver.New(nil, discovery.Builder())))
	Client.comicClient = http.NewClient(c.HTTPClientComic)
	Client.comicBnj2021CouponURL = c.Host.Comic + _cartoonBnj2021URI
	Client.mallCouponURL = c.Host.Mall + _mallCouponURI
	Client.mallPrizeURL = c.Host.Mall + _mallPrizeURI
	Client.liveGoldURL = c.Host.LiveCo + _liveGoldUri
	Client.liveDataBusPub = databus.New(c.DataBus.LiveItemPub)
	Client.actPlatPub = databus.New(c.DataBus.ActPlatPub)
	Client.v = validator.New()

	ac, err := Client.dao.GetAwardMap(context.Background(), 0)
	if err != nil {
		//TODO: panic here
	} else {
		Client.awardsConfigs.Store(&Configs{AwardMap: ac})
	}
	go Client.updateAwardConfigLoop()

	if Client.garbClient, err = garb.NewClient(nil); err != nil {
		panic(err)
	}
	if Client.garbCouponClient, err = garbCoupon.NewClient(nil); err != nil {
		panic(err)
	}
	if Client.garbDiyClient, err = garbDiy.NewClient(nil); err != nil {
		panic(err)
	}
	if Client.accountClient, err = account.NewClient(nil); err != nil {
		panic(err)
	}
	if Client.backListClient, err = blackList.NewClient(nil); err != nil {
		panic(err)
	}
	if Client.userAccClient, err = user.NewClient(nil); err != nil {
		panic(err)
	}
	if Client.accountCouponClient, err = accountCoupon.NewClient(nil); err != nil {
		panic(err)
	}
	if Client.classClient, err = class.NewClient(nil); err != nil {
		panic(err)
	}
	if Client.vipClient, err = vip.NewClient(nil); err != nil {
		panic(err)
	}

	Client.bindSvr = bind.New(c)
	Client.bindDao = bindDao.New(c)

	return
}

// SendAwardById: 根据奖励Id发放奖励
func (s *service) SendAwardById(ctx context.Context, mid int64, uniqueID, business string, awardId int64, updateCache bool) (info *api.RewardsSendAwardReply, err error) {
	info = &api.RewardsSendAwardReply{}
	ac, err := s.GetAwardConfigById(ctx, awardId)
	if err != nil {
		return
	}
	awardFunc, ok := awardsSendFuncMap[ac.Type]
	if !ok {
		log.Errorc(ctx, "award config id %v no such award type: %v", ac.Id, ac.Type)
		tool.Metric4RewardFail.WithLabelValues([]string{ac.Type, "get_func"}...).Inc()
		err = ecode.RewardsAwardSendFail
		return
	}
	err = s.dao.SendAwardByFunc(ctx, mid, uniqueID, business, ac, func() (map[string]string, error) {
		return awardFunc(ctx, ac, mid, uniqueID, business)
	})
	if err != nil {
		tool.Metric4RewardFail.WithLabelValues([]string{ac.Type, "award"}...).Inc()
		return
	}
	info.Mid = mid
	info.AwardId = ac.Id
	info.Name = ac.Name
	info.ActivityId = ac.ActivityId
	info.ActivityName = ac.ActivityName
	info.Type = ac.Type
	info.Icon = ac.IconUrl
	info.ReceiveTime = time.Now().Unix()
	info.ExtraInfo = ac.ExtraInfo
	if updateCache {
		s.dao.AddSingleRecordCache(mid, ac.ActivityId, info)
	}
	tool.Metric4RewardSuccess.WithLabelValues([]string{ac.Type}...).Inc()
	return
}

func (s *service) IsAwardAlreadySend(ctx context.Context, mid, awardId int64, uniqueId string) (send bool, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.IsAwardAlreadySend error: %v", err)
		}
	}()
	c, err := s.GetAwardConfigById(ctx, awardId)
	if err != nil {
		return
	}
	send, err = s.dao.IsAwardAlreadySend(ctx, mid, c.ActivityId, uniqueId)
	return
}

func (s *service) GetAwardRecordByMidAndActivityId(ctx context.Context, mid int64, activityIds []int64, limit int64) (res []*model.AwardSentInfo, err error) {
	return s.dao.GetAwardRecordByMidAndActivityIdsFromDB(ctx, mid, activityIds, limit)
}

func (s *service) GetAwardRecordByMidAndActivityIdWithCache(ctx context.Context, mid int64, activityIds []int64, limit int64) (res []*model.AwardSentInfo, err error) {
	return s.dao.GetAwardRecordByMidAndActivityIdsWithCache(ctx, mid, activityIds, limit)
}

func (s *service) GetAwardCountByMidAndActivityId(ctx context.Context, mid, activityId int64) (res int64, err error) {
	return s.dao.GetAwardCountByMidAndActivityIdWithCache(ctx, mid, activityId)
}

func (s *service) GetAwardCdKeysById(ctx context.Context, mid, id int64) (res []*model.CdKeyInfo, err error) {
	return s.dao.GetCdKeyById(ctx, mid, id)
}

func (s *service) GetAwardCdKeysByActivityId(ctx context.Context, mid, activityId int64) (res []*model.CdKeyInfo, err error) {
	return s.dao.GetCdKeyByActivityId(ctx, mid, activityId)
}

// SendAwardByIdAsync: 根据奖励Id发放奖励(异步)
// updateDB: 是否更新DB, 可避免消息队列丢失导致丢数据(目前恒为true)
// updateDB=true: 一致性高,容忍消息丢失
// updateDB=false: 性能高,需要调用方自身提供额外的对账机制
func (s *service) SendAwardByIdAsync(ctx context.Context, mid int64, uniqueID, business string, awardId int64, updateCache, updateDB bool) (info *api.RewardsSendAwardReply, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "s.SendAwardByIdAsync error: %v", err)
		}
	}()
	info = &api.RewardsSendAwardReply{}
	ac, err := s.GetAwardConfigById(ctx, awardId)
	if err != nil {
		err = ecode.ActivityAwardNotExpected
		return
	}
	send, err := s.dao.IsAwardAlreadySend(ctx, mid, ac.ActivityId, uniqueID)
	if err != nil {
		tool.Metric4RewardFail.WithLabelValues([]string{ac.Type, "check_send"}...).Inc()
		err = ecode.RewardsAwardSendFail
		return
	}
	if send {
		tool.Metric4RewardFail.WithLabelValues([]string{ac.Type, "already_send"}...).Inc()
		err = nil
		log.Infoc(ctx, "RewardsAwardAlreadySent for mid: %v, uniqueId: %v, business: %v, awardId: %v",
			mid, uniqueID, business, awardId)
		return
	}
	err = s.dao.InitAwardSentRecord(ctx, mid, uniqueID, business, ac /*updateDB*/, true)
	if err != nil {
		return
	}
	msg := &model.AsyncSendingAwardInfo{
		Mid:       mid,
		UniqueId:  uniqueID,
		Business:  business,
		AwardId:   awardId,
		AwardType: ac.Type,
		SendTime:  time.Now().Unix(),
	}
	bs, err := json.Marshal(msg)
	if err != nil {
		err = ecode.RewardsAwardSendFail
		return
	}
	for i := 0; i < 3; i++ {
		_, err = component.BackUpMQ.Do(ctx, "LPUSH", backoffKey4AwardSending(ac.Type, mid), string(bs))
		if err == nil {
			break
		}
	}

	if err == nil {
		info.Mid = mid
		info.AwardId = ac.Id
		info.Name = ac.Name
		info.ActivityId = ac.ActivityId
		info.ActivityName = ac.ActivityName
		info.Type = ac.Type
		info.Icon = ac.IconUrl
		info.ReceiveTime = time.Now().Unix()
		info.ExtraInfo = deepCopyMap(ac.ExtraInfo)
	}
	if updateCache {
		s.dao.AddSingleRecordCache(mid, ac.ActivityId, info)
	}
	return
}

func (s *service) AddActivityAddress(ctx context.Context, mid, activityId, addressId int64) (err error) {
	//检查对应活动下是否中过实体奖励
	entityAwardId := int64(0)
	list, err := s.dao.GetAwardRecordByMidAndActivityIdsFromDB(ctx, mid, []int64{activityId}, 50)
	for _, i := range list {
		if i.Type == rewardTypeEntity {
			entityAwardId = i.AwardId
			break
		}
	}
	if entityAwardId == 0 {
		err = ecode.ActivityAddrNotNeed
		return
	}
	awardConfig, err := s.GetAwardConfigById(ctx, entityAwardId)
	if err != nil {
		log.Errorc(ctx, "AddLotteryAddress s.GetAwardConfigById awardId(%d) mid(%d) error(%v)", entityAwardId, mid, err)
		err = ecode.ActivityAddrAddFail
		return
	}
	// 校验传输的地址id是否有效
	addr, err := s.getMemberAddress(ctx, addressId, mid)
	if err != nil {
		log.Errorc(ctx, "AddLotteryAddress s.getMemberAddress id(%d) mid(%d) error(%v)", addressId, mid, err)
		err = ecode.ActivityAddrAddFail
		return
	}
	if addr == nil || addr.ID == 0 {
		err = ecode.ActivityAddrAddFail
		return
	}
	//查看地址是否已填写
	aId, err := s.dao.IsActivityAddressExists(ctx, mid, activityId)
	if err != nil {
		err = ecode.ActivityAddrAddFail
		log.Errorc(ctx, "AddLotteryAddress s.dao.IsActivityAddressExists id(%d) mid(%d) error(%v)", addressId, mid, err)
		return
	}
	if aId != 0 {
		err = ecode.ActivityAddrHasAdd
		return
	}
	err = s.dao.AddActivityAddress(ctx, mid, awardConfig, addressId)
	return
}

func (s *service) GetActivityAddress(ctx context.Context, mid, activityId int64) (res *model.AddressInfo, err error) {
	res = &model.AddressInfo{}
	defer func() {
		if err != nil {
			log.Errorc(ctx, "GetActivityAddress error: %v", err)
		}
	}()
	addressId, err := s.dao.IsActivityAddressExists(ctx, mid, activityId)
	if err != nil {
		return
	}
	if addressId == 0 {
		err = ecode.ActivityAddrNotAdd
		return
	}
	res, err = s.getMemberAddress(ctx, addressId, mid)
	if err != nil {
		log.Errorc(ctx, "GetActivityAddress s.getMemberAddress id(%d) mid(%d) error(%v)", addressId, mid, err)
		return
	}
	return
}

func (s *service) AddTmpAwardSentInfoToCache(ctx context.Context, mid, awardId int64) (err error) {
	defer func() {
		if err != nil {
		}
		log.Errorc(ctx, "AddTmpAwardSentInfoToCache error: %v", err)
	}()
	info, err := s.GetAwardSentInfoById(ctx, awardId, mid)
	if err != nil {
		return
	}
	err = s.dao.AddSingleRecordCache(mid, info.ActivityId, info)
	return
}

// RetrySendAwardById: 根据奖励Id重试发放奖励
func (s *service) RetrySendAwardById(ctx context.Context, mid int64, uniqueID, business string, awardId int64) (err error) {
	ac, err := s.GetAwardConfigById(ctx, awardId)
	if err != nil {
		return
	}
	awardFunc, ok := awardsSendFuncMap[ac.Type]
	if !ok {
		err = fmt.Errorf("can not found award config for type: %v", ac.Type)
		return
	}
	err = s.dao.RetrySendAwardByFunc(ctx, mid, uniqueID, business, ac, func() (map[string]string, error) {
		return awardFunc(ctx, ac, mid, uniqueID, business)
	})
	if err != nil {
		return
	}
	return
}

func (s *service) ListAwardType(ctx context.Context, c *api.RewardsListAwardTypeReq) (reply *api.RewardsListAwardTypeReply, err error) {
	reply = &api.RewardsListAwardTypeReply{Types: make([]string, 0)}
	for typ := range awardsConfigMap {
		t := typ
		reply.Types = append(reply.Types, t)
	}
	return
}

func (s *service) RewardsCheckSentStatusReq(ctx context.Context, req *api.RewardsCheckSentStatusReq) (reply *api.RewardsCheckSentStatusResp, err error) {
	reply = &api.RewardsCheckSentStatusResp{}
	var state int64
	if req.AwardId != 0 {
		reply.Result, err = s.IsAwardAlreadySend(ctx, req.Mid, req.AwardId, req.UniqueId)
	} else {
		state, err = s.dao.GetAwardSendStatusInDBByMidAndUniqId(ctx, req.Mid, req.UniqueId)
		reply.Result = state == dao.AwardSentStateOK
	}
	return
}

func (s *service) AwardSendPreCheck(ctx context.Context, mid int64, uniqueID, business string, awardId int64) (err error) {
	ac, err := s.GetAwardConfigById(ctx, awardId)
	if err != nil {
		return
	}
	//通用的检查, 如库存
	checkFunc, ok := awardsCheckFuncMap[ac.Type]
	if !ok {
		//未定义检查函数, 返回nil
		return
	}
	err = checkFunc(ctx, ac, mid, uniqueID, business)
	return
}
