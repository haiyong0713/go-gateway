package di

import (
	"context"
	"time"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/service"
)

//go:generate kratos tool wire
type App struct {
	svc  *service.Service
	http *bm.Engine
}

func NewApp(svc *service.Service, h *bm.Engine) (app *App, closeFunc func(), err error) {
	app = &App{
		svc:  svc,
		http: h,
	}
	closeFunc = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		if err := h.Shutdown(ctx); err != nil {
			log.Error("httpSrv.Shutdown error(%v)", err)
		}
		cancel()
	}
	return
}
