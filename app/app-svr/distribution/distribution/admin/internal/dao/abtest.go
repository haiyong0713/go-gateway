package dao

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	abm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/abtest"

	"github.com/pkg/errors"
)

func (d *dao) SaveABTestConfigs(ctx context.Context, details []*abm.Detail) error {
	taishanBatchPutKey := make(map[string][]byte, len(details))
	for _, v := range details {
		keyInfo := abm.TaishanKeyInfos{
			ID:       v.ID,
			GroupIDs: []string{v.GroupID},
		}
		key := keyInfo.BuildKeys()[0]
		taishanBatchPutKey[key] = v.Config
	}
	req := d.kv.NewBatchPutReq(ctx, taishanBatchPutKey)
	resp, err := d.kv.BatchPut(ctx, req)
	if err != nil {
		return err
	}
	if !resp.AllSucceed {
		return errors.Errorf("Failed to put all config to taishan")
	}
	return nil
}

func (d *dao) FetchAbtestExpID(ctx context.Context, expValue string) (int64, error) {
	uri := fmt.Sprintf("%s%s%s", d.abtestHost, "/abserver/v2/variable/search/", expValue)
	var res struct {
		Code int64   `json:"code"`
		Data []int64 `json:"data"`
	}
	fmt.Println(expValue)
	if err := d.bmClient.Get(ctx, uri, "", url.Values{}, &res); err != nil {
		fmt.Println(err, uri)
		return 0, errors.Wrapf(err, "FetchAbtestExpID uri(%s)", uri)
	}
	if res.Code != http.StatusOK {
		fmt.Println(res.Code)
		return 0, errors.Errorf("FetchAbtestExpID code(%d), uri(%s)", res.Code, uri)
	}
	return res.Data[0], nil
}

func (d *dao) FetchAbtestExpInfo(ctx context.Context, expID string) (*abm.Infos, error) {
	uri := fmt.Sprintf("%s%s%s", d.abtestHost, "/abserver/v2/experiment/", expID)
	var res = struct {
		Code      int64  `json:"code"`
		RunStatus int64  `json:"runStatus"`
		UserName  string `json:"userName"`
		Name      string `json:"name"`
	}{}
	if err := d.bmClient.Get(ctx, uri, "", url.Values{}, &res); err != nil {
		return nil, errors.Wrapf(err, "FetchAbtestExpInfo uri(%s)", uri)
	}
	if res.Code != http.StatusOK {
		return nil, errors.Errorf("FetchAbtestExpInfo code(%d), uri(%s)", res.Code, uri)
	}
	return &abm.Infos{
		ID:      expID,
		Name:    res.Name,
		Creator: res.UserName,
		Status:  strconv.FormatInt(res.RunStatus, 10),
	}, nil
}

func (d *dao) FetchAbtestGroupIDWithName(ctx context.Context, expID string) (map[string]string, error) {
	uri := fmt.Sprintf("%s%s%s", d.abtestHost, "/abserver/v2/experiment/", expID)
	var res = struct {
		Code   int64 `json:"code"`
		Groups []struct {
			Name string `json:"name"`
			ID   int64  `json:"id"`
		} `json:"groups"`
	}{}
	if err := d.bmClient.Get(ctx, uri, "", url.Values{}, &res); err != nil {
		return nil, errors.Wrapf(err, "FetchAbtestGroupIDs uri(%s)", uri)
	}
	if res.Code != http.StatusOK {
		return nil, errors.Errorf("FetchAbtestExpInfo code(%d), uri(%s)", res.Code, uri)
	}
	groupInfos := make(map[string]string, len(res.Groups))
	for _, v := range res.Groups {
		groupInfos[strconv.FormatInt(v.ID, 10)] = v.Name
	}
	return groupInfos, nil
}

func (d *dao) BatchFetchAbtestExpID(ctx context.Context, expValues []string) (map[string][]int64, error) {
	var queryValue string
	for _, v := range expValues {
		if queryValue == "" {
			queryValue = v
			continue
		}
		queryValue = fmt.Sprintf("%s,%s", queryValue, v)
	}
	url := fmt.Sprintf("%s%s%s", d.abtestHost, "/abserver/v2/variable/search?variableNameList=", queryValue)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var res = struct {
		Code int64              `json:"code"`
		Data map[string][]int64 `json:"data"`
	}{}
	if err := d.bmClient.Do(ctx, req, &res); err != nil {
		fmt.Println(err)
		return nil, errors.Wrapf(err, "FetchAbtestGroupIDs uri(%s)", url)
	}
	if res.Code != http.StatusOK {
		return nil, errors.Errorf("FetchAbtestExpInfo code(%d), uri(%s)", res.Code, url)
	}
	return res.Data, nil
}

func (d *dao) BatchFetchAbtestExpInfo(ctx context.Context, expValueIDMap map[string]int64) ([]*abm.Infos, error) {
	var queryValue string
	for _, v := range expValueIDMap {
		if queryValue == "" {
			queryValue = fmt.Sprintf("%d", v)
			continue
		}
		queryValue = fmt.Sprintf("%s,%d", queryValue, v)
	}
	url := fmt.Sprintf("%s%s%s", d.abtestHost, "/abserver/v2/experiment?experimentIdList=", queryValue)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var res = struct {
		Code  int64 `json:"code"`
		Items []struct {
			RunStatus int64  `json:"runStatus"`
			UserName  string `json:"userName"`
			Name      string `json:"name"`
			ID        int64  `json:"id"`
			Groups    []struct {
				GroupVariables []struct {
					VarName string `json:"varName"`
				} `json:"groupVariables"`
			} `json:"groups"`
		} `json:"items"`
	}{}
	if err := d.bmClient.Do(ctx, req, &res); err != nil {
		fmt.Println(err)
		return nil, errors.Wrapf(err, "FetchAbtestGroupIDs uri(%s)", url)
	}
	if res.Code != http.StatusOK {
		return nil, errors.Errorf("FetchAbtestExpInfo code(%d), uri(%s)", res.Code, url)
	}
	var abInfos []*abm.Infos
	for _, v := range res.Items {
		abInfo := &abm.Infos{
			ID:        strconv.FormatInt(v.ID, 10),
			Name:      v.Name,
			Creator:   v.UserName,
			Status:    strconv.FormatInt(v.RunStatus, 10),
			FlagValue: v.Groups[0].GroupVariables[0].VarName,
		}
		abInfos = append(abInfos, abInfo)
	}
	return abInfos, nil
}
