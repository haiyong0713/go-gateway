package view

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-view/interface/model/trade"
	"go-gateway/app/app-svr/ugc-season/service/api"

	"github.com/pkg/errors"
)

const (
	_seasonType int64 = 1
)

func (s *Service) TradeProductInfo(ctx context.Context, req *trade.ProductInfoReq) (*trade.ProductInfoReply, error) {
	productInfo := &trade.ProductInfoReply{
		UserProtocolList: []*trade.UserProtocol{
			{
				Link:  "https://www.bilibili.com/blackboard/agreement-class-h5.html",
				Title: "《哔哩哔哩付费内容购买协议》",
			},
			{
				Link:  "https://pay.bilibili.com/paywallet-fe/doc/license.html",
				Title: "《B币用户协议》",
			},
		},
	}
	switch req.ProductType {
	case _seasonType:
		// 合集商品
		productDesc, err := s.seasonProductDesc(ctx, req.ProductID)
		if err != nil {
			log.Error("s.TradeProductInfo error:%v", err)
			return nil, err
		}
		productInfo.ProductDesc = productDesc
	default:
		log.Error("日志告警 付费UGC 未知商品类型:%+v", req)
		return nil, errors.Wrapf(ecode.Error(ecode.RequestErr, "服务开小差了，请稍后重试~"), "未知商品类型, req:%+v", req)
	}
	return productInfo, nil
}

func (s *Service) TradeOrderState(ctx context.Context, mid int64, req *trade.OrderStateReq) (*trade.OrderStateReply, error) {
	reply, err := s.tradeDao.TradeOrderStateInfo(ctx, mid, req.OrderID)
	if err != nil {
		log.Error("日志告警 付费UGC s.TradeOrderState mid:%d, req:%+v, error:%v", mid, req, err)
		return nil, err
	}
	return &trade.OrderStateReply{OrderState: int32(reply.GetOrderState())}, nil
}

func (s *Service) TradeOrderCreate(ctx context.Context, mid int64, req *trade.OrderCreateReq) (*trade.OrderCreateReply, error) {
	reply, err := s.tradeDao.TradeOrderCreate(ctx, mid, req.Build, req.ProductID, req.MobiApp, req.From)
	if err != nil {
		log.Error("日志告警 付费UGC s.TradeOrderCreate mid:%d, req:%+v, error:%v", mid, req, err)
		return nil, err
	}
	tradeOrder := &trade.TradeOrder{}
	tradeOrder.FromTradeCreateReply(reply)
	return &trade.OrderCreateReply{TradeOrder: tradeOrder}, nil
}

func (s *Service) seasonProductDesc(ctx context.Context, seasonID int64) (*trade.ProductDesc, error) {
	season, err := s.seasonDao.SeasonInfo(ctx, seasonID)
	if err != nil {
		return nil, err
	}
	if season.Season.AttrVal(api.SeasonAttrSnPay) != api.AttrSnYes {
		return nil, errors.Wrapf(ecode.Error(ecode.RequestErr, "服务开小差了，请稍后重试~"), "该合集不属于付费合集, ID:%d, attr:%d", seasonID, season.Season.Attribute)
	}
	productDesc := &trade.ProductDesc{
		Cover:      season.Season.Cover,
		Title:      season.Season.Title,
		ProductId:  season.Season.GoodsInfo.GoodsId,
		Price:      season.Season.GoodsInfo.GoodsPriceFmt,
		NeedCharge: season.Season.GoodsInfo.GoodsPriceFmt,
		PayBtn:     "购买",
		Desc: func() string {
			if season.Season.EpCount > season.Season.EpNum {
				return fmt.Sprintf("共%d个视频", season.Season.EpCount)
			}
			return fmt.Sprintf("共%d个视频", season.Season.EpNum)
		}(),
	}
	return productDesc, nil
}
