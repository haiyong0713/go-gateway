package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	pb "go-gateway/app/app-svr/app-gw/management/api"

	"github.com/pkg/errors"
)

func ParseClientInfo(dst *pb.SetDynPathReq, req *http.Request) error {
	clientInfo := req.Form.Get("client_info")
	if clientInfo == "" {
		return nil
	}
	dst.ClientInfo = &pb.ClientInfo{}
	if err := json.Unmarshal([]byte(clientInfo), dst.ClientInfo); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func ParseAction(dst *pb.SetBreakerAPIReq, req *http.Request) error {
	action := req.Form.Get("action")
	if action == "" {
		return nil
	}
	null := &pb.BreakerByNull{}
	if err := json.Unmarshal([]byte(action), null); err != nil {
		return errors.WithStack(err)
	}
	ba := &pb.BreakerAction{}
	switch null.Name {
	case "null", "":
		null.Name = "null"
		ba.Action = &pb.BreakerAction_Null{Null: null}
	case "ecode":
		ecode := &pb.BreakerByEcode{}
		if err := json.Unmarshal([]byte(action), ecode); err != nil {
			return errors.WithStack(err)
		}
		ba.Action = &pb.BreakerAction_Ecode{Ecode: ecode}
	case "placeholder":
		placeholder := &pb.BreakerByPlaceholder{}
		if err := json.Unmarshal([]byte(action), placeholder); err != nil {
			return errors.WithStack(err)
		}
		checker := make(map[string]interface{})
		err := json.Unmarshal([]byte(placeholder.Data), &checker)
		if err != nil {
			return errors.WithStack(err)
		}
		ba.Action = &pb.BreakerAction_Placeholder{Placeholder: placeholder}
	case "directly_backup":
		directlyBackup := &pb.BreakerByDirectlyBackup{}
		if err := json.Unmarshal([]byte(action), directlyBackup); err != nil {
			return errors.WithStack(err)
		}
		u, err := url.Parse(directlyBackup.BackupUrl)
		if err != nil {
			return errors.WithStack(err)
		}
		if u.Scheme == "" || u.Host == "" {
			return errors.WithStack(fmt.Errorf("invalid backup_url: %v", directlyBackup.BackupUrl))
		}
		ba.Action = &pb.BreakerAction_DirectlyBackup{
			DirectlyBackup: directlyBackup,
		}
	case "retry_backup":
		retryBackup := &pb.BreakerByRetryBackup{}
		if err := json.Unmarshal([]byte(action), retryBackup); err != nil {
			return errors.WithStack(err)
		}
		u, err := url.Parse(retryBackup.BackupUrl)
		if err != nil {
			return errors.WithStack(err)
		}
		if u.Scheme == "" || u.Host == "" {
			return errors.WithStack(fmt.Errorf("invalid backup_url: %v", retryBackup.BackupUrl))
		}
		ba.Action = &pb.BreakerAction_RetryBackup{RetryBackup: retryBackup}
	}
	dst.Action = ba
	return nil
}
