package ugctab

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/app-svr/resource/service/model"
	bvid2 "go-gateway/pkg/idsafe/bvid"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	dataFormat = "2006-01-02 15:04:00"
	// sql
	getEffectiveUgcSQL = `select 
			id, tab_type, tab, link_type, link, background,
			selected_color, txt_color, builds, ugc_type, arctype,
			tagid, upid, avid, avid_file
		from ugctab
		where online=1 and deleted=0 and stime <= ? and etime >= ?
		order by stime asc
	`
)

func (d *Dao) GetMysqlUgcTab(c context.Context) (ret []*model.UgcTabItem, err error) {
	now := time.Now().Format(dataFormat)
	rows, err := d.db.Query(c, getEffectiveUgcSQL, now, now)
	if err != nil {
		log.Error("db Query error: %s", err)
		return
	}
	ret = make([]*model.UgcTabItem, 0)
	defer rows.Close()
	for rows.Next() {
		item := &model.UgcTabItem{}
		if err = rows.Scan(&item.ID, &item.TabType, &item.Tab, &item.LinkType, &item.Link, &item.Bg, &item.Selected, &item.Color, &item.Builds,
			&item.UgcType, &item.Arctype, &item.Tagid, &item.Upid, &item.Avid, &item.AvidFile); err != nil {
			log.Error("rows Scan error: %s", err)
			return
		}
		ret = append(ret, item)
	}
	err = rows.Err()
	if err != nil {
		log.Error("rows error: %s", err)
	}
	return ret, err
}

func (d *Dao) FetchAvidFromFile(filePath string) (map[string]bool, string, error) {
	var (
		avidMap = make(map[string]bool)
		resp    *http.Response
		err     error
	)

	// 此处会调用各类bfs上的txt或者csv文件
	// nolint:gosec
	if resp, err = http.Get(filePath); err != nil {
		log.Error("FetchAvidFromFile get file error: %s", err)
		return avidMap, "", err
	}
	var body []byte
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("FetchAvidFromFile read body error: %s", err)
		return avidMap, "", err
	}
	raw := string(body)
	raw = strings.Replace(raw, "\r", "", -1)
	raw = strings.Replace(raw, " ", "", -1)
	rows := strings.Split(raw, "\n")
	result := make([]string, 0)
	for i := range rows {
		bvid := rows[i]
		if i == 0 {
			bvid = strings.Replace(bvid, "\uFEFF", "", -1)
		}
		if aid, e := bvid2.BvToAv(bvid); e != nil {
			log.Error("FetchAvidFromFile avid transfer error: %s, bvid: (%+v)", e, bvid)
			continue
		} else {
			result = append(result, strconv.FormatInt(aid, 10))
		}
	}

	for _, v := range result {
		avidMap[v] = true
	}

	return avidMap, strings.Join(result, ","), nil
}
