package model

import (
	"regexp"
	"strings"

	pb "go-gateway/app/app-svr/app-gw/management/api"

	"github.com/dgrijalva/jwt-go"
	"github.com/gogo/protobuf/proto"
)

const (
	TotalRule     = "total"
	RefererRule   = "referer"
	_regexQuotaId = `(.*)\|(http|grpc)\|(.*)\|(.*)`
)

type BreakerAPI struct {
	Api       string      `json:"api"`
	Ratio     int64       `json:"ratio"`
	Reason    string      `json:"reason"`
	Condition string      `json:"condition"`
	Action    interface{} `json:"action"`
	Enable    bool        `json:"enable"`
	Node      string      `json:"node"`
	Gateway   string      `json:"gateway"`
	FlowCopy  interface{} `json:"flow_copy"`
}

func (ba *BreakerAPI) SetBreakerAction(in *pb.BreakerAction) {
	switch action := in.Action.(type) {
	case *pb.BreakerAction_Null:
		ba.Action = struct{}{}
	case *pb.BreakerAction_Ecode:
		ba.Action = action.Ecode
	case *pb.BreakerAction_Placeholder:
		ba.Action = action.Placeholder
	case *pb.BreakerAction_RetryBackup:
		ba.Action = action.RetryBackup
	case *pb.BreakerAction_DirectlyBackup:
		ba.Action = action.DirectlyBackup
	default:
		ba.Action = struct{}{}
	}
}

func (ba *BreakerAPI) SetFlowCopy(in *pb.FlowCopy) {
	switch flow := in.Flow.(type) {
	case *pb.FlowCopy_Null:
		ba.FlowCopy = struct{}{}
	case *pb.FlowCopy_Ratio:
		ba.FlowCopy = flow.Ratio
	case *pb.FlowCopy_Qps:
		ba.FlowCopy = flow.Qps
	default:
		ba.FlowCopy = struct{}{}
	}
}

func (ba *BreakerAPI) FromProto(in *pb.BreakerAPI) {
	ba.Api = in.Api
	ba.Ratio = in.Ratio
	ba.Reason = in.Reason
	ba.Condition = in.Condition
	ba.Enable = in.Enable
	ba.Node = in.Node
	ba.Gateway = in.Gateway
	ba.SetBreakerAction(in.Action)
	ba.SetFlowCopy(in.FlowCopy)
}

type GatewayProfile struct {
	GatewayVersion string `json:"gateway_version"`
	SDKVersion     string `json:"sdk_version"`
	ConfigDigest   string `json:"config_digest"`
}

type JWTTokenPayload struct {
	Addr    string `json:"addr"`
	Node    string `json:"node"`
	Gateway string `json:"gateway"`
	jwt.StandardClaims
}

type AddConfigFileReq struct {
	AppID      string
	TreeID     int64
	ConfigMeta *pb.ConfigMeta
	Buffer     []byte
}

type CreateConfigBuildReq struct {
	Env       string `json:"env"`
	Zone      string `json:"zone"`
	BuildName string `json:"build_name"`
	TreeId    int64  `json:"tree_id"`
	Cookie    string `json:"cookie"`
}

type FetchConfigBuildMetaReq struct {
	TreeId int64  `json:"tree_id"`
	Cookie string `json:"cookie"`
	Env    string `json:"env"`
}

type ConfigBuildMeta struct {
	Token string `json:"token"`
}

type ReloadConfigReq struct {
	Host           string `json:"host"`
	Content        string `json:"content"`
	Digest         string `json:"digest"`
	OriginalDigest string `json:"original_digest"`
	IsGRPC         bool   `json:"is_grpc"`
}

type ReloadConfigReply struct {
	Loaded string `json:"loaded"`
	Digest string `json:"digest"`
}

type QuotaConfig struct {
	ServiceID string `json:"service_id"`
	Protocol  string `json:"protocol"`
	Uri       string `json:"uri"`
	Rule      string `json:"rule"`
}

func MatchBreakerAPI(bapi *pb.BreakerAPI, ba *pb.BreakerAPI) bool {
	return proto.Equal(bapi, ba)
}

func MatchDynPath(dp *pb.DynPath, val *pb.DynPath) bool {
	return proto.Equal(CopyDynPath(dp), CopyDynPath(val))
}

func CopyDynPath(req *pb.DynPath) *pb.DynPath {
	dup := &pb.DynPath{}
	*dup = *req
	dup.UpdatedAt = 0
	return dup
}

// nolint:gomnd
func ParseQuotaConfig(id string) (QuotaConfig, bool) {
	re := regexp.MustCompile(_regexQuotaId)
	subSlice := re.FindStringSubmatch(id)
	if len(subSlice) != 5 {
		return QuotaConfig{}, false
	}
	config := QuotaConfig{
		ServiceID: subSlice[1],
		Protocol:  subSlice[2],
		Uri:       subSlice[3],
		Rule:      subSlice[4],
	}
	return config, true
}

// nolint:gomnd
func ParseZone(id string) string {
	slice := strings.Split(id, ".")
	if len(slice) > 2 {
		return slice[1]
	}
	return ""
}

func ParseRuleType(rule string) string {
	if strings.HasPrefix(rule, RefererRule) {
		return RefererRule
	}
	return TotalRule
}
