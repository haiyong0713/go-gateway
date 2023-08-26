package brand

const (
	// VipBatchNotEnoughErr 资源池数量不足
	VipBatchNotEnoughErr = 69006
)

// CouponReply ...
type CouponReply struct {
	// CouponType ...
	CouponType int `json:"coupon_type"`
}

// FrontEndParams ...
type FrontEndParams struct {
	// Ip ip
	IP string
	// DeviceId ...
	DeviceID string
	// Ua ...
	Ua string
	// API ...
	API string
	// Referer ...
	Referer string
}
