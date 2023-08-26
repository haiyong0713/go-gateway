package service

import (
	"context"
	"strconv"

	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"
)

func (s *Service) CheckExpressionWithDevice(ctx context.Context, req *model.CheckExpressionReq) (*model.CheckExpressionReply, error) {
	dev := pd.NewCommonDevice(req.MobiApp, req.Device, req.Platform, req.Build)
	rst, err := pd.WithDevice(dev).ParseCondition(req.Expression).Finish()
	if err != nil {
		log.Error("%+v", err)
		return nil, xecode.Errorf(xecode.RequestErr, err.Error())
	}
	res := &model.CheckExpressionReply{}
	res.Result = strconv.FormatBool(rst)
	return res, nil
}

func (s *Service) CheckExpressionWithContext(ctx context.Context, req *model.CheckExpressionReq) (*model.CheckExpressionReply, error) {
	rst, err := pd.WithContext(ctx).ParseCondition(req.Expression).Finish()
	if err != nil {
		log.Error("%+v", err)
		return nil, xecode.Errorf(xecode.RequestErr, err.Error())
	}
	res := &model.CheckExpressionReply{}
	res.Result = strconv.FormatBool(rst)
	return res, nil
}
