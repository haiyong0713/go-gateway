package archive

import (
	"net/url"
	"strconv"
	"testing"
	"time"
)

func Test_flowControlSign(t *testing.T) {
	var (
		secret = "sfj992htcg4oomdrbn5y9fgetrrk1tfqekiobjyz" //找服务方下发
		source = "bilibili_car"                             //业务接入约定

		oid = int64(320002280)
		//oids       = []int64{440110738}
		businessID = 1
		ts         = strconv.FormatInt(time.Now().Unix(), 10)
	)

	params := url.Values{}
	params.Set("source", source)
	params.Set("oid", strconv.FormatInt(oid, 10)) //单个

	//params.Set("oids", xstr.JoinInts(oids)) //批量

	params.Set("business_id", strconv.Itoa(businessID))
	params.Set("ts", ts)

	t.Run("Test_flowControlSign", func(t *testing.T) {
		t.Logf("generate ts %s sign %s", ts, flowControlSign(params, secret))
	})
}
