package anticrawler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go-common/component/metadata/device"
	"go-common/component/metadata/locale"
	"go-common/component/metadata/network"
	"go-common/component/metadata/restriction"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler/model"

	"google.golang.org/grpc"
)

func ReportInterceptor() grpc.UnaryServerInterceptor {
	return reportInterceptor(_antiCrawler.send, grpcFilter())
}

func reportInterceptor(send Send, filter Filter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		if send == nil {
			return
		}
		mid := metadata.Int64(ctx, metadata.Mid)
		var buvid string
		if d, ok := device.FromContext(ctx); ok {
			buvid = d.Buvid
		}
		query, err1 := json.Marshal(req)
		if err1 != nil {
			log.Error("%+v", err1)
			return
		}
		v := resp
		if err != nil {
			v = err
		}
		body, err1 := json.Marshal(v)
		if err1 != nil {
			log.Error("%+v", err1)
			return
		}
		reqHeader := http.Header{}
		device, _ := device.FromContext(ctx) // 获取设备信息
		bs, _ := json.Marshal(device)
		reqHeader.Add("device", string(bs))
		// 获取客户端模式
		restriction, _ := restriction.FromContext(ctx)
		bs, _ = json.Marshal(restriction)
		reqHeader.Add("restriction", string(bs))
		// 获取位置信息
		locale, _ := locale.FromContext(ctx)
		bs, _ = json.Marshal(locale)
		reqHeader.Add("locale", string(bs))
		// 获取网络信息
		network, _ := network.FromContext(ctx)
		bs, _ = json.Marshal(network)
		reqHeader.Add("network", string(bs))
		reqHeaderJSON, _ := json.Marshal(reqHeader)
		sample := random()
		if filter != nil && filter(ctx) {
			sample = -1
		}
		data := &model.InfocMsg{
			Mid:            mid,
			Buvid:          buvid,
			Host:           os.Getenv("APP_ID"),
			Path:           info.FullMethod,
			Method:         "grpc",
			Header:         string(reqHeaderJSON),
			Query:          string(query),
			Body:           "", // TODO
			Referer:        "", // TODO
			IP:             metadata.String(ctx, metadata.RemoteIP),
			Ctime:          time.Now().Unix(),
			ResponseHeader: "", // TODO
			ResponseBody:   string(body),
			Sample:         sample,
		}
		if err1 := send(ctx, data); err1 != nil {
			log.Error("failed to send mogul session: %+v", err1)
		}
		return
	}
}

// grpcFilter return mid filter.
func grpcFilter() Filter {
	return func(ctx context.Context) bool {
		mid := metadata.Int64(ctx, metadata.Mid)
		var buvid string
		if d, ok := device.FromContext(ctx); ok {
			buvid = d.Buvid
		}
		return wList(mid, buvid)
	}
}
