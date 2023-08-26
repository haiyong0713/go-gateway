package bnj2021

import (
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/model/bnj"
	"go-gateway/app/web-svr/activity/job/tool"
)

const (
	metricKey4ExamStats = "bnj_exam_stats"

	sql4ExamStats = `
SELECT 
    ItemID, 
    OptID, 
    countDistinct(MID)
FROM operational_biz.exam_log
GROUP BY 
    ItemID, 
    OptID
`
	sqlTest = `
SELECT 
    (intDiv(toUInt32(EventTime), 100) * 100) * 1000 AS t, 
    count(1), 
    countDistinct(UserID)
FROM operational_biz.hits_v2
WHERE (EventTime >= toDateTime(1395014400)) AND (EventTime <= toDateTime(1395187200)) AND (CounterID IN (8740403, 8740476, 774944))
GROUP BY t
ORDER BY t ASC
`
)

var (
	examStatsRule *bnj.ExamStatsRule
)

func ASyncUpdateExamStats(ctx context.Context) error {
	if examStatsRule == nil || time.Now().Unix() >= examStatsRule.EndTime {
		return nil
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			if now.Unix() >= examStatsRule.EndTime {
				return nil
			}

			if now.Unix() >= examStatsRule.StartTime {
				if err := fetchAndUpdateExamStats(); err == nil {
					tool.IncrBizCountAndLatency(srvName, metricKey4ExamStats, now)
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func fetchAndUpdateExamStats() (err error) {
	if examStatsRule.Url == "" {
		return
	}

	var resp *http.Response
	b := strings.NewReader(sql4ExamStats)
	resp, err = http.Post(examStatsRule.Url, "Content-Type: text/plain", b)
	if err != nil {
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var bs []byte
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	req := new(api.ExamStatsReq)
	{
		req.Stats = make([]*api.OneExamStats, 0)
	}
	records := strings.Split(string(bs), "\n")
	for _, v := range records {
		if v == "" {
			continue
		}

		innerRecords := strings.Split(v, "\t")
		if len(innerRecords) == 3 {
			var itemID, optID, total int64
			itemID, err = strconv.ParseInt(innerRecords[0], 10, 64)
			optID, err = strconv.ParseInt(innerRecords[1], 10, 64)
			total, err = strconv.ParseInt(innerRecords[2], 10, 64)

			if err != nil {
				return
			}

			tmp := new(api.OneExamStats)
			{
				tmp.Id = itemID
				tmp.OptID = optID
				tmp.Total = total
			}

			req.Stats = append(req.Stats, tmp)
		}
	}

	if len(req.Stats) > 0 {
		_, err = client.ActivityClient.UpdateExamStats(context.Background(), req)
	}

	return
}
