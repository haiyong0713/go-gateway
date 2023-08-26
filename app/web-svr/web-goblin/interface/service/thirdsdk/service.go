package thirdsdk

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/web-goblin/interface/conf"
	"go-gateway/app/web-svr/web-goblin/interface/dao/thirdsdk"
	model "go-gateway/app/web-svr/web-goblin/interface/model/thirdsdk"
)

// Service service struct.
type Service struct {
	c   *conf.Config
	dao *thirdsdk.Dao
}

// New new service.
func New(c *conf.Config) *Service {
	s := &Service{
		c:   c,
		dao: thirdsdk.New(c),
	}
	return s
}

func (s *Service) AuthorBindState(ctx context.Context, mid int64) (*model.Author, error) {
	res := &model.Author{
		Mid: mid,
	}
	ok, err := s.dao.Invited(ctx, mid)
	if err != nil {
		return nil, err
	}
	if !ok {
		return res, nil
	}
	bind, err := s.dao.UserBind(ctx, mid)
	if err != nil {
		return nil, err
	}
	log.Info("authbindstate mid;%v,bind:%+v", mid, bind)
	if bind == nil {
		return res, nil
	}
	// invited 邀请并授权用户 true：白名单用户并且授权
	// bind_state 绑定状态，0：未绑定，1：已绑定
	// check_state 审核状态，0：未认证，1：已通过，3：审核中，4：已驳回
	switch bind.AuthorizationStatus {
	case "AUTHORIZED": //已授权
		res.Invited = true
	// case "UNAUTHORIZED", "CANCELED": //未授权或授权被取消
	default:
		return res, nil
	}
	switch bind.BindStatus {
	case "BINDED": //已绑定
		res.BindState = 1
		// case "UNBINDED": //未绑定
	}
	switch bind.VerificationStatus {
	case "UNVERIFIED": //未认证
		res.CheckState = 0
	case "VERIFYING": //认证中（材料已提交）
		res.CheckState = 3
	case "VERIFIED": //认证成功
		res.CheckState = 1
	case "FAILED": //认证失败
		res.CheckState = 4
		res.RefuseReason = bind.Reason
	}
	return res, nil
}
