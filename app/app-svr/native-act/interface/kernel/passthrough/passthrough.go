package passthrough

import (
	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"go-common/library/log"
)

var (
	marshal = jsonpb.Marshaler{EnumsAsInts: true}
)

func ResolveRawParamsOfReq(rawParams string, newParams proto.Message, setf func(proto.Message) bool) bool {
	if rawParams == "" {
		return false
	}
	if err := jsonpb.UnmarshalString(rawParams, newParams); err != nil {
		log.Error("Fail to UnmarshalString passthrough rawParams, params=%+v error=%+v", rawParams, err)
		return false
	}
	return setf(newParams)
}

func Marshal(in proto.Message) string {
	out, err := marshal.MarshalToString(in)
	if err != nil {
		log.Error("Fail to marshal proto.Message, msg=%+v error=%+v", in, err)
		return ""
	}
	return out
}
