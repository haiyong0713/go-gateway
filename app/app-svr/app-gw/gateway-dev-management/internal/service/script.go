package service

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

const (
	//nolint:gosec
	tokenGetURL   = "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%v&corpsecret=%v"
	sendMsgURL    = "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%v"
	hrCoreAuthURL = "https://hrcore.bilibili.co/auth-info/v1/jwt?destClient=ops.ehr-api.hr-core&client=main.app-svr.gateway-dev-management&secret=x0HkDxvAgZ666v_QBel1ImY9ohzOHpdBulaUaChhPvc="
	hrCoreInfoURL = "https://hrcore.bilibili.co/hrcore/open/user/keyword?adAccount=%v"
)

func (s *Service) NewScript(ctx context.Context, req *model.NewScriptReq) error {
	var (
		data []byte
		err  error
	)
	script := &model.Script{
		UserName: req.UserName,
		Type:     req.Type,
		APP:      req.App,
	}
	if req.Type == "restart" {
		parameter := &model.RestartParam{
			Zone: req.Zone,
		}
		data, err = json.Marshal(parameter)
		if err != nil {
			log.Error("%+v", err)
			return err
		}
	}
	script.Parameter = string(data)
	if err = s.dao.InsertScript(ctx, script); err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}

func (s *Service) GetScript(ctx context.Context, userID string) (*model.GetScriptReply, error) {
	scripts, err := s.dao.GetUserScript(ctx, userID)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	reply := &model.GetScriptReply{
		UserID:  userID,
		Scripts: scripts,
	}
	return reply, nil
}

func (s *Service) DoScript(ctx context.Context, id string, cookie string) (string, error) {
	script, err := s.dao.GetScript(ctx, id)
	if err != nil {
		log.Error("%+v", err)
		return "", err
	}
	if script.Type == "restart" {
		var param *model.RestartParam
		if err = json.Unmarshal([]byte(script.Parameter), &param); err != nil {
			log.Error("%+v", err)
			return "", err
		}
		token, err := s.GetAccessToken(ctx)
		if err != nil {
			log.Error("%+v", err)
			return "", err
		}
		userID, err := s.GetUserIDByName(ctx, script.UserName)
		if err != nil {
			log.Error("%+v", err)
			return "", err
		}
		content, err := s.Restart(ctx, script.APP, param.Zone, cookie)
		if err != nil {
			log.Error("%+v", err)
			return "", err
		}
		msgReq := &model.SendCardReq{
			Touser:  userID,
			Msgtype: "textcard",
			Agentid: agentId,
		}
		msgReq.Textcard.Title = "重启发布单"
		msgReq.Textcard.Description = "创建重启发布单成功，点击链接继续操作"
		msgReq.Textcard.Url = content
		if _, err = httpPost(fmt.Sprintf(sendMsgURL, token), msgReq, nil); err != nil {
			log.Error("%+v", err)
			return "", err
		}
		return "restart", nil
	}
	return "", nil
}

func (s *Service) GetAccessToken(ctx context.Context) (string, error) {
	var tokenReply *model.GetTokenReply
	data, err := httpGet(fmt.Sprintf(tokenGetURL, corpid, secret), nil)
	if err != nil {
		return "", err
	}
	if err = json.Unmarshal(data, &tokenReply); err != nil {
		return "", err
	}
	if tokenReply.Errcode != 0 {
		return "", err
	}
	token := tokenReply.AccessToken
	return token, nil
}

func (s *Service) GetScriptURL(ctx context.Context, userid string) (string, error) {
	var hmacSampleSecret []byte
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userid,
	})
	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(hmacSampleSecret)
	if err != nil {
		return "", err
	}
	release, err := s.ac.Get("scriptURL").String()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf(release, tokenString)
	return url, nil
}

func (s *Service) GetHRToken(ctx context.Context) (string, error) {
	data, err := httpGet(hrCoreAuthURL, nil)
	if err != nil {
		return "", err
	}
	var reply *model.HRCoreAuthReply
	if err = json.Unmarshal(data, &reply); err != nil {
		log.Warn("%s", string(data))
		return "", err
	}
	if reply.Code != 0 {
		return "", errors.New("get hrcore error")
	}
	return reply.Data.Token, nil
}

func (s *Service) GetUserIDByName(ctx context.Context, userName string) (string, error) {
	token, err := s.GetHRToken(ctx)
	if err != nil {
		return "", err
	}
	headers := make(map[string]string)
	headers["Authorization"] = token
	data, err := httpGet(fmt.Sprintf(hrCoreInfoURL, userName), headers)
	if err != nil {
		return "", err
	}
	var reply *model.HRCoreINFOReply
	if err = json.Unmarshal(data, &reply); err != nil {
		log.Warn("%s", string(data))
		return "", err
	}
	if reply.Code != 0 {
		return "", errors.Errorf("get userid error:%+v", reply.Message)
	}
	if len(reply.Data) == 0 {
		return "", errors.New("can not find user")
	}
	return reply.Data[0].WxAccount, nil
}
