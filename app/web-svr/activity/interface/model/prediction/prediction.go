package prediction

// PreParams .
type PreParams struct {
	Sid      int64  `form:"sid" validate:"min=1"`
	NickName string `form:"nick_name"`
	Point    int64  `form:"point" default:"0"`
}
