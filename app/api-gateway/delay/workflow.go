package delay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
)

const (
	_success              = 200
	_defaultNamespace     = "default"
	_createWorkflowURL    = "/api/v1/wft"           //创建wf
	_getWorkflowStatusURL = "/api/v1/wf/status"     //获取整个wf的状态
	_outputWorkflowURL    = "/api/v1/wf/outputs"    //单个节点的状态
	_resumeWorkflowURL    = "/api/v1/wf/resume"     //继续wf
	_logWorkflowURL       = "/api/v1/wf/archivelog" //单个节点的日志信息
	_stopWorkflowURL      = "/api/v1/wf/stop"       //停止整个wf的状态

)

func (d *dao) createWorkflowURI() string {
	return fmt.Sprintf("%s%s", d.host.Workflow, _createWorkflowURL)
}

func (d *dao) getWorkflowStatusURI() string {
	return fmt.Sprintf("%s%s", d.host.Workflow, _getWorkflowStatusURL)
}

func (d *dao) outputWorkflowURI() string {
	return fmt.Sprintf("%s%s", d.host.Workflow, _outputWorkflowURL)
}

func (d *dao) resumeWorkflowURI() string {
	return fmt.Sprintf("%s%s", d.host.Workflow, _resumeWorkflowURL)
}

func (d *dao) logWorkflowURI() string {
	return fmt.Sprintf("%s%s", d.host.Workflow, _logWorkflowURL)
}

func (d *dao) stopWorkflowURI() string {
	return fmt.Sprintf("%s%s", d.host.Workflow, _stopWorkflowURL)
}

func (d *dao) CreateWorkflow(c context.Context, apiName, codeAddress, codeVersion, imageAddr string, dt DeployType) (name, url string, err error) {
	var (
		res = struct {
			Code    int    `json:"status"`
			Message string `json:"message"`
			Data    struct {
				NameSpace string `json:"namespace"`
				Name      string `json:"name"`
				Url       string `json:"url"`
			} `json:"data"`
		}{}
		msg []byte
		req *http.Request
	)
	params := &CreateWorkflowParams{
		NameSpace:    _defaultNamespace,
		ResourceName: "test-api-gateway",
		EntryPoint:   string(dt),
	}
	params.Parameters = append(params.Parameters, fmt.Sprintf("name=%s", apiName))
	params.Parameters = append(params.Parameters, fmt.Sprintf("code_address=%s", codeAddress))
	params.Parameters = append(params.Parameters, fmt.Sprintf("code_version=%s", codeVersion))
	if dt == OnlyDeploy {
		params.Parameters = append(params.Parameters, fmt.Sprintf("image=%s", imageAddr))
	}
	if msg, err = json.Marshal(params); err != nil {
		return
	}
	if req, err = http.NewRequest(http.MethodPost, d.createWorkflowURI(), bytes.NewBuffer(msg)); err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err = d.httpCli.Do(c, req, &res); err != nil {
		log.Error("d.CreateWorkflow params(%s) url(%s) error(%v)", msg, d.createWorkflowURI(), err)
		return
	}
	if res.Code != _success {
		err = ecode.New(res.Code)
		log.Errorc(c, "d.CreateWorkflow params(%s) res(%+v)", msg, res)
		return
	}
	if res.Data.Name == "" {
		err = ecode.NothingFound
		log.Errorc(c, "d.CreateWorkflow params(%s) res(%+v)", msg, res)
		return
	}
	name = res.Data.Name
	url = res.Data.Url
	return
}

func (d *dao) GetWorkflowStatus(c context.Context, name string) (phase string, displayName string, err error) {
	var res = struct {
		Code    int    `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Name  string           `json:"name"`
			Phase string           `json:"phase"`
			Nodes map[string]*Node `json:"nodes"`
		} `json:"data"`
	}{}
	params := url.Values{}
	params.Set("name", name)
	if err = d.httpCli.Get(c, d.getWorkflowStatusURI(), "", params, &res); err != nil {
		log.Errorc(c, "d.GetWorkflowStatus name:%s error:%v", name, err)
		return
	}
	if res.Code != _success {
		err = ecode.New(res.Code)
		log.Errorc(c, "d.GetWorkflowStatus name:%s res:%+v error:%v", name, res, err)
	}
	if res.Data.Name == "" {
		err = ecode.NothingFound
		log.Errorc(c, "d.GetWorkflowStatus name:%s res:%+v", name, res)
		return
	}
	phase = res.Data.Phase
	for _, v := range res.Data.Nodes {
		if v.DisplayName == name {
			continue
		}
		displayName = v.DisplayName
		switch v.Phase {
		case NodeStatusRunning, NodeStatusPending:
			v.Phase = NodeStatusRunning
			displayName, phase = genDisplayName(displayName, v.Phase)
			return
		case NodeStatusFailed, NodeStatusError:
			v.Phase = NodeStatusFailed
			err = errors.New(v.Message)
			displayName, phase = genDisplayName(displayName, v.Phase)
			return
		case NodeStatusSucceeded:
			if displayName == DisplayNameProdDeploy {
				displayName, phase = genDisplayName(displayName, v.Phase)
				return
			}
		}
	}
	return
}

func genDisplayName(displayName, phase string) (string, string) {
	if strings.Contains(displayName, DisplayNameBuild) {
		return DisplayNameBuild, phase
	}
	if DisplayNameMap[displayName] <= DisplayUatSuspend {
		if displayName == DisplayNameUatSuspend && phase == NodeStatusRunning {
			return DisplayNameUatSuspend, NodeStatusSucceeded
		}
		return DisplayNameUatSuspend, phase
	}
	if DisplayNameMap[displayName] <= DisplayPreSuspend {
		if displayName == DisplayNamePreSuspend && phase == NodeStatusRunning {
			return DisplayNamePreSuspend, NodeStatusSucceeded
		}
		return DisplayNamePreSuspend, phase
	}
	if DisplayNameMap[displayName] <= DisplayProdSuspend {
		if displayName == DisplayNameProdSuspend && phase == NodeStatusRunning {
			return DisplayNameProdSuspend, NodeStatusSucceeded
		}
		return DisplayNameProdSuspend, phase
	}
	if DisplayNameMap[displayName] == DisplayProdDeploy {
		return DisplayNameProdDeploy, phase
	}
	return "", ""
}

func (d *dao) OutputWorkflow(c context.Context, name, displayName string) (text OutputsData, err error) {
	var res = struct {
		Code    int         `json:"status"`
		Message string      `json:"message"`
		Data    OutputsData `json:"data"`
	}{}
	params := url.Values{}
	params.Set("name", name)
	params.Set("display_name", displayName)
	if err = d.httpCli.Get(c, d.outputWorkflowURI(), "", params, &res); err != nil {
		log.Errorc(c, "d.OutputWorkflow name:%s displayName:%s error:%v", name, displayName, err)
		return
	}
	if res.Code != _success {
		err = ecode.New(res.Code)
		log.Errorc(c, "d.OutputWorkflow name:%s displayName:%s res:%+v error:%v", name, displayName, res, err)
		return
	}
	if len(res.Data.Parameters) == 0 {
		err = ecode.NothingFound
		log.Errorc(c, "d.OutputWorkflow name:%s displayName:%s res:%+v error:%v", name, displayName, res, err)
		return
	}
	text = res.Data
	return
}

func (d *dao) ResumeWorkflow(c context.Context, name, displayName string) (err error) {
	var (
		res = struct {
			Code    int    `json:"status"`
			Message string `json:"message"`
		}{}
		msg []byte
		req *http.Request
	)
	params := &ResumeWorkflowParams{
		NameSpace:   _defaultNamespace,
		Name:        name,
		DisplayName: displayName,
	}
	if msg, err = json.Marshal(params); err != nil {
		return
	}
	if req, err = http.NewRequest(http.MethodPost, d.resumeWorkflowURI(), bytes.NewBuffer(msg)); err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err = d.httpCli.Do(c, req, &res); err != nil {
		log.Error("d.ResumeWorkflow params(%s) url(%s) error(%v)", msg, d.createWorkflowURI(), err)
		return
	}
	if res.Code != _success {
		err = ecode.New(res.Code)
		log.Errorc(c, "d.ResumeWorkflow params(%s) res(%+v)", msg, res)
	}
	return
}

func (d *dao) GetLogWorkflow(c context.Context, name, displayName string) (text string, err error) {
	var res = struct {
		Code    int    `json:"status"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}{}
	params := url.Values{}
	params.Set("namespace", _defaultNamespace)
	params.Set("name", name)
	params.Set("display_name", displayName)
	if err = d.httpCli.Get(c, d.logWorkflowURI(), "", params, &res); err != nil {
		log.Errorc(c, "d.GetLogWorkflow name:%s displayName:%s error:%v", name, displayName, err)
		return
	}
	if res.Code != _success {
		err = ecode.New(res.Code)
		log.Errorc(c, "d.GetLogWorkflow name:%s displayName:%s res:%+v error:%v", name, displayName, res, err)
		return
	}
	text = res.Data
	return
}

func (d *dao) StopWorkflow(c context.Context, name string) (err error) {
	var res = struct {
		Code    int    `json:"status"`
		Message string `json:"message"`
	}{}
	params := url.Values{}
	params.Set("namespace", _defaultNamespace)
	params.Set("name", name)
	if err = d.httpCli.Post(c, d.stopWorkflowURI(), "", params, &res); err != nil {
		log.Errorc(c, "d.StopWorkflow name:%s error:%v", name, err)
		return
	}
	if res.Code != _success {
		err = ecode.New(res.Code)
		log.Errorc(c, "d.StopWorkflow name:%s res:%+v error:%v", name, res, err)
	}
	return
}
