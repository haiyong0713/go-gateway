package trade

import tradegrpc "git.bilibili.co/bapis/bapis-go/vas/trans/trade/service"

type ProductInfoReq struct {
	ProductID   int64 `form:"product_id" validate:"required"`
	ProductType int64 `form:"product_type" validate:"required"`
}

type ProductInfoReply struct {
	ProductDesc      *ProductDesc    `json:"product_desc"`
	UserProtocolList []*UserProtocol `json:"user_protocol_list"`
}

type UserProtocol struct {
	Link  string `json:"link"`
	Title string `json:"title"`
}

type ProductDesc struct {
	ProductId  string `json:"product_id"`
	Cover      string `json:"cover"`
	Title      string `json:"title"`
	Desc       string `json:"desc"`
	PayBtn     string `json:"pay_btn"`
	Price      string `json:"price"`
	NeedCharge string `json:"need_charge"`
}

type OrderStateReq struct {
	OrderID string `form:"order_id" validate:"required"`
}

type OrderStateReply struct {
	OrderState int32 `json:"order_state"`
}

type OrderCreateReq struct {
	ProductID string `form:"product_id" validate:"required"`
	Build     int64  `form:"build" validate:"required"`
	MobiApp   string `form:"mobi_app" validate:"required"`
	From      string `form:"from"`
}

type OrderCreateReply struct {
	TradeOrder *TradeOrder `json:"trade_order"`
}

type TradeOrder struct {
	CustomerID      string `json:"customerId"`
	DeviceType      int64  `json:"deviceType"`
	NotifyURL       string `json:"notifyUrl"`
	OrderCreateTime int64  `json:"orderCreateTime"`
	OrderExpire     int64  `json:"orderExpire"`
	OrderID         string `json:"orderId"`
	OriginalAmount  int64  `json:"originalAmount"`
	PayAmount       int64  `json:"payAmount"`
	ProductID       string `json:"productId"`
	ServiceType     int64  `json:"serviceType"`
	ShowTitle       string `json:"showTitle"`
	Sign            string `json:"sign"`
	SignType        string `json:"signType"`
	Timestamp       int64  `json:"timestamp"`
	TraceID         string `json:"traceId"`
	UID             int64  `json:"uid"`
	Version         string `json:"version"`
}

func (o *TradeOrder) FromTradeCreateReply(i *tradegrpc.TradeCreateReply) {
	if i == nil {
		return
	}
	o.CustomerID = i.CustomerId
	o.DeviceType = i.DeviceType
	o.NotifyURL = i.NotifyUrl
	o.OrderCreateTime = i.OrderCreateTime
	o.OrderExpire = i.OrderExpire
	o.OrderID = i.OrderId
	o.OriginalAmount = i.OriginalAmount
	o.PayAmount = i.PayAmount
	o.ProductID = i.ProductId
	o.ServiceType = i.ServiceType
	o.ShowTitle = i.ShowTitle
	o.Sign = i.Sign
	o.SignType = i.SignType
	o.Timestamp = i.Timestamp
	o.TraceID = i.TraceId
	o.UID = i.Uid
	o.Version = i.Version
}
