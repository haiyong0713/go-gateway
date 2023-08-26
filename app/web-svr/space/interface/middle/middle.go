package middle

import (
	"context"

	acpAPI "git.bilibili.co/bapis/bapis-go/account/service/account_control_plane"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/ecode"
	"go-gateway/app/web-svr/space/interface/conf"
)

const (
	_allowAction   = "space"
	_allowAction64 = "2020_64"
)

type Middle struct {
	acpClient acpAPI.AccountControlPlaneClient
}

// New new  Middle service.
func New(c *conf.Config) *Middle {
	acpRPC, err := acpAPI.NewClient(c.AccCPClient)
	if err != nil {
		panic(err)
	}
	middle := &Middle{
		acpClient: acpRPC,
	}
	return middle
}

func (s *Middle) Ban(c *bm.Context) {
	mid, ok := c.Get("mid")
	if !ok {
		return
	}
	if err := s.AccIsAllowed(c, mid.(int64)); err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}
}

// AccIsAllowed .
func (s *Middle) AccIsAllowed(c context.Context, mid int64) error {
	allowReply, err := s.acpClient.IsAllowedToDo(c, &acpAPI.IsAllowedToDoReq{Mid: mid, ControlAction: []string{_allowAction}})
	if err != nil || allowReply == nil {
		log.Error("AccIsAllowed mid(%d) allow(%+v) error(%v)", mid, allowReply, err)
		return ecode.SpaceBanUser
	}
	// 2020_64管控,优先判断
	if status, ok := allowReply.ControlActionStatus[_allowAction]; ok && status != nil && !status.Allowed && status.DeniedByControlRole == _allowAction64 {
		return xecode.ServiceUpdate
	}
	if !allowReply.AllAllowed {
		if ctlAction, ok := allowReply.ControlActionStatus[_allowAction]; ok && ctlAction != nil {
			return xecode.Int(int(ctlAction.ControlEcode))
		}
		return ecode.SpaceBanUser
	}
	return nil
}
