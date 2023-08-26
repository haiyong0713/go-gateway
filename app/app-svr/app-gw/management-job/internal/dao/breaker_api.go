package dao

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/database/taishan"
	"go-common/library/log"
	gwconfig "go-gateway/app/app-svr/app-gw/management-job/internal/model/gateway-config"
	pb "go-gateway/app/app-svr/app-gw/management/api"

	"github.com/pkg/errors"
)

const (
	_configHost = "discovery://main.common-arch.config-admin"
	_pushURL    = "/x/admin/config/openapi/v1/update-config-file"
	_rawURL     = "/x/admin/config/openapi/v1/get-config-file"
)

func breakerAPIKey(node, gateway, api string) string {
	builder := &strings.Builder{}
	builder.WriteString("{breakerapi-%s}/")
	args := []interface{}{node}
	if gateway != "" {
		builder.WriteString("%s/")
		args = append(args, gateway)
	}
	if api != "" {
		builder.WriteString("%s")
		args = append(args, api)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func grpcBreakerAPIKey(node, gateway, api string) string {
	builder := &strings.Builder{}
	builder.WriteString("{grpc-breakerapi-%s}/")
	args := []interface{}{node}
	if gateway != "" {
		builder.WriteString("%s/")
		args = append(args, gateway)
	}
	if api != "" {
		builder.WriteString("%s")
		args = append(args, api)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func fullRange(prefix string) (string, string) {
	return fmt.Sprintf("%s\x00", prefix), fmt.Sprintf("%s\xFF", prefix)
}

func filterBreakderAPI(in []*pb.BreakerAPI, filter func(*pb.BreakerAPI) bool) []*pb.BreakerAPI {
	filtered := []*pb.BreakerAPI{}
	for _, bapi := range in {
		if !filter(bapi) {
			continue
		}
		filtered = append(filtered, bapi)
	}
	return filtered
}

// ListBreakerAPI is
func (d *dao) ListBreakerAPI(ctx context.Context, node string, gateway string) ([]*pb.BreakerAPI, error) {
	key := breakerAPIKey(node, "", "")
	start, end := fullRange(key)
	out := []*pb.BreakerAPI{}
	req := d.taishan.NewScanReq([]byte(start), []byte(end), 100)
	for {
		reply, err := d.taishan.Scan(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, r := range reply.Records {
			bapi := &pb.BreakerAPI{
				Action: &pb.BreakerAction{},
			}
			if err := bapi.Unmarshal(r.Columns[0].Value); err != nil {
				log.Error("Failed to unmarshal breaker api: %+v", errors.WithStack(err))
				continue
			}
			out = append(out, bapi)
		}
		if !reply.HasNext {
			break
		}
		req.StartRec = &taishan.Record{
			Key: append(reply.NextKey, 0x00),
		}
	}

	filtered := filterBreakderAPI(out, func(bapi *pb.BreakerAPI) bool {
		if gateway == "" {
			return true
		}
		return bapi.Gateway == gateway
	})
	return filtered, nil
}

// GRPCListBreakerAPI is
func (d *dao) GRPCListBreakerAPI(ctx context.Context, node string, gateway string) ([]*pb.BreakerAPI, error) {
	key := grpcBreakerAPIKey(node, "", "")
	start, end := fullRange(key)
	out := []*pb.BreakerAPI{}
	req := d.taishan.NewScanReq([]byte(start), []byte(end), 100)
	for {
		reply, err := d.taishan.Scan(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, r := range reply.Records {
			bapi := &pb.BreakerAPI{
				Action: &pb.BreakerAction{},
			}
			if err := bapi.Unmarshal(r.Columns[0].Value); err != nil {
				log.Error("Failed to unmarshal breaker api: %+v", errors.WithStack(err))
				continue
			}
			out = append(out, bapi)
		}
		if !reply.HasNext {
			break
		}
		req.StartRec = &taishan.Record{
			Key: append(reply.NextKey, 0x00),
		}
	}

	filtered := filterBreakderAPI(out, func(bapi *pb.BreakerAPI) bool {
		if gateway == "" {
			return true
		}
		return bapi.Gateway == gateway
	})
	return filtered, nil
}

// PushConfigs is
func (d *dao) PushConfigs(ctx context.Context, req *gwconfig.PushConfigReq) error {
	params := url.Values{}
	params.Set("app_id", req.AppID)
	params.Set("tree_id", strconv.FormatInt(req.TreeID, 10))
	params.Set("env", req.ConfigMeta.Env)
	params.Set("zone", req.ConfigMeta.Zone)
	params.Set("filename", req.ConfigMeta.Filename)
	params.Set("build_name", req.ConfigMeta.BuildName)
	params.Set("force_release", "true")
	request, err := http.NewRequest("PUT", fmt.Sprintf("%s?%s", _configHost+_pushURL, params.Encode()), bytes.NewReader(req.Buffer))
	if err != nil {
		return err
	}
	token := fmt.Sprintf("Token %s", req.ConfigMeta.Token)
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/octet-stream")

	if err = d.httpClient.Do(ctx, request, nil); err != nil {
		return err
	}

	return nil
}

// RawConfigs is
func (d *dao) RawConfigs(ctx context.Context, req *gwconfig.RawConfigReq) ([]byte, error) {
	params := url.Values{}
	params.Set("app_id", req.AppID)
	params.Set("tree_id", strconv.FormatInt(req.TreeID, 10))
	params.Set("env", req.ConfigMeta.Env)
	params.Set("zone", req.ConfigMeta.Zone)
	params.Set("filename", req.ConfigMeta.Filename)
	params.Set("build_name", req.ConfigMeta.BuildName)

	request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", _configHost+_rawURL, params.Encode()), nil)
	if err != nil {
		return nil, err
	}
	token := fmt.Sprintf("Token %s", req.ConfigMeta.Token)
	request.Header.Set("Authorization", token)

	res, err := d.httpClient.Raw(ctx, request)
	if err != nil {
		return nil, err
	}

	return res, nil
}
