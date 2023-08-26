package feature

import (
	"context"
	"net/http"
	"net/url"

	"go-gateway/app/app-svr/app-feed/admin/model/tree"

	"github.com/pkg/errors"
)

func (d *Dao) FetchRoleTree(ctx context.Context, cookie string) ([]*tree.Node, error) {
	treeAuthURL, err := url.Parse(d.authURL)
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
	//nolint:gomnd
	if result.Status != 200 {
		return nil, errors.Errorf("Failed to request tree token: %+v", result)
	}
	if result.Data == nil {
		return nil, errors.New("tree token's result.Data is nil")
	}
	roleTreeURL, err := url.Parse(d.roleTreeURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req, err = http.NewRequest("GET", roleTreeURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Authorization-Token", result.Data.Token)
	reply := &tree.Resp{}
	if err := d.http.Do(ctx, req, reply); err != nil {
		return nil, err
	}
	return reply.Data, nil
}
