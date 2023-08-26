package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

const (
	statusURL     = "http://caster.bilibili.co/api/v1/applications/status?bu=%v&project=%v&app=%v&env=prod&page_size=10&page_num=1"
	calcURL       = "http://caster.bilibili.co/api/v1/deployment/available_resource/calc"
	deployURL     = "http://caster.bilibili.co/api/v1/deployments"
	startURL      = "http://caster.bilibili.co/api/v1/deployment/%v/start"
	resumeURL     = "http://caster.bilibili.co/api/v1/deployment/%v/resume"
	doneURL       = "http://caster.bilibili.co/api/v1/deployment/%v/done"
	rollbackURL   = "http://caster.bilibili.co/api/v1/deployment/%v/rollback"
	infoURL       = "http://caster.bilibili.co/api/v1/deployment/%v"
	revisionURL   = "http://caster.bilibili.co/api/v1/application/%v/revision?cluster_id=%v&revision_id=%v"
	casterAuthURL = "http://caster.bilibili.co/api/v1/auth"
)

func (s *Service) CasterAuth(ctx context.Context, cookie string) (string, error) {
	var reply *model.CasterAuthReply
	headers := s.RuleCookieHeader(cookie)
	data, err := httpGet(casterAuthURL, headers)
	if err != nil {
		return "", err
	}
	if err = json.Unmarshal(data, &reply); err != nil {
		return "", err
	}
	if reply.Status != http.StatusOK {
		return "", errors.New(reply.Message)
	}
	return reply.Data.Token, nil
}

func (s *Service) GetStatus(ctx context.Context, service string, token string) (*model.DeploymentDetail, error) {
	var reply *model.GetStatusReply
	headers := s.XTokenHeader(token)
	serviceSplit := strings.Split(service, ".")
	//nolint:gomnd
	if len(serviceSplit) != 3 {
		return nil, errors.New("wrong service format")
	}
	data, err := httpGet(fmt.Sprintf(statusURL, serviceSplit[0], serviceSplit[1], serviceSplit[2]), headers)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &reply); err != nil {
		return nil, err
	}
	if reply.Status != http.StatusOK {
		return nil, errors.New(reply.Message)
	}
	if len(reply.Data.Items) != 1 {
		return nil, errors.New("wrong service name")
	}
	return reply.Data.Items[0], nil
}

func (s *Service) Calc(ctx context.Context, info *model.DeploymentDetail, zone string, token string) (*model.AvailableResource, error) {
	var (
		reply    *model.CalcReply
		resource *model.CalcResource
	)
	if err := json.Unmarshal([]byte(info.ApplicationClusters[0].Resource), &resource); err != nil {
		return nil, err
	}
	var cluster *model.DeploymentCluster
	for i, cst := range info.ApplicationClusters {
		if cst.Cluster.Zone == zone {
			cluster = info.ApplicationClusters[i]
		}
	}
	if cluster == nil {
		return nil, errors.New("get cluster error")
	}
	req := &model.CalcReq{
		ClusterId:        cluster.Cluster.ID,
		ClusterName:      cluster.Cluster.Name,
		Constraints:      "",
		CpuPolicy:        cluster.CPUPolicy,
		ResourcePoolName: cluster.ResourcePool.Name,
		OverCommitFactor: cluster.OverCommitFactor,
		ResourceLimitId:  cluster.ResourceLimit.ID,
		CpuReq:           resource.CPUReq,
		CpuLimit:         resource.CPULimit,
		MemReq:           resource.MemReq,
		MemLimit:         resource.MemLimit,
		EstorageReq:      resource.EstorageReq,
		EstorageLimit:    resource.EstorageLimit,
	}
	headers := s.XTokenHeader(token)
	data, err := httpPost(calcURL, req, headers)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &reply); err != nil {
		return nil, err
	}
	if reply.Status != http.StatusOK {
		return nil, errors.New(reply.Message)
	}
	rst := &model.AvailableResource{}
	if err = copier.Copy(rst, req); err != nil {
		return nil, err
	}
	rst.AvailableResource = reply.Data
	return rst, nil
}

func (s *Service) Deploy(ctx context.Context, info *model.DeploymentDetail, resource *model.AvailableResource, zone string, token string) (*model.DeploymentReply, error) {
	var reply *model.DeploymentReply
	if len(info.ApplicationClusters) == 0 {
		return nil, errors.New("get info error")
	}
	var cluster *model.DeploymentCluster
	for i, cst := range info.ApplicationClusters {
		if cst.Cluster.Zone == zone {
			cluster = info.ApplicationClusters[i]
		}
	}
	if cluster == nil {
		return nil, errors.New("get cluster error")
	}
	req := &model.DeploymentReq{
		App:                             info.Name,
		AppName:                         info.Name,
		ApplicationID:                   info.ID,
		AutoPausePoint:                  true,
		AvailableResources:              resource,
		BatchSize:                       1,
		BatchTimeout:                    130,
		Cluster:                         cluster.Cluster.ID,
		ClusterID:                       cluster.Cluster.ID,
		ConfigurationPlatform:           cluster.CurrentPodTemplate.ConfigurationPlatform,
		ConfigurationPlatformBuild:      cluster.CurrentPodTemplate.ConfigurationPlatform,
		ConfigurationPlatformBuildInput: cluster.CurrentPodTemplate.ConfigurationPlatform,
		ConfigurationPlatformEnv:        cluster.CurrentPodTemplate.ConfigurationPlatformEnv,
		EnvInfo:                         "",
		Image:                           cluster.CurrentPodTemplate.Image,
		ImageBefore:                     cluster.CurrentPodTemplate.Cluster.DefaultRegistry,
		ImageTagType:                    "retag",
		ShowConfig:                      false,
		Strategy:                        "rolling",
		Type:                            "restart",
		Version:                         cluster.CurrentPodTemplate.Version,
		Versions:                        "",
	}
	headers := s.XTokenHeader(token)
	data, err := httpPost(deployURL, req, headers)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &reply); err != nil {
		return nil, err
	}
	if reply.Status != http.StatusOK {
		return nil, errors.New(reply.Message)
	}
	return reply, nil
}

func (s *Service) Restart(ctx context.Context, service string, cluster string, cookie string) (string, error) {
	release, err := s.ac.Get("releaseURL").String()
	if err != nil {
		return "", err
	}
	cToken, err := s.CasterAuth(ctx, cookie)
	if err != nil {
		return "", err
	}
	info, err := s.GetStatus(ctx, service, cToken)
	if err != nil {
		return "", errors.WithStack(err)
	}
	resource, err := s.Calc(ctx, info, cluster, cToken)
	if err != nil {
		return "", errors.WithStack(err)
	}
	reply, err := s.Deploy(ctx, info, resource, cluster, cToken)
	if err != nil {
		return "", errors.WithStack(err)
	}
	var hmacSampleSecret []byte
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"deployID": reply.Data.ID,
	})
	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(hmacSampleSecret)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(release, tokenString), nil
}

func (s *Service) StartDeploy(ctx context.Context, deployID int64, cookie string) (string, error) {
	var reply *model.StartReply
	cToken, err := s.CasterAuth(ctx, cookie)
	if err != nil {
		log.Error("%+v", err)
		return "", err
	}
	headers := s.XTokenHeader(cToken)
	data, err := httpPut(fmt.Sprintf(startURL, deployID), nil, headers)
	if err != nil {
		log.Error("%+v", err)
		return "", err
	}
	if err = json.Unmarshal(data, &reply); err != nil {
		log.Error("%+v", err)
		return "", err
	}
	if reply.Data.ID != deployID {
		log.Error("%+v", err)
		return "", errors.New("start deployment error")
	}
	return strconv.FormatInt(deployID, 10), nil
}

func (s *Service) ResumeDeploy(ctx context.Context, deployID int64, cookie string) (string, error) {
	var reply *model.ResumeReply
	cToken, err := s.CasterAuth(ctx, cookie)
	if err != nil {
		log.Error("%+v", err)
		return "", err
	}
	headers := s.XTokenHeader(cToken)
	data, err := httpPut(fmt.Sprintf(resumeURL, deployID), nil, headers)
	if err != nil {
		log.Error("%+v", err)
		return "", err
	}
	if err = json.Unmarshal(data, &reply); err != nil {
		log.Error("%+v", err)
		return "", err
	}
	if reply.Status != http.StatusOK {
		log.Error("%+v", reply.Message)
		return "", errors.New(reply.Message)
	}
	return strconv.FormatInt(deployID, 10), nil
}

func (s *Service) DoneDeploy(ctx context.Context, deployID int64, cookie string) (string, error) {
	var reply *model.ResumeReply
	cToken, err := s.CasterAuth(ctx, cookie)
	if err != nil {
		log.Error("%+v", err)
		return "", err
	}
	headers := s.XTokenHeader(cToken)
	data, err := httpPut(fmt.Sprintf(doneURL, deployID), nil, headers)
	if err != nil {
		log.Error("%+v", err)
		return "", err
	}
	if err = json.Unmarshal(data, &reply); err != nil {
		log.Error("%+v", err)
		return "", err
	}
	if reply.Status != http.StatusOK {
		log.Error("%+v", reply.Message)
		return "", errors.New(reply.Message)
	}
	return strconv.FormatInt(deployID, 10), nil
}

func (s *Service) RollbackDeploy(ctx context.Context, deployID int64, cookie string) (string, error) {
	var reply *model.ResumeReply
	cToken, err := s.CasterAuth(ctx, cookie)
	if err != nil {
		log.Error("%+v", err)
		return "", err
	}
	headers := s.XTokenHeader(cToken)
	data, err := httpPut(fmt.Sprintf(rollbackURL, deployID), nil, headers)
	if err != nil {
		log.Error("%+v", err)
		return "", err
	}
	if err = json.Unmarshal(data, &reply); err != nil {
		log.Error("%+v", err)
		return "", err
	}
	if reply.Status != http.StatusOK {
		log.Error("%+v", reply.Message)
		return "", errors.New(reply.Message)
	}
	return strconv.FormatInt(deployID, 10), nil
}

func (s *Service) GetRevision(ctx context.Context, req *model.GetRevisionReq, token string) (*model.GetRevisionReply, error) {
	var reply *model.GetRevisionReply
	headers := s.XTokenHeader(token)
	data, err := httpGet(fmt.Sprintf(revisionURL, req.AppID, req.ClusterID, req.Revision), headers)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &reply); err != nil {
		return nil, err
	}
	if reply.Status != http.StatusOK {
		return nil, errors.New(reply.Message)
	}
	return reply, nil
}

func (s *Service) GetDeploy(ctx context.Context, tokenString string, cookie string) (*model.DeployStatus, error) {
	var (
		reply *model.GetDeployReply
		info  *model.DeployStatus
	)
	var hmacSampleSecret []byte
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error("Unexpected signing method: %+v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return hmacSampleSecret, nil
	})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Error("%+v", err)
		return nil, err
	}
	deployID := claims["deployID"]
	cToken, err := s.CasterAuth(ctx, cookie)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	headers := s.XTokenHeader(cToken)
	data, err := httpGet(fmt.Sprintf(infoURL, deployID), headers)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if err = json.Unmarshal(data, &reply); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if reply.Status != http.StatusOK {
		log.Error("%+v", reply.Message)
		return nil, errors.New(reply.Message)
	}
	appID := reply.Data.Application.ID
	lastRevision := reply.Data.LastRevision
	revision := reply.Data.Revision
	cluterID := reply.Data.PodTemplate.Cluster.ID
	instances := reply.Data.PodTemplate.Instances
	info = s.GetDeployStage(reply)
	lastRevisionReq := &model.GetRevisionReq{
		AppID:     appID,
		ClusterID: cluterID,
		Revision:  lastRevision,
	}
	lastRevisionReply, err := s.GetRevision(ctx, lastRevisionReq, cToken)
	if err != nil || lastRevisionReply == nil || lastRevisionReply.Data == nil {
		info.LastReplica = "暂无数据"
	} else {
		info.LastReplica = fmt.Sprintf("%v/%v", lastRevisionReply.Data.Status.ReadyReplicas, instances)
	}
	revisionReq := &model.GetRevisionReq{
		AppID:     appID,
		ClusterID: cluterID,
		Revision:  revision,
	}
	revisionReply, err := s.GetRevision(ctx, revisionReq, cToken)
	if err != nil || revisionReply == nil || revisionReply.Data == nil {
		info.Replica = "暂无数据"
	} else {
		info.Replica = fmt.Sprintf("%v/%v", revisionReply.Data.Status.ReadyReplicas, instances)
	}
	return info, nil
}

func (s *Service) GetDeployStage(reply *model.GetDeployReply) *model.DeployStatus {
	var info *model.DeployStatus
	actions := reply.Data.Actions
	info = &model.DeployStatus{
		Id:       reply.Data.ID,
		Service:  reply.Data.Application.Name,
		Zone:     reply.Data.PodTemplate.Cluster.Zone,
		Version:  reply.Data.PodTemplate.Version,
		Current:  "未开始",
		Action:   "",
		Percent:  "0%",
		Start:    true,
		Rollback: false,
		Next:     false,
		Done:     false,
	}
	if reply.Data.Status == "finished" {
		info.Current = "已结单"
		info.Percent = "100%"
		info.Start = false
		info.Rollback = false
		info.Next = false
		info.Done = false
		return info
	}
	for _, action := range actions {
		var ok bool
		info, ok = s.judgeActionStatus(action, info)
		if ok {
			return info
		}
	}
	return info
}

func (s *Service) judgeActionStatus(action *model.Action, info *model.DeployStatus) (*model.DeployStatus, bool) {
	if action.Name == "rolling-rollback" && action.Status == "rolling" {
		info.Current = "回滚中"
		info.Percent = "100%"
		info.Start = false
		info.Rollback = false
		info.Next = false
		info.Done = false
		return info, true
	}
	if action.Name == "rolling-rollback" && action.Status == "finished" {
		info.Current = "回滚完成"
		info.Percent = "100%"
		info.Start = false
		info.Rollback = false
		info.Next = false
		info.Done = false
		return info, true
	}
	if action.Name == "pre-staging" && action.Status == "init" {
		info.Current = "初始化完成"
		info.Percent = "20%"
		info.Start = false
		info.Rollback = false
		info.Next = true
		info.Done = false
		return info, true
	}
	if action.Name == "pre-staging" && action.Status == "rolling" {
		info.Current = "单台中"
		info.Percent = "20%"
		info.Start = false
		info.Rollback = false
		info.Next = false
		info.Done = false
		return info, true
	}
	if action.Name == "pre-staging" && action.Status == "finished" {
		info.Current = "单台完成"
		info.Percent = "40%"
		info.Start = false
		info.Rollback = true
		info.Next = true
		info.Done = false
		return info, false
	}
	if action.Name == "canary" && action.Status == "rolling" {
		info.Current = "单台灰度中"
		info.Percent = "40%"
		info.Start = false
		info.Rollback = false
		info.Next = false
		info.Done = false
		return info, true
	}
	if action.Name == "canary" && action.Status == "finished" {
		info.Current = "单台灰度完成"
		info.Percent = "60%"
		info.Start = false
		info.Rollback = true
		info.Next = true
		info.Done = false
		return info, false
	}
	if action.Name == "rolling-update" && action.Status == "rolling" {
		info.Current = "滚动更新中"
		info.Percent = "60%"
		info.Start = false
		info.Rollback = false
		info.Next = false
		info.Done = false
		return info, true
	}
	if action.Name == "rolling-update" && action.Status == "finished" {
		info.Current = "滚动更新完成"
		info.Percent = "80%"
		info.Start = false
		info.Rollback = true
		info.Next = false
		info.Done = true
		return info, false
	}
	return info, true
}
