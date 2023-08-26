package service

import (
	"os"
	"testing"

	"go-gateway/app/app-svr/kvo/job/internal/dao"

	"go-common/library/conf/paladin"
	"go-common/library/net/trace"
)

var (
	svr *Service
)

func TestMain(m *testing.M) {
	var err error
	trace.Init(nil)
	defer trace.Close()
	if paladin.DefaultClient, err = paladin.NewFile("../../cmd/cmd/"); err != nil {
		panic(err)
	}
	redis, err := dao.NewRedis()
	if err != nil {
		panic(err)
	}
	db, err := dao.NewDB()
	if err != nil {
		panic(err)
	}
	daoDao, err := dao.New(redis, db)
	if err != nil {
		panic(err)
	}
	svr, _ = New(daoDao)
	os.Exit(m.Run())
}
