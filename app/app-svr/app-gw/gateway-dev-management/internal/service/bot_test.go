package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetServiceName(t *testing.T) {
	rst := getServiceName("1+main.app-svr.app-feed")
	assert.Equal(t, "main.app-svr.app-feed", rst)
}
