package space

// FollowParam def.
type FollowParam struct {
	MobiApp  string `form:"mobi_app" validate:"required"`
	Device   string `form:"device"`
	Platform string `form:"platform" validate:"required"`
	Build    int    `form:"build" validate:"required"`
	Vmid     int64  `form:"vmid" validate:"required"`
}
