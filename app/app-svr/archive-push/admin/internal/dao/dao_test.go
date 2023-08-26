package dao

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"go-common/library/conf/paladin"
)

var testD *Dao

func TestMain(m *testing.M) {
	flag.Set("conf", "../../configs")
	flag.Set("deploy_env", "uat")
	flag.Parse()
	//disableLich := os.Getenv("DISABLE_LICH") != ""
	//if !disableLich {
	//	if err := lich.Setup(); err != nil {
	//		panic(err)
	//	}
	//}
	var err error
	if err = paladin.Init(); err != nil {
		panic(err)
	}
	r, _, _ := NewRedis()
	db, _, _ := NewDB()
	orm, _, _ := NewORM()
	bmClient, _, _ := NewBMClient()
	httpClient := NewHTTPClient()
	archiveGRPCClient, _, _ := NewArchiveGRPC()
	tagGRPCClient, _, _ := NewTagGRPC()
	accountGRPCClient, _, _ := NewAccountGRPC()
	activityGRPCClient, _, _ := NewActivityGRPC()
	testD, _, err = New(r, db, orm, bmClient, httpClient, archiveGRPCClient, tagGRPCClient, accountGRPCClient, activityGRPCClient)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	ret := m.Run()
	//if !disableLich {
	//	_ = lich.Teardown()
	//}
	os.Exit(ret)
}
