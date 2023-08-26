package logcontext

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"

	"go-common/library/conf/paladin.v2"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"github.com/pkg/errors"
)

type logContextKey struct{}

type LogContext interface {
	ActionType() int
	BusinessId() int
	UserName() string
	ExtraContext() map[string]string
	ExtraContextValue(key string) (string, bool)
}

func FromContext(ctx context.Context) (LogContext, bool) {
	ssCtx, ok := ctx.Value(logContextKey{}).(LogContext)
	return ssCtx, ok
}

func NewContext(ctx context.Context, s LogContext) context.Context {
	ctx = context.WithValue(ctx, logContextKey{}, s)
	return ctx
}

type logContextImpl struct {
	actionType   int
	businessID   int
	userName     string
	extraContext map[string]string
}

func (l *logContextImpl) UserName() string {
	return l.userName
}

func (l *logContextImpl) ExtraContext() map[string]string {
	return l.extraContext
}

func (l *logContextImpl) ExtraContextValue(key string) (string, bool) {
	v, ok := l.extraContext[key]
	return v, ok
}

func (l *logContextImpl) BusinessId() int {
	return l.businessID
}

func (l *logContextImpl) ActionType() int {
	return l.actionType
}

func ActionLogHandler() bm.HandlerFunc {
	return func(c *bm.Context) {
		l := &logContextImpl{}
		usernameCookie, err := c.Request.Cookie("username")
		if err != nil {
			c.JSON(nil, errors.Wrapf(ecode.NothingFound, "No username"))
			c.Abort()
			return
		}
		l.userName = usernameCookie.Value
		bs, _ := ioutil.ReadAll(c.Request.Body)
		l.extraContext = map[string]string{
			"form": c.Request.Form.Encode(),
			"body": string(bs),
		}
		_ = c.Request.Body.Close()
		c.Request.GetBody = func() (io.ReadCloser, error) {
			r := bytes.NewReader(bs)
			return ioutil.NopCloser(r), nil
		}
		ac := &paladin.TOML{}
		if err := paladin.Watch("application.toml", ac); err != nil {
			return
		}
		switch c.Request.URL.Path {
		case "/x/admin/distribution/tus/multiple/edit/performance/save":
			l.businessID, _ = ac.Get("editBusinessID").Int()
		case "/x/admin/distribution/tus/multiple/save":
			l.businessID, _ = ac.Get("businessID").Int()
			l.actionType = 2
		case "/x/admin/distribution/tus/save":
			l.businessID, _ = ac.Get("businessID").Int()
			l.actionType = 1
		case "/x/admin/distribution/abtest/save":
			l.businessID, _ = ac.Get("businessID").Int()
			l.actionType = 0
		case "/x/admin/distribution/tus/multiple/version/add":
			l.businessID, _ = ac.Get("businessID").Int()
			l.actionType = 3
		case "/x/admin/distribution/tus/multiple/version/update/buildlmit":
			l.businessID, _ = ac.Get("businessID").Int()
			l.actionType = 3
		case "/x/admin/distribution/tus/multiple/version/delete":
			l.businessID, _ = ac.Get("businessID").Int()
			l.actionType = 3
		}
		c.Context = NewContext(c.Context, l)
	}
}
