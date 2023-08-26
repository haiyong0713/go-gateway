package v1

import (
	"context"
	"fmt"

	"go-common/library/net/rpc/warden"

	"google.golang.org/grpc"
	grpcmd "google.golang.org/grpc/metadata"
)

// AppID .
const AppID = "app.listener"

// NewClient new a Podcast/Listener Client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (ListenerClient, error) {
	client := warden.NewClient(cfg, opts...)
	cc, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", AppID))
	if err != nil {
		return nil, err
	}
	return NewListenerClient(cc), nil
}

// NewLegacyMusicClient new a legacy music Client
func NewLegacyMusicClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (MusicClient, error) {
	client := warden.NewClient(cfg, opts...)
	cc, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", AppID))
	if err != nil {
		return nil, err
	}
	return NewMusicClient(cc), nil
}

const (
	_appAuthMetaKey = "x-bili-internal-gw-auth"
	//_grpcDeviceBin      = "x-bili-device-bin"
	//_grpcNetworkBin     = "x-bili-network-bin"
	//_grpcRestrictionBin = "x-bili-restriction-bin"
	//
	//_headerRemoteIP     = "x-backend-bili-real-ip"
	//_headerRemoteIPPort = "x-backend-bili-real-ipport"
	//_headerWebcdnIP     = "X-Cache-Server-Addr"
)

// AttachAppAuthGWKey 将 appName/appKey 添加到grpc outgoing context里
func AttachAppAuthGWKey(ctx context.Context, appName, appKey string) context.Context {
	return setToOutgoingContext(ctx, _appAuthMetaKey, appName+" "+appKey)
}

// 类似于 AppendToOutgoingContext 但是会覆盖已有的key
func setToOutgoingContext(ctx context.Context, key, val string) context.Context {
	md, ok := grpcmd.FromOutgoingContext(ctx)
	if !ok {
		md = grpcmd.Pairs(key, val)
	} else {
		md = md.Copy()
		md.Set(key, val)
	}
	return grpcmd.NewOutgoingContext(ctx, md)
}

// AttachAppDeviceMeta 将 warden解析出的device重新写入请求的grpc metadata中
// 主要用于前端http服务转后端其他grpc时提供相应的metadata
//func AttachAppDeviceMeta(ctx context.Context, dev *wardenDevice.Device) context.Context {
//	if dev == nil {
//		return ctx
//	}
//	cd := grpcDevice.Device{
//		Build:       int32(dev.Build),
//		Buvid:       dev.Buvid,
//		MobiApp:     dev.RawMobiApp,
//		Platform:    dev.RawPlatform,
//		Device:      dev.Device,
//		Channel:     dev.Channel,
//		Brand:       dev.Brand,
//		Model:       dev.Model,
//		Osver:       dev.Osver,
//		FpLocal:     dev.FpLocal,
//		FpRemote:    dev.FpRemote,
//		VersionName: dev.VersionName,
//	}
//	data, _ := cd.Marshal()
//	return grpcmd.AppendToOutgoingContext(ctx, _grpcDeviceBin, string(data))
//}

// AttachAppNetworkMeta 用法同 AttachAppDeviceMeta
//func AttachAppNetworkMeta(ctx context.Context, net *wardenNetwork.Network) context.Context {
//	if net == nil {
//		return ctx
//	}
//	toAdd := make([]string, 0, 4)
//	if len(net.RemoteIP) > 0 {
//		toAdd = append(toAdd, _headerRemoteIP, net.RemoteIP)
//	}
//	if len(net.RemotePort) > 0 {
//		toAdd = append(toAdd, _headerRemoteIPPort, net.RemotePort)
//	}
//	if len(net.WebcdnIP) > 0 {
//		toAdd = append(toAdd, _headerWebcdnIP, net.WebcdnIP)
//	}
//	if len(net.Operator) > 0 {
//		cd := grpcNetwork.Network{
//			Type: grpcNetwork.NetworkType(net.Type),
//			Tf:   grpcNetwork.TFType(net.TF),
//			Oid:  net.Operator,
//		}
//		data, _ := cd.Marshal()
//		toAdd = append(toAdd, _grpcNetworkBin, string(data))
//	}
//
//	if len(toAdd) > 0 {
//		return grpcmd.AppendToOutgoingContext(ctx, toAdd...)
//	}
//	return ctx
//}

// AttachAppRestrictionMeta 用法同 AttachAppDeviceMeta
//func AttachAppRestrictionMeta(ctx context.Context, res *wardenRestrction.Restriction) context.Context {
//	if res == nil {
//		return ctx
//	}
//	cr := grpcRestriction.Restriction{
//		TeenagersMode: res.IsTeenagers,
//		LessonsMode:   res.IsLessons,
//		Review:        res.IsReview,
//		DisableRcmd:   res.DisableRcmd,
//	}
//	data, _ := cr.Marshal()
//	return grpcmd.AppendToOutgoingContext(ctx, _grpcRestrictionBin, string(data))
//}

// AppAuthGWKey 自动为每个请求加上相应的 appname/appkey
func AppAuthGWKey(appName, appKey string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(AttachAppAuthGWKey(ctx, appName, appKey), method, req, reply, cc, opts...)
	}
}

var _defaultPassthroughKeys = []string{
	"x-bili-restriction-bin",
	"x-bili-metadata-bin",
	"x-bili-exps-bin",
	"x-bili-fawkes-req-bin",
	"x-bili-device-bin",
	"x-bili-network-bin",
	"x-bili-locale-bin",
	"authorization",
	"user-agent",
}

// BiliMetaPassthrough 会默认透传一系列 bili 内部定义的metadata
// 额外需要透传的metadata可以由keys指定key
func BiliMetaPassthrough(keys ...string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		incomingMD, ok1 := grpcmd.FromIncomingContext(ctx)
		outgoingMD, ok2 := grpcmd.FromOutgoingContext(ctx)
		if !ok2 {
			outgoingMD = grpcmd.MD{}
		} else {
			outgoingMD = outgoingMD.Copy()
		}
		// 透传常见的客户端metadata
		if ok1 {
			for _, k := range _defaultPassthroughKeys {
				if ins := incomingMD.Get(k); len(ins) > 0 {
					outgoingMD.Set(k, ins...)
				}
			}
		}
		for _, key := range keys {
			if pass := incomingMD.Get(key); len(pass) > 0 {
				outgoingMD.Set(key, pass...)
			}
		}
		return invoker(grpcmd.NewOutgoingContext(ctx, outgoingMD), method, req, reply, cc, opts...)
	}
}
