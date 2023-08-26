package service

import (
	"flag"
	"fmt"
	"go-common/library/conf/paladin"
	"os"
	"testing"
)

var testService *Service

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
	testService, _, err = InitService()
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
