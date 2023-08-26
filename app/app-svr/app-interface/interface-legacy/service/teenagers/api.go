package teenagers

import (
	"context"
	"fmt"
	"time"

	membergrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/teenagers"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/usermodel"

	"github.com/pkg/errors"
)

const _teenagersUnlockErrorNumLimit = 5

// 修改密码
func (s *Service) ModifyPwd(ctx context.Context, req *usermodel.ModifyPwdReq) error {
	err := s.verifyTeenagersOrDynamicPwd(ctx, req.Mid, req.MobiApp, req.DeviceToken, req.OldPwd, int32(api.PwdFrom_TeenagersModifyPwdFrom), false)
	if err != nil {
		return err
	}
	user := &usermodel.User{
		Mid:          req.Mid,
		MobiApp:      req.MobiApp,
		DeviceToken:  req.DeviceToken,
		Password:     encrypt(req.NewPwd),
		State:        1,
		Model:        usermodel.TeenagersModel,
		Operation:    usermodel.OperationOpenSelf,
		PwdType:      usermodel.PwdTypeSelf,
		DevOperation: usermodel.OperationOpenSelf,
	}
	var sync bool
	if req.Mid > 0 {
		sync = true
	}
	if err := s.addUserModelAndLog(ctx, user, sync); err != nil {
		return err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		s.reportedPWDLog(ctx, req.Mid, req.NewPwd, user.Password, req.DeviceToken, req.DeviceModel)
	})
	return nil
}

// 验证密码
func (s *Service) VerifyPwd(ctx context.Context, req *usermodel.VerifyPwdReq) error {
	err := s.verifyTeenagersOrDynamicPwd(ctx, req.Mid, req.MobiApp, req.DeviceToken, req.Pwd, int32(req.PwdFrom), req.IsDynamic)
	if err != nil {
		return err
	}
	// 需要清除设备的青少年模式状态
	if req.CloseDevice {
		// 亲子平台模式退出登陆才需要清除设备的青少年模式状态
		if !req.IsDynamic || req.PwdFrom != api.PwdFrom_FamilyLogOutFrom {
			return ecode.RequestErr
		}
		// 如果需要关闭设备青少年模式
		user := &usermodel.User{
			Mid:          0,
			MobiApp:      req.MobiApp,
			DeviceToken:  req.DeviceToken,
			Password:     "",
			State:        0,
			Model:        usermodel.TeenagersModel,
			QuitTime:     xtime.Time(time.Now().Unix()),
			Operation:    usermodel.OperationQuitSelf,
			PwdType:      usermodel.PwdTypeSelf,
			DevOperation: usermodel.OperationQuitSelf,
		}
		if err := s.addUserModelAndLog(ctx, user, false); err != nil {
			log.Error("日志告警 青少年模式 关闭设备模式失败 req:%+v error:%+v", req, err)
		}
	}
	return nil
}

// 修改青少年模式状态
func (s *Service) UpdateStatus(ctx context.Context, req *usermodel.UpdateStatusReq) error {
	if req.Switch { // 开启青少年模式
		return s.openTeenagersModel(ctx, req.Mid, req.MobiApp, req.DeviceToken, req.DeviceModel, req.Pwd)
	}
	return s.closeTeenagersModel(ctx, req.Mid, req.MobiApp, req.DeviceToken, req.Pwd, req.PwdFrom)
}

// 获取特殊模式状态
func (s *Service) ModeStatus(ctx context.Context, req *usermodel.ModeStatusReq) ([]*usermodel.UserModel, error) {
	if req.Mid == 0 && req.DeviceToken == "" {
		return []*usermodel.UserModel{
			{Mid: req.Mid, Mode: "teenagers", Status: usermodel.NotSetStatus, Policy: &usermodel.Policy{Interval: 0}},
		}, nil
	}
	var (
		users    []*usermodel.User
		zoneIDs  []int64
		ageCheck *membergrpc.RealnameTeenAgeCheckReply
	)
	g := errgroup.WithCancel(ctx)
	g.Go(func(ctx context.Context) error {
		var err error
		users, err = s.userModelDao.UserModels(ctx, req.Mid, req.MobiApp, req.DeviceToken)
		return err
	})
	g.Go(func(ctx context.Context) error {
		reply, err := s.loc.Info2(ctx, req.IP)
		if err != nil {
			log.Error("s.ModeStatus err:%+v", err)
			return nil
		}
		zoneIDs = reply.GetZoneId()
		return nil
	})
	if req.Mid > 0 {
		g.Go(func(ctx context.Context) error {
			if rly, err := s.accountDao.RealnameTeenAgeCheck(ctx, req.Mid, req.IP); err == nil {
				ageCheck = rly
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("s.ModeStatus err:%+v", err)
		return nil, err
	}
	originTeen, lesson := extractModels(users)
	sync, teen := s.autoHandleTeen(ctx, req.Mid, req.MobiApp, req.DeviceToken, originTeen, ageCheck)
	suppleManualForce(originTeen, teen)
	var isForced bool
	if teen != nil && usermodel.IsForced(teen.Operation) {
		isForced = true
	}
	teen = s.autoQuitManualForce(ctx, teen, ageCheck)
	mustTeen := isMustTeen(teen, ageCheck)
	mustRealname := isMustRealname(teen, ageCheck)
	var parentControl bool
	if teen != nil && usermodel.IsParentControl(teen.Operation) {
		parentControl = true
	}
	if teen != nil && req.Mid > 0 && !sync {
		// 尝试将账号信息同步到设备中
		_ = s.cache.Do(ctx, func(ctx context.Context) {
			if err := s.syncDeviceTeenagersModel(ctx, req.MobiApp, req.DeviceToken, req.DeviceModel, parentControl, teen); err != nil {
				log.Error("日志告警 青少年模式 同步设备数据失败 s.syncDeviceTeenagersModel req:%+v, error:%+v", req, err)
			}
		})
	}
	var res []*usermodel.UserModel
	if teen == nil {
		res = append(res, &usermodel.UserModel{
			Mid:             req.Mid,
			Mode:            "teenagers",
			Status:          usermodel.NotSetStatus,
			Policy:          s.policy(req.Mid, nil, zoneIDs, req.MobiApp),
			IsForced:        isForced,
			MustTeen:        mustTeen,
			MustRealname:    mustRealname,
			IsParentControl: parentControl,
		})
	} else {
		res = append(res, &usermodel.UserModel{
			Mid:             req.Mid,
			Mode:            "teenagers",
			Status:          usermodel.Status(teen.State),
			Policy:          s.policy(req.Mid, teen, zoneIDs, req.MobiApp),
			IsForced:        isForced,
			MustTeen:        mustTeen,
			MustRealname:    mustRealname,
			IsParentControl: parentControl,
		})
	}
	if lesson == nil {
		res = append(res, &usermodel.UserModel{Mid: req.Mid, Mode: "lessons", Status: usermodel.CloseStatus})
	} else {
		res = append(res, &usermodel.UserModel{Mid: req.Mid, Mode: "lessons", Status: usermodel.Status(lesson.State), Wsxcde: lesson.Password})
	}
	return res, nil
}

// 人脸识别验证
func (s *Service) FacialRecognitionVerify(ctx context.Context, req *usermodel.FacialRecognitionVerifyReq) error {
	var operation int64
	switch req.From {
	case api.FacialRecognitionVerifyFrom_VerifyFromGuardian:
		operation = usermodel.OperationQuitGuardian
	case api.FacialRecognitionVerifyFrom_VerifyFromAppeal:
		operation = usermodel.OperationQuitAppeal
	}
	user := &usermodel.User{
		Mid:          req.Mid,
		MobiApp:      req.MobiApp,
		DeviceToken:  req.DeviceToken,
		Password:     "",
		State:        0,
		Model:        usermodel.TeenagersModel,
		QuitTime:     xtime.Time(time.Now().Unix()),
		Operation:    int(operation),
		PwdType:      usermodel.PwdTypeSelf,
		DevOperation: int(operation),
	}
	var sync bool
	if req.Mid > 0 {
		sync = true
	}
	if err := s.addUserModelAndLog(ctx, user, sync); err != nil {
		return err
	}
	if operation == usermodel.OperationQuitGuardian && req.Mid > 0 {
		if teen, _, err := s.userModels(ctx, req.Mid, "", ""); err == nil {
			s.quitManualForce(ctx, teen, "系统解除-监护人授权")
		}
	}
	return nil
}

// 验证青少年模式或者动态密码
func (s *Service) verifyTeenagersOrDynamicPwd(ctx context.Context, mid int64, mobiApp, deviceToken, pwd string, pwdFrom int32, isDynamic bool) error {
	day := int64(time.Now().YearDay())
	// 获取对应场景的次数
	num, err := s.userModelDao.GetCacheTeenagersUnlockErrorNum(ctx, deviceToken, mid, day, pwdFrom)
	if err != nil {
		return err
	}
	if num >= _teenagersUnlockErrorNumLimit {
		return errors.WithStack(ecode.Error(ecode.RequestErr, "今日次数已达上限，请明天再尝试或联系客服"))
	}
	// 动态密码验证
	if isDynamic {
		realPwd, err := s.familyDao.CacheTimelockPwd(ctx, mid)
		if err != nil {
			return ecode.ServerErr
		}
		if realPwd != pwd {
			// 错误增加失败次数
			res, err := s.userModelDao.SetCacheTeenagersUnlockErrorNum(ctx, deviceToken, mid, day, pwdFrom)
			if err != nil {
				return err
			}
			return errorByUnlockErrorNum(res)
		}
		_ = s.familyDao.DelCacheTimelockPwd(ctx, mid)
		return nil
	}
	// 非动态密码的验证,都使用青少年模式密码
	teen, _, err := s.userModels(ctx, mid, mobiApp, deviceToken)
	if err != nil {
		return err
	}
	if teen == nil {
		return ecode.NothingFound
	}
	if teen.State == 0 {
		return errors.WithStack(ecode.Error(ecode.RequestErr, "青少年模式已关闭"))
	}
	if teen.Password != encrypt(pwd) {
		// 错误增加失败次数
		res, err := s.userModelDao.SetCacheTeenagersUnlockErrorNum(ctx, deviceToken, mid, day, pwdFrom)
		if err != nil {
			return err
		}
		return errorByUnlockErrorNum(res)
	}
	return nil
}

// 开启青少年模式
func (s *Service) openTeenagersModel(ctx context.Context, mid int64, mobiApp, deviceToken, deviceModel, pwd string) error {
	teen, _, err := s.userModels(ctx, mid, mobiApp, deviceToken)
	if err != nil {
		return err
	}
	if teen != nil && teen.State == 1 {
		return errors.WithStack(ecode.Error(ecode.RequestErr, "青少年模式不能重复开启"))
	}
	user := &usermodel.User{
		Mid:          mid,
		MobiApp:      mobiApp,
		DeviceToken:  deviceToken,
		Password:     encrypt(pwd),
		State:        1,
		Model:        usermodel.TeenagersModel,
		Operation:    usermodel.OperationOpenSelf,
		PwdType:      usermodel.PwdTypeSelf,
		DevOperation: usermodel.OperationOpenSelf,
	}
	var sync bool
	if mid > 0 {
		sync = true
	}
	if err := s.addUserModelAndLog(ctx, user, sync); err != nil {
		return err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		s.reportedPWDLog(ctx, mid, pwd, user.Password, deviceToken, deviceModel)
	})
	return nil
}

// 关闭青少年模式
func (s *Service) closeTeenagersModel(ctx context.Context, mid int64, mobiApp, deviceToken, pwd string, pwdFrom api.PwdFrom) error {
	if err := s.verifyTeenagersOrDynamicPwd(ctx, mid, mobiApp, deviceToken, pwd, int32(pwdFrom), false); err != nil {
		return err
	}
	user := &usermodel.User{
		Mid:          mid,
		MobiApp:      mobiApp,
		DeviceToken:  deviceToken,
		Password:     "",
		State:        0,
		Model:        usermodel.TeenagersModel,
		QuitTime:     xtime.Time(time.Now().Unix()),
		Operation:    usermodel.OperationQuitSelf,
		PwdType:      usermodel.PwdTypeSelf,
		DevOperation: usermodel.OperationQuitSelf,
	}
	var sync bool
	if mid > 0 {
		sync = true
	}
	if err := s.addUserModelAndLog(ctx, user, sync); err != nil {
		return err
	}
	return nil
}

// 通过剩余次数返回相应的错误信息
func errorByUnlockErrorNum(num int64) error {
	if num >= _teenagersUnlockErrorNumLimit {
		return ecode.Error(ecode.RequestErr, "今日次数已达上限，请明天再尝试或联系客服")
	}
	if num <= 1 {
		return ecode.Error(ecode.RequestErr, "密码错误，请重试")
	}
	return ecode.Error(ecode.RequestErr, fmt.Sprintf("密码错误，今日还剩%d次机会", _teenagersUnlockErrorNumLimit-num))
}

// 同步设备青少年模式
func (s *Service) syncDeviceTeenagersModel(ctx context.Context, mobiApp, deviceToken, deviceModel string, parentControl bool, accTeen *usermodel.User) error {
	devTeen, _, err := s.userModels(ctx, 0, mobiApp, deviceToken)
	if err != nil {
		return err
	}
	if devTeen != nil &&
		devTeen.PwdType == accTeen.PwdType &&
		devTeen.Password == accTeen.Password &&
		devTeen.State == accTeen.State &&
		devTeen.Model == accTeen.Model {
		// 设备信息与账号相同时不需要再同步
		return nil
	}
	devUser := &usermodel.User{
		Mid:         0,
		MobiApp:     mobiApp,
		DeviceToken: deviceToken,
		Password:    accTeen.Password,
		State:       accTeen.State,
		Model:       accTeen.Model,
		Operation:   accTeen.Operation,
		QuitTime:    accTeen.QuitTime,
		PwdType:     accTeen.PwdType,
		DevOperation: func() int {
			// 如果是家长控制的状态,设备同步时改为同步客户端状态
			// 防止设备下发is_parent_control:true
			if parentControl {
				if accTeen.State == 1 {
					return usermodel.OperationOpenDevSync
				}
				return usermodel.OperationQuitDevSync
			}
			return accTeen.Operation
		}(),
	}
	if err := s.addUserModelAndLog(ctx, devUser, false); err != nil {
		return err
	}
	if accTeen.State == 1 {
		s.reportedPWDLog(ctx, 0, "", devUser.Password, devUser.DeviceToken, deviceModel)
	}
	return nil
}
