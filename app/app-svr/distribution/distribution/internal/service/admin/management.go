package service

import (
	"context"
	"encoding/json"

	"go-common/component/metadata/device"
	"go-common/library/log"
	"go-gateway/app/app-svr/distribution/distribution/api"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	"go-gateway/app/app-svr/distribution/distribution/internal/sessioncontext"

	safecenter "git.bilibili.co/bapis/bapis-go/passport/service/safecenter"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

type UserDeviceRequest struct {
	Mid int64 `form:"mid" validate:"required"`
}

type UserDeviceReply struct {
	Device []*DeviceMeta `json:"device"`
}

type DeviceMeta struct {
	Buvid   string `json:"buvid"`
	FpLocal string `json:"fp_local"`
	Time    string `json:"time"`
}

func (s *Service) UserDevice(ctx context.Context, req *UserDeviceRequest) (*UserDeviceReply, error) {
	deviceALL, err := s.safecenter.UserDevicesAll(ctx, &safecenter.UserDeviceMidReq{
		Mid: req.Mid,
	})
	if err != nil {
		return nil, err
	}
	reply := &UserDeviceReply{
		Device: make([]*DeviceMeta, 0, len(deviceALL.Infos)),
	}
	for _, d := range deviceALL.Infos {
		reply.Device = append(reply.Device, &DeviceMeta{
			Buvid:   d.DeviceId,
			Time:    d.Time,
			FpLocal: "",
		})
	}
	return reply, nil
}

type DevicePreferenceRequest struct {
	Buvid        string            `form:"buvid" validate:"required"`
	FpLocal      string            `form:"fp_local"`
	Mid          int64             `form:"mid"`
	ExtraContext map[string]string `form:"extra_context"`
}

type DevicePreferenceRequestContext struct {
	ctx context.Context
	req *DevicePreferenceRequest
}

func (d *DevicePreferenceRequestContext) Device() device.Device {
	dev, _ := device.FromContext(d.ctx)
	dev.Buvid = d.req.Buvid
	dev.FpLocal = d.req.FpLocal
	return dev
}

func (d *DevicePreferenceRequestContext) Mid() int64 {
	return d.req.Mid
}

func (d *DevicePreferenceRequestContext) ExtraContext() map[string]string {
	return d.req.ExtraContext
}

func (d *DevicePreferenceRequestContext) ExtraContextValue(key string) (string, bool) {
	v, ok := d.req.ExtraContext[key]
	return v, ok
}

type DevicePreferenceReply struct {
	Preference []*DevicePreferenceItem `json:"preference"`
}

type DevicePreferenceItem struct {
	FullyQualifiedName string          `json:"fully_qualified_name"`
	Data               json.RawMessage `json:"data"`
}

func preferenceJSONify(in *dynamic.Message) string {
	data, err := in.MarshalJSON()
	if err != nil {
		log.Error("Failed to marshal preference as json: %+v", err)
		return ""
	}
	return string(data)
}

func (s *Service) DevicePreference(ctx context.Context, req *DevicePreferenceRequest) (*DevicePreferenceReply, error) {
	ctx = sessioncontext.NewContext(ctx, &DevicePreferenceRequestContext{
		ctx: ctx,
		req: req,
	})

	preference, err := s.origin.UserPreference(ctx, &api.UserPreferenceReq{})
	if err != nil {
		return nil, err
	}

	reply := &DevicePreferenceReply{
		Preference: make([]*DevicePreferenceItem, 0, len(preference.Preference)),
	}
	for _, p := range preference.Preference {
		messageName, err := ptypes.AnyMessageName(p)
		if err != nil {
			log.Error("Failed to extract any message name: %q: %+v", p.TypeUrl, errors.WithStack(err))
			continue
		}
		pm, ok := preferenceproto.TryGetPreference(messageName)
		if !ok {
			log.Error("Failed to get preference meta: %q", messageName)
			continue
		}
		ctr := dynamic.NewMessage(pm.ProtoDesc)
		if err := ctr.Unmarshal(p.Value); err != nil {
			log.Error("Failed to unmarshal preference data: %+v", errors.WithStack(err))
			continue
		}
		reply.Preference = append(reply.Preference, &DevicePreferenceItem{
			FullyQualifiedName: pm.ProtoDesc.GetFullyQualifiedName(),
			Data:               json.RawMessage(preferenceJSONify(ctr)),
		})
	}

	return reply, nil
}
