package dao

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/database/taishan"
	"go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/app-gw/management/api"

	"github.com/pkg/errors"
)

const (
	_resources       = "/x/internal/quota/resources"
	_addResources    = "/x/internal/quota/resources/add"
	_updateResources = "/x/internal/quota/resources/update"
	_delResources    = "/x/internal/quota/resources/del"
	QuotaAddedErr    = 132011
)

func quotaMethodKey(node, gateway, api, rule string) string {
	builder := &strings.Builder{}
	builder.WriteString("{rate-limiter-%s}/%s")
	args := []interface{}{node, gateway}
	if api != "" {
		builder.WriteString("/%s")
		args = append(args, api)
		if rule != "" {
			builder.WriteString("/%s")
			args = append(args, rule)
		}
	}
	return fmt.Sprintf(builder.String(), args...)
}

func pluginKey(pluginName, field string) string {
	builder := &strings.Builder{}
	builder.WriteString("{plugin-gw}/%s")
	args := []interface{}{pluginName}
	if field != "" {
		builder.WriteString("/%s")
		args = append(args, field)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func (d *dao) GetPlugin(ctx context.Context, pluginName, field string) (*pb.Plugin, error) {
	pluginKey := pluginKey(pluginName, field)
	req := d.taishan.NewGetReq([]byte(pluginKey))
	record, err := d.taishan.Get(ctx, req)
	if err != nil {
		tsErr, ok := errors.Cause(err).(taishanError)
		if !ok {
			return nil, err
		}
		if tsErr.GetMsg() == "KeyNotFoundError" {
			return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("需要先配置quota的token: %+v", err))
		}
		return nil, err
	}
	plugin := &pb.Plugin{}
	if err := plugin.Unmarshal(record.Columns[0].Value); err != nil {
		return nil, err
	}
	return plugin, nil
}

func (d *dao) SetupPlugin(ctx context.Context, pluginName, field string, value *pb.Plugin) error {
	pluginKey := pluginKey(pluginName, field)
	plugin, err := value.Marshal()
	if err != nil {
		return err
	}
	putReq := d.taishan.NewPutReq([]byte(pluginKey), plugin, 0)
	return d.taishan.Put(ctx, putReq)
}

func (d *httpResourceDao) SetQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error {
	return d.dao.setQuotaMethod(ctx, req)
}

func (d *resourceDao) setQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error {
	key := quotaMethodKey(req.Node, req.Gateway, req.Api, req.Rule)
	value, err := req.Marshal()
	if err != nil {
		return err
	}
	putReq := d.taishan.NewPutReq([]byte(key), value, 0)
	return d.taishan.Put(ctx, putReq)
}

func (d *httpResourceDao) GetQuotaMethods(ctx context.Context, node, gateway string) ([]*pb.QuotaMethod, error) {
	return d.dao.getQuotaMethods(ctx, node, gateway)
}

func (d *resourceDao) getQuotaMethods(ctx context.Context, node, gateway string) ([]*pb.QuotaMethod, error) {
	key := quotaMethodKey(node, gateway, "", "")
	start, end := fullRange(key)
	out := []*pb.QuotaMethod{}
	req := d.taishan.NewScanReq([]byte(start), []byte(end), 100)
	for {
		reply, err := d.taishan.Scan(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, r := range reply.Records {
			bapi := &pb.QuotaMethod{}
			if err := bapi.Unmarshal(r.Columns[0].Value); err != nil {
				log.Error("Failed to unmarshal quota method: %+v", errors.WithStack(err))
				continue
			}
			out = append(out, bapi)
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

func (d *resourceDao) getQuotaMethod(ctx context.Context, key string) ([]byte, error) {
	req := d.taishan.NewGetReq([]byte(key))
	record, err := d.taishan.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return record.Columns[0].Value, nil
}

func (d *httpResourceDao) EnableQuotaMethod(ctx context.Context, req *pb.EnableLimiterReq) error {
	return d.dao.enableQuotaMethod(ctx, req)
}

func (d *resourceDao) enableQuotaMethod(ctx context.Context, req *pb.EnableLimiterReq) error {
	key := quotaMethodKey(req.Node, req.Gateway, req.Api, req.Rule)
	raw, err := d.getQuotaMethod(ctx, key)
	if err != nil {
		return err
	}
	qm := &pb.QuotaMethod{}
	if err := qm.Unmarshal(raw); err != nil {
		return err
	}
	qm.Enable = !req.Disable
	newRaw, err := qm.Marshal()
	if err != nil {
		return err
	}
	casReq := d.taishan.NewCASReq([]byte(key), raw, newRaw)
	return d.taishan.CAS(ctx, casReq)
}

func (d *httpResourceDao) DeleteQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error {
	return d.dao.deleteQuotaMethod(ctx, req)
}

func (d *resourceDao) deleteQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error {
	key := quotaMethodKey(req.Node, req.Gateway, req.Api, req.Rule)
	delReq := d.taishan.NewDelReq([]byte(key))
	return d.taishan.Del(ctx, delReq)
}

func (d *dao) QuotaResources(ctx context.Context, id, token string) ([]*pb.Limiter, error) {
	params := url.Values{}
	params.Set("id", id)
	request, err := d.http.NewRequest("GET", d.Hosts.ApiCo+_resources, "", params)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", token)
	var ret struct {
		Code int           `json:"code"`
		Msg  string        `json:"message"`
		Data []*pb.Limiter `json:"data"`
	}
	if err := d.http.Do(ctx, request, &ret); err != nil {
		return nil, err
	}
	if ret.Code != 0 {
		err := errors.Wrapf(ecode.Int(ret.Code), "Failed to get quota resources, id: %s, msg: %+v", id, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *dao) AddQuotaResources(ctx context.Context, req *pb.Limiter, token string) error {
	params := url.Values{}
	params.Set("id", req.Id)
	params.Set("capacity", strconv.FormatInt(req.Capacity, 10))
	params.Set("refresh", strconv.FormatInt(req.RefreshInterval, 10))
	params.Set("algo", strconv.FormatInt(req.Algorithm, 10))
	request, err := d.http.NewRequest("POST", d.Hosts.ApiCo+_addResources, "", params)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err := d.http.Do(ctx, request, &ret); err != nil {
		return err
	}
	if ret.Code == QuotaAddedErr {
		log.Error("Failed to add quota resources: %+v %+v", req, ret.Msg)
		return nil
	}
	if ret.Code != 0 {
		err := errors.Wrapf(ecode.Int(ret.Code), "Failed to add quota resources, req: %+v msg: %+v", req, ret.Msg)
		return err
	}
	return nil
}

func (d *dao) UpdateQuotaResources(ctx context.Context, req *pb.Limiter, token string) error {
	params := url.Values{}
	params.Set("id", req.Id)
	params.Set("capacity", strconv.FormatInt(req.Capacity, 10))
	params.Set("refresh", strconv.FormatInt(req.RefreshInterval, 10))
	params.Set("algo", strconv.FormatInt(req.Algorithm, 10))
	request, err := d.http.NewRequest("POST", d.Hosts.ApiCo+_updateResources, "", params)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err := d.http.Do(ctx, request, &ret); err != nil {
		return err
	}
	if ret.Code != 0 {
		err := errors.Wrapf(ecode.Int(ret.Code), "Failed to update quota resources, req: %+v, msg: %+v", req, ret.Msg)
		return err
	}
	return nil
}

func (d *dao) DeleteQuotaResources(ctx context.Context, id, token string) error {
	params := url.Values{}
	params.Set("id", id)
	request, err := d.http.NewRequest("POST", d.Hosts.ApiCo+_delResources, "", params)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err := d.http.Do(ctx, request, &ret); err != nil {
		return err
	}
	if ret.Code != 0 {
		err := errors.Wrapf(ecode.Int(ret.Code), "Failed to delete quota resources, id: %s, msg: %+v", id, ret.Msg)
		return err
	}
	return nil
}

func (d *dao) PluginList(ctx context.Context, req *pb.PluginListReq) ([]*pb.PluginListItem, error) {
	pluginKey := pluginKey(req.PluginName, "")
	start, end := fullRange(pluginKey)
	out := []*pb.PluginListItem{}
	scanReq := d.taishan.NewScanReq([]byte(start), []byte(end), 100)
	for {
		reply, err := d.taishan.Scan(ctx, scanReq)
		if err != nil {
			return nil, err
		}
		for _, r := range reply.Records {
			gw := &pb.Plugin{}
			if err := gw.Unmarshal(r.Columns[0].Value); err != nil {
				log.Error("Failed to unmarshal gateway: %+v", errors.WithStack(err))
				continue
			}
			item := &pb.PluginListItem{
				Plugin: gw,
				Key:    string(r.Key),
			}
			out = append(out, item)
		}
		if !reply.HasNext {
			break
		}
		scanReq.StartRec = &taishan.Record{
			Key: append(reply.NextKey, 0x00),
		}
	}
	return out, nil
}
