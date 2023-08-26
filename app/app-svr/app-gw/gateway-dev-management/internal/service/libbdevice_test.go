package service

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"

	"github.com/stretchr/testify/assert"
)

func Test_CheckExpressionWithDevice(t *testing.T) {
	s := &Service{}
	req := &model.CheckExpressionReq{
		MobiApp:    "iphone",
		Device:     "phone",
		Platform:   "",
		Build:      61400000,
		Expression: `mobi_app == "iphone" && build == 61400000`,
	}
	rst, err := s.CheckExpressionWithDevice(context.TODO(), req)
	assert.NoError(t, err)
	assert.Equal(t, "true", rst.Result)

	req = &model.CheckExpressionReq{
		MobiApp:    "iphone",
		Device:     "phone",
		Platform:   "",
		Build:      61400000,
		Expression: `mobi_app == "android" && build == 61400000`,
	}
	rst, err = s.CheckExpressionWithDevice(context.TODO(), req)
	assert.NoError(t, err)
	assert.Equal(t, "false", rst.Result)
}
