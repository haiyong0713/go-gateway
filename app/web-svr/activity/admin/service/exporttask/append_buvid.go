package exporttask

import (
	"context"
	"go-gateway/pkg/idsafe/bvid"
	"strconv"
)

type BvidAppend struct {
	Field string
}

func (b *BvidAppend) Append(c context.Context, taskRet []map[string]string) []map[string]string {
	for i, one := range taskRet {
		avid, _ := strconv.ParseInt(one[b.Field], 10, 64)
		if avid > 0 {
			taskRet[i]["bvid"], _ = bvid.AvToBv(avid)
		}
	}
	return taskRet
}
