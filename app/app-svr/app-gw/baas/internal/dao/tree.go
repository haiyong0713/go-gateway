package dao

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"go-gateway/app/app-svr/app-gw/baas/internal/model"

	"github.com/pkg/errors"
)

const (
	_auth          = "/v1/auth"
	_fetchRoleTree = "/v1/node/role/app"
)

// FetchRoleTree is
//nolint:gomnd
func (d *dao) FetchRoleTree(ctx context.Context, username, cookie string) ([]*model.Node, error) {
	treeAuthURL, err := url.Parse(d.Hosts.Easyst + _auth)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req, err := http.NewRequest("GET", treeAuthURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Cookie", cookie)
	result := &model.TokenResult{}
	if err := d.http.Do(ctx, req, result); err != nil {
		return nil, err
	}
	if result.Status != 200 {
		return nil, errors.Errorf("Failed to request tree token: %+v", result)
	}
	token := &model.Token{}
	if err := json.Unmarshal(result.Data, token); err != nil {
		return nil, errors.WithStack(err)
	}
	roleTreeURL, err := url.Parse(d.Hosts.Easyst + _fetchRoleTree)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req, err = http.NewRequest("GET", roleTreeURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Authorization-Token", token.Token)
	reply := &model.Resp{}
	if err := d.http.Do(ctx, req, reply); err != nil {
		return nil, err
	}
	return reply.Data, nil
}
