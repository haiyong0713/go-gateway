package brand

import (
	"context"
	xecode "go-common/library/ecode"
	"go-gateway/app/web-svr/activity/ecode"
	dao "go-gateway/app/web-svr/activity/interface/dao/brand"
	brandMdl "go-gateway/app/web-svr/activity/interface/model/brand"
	"testing"

	couponapi "git.bilibili.co/bapis/bapis-go/account/service/coupon"
	passportinfoapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	silverbulletapi "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
	vipresourceapi "git.bilibili.co/bapis/bapis-go/vip/resource/service"

	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAddCoupon(t *testing.T) {
	Convey("test addcoupon success and counpon type 1", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockDao := dao.NewMockDao(mockCtl)
		mockPassport := passportinfoapi.NewMockPassportUserClient(mockCtl)
		mockVipResource := vipresourceapi.NewMockResourceClient(mockCtl)
		mockSilverbullet := silverbulletapi.NewMockSilverbulletProxyClient(mockCtl)
		defer mockCtl.Finish()

		// mock
		infosMap := make(map[string]*silverbulletapi.RiskInfo)
		infosMap[strategyName] = &silverbulletapi.RiskInfo{Level: 0}
		mockSilverbullet.EXPECT().RiskInfo(Any(), Any()).Return(&silverbulletapi.RiskInfoReply{Infos: infosMap}, nil)
		mockDao.EXPECT().CacheAddCouponTimes(Any(), Any()).Return(int64(1), nil)
		mockPassport.EXPECT().CheckFreshUser(Any(), Any()).Return(&passportinfoapi.CheckFreshUserReply{IsNew: true}, nil)
		mockDao.EXPECT().CacheQPSLimit(Any(), Any()).Return(int64(1), nil)
		mockVipResource.EXPECT().ResourceUse(Any(), Any()).Return(&vipresourceapi.ResourceUseReply{}, nil)
		//
		s.dao = mockDao
		s.passportClient = mockPassport
		s.vipResourceClient = mockVipResource
		s.silverbulletClient = mockSilverbullet
		res, err := s.AddCoupon(context.Background(), 4111111, &brandMdl.FrontEndParams{
			IP:       "127.0.0.1",
			DeviceID: "",
			Ua:       "",
			API:      "",
			Referer:  "",
		})

		So(err, ShouldEqual, nil)
		So(res.CouponType, ShouldEqual, 1)
	}))
	Convey("test addcoupon success and counpon type 2", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockPassport := passportinfoapi.NewMockPassportUserClient(mockCtl)
		mockDao := dao.NewMockDao(mockCtl)
		mockCoupon := couponapi.NewMockCouponClient(mockCtl)
		mockSilverbullet := silverbulletapi.NewMockSilverbulletProxyClient(mockCtl)

		defer mockCtl.Finish()
		// mock
		infosMap := make(map[string]*silverbulletapi.RiskInfo)
		infosMap[strategyName] = &silverbulletapi.RiskInfo{Level: 0}
		mockDao.EXPECT().CacheAddCouponTimes(Any(), Any()).Return(int64(1), nil)
		mockSilverbullet.EXPECT().RiskInfo(Any(), Any()).Return(&silverbulletapi.RiskInfoReply{Infos: infosMap}, nil)
		mockPassport.EXPECT().CheckFreshUser(Any(), Any()).Return(&passportinfoapi.CheckFreshUserReply{IsNew: false}, nil)

		mockCoupon.EXPECT().AllowanceReceive(Any(), Any()).Return(nil, nil)
		mockCoupon.EXPECT().AllowanceReceive(Any(), Any()).Return(nil, nil)
		//
		s.dao = mockDao
		s.passportClient = mockPassport

		s.silverbulletClient = mockSilverbullet
		s.couponClient = mockCoupon
		res, err := s.AddCoupon(context.Background(), 4111111, &brandMdl.FrontEndParams{
			IP:       "127.0.0.1",
			DeviceID: "",
			Ua:       "",
			API:      "",
			Referer:  "",
		})

		So(err, ShouldEqual, nil)
		So(res.CouponType, ShouldEqual, 2)
	}))
	Convey("test addcoupon already get", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockDao := dao.NewMockDao(mockCtl)
		mockPassport := passportinfoapi.NewMockPassportUserClient(mockCtl)

		mockSilverbullet := silverbulletapi.NewMockSilverbulletProxyClient(mockCtl)
		defer mockCtl.Finish()

		infosMap := make(map[string]*silverbulletapi.RiskInfo)
		infosMap[strategyName] = &silverbulletapi.RiskInfo{Level: 0}
		mockSilverbullet.EXPECT().RiskInfo(Any(), Any()).Return(&silverbulletapi.RiskInfoReply{Infos: infosMap}, nil)
		mockPassport.EXPECT().CheckFreshUser(Any(), Any()).Return(&passportinfoapi.CheckFreshUserReply{IsNew: false}, nil)

		mockDao.EXPECT().CacheAddCouponTimes(Any(), Any()).Return(int64(2), nil)
		s.dao = mockDao
		s.passportClient = mockPassport

		s.silverbulletClient = mockSilverbullet

		_, err := s.AddCoupon(context.Background(), 4111111, &brandMdl.FrontEndParams{
			IP:       "127.0.0.1",
			DeviceID: "",
			Ua:       "",
			API:      "",
			Referer:  "",
		})

		So(err, ShouldEqual, ecode.ActivityBrandAwardOnceErr)
	}))

	Convey("test addcoupon risk error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockDao := dao.NewMockDao(mockCtl)
		mockPassport := passportinfoapi.NewMockPassportUserClient(mockCtl)

		mockSilverbullet := silverbulletapi.NewMockSilverbulletProxyClient(mockCtl)

		mockVipResource := vipresourceapi.NewMockResourceClient(mockCtl)

		defer mockCtl.Finish()
		infosMap := make(map[string]*silverbulletapi.RiskInfo)
		infosMap[strategyName] = &silverbulletapi.RiskInfo{Level: 3}
		mockSilverbullet.EXPECT().RiskInfo(Any(), Any()).Return(&silverbulletapi.RiskInfoReply{Infos: infosMap}, nil)
		mockDao.EXPECT().CacheAddCouponTimes(Any(), Any()).Return(int64(1), nil)
		mockPassport.EXPECT().CheckFreshUser(Any(), Any()).Return(&passportinfoapi.CheckFreshUserReply{IsNew: true}, nil)

		mockDao.EXPECT().CacheSetMinusCouponTimes(Any(), Any()).Return(nil)

		s.dao = mockDao
		s.passportClient = mockPassport

		s.silverbulletClient = mockSilverbullet
		s.vipResourceClient = mockVipResource
		_, err := s.AddCoupon(context.Background(), 4111111, &brandMdl.FrontEndParams{
			IP:       "127.0.0.1",
			DeviceID: "",
			Ua:       "",
			API:      "",
			Referer:  "",
		})

		So(err, ShouldEqual, ecode.ActivityBrandRiskErr)
	}))

	Convey("test addcoupon minus error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockDao := dao.NewMockDao(mockCtl)
		mockPassport := passportinfoapi.NewMockPassportUserClient(mockCtl)

		mockSilverbullet := silverbulletapi.NewMockSilverbulletProxyClient(mockCtl)
		mockVipResource := vipresourceapi.NewMockResourceClient(mockCtl)
		mockCoupon := couponapi.NewMockCouponClient(mockCtl)

		defer mockCtl.Finish()
		infosMap := make(map[string]*silverbulletapi.RiskInfo)
		infosMap[strategyName] = &silverbulletapi.RiskInfo{Level: 0}
		mockSilverbullet.EXPECT().RiskInfo(Any(), Any()).Return(&silverbulletapi.RiskInfoReply{Infos: infosMap}, nil)
		mockDao.EXPECT().CacheAddCouponTimes(Any(), Any()).Return(int64(1), nil)
		mockPassport.EXPECT().CheckFreshUser(Any(), Any()).Return(&passportinfoapi.CheckFreshUserReply{IsNew: true}, nil)

		mockDao.EXPECT().CacheQPSLimit(Any(), Any()).Return(int64(1), nil)
		mockVipResource.EXPECT().ResourceUse(Any(), Any()).Return(&vipresourceapi.ResourceUseReply{}, xecode.New(brandMdl.VipBatchNotEnoughErr))
		mockCoupon.EXPECT().AllowanceReceive(Any(), Any()).Return(nil, xecode.New(brandMdl.VipBatchNotEnoughErr))
		mockCoupon.EXPECT().AllowanceReceive(Any(), Any()).Return(nil, nil)
		mockDao.EXPECT().CacheSetMinusCouponTimes(Any(), Any()).Return(ecode.ActivityUpVipLimit)
		s.passportClient = mockPassport
		s.dao = mockDao
		s.silverbulletClient = mockSilverbullet
		s.couponClient = mockCoupon
		s.vipResourceClient = mockVipResource

		_, err := s.AddCoupon(context.Background(), 4111111, &brandMdl.FrontEndParams{
			IP:       "127.0.0.1",
			DeviceID: "",
			Ua:       "",
			API:      "",
			Referer:  "",
		})

		So(err, ShouldEqual, ecode.ActivityBrandCouponErr)
	}))

	Convey("test addcoupon add times error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockPassport := passportinfoapi.NewMockPassportUserClient(mockCtl)

		mockDao := dao.NewMockDao(mockCtl)
		mockSilverbullet := silverbulletapi.NewMockSilverbulletProxyClient(mockCtl)
		defer mockCtl.Finish()

		infosMap := make(map[string]*silverbulletapi.RiskInfo)
		infosMap[strategyName] = &silverbulletapi.RiskInfo{Level: 0}
		mockSilverbullet.EXPECT().RiskInfo(Any(), Any()).Return(&silverbulletapi.RiskInfoReply{Infos: infosMap}, nil)
		mockDao.EXPECT().CacheAddCouponTimes(Any(), Any()).Return(int64(2), xecode.New(brandMdl.VipBatchNotEnoughErr))
		mockPassport.EXPECT().CheckFreshUser(Any(), Any()).Return(&passportinfoapi.CheckFreshUserReply{IsNew: true}, nil)

		s.dao = mockDao
		s.passportClient = mockPassport

		s.silverbulletClient = mockSilverbullet

		_, err := s.AddCoupon(context.Background(), 4111111, &brandMdl.FrontEndParams{
			IP:       "127.0.0.1",
			DeviceID: "",
			Ua:       "",
			API:      "",
			Referer:  "",
		})

		So(err, ShouldNotBeNil)
	}))

	Convey("test addcoupon vip user error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockPassport := passportinfoapi.NewMockPassportUserClient(mockCtl)

		mockDao := dao.NewMockDao(mockCtl)
		mockSilverbullet := silverbulletapi.NewMockSilverbulletProxyClient(mockCtl)
		defer mockCtl.Finish()

		infosMap := make(map[string]*silverbulletapi.RiskInfo)
		infosMap[strategyName] = &silverbulletapi.RiskInfo{Level: 0}
		mockSilverbullet.EXPECT().RiskInfo(Any(), Any()).Return(&silverbulletapi.RiskInfoReply{Infos: infosMap}, nil)
		mockDao.EXPECT().CacheAddCouponTimes(Any(), Any()).Return(int64(1), nil)
		mockPassport.EXPECT().CheckFreshUser(Any(), Any()).Return(&passportinfoapi.CheckFreshUserReply{IsNew: true}, xecode.New(brandMdl.VipBatchNotEnoughErr))

		mockDao.EXPECT().CacheSetMinusCouponTimes(Any(), Any()).Return(nil)

		s.dao = mockDao
		s.silverbulletClient = mockSilverbullet
		s.passportClient = mockPassport

		_, err := s.AddCoupon(context.Background(), 4111111, &brandMdl.FrontEndParams{
			IP:       "127.0.0.1",
			DeviceID: "",
			Ua:       "",
			API:      "",
			Referer:  "",
		})

		So(err, ShouldNotBeNil)

	}))

	Convey("test addcoupon get risk error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockDao := dao.NewMockDao(mockCtl)
		mockPassport := passportinfoapi.NewMockPassportUserClient(mockCtl)

		mockSilverbullet := silverbulletapi.NewMockSilverbulletProxyClient(mockCtl)
		defer mockCtl.Finish()

		infosMap := make(map[string]*silverbulletapi.RiskInfo)
		infosMap[strategyName] = &silverbulletapi.RiskInfo{Level: 0}
		mockSilverbullet.EXPECT().RiskInfo(Any(), Any()).Return(&silverbulletapi.RiskInfoReply{Infos: infosMap}, xecode.New(brandMdl.VipBatchNotEnoughErr))
		mockDao.EXPECT().CacheAddCouponTimes(Any(), Any()).Return(int64(1), nil)
		mockPassport.EXPECT().CheckFreshUser(Any(), Any()).Return(&passportinfoapi.CheckFreshUserReply{IsNew: true}, nil)

		mockDao.EXPECT().CacheSetMinusCouponTimes(Any(), Any()).Return(nil)

		s.dao = mockDao
		s.silverbulletClient = mockSilverbullet
		s.passportClient = mockPassport

		_, err := s.AddCoupon(context.Background(), 4111111, &brandMdl.FrontEndParams{
			IP:       "127.0.0.1",
			DeviceID: "",
			Ua:       "",
			API:      "",
			Referer:  "",
		})

		So(err, ShouldNotBeNil)
	}))

	Convey("test addcoupon get risk map error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockDao := dao.NewMockDao(mockCtl)
		mockPassport := passportinfoapi.NewMockPassportUserClient(mockCtl)

		mockSilverbullet := silverbulletapi.NewMockSilverbulletProxyClient(mockCtl)
		defer mockCtl.Finish()

		infosMap := make(map[string]*silverbulletapi.RiskInfo)
		mockSilverbullet.EXPECT().RiskInfo(Any(), Any()).Return(&silverbulletapi.RiskInfoReply{Infos: infosMap}, nil)
		mockDao.EXPECT().CacheAddCouponTimes(Any(), Any()).Return(int64(1), nil)
		mockDao.EXPECT().CacheSetMinusCouponTimes(Any(), Any()).Return(nil)
		mockPassport.EXPECT().CheckFreshUser(Any(), Any()).Return(&passportinfoapi.CheckFreshUserReply{IsNew: true}, nil)

		s.dao = mockDao
		s.silverbulletClient = mockSilverbullet
		s.passportClient = mockPassport

		_, err := s.AddCoupon(context.Background(), 4111111, &brandMdl.FrontEndParams{
			IP:       "127.0.0.1",
			DeviceID: "",
			Ua:       "",
			API:      "",
			Referer:  "",
		})

		So(err, ShouldNotBeNil)
	}))

	Convey("test addcoupon qps error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockPassport := passportinfoapi.NewMockPassportUserClient(mockCtl)

		mockDao := dao.NewMockDao(mockCtl)
		mockVipResource := vipresourceapi.NewMockResourceClient(mockCtl)
		mockSilverbullet := silverbulletapi.NewMockSilverbulletProxyClient(mockCtl)
		defer mockCtl.Finish()

		// mock
		infosMap := make(map[string]*silverbulletapi.RiskInfo)
		infosMap[strategyName] = &silverbulletapi.RiskInfo{Level: 0}
		mockSilverbullet.EXPECT().RiskInfo(Any(), Any()).Return(&silverbulletapi.RiskInfoReply{Infos: infosMap}, nil)
		mockDao.EXPECT().CacheAddCouponTimes(Any(), Any()).Return(int64(1), nil)
		mockDao.EXPECT().CacheQPSLimit(Any(), Any()).Return(int64(151), nil)
		mockDao.EXPECT().CacheSetMinusCouponTimes(Any(), Any()).Return(nil)
		mockPassport.EXPECT().CheckFreshUser(Any(), Any()).Return(&passportinfoapi.CheckFreshUserReply{IsNew: true}, nil)

		//
		s.dao = mockDao
		s.vipResourceClient = mockVipResource
		s.silverbulletClient = mockSilverbullet
		s.passportClient = mockPassport

		_, err := s.AddCoupon(context.Background(), 4111111, &brandMdl.FrontEndParams{
			IP:       "127.0.0.1",
			DeviceID: "",
			Ua:       "",
			API:      "",
			Referer:  "",
		})

		So(err, ShouldEqual, ecode.ActivityQPSLimitErr)
	}))

	Convey("test addcoupon qps get error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockDao := dao.NewMockDao(mockCtl)
		mockPassport := passportinfoapi.NewMockPassportUserClient(mockCtl)

		mockVipResource := vipresourceapi.NewMockResourceClient(mockCtl)
		mockSilverbullet := silverbulletapi.NewMockSilverbulletProxyClient(mockCtl)
		defer mockCtl.Finish()

		// mock
		infosMap := make(map[string]*silverbulletapi.RiskInfo)
		infosMap[strategyName] = &silverbulletapi.RiskInfo{Level: 0}
		mockSilverbullet.EXPECT().RiskInfo(Any(), Any()).Return(&silverbulletapi.RiskInfoReply{Infos: infosMap}, nil)
		mockDao.EXPECT().CacheAddCouponTimes(Any(), Any()).Return(int64(1), nil)
		mockPassport.EXPECT().CheckFreshUser(Any(), Any()).Return(&passportinfoapi.CheckFreshUserReply{IsNew: true}, nil)

		mockDao.EXPECT().CacheQPSLimit(Any(), Any()).Return(int64(1), xecode.New(brandMdl.VipBatchNotEnoughErr))
		mockDao.EXPECT().CacheSetMinusCouponTimes(Any(), Any()).Return(nil)
		//
		s.dao = mockDao
		s.vipResourceClient = mockVipResource
		s.passportClient = mockPassport

		s.silverbulletClient = mockSilverbullet
		_, err := s.AddCoupon(context.Background(), 4111111, &brandMdl.FrontEndParams{
			IP:       "127.0.0.1",
			DeviceID: "",
			Ua:       "",
			API:      "",
			Referer:  "",
		})

		So(err, ShouldEqual, ecode.ActivityQPSLimitErr)
	}))

}
