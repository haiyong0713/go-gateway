package model

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type RoleContext struct {
	Node     string `form:"node" validate:"required"`
	AppName  string `form:"app_name"`
	Gateway  string `form:"gateway"`
	Cookie   string `form:"-"`
	Username string `form:"-"`
	Role     string `form:"-"`
}

func (rc *RoleContext) TargetGateway() string {
	if rc.AppName != "" {
		return rc.AppName
	}
	return rc.Gateway
}
