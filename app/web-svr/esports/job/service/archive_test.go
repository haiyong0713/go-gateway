package service

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	arcclient "git.bilibili.co/bapis/bapis-go/archive/service"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/esports/job/model"
)

// go test -v archive_test.go archive.go service.go pointdata.go leidata.go ftp.go match.go score.go bfs.go
func TestArchiveBiz(t *testing.T) {
	t.Run("init archive download shell script", testInitShellScript)
	t.Run("download baidu html", testDownloadBiz)
	t.Run("calculate archive score", testCalculateScore)
	t.Run("test genArchiveStats biz", testGenArchiveStatsByFileRaw)
	t.Run("test calculate biz", testScoreCalculate)
}

func testScoreCalculate(t *testing.T) {
	arc := new(arcclient.Arc)
	{
		arc.Title = "test"
		arc.Ctime = xtime.Time(time.Now().Add(-time.Hour * 24 * 45).Second())
	}

	stat := new(arcclient.Stat)
	{
		stat.View = 212529
		stat.Reply = 606
		stat.Fav = 11842
		stat.Coin = 4403
		stat.Danmaku = 1495
		stat.Share = 3757
		stat.Like = 4544
	}

	statFromDW := new(model.ArchiveStats)
	{
		statFromDW.ViewBefore14 = 212289
		statFromDW.ReplyBefore14 = 606
		statFromDW.FavoriteBefore14 = 11856
		statFromDW.CoinBefore14 = 4400
		statFromDW.DanmakuBefore14 = 1495
		statFromDW.ShareBefore14 = 3757
		statFromDW.LikeBefore14 = 4537
	}
	statFromDW.Rebuild(stat)
	score := calculateScore(statFromDW, arc.Ctime.Time())
	score, _ = strconv.ParseFloat(fmt.Sprintf("%.1f", score), 64)
	if score != 59.8 {
		t.Errorf("archice score should as %v, but now %v", 59.8, score)
	}
}

func testGenArchiveStatsByFileRaw(t *testing.T) {
	t.Run("raw with 8 field", genArchiveStatsWith8Fields)
	t.Run("raw with 7 field", genArchiveStatsWith7Fields)
}

func genArchiveStatsWith7Fields(t *testing.T) {
	raw := []string{
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
	}

	if _, err := genArchiveStats(raw); err == nil {
		t.Errorf("genArchiveStats >>> err(%v), expected(invalid field count)", err)
	}
}

func genArchiveStatsWith8Fields(t *testing.T) {
	raw := []string{
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
	}

	if stats, err := genArchiveStats(raw); err != nil || stats == nil {
		t.Errorf("genArchiveStats >>> err(%v), expected(%v); stats(%v), expected(%v)", err, nil, stats, "not nil")
	}
}

func testDownloadBiz(t *testing.T) {
	if files := downloadRemoteFiles2Local([]string{"www.baidu.com"}, 1); len(files) <= 0 {
		t.Errorf("downloadRemoteFiles2Local failed, err: files length is not %v", 1)
	}
}

func testInitShellScript(t *testing.T) {
	if err := initShellScript(); err != nil {
		t.Errorf("initShellScript failed, err: %v", err)
	}
}

func testCalculateScore(t *testing.T) {
	t.Run("test archive score which had published 30 days", calculate30Days)
	t.Run("test archive score which had published 1 days", calculate1Days)
}

func calculate30Days(t *testing.T) {
	stats := &model.ArchiveStats{
		AID:              1,
		CoinBefore14:     100,
		CoinIn14:         100,
		DanmakuBefore14:  100,
		DanmakuIn14:      100,
		FavoriteBefore14: 100,
		FavoriteIn14:     100,
		LikeBefore14:     100,
		LikeIn14:         100,
		ReplyBefore14:    100,
		ReplyIn14:        100,
		ShareBefore14:    100,
		ShareIn14:        100,
		ViewBefore14:     100,
		ViewIn14:         100,
	}
	pubTime := time.Now().Add(-time.Hour * 24 * 30)
	if score := calculateScore(stats, pubTime); score != 275 {
		t.Errorf("archice score should as %v, but now %v", 275, score)
	}
}

func calculate1Days(t *testing.T) {
	stats := &model.ArchiveStats{
		AID:              1,
		CoinBefore14:     0,
		CoinIn14:         100,
		DanmakuBefore14:  0,
		DanmakuIn14:      100,
		FavoriteBefore14: 0,
		FavoriteIn14:     100,
		LikeBefore14:     0,
		LikeIn14:         100,
		ReplyBefore14:    0,
		ReplyIn14:        100,
		ShareBefore14:    0,
		ShareIn14:        100,
		ViewBefore14:     0,
		ViewIn14:         100,
	}
	pubTime := time.Now().Add(-time.Hour * 24)
	score := calculateScore(stats, pubTime)
	score, _ = strconv.ParseFloat(fmt.Sprintf("%.1f", score), 64)
	if score != 7.7 {
		t.Errorf("archice score should as %v, but now %v", 7.7, score)
	}
}
