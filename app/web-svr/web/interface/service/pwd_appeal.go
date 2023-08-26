package service

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"

	feedadmingrpc "git.bilibili.co/bapis/bapis-go/platform/admin/app-feed"
	"go-common/library/ecode"
	"go-common/library/log"

	faecode "go-gateway/app/app-svr/app-feed/ecode"
	xecode "go-gateway/app/web-svr/web/ecode"
	"go-gateway/app/web-svr/web/interface/model"
)

const (
	_emptyMd5 = "d41d8cd98f00b204e9800998ecf8427e"
)

func (s *Service) AddPwdAppeal(c context.Context, req *model.AddPwdAppealReq, mid int64) error {
	// 老版本已登录：使用mid申诉
	// 老版本未登录：升级app后再申诉，新版本会进入青少年模式并同步密码给服务端
	if req.DeviceToken == "" && mid == 0 {
		return xecode.AppealUpgradeApp
	}
	if err := s.VerifyCaptcha(c, req.Mobile, req.Captcha, s.PwdAppealCaptchaDao); err != nil {
		return err
	}
	if req.Pwd == _emptyMd5 {
		req.Pwd = ""
	}
	createReq := &feedadmingrpc.CreatePwdAppealReq{
		Mid:         mid,
		DeviceToken: req.DeviceToken,
		Mobile:      req.Mobile,
		Mode:        req.Mode,
		UploadKey:   req.UploadKey,
		Pwd:         req.Pwd,
	}
	if _, err := s.dao.CreatePwdAppeal(c, createReq); err != nil {
		if ecode.EqualError(faecode.PwdAppealPendingExist, err) {
			return xecode.AppealExist
		}
		return ecode.ServerErr
	}
	return nil
}

func (s *Service) UploadPwdAppeal(c context.Context, key string, file multipart.File) (*model.UploadPwdAppealRly, error) {
	if key == "" {
		return nil, xecode.AppealUpgradeApp
	}
	photo, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorc(c, "Fail to read PwdAppeal photo, error=%+v", err)
		return nil, ecode.RequestErr
	}
	encryptPhoto, err := model.AesEncrypt(photo, []byte(s.c.PwdAppeal.EncryptKey))
	if err != nil {
		log.Error("Fail to encrypt PwdAppeal photo, error=%+v", err)
		return nil, ecode.ServerErr
	}
	uploadKey := model.GenerateAppealUploadKey(key)
	uploadRst, err := s.PwdAppealBoss.Upload(uploadKey, bytes.NewReader(encryptPhoto))
	if err != nil {
		return nil, ecode.ServerErr
	}
	log.Warnc(c, "UploadPwdAppeal success, key=%s uploadKey=%s uploadRst=%+v", key, uploadKey, uploadRst)
	return &model.UploadPwdAppealRly{UploadKey: uploadKey}, nil
}
