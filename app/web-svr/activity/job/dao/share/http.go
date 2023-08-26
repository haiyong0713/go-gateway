package share

import (
	"context"
	"fmt"

	"bytes"
	"encoding/json"
	"net/http"

	xecode "go-common/library/ecode"
	"go-gateway/app/web-svr/activity/job/model/share"

	"go-common/library/log"
)

const (
	// shareURLURI 分享url
	shareURLURI = "/x/share/fission"
)

// ShareURL ...
func (d *dao) ShareURL(ctx context.Context, business string, token string, addLinks []string) (*share.Share, error) {
	reqParam := struct {
		Business string   `json:"business"`
		Token    string   `json:"token"`
		AddLinks []string `json:"add_links"`
	}{business, token, addLinks}
	b, _ := json.Marshal(reqParam)
	reply := &share.Share{}
	reply.Location = make([]string, 0)
	req, err := http.NewRequest(http.MethodPost, d.shareURL, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res := struct {
		Code int `json:"code"`
		Data struct {
			Location []string `json:"location"`
		}
		Message string `json:"message"`
	}{}
	if err = d.client.Do(ctx, req, &res); err != nil {
		log.Errorc(ctx, "ShareURL  error(%v)", err)
		return nil, err
	}
	if res.Code != xecode.OK.Code() {
		err = fmt.Errorf("ShareURL params(%s) code(%d)", string(b), res.Code)
		log.Errorc(ctx, "%s,%s", err.Error(), res.Message)
		return nil, err
	}
	reply.Location = res.Data.Location
	return reply, nil
}

// ShareURL ...
func (d *dao) ShareRemoveURL(ctx context.Context, business string, token string, removeLinks []string) (*share.Share, error) {
	reqParam := struct {
		Business    string   `json:"business"`
		Token       string   `json:"token"`
		RemoveLinks []string `json:"remove_links"`
	}{business, token, removeLinks}
	b, _ := json.Marshal(reqParam)
	reply := &share.Share{}
	reply.Location = make([]string, 0)
	req, err := http.NewRequest(http.MethodPost, d.shareURL, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res := struct {
		Code int `json:"code"`
		Data struct {
			Location []string `json:"location"`
		}
		Message string `json:"message"`
	}{}
	if err = d.client.Do(ctx, req, &res); err != nil {
		log.Errorc(ctx, "ShareURL  error(%v)", err)
		return nil, err
	}
	if res.Code != xecode.OK.Code() {
		err = fmt.Errorf("ShareURL params(%s) code(%d)", string(b), res.Code)
		log.Errorc(ctx, "%s,%s", err.Error(), res.Message)
		return nil, err
	}
	reply.Location = res.Data.Location
	return reply, nil
}
