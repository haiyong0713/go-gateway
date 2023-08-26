package note

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"io/ioutil"
	"net/http"
)

type hotArcItem struct {
	ID   int64  `json:"id"`
	GOTO string `json:"goto"`
}

func (d *Dao) HotArchives(ctx context.Context) ([]int64, error) {

	req, err := http.NewRequest(http.MethodGet, d.c.NoteCfg.Host.HotArchiveHost, nil)
	if err != nil {
		log.Errorc(ctx, "d.HotArchives http.NewRequest url(%s) error(%+v)", d.c.NoteCfg.Host.HotArchiveHost, err)
		return nil, err
	}

	q := req.URL.Query()
	q.Set("cmd", "hot")
	q.Set("from", "10")

	req.URL.RawQuery = q.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorc(ctx, "d.HotArchives http.DefaultClient.Do url(%s %s) error(%+v)", d.c.NoteCfg.Host.HotArchiveHost, q.Encode(), err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, nil
	}
	respBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	resp := new(struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data []*hotArcItem `json:"data"`
	})

	err = json.Unmarshal(respBytes, &resp)
	if err != nil {
		log.Errorc(ctx, "d.HotArchives json.Unmarshal(%v) error(%v)", string(respBytes), err)
		return nil, err
	}

	if resp == nil || len(resp.Data) == 0 {
		return nil, err
	}

	var aids []int64

	for _, v := range resp.Data {
		if v.GOTO != "av" {
			continue
		}
		aids = append(aids, v.ID)
	}

	//log.Infoc(ctx, string(respBytes))

	return aids, nil
}
