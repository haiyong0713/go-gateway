package rewards

import (
	"encoding/json"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v . -count 1 -run TestConfigValidate
func TestConfigValidate(t *testing.T) {
	c1 := &model.MallCouponConfig{
		SourceId:         1,
		CouponId:         "testId",
		SourceActivityID: "activity",
	}
	bs, _ := json.Marshal(c1)
	err := Client.validateJsonStr("MallCoupon", string(bs))
	assert.Equal(t, nil, err)
	c2 := &model.MallCouponConfig{
		SourceId:         0,
		CouponId:         "testId",
		SourceActivityID: "activity",
	}
	bs, _ = json.Marshal(c2)
	err = Client.validateJsonStr("MallCoupon", string(bs))
	assert.NotEqual(t, nil, err)

}
