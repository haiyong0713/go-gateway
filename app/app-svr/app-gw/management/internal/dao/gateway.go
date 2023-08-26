package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/database/taishan"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model"

	"github.com/pkg/errors"
)

// SetGateway id
func (d *dao) SetGateway(ctx context.Context, req *pb.SetGatewayReq) error {
	key := gatewayKey(req.Node, req.AppName)
	gw := &pb.Gateway{
		AppName:        req.AppName,
		Node:           req.Node,
		TreeId:         req.TreeId,
		Configs:        req.Configs,
		UpdatedAt:      time.Now().Unix(),
		DiscoveryAppid: req.DiscoveryAppid,
		DiscoveryColor: req.DiscoveryColor,
		GrpcConfigs:    req.GrpcConfigs,
	}
	value, err := gw.Marshal()
	if err != nil {
		return err
	}
	putReq := d.taishan.NewPutReq([]byte(key), value, 0)
	return d.taishan.Put(ctx, putReq)
}

func gatewayKey(node, app_name string) string {
	builder := &strings.Builder{}
	builder.WriteString("{gateway}/%s")
	args := []interface{}{node}
	if app_name != "" {
		builder.WriteString("/%s")
		args = append(args, app_name)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func (d *dao) ListGateway(ctx context.Context, node string) ([]*pb.Gateway, error) {
	key := gatewayKey(node, "")
	start, end := fullRange(key)
	out := []*pb.Gateway{}
	req := d.taishan.NewScanReq([]byte(start), []byte(end), 100)
	for {
		reply, err := d.taishan.Scan(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, r := range reply.Records {
			gw := &pb.Gateway{
				Configs:     []*pb.ConfigMeta{},
				GrpcConfigs: []*pb.ConfigMeta{},
			}
			if err := gw.Unmarshal(r.Columns[0].Value); err != nil {
				log.Error("Failed to unmarshal gateway: %+v", errors.WithStack(err))
				continue
			}
			out = append(out, gw)
		}
		if !reply.HasNext {
			break
		}
		req.StartRec = &taishan.Record{
			Key: reply.NextKey,
		}
	}
	return out, nil
}

func (d *dao) DeleteGateway(ctx context.Context, req *pb.DeleteGatewayReq) error {
	key := gatewayKey(req.Node, req.AppName)
	delReq := d.taishan.NewDelReq([]byte(key))
	return d.taishan.Del(ctx, delReq)
}

func (d *dao) getRawGateway(ctx context.Context, node, app_name string) ([]byte, error) {
	key := gatewayKey(node, app_name)
	req := d.taishan.NewGetReq([]byte(key))
	record, err := d.taishan.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return record.Columns[0].Value, nil
}

// EnableALLGatewayConfig is
func (d *dao) EnableALLGatewayConfig(ctx context.Context, req *pb.UpdateALLGatewayConfigReq) error {
	raw, err := d.getRawGateway(ctx, req.Node, req.AppName)
	if err != nil {
		return err
	}
	key := gatewayKey(req.Node, req.AppName)
	gw := &pb.Gateway{}
	if err := gw.Unmarshal(raw); err != nil {
		return err
	}
	for _, conf := range gw.Configs {
		conf.Enable = !req.Disable
	}
	newRaw, err := gw.Marshal()
	if err != nil {
		return err
	}
	casReq := d.taishan.NewCASReq([]byte(key), raw, newRaw)
	return d.taishan.CAS(ctx, casReq)
}

func (d *dao) ProxyPage(ctx context.Context, host, suffix string) (*pb.GatewayProxyReply, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse url: %s", host)
	}
	u.Path = suffix
	req, err := d.http.NewRequest(http.MethodGet, u.String(), "", nil)
	if err != nil {
		return nil, err
	}
	resp, body, err := d.http.RawResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	header := make(map[string]pb.Header)
	for k, v := range resp.Header {
		val := pb.Header{Values: v}
		header[k] = val
	}
	gatewayProxyReply := &pb.GatewayProxyReply{
		Page:       body,
		StatusCode: int32(resp.StatusCode),
		Header:     header,
	}
	return gatewayProxyReply, nil
}

func (d *dao) GatewayProfile(ctx context.Context, host string, isGRPC bool) (*model.GatewayProfile, error) {
	var res struct {
		Code int                   `json:"code"`
		Data *model.GatewayProfile `json:"data"`
	}
	u, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse url: %s", host)
	}
	u.Path = "/_/profile"
	if isGRPC {
		u.Path = "/_/grpc-profile"
	}
	if err := d.http.Get(ctx, u.String(), metadata.String(ctx, metadata.RemoteIP), nil, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), u.String())
		return nil, err
	}
	if res.Data == nil {
		return nil, errors.New("response data is nil")
	}
	return res.Data, nil
}

func (d *dao) AddGatewayConfigFile(ctx context.Context, req *model.AddConfigFileReq) error {
	params := url.Values{}
	params.Set("app_id", req.AppID)
	params.Set("tree_id", strconv.FormatInt(req.TreeID, 10))
	params.Set("env", req.ConfigMeta.Env)
	params.Set("zone", req.ConfigMeta.Zone)
	params.Set("filename", req.ConfigMeta.Filename)
	params.Set("build_name", req.ConfigMeta.BuildName)
	params.Set("force_release", "true")
	pushURL, err := url.Parse(d.Hosts.Config)
	if err != nil {
		return errors.WithStack(err)
	}
	pushURL.RawQuery = params.Encode()
	pushURL.Path = "/admin/v1/update-config-file"
	httpReq, err := http.NewRequest("PUT", pushURL.String(), bytes.NewReader(req.Buffer))
	if err != nil {
		return errors.WithStack(err)
	}
	token := fmt.Sprintf("Token %s", req.ConfigMeta.Token)
	httpReq.Header.Set("Authorization", token)
	httpReq.Header.Set("Content-Type", "application/json")
	var data struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := d.http.Do(ctx, httpReq, &data); err != nil {
		return err
	}
	if data.Code != 0 {
		return errors.WithStack(ecode.Error(ecode.Int(data.Code), data.Message))
	}
	return nil
}

func (d *dao) CreateGatewayConfigBuild(ctx context.Context, req *model.CreateConfigBuildReq) error {
	params := url.Values{}
	params.Set("tree_id", strconv.FormatInt(req.TreeId, 10))
	createConfigBuildURL, err := url.Parse(d.Hosts.Config)
	if err != nil {
		return errors.WithStack(err)
	}
	createConfigBuildURL.Path = "/x/admin/config/v2/build/create"
	createConfigBuildURL.RawQuery = params.Encode()
	var createConfigBuildBody = &struct {
		Name string `json:"name"`
		Env  string `json:"env"`
		Zone string `json:"zone"`
	}{
		req.BuildName,
		req.Env,
		req.Zone,
	}
	body, err := json.Marshal(createConfigBuildBody)
	if err != nil {
		return errors.WithStack(err)
	}
	httpReq, err := http.NewRequest(http.MethodPost, createConfigBuildURL.String(), bytes.NewReader(body))
	if err != nil {
		return errors.WithStack(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Cookie", req.Cookie)
	var data struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := d.http.Do(ctx, httpReq, &data); err != nil {
		return err
	}
	if data.Code != 0 {
		return errors.WithStack(ecode.Error(ecode.Int(data.Code), data.Message))
	}
	return nil
}

// EnableALLGatewayConfig is
func (d *dao) EnableAllGRPCGatewayConfig(ctx context.Context, req *pb.UpdateALLGatewayConfigReq) error {
	raw, err := d.getRawGateway(ctx, req.Node, req.AppName)
	if err != nil {
		return err
	}
	key := gatewayKey(req.Node, req.AppName)
	gw := &pb.Gateway{}
	if err := gw.Unmarshal(raw); err != nil {
		return err
	}
	for _, conf := range gw.GrpcConfigs {
		conf.Enable = !req.Disable
	}
	newRaw, err := gw.Marshal()
	if err != nil {
		return err
	}
	casReq := d.taishan.NewCASReq([]byte(key), raw, newRaw)
	return d.taishan.CAS(ctx, casReq)
}
