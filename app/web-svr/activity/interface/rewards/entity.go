package rewards

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"

	"github.com/pkg/errors"
)

const (
	//实物奖励, 发送填写地址私信,用户填写地址后运营统一发货
	rewardTypeEntity = "Entity"
	getAddressURL    = "/api/basecenter/addr/view"
)

func init() {
	//实物奖励, 发送填写地址私信,用户填写地址后运营统一发货
	awardsSendFuncMap[rewardTypeEntity] = Client.entitySender
	awardsConfigMap[rewardTypeEntity] = &model.EmptyConfig{}
}

// 获取用户收货地址信息
func (s *service) getMemberAddress(c context.Context, id, mid int64) (val *model.AddressInfo, err error) {
	var res struct {
		Errno int                `json:"errno"`
		Msg   string             `json:"msg"`
		Data  *model.AddressInfo `json:"data"`
	}
	params := url.Values{}
	params.Set("app_id", s.c.Lottery.AppKey)
	params.Set("app_token", s.c.Lottery.AppToken)
	params.Set("id", strconv.FormatInt(id, 10))
	params.Set("uid", strconv.FormatInt(mid, 10))
	if err = s.httpClient.Get(c, s.c.Host.ShowCo+getAddressURL, "", params, &res); err != nil {
		log.Errorc(c, "getMemberAddress:dao.client.Get id(%d) mid(%d) error(%v)", id, mid, err)
		return
	}
	if res.Errno != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Errno), s.c.Host.ShowCo+getAddressURL+"?"+params.Encode())
	}
	val = res.Data
	return
}

// 不实际发放奖励. 可配合extraInfo返回自定义信息
func (s *service) entitySender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, business string) (extraInfo map[string]string, err error) {
	addressUri := strings.Replace(c.NotifyJumpUri2, "{{ACTIVITY_ID}}", strconv.FormatInt(c.ActivityId, 10), -1)
	s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, addressUri)
	return
}
