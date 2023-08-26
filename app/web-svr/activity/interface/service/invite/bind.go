package invite

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"

	mdl "go-gateway/app/web-svr/activity/interface/model/invite"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"

	passportinfoapi "git.bilibili.co/bapis/bapis-go/passport/service/user"

	sp "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
)

const (
	inviteBindAPI          = "/x/activity/invite/bind"
	inviteBindStrategyName = "activity_bind_tel"
	countryCodeChina       = 86
	countryCodeChinaStr    = "86"
	mobilePattern          = "^((13[0-9])|(14[1,4,5,6,7,8])|(15[^4])|(16[2,5-7])|(17[0-8])|(18[0-9])|(19[^4]))\\d{8}$"
)

var (
	mobileReg, _ = regexp.Compile(mobilePattern)
)

// InviteBind bind mobile.
func (s *Service) InviteBind(ctx context.Context, req *mdl.BindReq) (res *mdl.BindReply, err error) {
	var (
		isNew            = false
		fiToken          *mdl.FiToken
		now              = time.Now()
		inviterMid       int64
		allInviteLogPara = &mdl.AllInviteLog{
			ActivityUID: req.ActivityID,
			Tel:         req.Tel,
			Token:       req.Token,
			InvitedTime: now.Unix(),
		}
	)
	defer func() {
		// 0默认状态，1成功，2非中国手机号，3活动检测失败，4token检测失败，5已绑定，6被风控，7手机号验新报错，8手机号非新用户，9绑定关系插入数据库报错
		s.invite.AddAllInviteLog(ctx, allInviteLogPara)

		_, err2 := s.userShareLog(ctx, inviterMid, req.ActivityID)
		if err2 != nil {
			log.Errorc(ctx, "s.userShareLog error(%v)", err)
		}
	}()
	// // china tel
	if !isChinaMobile(req.Tel) {
		err = ecode.ActivityInviterShareNotAllow
		allInviteLogPara.InviteStatus = 2
		log.Warn("InviteBind tel[%s] is not chinamobile, token[%s]", req.Tel, req.Token)
		return
	}

	// check token
	if fiToken, err = s.checkToken(ctx, req.Token); err != nil || fiToken == nil {
		allInviteLogPara.InviteStatus = 4
		return
	}
	inviterMid = fiToken.Mid
	allInviteLogPara.Mid = fiToken.Mid
	allInviteLogPara.Source = fiToken.Source

	// check if tel is already bind
	if err = s.checkIsAlreadyTmpBind(ctx, req.Tel); err != nil {
		if err == ecode.ActivityInviterTelIsOld {
			err = nil
		}
		allInviteLogPara.InviteStatus = 5
		log.Warn("InviteBind tel[%s] is already exist, mid[%d]", req.Tel, inviterMid)
		return
	}

	// risk
	riskReq := &sp.RiskInfoReq{
		Mid:          inviterMid,
		Ip:           req.IP,
		DeviceId:     req.Buvid,
		Ua:           req.UA,
		Api:          inviteBindAPI,
		Referer:      req.Referer,
		Origin:       req.Origin,
		StrategyName: []string{inviteBindStrategyName},
		ExtraData:    map[string]string{"token": req.Token, "invite_mid": strconv.FormatInt(inviterMid, 10)},
	}
	isAllowed := s.isAllowedByRisk(ctx, riskReq)
	if !isAllowed {
		err = ecode.ActivityInviterJoinNotAllow
		allInviteLogPara.InviteStatus = 6
		log.Warn("InviteBind tel[%s] is not allowed by risk mid[%d]", req.Tel, inviterMid)
		return
	}
	// tel is new
	if isNew, err = s.isNewTel(ctx, req.Tel, countryCodeChina); err != nil {
		allInviteLogPara.InviteStatus = 7
		return
	}
	if !isNew {
		allInviteLogPara.InviteStatus = 8
		log.Warn("InviteBind tel[%s] is old mid[%d]", req.Tel, inviterMid)
		//err = ecode.FissionTelIsOld
		s.invite.AddOldUserCache(ctx, s.telHash(req.Tel))
		// return
	}
	// insert into mysql
	inviteRel := &mdl.InviteRelation{
		Mid:         fiToken.Mid,
		ActivityUID: req.ActivityID,
		Tel:         req.Tel,
		TelHash:     s.telHash(req.Tel),
		Token:       req.Token,
		InvitedTime: now.Unix(),
		ExpireTime:  now.Add(time.Duration(s.c.Invite.BindExpire)).Unix(),
		IP:          req.IP,
	}
	if err = s.addOrUpdateInviteRelation(ctx, inviteRel); err != nil {
		allInviteLogPara.InviteStatus = 9
		return
	}
	allInviteLogPara.InviteStatus = 1
	err1 := s.invite.SetMidBindInviter(ctx, s.telHash(req.Tel), fiToken.Mid, req.ActivityID)
	if err1 != nil {
		log.Errorc(ctx, "s.inviter.SetMidBindInviter telHas(%s) inviter(%d) activity_uid(%s)", s.telHash(req.Tel), fiToken.Mid, req.ActivityID)
	}
	return
}

func (s *Service) addOrUpdateInviteRelation(ctx context.Context, inviteRel *mdl.InviteRelation) error {
	if err := s.invite.AddInviteRelationLog(ctx, inviteRel); err != nil {
		return err
	}
	err := s.invite.AddInviteRelation(ctx, inviteRel)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			_, err = s.invite.SetInviteRelation(ctx, inviteRel)
			return err
		}
		return err
	}
	return nil
}

func (s *Service) checkIsAlreadyTmpBind(ctx context.Context, tel string) error {
	inviteRel, err := s.invite.GetInviteMidByTel(ctx, tel)
	if err != nil {
		return err
	}
	if inviteRel == nil {
		return nil
	}
	if inviteRel.InvitedMid > 0 {
		return ecode.ActivityInviterTelIsOld
	}
	if inviteRel.IsBlocked == mdl.UserIsBlocked {
		return ecode.ActivityMobileNotAllow
	}
	return nil
}

func (s *Service) userShareLog(ctx context.Context, mid int64, activityUID string) (*mdl.UserShareLog, error) {
	isCache := true
	award, err := s.invite.UserShareLogCache(ctx, mid, activityUID)
	if err != nil {
		isCache = false
	}
	if award != nil {
		return award, nil
	}

	res, err := s.invite.UserShareLog(ctx, mid, activityUID)
	if err != nil {
		return nil, err
	}
	if res == nil {
		// empty cache
		res = &mdl.UserShareLog{}
	}

	if isCache {
		s.fan.Do(ctx, func(ctx context.Context) {
			s.invite.AddUserShareLogCache(ctx, mid, activityUID, res)
		})
	}
	return res, nil
}

func isChinaMobile(mobile string) bool {
	return mobileReg.MatchString(mobile)
}

// isNewTel check tel is new from passport user service.
func (s *Service) isNewTel(ctx context.Context, tel string, countryCode int64) (bool, error) {
	req := &passportinfoapi.TelIsNewReq{
		CountryCode: countryCode,
		Tel:         tel,
	}
	t := tel[0:3] + "****" + tel[len(tel)-4:]
	reply, err := s.passportClient.TelIsNew(ctx, req)
	if err != nil {
		log.Error("PassportUser.TelIsNew tel:%s, country:%d, error:(%v)", t, countryCode, err)
		return false, err
	}
	return reply.IsNew, nil
}
