package rewards

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/admin/service/exporttask"
	"strings"

	"go-common/library/log"
)

const (
	sql4InsertCdKey = `
INSERT INTO rewards_cdkey_v2 (award_id, cdkey_content, unique_id)
VALUES %s;`

	sql4CountCDKey = `
SELECT COUNT(1) FROM rewards_cdkey_v2;
`

	sql4CountCDKeyByAwardId = `
SELECT COUNT(1) FROM rewards_cdkey_v2 
WHERE award_id = ?
	AND mid = 0
	AND activity_id = 0
	AND is_used = 0
`
)

func (d *Dao) UploadCdKey(ctx context.Context, userName string, batchInsertSize, awardId int64, keys []string) (err error) {
	if len(keys) == 0 {
		return
	}

	var count int64
	err = d.db.QueryRow(ctx, sql4CountCDKey).Scan(&count)
	if err != nil {
		return
	}

	tx, err := d.db.Begin(ctx)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = exporttask.SendWeChatTextMessage(ctx, []string{userName}, fmt.Sprintf("CdKey导入失败: %v", err))
			log.Errorc(ctx, "UploadCdKey error(%v)", err)
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(ctx, "UploadCdKey tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(ctx, "UploadCdKey tx.Commit() error(%v)", err)
			_ = exporttask.SendWeChatTextMessage(ctx, []string{userName}, fmt.Sprintf("CdKey导入失败: %v", err))
		} else {
			_ = exporttask.SendWeChatTextMessage(ctx, []string{userName}, fmt.Sprintf("CdKey导入成功。"))
		}
	}()
	params := strings.Builder{}
	values := make([]interface{}, 0)
	first := true
	for i := 0; i < len(keys); i++ {
		count++
		uniqIDStr := fmt.Sprintf("-%v", count)
		if first {
			params.WriteString(fmt.Sprintf("(?,?, ?)"))
			values = append(values, awardId, keys[i], uniqIDStr)
			first = false
		} else {
			params.WriteString(fmt.Sprintf(",(?,?, ?)"))
			values = append(values, awardId, keys[i], uniqIDStr)
		}
		if i%int(batchInsertSize) == 0 || i == len(keys)-1 {
			_, err = tx.Exec(fmt.Sprintf(sql4InsertCdKey, params.String()), values...)
			if err != nil {
				return
			}
			params = strings.Builder{}
			values = make([]interface{}, 0)
			first = true
		}
	}
	return
}

func (d *Dao) GetCdkeyCount(ctx context.Context, awardId int64) (count int64, err error) {
	err = d.db.QueryRow(ctx, sql4CountCDKeyByAwardId, awardId).Scan(&count)
	return
}
