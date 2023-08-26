package fawkes

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go-gateway/app/app-svr/fawkes/service/conf"
	wf "go-gateway/app/app-svr/fawkes/service/model/workflow"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func (d *Dao) CreateWorkflow(c context.Context, workflow *wf.Workflow, conf *conf.Comet) (workflowId string, err error) {
	if workflow == nil {
		return
	}
	params := make(map[string]interface{})
	params["title"] = workflow.Title
	params["workflow_name"] = workflow.Name
	params["sponsor"] = workflow.Operator
	for k, v := range workflow.Params {
		params[k] = v
	}
	reqBody, err := json.MarshalIndent(params, "", " ")
	if err != nil {
		log.Errorc(c, "MarshalIndent error %v", err)
		return
	}
	log.Infoc(c, "CreateWorkflow:\n%v", string(reqBody))
	payload := strings.NewReader(string(reqBody))
	toRequestUrl := conf.WorkflowUrl
	req, _ := http.NewRequest(http.MethodPost, toRequestUrl, payload)
	req.Header.Add("content-type", "application/json; charset=utf-8")
	req.Header.Add("x-secretid", conf.SecretID)
	req.Header.Add("x-signature", conf.Signature)
	var re struct {
		Code int64 `json:"code"`
		Data struct {
			WorkflowId string `json:"process_id"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err = d.httpClient.Do(c, req, &re); err != nil {
		log.Errorc(c, "d.httpClient.Do error %v", err)
		return
	}
	return re.Data.WorkflowId, err
}
