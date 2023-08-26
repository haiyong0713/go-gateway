package warden

import (
	"math/rand"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/request"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster/ab"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/prom"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

var (
	ToBackupServiceCode = prom.New().
		WithCounter("grpc_backup_call_count", []string{"method"}).
		WithState("grpc_backup_call_count_state", []string{"method"})
)

var _errStopAttempt = errors.New("ForceBackupPatcher: stop attempt")

type BackupRetryOption struct {
	Ratio                int64
	ForceBackupCondition string
	forceBackupCondition ab.Condition

	BackupAction      string
	BackupPlaceholder string
	BackupECode       int
	BackupTarget      string
}

func setupBackupRetry(r *request.Request, option *BackupRetryOption) {
	if option == nil {
		return
	}
	r.ApplyOptions(request.WithHandlerPatchers(NewForceBackupPatcher(option)))
}

// ForceBackupPatcher is
type ForceBackupPatcher struct {
	option  *BackupRetryOption
	Attempt int
}

func NewForceBackupPatcher(option *BackupRetryOption) *ForceBackupPatcher {
	return &ForceBackupPatcher{
		option:  option,
		Attempt: 1,
	}
}

// Name is
func (bp *ForceBackupPatcher) Name() string {
	return "ForceBackupPatcher"
}

// Matched is
func (bp *ForceBackupPatcher) Matched(r *request.Request) bool {
	if bp.option.Ratio == 0 {
		return false
	}
	if bp.option.Ratio < 100 && rand.Int63n(100)+1 > bp.option.Ratio {
		return false
	}
	t, ok := ab.FromContext(r.Context())
	if !ok {
		// all request should be matched if no ab environment is set
		return true
	}
	if bp.option.forceBackupCondition != nil {
		return bp.option.forceBackupCondition.Matches(t)
	}
	return true
}

func dummySuccessResponse(str string, v interface{}) error {
	pb, ok := v.(proto.Message)
	if !ok {
		return nil
	}
	return errors.WithStack(jsonpb.UnmarshalString(str, pb))
}

// Patch is
func (bp *ForceBackupPatcher) Patch(in request.Handlers) request.Handlers {
	out := in.Copy()
	out.Send.SetFrontNamed(request.NamedHandler{
		Name: "appgwsdk.warden.ab.ForceBackupPatcherSendHandler",
		Fn: func(r *request.Request) {
			switch bp.option.BackupAction {
			case "placeholder":
				err := dummySuccessResponse(bp.option.BackupPlaceholder, r.Data)
				if err != nil {
					log.Warn("Unrecognized backup placeholder: %s error: %+v", bp.option.BackupPlaceholder, err)
				}
				r.Error = _errStopAttempt
				return
			case "ecode":
				r.Data = nil
				r.Error = ecode.Int(bp.option.BackupECode)
				return
			case "retry_backup":
				if r.RetryCount < bp.Attempt {
					return
				}
				// TODO: configurable backup conn
				backupCC, err := pooledClientConn(bp.option.BackupTarget, nil)
				if err != nil {
					log.Warn("Failed to fetch a pooled conn: %q: %+v", bp.option.BackupTarget, err)
					return
				}
				r.Operation.CC = backupCC
				r.Operation.AppID = bp.option.BackupTarget
				ToBackupServiceCode.Incr(r.Operation.Method)
				return
			case "directly_backup":
				backupCC, err := pooledClientConn(bp.option.BackupTarget, nil)
				if err != nil {
					log.Warn("Failed to fetch a pooled conn: %q: %+v", bp.option.BackupTarget, err)
					return
				}
				r.Operation.CC = backupCC
				r.Operation.AppID = bp.option.BackupTarget
				ToBackupServiceCode.Incr(r.Operation.Method)
				return
			default:
				log.Warn("Unrecognized backup action: %s", bp.option.BackupAction)
			}
		},
	})
	out.Send.AfterEachFn = func(item request.HandlerListRunItem) bool {
		if item.Request.Error == _errStopAttempt {
			item.Request.Error = nil
			return false
		}
		return request.HandlerListStopOnError(item)
	}
	out.ValidateResponse.SetFrontNamed(request.NamedHandler{
		Name: "appgwsdk.warden.ab.ForceBackupPatcherCompleteAttemptHandler",
		Fn: func(r *request.Request) {
			if r.Error == _errStopAttempt {
				r.Error = nil
				return
			}
		},
	})
	return out
}
