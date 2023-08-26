package teenagers

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	membergrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"

	xecode "go-gateway/app/app-svr/app-interface/ecode"
	accountmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/account"
	model "go-gateway/app/app-svr/app-interface/interface-legacy/model/family"
	pushmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/push"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/usermodel"
)

func (s *Service) Aggregation(ctx context.Context, req *model.AggregationReq, mid int64) (*model.AggregationRly, error) {
	rly := &model.AggregationRly{}
	// 青少年模式状态
	if mid > 0 || (req.MobiApp != "" && req.DeviceToken != "") {
		rly.TeenagerStatus, rly.LessonStatus = func() (bool, bool) {
			teen, lesson, err := s.userModels(ctx, mid, req.MobiApp, req.DeviceToken)
			if err != nil {
				return false, false
			}
			return teen != nil && teen.State == int(usermodel.OpenStatus), lesson != nil && lesson.State == int(usermodel.OpenStatus)
		}()
	}
	// 亲子平台状态
	if mid > 0 {
		if fiRly, err := s.Identity(ctx, mid); err == nil && fiRly.Identity != model.IdentityNormal {
			rly.FamilyStatus = true
		}
	}
	return rly, nil
}

func (s *Service) TeenGuard(ctx context.Context, mid int64) (*model.TeenGuardRly, error) {
	rly := &model.TeenGuardRly{}
	rly.Url, rly.RelationType = func() (string, int64) {
		if resp, err := s.Identity(ctx, mid); err == nil && resp != nil {
			switch resp.Identity {
			case model.IdentityChild:
				return "https://www.bilibili.com/h5/teenagers/child/home?navhide=1", model.RelTypeChild
			case model.IdentityParent:
				return "https://www.bilibili.com/h5/teenagers/home?navhide=1", model.RelTypeParent
			}
		}
		return "https://www.bilibili.com/h5/teenagers/home?navhide=1", model.RelTypeNormal
	}()
	return rly, nil
}

func (s *Service) Identity(ctx context.Context, mid int64) (*model.IdentityRly, error) {
	if mid == 0 {
		return &model.IdentityRly{Identity: model.IdentityNormal}, nil
	}
	eg := errgroup.WithCancel(ctx)
	var parentRels []*model.FamilyRelation
	eg.Go(func(ctx context.Context) error {
		rly, err := s.userModelDao.FamilyRelsOfParent(ctx, mid)
		if err != nil {
			return err
		}
		parentRels = rly
		return nil
	})
	var childRel *model.FamilyRelation
	eg.Go(func(ctx context.Context) error {
		rly, err := s.userModelDao.FamilyRelsOfChild(ctx, mid)
		if err != nil {
			return err
		}
		childRel = rly
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Fail to fetch Identity material, mid=%+v error=%+v", mid, err)
		return nil, errors.WithMessage(ecode.ServerErr, "数据查询失败")
	}
	return &model.IdentityRly{
		Identity: identity(parentRels, childRel),
	}, nil
}

func (s *Service) CreateQrcode(ctx context.Context, mid int64) (*model.CreateFamilyQrcodeRly, error) {
	ticket := generateTicket()
	if err := s.familyDao.AddCacheQrcode(ctx, ticket, mid); err != nil {
		return nil, errors.WithMessage(ecode.ServerErr, "缓存ticket失败")
	}
	return &model.CreateFamilyQrcodeRly{
		Ticket: ticket,
		Url:    fmt.Sprintf("https://www.bilibili.com/h5/teenagers/child/accept?navhide=1&ticket=%+v", ticket),
	}, nil
}

func (s *Service) QrcodeInfo(ctx context.Context, req *model.QrcodeInfoReq, cmid int64) (*model.QrcodeInfoRly, error) {
	pmid, err := s.familyDao.CacheQrcode(ctx, req.Ticket)
	if err != nil {
		return nil, errors.WithMessage(ecode.ServerErr, "获取ticket信息失败")
	}
	if pmid <= 0 {
		return nil, ecode.NothingFound
	}
	accounts, err := s.accountDao.Infos3(ctx, []int64{pmid})
	if err != nil {
		return nil, errors.WithMessage(ecode.ServerErr, "获取用户信息失败")
	}
	account, ok := accounts[pmid]
	if !ok {
		return nil, errors.WithMessage(ecode.ServerErr, "用户信息为空")
	}
	isBinded, err := s.verifyChildBind(ctx, pmid, cmid)
	if err != nil {
		return nil, err
	}
	return &model.QrcodeInfoRly{
		Mid:      account.GetMid(),
		Name:     account.GetName(),
		Face:     account.GetFace(),
		IsBinded: isBinded,
	}, nil
}

func (s *Service) QrcodeStatus(ctx context.Context, req *model.QrcodeStatusReq) (*model.QrcodeStatusRly, error) {
	cmid, err := s.familyDao.CacheQrcodeBind(ctx, req.Ticket)
	if err != nil {
		return nil, errors.WithMessage(ecode.ServerErr, "获取绑定信息失败")
	}
	rly := &model.QrcodeStatusRly{}
	if cmid > 0 {
		rly.IsBinded = true
		rly.TeenagerStatus = func() bool {
			teen, _, err := s.userModels(ctx, cmid, "", "")
			if err != nil {
				return false
			}
			return teen != nil && teen.State == int(usermodel.OpenStatus)
		}()
	}
	return rly, nil
}

func (s *Service) ParentIndex(ctx context.Context, pmid int64) (*model.ParentIndexRly, error) {
	rels, err := s.userModelDao.FamilyRelsOfParent(ctx, pmid)
	if err != nil {
		return nil, errors.WithMessage(ecode.ServerErr, "数据查询失败")
	}
	cmids := extractChildMidsFromRelations(rels)
	accounts, teenUsers := s.fetchMaterialOfParentIndex(ctx, cmids)
	infos := make([]*model.ChildInfo, 0, len(cmids))
	for _, rel := range rels {
		if rel == nil || rel.ChildMid <= 0 {
			continue
		}
		info := &model.ChildInfo{
			Mid:            rel.ChildMid,
			TimelockStatus: rel.TimelockState == model.TlStateOpen,
		}
		if acc, ok := accounts[rel.ChildMid]; ok && acc != nil {
			info.Name = acc.GetName()
			info.Face = acc.GetFace()
		}
		if teen, ok := teenUsers[rel.ChildMid]; ok && teen != nil && teen.State == int(usermodel.OpenStatus) {
			info.TeenagerStatus = true
		}
		infos = append(infos, info)
	}
	return &model.ParentIndexRly{
		MaxBind:    model.MaxBind,
		ChildInfos: infos,
	}, nil
}

func (s *Service) fetchMaterialOfParentIndex(ctx context.Context, mids []int64) (map[int64]*accountgrpc.Info, map[int64]*usermodel.User) {
	if len(mids) == 0 {
		return nil, nil
	}
	eg := errgroup.WithContext(ctx)
	var accounts map[int64]*accountgrpc.Info
	eg.Go(func(ctx context.Context) error {
		if rly, err := s.accountDao.Infos3(ctx, mids); err == nil {
			accounts = rly
		}
		return nil
	})
	teenUsers := make(map[int64]*usermodel.User, len(mids))
	lock := sync.Mutex{}
	for _, v := range mids {
		cmid := v
		eg.Go(func(ctx context.Context) error {
			if teen, _, err := s.userModels(ctx, cmid, "", ""); err == nil && teen != nil {
				lock.Lock()
				defer lock.Unlock()
				teenUsers[cmid] = teen
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("Fail to fetch ParentIndex material, mids=%+v error=%+v", mids, err)
		return nil, nil
	}
	return accounts, teenUsers
}

func (s *Service) ParentUnbind(ctx context.Context, req *model.ParentUnbindReq, pmid int64) error {
	rel, err := s.userModelDao.FamilyRelsOfChild(ctx, req.ChildMid)
	if err != nil {
		return errors.WithMessage(ecode.ServerErr, "数据查询失败")
	}
	if rel == nil || rel.ParentMid != pmid {
		return nil
	}
	if err := s.userModelDao.UnbindFamily(ctx, rel); err != nil {
		return errors.WithMessage(ecode.ServerErr, "解绑失败")
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		_ = s.userModelDao.AddFamilyLogs(ctx, []*model.FamilyLog{
			{Mid: rel.ParentMid, Operator: model.OperatorUser, Content: fmt.Sprintf("解除与孩子%+v的绑定", rel.ChildMid)},
			{Mid: rel.ChildMid, Operator: model.OperatorUser, Content: fmt.Sprintf("家长%+v解除绑定", rel.ParentMid)},
		})
	})
	user := &usermodel.User{
		Mid:          rel.ChildMid,
		Password:     "",
		State:        int(usermodel.CloseStatus),
		Model:        usermodel.TeenagersModel,
		Operation:    usermodel.OperationQuitFyParentUnbind,
		QuitTime:     xtime.Time(time.Now().Unix()),
		PwdType:      usermodel.PwdTypeRandom,
		DevOperation: usermodel.OperationQuitFyParentUnbind,
	}
	_ = s.addUserModelAndLog(ctx, user, false)
	return nil
}

func (s *Service) ParentUpdateTeenager(ctx context.Context, req *model.ParentUpdateTeenagerReq, pmid int64) error {
	var user *usermodel.User
	switch req.Action {
	case model.ParentActionOpen:
		user = &usermodel.User{
			Mid:          req.ChildMid,
			Password:     encrypt(randomPwd(usermodel.PwdDigit)),
			State:        int(usermodel.OpenStatus),
			Model:        usermodel.TeenagersModel,
			Operation:    usermodel.OperationOpenFyParentControl,
			QuitTime:     0,
			PwdType:      usermodel.PwdTypeRandom,
			DevOperation: usermodel.OperationOpenFyParentControl,
		}
	case model.ParentActionClose:
		user = &usermodel.User{
			Mid:          req.ChildMid,
			Password:     "",
			State:        int(usermodel.CloseStatus),
			Model:        usermodel.TeenagersModel,
			Operation:    usermodel.OperationQuitFyParentControl,
			QuitTime:     xtime.Time(time.Now().Unix()),
			PwdType:      usermodel.PwdTypeRandom,
			DevOperation: usermodel.OperationQuitFyParentControl,
		}
	default:
		return errors.WithMessagef(ecode.RequestErr, "未知的action=%+v", req.Action)
	}
	rel, err := s.userModelDao.FamilyRelsOfChild(ctx, req.ChildMid)
	if err != nil {
		return errors.WithMessage(ecode.ServerErr, "数据查询失败")
	}
	if rel == nil || rel.ParentMid != pmid {
		return nil
	}
	if err := s.addUserModelAndLog(ctx, user, false); err != nil {
		return errors.WithMessage(ecode.ServerErr, "数据更新失败")
	}
	return nil
}

func (s *Service) ChildIndex(ctx context.Context, cmid int64) (*model.ChildIndexRly, error) {
	rel, err := s.userModelDao.FamilyRelsOfChild(ctx, cmid)
	if err != nil {
		return nil, errors.WithMessage(ecode.ServerErr, "数据查询失败")
	}
	if rel == nil || rel.ParentMid <= 0 {
		return &model.ChildIndexRly{}, nil
	}
	account, teen, _ := s.fetchMaterialOfChildIndex(ctx, rel.ParentMid, rel.ChildMid)
	return &model.ChildIndexRly{
		ParentName:     account.GetName(),
		ParentMid:      account.GetMid(),
		ParentFace:     account.GetFace(),
		TeenagerStatus: teen != nil && teen.State == int(usermodel.OpenStatus),
		TimelockStatus: rel.TimelockState == model.TlStateOpen,
	}, nil
}

func (s *Service) fetchMaterialOfChildIndex(ctx context.Context, pmid, cmid int64) (*accountgrpc.Info, *usermodel.User, error) {
	eg := errgroup.WithContext(ctx)
	var account *accountgrpc.Info
	eg.Go(func(ctx context.Context) error {
		if accounts, err := s.accountDao.Infos3(ctx, []int64{pmid}); err == nil {
			account = accounts[pmid]
		}
		return nil
	})
	var teen *usermodel.User
	eg.Go(func(ctx context.Context) error {
		if teenModel, _, err := s.userModels(ctx, cmid, "", ""); err == nil {
			teen = teenModel
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Fail to fetch ChildIndex material, pmid=%+v cmid=%+v error=%+v", pmid, cmid, err)
		return nil, nil, err
	}
	return account, teen, nil
}

func (s *Service) ChildBind(ctx context.Context, req *model.ChildBindReq, cmid int64) error {
	// 校验ticket
	pmid, err := s.familyDao.CacheQrcode(ctx, req.Ticket)
	if err != nil {
		return errors.WithMessage(ecode.ServerErr, "获取二维码信息失败")
	}
	if pmid == 0 {
		return xecode.FamilyInvalidQrcode
	}
	if pmid == cmid {
		return xecode.FamilyNotSupportBind
	}
	binded, err := s.verifyChildBind(ctx, pmid, cmid)
	if err != nil {
		return err
	}
	if binded {
		return nil
	}
	duration := s.dailyDurationOfBind(ctx, pmid, cmid)
	// 并发控制
	if err := s.familyDao.Lock(ctx, req.Ticket); err != nil {
		return xecode.FamilyLockExceed
	}
	defer func() {
		_ = s.familyDao.Unlock(ctx, req.Ticket)
	}()
	// 绑定
	if err := s.userModelDao.BindFamily(ctx, pmid, cmid, duration); err != nil {
		return errors.WithMessage(ecode.ServerErr, "绑定失败")
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		_ = s.userModelDao.AddFamilyLogs(ctx, []*model.FamilyLog{
			{Mid: pmid, Operator: model.OperatorUser, Content: fmt.Sprintf("成为%+v的家长", cmid)},
			{Mid: cmid, Operator: model.OperatorUser, Content: fmt.Sprintf("成为%+v的孩子", pmid)},
		})
		_ = s.familyDao.DelCacheQrcode(ctx, req.Ticket)
		_ = s.familyDao.AddCacheQrcodeBind(ctx, req.Ticket, cmid)
		// 孩子处于人工强拉则主动关闭|| 孩子已主动开启青少年模式则更新operator
		_ = s.bindCloseManualForce(ctx, cmid)
	})
	return nil
}

func (s *Service) dailyDurationOfBind(ctx context.Context, pmid, cmid int64) int64 {
	if rel, err := s.userModelDao.LatestFamilyRel(ctx, pmid, cmid); err == nil && rel != nil {
		return rel.DailyDuration
	}
	return model.DefaultDailyDuration
}

// 孩子处于人工强拉则主动关闭 || 孩子已主动开启青少年模式则更新operator
func (s *Service) bindCloseManualForce(ctx context.Context, cmid int64) error {
	teen, _, err := s.userModels(ctx, cmid, "", "")
	if err != nil {
		return err
	}
	// 孩子处于人工强拉则主动关闭
	func() {
		if teen == nil || teen.ManualForce == usermodel.ManualForceQuit {
			return
		}
		s.quitManualForce(ctx, teen, "系统解除-亲子平台绑定")
	}()
	//孩子已主动开启青少年模式则更新operator
	func() {
		if teen == nil || !usermodel.IsParentReopen(teen.Operation) {
			return
		}
		s.reOpenParentControl(ctx, teen)
	}()
	return nil
}

func (s *Service) verifyChildBind(ctx context.Context, pmid, cmid int64) (bool, error) {
	eg := errgroup.WithCancel(ctx)
	// 每个家长最多绑定3个孩子
	eg.Go(func(ctx context.Context) error {
		rels, err := s.userModelDao.FamilyRelsOfParent(ctx, pmid)
		if err != nil {
			return errors.WithMessage(ecode.ServerErr, "获取家长的绑定数据失败")
		}
		if len(rels) >= model.MaxBind {
			return xecode.FamilyExceedLimit
		}
		return nil
	})
	// 家长身份不能是孩子
	eg.Go(func(ctx context.Context) error {
		rel, err := s.userModelDao.FamilyRelsOfChild(ctx, pmid)
		if err != nil {
			return errors.WithMessage(ecode.ServerErr, "获取家长的孩子身份失败")
		}
		if rel != nil {
			return xecode.FamilyNotSupportBind
		}
		return nil
	})
	// 家长必须18+
	eg.Go(func(ctx context.Context) error {
		rly, err := s.accountDao.RealnameTeenAgeCheck(ctx, pmid, metadata.String(ctx, metadata.RemoteIP))
		if err != nil {
			return errors.WithMessage(ecode.RequestErr, "获取家长的认证数据失败")
		}
		if rly == nil || rly.Realname == accountmdl.RealnameNotVerified || !rly.After18 {
			return xecode.FamilyNotSupportBind
		}
		return nil
	})
	// 每个孩子最多绑定1个家长
	var binded bool
	eg.Go(func(ctx context.Context) error {
		childRel, err := s.userModelDao.FamilyRelsOfChild(ctx, cmid)
		if err != nil {
			return errors.WithMessage(ecode.RequestErr, "获取孩子的绑定数据失败")
		}
		if childRel != nil {
			if childRel.ParentMid != pmid {
				return xecode.FamilyNotSupportBind
			}
			binded = true
		}
		return nil
	})
	// 孩子身份不能是家长
	eg.Go(func(ctx context.Context) error {
		rel, err := s.userModelDao.FamilyRelsOfParent(ctx, cmid)
		if err != nil {
			return errors.WithMessage(ecode.ServerErr, "获取孩子的家长身份失败")
		}
		if len(rel) > 0 {
			return xecode.FamilyNotSupportBind
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Fail to fetch ChildBind material, cmid=%+v error=%+v", cmid, err)
		return false, err
	}
	return binded, nil
}

func (s *Service) ChildUnbind(ctx context.Context, cmid int64) error {
	rel, ageCheck, err := s.fetchChildUnbindMaterial(ctx, cmid)
	if err != nil {
		return err
	}
	if rel == nil {
		return nil
	}
	if ageCheck == nil || ageCheck.Realname == accountmdl.RealnameNotVerified {
		return xecode.FamilyNotRealnamed
	}
	if !ageCheck.After18 {
		return ecode.Error(ecode.RequestErr, "非18+用户")
	}
	// 解绑
	if err := s.userModelDao.UnbindFamily(ctx, rel); err != nil {
		return errors.WithMessage(ecode.ServerErr, "解绑失败")
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		_ = s.userModelDao.AddFamilyLogs(ctx, []*model.FamilyLog{
			{Mid: rel.ParentMid, Operator: model.OperatorUser, Content: fmt.Sprintf("孩子%+v申诉解除绑定", cmid)},
			{Mid: cmid, Operator: model.OperatorUser, Content: fmt.Sprintf("申诉解除与家长%+v的绑定", rel.ParentMid)},
		})
	})
	// 关闭青少年模式
	user := &usermodel.User{
		Mid:          cmid,
		Password:     "",
		State:        int(usermodel.CloseStatus),
		Model:        usermodel.TeenagersModel,
		Operation:    usermodel.OperationQuitFyChildUnbind,
		QuitTime:     xtime.Time(time.Now().Unix()),
		PwdType:      usermodel.PwdTypeRandom,
		DevOperation: usermodel.OperationQuitFyChildUnbind,
	}
	_ = s.addUserModelAndLog(ctx, user, false)
	return nil
}

func (s *Service) fetchChildUnbindMaterial(ctx context.Context, cmid int64) (*model.FamilyRelation, *membergrpc.RealnameTeenAgeCheckReply, error) {
	eg := errgroup.WithCancel(ctx)
	var rel *model.FamilyRelation
	eg.Go(func(ctx context.Context) error {
		rly, err := s.userModelDao.FamilyRelsOfChild(ctx, cmid)
		if err != nil {
			return errors.WithMessage(ecode.ServerErr, "获取孩子绑定信息失败")
		}
		rel = rly
		return nil
	})
	var ageCheck *membergrpc.RealnameTeenAgeCheckReply
	eg.Go(func(ctx context.Context) error {
		rly, err := s.accountDao.RealnameTeenAgeCheck(ctx, cmid, metadata.String(ctx, metadata.RemoteIP))
		if err != nil {
			return errors.WithMessage(ecode.ServerErr, "获取实名信息失败")
		}
		ageCheck = rly
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Fail to fetch ChildUnbind material, cmid=%+v error=%+v", cmid, err)
		return nil, nil, err
	}
	return rel, ageCheck, nil
}

func (s *Service) TimelockInfo(ctx context.Context, req *model.TimelockInfoReq, mid int64) (*model.TimelockInfoRly, error) {
	rel, err := s.userModelDao.FamilyRelsOfChild(ctx, req.ChildMid)
	if err != nil {
		return nil, errors.WithMessage(ecode.ServerErr, "数据查询失败")
	}
	if rel == nil || rel.ParentMid != mid {
		return nil, ecode.Error(ecode.RequestErr, "不存在亲子关系")
	}
	return &model.TimelockInfoRly{
		TimelockStatus: rel.TimelockState == model.TlStateOpen,
		DailyDuration:  rel.DailyDuration,
	}, nil
}

func (s *Service) UpdateTimelock(ctx context.Context, req *model.UpdateTimelockReq, mid int64) error {
	if !(req.Status == model.TlStateOpen || req.Status == model.TlStateClose) {
		return errors.WithMessagef(ecode.RequestErr, "未知的status=%+v", req.Status)
	}
	rel, err := s.userModelDao.FamilyRelsOfChild(ctx, req.ChildMid)
	if err != nil {
		return errors.WithMessage(ecode.ServerErr, "数据查询失败")
	}
	if rel == nil || rel.ParentMid != mid {
		return ecode.Error(ecode.RequestErr, "不存在亲子关系")
	}
	if rel.TimelockState == req.Status && rel.DailyDuration == req.DailyDuration {
		return nil
	}
	rel.TimelockState = req.Status
	rel.DailyDuration = req.DailyDuration
	if err := s.userModelDao.UpdateTimelock(ctx, rel); err != nil {
		return errors.WithMessage(ecode.ServerErr, "更新时间锁失败")
	}
	return nil
}

func (s *Service) TimelockPwd(ctx context.Context, req *model.TimelockPwdReq, mid int64) (*model.TimelockPwdRly, error) {
	rel, err := s.userModelDao.FamilyRelsOfChild(ctx, req.ChildMid)
	if err != nil {
		return nil, errors.WithMessage(ecode.ServerErr, "数据查询失败")
	}
	if rel == nil || rel.ParentMid != mid {
		return nil, errors.WithMessage(ecode.RequestErr, "没有绑定亲子关系")
	}
	pwd := randomPwd(usermodel.PwdDigit)
	if err := s.familyDao.AddCacheTimelockPwd(ctx, req.ChildMid, pwd); err != nil {
		return nil, errors.WithMessage(ecode.ServerErr, "缓存动态密码失败")
	}
	return &model.TimelockPwdRly{Pwd: pwd}, nil
}

func (s *Service) VerifyTimelockPwd(ctx context.Context, req *model.VerifyTimelockPwdReq, mid int64) (*model.VerifyTimelockPwdRly, error) {
	realPwd, err := s.familyDao.CacheTimelockPwd(ctx, mid)
	if err != nil {
		return nil, errors.WithMessage(ecode.ServerErr, "获取密码缓存失败")
	}
	isPassed := realPwd == req.Pwd
	if isPassed {
		_ = s.familyDao.DelCacheTimelockPwd(ctx, mid)
		return &model.VerifyTimelockPwdRly{IsPassed: isPassed}, nil
	}
	// 验证青少年模式密码
	users, err := s.userModelDao.UserModels(ctx, mid, "", "")
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	for _, user := range users {
		if user.Model == usermodel.TeenagersModel {
			if _, err := strconv.Atoi(req.Pwd); err != nil || len(req.Pwd) != 4 {
				return nil, ecode.Error(ecode.RequestErr, "开启时密码必须是4位长度数字")
			}
			isPassed = user.Password == encrypt(req.Pwd)
			break
		}
	}
	return &model.VerifyTimelockPwdRly{IsPassed: isPassed}, nil
}

func (s *Service) timelock(ctx context.Context, mid int64) *model.Timelock {
	rel, err := s.userModelDao.FamilyRelsOfChild(ctx, mid)
	if err != nil || rel == nil {
		return nil
	}
	return &model.Timelock{
		Switch:        rel.TimelockState == model.TlStateOpen,
		DailyDuration: rel.DailyDuration,
		PushTime:      model.TLPushTime,
		Push: &pushmdl.Message{
			Title:     "o(￣▽￣)d 准备休息倒计时...",
			Summary:   fmt.Sprintf("今日使用时间还剩%d分钟，时间锁预备弹出~", model.TLPushTime),
			Position:  1,                                  //顶部
			Duration:  3,                                  //3s
			Expire:    time.Now().AddDate(1, 0, 0).Unix(), //1年后
			MsgSource: pushmdl.MsgSourceTimelock,
			HideArrow: true,
		},
	}
}

func generateTicket() string {
	if ticket, err := gonanoid.Generate(model.TicketAlphabet, model.TicketLength); err == nil {
		return ticket
	}
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d%d", time.Now().Unix(), rand.Intn(1000))
}

func identity(parentRels []*model.FamilyRelation, childRel *model.FamilyRelation) string {
	if len(parentRels) > 0 && childRel == nil {
		return model.IdentityParent
	}
	if len(parentRels) == 0 && childRel != nil {
		return model.IdentityChild
	}
	if len(parentRels) > 0 && childRel != nil {
		log.Error(fmt.Sprintf("日志告警 亲子关系异常, mid=%+v", childRel.ChildMid))
	}
	return model.IdentityNormal
}

func extractChildMidsFromRelations(rels []*model.FamilyRelation) []int64 {
	cmids := make([]int64, 0, len(rels))
	for _, rel := range rels {
		if rel == nil || rel.ChildMid <= 0 {
			continue
		}
		cmids = append(cmids, rel.ChildMid)
	}
	return cmids
}
