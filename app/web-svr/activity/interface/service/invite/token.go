package invite

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	mdl "go-gateway/app/web-svr/activity/interface/model/invite"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"

	acp "git.bilibili.co/bapis/bapis-go/account/service/account_control_plane"
	passportUser "git.bilibili.co/bapis/bapis-go/passport/service/user"
	sp "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
)

const (
	getTokenAPI          = "/x/activity/invite/token"
	getTokenStrategyName = "activity_get_token"
	faceToken            = 1
)

// Token create token.
func (s *Service) Token(ctx context.Context, mid int64, activityID string, tp int64, source int64, frontEndParams *mdl.BaseInfo) (rs *mdl.TokenResp, err error) {
	rs = new(mdl.TokenResp)
	reply, err := s.userDetail(ctx, mid)
	if err != nil {
		return
	}
	if !reply.BindTel {
		err = ecode.ActivityInviterNoBindTelGetErr
		return
	}

	// risk check
	riskReq := &sp.RiskInfoReq{
		Mid:          mid,
		Ip:           frontEndParams.IP,
		DeviceId:     frontEndParams.Buvid,
		Ua:           frontEndParams.UA,
		Api:          getTokenAPI,
		Referer:      frontEndParams.Referer,
		Origin:       frontEndParams.Origin,
		StrategyName: []string{getTokenStrategyName},
	}
	isBlocked, err := s.isBlocked(ctx, riskReq)
	if err != nil {
		return
	}
	if isBlocked {
		err = ecode.ActivityInviterShareNotAllow
		return
	}

	rs.Token, err = s.createToken(ctx, mid, tp, activityID, source)
	if err == nil {
		s.fan.Do(ctx, func(ctx context.Context) {
			s.retry(func() error {
				if err := s.invite.AddUserShareLog(ctx, mid, activityID, time.Now()); err != nil {
					return err
				}
				return s.invite.ClearUserShareLogCache(ctx, mid, activityID)
			})
		})
	}
	return
}

func (s *Service) retry(f func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		err = f()
		if err == nil {
			return nil
		}
		log.Error("retry error:%v", err)
		time.Sleep(200 * time.Millisecond)
	}
	return err
}

func (s *Service) createToken(ctx context.Context, mid int64, tp int64, activityUID string, source int64) (token string, err error) {
	var (
		tokenID int64
		st      time.Duration
	)
	if tp == faceToken {
		st, _ = time.ParseDuration(s.faceTokenExpire)
	} else {
		st, _ = time.ParseDuration(s.tokenExpire)
	}
	expire := time.Now().Add(st).Unix()
	token, err = s.invite.CacheGetMidToken(ctx, mid, tp, activityUID, source)
	if err == nil {
		return token, nil
	}
	if err != nil {
		log.Errorc(ctx, "s.invite.CacheGetMidToken error(%v)", err)
	}
	p := fmt.Sprintf("%d%d%s%d%s", mid, tp, activityUID, source, s.tokenSalt)
	token = s.tokenMd5(p)
	_, err = s.invite.AddToken(ctx, mid, tp, expire, activityUID, token, source)
	if err != nil {
		token = ""
		return token, err
	}
	_ = s.invite.AddCacheToken(ctx, token, &mdl.FiToken{
		ID:         tokenID,
		Mid:        mid,
		Token:      token,
		ExpireTime: expire,
		Tp:         tp,
		Source:     source,
	})
	err1 := s.invite.CacheMidToken(ctx, mid, tp, activityUID, token, source)
	if err1 != nil {
		log.Errorc(ctx, "s.invite.CacheMidToken err(%v)", err1)
	}
	return
}

func (s *Service) tokenMd5(p string) string {
	hasher := md5.New()
	hasher.Write([]byte(p))
	return hex.EncodeToString(hasher.Sum(nil))
}

// userDetail 用户详情
func (s *Service) userDetail(ctx context.Context, mid int64) (*passportUser.UserDetailReply, error) {
	req := &passportUser.UserDetailReq{Mid: mid}
	reply, err := s.passportClient.UserDetail(ctx, req)
	if err != nil {
		log.Errorc(ctx, "PassportUser.UserDetail mid: %d error:(%v)", req.Mid, err)
		return nil, err
	}
	return reply, nil
}

func (s *Service) telHash(tel string) string {
	hash := md5.New()
	hash.Write([]byte(tel))
	if s.c.Invite.TelSalt != "" {
		hash.Write([]byte(s.c.Invite.TelSalt))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *Service) isBlocked(ctx context.Context, riskReq *sp.RiskInfoReq) (bool, error) {
	// risk check
	isAllowed := s.isAllowedByRisk(ctx, riskReq)
	if !isAllowed {
		log.Warn("is block by risk, mid: %d", riskReq.Mid)
		return true, nil
	}
	isBlocked, err := s.isMidBlocked(ctx, riskReq.Mid, riskReq.Ip)
	if err != nil {
		return false, err
	}
	if isBlocked {
		log.Warn("is block by account control, mid: %d", riskReq.Mid)
		return true, nil
	}
	return false, nil
}

func (s *Service) isAllowedByRisk(ctx context.Context, riskReq *sp.RiskInfoReq) bool {
	reply, err := s.silverbulletClient.RiskInfo(ctx, riskReq)
	if err != nil {
		return true
	}
	riskInfo, ok := reply.Infos[riskReq.StrategyName[0]]
	if !ok {
		return true
	}
	if riskInfo.Level != 0 {
		return false
	}
	return true
}

// isMidBlocked check if mid is block
func (s *Service) isMidBlocked(c context.Context, mid int64, ip string) (bool, error) {
	req := acp.HasControlRoleReq{Mid: mid, ControlRole: []string{"silence", "all_block", "risk_reg", "risk_login"}}
	reply, err := s.acpClient.HasControlRole(c, &req)
	if err != nil {
		log.Errorc(c, "HasControlRoleReq mid:%d error: (%v)", mid, err)
		return false, err
	}
	for _, cs := range reply.ControlRoleStatus {
		if cs.HasRole {
			return true, nil
		}
	}
	return false, nil
}

// checkToken check token.
func (s *Service) checkToken(ctx context.Context, token string) (*mdl.FiToken, error) {
	fiToken, err := s.invite.CacheToken(ctx, token)
	if fiToken != nil && err == nil {
		return fiToken, nil
	}
	fiToken, err = s.invite.SelToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if fiToken == nil {
		return nil, ecode.ActivityTokenNotFind
	}
	if err = s.invite.AddCacheToken(ctx, token, fiToken); err != nil {
		log.Errorc(ctx, "checkToken AddCacheToken error: %v", err)
	}
	return fiToken, nil
}

// Inviter ...
func (s *Service) Inviter(c context.Context, token string) (*mdl.InviterReply, error) {
	fiToken, err := s.checkToken(c, token)
	if err != nil {
		log.Errorc(c, "s.checkToken(%s)", token)
		return nil, ecode.ActivityTokenNotFind
	}
	midInfo, err := s.accClient.Info3(c, &accountapi.MidReq{Mid: fiToken.Mid})
	if err != nil {
		log.Errorc(c, "s.accClient.Info3: error(%v)", err)
		return nil, err
	}
	if midInfo == nil || midInfo.Info == nil {
		return nil, ecode.ActivityInviteMidNotFind
	}
	account := &mdl.Account{
		Mid:  midInfo.Info.Mid,
		Name: midInfo.Info.Name,
		Face: midInfo.Info.Face,
		Sign: midInfo.Info.Sign,
		Sex:  midInfo.Info.Sex,
	}
	res := &mdl.InviterReply{}
	res.Account = account
	return res, nil
}
