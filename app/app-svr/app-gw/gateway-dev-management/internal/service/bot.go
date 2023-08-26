package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"

	"github.com/pkg/errors"
	"github.com/sbzhu/weworkapi_golang/wxbizmsgcrypt"
)

const (
	corpid  = "wx0833ac9926284fa5" //企业ID
	agentId = 1000274              //应用ID
	//nolint:gosec
	secret     = "fRSO1_TabdDZiYIs5lP9u5OYc2i1c-EbDQdyZXx4w8s" //Secret
	receiverId = corpid
	//nolint:gosec
	wxToken        = "rgDxXGCl"
	encodingAeskey = "o2sUCF126Tv75J2RWYgslvwaRRrw6ocvllEWLMhkEnL"
	queryGetURL    = "http://hawkeye.bilibili.co/report/api/v2/query/datasource/%v?query=%v"
	cpuQuery       = `job:cpu_used{container_env_app_id=~"%s", container_env_deploy_env="prod"}`
	memoryQuery    = `job:mem_used{container_env_app_id=~"%s", container_env_deploy_env="prod"}`
	httpQPSQuery   = `sum(rate(http_server_requests_duration_ms_count{app=~"%v",env="prod",cluster=~".*"}[2m])) by (path)`
	grpcQPSQuery   = `sum(rate(grpc_server_requests_duration_ms_count{app=~"%v",method!~".*Ping",env="prod",cluster=~".*"}[2m])) by (method)`
	qpsTop         = 5
)

func (s *Service) BotVerify(ctx context.Context, req *model.BotVerifyReq) (string, error) {
	wxcpt := wxbizmsgcrypt.NewWXBizMsgCrypt(wxToken, encodingAeskey, receiverId, wxbizmsgcrypt.XmlType)
	echoStr, cryptErr := wxcpt.VerifyURL(req.MsgSignature, req.Timestamp, req.Nonce, req.Echostr)
	if cryptErr != nil {
		log.Error("%+v", cryptErr)
		return "", errors.New(cryptErr.ErrMsg)
	}
	return string(echoStr), nil
}

func (s *Service) BotCallback(ctx context.Context, req *model.BotCallbackReq, data []byte) (string, error) {
	wxcpt := wxbizmsgcrypt.NewWXBizMsgCrypt(wxToken, encodingAeskey, receiverId, wxbizmsgcrypt.XmlType)
	reqMsgSign := req.MsgSignature
	reqTimestamp := req.Timestamp
	reqNonce := req.Nonce
	reqData := data
	msg, cryptErr := wxcpt.DecryptMsg(reqMsgSign, reqTimestamp, reqNonce, reqData)
	if cryptErr != nil {
		log.Error("%+v", cryptErr)
		return "", errors.New(cryptErr.ErrMsg)
	}
	var msgContent *model.MsgContent
	err := xml.Unmarshal(msg, &msgContent)
	if err != nil {
		log.Error("%+v", err)
		return "", err
	}
	respData, ok := s.BotContent(ctx, msgContent)
	if !ok {
		return "", nil
	}
	msg, err = xml.Marshal(&respData)
	if err != nil {
		log.Error("%+v", err)
		return "", err
	}
	encryptMsg, cryptErr := wxcpt.EncryptMsg(string(msg), reqTimestamp, reqNonce)
	if cryptErr != nil {
		log.Error("%+v", cryptErr)
		return "", errors.New(cryptErr.ErrMsg)
	}
	return string(encryptMsg), nil
}

func (s *Service) BotContent(ctx context.Context, msgContent *model.MsgContent) (*model.WXRepTextMsg, bool) {
	respData := &model.WXRepTextMsg{
		ToUserName:   msgContent.ToUsername,
		FromUserName: msgContent.FromUsername,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      "bot还在建设中...:-)",
	}
	var err error
	content := msgContent.Content
	if msgContent.MsgType == "event" {
		if msgContent.Event == "click" && msgContent.EventKey == "sre" {
			respData.Content, err = s.GetCurrentSre(ctx)
			respData.Content = "本周SRE负责人为：" + respData.Content
			if err != nil {
				respData.Content = "获取sre轮值失败"
				log.Error("%+v", err)
			}
			return respData, true
		}
		return nil, false
	}
	if strings.Contains(content, "工号") {
		respData.Content = msgContent.FromUsername
		return respData, true
	}
	if service := getServiceName(content); service != "" && strings.Contains(content, "moni") {
		respData.Content, err = s.GetAllMoniData(ctx, service)
		if err != nil {
			log.Error("%+v", err)
		}
		return respData, true
	}
	respData.Content = "不能识别发送的内容"
	return respData, true
}

func (s *Service) GetAllMoniData(ctx context.Context, service string) (string, error) {
	var strErr = "查找失败"
	cpu, err := s.GetCPU(ctx, service)
	if err != nil {
		return strErr, errors.WithStack(err)
	}
	memory, err := s.GetMemory(ctx, service)
	if err != nil {
		return strErr, errors.WithStack(err)
	}
	http, err := s.GetHttpQPS(ctx, service)
	if err != nil {
		return strErr, errors.WithStack(err)
	}
	grpc, err := s.GetGrpcQPS(ctx, service)
	if err != nil {
		return strErr, errors.WithStack(err)
	}
	var separation = "--------------------------------------------------------------------\n"
	rst := cpu + separation + memory + separation + http + separation + grpc
	return rst, nil
}

func (s *Service) GetCPU(ctx context.Context, service string) (string, error) {
	moniToken, err := s.ac.Get("moniToken").String()
	if err != nil {
		return "", err
	}
	query := fmt.Sprintf(cpuQuery, service)
	reqURL := fmt.Sprintf(queryGetURL, 6, url.QueryEscape(query))
	headers := s.RuleTokenHeader(moniToken)
	data, err := httpGet(reqURL, headers)
	if err != nil {
		return "", err
	}
	var reply *model.GetKernelReply
	if err = json.Unmarshal(data, &reply); err != nil {
		log.Warn("%s", string(data))
		return "", err
	}
	if reply.Code != 0 {
		return "", errors.New("get cpu usage error")
	}
	rst, err := s.ProcessCPUOrMemoryData(ctx, reply.Data.Result)
	if err != nil {
		return "", err
	}
	rst = "Container CPU usage（已经除以核数)\n" + rst
	return rst, nil
}

func (s *Service) GetMemory(ctx context.Context, service string) (string, error) {
	moniToken, err := s.ac.Get("moniToken").String()
	if err != nil {
		return "", err
	}
	query := fmt.Sprintf(memoryQuery, service)
	reqURL := fmt.Sprintf(queryGetURL, 6, url.QueryEscape(query))
	headers := s.RuleTokenHeader(moniToken)
	data, err := httpGet(reqURL, headers)
	if err != nil {
		return "", err
	}
	var reply *model.GetKernelReply
	if err = json.Unmarshal(data, &reply); err != nil {
		log.Warn(string(data))
		return "", err
	}
	if reply.Code != 0 {
		return "", errors.New("get cpu memory error")
	}
	rst, err := s.ProcessCPUOrMemoryData(ctx, reply.Data.Result)
	if err != nil {
		return "", err
	}
	rst = "Container Memory （减去cache）使用率\n" + rst
	return rst, nil
}

func (s *Service) GetHttpQPS(ctx context.Context, service string) (string, error) {
	moniToken, err := s.ac.Get("moniToken").String()
	if err != nil {
		return "", err
	}
	query := fmt.Sprintf(httpQPSQuery, service)
	reqURL := fmt.Sprintf(queryGetURL, 1, url.QueryEscape(query))
	headers := s.RuleTokenHeader(moniToken)
	data, err := httpGet(reqURL, headers)
	if err != nil {
		return "", err
	}
	var reply *model.GetHTTPQPSReply
	if err = json.Unmarshal(data, &reply); err != nil {
		log.Warn(string(data))
		return "", err
	}
	if reply.Code != 0 {
		return "", errors.New("get http qps error")
	}
	rst, err := s.ProcessHTTPQPSData(ctx, reply.Data.Result)
	if err != nil {
		if rst != "" {
			return rst, nil
		}
		return "", err
	}
	rst = "HTTP服务QPS\n" + rst
	return rst, nil
}

func (s *Service) GetGrpcQPS(ctx context.Context, service string) (string, error) {
	moniToken, err := s.ac.Get("moniToken").String()
	if err != nil {
		return "", err
	}
	query := fmt.Sprintf(grpcQPSQuery, service)
	reqURL := fmt.Sprintf(queryGetURL, 1, url.QueryEscape(query))
	headers := s.RuleTokenHeader(moniToken)
	data, err := httpGet(reqURL, headers)
	if err != nil {
		return "", err
	}
	var reply *model.GetGRPCQPSReply
	if err = json.Unmarshal(data, &reply); err != nil {
		log.Warn(string(data))
		return "", err
	}
	if reply.Code != 0 {
		return "", errors.New("get grpc qps error")
	}
	rst, err := s.ProcessGRPCQPSData(ctx, reply.Data.Result)
	if err != nil {
		if rst != "" {
			return rst, nil
		}
		return "", err
	}
	rst = "GRPC服务QPS\n" + rst
	return rst, nil
}

func (s *Service) ProcessCPUOrMemoryData(ctx context.Context, data []*model.GetKernelResult) (string, error) {
	if len(data) == 0 {
		return "", errors.New("data is empty")
	}
	var info []Info
	for _, d := range data {
		if strings.Contains(d.Metric.ContainerEnvPodContainer, "overlord") {
			continue
		}
		v := d.Value[1].(string)
		current, err := strconv.ParseFloat(v, 32)
		current, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", current), 64)
		if err != nil {
			return "", err
		}
		info = append(info, Info{d.Metric.ContainerEnvPodName, current})
	}
	sort.Slice(info, func(i, j int) bool {
		return info[i].Data < info[j].Data
	})
	max := info[len(info)-1]
	min := info[0]
	avr := getAverage(info)
	avr, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", avr), 64)
	rst := fmt.Sprintf("目前最高:%v,数值%v%%\n", max.Name, max.Data) +
		fmt.Sprintf("目前最低:%v,数值%v%%\n", min.Name, min.Data) +
		fmt.Sprintf("目前平均数值%v%%\n", avr)
	return rst, nil
}

func (s *Service) ProcessHTTPQPSData(ctx context.Context, data []*model.GetHTTPQPSResult) (string, error) {
	var info []Info
	for _, d := range data {
		v := d.Value[1].(string)
		current, err := strconv.ParseFloat(v, 32)
		current, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", current), 64)
		if err != nil {
			return "", err
		}
		info = append(info, Info{d.Metric.Path, current})
	}
	rst := getQPSTop(info)
	return rst, nil
}

func (s *Service) ProcessGRPCQPSData(ctx context.Context, data []*model.GetGRPCQPSResult) (string, error) {
	var info []Info
	for _, d := range data {
		v := d.Value[1].(string)
		current, err := strconv.ParseFloat(v, 32)
		current, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", current), 64)
		if err != nil {
			return "", err
		}
		info = append(info, Info{d.Metric.Method, current})
	}
	rst := getQPSTop(info)
	return rst, nil
}

func getQPSTop(info []Info) string {
	sort.Slice(info, func(i, j int) bool {
		return info[i].Data > info[j].Data
	})
	var top = qpsTop
	if len(info) == 0 {
		return "查询结果为空"
	}
	if len(info) < top {
		top = len(info)
	}
	var rst string
	for i, v := range info {
		rst += fmt.Sprintf("top%d:%v,数值%v\n", i+1, v.Name, v.Data)
		if i == top-1 {
			break
		}
	}
	return rst
}

type Info struct {
	Name string
	Data float64
}

func getAverage(a []Info) float64 {
	var b float64
	for _, value := range a {
		b += value.Data
	}
	return b / float64(len(a))
}

func getServiceName(content string) string {
	reg := regexp.MustCompile(`[a-z|\-]+\.[a-z|\-]+\.[a-z|\-]+`)
	if reg == nil {
		return ""
	}
	rst := reg.FindAllStringSubmatch(content, -1)
	if len(rst) == 0 || len(rst[0]) == 0 {
		return ""
	}
	return rst[0][0]
}

func (s *Service) DashboardVerify(ctx context.Context, session string) (*model.DashboardVerifyReply, error) {
	caller := "gateway-dev-mgt"
	sessionID := session
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	apiKey := "9587a0ee92cd5870ac7fb519361c2abd"

	v := url.Values{}
	v.Set("caller", caller)
	v.Set("session_id", sessionID)
	v.Set("ts", ts)

	sign := Md5(v.Encode() + apiKey)

	resp, err := http.Post(
		"http://dashboard-mng.bilibili.co/api/session/verify",
		"application/x-www-form-urlencoded",
		strings.NewReader("caller="+caller+"&session_id="+sessionID+"&ts="+ts+"&sign="+sign))
	if err != nil {
		log.Error("%+v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("%+v", err)
	}
	var reply *model.DashboardVerifyReply
	if err = json.Unmarshal(body, &reply); err != nil {
		log.Warn(string(body))
		return nil, err
	}
	if reply.Code != 0 {
		return nil, errors.New("verify error")
	}
	return reply, err
}

func Md5(str string) string {
	data := []byte(str)
	md5Ctx := md5.New()
	md5Ctx.Write(data)
	cipherStr := md5Ctx.Sum(nil)

	return hex.EncodeToString(cipherStr)
}
