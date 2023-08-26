package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	xecode "go-gateway/app/app-svr/kvo/ecode"
	pb "go-gateway/app/app-svr/kvo/interface/api"
	"go-gateway/app/app-svr/kvo/interface/model"

	"go-common/library/log"

	infoc2 "go-common/library/log/infoc.v2"
)

func (s *Service) AddUserDoc(c context.Context, mid int64, req pb.ConfigModify, platform, buvid, module string) (err error) {
	var (
		data     []byte
		newCtx   = context.TODO()
		reqMap   map[string]string
		moduleID int
	)
	if req == nil {
		return
	}
	if module == "" {
		module = pb.DmPlayerConfig
	}
	if moduleID = pb.VerifyModuleKey(module); moduleID == 0 {
		err = xecode.KvoModuleNotExist
		return
	}
	msg := &struct {
		Body     pb.ConfigModify
		Mid      int64
		Platform string
		Buvid    string
	}{
		Body:     req,
		Mid:      mid,
		Platform: platform,
		Buvid:    buvid,
	}
	reqMap = req.ToMap()
	// 增量数据 不存在不落库
	if err = s.da.HMsetUserDoc(c, mid, buvid, moduleID, reqMap); err != nil {
		log.Warn("s.HMsetUserDoc(mid:%d, reqMap:%+v) err(%v)", mid, reqMap, err)
		err = nil
	}
	if data, err = json.Marshal(msg); err != nil {
		log.Error("json.Marshal(%+v) mid(%d) error(%v)", req, mid, err)
		return
	}
	act := &model.Action{Action: module, Data: data}
	switch mid {
	case 0:
		err = s.da.SendBuvidTaskAction(newCtx, buvid, act)
	default:
		err = s.da.SendTaskAction(newCtx, fmt.Sprint(mid), act)
	}
	_ = s.cacheLog.Do(newCtx, func(ctx context.Context) {
		if body, err := json.Marshal(req); err == nil {
			_ = s.sendBIData(ctx, &model.BILogStream{
				Business: module,
				Mid:      mid,
				Buvid:    buvid,
				Body:     string(body),
				Platform: platform,
				CTime:    time.Now().Unix(),
			})
		}
	})
	return
}

func (s *Service) userDoc(c context.Context, mid int64, buvid string, moduleKeyID int) (bs json.RawMessage, err error) {
	if bs, err = s.da.UserDocRds(c, mid, buvid, moduleKeyID); err != nil {
		err = nil
	}
	if bs != nil {
		if len(bs) == 0 {
			bs = nil
		}
		return
	}
	if bs, err = s.userDocTaiShan(c, mid, buvid, moduleKeyID); err != nil {
		log.Error("d.userDocTaiShan(mid:%d, buvid:%s, modulekey:%d) err(%v)", 0, buvid, moduleKeyID, err)
		return
	}
	if bs == nil {
		_ = s.da.SetUserDocRds(c, mid, buvid, moduleKeyID, []byte{})
		return
	}
	_ = s.da.SetUserDocRds(c, mid, buvid, moduleKeyID, bs)
	return
}

func (s *Service) sendBIData(ctx context.Context, l *model.BILogStream) (err error) {
	if s.infocLogStream == nil {
		return
	}
	payload := infoc2.NewLogStreamV(model.BILogStreamID, log.String(l.Business), log.Int64(l.Mid), log.String(l.Buvid),
		log.String(l.Body), log.String(l.Platform), log.Int64(l.CTime))
	if err = s.infocLogStream.Info(ctx, payload); err != nil {
		log.Warn("s.sendBIData() logstream(%+v) error(%v)", l, err)
	}
	return
}
