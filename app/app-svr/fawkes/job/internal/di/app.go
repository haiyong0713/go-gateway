package di

import (
	"context"
	"time"

	"go-gateway/app/app-svr/fawkes/job/internal/service/mod"
	packSvc "go-gateway/app/app-svr/fawkes/job/internal/service/pack"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

//go:generate kratos tool wire
type App struct {
	modSvc  *mod.Service
	packSvc *packSvc.Service
	http    *bm.Engine
}

func NewApp(modSvc *mod.Service, packSvc *packSvc.Service, h *bm.Engine) (app *App, closeFunc func(), err error) {
	app = &App{
		modSvc:  modSvc,
		packSvc: packSvc,
		http:    h,
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
