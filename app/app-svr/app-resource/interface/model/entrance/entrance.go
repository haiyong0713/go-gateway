package entrance

type BusinessInfocReq struct {
	Business string `form:"business" validate:"required"`
	Mid      int64  `form:"mid"`
	UpMid    int64  `form:"up_mid"`
}
