package brand

import (
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/brand"

	couponapi "git.bilibili.co/bapis/bapis-go/account/service/coupon"
	passportinfoapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	silverbulletapi "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
	vipresourceapi "git.bilibili.co/bapis/bapis-go/vip/resource/service"
	vipinfoapi "git.bilibili.co/bapis/bapis-go/vip/service/vipinfo"
)

// Service ...
type Service struct {
	c                      *conf.Config
	dao                    brand.Dao
	couponClient           couponapi.CouponClient
	vipInfoClient          vipinfoapi.VipInfoClient
	vipResourceClient      vipresourceapi.ResourceClient
	silverbulletClient     silverbulletapi.SilverbulletProxyClient
	passportClient         passportinfoapi.PassportUserClient
	CouponBatchToken       string
	CouponBatchToken2      string
	ResourceBatchToken     string
	ResourceAppkey         string
	CouponExperienceRemark string
	QPSLimitResourceCoupon int64
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: brand.New(c),
	}
	var err error

	if s.couponClient, err = couponapi.NewClient(c.CouponClient); err != nil {
		panic(err)
	}
	if s.vipInfoClient, err = vipinfoapi.NewClient(c.VipInfoClient); err != nil {
		panic(err)
	}
	if s.vipResourceClient, err = vipresourceapi.NewClient(c.ResourceClient); err != nil {
		panic(err)
	}
	if s.silverbulletClient, err = silverbulletapi.NewClient(c.SilverClient); err != nil {
		panic(err)
	}
	if s.passportClient, err = passportinfoapi.NewClient(c.PassportClient); err != nil {
		panic(err)
	}
	s.CouponBatchToken = c.Brand.CouponBatchToken
	s.ResourceBatchToken = c.Brand.ResourceBatchToken
	s.CouponBatchToken2 = c.Brand.CouponBatchToken2
	s.ResourceAppkey = c.Brand.ResourceAppkey
	s.CouponExperienceRemark = c.Brand.CouponExperienceRemark
	s.QPSLimitResourceCoupon = c.Brand.QPSLimitResourceCoupon
	return s
}

// Close ...
func (s *Service) Close() {
	s.dao.Close()
}
