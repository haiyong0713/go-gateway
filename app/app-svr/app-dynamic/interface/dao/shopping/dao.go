package shopping

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	"go-gateway/app/app-svr/app-dynamic/interface/model/shopping"
	mecode "go-gateway/ecode"

	"github.com/pkg/errors"
)

const (
	_listItemCards = "/mall-up-search/items/listItemCards"
	_itemcard      = "/mall-ugc/ugc/content/itemcard"
)

type Dao struct {
	c           *conf.Config
	client      *bm.Client
	itemcardURL string
}

func New(c *conf.Config) *Dao {
	return &Dao{
		c:           c,
		client:      bm.NewClient(c.HTTPClient),
		itemcardURL: c.Hosts.MallCo + _itemcard,
	}
}

func (d *Dao) ItemCard(c context.Context, ids []int64) (map[int64]*shopping.Item, error) {
	params := map[string]interface{}{}
	params["itemIds"] = ids
	paramsb, _ := json.Marshal(params)
	var data struct {
		Code int              `json:"code"`
		Data []*shopping.Item `json:"data"`
	}
	req, err := http.NewRequest(http.MethodPost, d.itemcardURL, strings.NewReader(string(paramsb)))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Error("%s create request failed, param:(%s)", d.itemcardURL, string(paramsb))
		return nil, mecode.ParamInvalid
	}
	if err = d.client.Do(c, req, &data); err != nil {
		log.Error("%s query failed, params:(%s) error(%v)", d.itemcardURL, string(paramsb), err)
		return nil, err
	}
	if data.Code != 0 {
		return nil, errors.Wrap(ecode.Int(data.Code), d.itemcardURL+"?"+string(paramsb))
	}
	res := map[int64]*shopping.Item{}
	for _, v := range data.Data {
		res[v.ID] = v
	}
	return res, nil
}

func (d *Dao) ListItemCards(ctx context.Context, ids []int64) (map[int64]*shopping.CardInfo, error) {
	params := map[string]interface{}{}
	params["bizIds"] = ids
	paramsb, _ := json.Marshal(params)
	var data struct {
		Code int                  `json:"code"`
		Data []*shopping.CardInfo `json:"data"`
	}
	url := d.c.Hosts.MallCo + _listItemCards
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(paramsb)))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Error("%s create request failed, param:(%s)", _listItemCards, string(paramsb))
		return nil, mecode.ParamInvalid
	}
	if err = d.client.JSON(ctx, req, &data); err != nil {
		log.Error("%s query failed, params:(%s) error(%v)", _listItemCards, string(paramsb), err)
		return nil, err
	}
	if data.Code != ecode.OK.Code() {
		log.Error("%+v", err)
		return nil, err
	}
	res := map[int64]*shopping.CardInfo{}
	for _, v := range data.Data {
		res[v.ItemsId] = v
	}
	return res, nil
}
