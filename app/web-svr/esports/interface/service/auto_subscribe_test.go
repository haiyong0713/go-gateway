package service

import (
	"context"
	"fmt"
	"testing"
	xtime "time"

	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/database/sql"
	commonECode "go-common/library/ecode"
	"go-common/library/time"

	"github.com/golang/mock/gomock"

	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/dao"
	"go-gateway/app/web-svr/esports/interface/mock"
	"go-gateway/app/web-svr/esports/interface/model"
)

var (
	srv *Service
)

// go test -v auto_subscribe.go auto_subscribe_test.go favorite.go grpc.go guess.go live.go match.go match_active.go pointdata.go s10.go s10_score_analysis.go s10_tab.go s9.go search.go  service.go
func TestAutoSubscribeBiz(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = time.Duration(10 * xtime.Second)
		cfg.ExecTimeout = time.Duration(10 * xtime.Second)
		cfg.TranTimeout = time.Duration(10 * xtime.Second)
	}

	component.GlobalDB = sql.NewMySQL(cfg)
	if err := component.GlobalDB.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	newCfg := &redis.Config{
		Name:  "local",
		Proto: "tcp",
		Addr:  "127.0.0.1:6379",
		Config: &pool.Config{
			IdleTimeout: time.Duration(10 * xtime.Second),
			Idle:        2,
			Active:      8,
		},
		WriteTimeout: time.Duration(10 * xtime.Second),
		ReadTimeout:  time.Duration(10 * xtime.Second),
		DialTimeout:  time.Duration(10 * xtime.Second),
	}
	globalCfg := new(conf.Config)
	{
		globalCfg.AutoSubCache = newCfg
	}

	component.InitRedis(globalCfg)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	favClient := mock.NewMockFavoriteClient(ctrl)
	favClient.EXPECT().MultiAdd(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	srv = new(Service)
	{
		srv.favClient = favClient
	}

	t.Run("auto sub with invalid auto_sub key", autoSubWithInvalidSubKey)
	t.Run("auto sub with auto_sub key 4 not subscribed case", autoSubWithNoSub)
	t.Run("auto sub with auto_sub key 4 not subscribed and fav case", autoSubWithNoSubAndFav)
	t.Run("auto sub with auto_sub key 4 not subscribed and server error case", autoSubWithFavServerError)
	t.Run("auto sub with auto_sub key 4 fetch auto sub status", fetchAutoSubStatus)
}

func fetchAutoSubStatus(t *testing.T) {
	ctx := context.Background()
	mid := int64(88888)
	seasonID := int64(888888)
	teamID := int64(888888)

	req := new(model.AutoSubRequest)
	{
		req.SeasonID = seasonID
		req.TeamIDList = []int64{teamID, 88888888}
	}

	subKey := dao.GenAutoSubUniqKey(req.SeasonID, teamID)
	autoSubMap[subKey] = []int64{1, 2, 3}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	favClient := mock.NewMockFavoriteClient(ctrl)
	favClient.EXPECT().MultiAdd(gomock.Any(), gomock.Any()).Return(nil, commonECode.ServerErr).AnyTimes()
	tmpSrv := new(Service)
	{
		tmpSrv.favClient = favClient
	}

	delAutoSubCacheKey(t, seasonID, mid, teamID)
	m, err := tmpSrv.AutoSubscribeStatus(ctx, mid, req)
	fmt.Println("AutoSubscribeStatus >>>", m)
	if err != nil {
		t.Error(err)
	}
}

func autoSubWithFavServerError(t *testing.T) {
	ctx := context.Background()
	mid := int64(66666)
	seasonID := int64(888888)
	teamID := int64(888888)

	req := new(model.AutoSubRequest)
	{
		req.SeasonID = seasonID
		req.TeamIDList = []int64{teamID}
	}

	delAutoSubCacheKey(t, seasonID, mid, teamID)
	subKey := dao.GenAutoSubUniqKey(req.SeasonID, teamID)
	autoSubMap[subKey] = []int64{1, 2, 3}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	favClient := mock.NewMockFavoriteClient(ctrl)
	favClient.EXPECT().MultiAdd(gomock.Any(), gomock.Any()).Return(nil, commonECode.ServerErr).AnyTimes()
	tmpSrv := new(Service)
	{
		tmpSrv.favClient = favClient
	}

	if err := tmpSrv.AutoSubscribe(ctx, mid, req); err != commonECode.ServerErr {
		t.Error(err)
	}
}

func autoSubWithInvalidSubKey(t *testing.T) {
	ctx := context.Background()
	req := new(model.AutoSubRequest)
	{
		req.SeasonID = 8888
		req.TeamIDList = []int64{88888}
	}
	if err := srv.AutoSubscribe(ctx, 666, req); err != commonECode.RequestErr {
		t.Error(err)
	}
}

func autoSubWithNoSub(t *testing.T) {
	ctx := context.Background()
	mid := int64(666)

	req := new(model.AutoSubRequest)
	{
		req.SeasonID = 888888
		req.TeamIDList = []int64{888888}
	}

	delAutoSubCacheKey(t, 888888, mid, 888888)
	subKey := dao.GenAutoSubUniqKey(req.SeasonID, 888888)
	autoSubMap[subKey] = make([]int64, 0)

	if err := srv.AutoSubscribe(ctx, mid, req); err != nil {
		t.Error(err)
	}
}

func autoSubWithNoSubAndFav(t *testing.T) {
	ctx := context.Background()
	mid := int64(666666)

	req := new(model.AutoSubRequest)
	{
		req.SeasonID = 888888
		req.TeamIDList = []int64{888888, 88888888}
	}

	delAutoSubCacheKey(t, 888888, mid, 888888)
	subKey := dao.GenAutoSubUniqKey(req.SeasonID, 888888)
	autoSubMap[subKey] = []int64{1, 2, 3}

	if err := srv.AutoSubscribe(ctx, mid, req); err != nil {
		t.Error(err)
	}
}

func delAutoSubCacheKey(t *testing.T, seasonId, mid, teamID int64) {
	if _, err := component.GlobalAutoSubCache.Do(context.Background(), "DEL", genAutoSubCacheKey(seasonId, mid, teamID)); err != nil {
		t.Error(err)
	}
}
