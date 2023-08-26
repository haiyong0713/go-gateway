package channel

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/utils/collection"
	"go-gateway/app/app-svr/app-channel/interface/conf"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

func makeContentFlowControlInfosV2Params(config *conf.CfcSvrConfig, aids []int64) *cfcgrpc.FlowCtlInfosReq {
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", config.Source)
	params.Set("business_id", strconv.FormatInt(config.BusinessID, 10))
	params.Set("ts", strconv.FormatInt(ts, 10))
	params.Set("oids", collection.JoinSliceInt(aids, ","))
	return &cfcgrpc.FlowCtlInfosReq{
		Oids:       aids,
		BusinessId: int32(config.BusinessID),
		Source:     config.Source,
		Sign:       getFlowCtlInfosReqSign(params, config.Secret),
		Ts:         ts,
	}
}

func getFlowCtlInfosReqSign(params url.Values, secret string) string {
	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	var buf bytes.Buffer
	buf.WriteString(tmp)
	buf.WriteString(secret)
	mh := md5.Sum(buf.Bytes())
	return hex.EncodeToString(mh[:])
}

//nolint:unparam
func getAttrBitValueFromInfosV2(reply *cfcgrpc.FlowCtlInfosV2Reply, aid int64, arcsAttrKey string) int32 {
	if reply == nil {
		return 0
	}
	val, ok := reply.ItemsMap[aid]
	if !ok {
		return 0
	}
	for _, v := range val.Items {
		if v.Key == arcsAttrKey {
			return v.Value
		}
	}
	return 0
}
