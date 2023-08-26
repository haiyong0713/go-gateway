package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-common/library/ecode"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model/rename"

	"github.com/pkg/errors"
)

func (rd *renameDao) Rename(ctx context.Context, in *rename.Rename) error {
	inj, err := json.Marshal(in)
	if err != nil {
		return err
	}
	kvReq := rd.kv.NewPutReq([]byte(renameKeyFormat(in.ID)), inj)
	if err := rd.kv.Put(ctx, kvReq); err != nil {
		return err
	}
	return nil
}

func (rd *renameDao) FetchRenameInfo(ctx context.Context, ID string) (*rename.Rename, error) {
	kvReq := rd.kv.NewGetReq([]byte(renameKeyFormat(ID)))
	record, err := rd.kv.Get(ctx, kvReq)
	if err != nil {
		if ecode.Cause(err) != ecode.NothingFound {
			return nil, err
		}
		return &rename.Rename{}, nil
	}
	renameInfo := &rename.Rename{}
	if err := json.Unmarshal(record.Columns[0].Value, renameInfo); err != nil {
		return nil, err
	}
	return renameInfo, nil
}

func (rd *renameDao) BatchFetchRenameInfo(ctx context.Context, ids []string) (map[string]*rename.Rename, error) {
	var keys []string
	for _, id := range ids {
		keys = append(keys, renameKeyFormat(id))
	}
	kvReq := rd.kv.NewBatchGetReq(ctx, keys)
	resp, err := rd.kv.BatchGet(ctx, kvReq)
	if err != nil {
		return nil, err
	}
	if !resp.AllSucceed {
		return nil, errors.Errorf("BatchFetchRenameInfo not all succeed")
	}
	renameMap := make(map[string]*rename.Rename)
	for _, v := range resp.Records {
		id := renameKeyDecode(string(v.Key))
		renameInfo := &rename.Rename{}
		if err := json.Unmarshal(v.Columns[0].Value, renameInfo); err != nil {
			continue
		}
		renameMap[id] = renameInfo
	}
	return renameMap, nil
}

func renameKeyFormat(in string) string {
	return fmt.Sprintf("rename_%s", in)
}

func renameKeyDecode(in string) string {
	return strings.ReplaceAll(in, "rename_", "")
}
