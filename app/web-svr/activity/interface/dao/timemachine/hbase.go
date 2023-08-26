package timemachine

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/timemachine"

	"github.com/tsuna/gohbase/hrpc"
)

const (
	_hBaseUpItemTableName = "dw:mid_year_report"
)

// reverse for string.
func hbaseRowKey(mid int64) string {
	s := strconv.FormatInt(mid, 10)
	rs := []rune(s)
	l := len(rs)
	for f, t := 0, l-1; f < t; f, t = f+1, t-1 {
		rs[f], rs[t] = rs[t], rs[f]
	}
	ns := string(rs)
	if l < 10 {
		for i := 0; i < 10-l; i++ {
			ns = ns + "0"
		}
	}
	return ns
}

// RawTimemachine .
func (d *Dao) RawTimemachine(c context.Context, mid int64) (data *timemachine.Item, err error) {
	var (
		result      *hrpc.Result
		ctx, cancel = context.WithTimeout(c, time.Duration(d.c.Hbase.RegionReadTimeout))
		key         = hbaseRowKey(mid)
		tableName   = _hBaseUpItemTableName
	)
	defer cancel()
	if result, err = d.hbase.GetStr(ctx, tableName, key); err != nil {
		log.Error("RawTimemachine d.hbase.GetStr tableName(%s)|mid(%d)|key(%v)|error(%v)", tableName, mid, key, err)
		return
	}
	if result == nil {
		return
	}
	data = &timemachine.Item{Mid: mid}
	for _, c := range result.Cells {
		if c == nil {
			continue
		}
		if !bytes.Equal(c.Family, []byte("m")) {
			continue
		}
		tmFillFields(data, c)
	}
	return
}

func (d *Dao) timemachineScan(c context.Context, startRow, endRow string) (err error) {
	var scanner hrpc.Scanner
	if scanner, err = d.hbase.ScanRangeStr(c, _hBaseUpItemTableName, startRow, endRow); err != nil {
		log.Error("TimemachineScan d.hbase.Scan(%s) error(%v)", _hBaseUpItemTableName, err)
		return
	}
	for {
		if d.tmProcStop != 0 {
			err = scanner.Close()
			return
		}
		result, e := scanner.Next()
		if e != nil {
			if e == io.EOF {
				return
			}
			log.Error("TimemachineScan scanner.Next error(%v)", e)
			continue
		}
		if result == nil {
			continue
		}
		item := &timemachine.Item{}
		for _, c := range result.Cells {
			if c == nil {
				continue
			}
			if !bytes.Equal(c.Family, []byte("m")) {
				continue
			}
			tmFillFields(item, c)
		}
		if item.Mid == 0 {
			log.Error("TimemachineScan item.Mid == 0")
		}
		if item.Mid > 0 {
			if e := func() (err error) {
				for i := 0; i < 3; i++ {
					if err = d.AddCacheTimemachine(c, item.Mid, item); err == nil {
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
				return
			}(); e != nil {
				log.Error("TimemachineScan.AddCacheTimemachine(%d) error(%v)", item.Mid, e)
			}
		}
	}
}

func tmFillFields(data *timemachine.Item, c *hrpc.Cell) {
	var (
		intVal   int64
		strVal   string
		floatVal float64
	)
	strVal = string(c.Value[:])
	if v, e := strconv.ParseInt(string(c.Value[:]), 10, 64); e == nil {
		intVal = v
	} else {
		if v, e := strconv.ParseFloat(string(c.Value[:]), 64); e == nil {
			floatVal = v
		}
	}
	switch {
	case bytes.Equal(c.Qualifier, []byte("mid")):
		data.Mid = intVal
	case bytes.Equal(c.Qualifier, []byte("visit_ds")):
		data.VisitDays = intVal
	case bytes.Equal(c.Qualifier, []byte("hour_vd")):
		data.HourVisitDays = strVal
	case bytes.Equal(c.Qualifier, []byte("max_vh")):
		data.MaxVisitDaysHour = intVal
	case bytes.Equal(c.Qualifier, []byte("vv")):
		data.Vv = intVal
	case bytes.Equal(c.Qualifier, []byte("tid")):
		data.MaxVvTid = int32(intVal)
	case bytes.Equal(c.Qualifier, []byte("tid_score")):
		data.Top6VvTidScore = strVal
	case bytes.Equal(c.Qualifier, []byte("sub_tid")):
		data.MaxVvSubtid = int32(intVal)
	case bytes.Equal(c.Qualifier, []byte("tags")):
		data.Top10VvTag = strVal
	case bytes.Equal(c.Qualifier, []byte("is_cn")):
		data.IsCoin = intVal
	case bytes.Equal(c.Qualifier, []byte("coin_tm")):
		data.CoinTime = strVal
	case bytes.Equal(c.Qualifier, []byte("coin_users")):
		data.CoinUsers = intVal
	case bytes.Equal(c.Qualifier, []byte("coin_av")):
		data.CoinAvid = intVal
	case bytes.Equal(c.Qualifier, []byte("ams_d")):
		data.PlayAmsDuration = intVal
	case bytes.Equal(c.Qualifier, []byte("fjs")):
		data.PlayFjs = intVal
	case bytes.Equal(c.Qualifier, []byte("gcs")):
		data.PlayGcs = intVal
	case bytes.Equal(c.Qualifier, []byte("like_sid")):
		data.BestLikeSid = int32(intVal)
	case bytes.Equal(c.Qualifier, []byte("is_ys")):
		data.IsNeedShowYingshi = intVal
	case bytes.Equal(c.Qualifier, []byte("movies")):
		data.PlayMovies = intVal
	case bytes.Equal(c.Qualifier, []byte("dramas")):
		data.PlayDramas = intVal
	case bytes.Equal(c.Qualifier, []byte("jlps")):
		data.PlayDocumentarys = intVal
	case bytes.Equal(c.Qualifier, []byte("zjs")):
		data.PlayZongyi = intVal
	case bytes.Equal(c.Qualifier, []byte("like_yinshi")):
		data.BestLikeYinshi = int32(intVal)
	case bytes.Equal(c.Qualifier, []byte("rhe")):
		data.IsReadHotEvent = intVal
	case bytes.Equal(c.Qualifier, []byte("fvt")):
		data.FirstViewTime = strVal
	case bytes.Equal(c.Qualifier, []byte("event_id")):
		data.EventID = intVal
	case bytes.Equal(c.Qualifier, []byte("like_up")):
		data.LikeBestUp = intVal
	case bytes.Equal(c.Qualifier, []byte("like_up_cr")):
		data.LikeUpBestCreate = intVal
	case bytes.Equal(c.Qualifier, []byte("like_liveup")):
		data.LikeBestLiveUp = intVal
	case bytes.Equal(c.Qualifier, []byte("like_liveup_d")):
		data.LikeLiveupPlayDuration = intVal
	case bytes.Equal(c.Qualifier, []byte("up")):
		data.IsValidup = intVal
	case bytes.Equal(c.Qualifier, []byte("crs")):
		data.Creates = intVal
	case bytes.Equal(c.Qualifier, []byte("cr_vv")):
		data.CreateVv = intVal
	case bytes.Equal(c.Qualifier, []byte("avs")):
		data.CreateAvs = intVal
	case bytes.Equal(c.Qualifier, []byte("rds")):
		data.CreateReads = intVal
	case bytes.Equal(c.Qualifier, []byte("av_vv")):
		data.AvVv = intVal
	case bytes.Equal(c.Qualifier, []byte("rd_vv")):
		data.ReadVv = intVal
	case bytes.Equal(c.Qualifier, []byte("best_av")):
		data.BestCreate = intVal
	case bytes.Equal(c.Qualifier, []byte("have_bf")):
		data.IsHaveBestFan = intVal
	case bytes.Equal(c.Qualifier, []byte("best_fan")):
		data.BestFanMid = intVal
	case bytes.Equal(c.Qualifier, []byte("bf_vv")):
		data.BestFanVv = intVal
	case bytes.Equal(c.Qualifier, []byte("live_up")):
		data.IsValidLiveUp = intVal
	case bytes.Equal(c.Qualifier, []byte("live_d")):
		data.LiveDays = intVal
	case bytes.Equal(c.Qualifier, []byte("ratio")):
		data.Ratio = floatVal
	case bytes.Equal(c.Qualifier, []byte("date")):
		data.MaxOnlineNumDate = strVal
	case bytes.Equal(c.Qualifier, []byte("cdn_num")):
		data.MaxOnlineNum = intVal
	case bytes.Equal(c.Qualifier, []byte("best_live_fan")):
		data.BestLiveFanMid = intVal
	case bytes.Equal(c.Qualifier, []byte("hour_pd")):
		data.HourPlayDays = strVal
	case bytes.Equal(c.Qualifier, []byte("max_ph")):
		data.MaxPlayHour = intVal
	case bytes.Equal(c.Qualifier, []byte("play_d")):
		data.PlayDays = intVal
	case bytes.Equal(c.Qualifier, []byte("blsr")):
		data.BestLikeSidRep = intVal
	case bytes.Equal(c.Qualifier, []byte("up_bestav")):
		data.UpBestAv = intVal
	case bytes.Equal(c.Qualifier, []byte("up_bestav_r")):
		data.UpBestAvRep = intVal
	case bytes.Equal(c.Qualifier, []byte("live_h")):
		data.LiveHour = floatVal
	case bytes.Equal(c.Qualifier, []byte("rbls")):
		data.RealBestLikeSid = intVal
	case bytes.Equal(c.Qualifier, []byte("rllpd")):
		data.RealLikeLiveupPlayDuration = floatVal
	case bytes.Equal(c.Qualifier, []byte("best_cr_type")):
		data.UpBestCreatType = intVal
	}
}
