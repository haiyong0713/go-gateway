package dao

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	bapimethod "go-gateway/app/app-svr/app-gw/management/bapi"
	"go-gateway/app/app-svr/app-gw/management/internal/model"

	"github.com/pkg/errors"
)

const (
	_fetchConfigPath = "/x/admin/config/v2/build/list"
)

func (d *dao) FetchConfig(ctx context.Context, id int64, cookie string) ([]*model.ConfigBuildItem, error) {
	params := url.Values{}
	params.Set("tree_id", strconv.FormatInt(id, 10))
	fetchConfigURL, err := url.Parse(d.Hosts.Config + _fetchConfigPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	fetchConfigURL.RawQuery = params.Encode()
	req, err := http.NewRequest(http.MethodGet, fetchConfigURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Cookie", cookie)
	var ret struct {
		Code int                      `json:"code"`
		Msg  string                   `json:"message"`
		Data []*model.ConfigBuildItem `json:"data"`
	}
	if err := d.http.Do(ctx, req, &ret); err != nil {
		return nil, err
	}
	if ret.Code != 0 {
		err := errors.Wrapf(ecode.Int(ret.Code), "Failed to FetchConfig. msg: %+v", ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *dao) ServerMetadata(ctx context.Context, appID string) ([]string, error) {
	metadataURL := fmt.Sprintf("discovery://%s/metadata", appID)
	var ret struct {
		Data map[string]struct {
			Method string `json:"method"`
		} `json:"data"`
	}
	if err := d.http.Get(ctx, metadataURL, "", nil, &ret); err != nil {
		return nil, err
	}
	result := make([]string, 0, len(ret.Data))
	for key := range ret.Data {
		result = append(result, key)
	}
	return result, nil
}

func (d *dao) GRPCServerMethods(ctx context.Context, appIDs []string) (map[string][]string, bool) {
	svrMd, ok := bapimethod.GetByAppIDs(appIDs)
	if !ok {
		return nil, false
	}
	ret := map[string][]string{}
	for appid, pkgs := range svrMd {
		for _, services := range pkgs {
			for _, methods := range services {
				ret[appid] = append(ret[appid], methods...)
			}
		}
	}
	return ret, true
}

func (d *dao) GRPCServerPackages(ctx context.Context, appIDs []string) (map[string]map[string][]string, bool) {
	svrMd, ok := bapimethod.GetByAppIDs(appIDs)
	if !ok {
		return nil, false
	}
	ret := map[string]map[string][]string{}
	for appid, pkgs := range svrMd {
		for pkg, services := range pkgs {
			for service := range services {
				pkgMap, ok := ret[appid]
				if !ok {
					ret[appid] = map[string][]string{}
					pkgMap = ret[appid]
				}
				pkgMap[pkg] = append(pkgMap[pkg], service)
			}
		}
	}
	return ret, true
}
