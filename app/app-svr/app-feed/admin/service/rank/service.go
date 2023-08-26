package rank

import (
	"strconv"
	"strings"
	"time"

	bm "go-common/library/net/http/blademaster"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	actclient "git.bilibili.co/bapis/bapis-go/activity/service"
	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/rank"
	rankModel "go-gateway/app/app-svr/app-feed/admin/model/rank"
)

// Service is rank service
type Service struct {
	c         *conf.Config
	dao       *rank.Dao
	arcClient arcgrpc.ArchiveClient
	actClient actclient.ActivityClient
	// account grpc
	accClient account.AccountClient
	Client    *bm.Client
}

// New new a rank service
func New(c *conf.Config) (s *Service) {
	var (
		err error
	)
	s = &Service{
		c:      c,
		dao:    rank.New(c),
		Client: bm.NewClient(c.HTTPClient.Read),
	}
	if s.arcClient, err = arcgrpc.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.actClient, err = actclient.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.accClient, err = account.NewClient(c.AccountGRPC); err != nil {
		panic(err)
	}

	// 定时任务，每天凌晨计算出所有给数据平台用的原始视频列表
	//nolint:biligowordcheck
	go s.JobGenDataplatAvRank()

	// 定时检查当前所有榜单配置，是否要继续流转状态
	//nolint:biligowordcheck
	go s.JobUpdateRankState()

	return
}

func genLogDate() (logDate string) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	logDate = strings.Replace(yesterday, "-", "", 2)
	return
}

func string2Int64Array(in string) (out []int64) {
	if in == "" {
		return
	}
	for _, s := range strings.Split(in, ",") {
		if s != "" {
			num, _ := strconv.ParseInt(s, 10, 64)
			out = append(out, num)
		}
	}
	return
}

func hiveString2Int64Array(in string) (out []int64) {
	if in == "" {
		return
	}
	for _, s := range strings.Split(in, "\u0001") {
		if s != "" {
			num, _ := strconv.ParseInt(s, 10, 64)
			out = append(out, num)
		}
	}
	return
}

type RankSortList []rankModel.RankDetailAVItem

// Len() 排序用
func (s RankSortList) Len() int {
	return len(s)
}

// Less() 排序用
func (s RankSortList) Less(i, j int) bool {
	return s[i].Score.Total > s[j].Score.Total
}

// Swap() 排序用
func (s RankSortList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
