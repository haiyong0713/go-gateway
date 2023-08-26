package lottery

type BluetoothUpListParam struct {
	Bid int64 `form:"bid" validate:"required"`
	Pn  int   `form:"pn" default:"1"`
	Ps  int   `form:"ps" default:"20"`
}

type EditBluetoothUpParam struct {
	ID   int64  `form:"bid" validate:"required"`
	Bid  int64  `form:"bid" validate:"required"`
	Mid  int64  `form:"mid"`
	Key  string `form:"key"`
	Desc string `form:"desc"`
}
