package dao

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model"

	"go-common/library/ecode"

	tus "git.bilibili.co/bapis/bapis-go/datacenter/service/titan"

	"github.com/pkg/errors"
)

func (td *tusEditDao) FetchTargetTusValue(ctx context.Context, tusValues []string, mid int64) (tusValue string, err error) {
	reply, err := td.tus.CheckTagBatch(ctx, &tus.TusBatchRequest{
		Uid:       strconv.FormatInt(mid, 10),
		UidType:   "mid",
		BizType:   "gateway",
		Condition: buildConditionForTus(tusValues),
		Sign:      fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s%s", "2dd8dac606d1", strconv.FormatInt(mid, 10))))),
	})
	if err != nil {
		return "", err
	}
	if reply.Code != 0 {
		return "", errors.Errorf("tus CheckTagBatch error code(%d)", reply.Code)
	}
	if len(tusValues) != len(reply.Hits) {
		return "", errors.Errorf("tus reply length not match with tus values")
	}
	for index, hit := range reply.Hits {
		if hit {
			return tusValues[index], nil
		}
	}
	return model.DefaultTusValue, nil
}

func buildConditionForTus(tusValues []string) []string {
	var conditions []string
	for _, v := range tusValues {
		conditions = append(conditions, fmt.Sprintf("tag_%s==1", v))
	}
	return conditions
}

func (td *tusEditDao) MigrateTusValueWithMids(ctx context.Context, tusValue string, mids map[int64]string) error {
	const _migrate = 2
	nowStr := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	req := &tus.CrowdTransParam{
		Sign:             getTitanSign(nowStr, "APP云控", "0f06d32ba2e1b42e5ca34725cb3f87818cdfffa9"),
		Timestamp:        nowStr,
		User:             "APP云控",
		BusinessDomainId: 12,
		Operation:        _migrate,
		TransInfos: buildTransInfo(mids, func(originTusValue string, mid int64) *tus.TransInfo {
			return &tus.TransInfo{
				Uid:           strconv.FormatInt(mid, 10),
				SourceCrowdId: originTusValue,
				DestCrowdId:   tusValue,
			}
		}),
	}
	reply, err := td.crowed.CrowdTransplant(ctx, req)
	if err != nil {
		return err
	}
	if reply.Code != int64(ecode.OK) {
		return errors.Errorf("MigrateTusValueWithMids error code(%d), failed mids(%+v)", reply.Code, reply.TransInfos)
	}
	return nil
}

func (td *tusEditDao) MigrateTusValueToDefaultWithMids(ctx context.Context, tusValue string, mids map[int64]string) error {
	const _delete = 0
	nowStr := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	req := &tus.CrowdTransParam{
		Sign:             getTitanSign(nowStr, "APP云控", "0f06d32ba2e1b42e5ca34725cb3f87818cdfffa9"),
		Timestamp:        nowStr,
		User:             "APP云控",
		BusinessDomainId: 12,
		Operation:        _delete,
		TransInfos: buildTransInfo(mids, func(originTusValue string, mid int64) *tus.TransInfo {
			return &tus.TransInfo{
				Uid:           strconv.FormatInt(mid, 10),
				SourceCrowdId: originTusValue,
			}
		}),
	}
	reply, err := td.crowed.CrowdTransplant(ctx, req)
	if err != nil {
		return err
	}
	if reply.Code != int64(ecode.OK) {
		return errors.Errorf("MigrateTusValueToDefaultWithMids error code(%d), failed mids(%+v)", reply.Code, reply.TransInfos)
	}
	return nil
}

func (td *tusEditDao) PutinTusValueWithMids(ctx context.Context, tusValue string, mids map[int64]string) error {
	const _add = 1
	nowStr := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	req := &tus.CrowdTransParam{
		Sign:             getTitanSign(nowStr, "APP云控", "0f06d32ba2e1b42e5ca34725cb3f87818cdfffa9"),
		Timestamp:        nowStr,
		User:             "APP云控",
		BusinessDomainId: 12,
		Operation:        _add,
		TransInfos: buildTransInfo(mids, func(originTusValue string, mid int64) *tus.TransInfo {
			return &tus.TransInfo{
				Uid:         strconv.FormatInt(mid, 10),
				DestCrowdId: tusValue,
			}
		}),
	}
	reply, err := td.crowed.CrowdTransplant(ctx, req)
	if err != nil {
		return err
	}
	if reply.Code != int64(ecode.OK) {
		return errors.Errorf("PutinTusValueWithMids error code(%d), failed mids(%+v)", reply.Code, mids)
	}
	return nil
}

func buildTransInfo(mids map[int64]string, fn func(originTusValue string, mid int64) *tus.TransInfo) []*tus.TransInfo {
	var out []*tus.TransInfo
	for mid, originTusValue := range mids {
		out = append(out, fn(originTusValue, mid))
	}
	return out
}

func encryptMD5(message string) ([]byte, error) {
	if message == "" {
		return nil, errors.New("message is not null")
	}
	d := []byte(message)
	m := md5.New()
	m.Write(d)
	return m.Sum(nil), nil
}

func byte2hex(bytes []byte) string {
	sign := hex.EncodeToString(bytes)
	return strings.ToUpper(sign)
}

func signToRequest(params map[string]string, secret string) string {
	var keys []string
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var query string
	query = query + secret
	for _, v := range keys {
		query = query + v + params[v]
	}
	query = query + secret
	fmt.Println("query:" + query)
	bytes, _ := encryptMD5(query)
	return byte2hex(bytes)
}

func getTitanSign(timestamp, user, secret string) string {
	params := make(map[string]string)
	params["user"] = user
	params["timestamp"] = timestamp
	sign := signToRequest(params, secret)
	return strings.ToUpper(sign)
}
