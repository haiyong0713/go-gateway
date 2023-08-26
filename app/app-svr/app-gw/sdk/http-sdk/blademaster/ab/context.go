package ab

import (
	v1 "go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster/ab/internal/proto"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const (
	_headerKey = "x-exp-abtest-bin"
)

var (
	// ErrUnknownState is an error if state cannot be unmarshal properly.
	ErrUnknownState = errors.New("ab: unknown state")
)

// StateType stands for different types of internal layer states of experiment.
type StateType int32

const (
	StateUnknown StateType = iota
	ExpHit
	ExpHittable
	LayerNoHit
	LayerConflict
)

// State holds snapshot of T, which can be passed to downstream sub-systems.
type State struct {
	Type  StateType
	Value int64
}

// Carrier is abstraction holding states.
type Carrier interface {
	Get() ([]State, []KV)
	Set([]State, []KV)
}

// GRPCCarrier holds states in grpc metadata.
type GRPCCarrier metadata.MD

// Get returns unmarshaled states from grpc metadata.
func (g GRPCCarrier) Get() ([]State, []KV) {
	if s, ok := g[_headerKey]; ok && len(s) > 0 {
		return Unmarshal([]byte(s[0]))
	}
	return nil, nil
}

// Set inject marshaled states into grpc metadata.
func (g GRPCCarrier) Set(states []State, kvs []KV) {
	b, err := Marshal(states, kvs)
	if err != nil {
		return
	}
	if len(b) > 0 {
		g[_headerKey] = append(g[_headerKey], string(b))
	}
}

func Marshal(states []State, kvs []KV) (b []byte, err error) {
	ctx := &v1.Context{
		Env: make(map[string]*v1.Val),
	}
	for _, st := range states {
		ctx.States = append(ctx.States, &v1.State{
			Type:  v1.StateType(st.Type),
			Value: st.Value,
		})
	}
	for _, kv := range kvs {
		v := fromKV(kv)
		ctx.Env[kv.Key] = &v
	}
	return proto.Marshal(ctx)
}

func Unmarshal(b []byte) (states []State, kvs []KV) {
	var ctx v1.Context
	if err := proto.Unmarshal(b, &ctx); err != nil {
		return nil, nil
	}
	for _, st := range ctx.States {
		states = append(states, State{
			Type:  StateType(st.Type),
			Value: st.Value,
		})
	}
	for k, v := range ctx.Env {
		kv := toKV(k, v)
		kvs = append(kvs, kv)
	}
	return
}

func fromKV(kv KV) (v v1.Val) {
	v.Type = v1.VarType(kv.Type)
	switch kv.Type {
	case typeString:
		v.SVal = kv.String
	case typeInt64:
		v.IVal = kv.Int64
	case typeFloat64:
		v.FVal = kv.Float64
	case typeBool:
		v.BVal = kv.Bool
	case typeVersion:
		v.SVal = kv.Version.String()
	default:
	}
	return
}

func toKV(k string, v *v1.Val) (kv KV) {
	kv.Key = k
	kv.Type = varType(v.Type)
	switch v.Type {
	case v1.VarType_TYPE_STRING:
		kv.String = v.SVal
	case v1.VarType_TYPE_INT64:
		kv.Int64 = v.IVal
	case v1.VarType_TYPE_FLOAT64:
		kv.Float64 = v.FVal
	case v1.VarType_TYPE_BOOL:
		kv.Bool = v.BVal
	case v1.VarType_TYPE_VERSION:
		kv.Version, _ = newVersion(v.SVal)
	default:
	}
	return
}
