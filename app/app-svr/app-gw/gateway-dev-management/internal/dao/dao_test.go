package dao

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin"
)

var (
	d *dao
)

func ctx() context.Context {
	return context.Background()
}

func TestMain(m *testing.M) {
	flag.Set("conf", "../../configs")
	flag.Parse()
	var err error
	if err = paladin.Init(); err != nil {
		panic(err)
	}
	var cf func()
	if d, cf, err = newTestDao(); err != nil {
		panic(err)
	}
	ret := m.Run()
	cf()
	os.Exit(ret)
}

//func TestInsertServiceConfig(t *testing.T) {
//	s := []model.Stuff{
//		{
//			Id:       "006345",
//			Username: "huyang",
//		},
//		{
//			Id:       "002487",
//			Username: "zhoujiahui",
//		},
//		{
//			Id:       "009123",
//			Username: "ruanshuaishuai",
//		},
//		{
//			Id:       "008599",
//			Username: "dongzhengyi",
//		},
//		{
//			Id:       "000952",
//			Username: "zhangxin",
//		},
//		{
//			Id:       "014883",
//			Username: "lizhenbang",
//		},
//		{
//			Id:       "023043",
//			Username: "xialinjuan01",
//		},
//		{
//			Id:       "015626",
//			Username: "zoujunhui",
//		},
//		{
//			Id:       "021315",
//			Username: "liuyijun02",
//		},
//		{
//			Id:       "000285",
//			Username: "sunyu",
//		},
//		{
//			Id:       "003706",
//			Username: "laiying",
//		},
//	}
//	data, _ := json.Marshal(s)
//	gs := &model.GatewaySchedule{Key: "service_schedule", Value: string(data)}
//	err := d.InsertConfig(ctx(), gs)
//	assert.NoError(t, err)
//}

//func TestInsertSreConfig(t *testing.T) {
//	s := []model.Stuff{
//		{
//			Id:       "015626",
//			Username: "zoujunhui",
//		},
//		{
//			Id:       "021315",
//			Username: "liuyijun02",
//		},
//		{
//			Id:       "000285",
//			Username: "sunyu",
//		},
//		{
//			Id:       "003706",
//			Username: "laiying",
//		},
//		{
//			Id:       "006345",
//			Username: "huyang",
//		},
//		{
//			Id:       "002487",
//			Username: "zhoujiahui",
//		},
//		{
//			Id:       "009123",
//			Username: "ruanshuaishuai",
//		},
//		{
//			Id:       "008599",
//			Username: "dongzhengyi",
//		},
//		{
//			Id:       "000952",
//			Username: "zhangxin",
//		},
//		{
//			Id:       "014883",
//			Username: "lizhenbang",
//		},
//		{
//			Id:       "023043",
//			Username: "xialinjuan01",
//		},
//	}
//	data, _ := json.Marshal(s)
//	gs := &model.GatewaySchedule{Key: "sre_schedule", Value: string(data)}
//	err := d.InsertConfig(ctx(), gs)
//	assert.NoError(t, err)
//}

//func TestInsertArrange(t *testing.T) {
//	s := 0
//	data, _ := json.Marshal(s)
//	gs := &model.GatewaySchedule{Key: "service_arrange", Value: string(data)}
//	err := d.InsertConfig(ctx(), gs)
//	assert.NoError(t, err)
//}
