package middle

import (
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"

	mng "go-common/library/net/http/blademaster/middleware/permit/api"

	"github.com/pkg/errors"
)

const (
	_sessUIDKey    = "uid"
	_sessUnKey     = "username"
	_sessIDKey     = "_AJSESSIONID"
	CtxPermissions = "permissions"
)

// Permit manager auth .
type Permit struct {
	mng.PermitClient
}

// New .
func New(c *warden.ClientConfig) *Permit {
	permitClient, err := mng.NewClient(c)
	if err != nil {
		panic(errors.WithMessage(err, "Failed to dial mng rpc server"))
	}
	return &Permit{
		PermitClient: permitClient,
	}
}

// Verify .
func (p *Permit) Verify() bm.HandlerFunc {
	return func(ctx *bm.Context) {
		_, username, err := p.login(ctx)
		if err != nil {
			ctx.JSON(nil, ecode.Unauthorized)
			ctx.Abort()
			return
		}
		ctx.Set(_sessUnKey, username)
	}
}

// Permit .
func (p *Permit) Permit(permit string) bm.HandlerFunc {
	return func(ctx *bm.Context) {
		_, username, err := p.login(ctx)
		if err != nil {
			ctx.JSON(nil, ecode.Unauthorized)
			ctx.Abort()
			return
		}
		ctx.Set(_sessUnKey, username)
		if md, ok := metadata.FromContext(ctx); ok {
			md[metadata.Username] = username
		}
		perReply, err := p.Permissions(ctx, &mng.PermissionReq{Username: username})
		if err != nil {
			if ecode.NothingFound.Equal(err) && permit != "" {
				ctx.JSON(nil, ecode.AccessDenied)
				ctx.Abort()
			}
			return
		}
		ctx.Set(_sessUIDKey, perReply.Uid)
		if md, ok := metadata.FromContext(ctx); ok {
			md[metadata.Uid] = perReply.Uid
		}
		if len(perReply.Perms) > 0 {
			ctx.Set(CtxPermissions, perReply.Perms)
		}
		if !p.permitCheck(permit, perReply.Perms) {
			ctx.JSON(nil, ecode.AccessDenied)
			ctx.Abort()
			return
		}
	}
}

// login .
//
//nolint:unparam
func (p *Permit) login(ctx *bm.Context) (sid, uname string, err error) {
	var dsbsid string
	dsbck, err := ctx.Request.Cookie(_sessIDKey)
	if err == nil {
		dsbsid = dsbck.Value
	}
	if dsbsid == "" {
		err = ecode.Unauthorized
		return
	}
	reply, err := p.Login(ctx, &mng.LoginReq{Mngsid: "", Dsbsid: dsbsid})
	if err != nil {
		log.Error("mng rpc Login error(%v)", err)
		return
	}
	sid = reply.Sid
	uname = reply.Username
	return
}

// permitCheck .
func (p *Permit) permitCheck(permit string, permissions []string) bool {
	if permit == "" {
		return true
	}
	permits := strings.Split(permit, ",")
	for _, ps := range permits {
		//just one auth is ok,will return
		if p.permit(ps, permissions) {
			return true
		}
	}
	return false
}

// permit .
func (p *Permit) permit(permit string, permissions []string) bool {
	if permit == "" {
		return true
	}
	for _, p := range permissions {
		if p == permit {
			return true
		}
	}
	return false
}
