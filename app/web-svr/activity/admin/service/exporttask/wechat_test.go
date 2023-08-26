package exporttask

import (
	"context"
	"fmt"
	bm "go-common/library/net/http/blademaster"
	xtime "go-common/library/time"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	httpClient = bm.NewClient(&bm.ClientConfig{
		App: &bm.App{
			Key:    "53e2fa226f5ad348",
			Secret: "3cf6bd1b0ff671021da5f424fea4b04a",
		},
		Dial:      xtime.Duration(5 * time.Second),
		Timeout:   xtime.Duration(5 * time.Second),
		KeepAlive: xtime.Duration(5 * time.Second),
	})
}

func TestGetMemberInfo(t *testing.T) {
	err := GetMemberInfo()
	assert.Equal(t, nil, err)
	fmt.Println(len(userID))
}

func TestGetMinDeptToFetch(t *testing.T) {
	ctx := context.Background()
	token, err := WeChatAccessToken(ctx)
	assert.Equal(t, nil, err)
	depts, err := GetMinDeptToFetch(ctx, token)
	assert.Equal(t, nil, err)
	fmt.Println(depts)
}
