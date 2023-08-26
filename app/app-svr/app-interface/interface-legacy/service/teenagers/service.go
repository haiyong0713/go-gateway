package teenagers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"

	membergrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
	"go-common/library/ecode"
	"go-common/library/exp/ab"
	"go-common/library/log"
	"go-common/library/net/metadata"
	rpt "go-common/library/queue/databus/report"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/sync/pipeline/fanout"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/account"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/bgroup"
	"go-gateway/app/app-svr/app-interface/interface-legacy/dao/family"
	locdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/location"
	teendao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/teenagers"
	usermodeldao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/usermodel"
	accmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/account"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/usermodel"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

const (
	_chineseMainland = 22
)

var (
	_teenDailyDialogAb = ab.Int("teen_window_type", "青少年每日弹窗样式", 0)
)

type Service struct {
	c            *conf.Config
	teenDao      *teendao.Dao
	loc          *locdao.Dao
	userModelDao usermodeldao.Dao
	accountDao   *account.Dao
	cache        *fanout.Fanout
	familyDao    *family.Dao
	bgroupDao    *bgroup.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:            c,
		teenDao:      teendao.New(c),
		loc:          locdao.New(c),
		userModelDao: usermodeldao.New(c),
		accountDao:   account.New(c),
		cache:        fanout.New("teenagers"),
		familyDao:    family.NewDao(c),
		bgroupDao:    bgroup.NewDao(c),
	}
	return
}

func (s *Service) Status(ctx context.Context, mid int64) (*usermodel.ModelStatus, error) {
	users, err := s.userModelDao.UserModels(ctx, mid, "", "")
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	for _, user := range users {
		if user.Model == usermodel.TeenagersModel {
			return &usermodel.ModelStatus{Wsxcde: user.Password, TeenagersStatus: usermodel.Status(user.State)}, nil
		}
	}
	return &usermodel.ModelStatus{TeenagersStatus: usermodel.NotSetStatus}, nil
}

func (s *Service) UserModel(ctx context.Context, mobiApp, deviceToken string, mid int64, ip, buvid string) ([]*usermodel.UserModel, error) {
	var dialogAb int64
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsMobiAppIPhone().And().Build("<", 67800000)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build("<", 6780000)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build("<", 34500000)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroidHD().And().Build("<", 1230000)
	}).MustFinish() {
		dialogAb = 2
	}
	if mid == 0 && deviceToken == "" {
		return []*usermodel.UserModel{
			{Mid: mid, Mode: "teenagers", Status: usermodel.NotSetStatus, Policy: &usermodel.Policy{Interval: 0}, DailyDialogAb: dialogAb},
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
		users, err = s.userModelDao.UserModels(ctx, mid, mobiApp, deviceToken)
		return err
	})
	g.Go(func(ctx context.Context) error {
		reply, err := s.loc.Info2(ctx, ip)
		if err != nil {
			log.Error("%+v", err)
			return nil
		}
		zoneIDs = reply.GetZoneId()
		return nil
	})
	if mid > 0 {
		g.Go(func(ctx context.Context) error {
			if rly, err := s.accountDao.RealnameTeenAgeCheck(ctx, mid, ip); err == nil {
				ageCheck = rly
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	originTeen, lesson := extractModels(users)
	_, teen := s.autoHandleTeen(ctx, mid, mobiApp, deviceToken, originTeen, ageCheck)
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
	var res []*usermodel.UserModel
	if teen == nil {
		res = append(res, &usermodel.UserModel{
			Mid:             mid,
			Mode:            "teenagers",
			Status:          usermodel.NotSetStatus,
			Policy:          s.policy(mid, nil, zoneIDs, mobiApp),
			IsForced:        isForced,
			MustTeen:        mustTeen,
			MustRealname:    mustRealname,
			IsParentControl: parentControl,
			DailyDialogAb:   dialogAb,
		})
	} else {
		res = append(res, &usermodel.UserModel{
			Mid:             mid,
			Mode:            "teenagers",
			Status:          usermodel.Status(teen.State),
			Wsxcde:          teen.Password,
			Policy:          s.policy(mid, teen, zoneIDs, mobiApp),
			IsForced:        isForced,
			MustTeen:        mustTeen,
			MustRealname:    mustRealname,
			IsParentControl: parentControl,
			DailyDialogAb:   dialogAb,
		})
	}
	if lesson == nil {
		res = append(res, &usermodel.UserModel{Mid: mid, Mode: "lessons", Status: usermodel.CloseStatus})
	} else {
		res = append(res, &usermodel.UserModel{Mid: mid, Mode: "lessons", Status: usermodel.Status(lesson.State), Wsxcde: lesson.Password})
	}
	return res, nil
}

func (s *Service) autoHandleTeen(ctx context.Context, mid int64, mobiApp, deviceToken string, teen *usermodel.User, ageCheck *membergrpc.RealnameTeenAgeCheckReply) (bool, *usermodel.User) {
	if mid == 0 {
		return false, teen
	}
	if s.ifForceOpen(ctx, teen, ageCheck, mid) {
		if user, err := s.forceOpen(ctx, mid, mobiApp, deviceToken); err == nil {
			return true, user
		}
	} else if s.ifForceQuit(ctx, teen, ageCheck) {
		if user, err := s.forceQuit(ctx, mid, mobiApp, deviceToken); err == nil {
			return true, user
		}
	}
	return false, teen
}

// 强制拉入
func (s *Service) ifForceOpen(ctx context.Context, user *usermodel.User, ageCheck *membergrpc.RealnameTeenAgeCheckReply, mid int64) bool {
	// 版本控制
	if !feature.GetBuildLimit(ctx, "service.TeenageForce", nil) {
		return false
	}
	// 14-用户
	if ageCheck == nil || ageCheck.Realname != accmdl.RealnameVerified || ageCheck.After14 {
		return false
	}
	if s.c.Teenagers.ForceOnlineTime > ageCheck.Rtime {
		// 存量14- 受服务端开关控制
		if !s.c.Teenagers.ForceOpen {
			return false
		}
	} else if !ageCheck.IsFaceid {
		// 增量14- 判断认证渠道
		return false
	}
	// 已绑定亲子关系
	if rel, err := s.userModelDao.FamilyRelsOfChild(ctx, mid); err == nil && rel != nil {
		return false
	}
	// 未在青少年模式内
	if user != nil {
		if user.State == int(usermodel.OpenStatus) {
			return false
		}
		return time.Now().Unix()-int64(user.QuitTime) > s.c.Teenagers.ForceOpenInterval
	}
	return true
}

func (s *Service) forceOpen(ctx context.Context, mid int64, mobiApp, deviceToken string) (*usermodel.User, error) {
	pwd := randomPwd(usermodel.PwdDigit)
	user, err := s.UserModelUpdate(ctx, mobiApp, deviceToken, mid, pwd, "", 1, true, usermodel.TeenagersModel, usermodel.OperationOpenForce, usermodel.PwdTypeRandom)
	if err != nil {
		log.Error("Fail to handle teenager forceOpen, mid=%d mobiApp=%s deviceToken=%s error=%+v", mid, mobiApp, deviceToken, err)
		return nil, err
	}
	return user, nil
}

// 符合条件强制解除
func (s *Service) ifForceQuit(ctx context.Context, user *usermodel.User, ageCheck *membergrpc.RealnameTeenAgeCheckReply) bool {
	// 服务端开关开启状态
	if !s.c.Teenagers.ForceClose {
		return false
	}
	// 版本控制
	if !feature.GetBuildLimit(ctx, "service.TeenageForce", nil) {
		return false
	}
	// 14+用户
	if ageCheck != nil && ageCheck.Realname == accmdl.RealnameVerified && !ageCheck.After14 {
		return false
	}
	// 强制开启状态
	return user != nil && user.State == int(usermodel.OpenStatus) && user.Operation == usermodel.OperationOpenForce
}

func (s *Service) forceQuit(ctx context.Context, mid int64, mobiApp, deviceToken string) (*usermodel.User, error) {
	user, err := s.UserModelUpdate(ctx, mobiApp, deviceToken, mid, "", "", 0, true, usermodel.TeenagersModel, usermodel.OperationQuitForce, usermodel.PwdTypeSelf)
	if err != nil {
		log.Error("Fail to handle teenager forceQuit, mid=%d mobiApp=%s deviceToken=%s error=%+v", mid, mobiApp, deviceToken, err)
		return nil, err
	}
	return user, nil
}

func (s *Service) UpdateTeenager(ctx context.Context, req *usermodel.UpdateTeenagerReq, mid int64) error {
	var operation int
	switch req.TeenagersStatus {
	case 0:
		if op, ok := usermodel.From2OpOfQuit[req.From]; ok {
			operation = op
		}
	case 1:
		if op, ok := usermodel.From2OpOfOpen[req.From]; ok {
			operation = op
		}
	}
	if _, err := s.UserModelUpdate(ctx, req.MobiApp, req.DeviceToken, mid, req.Pwd, req.Wsxcde, req.TeenagersStatus, req.Sync, usermodel.TeenagersModel, operation, usermodel.PwdTypeSelf); err != nil {
		return err
	}
	if operation == usermodel.OperationQuitGuardian && mid > 0 {
		if teen, _, err := s.userModels(ctx, mid, "", ""); err == nil {
			s.quitManualForce(ctx, teen, "系统解除-监护人授权")
		}
	}
	if req.TeenagersStatus == 1 {
		_ = s.cache.Do(ctx, func(ctx context.Context) {
			s.reportedPWDLog(ctx, mid, req.Pwd, req.Wsxcde, req.DeviceToken, req.DeviceModel)
		})
	}
	return nil
}

func (s *Service) UserModelUpdate(ctx context.Context, mobiApp, deviceToken string, mid int64, pwd, wsxcde string, status int, sync bool, model usermodel.Model, operation, pwdType int) (*usermodel.User, error) {
	var password string
	switch status {
	case 0: // 0:关闭
		if pwd != "" || wsxcde != "" {
			return nil, ecode.Error(ecode.RequestErr, "关闭时密码必须为空")
		}
	case 1: // 1:开启
		if pwd == "" && wsxcde == "" {
			return nil, ecode.Error(ecode.RequestErr, "开启时密码必须要传")
		}
		password = wsxcde
		if pwd != "" {
			if _, err := strconv.Atoi(pwd); err != nil || len(pwd) != 4 {
				return nil, ecode.Error(ecode.RequestErr, "开启时密码必须是4位长度数字")
			}
			password = encrypt(pwd)
		}
	}
	var quitTime xtime.Time
	if status == 0 {
		quitTime = xtime.Time(time.Now().Unix())
	}
	user := &usermodel.User{
		Mid:         mid,
		MobiApp:     mobiApp,
		DeviceToken: deviceToken,
		Password:    password,
		State:       status,
		Model:       model,
		Operation:   operation,
		QuitTime:    quitTime,
		PwdType:     pwdType,
		DevOperation: func() int {
			// 如果是家长控制的状态,设备同步时改为同步客户端状态
			if operation == usermodel.OperationOpenParentForcedSync {
				return usermodel.OperationOpenDevSync
			}
			if operation == usermodel.OperationQuitParentForcedSync {
				return usermodel.OperationQuitDevSync
			}
			return operation
		}(),
	}
	if err := s.addUserModelAndLog(ctx, user, sync); err != nil {
		log.Error("UserModelUpdate mid:%d,user:%+v,error%+v", mid, user, err)
		return nil, err
	}
	return user, nil
}

// nolint:gocognit,gomnd
func (s *Service) policy(mid int64, user *usermodel.User, zoneIDs []int64, mobiApp string) *usermodel.Policy {
	interval := func() int64 {
		if mobiApp == "iphone_i" || mobiApp == "android_i" {
			return 0
		}
		var countryID, provinceID, cityID int64
		for index, zoneID := range zoneIDs { // zone_id[0]目前无意义，zone_id[1] 代表国家维度，zone_id[2]代表省份或者直辖市维度，zone_id[3]代表城市维度
			switch index {
			case 1:
				countryID = zoneID
			case 2:
				provinceID = zoneID
			case 3:
				cityID = zoneID
			default:
				continue
			}
		}
		switch countryID {
		case 0, 947912704, 402653184, 977272832: // 非外网的IP zone_id[1]，共享地址：947912704，局域网：402653184，本机地址：977272832
			return s.c.Teenagers.OuterInterval
		}
		for _, zone := range s.c.Teenagers.NoneZone {
			if zone == nil {
				continue
			}
			var zoneID int64
			switch zone.Index {
			case 1:
				zoneID = countryID
			case 2:
				zoneID = provinceID
			default:
				continue
			}
			if vals, ok := zone.Zone[strconv.FormatInt(zoneID, 10)]; ok {
				if len(vals) == 0 {
					return 0
				}
				if cityID != 0 {
					for _, city := range vals {
						if cityID == city {
							return 0
						}
					}
				}
			}
		}
		if mid == 0 || user != nil { // 1.游客 2.历史成功开启过青少年模式的用户
			return s.c.Teenagers.OuterInterval
		}
		if citys, ok := s.c.Teenagers.OuterZone[strconv.FormatInt(provinceID, 10)]; ok {
			if len(citys) == 0 || cityID == 0 {
				return s.c.Teenagers.OuterInterval
			}
			for _, city := range citys {
				if cityID == city {
					return s.c.Teenagers.OuterInterval
				}
			}
		}
		return s.c.Teenagers.InnerInterval
	}()
	useLocalTime := true // 默认使用本地时间
	if len(zoneIDs) > 1 {
		// 第2个zone id,右移22位数值是1表示为中国大陆地区
		// 如果不是中国大陆地区就使用本地时间
		useLocalTime = (zoneIDs[1] >> _chineseMainland) != 1
	}
	return &usermodel.Policy{
		Interval:     interval,
		UseLocalTime: useLocalTime,
	}
}

func (s *Service) SetAntiAddictionTime(ctx context.Context, param *usermodel.AntiAddiction) error {
	if err := s.userModelDao.SetCacheAntiAddictionTime(ctx, param.DeviceToken, param.MID, param.Day, param.UseTime); err != nil {
		log.Error("service.SetAntiAddictionTime set cache error,err:(%v),param:(%+v)", err, param)
		return err
	}
	return nil
}

func (s *Service) GetAntiAddictionTime(ctx context.Context, param *usermodel.AntiAddiction) (int64, error) {
	if param.MID == 0 {
		mid, err := s.userModelDao.GetCacheAntiAddictionMID(ctx, param.DeviceToken, param.Day)
		if err != nil {
			log.Error("service.GetAntiAddictionTime get cache mid error,err:(%v),param:(%+v)", err, param)
			return 0, err
		}
		param.MID = mid
	}
	res, err := s.userModelDao.GetCacheAntiAddictionTime(ctx, param.DeviceToken, param.MID, param.Day)
	if err != nil {
		log.Error("service.GetAntiAddictionTime get cache error,err:(%v),param:(%+v)", err, param)
		return 0, err
	}
	return res, nil
}

func (s *Service) autoQuitManualForce(ctx context.Context, user *usermodel.User, ageCheck *membergrpc.RealnameTeenAgeCheckReply) *usermodel.User {
	// 18+
	if !(ageCheck != nil && ageCheck.Realname == accmdl.RealnameVerified && ageCheck.After18) {
		return user
	}
	newUser := cloneUser(user)
	s.quitManualForce(ctx, newUser, "系统解除-实名认证")
	return newUser
}

func (s *Service) quitManualForce(ctx context.Context, teen *usermodel.User, content string) {
	// 人工强拉已开启
	if !(teen != nil && teen.ManualForce == usermodel.ManualForceOpen) {
		return
	}
	teen.ManualForce = usermodel.ManualForceQuit
	teen.MfOperator = usermodel.OperatorSystem
	teen.MfTime = xtime.Time(time.Now().Unix())
	if err := s.userModelDao.UpdateManualForceAndCache(ctx, teen); err != nil {
		return
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		_ = s.userModelDao.AddManualForceLog(ctx, &usermodel.ManualForceLog{
			Mid:      teen.Mid,
			Operator: usermodel.OperatorSystem,
			Content:  content,
		})
	})
}

func (s *Service) reOpenParentControl(ctx context.Context, teen *usermodel.User) {
	if teen == nil {
		return
	}
	//更新表&cache
	if err := s.userModelDao.UpdateOperation(ctx, teen.ID, teen.Mid, usermodel.OperationOpenParentControlReopen); err != nil {
		log.Error("s.userModelDao.UpdateOperation(%d)", teen.ID)
		return
	}
	//更新日志
	_ = s.userModelDao.AddSpecialModeLog(ctx, &usermodel.SpecialModeLog{
		RelatedKey:  usermodel.LogRelatedKey(usermodel.RelatedKeyTypeUser, teen.Mid),
		OperatorUid: 0,
		Operator:    usermodel.LogOperator(usermodel.OperationOpenParentControlReopen),
		Content:     usermodel.LogContent(usermodel.OperationOpenParentControlReopen, teen.State),
	})
}

func (s *Service) userModels(ctx context.Context, mid int64, mobiApp string, deviceToken string) (teen *usermodel.User, lesson *usermodel.User, err error) {
	userModels, err := s.userModelDao.UserModels(ctx, mid, mobiApp, deviceToken)
	if err != nil {
		return nil, nil, err
	}
	teen, lesson = extractModels(userModels)
	return teen, lesson, nil
}

func (s *Service) addUserModelAndLog(ctx context.Context, user *usermodel.User, sync bool) error {
	userId, devId, err := s.userModelDao.AddUserModel(ctx, user, sync)
	if err != nil {
		return err
	}
	_ = s.cache.Do(ctx, func(ctx context.Context) {
		var (
			operatorName, operatorAccContent, operatorDevContent string
		)
		operatorName = usermodel.LogOperator(user.Operation)
		if userId > 0 {
			operatorAccContent = usermodel.LogContent(user.Operation, user.State)
		}
		if devId > 0 {
			operatorDevContent = usermodel.LogContent(user.DevOperation, user.State)
		}
		s.reportedOperatorLog(ctx, user.Mid, 0, user.Model, user.DeviceToken, operatorName, operatorAccContent, operatorDevContent)
	})
	return nil
}

func extractModels(users []*usermodel.User) (teen *usermodel.User, lesson *usermodel.User) {
	for _, user := range users {
		switch user.Model {
		case usermodel.TeenagersModel:
			teen = user
		case usermodel.LessonsModel:
			lesson = user
		default:
			continue
		}
	}
	return teen, lesson
}

func randomPwd(digit int) string {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	var pwd string
	for i := 0; i < digit; i++ {
		pwd += strconv.Itoa(rd.Intn(10))
	}
	return pwd
}

func isMustTeen(user *usermodel.User, ageCheck *membergrpc.RealnameTeenAgeCheckReply) bool {
	if user == nil {
		return false
	}
	if user.ManualForce != usermodel.ManualForceOpen {
		return false
	}
	if ageCheck != nil && ageCheck.Realname == accmdl.RealnameVerified && ageCheck.After18 {
		return false
	}
	return true
}

func isMustRealname(user *usermodel.User, ageCheck *membergrpc.RealnameTeenAgeCheckReply) bool {
	if user == nil {
		return false
	}
	if user.ManualForce != usermodel.ManualForceOpen {
		return false
	}
	return ageCheck != nil && ageCheck.Realname != accmdl.RealnameVerified
}

func cloneUser(user *usermodel.User) *usermodel.User {
	if user == nil {
		return nil
	}
	tmp := *user
	return &tmp
}

func suppleManualForce(origin, now *usermodel.User) {
	if origin == nil || now == nil {
		return
	}
	now.ManualForce = origin.ManualForce
	now.MfOperator = origin.MfOperator
	now.MfTime = origin.MfTime
}

func encrypt(pwd string) string {
	h := md5.New()
	_, _ = h.Write([]byte(pwd))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Service) reportedPWDLog(ctx context.Context, mid int64, pwd, wsxcde, deviceToken, deviceModel string) {
	var err error
	if pwd == "" {
		if pwd, err = s.userModelDao.GetTeenagerModelPWD(ctx, wsxcde); err != nil {
			log.Error("s.reportedPWDLog mid:%d, deviceToken:%s, wsxcde:%s, error:%v", mid, deviceToken, wsxcde, err)
			return
		}
	}
	rptFunc := func() error {
		return rpt.User(&rpt.UserInfo{
			Mid:      mid,
			Business: 92,
			Action:   "teenagers_pwd_set",
			Ctime:    time.Now(),
			IP:       metadata.String(ctx, metadata.RemoteIP),
			Buvid:    deviceToken,
			Content: map[string]interface{}{
				"device_model": deviceModel,
				"password":     pwd,
			},
		})
	}
	for i := 0; i < 3; i++ {
		if err := rptFunc(); err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) reportedOperatorLog(ctx context.Context, mid, operatorUid int64, model usermodel.Model, deviceToken, operatorName, operatorAccContent, operatorDevContent string) {
	action := "teenagers_mode_log"
	if model == usermodel.LessonsModel {
		action = "lessons_mode_log"
	}
	rptFunc := func() error {
		return rpt.User(&rpt.UserInfo{
			Mid:      mid,
			Business: 369,
			Action:   action,
			Ctime:    time.Now(),
			IP:       metadata.String(ctx, metadata.RemoteIP),
			Buvid:    deviceToken,
			Content: map[string]interface{}{
				"operator_uid":         operatorUid,
				"operator_name":        operatorName,
				"operator_acc_content": operatorAccContent,
				"operator_dev_content": operatorDevContent,
			},
		})
	}
	for i := 0; i < 3; i++ {
		if err := rptFunc(); err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}
