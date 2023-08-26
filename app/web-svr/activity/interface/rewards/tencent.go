package rewards

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/rewards"
	"go-gateway/app/web-svr/activity/interface/tool"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	//腾讯奖品
	rewardTypeTencentGame = "TencentGame"
)

func init() {
	awardsSendFuncMap[rewardTypeTencentGame] = Client.tencentGameAwardSender

	awardsCheckFuncMap[rewardTypeTencentGame] = Client.tencentGameAwardChecker

	awardsConfigMap[rewardTypeTencentGame] = &model.TencentGameConfig{}
}

type tentCentGameAwardParams struct {
	FlowId     string `json:"flowId"`
	SerialCode string `json:"serialCode"`
}

/*
{
    "iRet":0,
    "jData":{
        "data":[
            {
                "errno":0,
                "errmsg":"ok",
                "alias":"reward",
                "cName":"【通用】奖励发放接口",
                "label":"result",
                "data":{
                    "iRet":0,
                    "message":"恭喜您获得了礼包： 刀锋之影 泰隆(7天)+腥红之月 泰隆(7天) , 请注意：游戏虚拟道具奖品将会在24小时内到账",
                    "packageId":"2589772",
                    "packageName":"刀锋之影 泰隆(7天)+腥红之月 泰隆(7天)",
                    "packageNum":"1",
                    "sPackageOtherInfo":"",
                    "sPackageRealFlag":"0"
                }
            }
        ],
        "errno":0,
        "errmsg":"ok",
        "tag":1,
        "flowName":"观看直播时长任务奖励",
        "desc":""
    },
    "sMsg":"ok",
    "tid":"187380792505182383"
}
*/

type tentCentGameAwardResponse struct {
	IRet  int64                          `json:"iRet"`
	SMsg  string                         `json:"sMsg"`
	Tid   string                         `json:"tid"`
	JData *tentCentGameAwardResponseData `json:"jData"`
}

type tentCentGameAwardResponseData struct {
	Data []*tentCentGameAwardResponseInner1Data `json:"data"`
}

type tentCentGameAwardResponseInner1Data struct {
	Data *tentCentGameAwardResponseInner2Data `json:"data"`
}

type tentCentGameAwardResponseInner2Data struct {
	PackageName string `json:"packageName"`
	PackageNum  string `json:"packageNum"`
}

func (s *service) tencentGameAwardChecker(ctx context.Context, c *api.RewardsAwardInfo, mid int64, _, _ string) (err error) {
	config := &model.TencentGameConfig{}
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	bindParams, err := s.bindSvr.GetBindInfo(ctx, config.AccountInfoId, mid, 0)
	if err != nil {
		return
	}

	if bindParams.ConfigInfo.BindType != bindParams.BindInfo.BindType {
		return ecode.ActivityNotBind
	}

	return
}

func (s *service) tencentGameAwardSender(ctx context.Context, c *api.RewardsAwardInfo, mid int64, uniqueID, _ string) (extraInfo map[string]string, err error) {
	config := &model.TencentGameConfig{}
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.tencentGameAwardSender mids(%v) uniqueId(%v) config: %+v, error:%v", mid, uniqueID, c.JsonStr, err)
			return
		}
		s.sendAwardNotifyCard(mid, c, c.NotifyJumpUri1, c.NotifyJumpUri2)
	}()
	if err = json.Unmarshal([]byte(c.JsonStr), &config); err != nil {
		return
	}
	bindParams, err := s.bindSvr.GetBindInfo(ctx, config.AccountInfoId, mid, 0)
	if err != nil {
		return
	}
	gameConfig, err := s.bindSvr.GetGameConfig(ctx, bindParams.ConfigInfo.GameType)
	if err != nil {
		return
	}
	if bindParams.ConfigInfo.BindType != bindParams.BindInfo.BindType {
		err = ecode.ActivityNotBind
		return
	}
	privateParams := &tentCentGameAwardParams{
		FlowId:     config.FlowId,
		SerialCode: uniqueID,
	}
	bs, err := json.Marshal(privateParams)
	if err != nil {
		return
	}

	var code string
	err = retry.WithAttempts(ctx, "tencentGameAwardSender.GetCode", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
		code, err = s.bindDao.GetCode(ctx, gameConfig.ClientId, gameConfig.Business, mid)
		return err
	})
	publicParamsMap := make(map[string]string, 0)
	publicParamsMap["logintype"] = bindParams.BindInfo.AccountInfo.AccountType
	publicParamsMap["livePlatId"] = gameConfig.AppId
	publicParamsMap["actId"] = bindParams.ConfigInfo.ActId
	publicParamsMap["gameId"] = gameConfig.GameName
	publicParamsMap["v"] = gameConfig.Version
	publicParamsMap["code"] = code
	publicParamsMap["nonce"] = tool.RandStringRunes(8)
	publicParamsMap["t"] = strconv.FormatInt(time.Now().Unix(), 10)
	sig := tool.TencentMd5Sign(gameConfig.SignKey, publicParamsMap)
	publicParams := url.Values{}
	for k, v := range publicParamsMap {
		publicParams.Add(k, v)
	}
	publicParams.Add("sig", sig)
	publicParams.Add("apiName", "ApiRequest")
	publicParams.Encode()
	res := &tentCentGameAwardResponse{}
	extraInfo = make(map[string]string)
	var req *http.Request
	var bodyBs []byte
	err = retry.WithAttempts(ctx, "tencentGameAwardSender.HTTP", 3, netutil.DefaultBackoffConfig, func(c context.Context) (err error) {
		req, err = http.NewRequest("GET", fmt.Sprintf("https://open.livelink.qq.com/livelink/?%v", publicParams.Encode()), bytes.NewReader(bs))
		if err != nil {
			log.Errorc(ctx, "rewards.tencentGameAwardSender http.NewRequest error: %v", err)
			return
		}
		bodyBs, err = s.httpClient.Raw(ctx, req)
		if err != nil {
			log.Errorc(ctx, "rewards.tencentGameAwardSender do http call error: %v", err)
			return
		}
		err = json.Unmarshal(bodyBs, res)
		if err != nil {
			log.Errorc(ctx, "rewards.tencentGameAwardSender decode http resp body error: %v", err)
			return
		}

		if res.IRet == -3018 {
			log.Errorc(ctx, "rewards.tencentGameAwardSender: -3018 该订单已存在了")
			err = nil
			return
		}
		if res.IRet != 0 {
			err = fmt.Errorf("response: %+v", res)
			log.Errorc(ctx, "rewards.tencentGameAwardSender res.IRet error: %v", err)
		}
		return

	})
	if err != nil {
		log.Errorc(ctx, "rewards.tencentGameAwardSender req url(%v), body (%v) resp(%v), error: %v", req.URL, string(bs), res, err)
	} else {
		log.Infoc(ctx, "rewards.tencentGameAwardSender success req url(%v), body (%v) resp(%v), error: %v", req.URL, string(bs), string(bodyBs), err)
		if res != nil && res.JData != nil && len(res.JData.Data) != 0 &&
			res.JData.Data[0] != nil && &res.JData.Data[0].Data != nil {
			extraInfo["tencent_package_name"] = res.JData.Data[0].Data.PackageName
			extraInfo["tencent_package_num"] = res.JData.Data[0].Data.PackageNum
		}
	}
	return
}

func (s *service) GetParentUniqueId(ctx context.Context, awardId int64, uniqueId string) (parentUniqueId string, err error) {
	//目前不支持礼包, 直接返回原有uniqueId即可
	parentUniqueId = uniqueId
	err = nil
	return
}

func (s *service) GetTencentAwardAccountId(ctx context.Context, awardId int64) (accountInfoId int64, err error) {
	awardConfig, err := s.GetAwardConfigById(ctx, awardId)
	if err != nil {
		return
	}
	if awardConfig.Type != rewardTypeTencentGame {
		err = fmt.Errorf("award is not rewardTypeTencentGame")
	}
	config := &model.TencentGameConfig{}
	if err = json.Unmarshal([]byte(awardConfig.JsonStr), &config); err != nil {
		return
	}
	accountInfoId = config.AccountInfoId
	return
}
