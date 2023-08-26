package act

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go-common/component/metadata/device"
	"go-common/library/log"
	infocv2 "go-common/library/log/infoc.v2"
)

type Report struct {
	infoc infocv2.Infoc
}

func NewReport(cfg *infocv2.Config) *Report {
	infoc, err := infocv2.New(cfg)
	if err != nil {
		panic(err)
	}
	return &Report{infoc: infoc}
}

type PageViewReport struct {
	Ctime    string //上报时间
	Mid      int64
	PageID   int64
	FromType int32 //页面发起类型
	Type     int64 //页面类型
	MobiApp  string
	Build    int64
	Plat     int8
}

func (r *Report) reportPageView(c context.Context, logID string, data *PageViewReport) error {
	if logID == "" {
		return errors.New("logID is empty")
	}
	if data.Ctime == "" {
		data.Ctime = time.Now().Format("2006-01-02 15:04:05")
	}
	if dev, ok := device.FromContext(c); ok {
		data.Plat = dev.Plat()
	}
	args := []log.D{
		log.KV("ctime", data.Ctime),
		log.KV("mid", data.Mid),
		log.KV("page_id", data.PageID),
		log.KV("from_type", data.FromType),
		log.KV("type", data.Type),
		log.KV("mobi_app", data.MobiApp),
		log.KV("build", data.Build),
		log.KV("plat", data.Plat),
	}
	payload := infocv2.NewLogStreamV(logID, args...)
	if err := r.infoc.Info(c, payload); err != nil {
		log.Error("Fail to report PageView, data=%+v error=%+v", data, err)
		return err
	}
	return nil
}

type ModuleViewReport struct {
	Ctime    string //上报时间
	ModuleID int64
	Category int64 //组件类型
	PageID   int64
	FromType int32 //页面发起类型
}

func (r *Report) reportModuleView(c context.Context, logID string, data *ModuleViewReport) error {
	if logID == "" {
		return errors.New("logID is empty")
	}
	if data.Ctime == "" {
		data.Ctime = time.Now().Format("2006-01-02 15:04:05")
	}
	args := []log.D{
		log.KV("ctime", data.Ctime),
		log.KV("module_id", data.ModuleID),
		log.KV("category", data.Category),
		log.KV("page_id", data.PageID),
		log.KV("from_type", data.FromType),
	}
	payload := infocv2.NewLogStreamV(logID, args...)
	if err := r.infoc.Info(c, payload); err != nil {
		log.Error("Fail to report ModuleView, data=%+v error=%+v", data, err)
		return err
	}
	return nil
}
