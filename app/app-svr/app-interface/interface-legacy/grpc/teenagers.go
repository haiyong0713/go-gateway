package grpc

import (
	"context"
	"strconv"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/teenagers"
	"go-gateway/app/app-svr/app-interface/interface-legacy/http"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/usermodel"
	"go-gateway/app/app-svr/app-interface/interface-legacy/service/teenagers"
)

var conut = 0

type TeenagersServer struct {
	srcSvr *teenagers.Service
}

func newTeenagers(ws *warden.Server, svr *http.Server) error {
	s := &TeenagersServer{
		srcSvr: svr.TeenSvr,
	}
	api.RegisterTeenagersServer(ws.Server(), s)
	// 用户鉴权
	auther := mauth.New(nil)
	ws.Add("/bilibili.app.interface.v1.Teenagers/ModifyPwd", auther.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	ws.Add("/bilibili.app.interface.v1.Teenagers/VerifyPwd", auther.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	ws.Add("/bilibili.app.interface.v1.Teenagers/UpdateStatus", auther.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	ws.Add("/bilibili.app.interface.v1.Teenagers/ModeStatus", auther.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	ws.Add("/bilibili.app.interface.v1.Teenagers/FacialRecognitionVerify", auther.UnaryServerInterceptor(true), svr.FeatureSvc.BuildLimitGRPC())
	return nil
}

// 修改密码
func (s *TeenagersServer) ModifyPwd(ctx context.Context, req *api.ModifyPwdReq) (*api.ModifyPwdReply, error) {
	var mid int64
	// 获取鉴权mid
	if au, ok := auth.FromContext(ctx); ok {
		mid = au.Mid
	}
	// 获取设备信息
	dev, ok := device.FromContext(ctx)
	if !ok {
		return nil, ecode.RequestErr
	}
	deviceToken := dev.Buvid
	if req.DeviceToken != "" {
		deviceToken = req.DeviceToken
	}
	if deviceToken == "" {
		return nil, ecode.RequestErr
	}
	if req.OldPwd == req.NewPwd {
		return nil, ecode.Error(ecode.RequestErr, "新老密码不能相同")
	}
	if _, err := strconv.Atoi(req.OldPwd); err != nil || len(req.OldPwd) != 4 {
		return nil, ecode.Error(ecode.RequestErr, "老密码必须是4位长度数字")
	}
	if _, err := strconv.Atoi(req.NewPwd); err != nil || len(req.NewPwd) != 4 {
		return nil, ecode.Error(ecode.RequestErr, "新密码必须是4位长度数字")
	}
	svrReq := &usermodel.ModifyPwdReq{
		Mid:         mid,
		MobiApp:     dev.RawMobiApp,
		DeviceToken: deviceToken,
		DeviceModel: dev.Model,
		OldPwd:      req.OldPwd,
		NewPwd:      req.NewPwd,
	}
	if err := s.srcSvr.ModifyPwd(ctx, svrReq); err != nil {
		log.Error("s.ModifyPwd req:%+v err:%+v", svrReq, err)
		return nil, err
	}
	return &api.ModifyPwdReply{}, nil
}

// 验证密码
func (s *TeenagersServer) VerifyPwd(ctx context.Context, req *api.VerifyPwdReq) (*api.VerifyPwdReply, error) {
	var mid int64
	// 获取鉴权mid
	if au, ok := auth.FromContext(ctx); ok {
		mid = au.Mid
	}
	// 获取设备信息
	dev, ok := device.FromContext(ctx)
	if !ok {
		return nil, ecode.RequestErr
	}
	deviceToken := dev.Buvid
	if req.DeviceToken != "" {
		deviceToken = req.DeviceToken
	}
	if deviceToken == "" {
		return nil, ecode.RequestErr
	}
	// 验证场景
	if _, ok := api.PwdFrom_name[int32(req.PwdFrom)]; !ok {
		return nil, ecode.RequestErr
	}
	// 未知来源
	if req.PwdFrom == api.PwdFrom_UnknownFrom {
		return nil, ecode.RequestErr
	}
	// 动态密码验证，需要登陆状态
	if req.IsDynamic && mid <= 0 {
		return nil, ecode.NoLogin
	}
	svrReq := &usermodel.VerifyPwdReq{
		Mid:         mid,
		MobiApp:     dev.RawMobiApp,
		DeviceToken: deviceToken,
		Pwd:         req.Pwd,
		PwdFrom:     req.PwdFrom,
		IsDynamic:   req.IsDynamic,
		CloseDevice: req.CloseDevice,
	}
	if err := s.srcSvr.VerifyPwd(ctx, svrReq); err != nil {
		log.Error("s.VerifyPwd req:%+v err:%+v", svrReq, err)
		return nil, err
	}
	return &api.VerifyPwdReply{}, nil
}

// 修改青少年模式状态
func (s *TeenagersServer) UpdateStatus(ctx context.Context, req *api.UpdateStatusReq) (*api.UpdateStatusReply, error) {
	var mid int64
	// 获取鉴权mid
	if au, ok := auth.FromContext(ctx); ok {
		mid = au.Mid
	}
	// 获取设备信息
	dev, ok := device.FromContext(ctx)
	if !ok {
		return nil, ecode.RequestErr
	}
	deviceToken := dev.Buvid
	if req.DeviceToken != "" {
		deviceToken = req.DeviceToken
	}
	if deviceToken == "" {
		return nil, ecode.RequestErr
	}
	if _, err := strconv.Atoi(req.Pwd); err != nil || len(req.Pwd) != 4 {
		return nil, ecode.Error(ecode.RequestErr, "密码必须是4位长度数字")
	}
	if !req.Switch {
		// 关闭场景验证
		if req.PwdFrom != api.PwdFrom_FamilyQuitFrom && req.PwdFrom != api.PwdFrom_TeenagersQuitPwdFrom {
			return nil, ecode.RequestErr
		}
	}
	svrReq := &usermodel.UpdateStatusReq{
		Mid:         mid,
		MobiApp:     dev.RawMobiApp,
		DeviceToken: deviceToken,
		DeviceModel: dev.Model,
		Switch:      req.Switch,
		Pwd:         req.Pwd,
		PwdFrom:     req.PwdFrom,
	}
	if err := s.srcSvr.UpdateStatus(ctx, svrReq); err != nil {
		log.Error("s.UpdateStatus req:%+v err:%+v", svrReq, err)
		return nil, err
	}
	return &api.UpdateStatusReply{}, nil
}

// 获取特殊模式状态
func (s *TeenagersServer) ModeStatus(ctx context.Context, req *api.ModeStatusReq) (*api.ModeStatusReply, error) {
	var mid int64
	// 获取鉴权mid
	if au, ok := auth.FromContext(ctx); ok {
		mid = au.Mid
	}
	// 获取网络信息
	network, ok := network.FromContext(ctx)
	if !ok {
		return nil, ecode.RequestErr
	}
	// 获取设备信息
	dev, ok := device.FromContext(ctx)
	if !ok {
		return nil, ecode.RequestErr
	}
	deviceToken := dev.Buvid
	if req.DeviceToken != "" {
		deviceToken = req.DeviceToken
	}
	if deviceToken == "" {
		return nil, ecode.RequestErr
	}
	svrReq := &usermodel.ModeStatusReq{
		Mid:         mid,
		MobiApp:     dev.RawMobiApp,
		DeviceToken: deviceToken,
		DeviceModel: dev.Model,
		IP:          network.RemoteIP,
	}
	userModels, err := s.srcSvr.ModeStatus(ctx, svrReq)
	if err != nil {
		log.Error("s.ModeStatus req:%+v err:%+v", svrReq, err)
		return nil, err
	}
	reply := &api.ModeStatusReply{}
	for _, userModel := range userModels {
		reply.UserModels = append(reply.UserModels, &api.UserModel{
			Mid:  userModel.Mid,
			Mode: userModel.Mode,
			Wsxcde: func() string {
				if userModel.Mode == "teenagers" {
					// 青少年模式不返回密码
					return ""
				}
				return userModel.Wsxcde
			}(),
			Status: api.ModelStatus(userModel.Status),
			Policy: func() *api.Policy {
				if userModel.Policy != nil {
					return &api.Policy{
						Interval:     userModel.Policy.Interval,
						UseLocalTime: userModel.Policy.UseLocalTime,
					}
				}
				return nil
			}(),
			IsForced:        userModel.IsForced,
			MustTeen:        userModel.MustTeen,
			MustRealName:    userModel.MustRealname,
			IsParentControl: userModel.IsParentControl,
		})
	}
	return reply, err
}

// 人脸识别验证
func (s *TeenagersServer) FacialRecognitionVerify(ctx context.Context, req *api.FacialRecognitionVerifyReq) (*api.FacialRecognitionVerifyReply, error) {
	var mid int64
	// 获取鉴权mid
	if au, ok := auth.FromContext(ctx); ok {
		mid = au.Mid
	}
	// 获取设备信息
	dev, ok := device.FromContext(ctx)
	if !ok {
		return nil, ecode.RequestErr
	}
	deviceToken := dev.Buvid
	if req.DeviceToken != "" {
		deviceToken = req.DeviceToken
	}
	if deviceToken == "" {
		return nil, ecode.RequestErr
	}
	// 验证场景
	if _, ok := api.FacialRecognitionVerifyFrom_name[int32(req.From)]; !ok {
		return nil, ecode.RequestErr
	}
	if req.From == api.FacialRecognitionVerifyFrom_VerifyUnknownFrom {
		return nil, ecode.RequestErr
	}
	svrReq := &usermodel.FacialRecognitionVerifyReq{
		Mid:         mid,
		MobiApp:     dev.RawMobiApp,
		DeviceToken: deviceToken,
		From:        req.From,
	}
	if err := s.srcSvr.FacialRecognitionVerify(ctx, svrReq); err != nil {
		log.Error("s.FacialRecognitionVerify req:%+v err:%+v", svrReq, err)
		return nil, err
	}
	return &api.FacialRecognitionVerifyReply{}, nil
}
