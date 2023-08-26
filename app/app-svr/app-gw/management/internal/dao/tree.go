package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-gw/management/internal/model/tree"

	"github.com/pkg/errors"
)

const (
	_auth          = "/v1/auth"
	_fetchRoleTree = "/v1/node/role/app"
)

// nolint:gomnd
func (d *dao) fetchRoleTree(ctx context.Context, _, cookie string) ([]*tree.Node, error) {
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
	result := &tree.TokenResult{}
	if err := d.http.Do(ctx, req, result); err != nil {
		return nil, err
	}
	if result.Status != 200 {
		return nil, errors.Errorf("Failed to request tree token: %+v", result)
	}
	token := &tree.Token{}
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
	reply := &tree.Resp{}
	if err := d.http.Do(ctx, req, reply); err != nil {
		return nil, err
	}
	return reply.Data, nil
}

func gwUserTreeKey(username string) string {
	return fmt.Sprintf("{gw-user-tree}/%s", username)
}

func (d *dao) cachedRoleTree(ctx context.Context, username string) ([]*tree.Node, error) {
	key := gwUserTreeKey(username)
	req := d.taishan.NewGetReq([]byte(key))
	record, err := d.taishan.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	out := []*tree.Node{}
	if err := json.Unmarshal(record.Columns[0].Value, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (d *dao) setRoleTreeCache(ctx context.Context, username string, node []*tree.Node) error {
	key := gwUserTreeKey(username)
	value, err := json.Marshal(node)
	if err != nil {
		return err
	}
	req := d.taishan.NewPutReq([]byte(key), value, 1800)
	if err := d.taishan.Put(ctx, req); err != nil {
		return err
	}
	return nil
}

// FetchRoleTree is
func (d *dao) FetchRoleTree(ctx context.Context, username, cookie string) ([]*tree.Node, error) {
	node, err := d.cachedRoleTree(ctx, username)
	if err == nil {
		return node, nil
	}
	if err != nil {
		log.Error("Failed to fetch role tree from cache: %+v", err)
	}
	node, err = d.fetchRoleTree(ctx, username, cookie)
	if err != nil {
		return nil, err
	}
	if err := d.setRoleTreeCache(ctx, username, node); err != nil {
		log.Error("Failed to set role tree cache: %s: %+v", username, err)
	}
	return node, nil
}
